package config

import "os"

type Config struct {
	DatabaseURL string
	SQSQueueURL string
	AppEnv      string
}

func Load() *Config {
	return &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		SQSQueueURL: os.Getenv("SQS_QUEUE_URL"),
		AppEnv:      getEnvWithDefault("APP_ENV", "development"),
	}
}

// getEnvWithDefault 環境変数を取得（デフォルト値あり）
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
