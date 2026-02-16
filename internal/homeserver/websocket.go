package homeserver

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/varsotech/prochat-server/internal/oauth"
)

// Upgrader is used to upgrade HTTP connections to WebSocket connections.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (o *Routes) ws(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Info("unable to upgrade websocket", "error", err)
		return
	}
	defer conn.Close()

	slog.Info("new websocket connection")

	// First message must be the authorization header
	_, authorizationHeader, err := conn.ReadMessage()
	if err != nil {
		slog.Info("unable to read websocket authorization header", "error", err)
		return
	}

	slog.Info("authorizationHeader", "authorizationHeader", authorizationHeader)

	// TODO: Should r.Context() be used here?
	_, err = o.authorizer.Authorize(r.Context(), string(authorizationHeader))
	if errors.Is(err, oauth.UnauthorizedError) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if err != nil {
		slog.Error("failed to authorize request", "error", err)
		return
	}

	slog.Info("websocket connection is authorized")

	// Listen for incoming messages
	for {
		// Read message from the client
		_, message, err := conn.ReadMessage()
		if err != nil {
			slog.Error("failed to read websocket message", "error", err)
			break
		}
		fmt.Printf("Received: %s\n", message)
		// Echo the message back to the client
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			slog.Error("failed to write websocket essage", "error", err)
			break
		}
	}
}
