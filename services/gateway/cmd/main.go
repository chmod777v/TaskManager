package main

import (
	"cmd/main.go/internal/config"
	"cmd/main.go/internal/handler"
	"cmd/main.go/internal/logger"
	my_middleware "cmd/main.go/internal/middlewares"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// go run cmd\main.go -config config\config.yaml
func main() {
	cfg := config.LoadConfig()
	logger.InitLogger(cfg.Env)
	slog.Info("Cfg, Logger launched successfully")

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(my_middleware.Logger) //  !!!!!

	router.Route("/users", func(r chi.Router) {
		r.Use(my_middleware.Auth)
	})
	router.Route("/tasks", func(r chi.Router) {
		// /tasks/my - для обычного разраба
		router.Group(func(r chi.Router) {
			r.Use(my_middleware.Auth)
		})
	})
	router.Route("/assign", func(r chi.Router) {
		r.Use(my_middleware.Auth)
	})
	router.Post("/auth", handler.Auth)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  cfg.Server.Timeout,
		WriteTimeout: cfg.Server.Timeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		slog.Error("Error starting server", "ERROR", err.Error())
	}

}
