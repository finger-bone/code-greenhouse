package shared

import (
	"judge/jConfig"
	"judge/router"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func ProcessGitServerRequestPath(
	logger *zap.Logger,
	config *jConfig.JudgeConfig,
	db *gorm.DB,
	c *fiber.Ctx,
) (string, string, string, string, string, error) {
	path := string(c.Context().Request.URI().Path())
	if strings.HasPrefix(path, "/api") {
		path = strings.TrimPrefix(path, "/api")
	}
	path = strings.TrimPrefix(path, router.GIT_SERVER_PREFIX)
	path = strings.TrimPrefix(path, "/")
	splittedPath := strings.Split(path, "/")
	if len(splittedPath) <= 3 {
		return "", "", "", "", "", &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: "Invalid path",
		}
	}
	provider := splittedPath[0]
	subject := splittedPath[1]
	challengeFolderName := splittedPath[2]
	repoId := splittedPath[3]
	prefix := strings.Join(splittedPath[:4], "/")

	return provider, subject, challengeFolderName, repoId, path[len(prefix):], nil
}
