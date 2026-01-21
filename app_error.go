// File: /app_error.go
package main

type AppError struct {
	Status  int
	Code    string
	Message string
	Details any
}

func (e *AppError) Error() string {
	return e.Code + ": " + e.Message
}
