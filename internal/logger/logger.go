package logger

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

func Init(logFile string) {
	var handler slog.Handler
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			handler = slog.NewTextHandler(os.Stderr, nil)
		}
		handler = slog.NewTextHandler(f, nil)
	} else {
		handler = slog.NewTextHandler(os.Stderr, nil)
	}
	Logger = slog.New(handler)
}
