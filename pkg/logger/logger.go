package logger

import (
	"log/slog"
	"os"
	"strings"
)

var defaultLogger *slog.Logger

func init() {
	// Default to INFO level
	SetLogLevel("info")
}

var LogLevelDebug = strings.ToLower(slog.LevelDebug.String())
var LogLevelInfo = strings.ToLower(slog.LevelInfo.String())
var LogLevelWarn = strings.ToLower(slog.LevelWarn.String())
var LogLevelError = strings.ToLower(slog.LevelError.String())

var LogLevel = []string{
	LogLevelDebug,
	LogLevelInfo,
	LogLevelWarn,
	LogLevelError,
}

// SetLogLevel configures the global logger with the specified level
func SetLogLevel(level string) {
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case LogLevelDebug:
		logLevel = slog.LevelDebug
	case LogLevelInfo:
		logLevel = slog.LevelInfo
	case LogLevelWarn:
		logLevel = slog.LevelWarn
	case LogLevelError:
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	})
	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

// Info logs an info message
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}
