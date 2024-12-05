package repository

import (
	"judge/jConfig"
	"judge/middleware"
	"judge/router"
	"judge/shared"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/sosedoff/gitkit"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func BuildGitServerHandler(
	logger *zap.Logger,
	config *jConfig.JudgeConfig,
	db *gorm.DB,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		provider, subject, challengeFolderName, repoId, suf, err := shared.ProcessGitServerRequestPath(
			logger,
			config,
			db,
			c,
		)
		logger.Debug(
			"Git server request",
			zap.String("provider", provider),
			zap.String("subject", subject),
			zap.String("challengeFolderName", challengeFolderName),
			zap.String("repoId", repoId),
			zap.String("path", suf),
		)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(router.BuildError(
				"Invalid path",
			))
		}
		if subject != c.Locals(middleware.SUBJECT_LOCAL_KEY).(string) {
			logger.Debug(
				"Subject mismatch",
				zap.String("expected", subject),
				zap.String("actual", c.Locals(middleware.SUBJECT_LOCAL_KEY).(string)),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(router.BuildError(
				"Subject mismatch",
			))
		}
		if provider != c.Locals(middleware.PROVIDER_LOCAL_KEY).(string) {
			return c.Status(fiber.StatusUnauthorized).JSON(router.BuildError(
				"Provider mismatch",
			))
		}

		repoRoot := filepath.Join(
			config.RepositoryStorage.StorageFolder,
			provider,
			subject,
			challengeFolderName,
		)

		wrapper := adaptor.HTTPHandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				r.URL.Path = filepath.Join(
					repoId,
					suf,
				)
				if !strings.HasPrefix(r.URL.Path, "/") {
					r.URL.Path = "/" + r.URL.Path
				}
				logger.Debug("Git server request", zap.String("path", r.URL.Path))
				gitkit.New(
					gitkit.Config{
						Dir:        repoRoot,
						Auth:       false,
						AutoCreate: true,
					},
				).ServeHTTP(w, r)
			},
		)

		return wrapper(c)
	}
}
