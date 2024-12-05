package bootstrap

import (
	"judge/jConfig"

	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

func bootstrapDocker(logger *zap.Logger, config *jConfig.JudgeConfig) (*client.Client, error) {
	socketPath := config.Testing.DockerSocket
	// use a local docker socket
	dockerClient, err := client.NewClientWithOpts(
		client.WithHost(socketPath),
	)
	if err != nil {
		logger.Error("Failed to create Docker client", zap.Error(err))
		return nil, err
	}
	return dockerClient, nil
}
