package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Dependency struct {
	Name  string
	Check func(context.Context) error
}

type healthHandler struct {
	dependencies []Dependency
}

func newHealthHandler(dependencies []Dependency) healthHandler {
	return healthHandler{dependencies: dependencies}
}

func (h healthHandler) live(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (h healthHandler) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	unavailable := make([]string, 0)
	for _, dependency := range h.dependencies {
		if err := dependency.Check(ctx); err != nil {
			unavailable = append(unavailable, dependency.Name)
		}
	}
	if len(unavailable) > 0 {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status":       "unavailable",
			"dependencies": unavailable,
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
