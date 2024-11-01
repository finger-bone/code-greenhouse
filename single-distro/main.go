package main

import (
	"fmt"
	"judge/bootstrap"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app, logger, config := bootstrap.BuildApp()
	// reroute /api to app
	proxyApp := fiber.New()
	proxyApp.Mount("/api", app)
	proxyApp.All("/*", func(c *fiber.Ctx) error {
		path := c.Params("*")
		if path == "" {
			path = "index.html"
		}
		return c.SendFile(fmt.Sprintf("frontend/dist/%s", path))
	})
	defer logger.Sync()
	logger.Info("Judge Started")
	proxyApp.Listen(fmt.Sprintf("%s:%d", config.Server.HostAddr, config.Server.HostPort))
}
