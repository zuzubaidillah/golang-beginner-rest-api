// File: /main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type apiResponse map[string]any

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
	userService := NewUserService(store)
	userHandler := NewUsersHandler(userService)

	mux.HandleFunc("/users", userHandler.HandleUsers)
	mux.HandleFunc("/users/", userHandler.HandleUserRoutes)

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

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("REST server listening on %s", addr)

	// pasang logger middleware untuk semua request
	handler := requestLogger(mux)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
