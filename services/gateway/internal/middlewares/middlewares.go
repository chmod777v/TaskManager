package my_middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"gateway/internal/grpc/auth"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entry := slog.With(
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("request_id", middleware.GetReqID(r.Context())),
			slog.String("remote_addr", r.RemoteAddr),
			//slog.String("user_agent", r.UserAgent()),
		)
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()

		next.ServeHTTP(ww, r)

		entry.Info("request",
			slog.Int("status", ww.Status()),
			slog.Int("bytes", ww.BytesWritten()),
			slog.String("duration", time.Since(start).String()),
		)
	})
}

func Auth(authGrpcClient *auth.Client, level int32) func(next http.Handler) http.Handler { //level: 1-работник, 2-начальник и админ
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
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
			valide, err := authGrpcClient.Validate(ctx, cookie.Value)

			if err != nil {
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, map[string]string{
					"Error": "Server error",
				})
				slog.Error("AuthMiddleware error", "ERROR", err.Error())
				return
			}
			if valide.Error != "" {
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, map[string]string{
					"Error": "Server error",
				})
				slog.Error("AuthMiddleware grpc server error", "ERROR", valide.Error)
				return
			}

			if valide.Valid == 0 {
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, map[string]string{
					"Error": "Ваши данные невалидны",
				})
				return
			}
			if valide.Valid < level {
				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, map[string]string{
					"Error": "У вас нет доступа",
				})
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
