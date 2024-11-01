package tester

import (
	"judge/jConfig"
	"time"

	"github.com/docker/docker/client"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func StartListener(
	logger *zap.Logger,
	config *jConfig.JudgeConfig,
	db *gorm.DB,
	docker *client.Client,
) {
	queue := GetTestingQueue(config)
	for {
		// wait for semaphore
		<-queue.Semaphore
		// get task from queue
		task := <-queue.PendingQueue
		logger.Info("Got task", zap.String("repository_id", task.RepositoryId), zap.Int("serial", task.Serial), zap.Int("stage", task.Stage))
		// run task
		err := runTask(logger, config, db, docker, &task)
		if err != nil {
			logger.Error("Failed to run task", zap.Error(err))
			task.TestingRecord.Status = StatusError
			task.TestingRecord.RunEndTime = time.Now().Format(time.RFC3339)
			// update the task status
			if err := db.Save(&task.TestingRecord).Error; err != nil {
				logger.Error("Failed to update task status", zap.Error(err))
			}
			continue
		}
		// release semaphore
		queue.Semaphore <- true
	}
}
