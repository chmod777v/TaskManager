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

	var accesslevel int
	err := s.Dbpool.QueryRow(ctx,
		"SELECT accesslevel FROM users WHERE login=$1 and key=$2",
		req.Login, req.Key).Scan(&accesslevel)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			resp := &usersv1.ValidateResponse{
				Valid: 0,
				Error: "",
			}
			return resp, nil
		}
		resp := &usersv1.ValidateResponse{
			Valid: 0,
			Error: "Validate err, error sending query to database: " + err.Error(),
		}
		slog.Error("Validate err, error sending query to database", "ERROR", err.Error())
		return resp, nil
	}

	resp := &usersv1.ValidateResponse{
		Valid: int32(accesslevel),
		Error: "",
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
		resp := &usersv1.CreateResponse{
			Id:    0,
			Error: "Create err, error sending query to database: " + err.Error(),
		}
		slog.Error("Create err, error sending query to database", "ERROR", err.Error())
		return resp, nil
	}

	resp := &usersv1.CreateResponse{
		Id:    int64(id),
		Error: "",
	}
	return resp, nil
}

func (s *Server) UserInformation(ctx context.Context, req *usersv1.UserInformationRequest) (*usersv1.UserInformationResponse, error) {
	slog.Debug("UserInformation request", "Token", req.Token)

	//Запрос к сервису Auth
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	getLogin, err := s.AuthGrpcClient.GetLogin(ctx, req.Token)
	if err != nil {
		slog.Error("UserInformation error", "ERROR", err.Error())
		return nil, err
	}
	if getLogin.Login == "" { //не валиден
		return nil, nil
	}

	//Запроc в БД
	var key string
	var accesslevel int32
	err = s.Dbpool.QueryRow(ctx,
		"SELECT key, accesslevel FROM users WHERE login=$1",
		getLogin.Login).Scan(&key, &accesslevel)
	if err != nil {
		slog.Error("UserInformation error", "ERROR", err.Error())
		return nil, status.Errorf(codes.Internal, "UserInformation err, error sending query to database: %v", err)
	}

	//
	resp := &usersv1.UserInformationResponse{
		Login:       getLogin.Login,
		Key:         key,
		Accesslevel: accesslevel,
	}
	return resp, nil
}
