package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/karabas/yakamoz/internal/uuid"
)

type APIError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func WriteError(w http.ResponseWriter, status int, code, message string, details map[string]any) {
	WriteJSON(w, status, APIError{Code: code, Message: message, Details: details})
}

func DecodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func RequestID(r *http.Request) string {
	if value := r.Header.Get("X-Request-ID"); value != "" {
		return value
	}
	return uuid.NewString()
}

func Middleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := RequestID(r)
		w.Header().Set("X-Request-ID", requestID)
		start := time.Now()
		defer func() {
			if recovered := recover(); recovered != nil {
				WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error", nil)
			}
			logger.Info("http request", "request_id", requestID, "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
		}()
		next.ServeHTTP(w, r)
	})
}
