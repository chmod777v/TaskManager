package authgrpc

import (
	"auth-service/internal/grpc/users"
	"context"
	"log/slog"
	authv1 "taskmanager/gen/go/auth"
	"time"
)

type Server struct {
	authv1.UnimplementedAuthServer
	UsersGrpcClient *users.Client
}

func (s *Server) Validate(ctx context.Context, req *authv1.ValidateRequest) (*authv1.ValidateResponse, error) {
	slog.Debug("Validate request", "Token", req.Token)
	//Проверка токена из редис
	resp := &authv1.ValidateResponse{
		Valid: 1, //0-невалиден 1-работник 2-начальник/админ
		Error: "",
	}
	return resp, nil
}

func (s *Server) Authenticate(ctx context.Context, req *authv1.AuthenticateRequest) (*authv1.AuthenticateResponse, error) {
	slog.Debug("Authenticate request", "Login", req.Login, "Key", req.Key)

	//Запрос к сервису users
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	validate, err := s.UsersGrpcClient.Validate(ctx, req.Login, req.Key)

	if err != nil {
		slog.Error("Authenticate error", "ERROR", err.Error())
		return nil, err
	}
	if validate.Error != "" {
		resp := &authv1.AuthenticateResponse{
			Success: false,
			Token:   "",
			Error:   validate.Error,
		}
		slog.Error("Authenticate grpc server error", "ERROR", validate.Error)
		return resp, nil
	}

	if validate.Valid == 0 { // Не авторизован
		resp := &authv1.AuthenticateResponse{
			Success: false,
			Token:   "",
			Error:   "",
		}
		return resp, nil
	}

	//Создание новой редис сессии (содержит уровень доступа)
	resp := &authv1.AuthenticateResponse{
		Success: true,
		Token:   "12345",
		Error:   "",
	}
	return resp, nil
}
