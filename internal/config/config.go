package config

import (
	"os"
	"path/filepath"

	"aisearch/internal/model"

	"gopkg.in/yaml.v3"
)

type Config = model.Config
type ServerConfig = model.ServerConfig
type LogConfig = model.LogConfig
type MySQLConfig = model.MySQLConfig
type RedisConfig = model.RedisConfig

// resolveConfigPath 从环境变量或上级目录的 manifest/config 中查找配置文件。
func resolveConfigPath(filename string) string {
	if p := os.Getenv("APP_CONFIG"); p != "" {
		return p
	}

	defaultPath := filepath.Join("manifest", "config", filename)

	dir, err := os.Getwd()
	if err != nil {
		return defaultPath
	}

	for {
		candidate := filepath.Join(dir, defaultPath)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return defaultPath
}

// Load 加载当前运行环境的完整配置。
func Load() Config {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	cfg := mustLoadYAML(resolveConfigPath("config.yaml"))

	if env == "prod" {
		override := mustLoadYAML(resolveConfigPath("config.prod.yaml"))
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

// mustLoadYAML 读取并解析 YAML 配置，失败时终止初始化。
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

// merge 将非零的环境配置覆盖到基础配置。
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
