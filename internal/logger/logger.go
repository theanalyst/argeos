package logger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
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

func makeStdErrHandler(logopts *slog.HandlerOptions) slog.Handler {
	return slog.NewTextHandler(os.Stderr, logopts)
}

func Init(logFile string) {
	gLogLevel.Set(slog.LevelInfo)
	logopts := slog.HandlerOptions{Level: &gLogLevel}
	var handler slog.Handler
	if logFile != "" {
		err := os.MkdirAll(filepath.Dir(logFile), 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating log directory: %s", err)
			handler = makeStdErrHandler(&logopts)
			Logger = slog.New(handler)
			return
		}

		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening log file: %s", err)
			handler = makeStdErrHandler(&logopts)
			Logger = slog.New(handler)
			return
		}

		handler = slog.NewTextHandler(f, &logopts)
	} else {
		handler = makeStdErrHandler(&logopts)
	}
	Logger = slog.New(handler)
}
