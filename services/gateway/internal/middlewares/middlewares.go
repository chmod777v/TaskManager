package my_middleware

import (
	"log/slog"
	"net/http"
)

func Logger(next http.Handler) http.Handler {
	slog.Info("logger middleware enabled")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		next.ServeHTTP(w, r)
	})
}

func Auth(next http.Handler) http.Handler {
	slog.Info("Auth middleware enabled")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}
