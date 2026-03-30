package config

import (
	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Claude   ClaudeConfig
	Security SecurityConfig
}

type ServerConfig struct {
	Port string
	Mode string // debug, release, test
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type ClaudeConfig struct {
	APIKey  string
	Model   string
	Timeout int // seconds
}

type SecurityConfig struct {
	AESKey string // 32-byte hex-encoded key for AES-256-GCM
}

// Load reads configuration from environment and config files.
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.AutomaticEnv()

	// Defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "crayfish_user")
	viper.SetDefault("database.password", "crayfish_dev_password")
	viper.SetDefault("database.dbname", "crayfish_travel")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("redis.addr", "localhost:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("claude.model", "claude-sonnet-4-20250514")
	viper.SetDefault("claude.timeout", 10)
	// Dev-only default AES key (32 bytes hex = 64 hex chars)
	viper.SetDefault("security.aeskey", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")

	_ = viper.ReadInConfig() // ok if config file not found

	cfg := &Config{
		Server: ServerConfig{
			Port: viper.GetString("server.port"),
			Mode: viper.GetString("server.mode"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("database.host"),
			Port:     viper.GetInt("database.port"),
			User:     viper.GetString("database.user"),
			Password: viper.GetString("database.password"),
			DBName:   viper.GetString("database.dbname"),
			SSLMode:  viper.GetString("database.sslmode"),
		},
		Redis: RedisConfig{
			Addr:     viper.GetString("redis.addr"),
			Password: viper.GetString("redis.password"),
			DB:       viper.GetInt("redis.db"),
		},
		Claude: ClaudeConfig{
			APIKey:  viper.GetString("claude.apikey"),
			Model:   viper.GetString("claude.model"),
			Timeout: viper.GetInt("claude.timeout"),
		},
		Security: SecurityConfig{
			AESKey: viper.GetString("security.aeskey"),
		},
	}

	return cfg, nil
}
