package tasks_handler

import "net/http"

func TasksCreate(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("some data"))
}
func TasksVie(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("some data"))
}
func TasksDelete(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("some data"))
}
func TasksUpdate(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("some data"))
}
func TasksMy(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("some data"))
}
