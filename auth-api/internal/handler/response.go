package handler

import (
	"auth-api/internal/domain"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
}

type successResponse struct {
	Data any `json:"data"`
}

// writeJSON sends a JSON response with the given status code
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(successResponse{Data: data})
}

// writeError inspects the error type and sends the right HTTP response
func writeError(w http.ResponseWriter, err error) {
	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		// log internal detail if present — never send it to client
		if appErr.Err != nil {
			slog.Error("request error", "err", appErr.Err, "code", appErr.Code)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(appErr.Code)
		json.NewEncoder(w).Encode(errorResponse{Error: appErr.Message})
		return
	}

	// completely unexpected error
	slog.Error("unexpected error", "err", err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(errorResponse{Error: "internal server error"})
}
