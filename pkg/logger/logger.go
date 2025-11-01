package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

func GetLogLevel(level string) (slog.Level, error) {
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
		return slog.LevelInfo, fmt.Errorf("invalid log level: %s", level)
	}
	return logLevel, nil
}

// SetLogLevel configures the global logger with the specified level
func SetLogLevel(level string) {
	logLevel, _ := GetLogLevel(level)
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

func formatValuesForUsageDocs(values []string) string {
	return fmt.Sprintf("{%s}", strings.Join(values, "|"))
}

type logLevelValue struct {
	string *string
}

func (e *logLevelValue) Set(value string) error {
	_, err := GetLogLevel(value)
	if err != nil {
		return fmt.Errorf("valid values are %s", formatValuesForUsageDocs(LogLevel))
	}
	*e.string = value
	return nil
}
func (e *logLevelValue) String() string {
	return *e.string
}

func (e *logLevelValue) Type() string {
	return "string"
}

func AddCmdFlag(cmd *cobra.Command, flagSet *pflag.FlagSet, p *string, name, shorthand string) *pflag.Flag {
	*p = LogLevelInfo
	val := &logLevelValue{string: p}
	f := flagSet.VarPF(val, name, shorthand, fmt.Sprintf("%s: %s", "Set log level", formatValuesForUsageDocs(LogLevel)))
	_ = cmd.RegisterFlagCompletionFunc(name, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return LogLevel, cobra.ShellCompDirectiveNoFileComp
	})
	return f
}
