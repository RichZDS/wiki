package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var Log *log.Logger

// dailyWriter 按日期自动滚动日志文件，跨天时自动切换到新文件。
type dailyWriter struct {
	mu   sync.Mutex
	dir  string
	file *os.File
	day  string
}

func (w *dailyWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	today := time.Now().Format("2006-01-02")
	if w.day != today {
		w.rotate(today)
	}
	return w.file.Write(p)
}

func (w *dailyWriter) rotate(today string) {
	if w.file != nil {
		w.file.Close()
	}
	os.MkdirAll(w.dir, 0755)
	path := filepath.Join(w.dir, today+".log")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	w.file = f
	w.day = today
}

// Init 初始化日志，同时写入 stdout 和按日期滚动的 log/YYYY-MM-DD.log 文件。
func Init(env string) {
	dw := &dailyWriter{dir: "log"}
	dw.rotate(time.Now().Format("2006-01-02"))

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
