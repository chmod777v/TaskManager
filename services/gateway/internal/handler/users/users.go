package users_handler

import (
	"context"
	"gateway/internal/grpc/users"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/render"
)

type UsersRequest struct {
	Login       string `json:"Login"`
	Key         string `json:"Key"`
	Accesslevel int32  `json:"Accesslevel"`
}

func UsersGetAll(usersGrpcClient *users.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("some data"))
	}
}

func UsersCreate(usersGrpcClient *users.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var usersReq UsersRequest
		if err := render.DecodeJSON(r.Body, &usersReq); err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{
				"error": "Invalid request format",
			})
			return
		}
		if usersReq.Login == "" || usersReq.Key == "" || usersReq.Accesslevel == 0 {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{
				"Error": "Login or key are required",
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		create, err := usersGrpcClient.Create(ctx, usersReq.Login, usersReq.Key, usersReq.Accesslevel)

		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{
				"Error": "Server error",
			})
			slog.Error("UsersCreate error", "ERROR", err.Error())
			return
		}
		if create.Error != "" {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{
				"Error": "Server error",
			})
			slog.Error("Users grpc server error", "ERROR", create.Error)
			return
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, map[string]int64{
			"Id": create.Id,
		})
	}
}

func UsersView(usersGrpcClient *users.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("some data"))
	}
}

func UsersDelete(usersGrpcClient *users.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("some data"))
	}
}
