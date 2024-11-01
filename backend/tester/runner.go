package tester

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"judge/challenge"
	"judge/jConfig"
	"judge/schema"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const STAGE_ENV_KEY = "STAGE"

func getStartPoint(
	challengeRecord *challenge.Challenge,
	startpoint string,
) (*challenge.StartPoint, error) {
	if startpoint == "" {
		return nil, nil
	}
	var r challenge.StartPoint
	for _, v := range challengeRecord.StartPoints {
		if v.Name == startpoint {
			r = v
			break
		}
	}
	if r.Name == "" {
		return nil, nil
	}
	return &r, nil
}

func handleTaskWaitingTimeout(
	logger *zap.Logger,
	db *gorm.DB,
	task *TestingTask,
	timeoutMinutes int,
) error {
	if time.Since(task.WaitingStartTime) > time.Duration(timeoutMinutes)*time.Second {
		task.TestingRecord.Status = StatusWaitingTimeout
		task.TestingRecord.RunEndTime = time.Now().Format(time.RFC3339)
		logger.Debug("Task waiting timeout", zap.String("repository_id", task.RepositoryId), zap.Int("serial", task.Serial), zap.Int("stage", task.Stage))
		return db.Save(task.TestingRecord).Error
	}
	return nil
}

func initializeTaskExecution(
	logger *zap.Logger,
	db *gorm.DB,
	task *TestingTask,
) error {
	logger.Debug("Running task",
		zap.String("repository_id", task.RepositoryId),
		zap.Int("serial", task.Serial),
		zap.Int("stage", task.Stage))

	task.TestingRecord.Status = StatusRunning
	task.TestingRecord.RunStartTime = time.Now().Format(time.RFC3339)
	return db.Save(task.TestingRecord).Error
}

func setupExecutionPaths(
	config *jConfig.JudgeConfig,
	repositoryRecord *schema.Repository,
	startpoint *challenge.StartPoint,
	runId string,
) (string, string, string, error) {
	repositoryPath := filepath.Join(
		config.RepositoryStorage.StorageFolder,
		repositoryRecord.Provider,
		repositoryRecord.Subject,
		repositoryRecord.ChallengeFolderName,
		repositoryRecord.RepositoryId,
	)

	dockerfilePath := filepath.Join(repositoryPath, startpoint.Dockerfile)
	tempStoragePath := filepath.Join(config.Testing.TmpStorageFolder, runId)

	if err := os.MkdirAll(tempStoragePath, 0755); err != nil {
		return "", "", "", err
	}

	return repositoryPath, dockerfilePath, tempStoragePath, nil
}

func buildDockerImage(
	docker *client.Client,
	dockerfilePath string,
	imageName string,
) error {
	tar, err := archive.TarWithOptions(
		filepath.Join(
			filepath.Dir(dockerfilePath),
		), &archive.TarOptions{},
	)
	if err != nil {
		return err
	}

	buildResponse, err := docker.ImageBuild(
		context.Background(),
		tar,
		types.ImageBuildOptions{
			Dockerfile: filepath.Base(dockerfilePath),
			Tags:       []string{imageName},
			Remove:     true,
		},
	)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, buildResponse.Body); err != nil {
		return err
	}

	defer buildResponse.Body.Close()
	return nil
}

func createAndStartContainer(
	docker *client.Client,
	imageName string,
	containerName string,
	reportMountPath string,
	timeout time.Duration, // 超时时间参数
	stageEnvKey string,
	stageEnvValue string,
) (string, bool, error) {
	containerConfig := &container.Config{
		Image: imageName,
		Tty:   true,
		Env: []string{
			fmt.Sprintf("%s=%s", stageEnvKey, stageEnvValue),
		},
	}

	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: reportMountPath,
				Target: "/mnt/report",
			},
		},
	}

	containerHandle, err := docker.ContainerCreate(
		context.Background(),
		containerConfig,
		hostConfig,
		nil,
		nil,
		containerName,
	)
	if err != nil {
		return "", false, err
	}

	// 启动容器
	if err := docker.ContainerStart(context.Background(), containerHandle.ID, container.StartOptions{}); err != nil {
		return "", false, err
	}

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 等待容器完成或超时
	statusCh, errCh := docker.ContainerWait(ctx, containerHandle.ID, container.WaitConditionNotRunning)
	select {
	case <-ctx.Done(): // 超时
		// 超时则强制停止容器
		if stopErr := docker.ContainerStop(context.Background(), containerHandle.ID, container.StopOptions{}); stopErr != nil {
			return "", true, fmt.Errorf("failed to stop container on timeout: %w", stopErr)
		}
		return "", true, fmt.Errorf("container execution timed out")

	case err := <-errCh: // 容器执行出错
		if err != nil {
			return "", false, err
		}

	case <-statusCh: // 容器完成
	}

	// 获取容器日志
	out, err := docker.ContainerLogs(context.Background(), containerHandle.ID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return "", false, err
	}
	defer out.Close()

	// 将日志输出转换为字符串
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, out); err != nil {
		return "", false, err
	}
	return buf.String(), false, nil
}

