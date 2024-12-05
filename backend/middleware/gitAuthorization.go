package middleware

import (
	"encoding/base64"
	"judge/jConfig"
	"judge/router"
	"judge/schema"
	"judge/shared"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func BuildGitAuthorizationMiddleWare(logger *zap.Logger, config *jConfig.JudgeConfig, db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if config.Authentication.SingleUser {
			c.Locals(SUBJECT_LOCAL_KEY, "subject")
			c.Locals(PROVIDER_LOCAL_KEY, "localhost")
			return c.Next()
		}
		token := c.Get("Authorization")
		if token == "" {
			// 当没有 Authorization 头时，返回 401 并要求 Basic Auth
			c.Set("WWW-Authenticate", `Basic realm="Git Server"`)
			c.SendStatus(fiber.StatusUnauthorized)
			return c.SendString("Authorization required")
		}
		provider, subject, _, _, _, err := shared.ProcessGitServerRequestPath(
			logger,
			config,
			db,
			c,
		)
		if err != nil {
			logger.Error("Failed to process git server request path", zap.Error(err))
			return c.Status(fiber.StatusUnauthorized).JSON(router.BuildError("Failed to process git server request path"))
		}

		if strings.HasPrefix(token, "Basic ") {
			token = strings.TrimPrefix(token, "Basic ")
		}
		// decode the basic token
		decoded, err := base64.StdEncoding.DecodeString(token)
		username, password, ok := strings.Cut(string(decoded), ":")
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(router.BuildError("Failed to decode basic token"))
		}
		decodedProvider, decodedSubject := shared.DecodeUserGitName(logger, username)
		logger.Debug(
			"Decoded username and provider",
			zap.String("decodedUsername", decodedSubject),
			zap.String("decodedProvider", decodedProvider),
			zap.String("provider", provider),
		)
		if decodedSubject != subject || decodedProvider != provider {
			return c.Status(fiber.StatusUnauthorized).JSON(router.BuildError("Mismatched username or provider"))
		}
		var passwordRecord schema.UserBasicAuthentication
		err = db.Where("subject = ? AND provider = ?", subject, provider).First(&passwordRecord).Error
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(router.BuildError("Failed to find password record"))
		}
		// verify password
		err = bcrypt.CompareHashAndPassword([]byte(passwordRecord.AuthenticationText), []byte(password))
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(router.BuildError("Failed to verify password"))
		}

		c.Locals(SUBJECT_LOCAL_KEY, subject)
		c.Locals(PROVIDER_LOCAL_KEY, provider)
		logger.Debug("Git authorization middleware passed", zap.String("subject", subject), zap.String("provider", provider))
		return c.Next()
	}
}
