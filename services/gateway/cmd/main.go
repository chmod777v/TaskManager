package main

import (
	"fmt"
	"gateway/internal/config"
	"gateway/internal/grpc/auth"
	"gateway/internal/grpc/tasks"
	"gateway/internal/grpc/users"
	"gateway/internal/logger"
	"gateway/internal/server"
	"log/slog"
	"net/http"
)

// go run cmd\main.go -config config\config.yaml
func main() {
	cfg := config.LoadConfig()
	logger.InitLogger(cfg.Env)
	slog.Info("Cfg, Logger launched successfully")

	//Services
	authGrpcClient := auth.NewClient(cfg.Services.Auth.Host, cfg.Services.Auth.GrpcPort)
	if authGrpcClient == nil {
		return
	}
	defer authGrpcClient.Close()

	usersGrpcClient := users.NewClient(cfg.Services.Users.Host, cfg.Services.Users.GrpcPort)
	if usersGrpcClient == nil {
		return
	}
	defer usersGrpcClient.Close()

	tasksGrpcClient := tasks.NewClient(cfg.Services.Tasks.Host, cfg.Services.Tasks.GrpcPort)
	if tasksGrpcClient == nil {
		return
	}
	defer tasksGrpcClient.Close()

	//Server
	router := server.NewRouter(authGrpcClient, usersGrpcClient, tasksGrpcClient)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.HttpPort)
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
	}()
	slog.Info("Server started", "LINK", addr)

	server.Shutdown(serv, serverErr)
}
