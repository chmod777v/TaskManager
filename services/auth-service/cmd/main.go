package main

import (
	"cmd/main.go/internal/config"
	authgrpc "cmd/main.go/internal/grpc"
	"cmd/main.go/internal/logger"
	"fmt"
	"log/slog"
	"net"
	authv1 "taskmanager/gen/go/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.LoadConfig()
	logger.InitLogger(cfg.Env)
	slog.Info("Cfg, Logger launched successfully")

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.GrpcPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic("Failed to listen:" + err.Error())
	}

	server := grpc.NewServer()
	reflection.Register(server) //сообщает методы

	authv1.RegisterAuthServer(server, &authgrpc.Server{})
	slog.Info("Server started", "LINK", listener.Addr())

	if err = server.Serve(listener); err != nil {
		panic("Failed to serve:" + err.Error())
	}
}
