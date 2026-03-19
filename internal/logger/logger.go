package logger

import (
	"log/slog"
	"os"
)

func New(level int) *slog.Logger {
	return slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.Level(level),
			AddSource: true,
		}),
	)
}
