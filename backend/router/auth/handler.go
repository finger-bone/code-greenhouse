package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"judge/jConfig"
	"judge/router"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type AccessTokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
	// RedirectURI  string `json:"redirect_uri"`
}

// 响应体定义
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error,omitempty"`
	ErrorDesc   string `json:"error_description,omitempty"`
}

// 获取访问令牌函数
func requestAccessToken(tokenUrl string, reqBody AccessTokenRequest) (*AccessTokenResponse, error) {
	// 将请求体编码为 JSON
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request body: %w", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", tokenUrl, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析响应体
	var tokenResp AccessTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	// 检查是否包含错误
	if tokenResp.Error != "" {
		return nil, errors.New(tokenResp.ErrorDesc)
	}

	return &tokenResp, nil
}

func SetupAuthRouter(logger *zap.Logger, config *jConfig.JudgeConfig, group *fiber.Router) {
	(*group).Get("/providers", func(c *fiber.Ctx) error {
		providers := make([]string, 0)
		for _, server := range config.Authentication.AuthenticationServers {
			if server.Enabled {
				providers = append(providers, server.ProviderName)
			}
		}
		return c.JSON(router.BuildResponse(
			struct {
				Providers []string `json:"providers"`
			}{
				Providers: providers,
			},
		))
	})
	(*group).Get("/single-user", func(c *fiber.Ctx) error {
		return c.JSON(router.BuildResponse(
			struct {
				SingleUser bool `json:"enabled"`
			}{
				SingleUser: config.Authentication.SingleUser,
			},
		))
	})
	(*group).Get("/token", func(c *fiber.Ctx) error {
		providerName := c.Query("provider")
		if providerName == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No provider found in query"})
		}

		authServerConfig := config.GetAuthenticationServerConfigByProviderName(providerName)
		if authServerConfig == nil || !authServerConfig.Enabled {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid or disabled auth provider"})
		}

		tokenUrl := authServerConfig.TokenUrl
		if tokenUrl == "" && providerName == "github" {
			tokenUrl = "https://github.com/login/oauth/access_token"
		}

		authorizationCode := c.Query("code")
		if authorizationCode == "" {
			return c.Status(fiber.StatusBadRequest).JSON(
				router.BuildError("No authorization code found in query"),
			)
		}

		// 创建访问令牌请求体
		reqBody := AccessTokenRequest{
			ClientID:     authServerConfig.ClientId,
			ClientSecret: authServerConfig.ClientSecret,
			Code:         authorizationCode,
			// RedirectURI:  c.Query("redirect_url"),
		}

		// 代理请求以获取访问令牌
		tokenResp, err := requestAccessToken(tokenUrl, reqBody)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		// 返回令牌响应
		return c.JSON(
			router.BuildResponse(
				struct {
					AccessToken string `json:"accessToken"`
				}{
					AccessToken: tokenResp.AccessToken,
				},
			),
		)
	})
	(*group).Get("/url", func(c *fiber.Ctx) error {
		if config.Authentication.SingleUser {
			return c.JSON(router.BuildError("Single user mode is enabled"))
		}
		provider := c.Query("provider")
		if provider == "" {
			return c.Status(fiber.StatusBadRequest).JSON(router.BuildError("No provider found in query"))
		}
		authServerConfig := config.GetAuthenticationServerConfigByProviderName(provider)
		if authServerConfig == nil {
			return c.Status(fiber.StatusBadRequest).JSON(router.BuildError("No auth server config found"))
		}
		if !authServerConfig.Enabled {
			return c.Status(fiber.StatusBadRequest).JSON(router.BuildError("Auth provider is not enabled"))
		}
		redirect_url := c.Query("redirect_url")
		oauth2Config := &oauth2.Config{
			ClientID:     authServerConfig.ClientId,
			ClientSecret: authServerConfig.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  authServerConfig.AuthUrl,
				TokenURL: authServerConfig.TokenUrl,
			},
			RedirectURL: redirect_url,
			Scopes:      authServerConfig.UserScopes,
		}
		if provider == "github" {
			oauth2Config.Endpoint = github.Endpoint
		}
		if provider == "google" {
			oauth2Config.Endpoint = google.Endpoint
		}
		state := c.Query("state")
		authURL := oauth2Config.AuthCodeURL(state)
		return c.JSON(router.BuildResponse(
			struct {
				Url string `json:"url"`
			}{
				Url: authURL,
			},
		))
	})
}
