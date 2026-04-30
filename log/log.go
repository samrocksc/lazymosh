package log

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// Level represents logging verbosity.
type Level int

const (
	LevelError Level = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

func (l Level) String() string {
	switch l {
	case LevelError:
		return "ERROR"
	case LevelWarn:
		return "WARN"
	case LevelInfo:
		return "INFO"
	case LevelDebug:
		return "DEBUG"
	}
	return "UNKNOWN"
}

// Logger is a minimal structured logger with verbosity support.
type Logger struct {
	mu  sync.Mutex
	lvl Level
}

var defaultLogger = &Logger{lvl: LevelInfo}

// SetLevel sets the global log level.
func SetLevel(lvl Level) {
	defaultLogger.mu.Lock()
	defaultLogger.lvl = lvl
	defaultLogger.mu.Unlock()
}

// SetLevelFromString parses "error","warn","info","debug" and sets the level.
func SetLevelFromString(s string) bool {
	switch s {
	case "error":
		SetLevel(LevelError)
	case "warn":
		SetLevel(LevelWarn)
	case "info":
		SetLevel(LevelInfo)
	case "debug":
		SetLevel(LevelDebug)
	default:
		return false
	}
	return true
}

// Debug logs a debug message.
func Debug(format string, args ...any) {
	defaultLogger.log(LevelDebug, format, args...)
}

// Info logs an info message.
func Info(format string, args ...any) {
	defaultLogger.log(LevelInfo, format, args...)
}

// Warn logs a warning message.
func Warn(format string, args ...any) {
	defaultLogger.log(LevelWarn, format, args...)
}

// Error logs an error message.
func Error(format string, args ...any) {
	defaultLogger.log(LevelError, format, args...)
}

// Fatal logs an error and exits.
func Fatal(format string, args ...any) {
	defaultLogger.log(LevelError, format, args...)
	os.Exit(1)
}

func (l *Logger) log(lvl Level, format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if lvl > l.lvl {
		return
	}
	prefix := fmt.Sprintf("%s %-5s ", time.Now().Format("15:04:05"), lvl.String())
	fmt.Fprintf(os.Stderr, prefix+format+"\n", args...)
}
