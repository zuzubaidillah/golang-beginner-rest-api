// File: /main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type apiResponse map[string]any

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type UserStore struct {
	mu     sync.RWMutex
	nextID int
	items  map[int]User
}

func NewUserStore() *UserStore {
	return &UserStore{
		nextID: 1,
		items:  make(map[int]User),
	}
}

func (s *UserStore) Create(name string) User {
	s.mu.Lock()
	defer s.mu.Unlock()

	u := User{
		ID:        s.nextID,
		Name:      name,
		CreatedAt: time.Now().UTC(),
	}
	s.items[u.ID] = u
	s.nextID++
	return u
}

func (s *UserStore) Get(id int) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.items[id]
	return u, ok
}

func (s *UserStore) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.items[id]; !ok {
		return false
	}
	delete(s.items, id)
	return true
}

func (s *UserStore) List() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]User, 0, len(s.items))
	for _, u := range s.items {
		out = append(out, u)
	}
	return out
}

// middleware logger (simple, beginner friendly)
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("IN  %s %s", r.Method, r.URL.RequestURI())

		next.ServeHTTP(w, r)

		log.Printf("OUT %s %s (%s)", r.Method, r.URL.RequestURI(), time.Since(start))
	})
}

func main() {
	port := flag.Int("port", 8080, "HTTP port for REST server")
	flag.Parse()

	mux := http.NewServeMux()

	store := NewUserStore()

	// GET /
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			writeJSON(w, http.StatusNotFound, apiResponse{
				"error": "not_found",
				"path":  r.URL.Path,
			})
			return
		}

		writeJSON(w, http.StatusOK, apiResponse{
			"service": "golang-beginner-rest",
			"routes": []string{
				"GET /health",
				"GET /time",
			},
		})
	})

	// GET /health
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if !requireMethod(w, r, http.MethodPost) {
			return
		}

		writeJSON(w, http.StatusOK, apiResponse{
			"status": "ok",
		})
	})

	// GET /time
	mux.HandleFunc("/time", func(w http.ResponseWriter, r *http.Request) {
		if !requireMethod(w, r, http.MethodPost) {
			return
		}

		writeJSON(w, http.StatusOK, apiResponse{
			"time": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// GET echo with query params
	mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		if !requireMethod(w, r, http.MethodPost) {
			return
		}

		qName := strings.TrimSpace(r.URL.Query().Get("name"))

		if qName == "" {
			writeJSON(w, http.StatusBadRequest, apiResponse{
				"error": "name_required",
				"path":  r.URL.Path,
			})
			return
		}

		writeJSON(w, http.StatusOK, apiResponse{
			"name": qName,
		})

	})

	// POST /sum
	// File: /main.go (di dalam main())
	mux.HandleFunc("/sum", func(w http.ResponseWriter, r *http.Request) {
		if !requireMethod(w, r, http.MethodPost) {
			return
		}

		type sumRequest struct {
			A *int `json:"a"`
			B *int `json:"b"`
		}
		var req sumRequest

		if err := readJSON(w, r, &req); err != nil {
			errorJSON(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
			return
		}

		if req.A == nil || req.B == nil {
			errorJSON(w, http.StatusBadRequest, "validation_failed", "missing required fields", []string{
				"a is required",
				"b is required",
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"result": *req.A + *req.B,
		})
	})

	// POST /mul
	// File: /main.go (di dalam main())
	mux.HandleFunc("/mul", func(w http.ResponseWriter, r *http.Request) {
		if !requireMethod(w, r, http.MethodPost) {
			return
		}

		type mulRequest struct {
			A *int `json:"a"`
			B *int `json:"b"`
		}
		var req mulRequest

		if err := readJSON(w, r, &req); err != nil {
			errorJSON(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
			return
		}

		if req.A == nil || req.B == nil {
			errorJSON(w, http.StatusBadRequest, "validation_failed", "missing required fields", []string{
				"a is required",
				"b is required",
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"result": *req.A * *req.B,
		})
	})

	// GET /users/{id} and GET /users/{id}/profile
	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		const prefix = "/users/"
		path := r.URL.Path
		if len(path) <= len(prefix) {
			errorJSON(w, http.StatusBadRequest, "invalid_path", "user id is required", nil)
			return
		}

		rest := strings.Trim(path[len(prefix):], "/")
		parts := strings.Split(rest, "/")

		idStr := strings.TrimSpace(parts[0])
		if idStr == "" {
			errorJSON(w, http.StatusBadRequest, "invalid_path", "user id is required", nil)
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			errorJSON(w, http.StatusBadRequest, "invalid_path", "user id must be a positive integer", nil)
			return
		}

		// /users/{id}
		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				u, ok := store.Get(id)
				if !ok {
					errorJSON(w, http.StatusNotFound, "invalid_path", "not_found", nil)
					return
				}
				writeJSON(w, http.StatusOK, u)
				return

			case http.MethodDelete:
				if ok := store.Delete(id); !ok {
					errorJSON(w, http.StatusNotFound, "invalid_path", "not_found", nil)
					return
				}
				writeJSON(w, http.StatusOK, apiResponse{
					"deleted": true,
					"id":      id,
				})
				return

			default:
				if !requireMethod(w, r, http.MethodPost) {
					return
				}
				return
			}
		}

		// /users/{id}/profile
		if len(parts) == 2 && parts[1] == "profile" {
			if !requireMethod(w, r, http.MethodPost) {
				return
			}

			// optional: pastikan user ada
			if _, ok := store.Get(id); !ok {
				writeJSON(w, http.StatusNotFound, apiResponse{"error": "not_found"})
				return
			}

			writeJSON(w, http.StatusOK, apiResponse{
				"id":      id,
				"profile": true,
			})
			return
		}

		// /users/{id}/orders/{orderId}
		if len(parts) == 3 && parts[1] == "orders" {
			if r.Method != http.MethodGet {
				writeJSON(w, http.StatusMethodNotAllowed, apiResponse{
					"error":  "method_not_allowed",
					"method": r.Method,
				})
				return
			}

			orderId := strings.TrimSpace(parts[2])
			if orderId == "" {
				writeJSON(w, http.StatusBadRequest, apiResponse{
					"error":   "invalid_path",
					"message": "orderId is required",
				})
				return
			}

			// optional: pastikan user ada
			if _, ok := store.Get(id); !ok {
				writeJSON(w, http.StatusNotFound, apiResponse{"error": "not_found"})
				return
			}

			writeJSON(w, http.StatusOK, apiResponse{
				"id":      id,
				"orderId": orderId,
			})
			return
		}

		writeJSON(w, http.StatusNotFound, apiResponse{
			"error": "not_found",
			"path":  r.URL.Path,
		})
	})

	// POST /users (create) and GET /users (list)
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		// File: /main.go
		case http.MethodPost:
			type createUserRequest struct {
				Name string `json:"name"`
			}

			var req createUserRequest
			if err := readJSON(w, r, &req); err != nil {
				errorJSON(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
				return
			}

			name := strings.TrimSpace(req.Name)
			if name == "" {
				errorJSON(w, http.StatusBadRequest, "validation_failed", "missing required fields", []string{
					"name is required",
				})
				return
			}

			u := store.Create(name)
			writeJSON(w, http.StatusCreated, u)
			return

		case http.MethodGet:
			users := store.List()
			writeJSON(w, http.StatusOK, apiResponse{
				"items": users,
				"count": len(users),
			})
			return

		default:
			if !requireMethod(w, r, http.MethodPost) {
				return
			}
			return
		}
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("REST server listening on %s", addr)

	// pasang logger middleware untuk semua request
	handler := requestLogger(mux)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
