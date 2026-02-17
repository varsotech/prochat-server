package identity

import (
	"log/slog"
	"net/http"

	homeserverv1 "github.com/varsotech/prochat-server/internal/models/gen/homeserver/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

func (s *Routes) wellKnown(w http.ResponseWriter, r *http.Request) {
	wellKnown := homeserverv1.WellKnown{
		PublicKey: s.publicKey,
	}

	data, err := protojson.Marshal(&wellKnown)
	if err != nil {
		slog.Error("failed to marshal well known response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		slog.Info("failed to write wellknown response", "error", err)
		return
	}
}
