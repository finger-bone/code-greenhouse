package tester

import (
	"judge/challenge"
	"judge/jConfig"
	"judge/middleware"
	"judge/router"
	"judge/schema"
	"time"

	"github.com/docker/docker/client"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	StatusPending        = "pending"
	StatusRunning        = "running"
	StatusSuccess        = "success"
	StatusFailed         = "failed"
	StatusError          = "error"
	StatusWaitingTimeout = "waitingTimeout"
	StatusRunningTimeout = "runningTimeout"
)

type TestingTask struct {
	RepositoryId     string
	Serial           int
	Stage            int
	Challenge        challenge.Challenge
	TestingRecord    *schema.Testing
	WaitingStartTime time.Time
}

func pushToPending(
	logger *zap.Logger,
	config *jConfig.JudgeConfig,
	db *gorm.DB,
	repositoryId string,
	stage int,
) error {
	repositoryRecord := &schema.Repository{}
	err := db.Where("repository_id = ?", repositoryId).First(repositoryRecord).Error
	if err != nil {
		logger.Error("Failed to get repository record", zap.Error(err))
		return err
	}
	folderName := repositoryRecord.ChallengeFolderName
	challengeRecord, err := challenge.ParseChallenge(
		logger,
		&config.Challenge,
		folderName,
	)
	if err != nil {
		logger.Error("Failed to parse challenge", zap.Error(err))
		return err
	}
	var repositoryTestingSerial schema.RepositoryTestingSerial
	err = db.Where("repository_id = ?", repositoryId).First(&repositoryTestingSerial).Error
	if err != nil {
		repositoryTestingSerial = schema.RepositoryTestingSerial{
			RepositoryId: repositoryId,
			NextSerial:   1,
		}
		err = db.Save(&repositoryTestingSerial).Error
		if err != nil {
			logger.Error("Failed to create repository testing serial", zap.Error(err))
			return err
		}
	}
	serial := repositoryTestingSerial.NextSerial
	// update next serial
	repositoryTestingSerial.NextSerial = serial + 1
	err = db.Save(&repositoryTestingSerial).Error
	if err != nil {
		logger.Error("Failed to update repository testing serial", zap.Error(err))
		return err
	}

	testingQueue := GetTestingQueue(config)
	testingRecord := schema.Testing{
		RepositoryId: repositoryId,
		Serial:       int32(serial),
		Stage:        int32(stage),
		Status:       StatusPending,
		CreateTime:   time.Now().Format(time.RFC3339),
	}
	err = db.Save(&testingRecord).Error
	if err != nil {
		logger.Error("Failed to create testing record", zap.Error(err))
		return err
	}
	testingQueue.PendingQueue <- TestingTask{
		RepositoryId:     repositoryId,
		Serial:           serial,
		Stage:            stage,
		Challenge:        *challengeRecord,
		TestingRecord:    &testingRecord,
		WaitingStartTime: time.Now(),
	}
	return nil
}

func BuildPushToPendingHandler(
	logger *zap.Logger,
	config *jConfig.JudgeConfig,
	db *gorm.DB,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		repositoryId := c.Query("repo")
		stage := c.QueryInt("stage", -1)

		repositoryRecord := &schema.Repository{}
		err := db.Where("repository_id = ?", repositoryId).First(repositoryRecord).Error
		if err != nil {
			logger.Error("Failed to get repository record", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to get repository record",
				"error":   err.Error(),
			})
		}
		if repositoryRecord.Provider != c.Locals(middleware.PROVIDER_LOCAL_KEY).(string) {
			logger.Error("Repository ID mismatch", zap.String("repository_id", repositoryId))
			return c.Status(fiber.StatusUnauthorized).JSON(router.BuildError(
				"Repository ID mismatch",
			))
		}
		if repositoryRecord.Subject != c.Locals(middleware.SUBJECT_LOCAL_KEY).(string) {
			logger.Error("Subject mismatch", zap.String("repository_id", repositoryId))
			return c.Status(fiber.StatusUnauthorized).JSON(router.BuildError(
				"Subject mismatch",
			))
		}
		if stage == -1 {
			stage = int(repositoryRecord.Stage)
		}

		err = pushToPending(logger, config, db, repositoryId, stage)
		if err != nil {
			logger.Error("Failed to push to pending", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(router.BuildError(
				"Failed to push to pending",
			))
		}
		return c.Status(fiber.StatusOK).JSON(router.BuildResponse(
			struct {
				Message string `json:"message"`
			}{
				Message: "Successfully pushed to pending",
			},
		))
	}
}

func SetupTestingRouter(
	logger *zap.Logger,
	config *jConfig.JudgeConfig,
	db *gorm.DB,
	docker *client.Client,
	group *fiber.Router,
) {
	(*group).Post(
		"/pending",
		middleware.BuildAuthorizationMiddleWare(logger, config, db),
		BuildPushToPendingHandler(logger, config, db),
	)
}
