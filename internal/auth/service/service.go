package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/varsotech/prochat-server/internal/auth/internal/argon2"
	"github.com/varsotech/prochat-server/internal/auth/internal/authrepo"
	"github.com/varsotech/prochat-server/internal/auth/internal/username"
	"github.com/varsotech/prochat-server/internal/pkg/postgres"
	"log/slog"
	"net/http"
)

var InternalError = Error{ExternalMessage: "Internal error", HTTPCode: http.StatusInternalServerError}
var UnauthorizedError = Error{ExternalMessage: "Unauthorized", HTTPCode: http.StatusUnauthorized}
var IncorrectCredentialsError = Error{ExternalMessage: "Incorrect credentials", HTTPCode: http.StatusUnauthorized}
var LoginNotProvided = Error{ExternalMessage: "Email or username not provided", HTTPCode: http.StatusBadRequest}
var EmailNotProvided = Error{ExternalMessage: "Email not provided", HTTPCode: http.StatusBadRequest}
var PasswordNotProvided = Error{ExternalMessage: "Password not provided", HTTPCode: http.StatusBadRequest}
var UsernameOrEmailTakenError = Error{ExternalMessage: "Username or email already taken", HTTPCode: http.StatusConflict}
var EmailTakenError = Error{ExternalMessage: "Email already taken", HTTPCode: http.StatusConflict}
var UsernameTakenError = Error{ExternalMessage: "Username already taken", HTTPCode: http.StatusConflict}

type Service struct {
	postgresClient *postgres.Queries
	authRepo       *authrepo.Repo
}

func New(pgClient *pgxpool.Pool, redisClient *redis.Client) *Service {
	return &Service{
		postgresClient: postgres.New(pgClient),
		authRepo:       authrepo.New(redisClient),
	}
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

type RefreshResult struct {
	AccessToken  string
	RefreshToken string
}

func (h Service) Refresh(ctx context.Context, refreshToken string) (RefreshResult, error) {
	refreshTokenPairResult, err := h.authRepo.RefreshTokenPair(ctx, refreshToken)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("failed refresh token pair: %w: %w", InternalError, err)
	}

	return RefreshResult{
		AccessToken:  refreshTokenPairResult.AccessToken,
		RefreshToken: refreshTokenPairResult.RefreshToken,
	}, nil
}

type LoginParams struct {
	Login    string
	Password string
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
}

func (h Service) Login(ctx context.Context, params LoginParams) (LoginResult, error) {
	if params.Login == "" {
		return LoginResult{}, LoginNotProvided
	}
	if params.Password == "" {
		return LoginResult{}, PasswordNotProvided
	}

	user, err := h.postgresClient.GetUserByLogin(ctx, params.Login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return LoginResult{}, fmt.Errorf("failed getting user by login: %w: %w", IncorrectCredentialsError, err)
		}
		slog.Error("failed to get user by login", "error", err)
		return LoginResult{}, fmt.Errorf("failed getting user by login: %w: %w", InternalError, err)
	}

	if !user.PasswordHash.Valid {
		return LoginResult{}, fmt.Errorf("no password set for user: %w", IncorrectCredentialsError)
	}

	passwordsMatch, err := argon2.Compare(params.Password, user.PasswordHash.String)
	if err != nil {
		return LoginResult{}, fmt.Errorf("failed comparing passwords for user: %w: %w", InternalError, err)
	}

	if !passwordsMatch {
		return LoginResult{}, fmt.Errorf("incorrect password: %w", IncorrectCredentialsError)
	}

	issueTokenPairResult, err := h.authRepo.IssueTokenPair(ctx, user.ID)
	if err != nil {
		return LoginResult{}, fmt.Errorf("failed issuing token pair for login: %w: %w", InternalError, err)
	}

	return LoginResult{
		AccessToken:  issueTokenPairResult.AccessToken,
		RefreshToken: issueTokenPairResult.RefreshToken,
	}, nil
}

type RegisterParams struct {
	DisplayName DisplayName
	Username    *Username
	Email       *Email
	Password    *Password
}

type RegisterResult struct {
	AccessToken  string
	RefreshToken string
}

func (h Service) Register(ctx context.Context, params RegisterParams) (RegisterResult, error) {
	// TODO: Multiple attempts in case of collisions. Unlikely at the moment
	if params.Username == nil {
		usernameStr := username.Generate()
		userName, err := NewUsername(usernameStr)
		if err != nil {
			slog.Error("failed generating valid username", "error", err)
			return RegisterResult{}, fmt.Errorf("failed generating valid username: %w", InternalError)
		}
		params.Username = &userName
	}

	id, err := uuid.NewV7()
	if err != nil {
		return RegisterResult{}, fmt.Errorf("failed generating valid uuidv7: %w: %w", InternalError, err)
	}

	if params.Email == nil && params.Password == nil {
		// If both email and password weren't provided, anonymously register the user
		_, err = h.postgresClient.CreateAnonymousUser(ctx, postgres.CreateAnonymousUserParams{
			ID:       id,
			Username: string(*params.Username),
		})
		if err != nil {
			return RegisterResult{}, fmt.Errorf("failed creating anonymous user: %w: %w", InternalError, err)
		}
	} else {
		if params.Email == nil {
			return RegisterResult{}, fmt.Errorf("email not provided but password was: %w", EmailNotProvided)
		}

		if params.Password == nil {
			return RegisterResult{}, fmt.Errorf("password was provided but email was not: %w", PasswordNotProvided)
		}

		argon2idPassword, err := argon2.Hash(string(*params.Password), nil)
		if err != nil {
			return RegisterResult{}, fmt.Errorf("failed hashing argon2 password: %w: %w", InternalError, err)
		}

		_, err = h.postgresClient.CreateUser(ctx, postgres.CreateUserParams{
			ID:           id,
			Username:     string(*params.Username),
			Email:        pgtype.Text{String: string(*params.Email), Valid: true},
			PasswordHash: pgtype.Text{String: argon2idPassword, Valid: true},
		})
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if pgErr.ConstraintName == "users_username_key" {
				return RegisterResult{}, fmt.Errorf("username already taken: %w: %w", UsernameTakenError, err)
			}

			if pgErr.ConstraintName == "users_email_key" {
				return RegisterResult{}, fmt.Errorf("email already taken: %w: %w", EmailTakenError, err)
			}

			// Fallback to less specific error, shouldn't reach here
			return RegisterResult{}, fmt.Errorf("email or username already taken: %w: %w", UsernameOrEmailTakenError, err)
		}
		if err != nil {
			return RegisterResult{}, fmt.Errorf("failed creating user: %w: %w", InternalError, err)
		}
	}

	issueTokenPairResult, err := h.authRepo.IssueTokenPair(ctx, id)
	if err != nil {
		return RegisterResult{}, fmt.Errorf("failed issuing token pair for registration: %w: %w", InternalError, err)
	}

	return RegisterResult{
		AccessToken:  issueTokenPairResult.AccessToken,
		RefreshToken: issueTokenPairResult.RefreshToken,
	}, nil
}

type LogoutParams struct {
	AccessToken string
}

func (h Service) Logout(ctx context.Context, params LogoutParams) error {
	accessTokenData, found, err := h.authRepo.GetAccessTokenData(ctx, params.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to get user id from refresh token: %w: %w", InternalError, err)
	}
	if !found {
		return UnauthorizedError
	}

	err = h.authRepo.DeleteTokenPair(ctx, params.AccessToken, accessTokenData.RefreshToken)
	if err != nil {
		return fmt.Errorf("failed to delete token pair: %w: %w", InternalError, err)
	}

	return nil
}
