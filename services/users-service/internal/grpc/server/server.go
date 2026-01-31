package usersgrpc

import (
	"context"
	"errors"
	"log/slog"
	usersv1 "taskmanager/gen/go/users"
	"time"
	"users-service/internal/grpc/auth"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	usersv1.UnimplementedUsersServer
	Dbpool         *pgxpool.Pool
	AuthGrpcClient *auth.Client
}

func (s *Server) Validate(ctx context.Context, req *usersv1.ValidateRequest) (*usersv1.ValidateResponse, error) {
	slog.Debug("Validate request", "Login", req.Login, "Key", req.Key)

	var id, accesslevel int
	err := s.Dbpool.QueryRow(ctx,
		"SELECT id,accesslevel FROM users WHERE login=$1 and key=$2",
		req.Login, req.Key).Scan(&id, &accesslevel)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			resp := &usersv1.ValidateResponse{
				Valid: 0,
			}
			return resp, nil
		}
		slog.Error("Validate err, error sending query to database", "ERROR", err.Error())
		return nil, status.Errorf(codes.Internal, "Validate err, error sending query to database: %v", err)
	}

	resp := &usersv1.ValidateResponse{
		Id:    int64(id),
		Valid: int32(accesslevel),
	}
	return resp, nil
}

func (s *Server) Create(ctx context.Context, req *usersv1.CreateRequest) (*usersv1.CreateResponse, error) {
	slog.Debug("Create request", "Login", req.Login, "Key", req.Key, "Accesslevel", req.Accesslevel)

	var id int
	err := s.Dbpool.QueryRow(ctx,
		"INSERT INTO users (login, key, accesslevel) VALUES ($1, $2, $3) RETURNING id",
		req.Login, req.Key, req.Accesslevel).Scan(&id)
	if err != nil {
		slog.Error("Create err, error sending query to database", "ERROR", err.Error())
		return nil, status.Errorf(codes.Internal, "Create err, error sending query to database: %v", err)
	}

	resp := &usersv1.CreateResponse{
		Id: int64(id),
	}
	return resp, nil
}

func (s *Server) UserInformation(ctx context.Context, req *usersv1.UserInformationRequest) (*usersv1.UserInformationResponse, error) {
	slog.Debug("UserInformation request", "Token", req.Token)

	//Запрос к сервису Auth
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	getId, err := s.AuthGrpcClient.GetId(ctx, req.Token)
	if err != nil {
		slog.Error("UserInformation error", "ERROR", err.Error())
		return nil, err
	}
	if getId.Id == 0 { //не валиден
		return nil, nil
	}

	//Запроc в БД
	var key, login string
	var accesslevel int32
	err = s.Dbpool.QueryRow(ctx,
		"SELECT login, key, accesslevel FROM users WHERE id=$1",
		getId.Id).Scan(&login, &key, &accesslevel)
	if err != nil {
		slog.Error("UserInformation error", "ERROR", err.Error())
		return nil, status.Errorf(codes.Internal, "UserInformation err, error sending query to database: %v", err)
	}

	//
	resp := &usersv1.UserInformationResponse{
		Login:       login,
		Key:         key,
		Accesslevel: accesslevel,
	}
	return resp, nil
}
