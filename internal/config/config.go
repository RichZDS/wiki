package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName  string
	Env      string
	Port     string
	LogLevel string
	DB       DBConfig
	Redis    RedisConfig
	DateTime DateTimeConfig
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       string
}

type DateTimeConfig struct {
	TimeZone string
}

type envConfig struct {
	AppName  string
	Port     string
	LogLevel string
	DB       DBConfig
	Redis    RedisConfig
	DateTime DateTimeConfig
}

var envDefaults = map[string]envConfig{
	"dev": {
		AppName:  "aisearch",
		Port:     "8080",
		LogLevel: "debug",
		DB: DBConfig{
			Host:     "localhost",
			Port:     "3306",
			User:     "root",
			Password: "12345678",
			DBName:   "aisearch_dev",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     "6379",
			Password: "",
			DB:       "0",
		},
		DateTime: DateTimeConfig{
			TimeZone: "Asia/Shanghai",
		},
	},
	"prod": {
		AppName:  "aisearch",
		Port:     "8081",
		LogLevel: "info",
		DB: DBConfig{
			Host:     "prod-db.example.com",
			Port:     "3306",
			User:     "prod_user",
			Password: "",
			DBName:   "aisearch",
		},
		Redis: RedisConfig{
			Host:     "prod-redis.example.com",
			Port:     "6379",
			Password: "",
			DB:       "0",
		},
		DateTime: DateTimeConfig{
			TimeZone: "UTC",
		},
	},
}

func LoadEnvFile() {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	envFile := ".env." + env
	if err := godotenv.Load(envFile); err != nil {
		log.Printf("warning: %s not found, using environment variables", envFile)
	}
}

func Load() Config {
	env := getEnv("APP_ENV", "dev")
	defaults := envDefaults[env]

	return Config{
		AppName:  getEnv("APP_NAME", defaults.AppName),
		Env:     env,
		Port:    getEnv("APP_PORT", defaults.Port),
		LogLevel: getEnv("LOG_LEVEL", defaults.LogLevel),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", defaults.DB.Host),
			Port:     getEnv("DB_PORT", defaults.DB.Port),
			User:     getEnv("DB_USER", defaults.DB.User),
			Password: getEnv("DB_PASSWORD", defaults.DB.Password),
			DBName:   getEnv("DB_NAME", defaults.DB.DBName),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", defaults.Redis.Host),
			Port:     getEnv("REDIS_PORT", defaults.Redis.Port),
			Password: getEnv("REDIS_PASSWORD", defaults.Redis.Password),
			DB:       getEnv("REDIS_DB", defaults.Redis.DB),
		},
		DateTime: DateTimeConfig{
			TimeZone: getEnv("TIME_ZONE", defaults.DateTime.TimeZone),
		},
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
