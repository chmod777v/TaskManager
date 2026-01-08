package auth

import (
	"context"
	"fmt"
	"log/slog"
	authv1 "taskmanager/gen/go/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	Client authv1.AuthClient
}

func NewClient(host string, port int) *Client {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Failed to conect to server:", "ERROR", err.Error())
	}
	slog.Info("Conect to gRPC server:", "Host", addr)
	return &Client{
		Client: authv1.NewAuthClient(conn),
	}
}

func (c *Client) Validate(ctx context.Context, token string) (*authv1.ValidateResponse, error) {
	resp, err := c.Client.Validate(ctx, &authv1.ValidateRequest{
		Token: token,
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) Authenticate(ctx context.Context, login, key string) (*authv1.AuthenticateResponse, error) {
	resp, err := c.Client.Authenticate(ctx, &authv1.AuthenticateRequest{
		Login: login,
		Key:   key,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}
