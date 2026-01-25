package auth_handler

import (
	"context"
	"gateway/internal/grpc/auth"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/render"
)

type authRequest struct {
	Login string `json:"Login"`
	Key   string `json:"Key"`
}

func Auth(authGrpcClient *auth.Client) http.HandlerFunc { //Аутентификация по логину и ключу
	return func(w http.ResponseWriter, r *http.Request) {
		var authReq authRequest
		if err := render.DecodeJSON(r.Body, &authReq); err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{
				"error": "Invalid request format",
			})
			return
		}
		if authReq.Login == "" || authReq.Key == "" {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{
				"Error": "Login or key are required",
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		authenticate, err := authGrpcClient.Authenticate(ctx, authReq.Login, authReq.Key)

		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{
				"Error": "Server error",
			})
			slog.Error("Auth error", "ERROR", err.Error())
			return
		}
		if authenticate.Error != "" {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{
				"Error": "Server error",
			})
			slog.Error("Auth grpc server error", "ERROR", authenticate.Error)
			return
		}

		if !authenticate.Success {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, map[string]string{
				"Error": "Вы не авторизованы",
			})
			return
		}

		cookie := &http.Cookie{
			Name:     "auth_token",
			Value:    authenticate.Token,
			Path:     "/",
			MaxAge:   172800,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		}
		http.SetCookie(w, cookie)
		w.WriteHeader(http.StatusOK)
	}
}
