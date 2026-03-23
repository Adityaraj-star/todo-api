package main

import (
	"log"
	"net/http"

	"github.com/Adityaraj-star/todo-api/handler"
	"github.com/Adityaraj-star/todo-api/middleware"
	"github.com/Adityaraj-star/todo-api/store"
)

func main() {
	todoStore := store.NewTodoStore()
	todoHandler := handler.NewTodoHandler(todoStore)

	mux := http.NewServeMux()

	mux.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			todoHandler.List(w, r)
		case http.MethodPost:
			todoHandler.Create(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/todos/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			todoHandler.Get(w, r)
		case http.MethodPut:
			todoHandler.Update(w, r)
		case http.MethodDelete:
			todoHandler.Delete(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	loggedMux := middleware.Logger(mux)

	log.Println("server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", loggedMux))
}
