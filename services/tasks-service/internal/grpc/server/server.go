package tasksgrpc

import (
	"context"
	"log/slog"
	tasksv1 "taskmanager/gen/go/tasks"
	"tasks-service/internal/grpc/auth"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	tasksv1.UnimplementedTasksServer
	Dbpool         *pgxpool.Pool
	AuthGrpcClient *auth.Client
}

func (s *Server) TasksMy(ctx context.Context, req *tasksv1.TasksMyRequest) (*tasksv1.TasksMyResponse, error) {
	slog.Debug("TasksMy request", "Token", req.Token)

	//Запрос к сервису Auth
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	getId, err := s.AuthGrpcClient.GetId(ctx, req.Token)
	if err != nil {
		slog.Error("TasksMy error", "ERROR", err.Error())
		return nil, err
	}
	if getId.Id == 0 { //не валиден
		return nil, nil
	}

	//Запроc в БД
	rows, err := s.Dbpool.Query(ctx,
		"SELECT header, task FROM tasks_users JOIN tasks ON task_id = id WHERE user_id = $1;",
		getId.Id)
	if err != nil {
		slog.Error("TasksMy error, error sending query to database", "ERROR", err.Error())
		return nil, status.Errorf(codes.Internal, "TasksMy err, error sending query to database: %v", err)
	}
	defer rows.Close()

	//Парс
	var tasks []*tasksv1.TasksMyResponse_Task
	for rows.Next() {
		var header, task string

		if err := rows.Scan(&header, &task); err != nil {
			slog.Error("TasksMy error, error scanning row", "ERROR", err.Error())
			return nil, status.Errorf(codes.Internal, "TasksMy error, error scanning row: %v", err)
		}

		tasks = append(tasks, &tasksv1.TasksMyResponse_Task{
			Header: header,
			Task:   task,
		})
	}

	resp := &tasksv1.TasksMyResponse{
		Tasks: tasks,
	}
	return resp, nil
}
