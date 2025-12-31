package logger

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

func InitLogger(env string) {
	var handler slog.Handler
	switch env {
	case "local":
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: "15:04:05",
			NoColor:    false,
		})
	case "development":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	case "production":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}
	slog.SetDefault(slog.New(handler))
}
