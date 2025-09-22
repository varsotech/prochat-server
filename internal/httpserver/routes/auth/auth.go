package auth

import (
	"github.com/jackc/pgx/v5/pgxpool"
	prochatv1 "github.com/varsotech/prochat-server/gen/prochat/v1"
	"github.com/varsotech/prochat-server/internal/database/postgres"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type Routes struct {
	PostgresClient *pgxpool.Pool
}

func NewRoutes(postgresClient *pgxpool.Pool) Routes {
	return Routes{PostgresClient: postgresClient}
}

func (o Routes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/test", func(writer http.ResponseWriter, request *http.Request) {
		select {
		case <-request.Context().Done():
			slog.Info("context done", "error", request.Context().Err())
			return
		case <-time.After(30 * time.Second):
			slog.Info("timeout")
			return
		}
	})
	mux.HandleFunc("POST /api/v1/auth/login", o.login)
	mux.HandleFunc("POST /api/v1/auth/register", o.register)
}

func (o Routes) login(w http.ResponseWriter, r *http.Request) {
	slog.Info("handling login")
	w.WriteHeader(http.StatusOK)
}

func (o Routes) register(w http.ResponseWriter, r *http.Request) {
	slog.Info("handling registration")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("error reading request body", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	var req prochatv1.RegisterRequest
	err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(body, &req)
	if err != nil {
		slog.Error("error reading request body", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "", http.StatusBadRequest)
	}

	queries := postgres.New(o.PostgresClient)
	
	// TODO for PR: Implement snowflake or revert to serialized ID
	_, err = queries.CreateAnonymousUser(r.Context(), postgres.CreateAnonymousUserParams{ID: , Username: })
	if err != nil {
		slog.Error("error creating anonymous user", "error", err)
		http.Error(w, "", http.StatusInternalServerError)
	}
	
	w.WriteHeader(http.StatusOK)
}
