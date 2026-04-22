package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Adityaraj-star/todo-api/model"
	"github.com/Adityaraj-star/todo-api/store"
	"github.com/google/uuid"
)

type TodoHandler struct {
	store *store.TodoStore
}

func NewTodoHandler(s *store.TodoStore) *TodoHandler {
	return &TodoHandler{
		store: s,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func (h *TodoHandler) List(w http.ResponseWriter, r *http.Request) {
	todos := h.store.GetAll()
	writeJSON(w, http.StatusOK, todos)
}

func (h *TodoHandler) Get(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path // "/todos/123"
	parts := strings.Split(path, "/")
	id := parts[2]

	todo, err := h.store.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, todo)
}

func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	todo := model.Todo{
		ID:        uuid.NewString(),
		Title:     req.Title,
		Status:    model.StatusTodo,
		CreatedAt: time.Now(),
	}

	created := h.store.Create(todo)
	writeJSON(w, http.StatusCreated, created)
}

func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	id := parts[2]

	var req struct {
		Title  string `json:"title"`
		Status string `json:"status"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	existing, err := h.store.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "todo not found")
		return
	}

	todo := model.Todo{
		ID:        id,
		Title:     req.Title,
		Status:    req.Status,
		CreatedAt: existing.CreatedAt,
	}

	updated, err := h.store.Update(todo)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	id := parts[2]

	err := h.store.Delete(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
