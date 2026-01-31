package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gateway/internal/grpc/auth"
	"gateway/internal/grpc/tasks"
	"gateway/internal/grpc/users"
	asignment_handler "gateway/internal/handler/asignment"
	auth_handler "gateway/internal/handler/auth"
	tasks_handler "gateway/internal/handler/tasks"
	users_handler "gateway/internal/handler/users"
	my_middleware "gateway/internal/middlewares"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(authGrpcClient *auth.Client, usersGrpcClient *users.Client, tasksGrpcClient *tasks.Client) *chi.Mux {
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

	router.Route("/users", func(rout chi.Router) {
		rout.Get("/", users_handler.UsersMy(usersGrpcClient))

		rout.Group(func(r chi.Router) {
			r.Use(my_middleware.Auth(authGrpcClient, 2))

			r.Post("/", users_handler.UsersCreate(usersGrpcClient))
			r.Get("/all", users_handler.UsersGetAll(usersGrpcClient))
			r.Get("/{id}", users_handler.UsersView(usersGrpcClient))
			r.Delete("/{id}", users_handler.UsersDelete(usersGrpcClient))
		})
	})
	router.Route("/tasks", func(rout chi.Router) {
		rout.Get("/", tasks_handler.TasksMy(tasksGrpcClient))

		rout.Group(func(r chi.Router) {
			r.Use(my_middleware.Auth(authGrpcClient, 2))

			r.Post("/", tasks_handler.TasksCreate)
			r.Get("/all", tasks_handler.TasksGetAll)
			r.Get("/{id}", tasks_handler.TasksVie)
			r.Delete("/{id}", tasks_handler.TasksDelete)
			r.Put("/{id}", tasks_handler.TasksUpdate)
		})
	})
	router.Route("/assign", func(r chi.Router) {
		r.Use(my_middleware.Auth(authGrpcClient, 2))

		r.Post("/", asignment_handler.AssignAppoint)
		r.Get("/user/{id}", asignment_handler.AssignVie)
		r.Delete("/user/{id}", asignment_handler.AssignDelete)
	})
	router.Post("/auth", auth_handler.Auth(authGrpcClient))

	return router
}

func Shutdown(server *http.Server, serverErr chan error) {
	defer close(serverErr)
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-done:
		slog.Info("Shutdown")
	case err := <-serverErr:
		slog.Error("Server error", "ERROR", err)
	}

	//SHUTDOWN
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Failed to stop server", "ERROR:", err.Error())
		server.Close()
	}

	slog.Info("Server stopped successfully")
}
