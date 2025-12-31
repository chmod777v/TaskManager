package main

import (
	"cmd/main.go/internal/config"
	"cmd/main.go/internal/logger"
	"log/slog"
)

func main() {
	cfg := config.LoadConfig()
	logger.InitLogger(cfg.Env)
	slog.Info("Cfg, Logger launched successfully")

}
