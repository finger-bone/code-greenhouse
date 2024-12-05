package bootstrap

import (
	"judge/jConfig"
	"judge/middleware"
	"time"

	"github.com/docker/docker/client"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func bootstrapServer(
	logger *zap.Logger,
	config *jConfig.JudgeConfig,
	db *gorm.DB,
	docker *client.Client,
) *fiber.App {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		stop := time.Now()
		logger.Debug("Request",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("latency", stop.Sub(start)),
		)
		return err
	})

	bootstrapHandler(logger, config, db, docker, app)

	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	app.Get("/ping/casual", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	app.Get("/ping/auth", middleware.BuildAuthorizationMiddleWare(logger, config, db), func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	logger.Info("Server Starting")

	return app
}
