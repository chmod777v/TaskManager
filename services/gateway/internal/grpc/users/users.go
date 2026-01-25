package users

import (
	"context"
	"fmt"
	"log/slog"
	usersv1 "taskmanager/gen/go/users"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	Client usersv1.UsersClient
}

func NewClient(host string, port int) *Client {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Failed to conect to server:", "ERROR", err.Error())
		return nil
	}
	slog.Info("Conect to users service:", "Host", addr)
	return &Client{
		conn:   conn,
		Client: usersv1.NewUsersClient(conn),
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
		slog.Error("Failed to close users-service connection", "ERROR", "timeout while closing connection: "+ctx.Err().Error())

	case err := <-done:
		if err != nil {
			slog.Error("Failed to close users-service connection", "ERROR", err)
		} else {
			slog.Info("Users-service connection closed successfully")
		}
	}
}

func (c *Client) Validate(ctx context.Context, login, key string) (*usersv1.ValidateResponse, error) {
	if c.Client == nil {
		return nil, fmt.Errorf("Users gRPC client is not initialized")
	}
	resp, err := c.Client.Validate(ctx, &usersv1.ValidateRequest{
		Login: login,
		Key:   key,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) Create(ctx context.Context, login, key string, accesslevel int32) (*usersv1.CreateResponse, error) {
	if c.Client == nil {
		return nil, fmt.Errorf("Users gRPC client is not initialized")
	}
	resp, err := c.Client.Create(ctx, &usersv1.CreateRequest{
		Login:       login,
		Key:         key,
		Accesslevel: accesslevel,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) UserInformation(ctx context.Context, token string) (*usersv1.UserInformationResponse, error) {
	if c.Client == nil {
		return nil, fmt.Errorf("Users gRPC client is not initialized")
	}
	resp, err := c.Client.UserInformation(ctx, &usersv1.UserInformationRequest{
		Token: token,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}
