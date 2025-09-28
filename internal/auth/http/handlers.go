package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	prochatv1 "github.com/varsotech/prochat-server/internal/models/gen/prochat/v1"
	"github.com/varsotech/prochat-server/internal/pkg/postgres"
	"github.com/varsotech/prochat-server/internal/pkg/redis/authrepo"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

const (
	accessTokenCookieName  = "prochat_accesstoken"
	refreshTokenCookieName = "prochat_refreshtoken"
)

type Handlers struct {
	postgresClient *postgres.Queries
	authRepo       *authrepo.Repo
}

func NewHandlers(postgresConnectionPool *pgxpool.Pool, redisClient *redis.Client) Handlers {
	postgresClient := postgres.New(postgresConnectionPool)
	authRepo := authrepo.New(redisClient)
	return Handlers{postgresClient: postgresClient, authRepo: authRepo}
}

//func (h Handlers) LoginProtectionMiddleware(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		accessTokenCookie, err := r.Cookie(accessTokenCookieName)
//		if errors.Is(err, http.ErrNoCookie) {
//			http.Error(w, "Unauthorized", http.StatusUnauthorized)
//		}
//		if err != nil {
//			slog.Error("error getting access token cookie", "error", err)
//			http.Error(w, "error getting cookie", http.StatusUnauthorized)
//			return
//		}
//
//		userId, found, err := h.authRepo.GetUserIdFromAccessToken(r.Context(), accessTokenCookie.Value)
//		if err != nil {
//			slog.Error("error getting user id from access token", "error", err)
//			http.Error(w, "Internal error", http.StatusInternalServerError)
//			return
//		}
//
//		if !found {
//			http.Error(w, "Unauthorized", http.StatusUnauthorized)
//			return
//		}
//
//		next.ServeHTTP(w, r)
//	})
//}

func (h Handlers) Refresh(w http.ResponseWriter, r *http.Request) {
	slog.Info("handling refresh")

	refreshTokenCookie, err := r.Cookie(refreshTokenCookieName)
	if err != nil {
		slog.Error("error getting refreshToken cookie", "error", err)
		http.Error(w, "error getting cookie", http.StatusUnauthorized)
		return
	}

	userId, found, err := h.authRepo.GetUserIdFromRefreshToken(r.Context(), refreshTokenCookie.Value)
	if err != nil {
		slog.Error("error getting refresh token user id", "error", err)
		http.Error(w, "error getting refresh token user id", http.StatusInternalServerError)
		return
	}

	if !found {
		slog.Error("refresh token not found")
		http.Error(w, "refresh token not found", http.StatusUnauthorized)
		return
	}

	accessToken, refreshToken, err := h.authRepo.RefreshTokenPair(r.Context(), userId)
	if err != nil {
		slog.Error("error refreshing access token", "error", err)
		http.Error(w, "error refreshing access token", http.StatusInternalServerError)
		return
	}

	h.setTokenPairCookies(w, accessToken, refreshToken)
	w.WriteHeader(http.StatusOK)
}

func (h Handlers) Login(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("error reading request body", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	var req prochatv1.LoginRequest
	err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(body, &req)
	if err != nil {
		slog.Info("unable to read request body", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	user, err := h.postgresClient.GetUserByLogin(r.Context(), req.Login)
	if err != nil {
		slog.Info("user not found", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "Invalid credentials", http.StatusNotFound)
		return
	}

	if !user.PasswordHash.Valid {
		// TODO: Test this
		slog.Info("no password found for account", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	passwordsMatch, err := comparePassword(req.Password, user.PasswordHash.String)
	if err != nil {
		slog.Error("failed comparing passwords", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if !passwordsMatch {
		slog.Info("wrong password", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	err = h.issueAndSetTokenPairCookies(r.Context(), w, user.ID)
	if err != nil {
		slog.Error("error generating token pair cookies", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h Handlers) Register(w http.ResponseWriter, r *http.Request) {
	slog.Info("handling registration")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("error reading request body", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	var req prochatv1.RegisterRequest
	err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(body, &req)
	if err != nil {
		slog.Info("unable to read request body", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "Bad request", http.StatusBadRequest)
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
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if req.Email == "" && req.Password == "" {
		// If no email and password provided, anonymously register the user
		_, err = h.postgresClient.CreateAnonymousUser(r.Context(), postgres.CreateAnonymousUserParams{
			ID:       id,
			Username: req.Username,
		})
		if err != nil {
			slog.Error("error creating anonymous user", "error", err, "request_uri", r.RequestURI)
			http.Error(w, "Internal error", http.StatusInternalServerError)
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

		_, err = h.postgresClient.CreateUser(r.Context(), postgres.CreateUserParams{
			ID:           id,
			Username:     req.Username,
			Email:        pgtype.Text{String: req.Email, Valid: true},
			PasswordHash: pgtype.Text{String: argon2idPassword, Valid: true},
		})
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if pgErr.ConstraintName == "users_username_key" {
				http.Error(w, "Username already taken", http.StatusBadRequest)
				return
			}

			if pgErr.ConstraintName == "users_email_key" {
				http.Error(w, "Email already taken", http.StatusBadRequest)
				return
			}

			// Fallback to less specific error, shouldn't reach here
			slog.Error("unknown constraint name, falling back to generic error message", "error", err, "request_uri", r.RequestURI)
			http.Error(w, "Username or email already taken", http.StatusBadRequest)
			return
		}
		if err != nil {
			slog.Error("error creating user", "error", err, "request_uri", r.RequestURI)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}

	err = h.issueAndSetTokenPairCookies(r.Context(), w, id)
	if err != nil {
		slog.Error("error generating token pair cookies", "error", err, "request_uri", r.RequestURI)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h Handlers) issueAndSetTokenPairCookies(ctx context.Context, w http.ResponseWriter, id uuid.UUID) error {
	accessToken, err := h.authRepo.IssueAccessToken(ctx, id)
	if err != nil {
		return fmt.Errorf("failed issuing access token")
	}

	refreshToken, err := h.authRepo.IssueRefreshToken(ctx, id, accessToken)
	if err != nil {
		return fmt.Errorf("failed issuing refresh token")
	}

	h.setTokenPairCookies(w, accessToken, refreshToken)
	return nil
}

func (h Handlers) setTokenPairCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	accessTokenCookie := createCookie(accessTokenCookieName, accessToken, "/", authrepo.AccessTokenMaxAge)
	refreshTokenCookie := createCookie(refreshTokenCookieName, refreshToken, "/api/v1/auth/refresh", authrepo.RefreshTokenMaxAge)

	http.SetCookie(w, &accessTokenCookie)
	http.SetCookie(w, &refreshTokenCookie)
}
