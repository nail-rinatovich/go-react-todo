package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type Store struct {
	mu     sync.RWMutex
	nextID int
	todos  []Todo
}

func (s *Store) list() []Todo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Todo, len(s.todos))
	copy(out, s.todos)
	return out
}

func (s *Store) create(title string) Todo {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := Todo{ID: s.nextID, Title: title, Completed: false}
	s.nextID++
	s.todos = append(s.todos, t)
	return t
}

func (s *Store) update(id int, title *string, completed *bool) (Todo, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.todos {
		if s.todos[i].ID == id {
			if title != nil {
				s.todos[i].Title = *title
			}
			if completed != nil {
				s.todos[i].Completed = *completed
			}
			return s.todos[i], true
		}
	}
	return Todo{}, false
}

func (s *Store) delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.todos {
		if s.todos[i].ID == id {
			s.todos = append(s.todos[:i], s.todos[i+1:]...)
			return true
		}
	}
	return false
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	store := &Store{nextID: 1, todos: []Todo{}}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/todos", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(store.list())
		case http.MethodPost:
			var body struct {
				Title string `json:"title"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Title) == "" {
				http.Error(w, `{"error":"title is required"}`, http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(store.create(strings.TrimSpace(body.Title)))
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/todos/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		idStr := strings.TrimPrefix(r.URL.Path, "/api/todos/")
		id, err := strconv.Atoi(idStr)
		if err != nil || id < 1 {
			http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodPatch:
			var body struct {
				Title     *string `json:"title"`
				Completed *bool   `json:"completed"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
				return
			}
			todo, ok := store.update(id, body.Title, body.Completed)
			if !ok {
				http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(todo)
		case http.MethodDelete:
			if !store.delete(id) {
				http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           cors(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Println("API http://localhost:8080")
	log.Fatal(srv.ListenAndServe())
}
