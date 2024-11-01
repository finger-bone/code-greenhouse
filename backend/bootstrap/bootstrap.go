package bootstrap

import (
	"fmt"

	"judge/jConfig"
	"judge/tester"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func BuildApp() (*fiber.App, *zap.Logger, *jConfig.JudgeConfig) {
	config := jConfig.GetJudgeConfig()

	logger := bootstrapLogger(&config.Logger)

	logger.Info("Starting Judge With Config,",
		zap.String("config", fmt.Sprintf("%+v", config)),
	)

	db := bootstrapDatabase(logger, &config.Database)

	bootstrapFolders(logger, &config)

	dockerClient, err := bootstrapDocker(logger, &config)
	if err != nil {
		logger.Fatal("Failed to create Docker client", zap.Error(err))
	}

	go tester.StartListener(logger, &config, db, dockerClient)
	return bootstrapServer(logger, &config, db, dockerClient), logger, &config
}

func Bootstrap() {
	app, logger, config := BuildApp()
	defer app.Shutdown()
	defer logger.Sync()
	logger.Info("Judge Started")
	app.Listen(fmt.Sprintf("%s:%d", config.Server.HostAddr, config.Server.HostPort))
	logger.Info("Judge Stopped")
}
