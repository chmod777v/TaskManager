package authgrpc

import (
	"auth-service/internal/grpc/users"
	"context"
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	authv1 "taskmanager/gen/go/auth"
	"time"

	"github.com/redis/go-redis/v9"
)

type Server struct {
	authv1.UnimplementedAuthServer
	UsersGrpcClient *users.Client
	RedisClient     *redis.Client
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

	//Генерация токена
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		resp := &authv1.AuthenticateResponse{
			Success: false,
			Token:   "",
			Error:   "Generate token error: " + err.Error(),
		}
		slog.Error("Generate token error", "ERROR", err.Error())
		return resp, nil
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)

	//Создание новой редис сессии (содержит уровень доступа)
	key := "session:" + token
	err = s.RedisClient.HSet(ctx, key,
		"accessLevel", validate.Valid,
	).Err()

	if err != nil {
		resp := &authv1.AuthenticateResponse{
			Success: false,
			Token:   "",
			Error:   "HSet redis error: " + err.Error(),
		}
		slog.Error("HSet redis error", "ERROR", err.Error())
		return resp, nil
	}
	err = s.RedisClient.Expire(ctx, key, 48*time.Hour).Err()
	if err != nil {
		resp := &authv1.AuthenticateResponse{
			Success: false,
			Token:   "",
			Error:   "Expire redis error: " + err.Error(),
		}
		slog.Error("Expire redis error", "ERROR", err.Error())
		return resp, nil
	}

	resp := &authv1.AuthenticateResponse{
		Success: true,
		Token:   token,
		Error:   "",
	}
	return resp, nil
}
