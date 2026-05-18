package config

import (
	"fmt"
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
	Host            string // 数据库地址
	Port            string // 数据库端口
	User            string // 数据库用户名
	Password        string // 数据库密码
	DBName          string // 数据库名称
	MaxIdleConns    int    // 最大空闲连接数
	MaxOpenConns    int    // 最大打开连接数
	ConnMaxLifetime int    // 连接最大生命周期（秒）
	Charset         string // 字符集
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
			Host:            "localhost",
			Port:            "3306",
			User:            "root",
			Password:        "12345678",
			DBName:          "aisearch_dev",
			MaxIdleConns:    10,
			MaxOpenConns:    100,
			ConnMaxLifetime: 3600,
			Charset:         "utf8mb4",
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
			Host:            "prod-db.example.com",
			Port:            "3306",
			User:            "prod_user",
			Password:        "",
			DBName:          "aisearch",
			MaxIdleConns:    20,
			MaxOpenConns:    200,
			ConnMaxLifetime: 7200,
			Charset:         "utf8mb4",
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
		Env:      env,
		Port:     getEnv("APP_PORT", defaults.Port),
		LogLevel: getEnv("LOG_LEVEL", defaults.LogLevel),
		DB: DBConfig{
			Host:            getEnv("DB_HOST", defaults.DB.Host),
			Port:            getEnv("DB_PORT", defaults.DB.Port),
			User:            getEnv("DB_USER", defaults.DB.User),
			Password:        getEnv("DB_PASSWORD", defaults.DB.Password),
			DBName:          getEnv("DB_NAME", defaults.DB.DBName),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", defaults.DB.MaxIdleConns),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", defaults.DB.MaxOpenConns),
			ConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", defaults.DB.ConnMaxLifetime),
			Charset:         getEnv("DB_CHARSET", defaults.DB.Charset),
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

func getEnvAsInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	var result int
	fmt.Sscanf(value, "%d", &result)
	return result
}