const REPORT_MESSAGE_FILE = "message.md"
const REPORT_RESULT_FILE = "result"

func readReport(
	logger *zap.Logger,
	reportPath string,
) (bool, string, error) {
	reportMessageFile := filepath.Join(reportPath, REPORT_MESSAGE_FILE)
	reportResultFile := filepath.Join(reportPath, REPORT_RESULT_FILE)
	logger.Debug("Reading report", zap.String("report_path", reportPath))
	reportMessage, err := os.ReadFile(reportMessageFile)
	if err != nil {
		logger.Error("Failed to read report message file", zap.Error(err))
		return false, "", err
	}
	reportResult, err := os.ReadFile(reportResultFile)
	if err != nil {
		logger.Error("Failed to read report result file", zap.Error(err))
		return false, "", err
	}
	if strings.HasPrefix(strings.ToLower(string(reportResult)), "t") {
		return true, string(reportMessage), nil
	}
	if strings.HasPrefix(strings.ToLower(string(reportResult)), "f") {
		return false, string(reportMessage), nil
	}
	logger.Warn("Unrecognized report result, supposing it failed", zap.String("result", string(reportResult)))
	return false, string(reportMessage), nil
}

func runTask(
	logger *zap.Logger,
	config *jConfig.JudgeConfig,
	db *gorm.DB,
	docker *client.Client,
	task *TestingTask,
) error {
	if err := handleTaskWaitingTimeout(logger, db, task, config.Testing.PendingQueueTimeoutInMinute); err != nil {
		logger.Error("Failed to handle task timeout", zap.Error(err))
		return err
	}

	if err := initializeTaskExecution(logger, db, task); err != nil {
		logger.Error("Failed to initialize task execution", zap.Error(err))
		return err
	}

	repositoryRecord := &schema.Repository{}
	if err := db.Where("repository_id = ?", task.RepositoryId).First(repositoryRecord).Error; err != nil {
		logger.Error("Failed to get repository record", zap.Error(err))
		return err
	}

	startpoint, err := getStartPoint(&task.Challenge, repositoryRecord.Startpoint)
	if err != nil {
		logger.Error("Failed to get startpoint", zap.Error(err))
		return err
	}

	runId := fmt.Sprintf("%s-%s-%s-%d-%d",
		task.RepositoryId,
		repositoryRecord.Provider,
		repositoryRecord.Subject,
		task.Serial,
		task.Stage)
	runId = fmt.Sprintf("%s-%d", runId, time.Now().UnixMilli())
	runId = strings.ToLower(runId)
	imageName := fmt.Sprintf("image-%s", runId)
	containerName := fmt.Sprintf("container-%s", runId)
	_, dockerfilePath, tempStoragePath, err := setupExecutionPaths(config, repositoryRecord, startpoint, runId)
	defer cleanUp(logger, docker, imageName, containerName, tempStoragePath)

	if err != nil {
		logger.Error("Failed to setup execution paths", zap.Error(err))
		return err
	}

	logger.Debug("Dockerfile path", zap.String("dockerfilePath", dockerfilePath))
	if err := buildDockerImage(docker, dockerfilePath, imageName); err != nil {
		logger.Error("Failed to build docker image", zap.Error(err))
		return err
	}

	reportMountPath := filepath.Join(tempStoragePath, "report")
	// convert to absolute path
	reportMountPath, err = filepath.Abs(reportMountPath)
	err = os.MkdirAll(reportMountPath, 0755)
	if err != nil {
		logger.Error("Failed to create report mount path", zap.Error(err))
		return err
	}
	logger.Debug("Report mount path", zap.String("reportMountPath", reportMountPath))
	log, stoppedBecauseTimeout, err := createAndStartContainer(
		docker,
		imageName,
		containerName,
		reportMountPath,
		time.Duration(config.Testing.RunningTimeoutInMinute)*time.Minute,
		STAGE_ENV_KEY,
		fmt.Sprintf("%d", task.Stage),
	)

	if err != nil {
		logger.Error("Failed to create and start container", zap.Error(err))
		task.TestingRecord.Status = StatusError
		task.TestingRecord.Log = log
		task.TestingRecord.RunEndTime = time.Now().Format(time.RFC3339)
		err = db.Save(&task.TestingRecord).Error
		if err != nil {
			logger.Error("Failed to save task record", zap.Error(err))
		}
		return err
	}

	if stoppedBecauseTimeout {
		task.TestingRecord.Status = StatusRunningTimeout
		task.TestingRecord.Log = log
		task.TestingRecord.RunEndTime = time.Now().Format(time.RFC3339)
		err = db.Save(&task.TestingRecord).Error
		if err != nil {
			logger.Error("Failed to save task record", zap.Error(err))
		}
		return err
	}

	task.TestingRecord.Status = StatusSuccess
	task.TestingRecord.Log = log
	task.TestingRecord.RunEndTime = time.Now().Format(time.RFC3339)
	err = db.Save(&task.TestingRecord).Error
	if err != nil {
		logger.Error("Failed to save task record", zap.Error(err))
	}

	pass, message, err := readReport(logger, reportMountPath)
	if err != nil {
		logger.Error("Failed to read report", zap.Error(err))
		return err
	}
	if !pass {
		task.TestingRecord.Status = StatusFailed
		task.TestingRecord.Log = log
		task.TestingRecord.Message = message
		task.TestingRecord.RunEndTime = time.Now().Format(time.RFC3339)
		err = db.Save(&task.TestingRecord).Error
		if err != nil {
			logger.Error("Failed to save task record", zap.Error(err))
		}
		return nil
	}
	task.TestingRecord.Status = StatusSuccess
	task.TestingRecord.Log = log
	task.TestingRecord.Message = message
	task.TestingRecord.RunEndTime = time.Now().Format(time.RFC3339)
	err = db.Save(&task.TestingRecord).Error
	if err != nil {
		logger.Error("Failed to save task record", zap.Error(err))
	}
	// advance the stage of repo by one
	if repositoryRecord.Stage <= int32(task.Stage)+1 {
		repositoryRecord.Stage = int32(task.Stage) + 1
		err = db.Save(&repositoryRecord).Error
		if err != nil {
			logger.Error("Failed to save repository record", zap.Error(err))
		}
	}
	err = db.Save(&repositoryRecord).Error
	if err != nil {
		logger.Error("Failed to save repository record", zap.Error(err))
		return err
	}
	return nil
}

