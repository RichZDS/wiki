package config

import "os"

type Config struct {
	AppName string
	Env     string
	Port    string
}

func Load() Config {
	return Config{
		AppName: getEnv("APP_NAME", "aisearch"),
		Env:     getEnv("APP_ENV", "dev"),
		Port:    getEnv("APP_PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
