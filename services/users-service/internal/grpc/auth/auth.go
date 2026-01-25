package auth

import (
	"context"
	"fmt"
	"log/slog"
	authv1 "taskmanager/gen/go/auth"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	Client authv1.AuthClient
}

func NewClient(host string, port int) *Client {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Failed to conect to server:", "ERROR", err.Error())
		return nil
	}
	slog.Info("Conect to auth service:", "Host", addr)
	return &Client{
		conn:   conn,
		Client: authv1.NewAuthClient(conn),
	}
}
func (c *Client) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if c.conn == nil {
		return
	}
	done := make(chan error, 1)
	go func() {
		done <- c.conn.Close()
	}()

	select {
	case <-ctx.Done():
		slog.Error("Failed to close auth-service connection", "ERROR", "timeout while closing connection: "+ctx.Err().Error())

	case err := <-done:
		if err != nil {
			slog.Error("Failed to close auth-service connection", "ERROR", err)
		} else {
			slog.Info("Auth-service connection closed successfully")
		}
	}
}

func (c *Client) GetLogin(ctx context.Context, token string) (*authv1.GetLoginResponse, error) {
	if c.Client == nil {
		return nil, fmt.Errorf("gRPC client is not initialized")
	}
	resp, err := c.Client.GetLogin(ctx, &authv1.GetLoginRequest{
		Token: token,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}
