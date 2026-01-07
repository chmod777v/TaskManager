package main

import (
	"cmd/main.go/internal/config"
	"cmd/main.go/internal/grpc/auth"
	"cmd/main.go/internal/logger"
	"cmd/main.go/internal/server"
	"fmt"
	"log/slog"
	"net/http"
)

// go run cmd\main.go -config config\config.yaml
func main() {
	cfg := config.LoadConfig()
	logger.InitLogger(cfg.Env)
	slog.Info("Cfg, Logger launched successfully")

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.HttpPort)
	authGrpcClient := auth.NewClient(cfg.Services.Auth.Host, cfg.Services.Auth.GrpcPort)
	router := server.NewRouter(authGrpcClient)

	serv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.Server.Timeout,
		WriteTimeout: cfg.Server.Timeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		if err := serv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
		close(serverErr)
	}()
	slog.Info("Server started", "LINK", addr)

	server.Shutdown(serv, serverErr)
}
