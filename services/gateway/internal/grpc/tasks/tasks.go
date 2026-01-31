package tasks

import (
	"context"
	"fmt"
	"log/slog"
	tasksv1 "taskmanager/gen/go/tasks"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	Client tasksv1.TasksClient
}

func NewClient(host string, port int) *Client {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Failed to conect to tasks service:", "ERROR", err.Error())
		return nil
	}
	slog.Info("Conect to tasks service:", "Host", addr)
	return &Client{
		conn:   conn,
		Client: tasksv1.NewTasksClient(conn),
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
		slog.Error("Failed to close tasks-service connection", "ERROR", "timeout while closing connection: "+ctx.Err().Error())

	case err := <-done:
		if err != nil {
			slog.Error("Failed to close tasks-service connection", "ERROR", err)
		} else {
			slog.Info("Tasks-service connection closed successfully")
		}
	}
}
func (c *Client) TasksMy(ctx context.Context, token string) (*tasksv1.TasksMyResponse, error) {
	if c.Client == nil {
		return nil, fmt.Errorf("Tasks gRPC client is not initialized")
	}
	resp, err := c.Client.TasksMy(ctx, &tasksv1.TasksMyRequest{
		Token: token,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}
