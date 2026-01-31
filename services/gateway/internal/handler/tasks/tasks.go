package tasks_handler

import (
	"context"
	"gateway/internal/grpc/tasks"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/render"
)

func TasksMy(tasksGrpcClient *tasks.Client) http.HandlerFunc { //Посмотреть свои задачи
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
		tasks, err := tasksGrpcClient.TasksMy(ctx, cookie.Value)

		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{
				"Error": "Server error",
			})
			slog.Error("TasksMy error", "ERROR", err.Error())
			return
		}
		if tasks.Tasks == nil {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, map[string]string{
				"Error": "Ваши данные невалидны",
			})
			return
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, map[string]interface{}{
			"Tasks": tasks.Tasks,
		})

	}
}
func TasksCreate(w http.ResponseWriter, r *http.Request) { //Создать задачу
	w.Write([]byte("some data"))
}
func TasksGetAll(w http.ResponseWriter, r *http.Request) { //Получение таблицы со всеми задачами (Id:Header)
	w.Write([]byte("some data"))
}
func TasksVie(w http.ResponseWriter, r *http.Request) { //Получение деталей задачи по id
	w.Write([]byte("some data"))
}
func TasksDelete(w http.ResponseWriter, r *http.Request) { //Удаление задачи по id
	w.Write([]byte("some data"))
}
func TasksUpdate(w http.ResponseWriter, r *http.Request) { //Обновить задачу
	w.Write([]byte("some data"))
}
