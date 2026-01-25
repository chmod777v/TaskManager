package asignment_handler

import "net/http"

func AssignAppoint(w http.ResponseWriter, r *http.Request) { //Назначить задачу разработчику
	w.Write([]byte("some data"))
}
func AssignVie(w http.ResponseWriter, r *http.Request) { //Задачи конкретного разработчика
	w.Write([]byte("some data"))
}
func AssignDelete(w http.ResponseWriter, r *http.Request) { //Удаление задачи разработчика по его id
	w.Write([]byte("some data"))
}
