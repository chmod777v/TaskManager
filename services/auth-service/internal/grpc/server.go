package authgrpc

import (
	"context"
	"log/slog"
	authv1 "taskmanager/gen/go/auth"
)

type Server struct {
	authv1.UnimplementedAuthServer
}

func (s *Server) Validate(ctx context.Context, req *authv1.ValidateRequest) (*authv1.ValidateResponse, error) {
	slog.Debug("request", "Token", req.Token)
	resp := &authv1.ValidateResponse{
		Valid: false,
		Error: "",
	}
	return resp, nil
}

func (s *Server) Authenticate(ctx context.Context, req *authv1.AuthenticateRequest) (*authv1.AuthenticateResponse, error) {
	resp := &authv1.AuthenticateResponse{
		Success: true,
		Token:   "123",
		Error:   "",
	}
	return resp, nil
}
