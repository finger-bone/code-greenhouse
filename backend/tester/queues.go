package tester

import (
	"judge/jConfig"
)

type TestingQueue struct {
	PendingQueue chan TestingTask
	Semaphore    chan bool
}

var i *TestingQueue = nil

func GetTestingQueue(config *jConfig.JudgeConfig) *TestingQueue {
	if i == nil {
		i = &TestingQueue{
			PendingQueue: make(chan TestingTask, config.Testing.PendingQueueSize),
			Semaphore:    make(chan bool, config.Testing.MaxConcurrentWorkers),
		}
		for idx := 0; idx < config.Testing.MaxConcurrentWorkers; idx++ {
			i.Semaphore <- true
		}
	}
	return i
}
