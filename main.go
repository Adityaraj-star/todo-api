package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/Adityaraj-star/todo-api/auth"
	"github.com/Adityaraj-star/todo-api/db"
	"github.com/Adityaraj-star/todo-api/handler"
	"github.com/Adityaraj-star/todo-api/middleware"
	"github.com/Adityaraj-star/todo-api/store"
)

func main() {
	auth.RequireJWTSecret() 
	
	var todoStore store.TodoStore
	var userStore store.UserStore


	if os.Getenv("STORE") == "memory" {
		log.Println("using in-memory store")
		todoStore = store.NewMemoryTodoStore()
		userStore = store.NewMemoryUserStore()
	} else {
		conn, err := db.Connect()
		if err != nil {
			log.Fatalf("failed to connect to db: %v", err)
		}
		defer conn.Close()

		if err := db.Migrate(conn); err != nil {
			log.Fatalf("failed to run migrations: %v", err)
		}

		log.Println("using postgres store")
		todoStore = store.NewPostgresTodoStore(conn)
		userStore = store.NewPostgresUserStore(conn)
	}
	
	todoHandler := handler.NewTodoHandler(todoStore)
	authHandler := handler.NewAuthHandler(userStore)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/register", authHandler.Register)
	r.Post("/login", authHandler.Login)

	r.Route("/todos", func(r chi.Router) {
		r.Use(middleware.RequireAuth)
		r.Get("/", todoHandler.List)
		r.Post("/", todoHandler.Create)
		r.Get("/{id}", todoHandler.Get)
		r.Put("/{id}", todoHandler.Update)
		r.Delete("/{id}", todoHandler.Delete)
	})

	log.Println("server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
