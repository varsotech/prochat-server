package http

import (
	"errors"
	"github.com/varsotech/prochat-server/internal/auth/service"
	"log/slog"
	"net/http"
)

func writeServiceError(w http.ResponseWriter, err error) {
	var serviceErr service.Error
	if !errors.As(err, &serviceErr) {
		slog.Error("not a service error", "error", err)
		http.Error(w, service.InternalError.ExternalMessage, service.InternalError.HTTPCode)
	}

	// Log internal errors
	if serviceErr.HTTPCode == http.StatusInternalServerError {
		slog.Error("user got internal error: %w", err)
	}

	http.Error(w, serviceErr.ExternalMessage, serviceErr.HTTPCode)
}
