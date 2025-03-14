package logger

import (
	"log/slog"
	"os"
	"strings"
)

var Logger *slog.Logger
var gLogLevel slog.LevelVar

func SetLogLevel(level slog.Level) {
	gLogLevel.Set(level)
}

func SetLogLevelfromString(level string) {
	_level := strings.ToUpper(strings.TrimSpace(level))
	switch _level {
	case "DEBUG":
		SetLogLevel(slog.LevelDebug)
	case "INFO":
		SetLogLevel(slog.LevelInfo)
	case "WARN":
		SetLogLevel(slog.LevelWarn)
	case "ERROR":
		SetLogLevel(slog.LevelError)
	default:
		SetLogLevel(slog.LevelInfo)
	}
}

func Init(logFile string) {
	gLogLevel.Set(slog.LevelInfo)
	logopts := slog.HandlerOptions{Level: &gLogLevel}
	var handler slog.Handler
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			handler = slog.NewTextHandler(os.Stderr, &logopts)
		}
		handler = slog.NewTextHandler(f, &logopts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, &logopts)
	}
	Logger = slog.New(handler)
}
