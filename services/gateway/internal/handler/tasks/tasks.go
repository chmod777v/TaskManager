package tasks_handler

import "net/http"

func TasksMy(w http.ResponseWriter, r *http.Request) { //Посмотреть свои задачи
	w.Write([]byte("some data"))
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
