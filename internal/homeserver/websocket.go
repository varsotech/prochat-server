package homeserver

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/varsotech/prochat-server/internal/homeserver/oauth"
	homeserverv1 "github.com/varsotech/prochat-server/internal/models/gen/homeserver/v1"
	"google.golang.org/protobuf/proto"
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

	auth, err := o.authorizer.Authorize(r.Context(), string(authorizationHeader))
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

		var homeserverMessage homeserverv1.Message
		err = proto.Unmarshal(message, &homeserverMessage)
		if err != nil {
			slog.Info("failed to unmarshal homeserver message", "error", err)
			continue
		}

		response := o.handlers.Handle(r.Context(), auth, &homeserverMessage)

		responseData, err := proto.Marshal(response)
		if err != nil {
			slog.Info("failed to marshal response", "error", err)
			continue
		}

		slog.Info("websocket sending response message", "message", string(responseData))

		if err := conn.WriteMessage(websocket.BinaryMessage, responseData); err != nil {
			slog.Error("failed to write websocket message", "error", err)
			break
		}
	}
}
