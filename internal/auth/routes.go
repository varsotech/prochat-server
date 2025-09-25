package auth

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	prochatv1 "github.com/varsotech/prochat-server/gen/prochat/v1"
	"github.com/varsotech/prochat-server/internal/auth/internal/authrepo"
	"github.com/varsotech/prochat-server/internal/database/postgres"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type Routes struct {
	postgresClient *pgxpool.Pool
	authRepo       *authrepo.Repo
}

func NewRoutes(postgresClient *pgxpool.Pool, redisClient *redis.Client) Routes {
	authRepo := authrepo.New(redisClient)
	return Routes{postgresClient: postgresClient, authRepo: authRepo}
}

func (o Routes) Login(w http.ResponseWriter, r *http.Request) {
	slog.Info("handling login")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("error reading request body", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	var req prochatv1.LoginRequest
	err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(body, &req)
	if err != nil {
		slog.Info("unable to read request body", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	queries := postgres.New(o.postgresClient)

	user, err := queries.GetUserByLogin(r.Context(), req.Login)
	if err != nil {
		slog.Info("user not found", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "", http.StatusNotFound)
		return
	}

	if !user.PasswordHash.Valid {
		// TODO: Test this
		slog.Info("no password found for account", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	passwordsMatch, err := comparePassword(req.Password, user.PasswordHash.String)
	if err != nil {
		slog.Error("failed comparing passwords", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if !passwordsMatch {
		slog.Info("wrong password", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	refreshToken, err := o.authRepo.CreateRefreshToken(r.Context(), user.ID)
	if err != nil {
		slog.Error("error creating refresh token", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	accessToken, err := o.authRepo.CreateAccessToken(r.Context(), user.ID)
	if err != nil {
		slog.Error("error creating access token", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	response := prochatv1.LoginResponse{
		RefreshToken: refreshToken.String(),
		AccessToken:  accessToken.String(),
	}

	responseString, err := protojson.Marshal(&response)
	if err != nil {
		slog.Error("error marshaling response", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(responseString)
	if err != nil {
		slog.Error("error writing response", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

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

	// TODO: Multiple attempts in case of collisions. Unlikely at the moment
	if req.Username == "" {
		req.Username = generateUsername()
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

	queries := postgres.New(o.postgresClient)

	if req.Email == "" && req.Password == "" {
		// If no email and password provided, anonymously register the user
		_, err = queries.CreateAnonymousUser(r.Context(), postgres.CreateAnonymousUserParams{
			ID:       id,
			Username: req.Username,
		})
		if err != nil {
			slog.Error("error creating anonymous user", "error", err, "request_uri", r.RequestURI)
			http.Error(w, "", http.StatusInternalServerError)
		}
	} else {
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

		_, err = queries.CreateUser(r.Context(), postgres.CreateUserParams{
			ID: id, Username: req.Username,
			Email:        pgtype.Text{String: req.Email, Valid: true},
			PasswordHash: pgtype.Text{String: argon2idPassword, Valid: true},
		})
		if err != nil {
			slog.Error("error creating anonymous user", "error", err, "request_uri", r.RequestURI)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}

	refreshToken, err := o.authRepo.CreateRefreshToken(r.Context(), id)
	if err != nil {
		slog.Error("error creating refresh token", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	accessToken, err := o.authRepo.CreateAccessToken(r.Context(), id)
	if err != nil {
		slog.Error("error creating access token", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	response := prochatv1.RegisterResponse{
		RefreshToken: refreshToken.String(),
		AccessToken:  accessToken.String(),
	}

	responseString, err := protojson.Marshal(&response)
	if err != nil {
		slog.Error("error marshaling response", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(responseString)
	if err != nil {
		slog.Error("error writing response", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}
