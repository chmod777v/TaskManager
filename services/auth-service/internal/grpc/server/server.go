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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	authv1.UnimplementedAuthServer
	RedisClient     *redis.Client
	UsersGrpcClient *users.Client
}

func (s *Server) Validate(ctx context.Context, req *authv1.ValidateRequest) (*authv1.ValidateResponse, error) {
	slog.Debug("Validate request", "Token", req.Token)

	redisResp, err := s.RedisClient.HGet(ctx, "session:"+req.Token, "accessLevel").Int()
	if err != nil {
		if errors.Is(err, redis.Nil) { //не валиден
			resp := &authv1.ValidateResponse{
				Valid: 0,
			}
			return resp, nil
		}
		return nil, status.Errorf(codes.Internal, "Redis validation err: %v", err)
	}

	resp := &authv1.ValidateResponse{
		Valid: int32(redisResp), //0-невалиден 1-работник 2-начальник/админ
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
		return nil, status.Errorf(codes.Internal, "Authenticate error: %v", err)
	}

	if validate.Valid == 0 { // Не авторизован
		resp := &authv1.AuthenticateResponse{
			Success: false,
			Token:   "",
		}
		return resp, nil
	}

	//Генерация токена
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, status.Errorf(codes.Internal, "Generate token error: %v", err)
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)

	//Создание новой редис сессии (содержит уровень доступа и login)
	key := "session:" + token
	err = s.RedisClient.HSet(ctx, key,
		"id", validate.Id,
		"accessLevel", validate.Valid,
	).Err()

	if err != nil {
		return nil, status.Errorf(codes.Internal, "HSet redis error: %v", err)
	}
	err = s.RedisClient.Expire(ctx, key, 48*time.Hour).Err()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Expire redis error: %v", err)
	}

	resp := &authv1.AuthenticateResponse{
		Success: true,
		Token:   token,
	}
	return resp, nil
}

func (s *Server) GetId(ctx context.Context, req *authv1.GetIdRequest) (*authv1.GetIdResponse, error) {
	slog.Debug("GetId request", "Token", req.Token)

	redisResp, err := s.RedisClient.HGet(ctx, "session:"+req.Token, "id").Int()
	if err != nil {
		if errors.Is(err, redis.Nil) { //не валиден
			return nil, nil
		}
		return nil, status.Errorf(codes.Internal, "GetId err, HGet redis error: %v", err)
	}
	resp := &authv1.GetIdResponse{
		Id: int64(redisResp),
	}
	return resp, nil
}
