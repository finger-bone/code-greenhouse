package bootstrap

import (
	"os"

	"judge/jConfig"

	"go.uber.org/zap"
)

func bootstrapFolders(logger *zap.Logger, config *jConfig.JudgeConfig) {

	// Create storage folder for challenges
	if _, err := os.Stat(config.Challenge.StorageFolder); os.IsNotExist(err) {
		err = os.Mkdir(config.Challenge.StorageFolder, os.ModePerm)
		if err != nil {
			logger.Panic("Failed to create storage folder for challenges.")
		}
	}

	// Create storage folder for repositories
	if _, err := os.Stat(config.RepositoryStorage.StorageFolder); os.IsNotExist(err) {
		err = os.Mkdir(config.RepositoryStorage.StorageFolder, os.ModePerm)
		if err != nil {
			logger.Panic("Failed to create storage folder for repositories.")
		}
	}

}
