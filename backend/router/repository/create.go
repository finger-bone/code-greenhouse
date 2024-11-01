package repository

import (
	"fmt"
	"io"
	"judge/challenge"
	"judge/jConfig"
	"judge/middleware"
	"judge/router"
	"judge/schema"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func generateFlakeId() string {
	// get a random bytes array of size 2
	randomPart := []byte{
		byte(rand.Intn(256)),
		byte(rand.Intn(256)),
	}

	raw := fmt.Sprintf(
		"%s%d",
		string(randomPart),
		time.Now().UnixMilli()-time.Date(2024, 10, 23, 23, 59, 59, 0, time.UTC).UnixMilli(),
	)
	encoded := base58.Encode([]byte(raw))

	return encoded
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	stat, err := srcFile.Stat()
	if err != nil {
		return err
	}
	return dstFile.Chmod(stat.Mode())
}

func copyDir(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

func createRepositoryFiles(logger *zap.Logger, judgeConfig *jConfig.JudgeConfig, provider, subject, folderName string, startpoint *challenge.StartPoint, repoId string) error {
	// first, create a folder under
	// startpointRootPath := fmt.Sprintf("%s/%s/%s", judgeConfig.Challenge.StorageFolder, folderName, startpoint.Root)
	startpointRootPath := filepath.Join(judgeConfig.Challenge.StorageFolder, folderName, startpoint.Root)
	repositoryPath := filepath.Join(
		judgeConfig.RepositoryStorage.StorageFolder,
		provider,
		subject,
		folderName,
		repoId,
	)
	if err := os.MkdirAll(repositoryPath, 0755); err != nil {
		logger.Error("Failed to create repository folder",
			zap.String("path", repositoryPath),
			zap.Error(err),
		)
		return err
	}
	// copy all the files from startpointRootPath to repositoryPath
	if err := copyDir(startpointRootPath, repositoryPath); err != nil {
		logger.Error("Failed to copy files from startpoint root to repository",
			zap.String("startpoint", startpoint.Name),
			zap.String("startpointRootPath", startpointRootPath),
			zap.String("repositoryPath", repositoryPath),
			zap.Error(err),
		)
		return err
	}
	// init a git repository
	repo, err := git.PlainInit(repositoryPath, false)
	if err != nil {
		logger.Error("Failed to initialize git repository",
			zap.String("repositoryPath", repositoryPath),
			zap.Error(err),
		)
		return err
	}

	// add all files to the repository
	worktree, err := repo.Worktree()
	if err != nil {
		logger.Error("Failed to get worktree",
			zap.String("repositoryPath", repositoryPath),
			zap.Error(err),
		)
		return err
	}

	_, err = worktree.Add(".")
	if err != nil {
		logger.Error("Failed to add files to git repository",
			zap.String("repositoryPath", repositoryPath),
			zap.Error(err),
		)
		return err
	}

	// commit the changes
	commit, err := worktree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "gardener-bot",
			Email: "gardener-bot@greenhouse.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		logger.Error("Failed to commit changes",
			zap.String("repositoryPath", repositoryPath),
			zap.Error(err),
		)
		return err
	}

	// log the commit
	obj, err := repo.CommitObject(commit)
	if err != nil {
		logger.Error("Failed to get commit object",
			zap.String("repositoryPath", repositoryPath),
			zap.Error(err),
		)
		return err
	}
	// git config receive.denyCurrentBranch updateInstead
	cfg, err := repo.Config()
	if err != nil {
		logger.Error("Failed to get git config",
			zap.String("repositoryPath", repositoryPath),
			zap.Error(err))
		return err
	}

	cfg.Raw.AddOption("receive", "", "denyCurrentBranch", "updateInstead")

	err = repo.SetConfig(cfg)
	if err != nil {
		logger.Error("Failed to set git config",
			zap.String("repositoryPath", repositoryPath),
			zap.Error(err))
		return err
	}

	// add a hook to the repository
	hookPath := filepath.Join(repositoryPath, ".git", "hooks", "post-commit")
	hookContent := fmt.Sprintf(`
	# !/bin/sh
	# This hook is used to initiate a test after a commit is made
	curl -X POST %s:%d/testing/pending?repo=%s
	`, judgeConfig.Server.HostAddr, judgeConfig.Server.HostPort, repoId)
	if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
		logger.Error("Failed to write hook file",
			zap.String("hookPath", hookPath),
			zap.Error(err),
		)
		return err
	}

	logger.Info("Repository initialized and initial commit created",
		zap.String("repositoryPath", repositoryPath),
		zap.String("commit", obj.String()),
	)

	return nil
}

func BuildNewRepositoryHandler(logger *zap.Logger, config *jConfig.JudgeConfig, db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		folderName := c.Query("folder")
		startpointName := c.Query("startpoint")
		if folderName == "" {
			return c.Status(fiber.StatusBadRequest).JSON(router.BuildError(
				"Folder name is required",
			))
		}
		challengeInfo, err := challenge.ParseChallenge(logger, &config.Challenge, folderName)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(router.BuildError(
				"Failed to parse challenge",
			))
		}
		startpoint := challengeInfo.FindStartPoint(startpointName)
		if startpoint == nil {
			return c.Status(fiber.StatusBadRequest).JSON(router.BuildError(
				"Startpoint not found",
			))
		}

		provider := c.Locals(middleware.PROVIDER_LOCAL_KEY).(string)
		subject := c.Locals(middleware.SUBJECT_LOCAL_KEY).(string)
		repoId := generateFlakeId()
		createRepositoryFiles(logger, config, provider, subject, folderName, startpoint, repoId)
		db.Create(&schema.Repository{
			RepositoryId:        repoId,
			Subject:             subject,
			Provider:            provider,
			ChallengeFolderName: folderName,
			Startpoint:          startpoint.Name,
			Stage:               0,
			TotalStages:         int32(len(challengeInfo.Stages)),
			CreateTime:          time.Now().Format(time.RFC3339),
			UpdateTime:          time.Now().Format(time.RFC3339),
		})
		return c.JSON(router.BuildResponse(
			struct {
				RepositoryId string `json:"repositoryId"`
			}{
				RepositoryId: repoId,
			},
		))
	}
}
