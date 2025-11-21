package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Services ServicesConfig
}

type AppConfig struct {
	Name string
	Env  string
	Port int
}

type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	Database     string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	PoolSize int
}

type JWTConfig struct {
	Secret            string
	ExpireHours       int
	RefreshExpireDays int
}

type ServicesConfig struct {
	UserServiceURL    string
	ChatServiceURL    string
	RelayServiceURL   string
	BillingServiceURL string
}

func Load() (*Config, error) {
	// 尝试加载 .env 文件
	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "Oblivious"),
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnvAsInt("APP_PORT", 8080),
		},
		Database: DatabaseConfig{
			Host:         getEnv("DATABASE_HOST", "localhost"),
			Port:         getEnvAsInt("DATABASE_PORT", 5432),
			User:         getEnv("DATABASE_USER", "postgres"),
			Password:     getEnv("DATABASE_PASSWORD", "password"),
			Database:     getEnv("DATABASE_NAME", "oblivious"),
			SSLMode:      getEnv("DATABASE_SSLMODE", "disable"),
			MaxOpenConns: getEnvAsInt("DATABASE_MAX_OPEN_CONNS", 100),
			MaxIdleConns: getEnvAsInt("DATABASE_MAX_IDLE_CONNS", 10),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			PoolSize: getEnvAsInt("REDIS_POOL_SIZE", 10),
		},
		JWT: JWTConfig{
			Secret:            getEnv("JWT_SECRET", "your-secret-key"),
			ExpireHours:       getEnvAsInt("JWT_EXPIRE_HOURS", 2),
			RefreshExpireDays: getEnvAsInt("REFRESH_TOKEN_EXPIRE_DAYS", 7),
		},
		Services: ServicesConfig{
			UserServiceURL:    getEnv("USER_SERVICE_URL", "http://localhost:8081"),
			ChatServiceURL:    getEnv("CHAT_SERVICE_URL", "http://localhost:8082"),
			RelayServiceURL:   getEnv("RELAY_SERVICE_URL", "http://localhost:8083"),
			BillingServiceURL: getEnv("BILLING_SERVICE_URL", "http://localhost:8088"),
		},
	}

	// 验证必要配置
	if cfg.JWT.Secret == "your-secret-key" && cfg.App.Env == "production" {
		return nil, fmt.Errorf("JWT_SECRET must be set in production")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}


