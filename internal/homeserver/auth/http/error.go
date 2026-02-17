package http

import (
	"errors"
	"log/slog"
	"net/http"

	service2 "github.com/varsotech/prochat-server/internal/homeserver/auth/service"
)

func writeServiceError(w http.ResponseWriter, err error) {
	var serviceErr service2.Error
	if !errors.As(err, &serviceErr) {
		slog.Error("not a service error", "error", err)
		http.Error(w, service2.InternalError.ExternalMessage, service2.InternalError.HTTPCode)
	}

	// Log internal errors
	if serviceErr.HTTPCode == http.StatusInternalServerError {
		slog.Error("user got internal error", "error", err)
	}

	http.Error(w, serviceErr.ExternalMessage, serviceErr.HTTPCode)
}
