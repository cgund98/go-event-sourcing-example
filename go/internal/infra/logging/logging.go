package logging

import (
	"log/slog"
	"os"
	"strings"
)

func getLevel() slog.Level {
	levelStr := os.Getenv("LOG_LEVEL")

	if strings.ToUpper(levelStr) == "DEBUG" {
		return slog.LevelDebug
	} else if strings.ToUpper(levelStr) == "WARN" {
		return slog.LevelWarn
	} else {
		return slog.LevelInfo
	}
}

var Logger *slog.Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	Level: getLevel(),
}))

func Fatalf(msg string, args ...any) {
	Logger.Error(msg, args...)
	os.Exit(1)
}
