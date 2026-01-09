package usersgrpc

import (
	"context"
	"log/slog"
	usersv1 "taskmanager/gen/go/users"
)

type Server struct {
	usersv1.UnimplementedUsersServer
}

func (s *Server) Validate(ctx context.Context, req *usersv1.ValidateRequest) (*usersv1.ValidateResponse, error) {
	slog.Debug("Validate request", "Login", req.Login, "Key", req.Key)
	resp := &usersv1.ValidateResponse{
		Valid: 1,
		Error: "",
	}
	return resp, nil
}
