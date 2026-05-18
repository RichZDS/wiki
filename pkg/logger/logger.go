// Package logger 提供日志记录功能，日志按日期分割存储在 log 目录下。
package logger

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

// Log 是全局日志实例，通过 GetLogger 获取。
var Log *log.Logger

// Init 初始化日志系统。
// 根据当前日期创建日志文件（log/YYYY-MM-DD.log），日志同时写入文件和控制台。
// env 参数目前未使用，保留用于未来根据环境调整日志级别。
func Init(env string) {
	logDir := "log"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("failed to create log directory: %v", err)
	}

	filename := time.Now().Format("2006-01-02") + ".log"
	logPath := filepath.Join(logDir, filename)

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}

	Log = log.New(file, "", log.LstdFlags)
	Log.SetOutput(os.Stdout)
}

// GetLogger 返回全局日志实例。
// 如果日志未初始化，先调用 Init("dev") 进行初始化。
func GetLogger() *log.Logger {
	if Log == nil {
		Init("dev")
	}
	return Log
}
