package logger

import (
	"io"
	"log"
	"os"
	"time"

	"aisearch/internal/model"
)

var Log *log.Logger

// Init 初始化日志，同时写入 stdout 和按日期滚动的 log/YYYY-MM-DD.log 文件。
func Init(env string) {
	dw := &model.DailyWriter{Dir: "log"}
	dw.Rotate(time.Now().Format("2006-01-02"))

	prefix := ""
	if env == "dev" {
		prefix = "[DEV] "
	}
	Log = log.New(io.MultiWriter(os.Stdout, dw), prefix, log.LstdFlags)
}

// GetLogger 返回全局 logger，未初始化时使用 dev 作为默认值。
func GetLogger() *log.Logger {
	if Log == nil {
		Init("dev")
	}
	return Log
}
