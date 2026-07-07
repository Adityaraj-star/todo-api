package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/Adityaraj-star/todo-api/handler"
	"github.com/Adityaraj-star/todo-api/middleware"
	"github.com/Adityaraj-star/todo-api/store"
)

func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET", "test-secret")
	os.Exit(m.Run())
}


func newTestServer() *httptest.Server {
	todoStore := store.NewMemoryTodoStore()
	userStore := store.NewMemoryUserStore()

	todoHandler := handler.NewTodoHandler(todoStore)
	authHandler := handler.NewAuthHandler(userStore)

	r := chi.NewRouter()
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

	return httptest.NewServer(r)
}

func postJSON(t *testing.T, url string, body any, token string) *http.Response {
	t.Helper()
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return resp
}

func registerAndLogin(t *testing.T, srv *httptest.Server, username string) string {
	t.Helper()

	resp := postJSON(t, srv.URL+"/register", map[string]string{
		"username": username,
		"password": "password123",
	}, "")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register failed, status %d", resp.StatusCode)
	}

	resp = postJSON(t, srv.URL+"/login", map[string]string{
		"username": username,
		"password": "password123",
	}, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login failed, status %d", resp.StatusCode)
	}
	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	if body["token"] == "" {
		t.Fatal("expected a token back from login")
	}
	return body["token"]
}

func TestRegisterDuplicateUsername(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	registerAndLogin(t, srv, "duplicate-user")

	resp := postJSON(t, srv.URL+"/register", map[string]string{
		"username": "duplicate-user",
		"password": "password123",
	}, "")
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409 for duplicate username, got %d", resp.StatusCode)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	registerAndLogin(t, srv, "wrongpass-user")

	resp := postJSON(t, srv.URL+"/login", map[string]string{
		"username": "wrongpass-user",
		"password": "not-the-right-password",
	}, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for wrong password, got %d", resp.StatusCode)
	}
}

func TestTodosRequireAuth(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/todos/")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", resp.StatusCode)
	}
}

func TestTodoCRUDFlow(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	token := registerAndLogin(t, srv, "crud-user")

	resp := postJSON(t, srv.URL+"/todos/", map[string]string{"title": "learn go"}, token)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create failed, status %d", resp.StatusCode)
	}
	var created map[string]any
	json.NewDecoder(resp.Body).Decode(&created)
	id := created["id"].(string)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/todos/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get failed, status %d", resp.StatusCode)
	}

	body, _ := json.Marshal(map[string]string{"title": "learn go properly", "status": "in-progress"})
	req, _ = http.NewRequest(http.MethodPut, srv.URL+"/todos/"+id, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update failed, status %d", resp.StatusCode)
	}

	req, _ = http.NewRequest(http.MethodDelete, srv.URL+"/todos/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete failed, status %d", resp.StatusCode)
	}

	req, _ = http.NewRequest(http.MethodGet, srv.URL+"/todos/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", resp.StatusCode)
	}
}

func TestTodosAreScopedPerUser(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	tokenA := registerAndLogin(t, srv, "user-a")
	tokenB := registerAndLogin(t, srv, "user-b")

	resp := postJSON(t, srv.URL+"/todos/", map[string]string{"title": "user a's todo"}, tokenA)
	var created map[string]any
	json.NewDecoder(resp.Body).Decode(&created)
	id := created["id"].(string)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/todos/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+tokenB)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 when user B fetches user A's todo, got %d", resp.StatusCode)
	}
}