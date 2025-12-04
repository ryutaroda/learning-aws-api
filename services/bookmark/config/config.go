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
		AppEnv:      os.Getenv("APP_ENV"),
	}
}
