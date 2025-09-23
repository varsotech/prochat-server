package auth

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	prochatv1 "github.com/varsotech/prochat-server/gen/prochat/v1"
	"github.com/varsotech/prochat-server/internal/database/postgres"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type Routes struct {
	PostgresClient *pgxpool.Pool
}

func NewRoutes(postgresClient *pgxpool.Pool) Routes {
	return Routes{PostgresClient: postgresClient}
}

func (o Routes) Login(w http.ResponseWriter, r *http.Request) {
	slog.Info("handling login")
	w.WriteHeader(http.StatusOK)
}

func (o Routes) Register(w http.ResponseWriter, r *http.Request) {
	slog.Info("handling registration")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("error reading request body", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	var req prochatv1.RegisterRequest
	err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(body, &req)
	if err != nil {
		slog.Info("unable to read request body", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	// Force username to be lowercase
	req.Username = strings.ToLower(req.Username)

	if isValid, msg := validateUsername(req.Username); !isValid {
		slog.Info(msg, "error", err, "request_uri", r.RequestURI)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	id, err := uuid.NewV7()
	if err != nil {
		slog.Error("error generating uuid", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	queries := postgres.New(o.PostgresClient)

	// If no email and password provided, anonymously register the user
	if req.Email == "" && req.Password == "" {
		_, err = queries.CreateAnonymousUser(r.Context(), postgres.CreateAnonymousUserParams{ID: id, Username: req.Username})
		if err != nil {
			slog.Error("error creating anonymous user", "error", err, "request_uri", r.RequestURI)
			http.Error(w, "", http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	// Email and password provided, register user
	if isValid, msg := validateEmail(req.Email); !isValid {
		slog.Info(msg, "error", err, "request_uri", r.RequestURI)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	if isValid, msg := validatePassword(req.Password); !isValid {
		slog.Info(msg, "error", err, "request_uri", r.RequestURI)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	argon2idPassword, err := hashPassword(req.Password, nil)
	if err != nil {
		slog.Error("error hashing password", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	_, err = queries.CreateUser(r.Context(), postgres.CreateUserParams{ID: id, Username: req.Username, Email: pgtype.Text{String: req.Email, Valid: true}, PasswordHash: pgtype.Text{String: argon2idPassword, Valid: true}})
	if err != nil {
		slog.Error("error creating anonymous user", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	return
}
