package authgrpc

import (
	"auth-service/internal/grpc/users"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
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

	redisResp, err := s.RedisClient.HGet(ctx, "session:"+req.Token, "accessLevel").Int()
	if err != nil {
		if errors.Is(err, redis.Nil) { //не валиден
			resp := &authv1.ValidateResponse{
				Valid: 0,
				Error: "",
			}
			return resp, nil
		}
		return ErrorValidate("Redis validation err", err.Error()), nil
	}

	resp := &authv1.ValidateResponse{
		Valid: int32(redisResp), //0-невалиден 1-работник 2-начальник/админ
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
		return ErrorAuthenticate("Users-service error", validate.Error), nil
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
		return ErrorAuthenticate("Generate token error", err.Error()), nil
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)

	//Создание новой редис сессии (содержит уровень доступа и login)
	key := "session:" + token
	err = s.RedisClient.HSet(ctx, key,
		"login", req.Login,
		"accessLevel", validate.Valid,
	).Err()

	if err != nil {
		return ErrorAuthenticate("HSet redis error", err.Error()), nil
	}
	err = s.RedisClient.Expire(ctx, key, 48*time.Hour).Err()
	if err != nil {
		return ErrorAuthenticate("Expire redis error", err.Error()), nil
	}

	resp := &authv1.AuthenticateResponse{
		Success: true,
		Token:   token,
		Error:   "",
	}
	return resp, nil
}

func ErrorAuthenticate(message, err string) *authv1.AuthenticateResponse {
	resp := &authv1.AuthenticateResponse{
		Success: false,
		Token:   "",
		Error:   message + ": " + err,
	}
	slog.Error(message, "ERROR", err)
	return resp
}

func ErrorValidate(message, err string) *authv1.ValidateResponse {
	resp := &authv1.ValidateResponse{
		Valid: 0,
		Error: message + ": " + err,
	}
	slog.Error(message, "ERROR", err)
	return resp
}
