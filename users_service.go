// File: /users_service.go
package main

import (
	"net/http"
	"strings"
)

type UserService struct {
	store *UserStore
}

func NewUserService(store *UserStore) *UserService {
	return &UserService{store: store}
}

func (s *UserService) CreateUser(name string) (User, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return User{}, &AppError{
			Status:  http.StatusBadRequest,
			Code:    "validation_failed",
			Message: "missing required fields",
			Details: []string{"name is required"},
		}
	}

	u := s.store.Create(name)
	return u, nil
}

func (s *UserService) GetUser(id int) (User, error) {
	u, ok := s.store.Get(id)
	if !ok {
		return User{}, &AppError{
			Status:  http.StatusNotFound,
			Code:    "not_found",
			Message: "resource not found",
		}
	}
	return u, nil
}

func (s *UserService) DeleteUser(id int) error {
	if ok := s.store.Delete(id); !ok {
		return &AppError{
			Status:  http.StatusNotFound,
			Code:    "not_found",
			Message: "resource not found",
		}
	}
	return nil
}

func (s *UserService) ListUsers() []User {
	return s.store.List()
}
