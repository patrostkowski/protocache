package logger

import (
	"log/slog"
	"os"
)

var (
	Logger *slog.Logger

	Info  func(msg string, args ...any)
	Warn  func(msg string, args ...any)
	Error func(msg string, args ...any)
	Debug func(msg string, args ...any)
)

func init() {
	Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // Allow all levels
	}))

	Info = Logger.Info
	Warn = Logger.Warn
	Error = Logger.Error
	Debug = Logger.Debug
}
