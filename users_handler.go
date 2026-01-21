// File: /users_handler.go
package main

import (
	"net/http"
	"strconv"
	"strings"
)

type UsersHandler struct {
	svc *UserService
}

func NewUsersHandler(svc *UserService) *UsersHandler {
	return &UsersHandler{svc: svc}
}

// /users -> GET list, POST create
func (h *UsersHandler) HandleUsers(w http.ResponseWriter, r *http.Request) {
	if !requireMethods(w, r, http.MethodGet, http.MethodPost) {
		return
	}

	switch r.Method {
	case http.MethodGet:
		users := h.svc.ListUsers()
		writeJSON(w, http.StatusOK, apiResponse{
			"items": users,
			"count": len(users),
		})
		return

	case http.MethodPost:
		type createUserRequest struct {
			Name string `json:"name"`
		}
		var req createUserRequest

		if err := readJSON(w, r, &req); err != nil {
			errorJSON(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
			return
		}

		u, err := h.svc.CreateUser(req.Name)
		if err != nil {
			h.writeAppError(w, err)
			return
		}

		writeJSON(w, http.StatusCreated, u)
		return
	}
}

// /users/{id}, /users/{id}/profile, /users/{id}/orders/{orderId}
func (h *UsersHandler) HandleUserRoutes(w http.ResponseWriter, r *http.Request) {
	const prefix = "/users/"
	path := r.URL.Path
	if len(path) <= len(prefix) {
		errorJSON(w, http.StatusBadRequest, "invalid_path", "user id is required", nil)
		return
	}

	rest := strings.Trim(path[len(prefix):], "/")
	parts := strings.Split(rest, "/")

	id, err := parsePositiveInt(parts[0])
	if err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid_path", "user id must be a positive integer", nil)
		return
	}

	// /users/{id}
	if len(parts) == 1 {
		if !requireMethods(w, r, http.MethodGet, http.MethodDelete) {
			return
		}

		switch r.Method {
		case http.MethodGet:
			u, err := h.svc.GetUser(id)
			if err != nil {
				h.writeAppError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, u)
			return

		case http.MethodDelete:
			if err := h.svc.DeleteUser(id); err != nil {
				h.writeAppError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, apiResponse{
				"deleted": true,
				"id":      id,
			})
			return
		}
	}

	// /users/{id}/profile
	if len(parts) == 2 && parts[1] == "profile" {
		if !requireMethods(w, r, http.MethodGet) {
			return
		}

		// pastikan user ada
		if _, err := h.svc.GetUser(id); err != nil {
			h.writeAppError(w, err)
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
		if !requireMethods(w, r, http.MethodGet) {
			return
		}

		orderId := strings.TrimSpace(parts[2])
		if orderId == "" {
			errorJSON(w, http.StatusBadRequest, "invalid_path", "orderId is required", nil)
			return
		}

		// pastikan user ada
		if _, err := h.svc.GetUser(id); err != nil {
			h.writeAppError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, apiResponse{
			"id":      id,
			"orderId": orderId,
		})
		return
	}

	errorJSON(w, http.StatusNotFound, "not_found", "resource not found", apiResponse{
		"path": r.URL.Path,
	})
}

func (h *UsersHandler) writeAppError(w http.ResponseWriter, err error) {
	if ae, ok := err.(*AppError); ok {
		errorJSON(w, ae.Status, ae.Code, ae.Message, ae.Details)
		return
	}
	errorJSON(w, http.StatusInternalServerError, "internal_error", "unexpected error", nil)
}

func parsePositiveInt(s string) (int, error) {
	s = strings.TrimSpace(s)
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return 0, strconv.ErrSyntax
	}
	return n, nil
}
