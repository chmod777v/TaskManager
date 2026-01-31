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

func UsersMy(usersGrpcClient *users.Client) http.HandlerFunc { //Просмотр своего профиля
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{
				"Error": "Токен не найден",
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		information, err := usersGrpcClient.UserInformation(ctx, cookie.Value)

		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{
				"Error": "Server error",
			})
			slog.Error("UsersMy error", "ERROR", err.Error())
			return
		}
		if information.Login == "" {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, map[string]string{
				"Error": "Ваши данные невалидны",
			})
			return
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, map[string]interface{}{
			"Login":       information.Login,
			"Key":         information.Key,
			"AccessLevel": information.Accesslevel,
		})
	}
}

func UsersCreate(usersGrpcClient *users.Client) http.HandlerFunc { //Создать пользователя
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
				"Error": "Login or key or accesslevel are required",
			})
			return
		}
		if !(usersReq.Accesslevel == 1 || usersReq.Accesslevel == 2) {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{
				"Error": "Level must be 1 or 2",
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

		render.Status(r, http.StatusOK)
		render.JSON(w, r, map[string]int64{
			"Id": create.Id,
		})
	}
}

func UsersGetAll(usersGrpcClient *users.Client) http.HandlerFunc { //Получение таблицы с пользователями (Id:Login)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("some data"))
	}
}

func UsersView(usersGrpcClient *users.Client) http.HandlerFunc { //Получение деталей пользователя по id
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("some data"))
	}
}

func UsersDelete(usersGrpcClient *users.Client) http.HandlerFunc { //Удаление пользователя по id
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("some data"))
	}
}