func cleanUp(
	logger *zap.Logger,
	docker *client.Client,
	imageName string,
	containerName string,
	tempStoragePath string,
) error {
	containers, err := docker.ContainerList(context.Background(), container.ListOptions{
		All: true,
	})
	if err != nil {
		logger.Error("Failed to list containers", zap.Error(err))
		return err
	}

	var containerID string
	for _, c := range containers {
		if c.Names[0] == "/"+containerName { // Docker prepends a '/' to container names
			containerID = c.ID
			break
		}
	}
	if containerID == "" {
		logger.Warn("Container not found", zap.String("containerName", containerName))
	} else {
		if err := docker.ContainerRemove(context.Background(), containerID, container.RemoveOptions{Force: true}); err != nil {
			logger.Error("Failed to remove container", zap.Error(err))
			return err
		}
	}

	if _, err := docker.ImageRemove(context.Background(), imageName, image.RemoveOptions{Force: true}); err != nil {
		logger.Error("Failed to remove image", zap.Error(err))
		return err
	}

	if _, err := docker.ImagesPrune(context.Background(), filters.Args{}); err != nil {
		logger.Error("Failed to prune dangling images", zap.Error(err))
		return err
	}

	if err := os.RemoveAll(tempStoragePath); err != nil {
		logger.Error("Failed to remove temp storage path", zap.Error(err))
		return err
	}

	return nil
}
