package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/varsotech/prochat-server/internal/auth/service"
)

func writeServiceError(w http.ResponseWriter, err error) {
	var serviceErr service.Error
	if !errors.As(err, &serviceErr) {
		slog.Error("not a service error", "error", err)
		http.Error(w, service.InternalError.ExternalMessage, service.InternalError.HTTPCode)
	}

	// Log internal errors
	if serviceErr.HTTPCode == http.StatusInternalServerError {
		slog.Error("user got internal error", "error", err)
	}

	http.Error(w, serviceErr.ExternalMessage, serviceErr.HTTPCode)
}
