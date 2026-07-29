package api

import (
	"context"
	"log"
	"net/http"
	"time"
)

type HealthChecker interface {
	Ping(context.Context) error
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleReady(checker HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if checker == nil {
			writeError(w, http.StatusServiceUnavailable, "not_ready")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := checker.Ping(ctx); err != nil {
			log.Printf("readiness database check failed: %v", err)
			writeError(w, http.StatusServiceUnavailable, "not_ready")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	}
}
