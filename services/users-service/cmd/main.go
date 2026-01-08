package main

import (
	"log/slog"
	"users-service/internal/config"
	"users-service/internal/logger"
)

func main() {
	cfg := config.LoadConfig()
	logger.InitLogger(cfg.Env)
	slog.Info("Cfg, Logger launched successfully")
}
