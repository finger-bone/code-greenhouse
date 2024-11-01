package middleware

import (
	"encoding/json"
	"errors"
	"io"
	"judge/jConfig"
	"judge/router"
	"judge/schema"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func getUserInfoUrl(logger *zap.Logger, config *jConfig.JudgeConfig, provider string) string {
	authenticationServerConfig := config.GetAuthenticationServerConfigByProviderName(provider)
	if authenticationServerConfig == nil {
		logger.Error("No authentication server config found for provider", zap.String("provider", provider))
		return ""
	}
	if authenticationServerConfig.Enabled && authenticationServerConfig.UserInfoUrl != "" {
		return authenticationServerConfig.UserInfoUrl
	}
	if authenticationServerConfig.Enabled && authenticationServerConfig.AuthUrl == "" {
		switch provider {
		case "github":
			return "https://api.github.com/user"
		case "google":
			return "https://www.googleapis.com/oauth2/v3/userinfo"
		}
	}
	logger.Debug("No user info url found for provider", zap.String("provider", provider))
	return ""
}

func getUserInfo(logger *zap.Logger, config *jConfig.JudgeConfig, userInfoUrl string, token string) (string, error) {
	req, err := http.NewRequest("GET", userInfoUrl, nil)
	if err != nil {
		logger.Error("Failed to create request", zap.Error(err))
		return "", err
	}

	req.Header.Set("Authorization", token)
	client := &http.Client{
		Timeout: time.Duration(config.Authentication.AuthenticationTimeoutInSecond) * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to send request", zap.Error(err))
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Failed to get user info", zap.Int("status_code", resp.StatusCode))
		return "", errors.New("failed to get user info, status code: " + resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response body", zap.Error(err))
		return "", err
	}

	return string(body), nil
}

func createUserIfNotExist(logger *zap.Logger, db *gorm.DB, subject string, provider string) error {
	var user schema.User
	db.Where("subject = ? AND provider = ?", subject, provider)
	if err := db.Where("subject = ? AND provider = ?", subject, provider).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user = schema.User{
				Subject:    subject,
				Provider:   provider,
				CreateTime: time.Now().Format(time.RFC3339),
				UpdateTime: time.Now().Format(time.RFC3339),
			}
			if err := db.Create(&user).Error; err != nil {
				logger.Error("Failed to create user", zap.Error(err))
				return err
			}
		} else {
			logger.Error("Failed to query user", zap.Error(err))
			return err
		}
	}
	return nil
}

func getSubject(logger *zap.Logger, userInfo string, provider string) string {
	var userInfoMap map[string]interface{}
	if err := json.Unmarshal([]byte(userInfo), &userInfoMap); err != nil {
		logger.Error("Failed to unmarshal userInfo", zap.Error(err))
		return ""
	}

	switch provider {
	case "github":
		return userInfoMap["login"].(string)
	case "google":
		return userInfoMap["sub"].(string)
	case "keycloak":
		return userInfoMap["sub"].(string)
	}

	logger.Warn("The provider is not recorded, automatically finding subject", zap.String("provider", provider))
	possibleKeys := []string{"sub", "login", "email", "openid", "id", "user", "name", "username"}
	for _, key := range possibleKeys {
		if value, ok := userInfoMap[key]; ok {
			return value.(string)
		}
	}

	logger.Error("Failed to get subject from an unrecorded provider. Will use the whole userInfo as subject.", zap.String("provider", provider))
	return userInfo
}

const USER_INFO_LOCAL_KEY = "userInfo"
const SUBJECT_LOCAL_KEY = "subject"
const PROVIDER_LOCAL_KEY = "provider"
const SINGLE_USER_PROVIDER = "localhost"
const SINGLE_USER_SUBJECT = "subject"

func BuildAuthorizationMiddleWare(logger *zap.Logger, config *jConfig.JudgeConfig, db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if config.Authentication.SingleUser {
			c.Locals(SUBJECT_LOCAL_KEY, SINGLE_USER_SUBJECT)
			c.Locals(PROVIDER_LOCAL_KEY, SINGLE_USER_PROVIDER)
			c.Locals(USER_INFO_LOCAL_KEY, "{}")
			err := createUserIfNotExist(logger, db, SINGLE_USER_SUBJECT, SINGLE_USER_PROVIDER)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(router.BuildError("Failed to create user"))
			}
			return c.Next()
		}

		provider := c.Get("Provider")

		if provider == "" {
			provider = c.Query("provider")
		}

		c.Locals(PROVIDER_LOCAL_KEY, provider)
		if provider == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(router.BuildError("No provider found in header"))
		}

		token := c.Get("Authorization")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(router.BuildError("No token found in header"))
		}
		userInfoUrl := getUserInfoUrl(logger, config, provider)
		if userInfoUrl == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(router.BuildError("No user info url found"))
		}
		userInfo, err := getUserInfo(logger, config, userInfoUrl, token)
		if err != nil {
			logger.Error("Failed to get user info", zap.Error(err))
			return c.Status(fiber.StatusUnauthorized).JSON(router.BuildError("Failed to get user info"))
		}

		c.Locals(USER_INFO_LOCAL_KEY, userInfo)

		subject := getSubject(logger, userInfo, provider)
		if subject == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(router.BuildError("Failed to get subject"))
		}
		c.Locals(SUBJECT_LOCAL_KEY, subject)

		err = createUserIfNotExist(logger, db, subject, provider)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(router.BuildError("Failed to create user"))
		}

		return c.Next()
	}
}
