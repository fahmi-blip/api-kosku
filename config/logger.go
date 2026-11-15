package config

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	writer io.Writer
	level  string
	mu     sync.Mutex
}

func NewLogger() *Logger {
	if err := os.MkdirAll("logs", 0755); err != nil {
		panic("Gagal membuat folder logs: " + err.Error())
	}

	rotator := &lumberjack.Logger{
		Filename:   filepath.Join("logs", "app.log"),
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     14,
		Compress:   true,
	}

	writer := io.MultiWriter(os.Stdout, rotator)
	log.SetOutput(writer)

	return &Logger{
		writer: writer,
		level:  strings.ToLower(GetEnv("LOG_LEVEL", "info")),
	}
}

func (l *Logger) Writer() io.Writer {
	return l.writer
}

func (l *Logger) logf(level, msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now().Format("2006-01-02T15:04:05.000Z07:00")
	var formatted string
	if len(args) > 0 {
		formatted = fmt.Sprintf(msg, args...)
	} else {
		formatted = msg
	}
	fmt.Fprintf(l.writer, "[%s] %-5s %s\n", now, level, formatted)
}

func (l *Logger) Info(msg string, args ...any) {
	l.logf("INFO", msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.logf("WARN", msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.logf("ERROR", msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	if l.level == "debug" {
		l.logf("DEBUG", msg, args...)
	}
}
