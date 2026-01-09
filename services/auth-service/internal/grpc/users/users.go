package users

import (
	"context"
	"fmt"
	"log/slog"
	usersv1 "taskmanager/gen/go/users"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	Client usersv1.UsersClient
}

func NewClient(host string, port int) *Client {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Failed to conect to server:", "ERROR", err.Error())
	}
	slog.Info("Conect to gRPC server:", "Host", addr)
	return &Client{
		Client: usersv1.NewUsersClient(conn),
	}
}

func (c *Client) Validate(ctx context.Context, login, key string) (*usersv1.ValidateResponse, error) {
	resp, err := c.Client.Validate(ctx, &usersv1.ValidateRequest{
		Login: login,
		Key:   key,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}
