package jConfig

import (
	"os"

	"github.com/BurntSushi/toml"
)

type LoggerConfig struct {
	Filename           string
	MaxSizeInMegabytes int
	MaxBackups         int
	MaxAgeInDays       int
	Compress           bool
	Level              string
}

type RepositoryStorageConfig struct {
	StorageFolder string
}

type DatabaseConfig struct {
	DbFile string
}

type ServerConfig struct {
	HostPort int
	HostAddr string
}

type ChallengeConfig struct {
	StorageFolder          string
	IgnorePatterns         []string
	MarkdownStyleSheetPath string
}

type AuthenticationServerConfig struct {
	ProviderName string
	ClientId     string
	ClientSecret string
	UserScopes   []string
	Enabled      bool
	AuthUrl      string
	TokenUrl     string
	UserInfoUrl  string
}

type AuthenticationConfig struct {
	AuthenticationServers         []AuthenticationServerConfig `toml:"server"`
	SingleUser                    bool
	AuthenticationTimeoutInSecond int
}

type TestingConfig struct {
	PendingQueueSize            int
	PendingQueueTimeoutInMinute int
	MaxConcurrentWorkers        int
	RunningTimeoutInMinute      int
	DockerSocket                string
	TmpStorageFolder            string
}

type JudgeConfig struct {
	Server            ServerConfig            `toml:"server"`
	Database          DatabaseConfig          `toml:"db"`
	Challenge         ChallengeConfig         `toml:"challenge"`
	RepositoryStorage RepositoryStorageConfig `toml:"repo"`
	Logger            LoggerConfig            `toml:"logger"`
	Authentication    AuthenticationConfig    `toml:"auth"`
	Testing           TestingConfig           `toml:"testing"`
}

func ParseJudgeConfig(path string) JudgeConfig {
	var config JudgeConfig
	if _, err := toml.DecodeFile(path, &config); err != nil {
		panic(err)
	}
	return config
}

func GetJudgeConfig() JudgeConfig {
	args := os.Args
	var configFile string
	if len(args) > 1 {
		configFile = args[1]
	} else {
		configFile = "config.toml"
	}
	config := ParseJudgeConfig(configFile)
	return config
}

func (c *JudgeConfig) GetAuthenticationServerConfigByProviderName(provider string) *AuthenticationServerConfig {
	for _, server := range c.Authentication.AuthenticationServers {
		if server.ProviderName == provider {
			return &server
		}
	}
	return nil
}
