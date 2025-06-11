package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"middleman/pkg/config"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

var levelStrings = map[Level]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
	LevelFatal: "FATAL",
}

var (
	GlobalLogger *Logger
	once         sync.Once
)

type Logger struct {
	mu         sync.Mutex
	fileWriter io.WriteCloser
	console    io.Writer
	level      Level
}

func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelStr := levelStrings[level]
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] [%s] %s\n", timestamp, levelStr, message)

	if l.fileWriter != nil {
		_, _ = l.fileWriter.Write([]byte(logLine))
	}

	if level >= LevelError {
		logLine = fmt.Sprintf("\033[31m%s\033[0m", logLine)
	}
	_, _ = l.console.Write([]byte(logLine))
}

func New(logFile string, level Level) (*Logger, error) {
	if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
		return nil, fmt.Errorf("无法创建日志目录: %w", err)
	}

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("无法打开日志文件: %w", err)
	}

	return &Logger{
		fileWriter: file,
		console:    os.Stdout,
		level:      level,
	}, nil
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.fileWriter != nil {
		return l.fileWriter.Close()
	}
	return nil
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// Debug 输出调试级别的日志
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Info 输出信息级别的日志
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Warn 输出警告级别的日志
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

// Error 输出错误级别的日志
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// Fatal 输出致命错误级别的日志并退出程序
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(LevelFatal, format, args...)
	os.Exit(1)
}

func init() {
	once.Do(func() {
		var level Level
		conf := config.GetConf()
		switch conf.LogLevel {
		case "debug":
			level = 0
		case "info":
			level = 1
		case "warning":
			level = 2
		case "error":
			level = 3
		default:
			level = 4
		}
		logger, err := New("logs/middleman.log", level)
		if err != nil {
			fmt.Printf("无法创建日志记录器: %v\n", err)
			return
		}
		GlobalLogger = logger
	})
}

func GetLogger() *Logger {
	return GlobalLogger
}
