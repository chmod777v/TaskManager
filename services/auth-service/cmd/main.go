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

	//GRPC services
	usersGrpcClient := users.NewClient(cfg.Services.Users.Host, cfg.Services.Users.GrpcPort)
	if usersGrpcClient == nil {
		return
	}
	defer usersGrpcClient.Close() //CLOSE

	//Redis
	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: cfg.Redis.Password,
		DB:       0,
	})

	ping, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		slog.Error("Failed to connect to the redis", "ERROR", err.Error())
		return
	}
	slog.Info("Redis launched successfully", "Response to ping", ping, "Host", redisAddr)

	defer func() { //CLOSE
		if err := redisClient.Close(); err != nil {
			slog.Error("Failed to close redis client", "ERROR", err.Error())
		} else {
			slog.Info("Redis client closed successfully")
		}

	}()

	//Server
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.GrpcPort)
	listener, err := net.Listen("tcp", serverAddr)
	if err != nil {
		slog.Error("Failed to listen", "ERROR", err.Error())
		return
	}

	server := grpc.NewServer()
	reflection.Register(server) //сообщает методы

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
	Shutdown(server, serverErr)
}

func Shutdown(server *grpc.Server, serverErr chan error) {
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
}
