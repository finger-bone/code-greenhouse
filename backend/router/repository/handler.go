package repository

import (
	"judge/jConfig"
	"judge/middleware"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func SetupRepositoryRouter(logger *zap.Logger, config *jConfig.JudgeConfig, db *gorm.DB, group *fiber.Router) {
	(*group).Post(
		"/project",
		middleware.BuildAuthorizationMiddleWare(logger, config, db),
		BuildNewRepositoryHandler(logger, config, db),
	)
	(*group).All(
		"/git/*",
		middleware.BuildGitAuthorizationMiddleWare(logger, config, db),
		BuildGitServerHandler(logger, config, db),
	)
}
