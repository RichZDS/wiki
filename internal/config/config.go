package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Log      LogConfig      `yaml:"log"`
	MySQL    MySQLConfig    `yaml:"mysql"`
	Redis    RedisConfig    `yaml:"redis"`
	Env      string         `yaml:"-"`
}

type ServerConfig struct {
	Port    string `yaml:"port"`
	AppName string `yaml:"app_name"`
}

type LogConfig struct {
	Level string `yaml:"level"`
}

type MySQLConfig struct {
	Host            string `yaml:"host"`
	Port            string `yaml:"port"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	Database        string `yaml:"database"`
	Charset         string `yaml:"charset"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

func Load() Config {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	cfg := mustLoadYAML("config.yaml")

	if env == "prod" {
		override := mustLoadYAML("config.prod.yaml")
		cfg = merge(cfg, override)
	}

	cfg.Env = env

	// 敏感字段允许环境变量覆盖
	if pw := os.Getenv("MYSQL_PASSWORD"); pw != "" {
		cfg.MySQL.Password = pw
	}
	if pw := os.Getenv("REDIS_PASSWORD"); pw != "" {
		cfg.Redis.Password = pw
	}

	return cfg
}

func mustLoadYAML(path string) Config {
	data, err := os.ReadFile(path)
	if err != nil {
		panic("无法读取配置文件 " + path + ": " + err.Error())
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic("解析配置文件 " + path + " 失败: " + err.Error())
	}
	return cfg
}

func merge(base, override Config) Config {
	if override.Server.Port != "" {
		base.Server.Port = override.Server.Port
	}
	if override.Server.AppName != "" {
		base.Server.AppName = override.Server.AppName
	}
	if override.Log.Level != "" {
		base.Log.Level = override.Log.Level
	}

	if override.MySQL.Host != "" {
		base.MySQL.Host = override.MySQL.Host
	}
	if override.MySQL.Port != "" {
		base.MySQL.Port = override.MySQL.Port
	}
	if override.MySQL.User != "" {
		base.MySQL.User = override.MySQL.User
	}
	if override.MySQL.Password != "" {
		base.MySQL.Password = override.MySQL.Password
	}
	if override.MySQL.Database != "" {
		base.MySQL.Database = override.MySQL.Database
	}
	if override.MySQL.Charset != "" {
		base.MySQL.Charset = override.MySQL.Charset
	}
	if override.MySQL.MaxIdleConns != 0 {
		base.MySQL.MaxIdleConns = override.MySQL.MaxIdleConns
	}
	if override.MySQL.MaxOpenConns != 0 {
		base.MySQL.MaxOpenConns = override.MySQL.MaxOpenConns
	}
	if override.MySQL.ConnMaxLifetime != 0 {
		base.MySQL.ConnMaxLifetime = override.MySQL.ConnMaxLifetime
	}

	if override.Redis.Host != "" {
		base.Redis.Host = override.Redis.Host
	}
	if override.Redis.Port != "" {
		base.Redis.Port = override.Redis.Port
	}
	if override.Redis.Password != "" {
		base.Redis.Password = override.Redis.Password
	}
	if override.Redis.DB != 0 {
		base.Redis.DB = override.Redis.DB
	}

	return base
}
