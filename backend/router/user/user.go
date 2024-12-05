package user

import (
	"judge/jConfig"
	"judge/middleware"
	"judge/router"
	"judge/schema"
	"judge/shared"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func BuildUserInfoHandler(logger *zap.Logger, config *jConfig.JudgeConfig, db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userInfo := c.Locals(middleware.USER_INFO_LOCAL_KEY).(string)
		if userInfo == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(router.BuildError("No user info found"))
		}

		return c.JSON(router.BuildResponse(
			struct {
				UserInfo string `json:"userInfo"`
			}{
				UserInfo: userInfo,
			},
		))
	}
}

func BuildUserUpdateGitPasswordHandler(logger *zap.Logger, config *jConfig.JudgeConfig, db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		subject := c.Locals(middleware.SUBJECT_LOCAL_KEY).(string)
		provider := c.Locals(middleware.PROVIDER_LOCAL_KEY).(string)
		newPassword := c.Query("newPassword")
		if newPassword == "" {
			return c.Status(fiber.StatusBadRequest).JSON(router.BuildError("No new password found"))
		}

		// hash the new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(router.BuildError("Failed to hash password"))
		}

		var passwordRecord schema.UserBasicAuthentication
		err = db.Where("subject = ? AND provider = ?", subject, provider).First(&passwordRecord).Error
		if err == nil {
			// Update existing record
			passwordRecord.AuthenticationText = string(hashedPassword)
			db.Save(&passwordRecord)
		} else {
			// Create new record
			passwordRecord = schema.UserBasicAuthentication{
				Subject:            subject,
				Provider:           provider,
				AuthenticationText: string(hashedPassword),
			}
			db.Create(&passwordRecord)
		}
		return c.Status(fiber.StatusOK).JSON(router.BuildResponse(
			struct {
				Altered bool `json:"altered"`
			}{
				Altered: true,
			},
		))
	}
}

func BuildUserGitNameHandler(logger *zap.Logger, config *jConfig.JudgeConfig, db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		subject := c.Locals(middleware.SUBJECT_LOCAL_KEY).(string)
		provider := c.Locals(middleware.PROVIDER_LOCAL_KEY).(string)
		encoded := shared.EncodeUserGitName(logger, provider, subject)
		return c.JSON(router.BuildResponse(
			struct {
				GitName string `json:"gitName"`
			}{
				GitName: encoded,
			},
		))
	}
}

func SetupUserRouter(logger *zap.Logger, config *jConfig.JudgeConfig, db *gorm.DB, group *fiber.Router) error {
	(*group).Use(middleware.BuildAuthorizationMiddleWare(logger, config, db))
	(*group).Get("/info", BuildUserInfoHandler(logger, config, db))
	(*group).Post("/password", BuildUserUpdateGitPasswordHandler(logger, config, db))
	(*group).Get("/name", BuildUserGitNameHandler(logger, config, db))
	(*group).Get("/subject", func(c *fiber.Ctx) error {
		return c.JSON(router.BuildResponse(
			struct {
				Subject string `json:"subject"`
			}{
				Subject: c.Locals(middleware.SUBJECT_LOCAL_KEY).(string),
			},
		))
	})
	return nil
}
