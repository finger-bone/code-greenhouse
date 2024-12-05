package challenge

import (
	"judge/jConfig"
	"os"
	"path/filepath"
	"regexp"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
)

const ATTRIBUTE_FILE_NAME = "attribute.toml"

type StartPoint struct {
	Name        string
	Description []string
	Root        string
	Dockerfile  string
}

type Stage struct {
	Name           string
	Description    []string
	NoteFileOrPath string
	NoteFileType   string
}

type Basic struct {
	Author      string
	Source      string
	Title       string
	Description []string
}

type Challenge struct {
	FolderName  string       `toml:"-"`
	Basic       Basic        `toml:"basic"`
	StartPoints []StartPoint `toml:"startpoints"`
	Stages      []Stage      `toml:"stages"`
}

func ParseChallenge(logger *zap.Logger, config *jConfig.ChallengeConfig, folderName string) (*Challenge, error) {
	// challengeAttributeFilePath := fmt.Sprintf("%s/%s/%s", config.StorageFolder, folderName, ATTRIBUTE_FILE_NAME)
	challengeAttributeFilePath := filepath.Join(config.StorageFolder, folderName, ATTRIBUTE_FILE_NAME)
	var challenge Challenge
	if _, err := toml.DecodeFile(challengeAttributeFilePath, &challenge); err != nil {
		logger.Error("Failed to parse challenge attribute file",
			zap.String("path", challengeAttributeFilePath),
			zap.Error(err),
		)
		return nil, err
	}
	challenge.FolderName = folderName

	return &challenge, nil
}

func ParseAllChallenges(logger *zap.Logger, config *jConfig.ChallengeConfig) ([]Challenge, error) {
	entries, err := os.ReadDir(config.StorageFolder)
	if err != nil {
		logger.Error("Failed to read challenge storage folder",
			zap.String("path", config.StorageFolder),
			zap.Error(err),
		)
		return nil, err
	}

	ignorePatterns := make([]*regexp.Regexp, len(config.IgnorePatterns))
	for i, pattern := range config.IgnorePatterns {
		ignorePatterns[i] = regexp.MustCompile(pattern)
	}

	var folderNames []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		for _, pattern := range ignorePatterns {
			if pattern.MatchString(entry.Name()) {
				continue
			}
		}
		folderNames = append(folderNames, entry.Name())
	}

	challenges := make([]Challenge, 0, len(folderNames))
	for _, folderName := range folderNames {
		challenge, err := ParseChallenge(logger, config, folderName)
		if err != nil {
			logger.Error("Failed to parse challenge",
				zap.String("folderName", folderName),
				zap.Error(err),
			)
			return nil, err
		}
		challenges = append(challenges, *challenge)
	}

	return challenges, nil
}

func (c *Challenge) FindStartPoint(name string) *StartPoint {
	for _, startPoint := range c.StartPoints {
		if startPoint.Name == name {
			return &startPoint
		}
	}
	return nil
}
