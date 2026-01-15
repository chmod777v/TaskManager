package main

import (
	"auth-service/internal/config"
	authgrpc "auth-service/internal/grpc/server"
	"auth-service/internal/grpc/users"
	"auth-service/internal/logger"
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	authv1 "taskmanager/gen/go/auth"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.LoadConfig()
	logger.InitLogger(cfg.Env)
	slog.Info("Cfg, Logger launched successfully")

	//Redis
	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: cfg.Redis.Password,
		DB:       0,
	})

	ping, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		panic("Failed to connect to the redis: " + err.Error())
	}
	slog.Info("Redis launched successfully", "Response to ping", ping, "Host", redisAddr)

	//Server
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.GrpcPort)
	listener, err := net.Listen("tcp", serverAddr)
	if err != nil {
		panic("Failed to listen: " + err.Error())
	}

	server := grpc.NewServer()
	reflection.Register(server) //сообщает методы

	usersGrpcClient := users.NewClient(cfg.Services.Users.Host, cfg.Services.Users.GrpcPort)
	authv1.RegisterAuthServer(server, &authgrpc.Server{
		UsersGrpcClient: usersGrpcClient,
		RedisClient:     redisClient,
	})

	serverErr := make(chan error, 1)
	go func() {
		if err := server.Serve(listener); err != nil && err != grpc.ErrServerStopped {
			serverErr <- err
		}
	}()
	slog.Info("Server started", "LINK", listener.Addr())
	Shutdown(server, serverErr, redisClient)
}

func Shutdown(server *grpc.Server, serverErr chan error, redisClient *redis.Client) {
	defer close(serverErr)
	//Waiting for a signal
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(interruptChan)

	select {
	case <-interruptChan:
		slog.Info("Shutdown")
	case err := <-serverErr:
		slog.Error("Server error", "ERROR", err)
	}

	//gRPC server shutdown
	slog.Info("Stopping server...")

	done := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(done)
	}()

	const shutdownTimeout = 10 * time.Second
	select {
	case <-done:
		slog.Info("Server stopped successfully")
	case <-time.After(shutdownTimeout):
		slog.Info("Graceful shutdown timeout, forcing stop")
		server.Stop() // Принудительная остановка
		slog.Info("gRPC server stopped")
	}

	//Redis
	err := redisClient.Close()
	if err != nil {
		slog.Error("Failed to close redis client", "ERROR", err.Error())
	}
	slog.Info("Redis client closed successfully")
}
