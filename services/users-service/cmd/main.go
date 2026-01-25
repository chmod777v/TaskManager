package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	usersv1 "taskmanager/gen/go/users"
	"time"
	"users-service/internal/config"
	"users-service/internal/grpc/auth"
	usersgrpc "users-service/internal/grpc/server"
	"users-service/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.LoadConfig()
	logger.InitLogger(cfg.Env)
	slog.Info("Cfg, Logger launched successfully")

	//GRPC services
	authGrpcClient := auth.NewClient(cfg.Services.Auth.Host, cfg.Services.Auth.GrpcPort)
	if authGrpcClient == nil {
		return
	}
	defer authGrpcClient.Close()

	//DB
	dbLink := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.Db.Username, cfg.Db.Password, cfg.Db.Host, cfg.Db.Port, cfg.Db.DbName)
	dbpool, err := pgxpool.New(context.Background(), dbLink)
	if err != nil {
		slog.Error("Failed to connect to the postgreSQL", "ERROR", err.Error())
		return
	}
	slog.Info("Database connection successfully")

	defer func() {
		dbpool.Close()
		dbpool = nil
		slog.Info("Database connection closed successfully")
	}()

	//server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.GrpcPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("Failed to listen", "ERROR", err.Error())
		return
	}

	server := grpc.NewServer()
	reflection.Register(server) //сообщает методы

	usersv1.RegisterUsersServer(server, &usersgrpc.Server{Dbpool: dbpool, AuthGrpcClient: authGrpcClient})

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
