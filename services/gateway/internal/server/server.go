package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cmd/main.go/internal/grpc/auth"
	"cmd/main.go/internal/handler"
	my_middleware "cmd/main.go/internal/middlewares"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(authGrpcClient *auth.Client) *chi.Mux {
	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	router.Use(middleware.Recoverer) //Для перехвата паник
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(my_middleware.Logger)

	router.Route("/users", func(r chi.Router) {
		r.Use(my_middleware.Auth(authGrpcClient, 2))

		r.Get("/", handler.UsersGetAll)
		r.Post("/", handler.UsersCreate)
		r.Get("/{id}", handler.UsersView)
		r.Delete("/{id}", handler.UsersDelete)
	})
	router.Route("/tasks", func(rout chi.Router) {
		rout.Use(my_middleware.Auth(authGrpcClient, 1))
		rout.Get("/my", handler.TasksMy)

		rout.Group(func(r chi.Router) {
			r.Use(my_middleware.Auth(authGrpcClient, 2))

			r.Post("/", handler.TasksCreate)
			r.Get("/{id}", handler.TasksVie)
			r.Delete("/{id}", handler.TasksDelete)
			r.Put("/{id}", handler.TasksUpdate)
		})
	})
	router.Route("/assign", func(r chi.Router) {
		r.Use(my_middleware.Auth(authGrpcClient, 2))

		r.Post("/", handler.AssignAppoint)
		r.Get("/user/{id}", handler.AssignVie)
		r.Delete("/user/{id}", handler.AssignDelete)
	})
	router.Post("/auth", handler.Auth)

	return router
}

func Shutdown(server *http.Server, serverErr <-chan error) {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-done:
		slog.Info("Shutdown")
	case err := <-serverErr:
		slog.Error("Server error", "ERROR", err)
	}

	//SHUTDOWN
	slog.Info("Stopping server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Failed to stop server", "ERROR:", err.Error())
		server.Close()
	}

	slog.Info("Server stopped successfully")
}
