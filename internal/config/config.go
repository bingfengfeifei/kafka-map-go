package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server            ServerConfig             `yaml:"server"`
	Database          DatabaseConfig           `yaml:"database"`
	Default           DefaultConfig            `yaml:"default"`
	Cache             CacheConfig              `yaml:"cache"`
	BootstrapClusters []BootstrapClusterConfig `yaml:"bootstrap_clusters"`
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

type BootstrapClusterConfig struct {
	Name             string `yaml:"name"`
	Servers          string `yaml:"servers"`
	SecurityProtocol string `yaml:"security_protocol"`
	SaslMechanism    string `yaml:"sasl_mechanism"`
	SaslUsername     string `yaml:"sasl_username"`
	SaslPassword     string `yaml:"sasl_password"`
	AuthUsername     string `yaml:"auth_username"`
	AuthPassword     string `yaml:"auth_password"`
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

	if err := applyEnvOverrides(&cfg); err != nil {
		return err
	}

	GlobalConfig = &cfg
	return nil
}

func applyEnvOverrides(cfg *Config) error {
	if v := strings.TrimSpace(os.Getenv("KAFKA_MAP_SERVER_PORT")); v != "" {
		port, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("invalid KAFKA_MAP_SERVER_PORT: %w", err)
		}
		cfg.Server.Port = port
	}

	if v := strings.TrimSpace(os.Getenv("KAFKA_MAP_DATABASE_PATH")); v != "" {
		cfg.Database.Path = v
	}

	if v := firstEnv("DEFAULT_USERNAME", "KAFKA_MAP_DEFAULT_USERNAME"); v != "" {
		cfg.Default.Username = v
	}

	if v := firstEnv("DEFAULT_PASSWORD", "KAFKA_MAP_DEFAULT_PASSWORD"); v != "" {
		cfg.Default.Password = v
	}

	if v := strings.TrimSpace(os.Getenv("KAFKA_MAP_CACHE_TOKEN_EXPIRATION")); v != "" {
		exp, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("invalid KAFKA_MAP_CACHE_TOKEN_EXPIRATION: %w", err)
		}
		cfg.Cache.TokenExpiration = exp
	}

	if v := strings.TrimSpace(os.Getenv("KAFKA_MAP_CACHE_MAX_TOKENS")); v != "" {
		maxTokens, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("invalid KAFKA_MAP_CACHE_MAX_TOKENS: %w", err)
		}
		cfg.Cache.MaxTokens = maxTokens
	}

	bootstrap := BootstrapClusterConfig{
		Name:             firstEnv("DEFAULT_CLUSTER_NAME", "KAFKA_MAP_BOOTSTRAP_NAME"),
		Servers:          firstEnv("DEFAULT_CLUSTER_SERVERS", "KAFKA_MAP_BOOTSTRAP_SERVERS"),
		SecurityProtocol: firstEnv("DEFAULT_CLUSTER_SECURITY_PROTOCOL", "KAFKA_MAP_BOOTSTRAP_SECURITY_PROTOCOL"),
		SaslMechanism:    firstEnv("DEFAULT_CLUSTER_SASL_MECHANISM", "KAFKA_MAP_BOOTSTRAP_SASL_MECHANISM"),
		SaslUsername:     firstEnv("DEFAULT_CLUSTER_SASL_USERNAME", "KAFKA_MAP_BOOTSTRAP_SASL_USERNAME"),
		SaslPassword:     firstEnv("DEFAULT_CLUSTER_SASL_PASSWORD", "KAFKA_MAP_BOOTSTRAP_SASL_PASSWORD"),
		AuthUsername:     firstEnv("DEFAULT_CLUSTER_AUTH_USERNAME", "KAFKA_MAP_BOOTSTRAP_AUTH_USERNAME"),
		AuthPassword:     firstEnv("DEFAULT_CLUSTER_AUTH_PASSWORD", "KAFKA_MAP_BOOTSTRAP_AUTH_PASSWORD"),
	}
	if strings.TrimSpace(bootstrap.Name) != "" && strings.TrimSpace(bootstrap.Servers) != "" {
		cfg.BootstrapClusters = append(cfg.BootstrapClusters, bootstrap)
	}

	return nil
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return ""
}
