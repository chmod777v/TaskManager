package my_middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"cmd/main.go/internal/grpc/auth"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func Logger(next http.Handler) http.Handler {
	slog.Info("logger middleware enabled")
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

func Auth(authGrpcClient *auth.Client, level int8) func(next http.Handler) http.Handler { //level: 1-работник, 2-начальник и админ
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			valide, err := authGrpcClient.Validate(context.Background(), "123")
			if err != nil {
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, map[string]string{
					"Message": "Error",
				})
				slog.Error("AuthL1 error", "ERROR", err.Error())
				return
			}
			if !valide { // ДОП ПРОВЕРКА С level!!!
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, map[string]string{
					"Message": "User not valid",
				})
				slog.Debug("User not valide")
				return
			}
			slog.Debug("User valide")
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
