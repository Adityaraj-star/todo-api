package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Adityaraj-star/todo-api/middleware"
	"github.com/Adityaraj-star/todo-api/model"
	"github.com/Adityaraj-star/todo-api/store"
)

type TodoHandler struct {
	store store.TodoStore
}

func NewTodoHandler(s store.TodoStore) *TodoHandler {
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

func userID(r *http.Request) (string, bool) {
	return middleware.UserIDFromContext(r.Context())
}

func (h *TodoHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	q := r.URL.Query()
	params := store.ListParams{
		Status: q.Get("status"),
	}
	if limit, err := strconv.Atoi(q.Get("limit")); err == nil {
		params.Limit = limit
	}
	if offset, err := strconv.Atoi(q.Get("offset")); err == nil {
		params.Offset = offset
	}

	todos, err := h.store.GetAll(uid, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch todos")
		return
	}
	writeJSON(w, http.StatusOK, todos)
}

func (h *TodoHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid, ok := userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")

	todo, err := h.store.GetByID(id, uid)
	if err == store.ErrNotFound {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch todo")
		return
	}
	writeJSON(w, http.StatusOK, todo)
}

func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, ok := userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	todo := model.Todo{
		ID:        uuid.NewString(),
		UserID:    uid,
		Title:     req.Title,
		Status:    model.StatusTodo,
		CreatedAt: time.Now(),
	}

	created, err := h.store.Create(todo)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create todo")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {
	uid, ok := userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")

	var req struct {
		Title  string `json:"title"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	existing, err := h.store.GetByID(id, uid)
	if err == store.ErrNotFound {
		writeError(w, http.StatusNotFound, "todo not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch todo")
		return
	}

	todo := model.Todo{
		ID:        id,
		UserID:    uid,
		Title:     req.Title,
		Status:    req.Status,
		CreatedAt: existing.CreatedAt,
	}

	updated, err := h.store.Update(todo)
	if err == store.ErrNotFound {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update todo")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, ok := userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")

	err := h.store.Delete(id, uid)
	if err == store.ErrNotFound {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete todo")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}