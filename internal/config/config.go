package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Default  DefaultConfig  `yaml:"default"`
	Cache    CacheConfig    `yaml:"cache"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type DefaultConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type CacheConfig struct {
	TokenExpiration int `yaml:"token_expiration"`
	MaxTokens       int `yaml:"max_tokens"`
}

var GlobalConfig *Config

func Load(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	GlobalConfig = &cfg
	return nil
}
