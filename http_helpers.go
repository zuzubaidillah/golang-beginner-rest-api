package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(payload)
}

func errorJSON(w http.ResponseWriter, status int, code string, message string, details any) {
	resp := map[string]any{
		"error":   code,
		"message": message,
	}
	if details != nil {
		resp["details"] = details
	}
	writeJSON(w, status, resp)
}

func readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	// limit 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return err
	}

	// pastikan tidak ada JSON tambahan
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("unexpected extra JSON content")
	}

	return nil
}

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}

	errorJSON(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", map[string]any{
		"method": r.Method,
		"allow":  method,
	})
	return false
}

func requireMethods(w http.ResponseWriter, r *http.Request, allowed ...string) bool {
	for _, m := range allowed {
		if r.Method == m {
			return true
		}
	}

	// header Allow (best practice HTTP)
	w.Header().Set("Allow", strings.Join(allowed, ", "))

	errorJSON(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", map[string]any{
		"method": r.Method,
		"allow":  allowed,
	})
	return false
}
