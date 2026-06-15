package model

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type SnowflakeGenerator struct {
	Mu       sync.Mutex
	WorkerID int64
	Sequence int64
	LastTs   int64
}

type DailyWriter struct {
	Mu   sync.Mutex
	Dir  string
	File *os.File
	Day  string
}

// Write 按日期切换日志文件并写入日志内容。
func (w *DailyWriter) Write(data []byte) (int, error) {
	w.Mu.Lock()
	defer w.Mu.Unlock()

	today := time.Now().Format("2006-01-02")
	if w.Day != today {
		w.Rotate(today)
	}
	if w.File == nil {
		return 0, io.ErrClosedPipe
	}
	return w.File.Write(data)
}

// Rotate 将日志输出切换到指定日期的文件。
func (w *DailyWriter) Rotate(today string) {
	if w.File != nil {
		_ = w.File.Close()
	}
	_ = os.MkdirAll(w.Dir, 0755)
	path := filepath.Join(w.Dir, today+".log")
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		w.File = nil
		return
	}
	w.File = file
	w.Day = today
}

type GormLoggerWriter struct {
	Logger *log.Logger
}

// Printf 将 GORM 日志转发到项目日志实例。
func (w *GormLoggerWriter) Printf(format string, args ...any) {
	w.Logger.Printf(format, args...)
}
