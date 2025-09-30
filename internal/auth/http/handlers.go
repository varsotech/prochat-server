package http

import (
	"errors"
	"github.com/varsotech/prochat-server/internal/auth/internal/authrepo"
	"github.com/varsotech/prochat-server/internal/auth/service"
	prochatv1 "github.com/varsotech/prochat-server/internal/models/gen/prochat/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"log/slog"
	"net/http"
)

const (
	accessTokenCookieName  = "prochat_accesstoken"
	refreshTokenCookieName = "prochat_refreshtoken"
	accessTokenCookiePath  = "/"
	refreshTokenCookiePath = "/api/v1/auth/refresh"
)

//func (h Handlers) LoginProtectionMiddleware(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

//		next.ServeHTTP(w, r)
//	})
//}

func (o *Service) refresh(w http.ResponseWriter, r *http.Request) {
	refreshTokenCookie, err := r.Cookie(refreshTokenCookieName)
	if err != nil {
		slog.Error("error getting refreshToken cookie", "error", err)
		http.Error(w, "error getting cookie", http.StatusUnauthorized)
		return
	}

	refreshResult, err := o.service.Refresh(r.Context(), refreshTokenCookie.Value)
	if err != nil {
		slog.Error("refresh failed", "error", err, "request_uri", r.RequestURI)

		var serviceErr service.Error
		if errors.As(err, &serviceErr) {
			http.Error(w, serviceErr.ExternalMessage, serviceErr.HTTPCode)
			return
		}

		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	o.setTokenPairCookies(w, refreshResult.AccessToken, refreshResult.RefreshToken)
	w.WriteHeader(http.StatusOK)
}

func (o *Service) login(w http.ResponseWriter, r *http.Request) {
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

	loginResult, err := o.service.Login(r.Context(), service.LoginParams{
		Login:    req.Login,
		Password: req.Password,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	o.setTokenPairCookies(w, loginResult.AccessToken, loginResult.RefreshToken)
	w.WriteHeader(http.StatusOK)
}

func (o *Service) register(w http.ResponseWriter, r *http.Request) {
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

	var userName *service.Username
	if req.Username != "" {
		validatedUserName, err := service.NewUsername(req.Username)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		userName = &validatedUserName
	}

	var email *service.Email
	if req.Email != "" {
		validatedEmail, err := service.NewEmail(req.Email)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		email = &validatedEmail
	}

	var password *service.Password
	if req.Password != "" {
		validatedPassword, err := service.NewPassword(req.Password)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		password = &validatedPassword
	}

	displayName, err := service.NewDisplayName(req.DisplayName)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	registerResult, err := o.service.Register(r.Context(), service.RegisterParams{
		DisplayName: displayName,
		Username:    userName,
		Email:       email,
		Password:    password,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	o.setTokenPairCookies(w, registerResult.AccessToken, registerResult.RefreshToken)
	w.WriteHeader(http.StatusOK)
}

func (o *Service) logout(w http.ResponseWriter, r *http.Request) {
	accessTokenData, err := o.Authenticate(r)
	if errors.Is(err, UnauthorizedError) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if err != nil {
		slog.Error("failed to authenticate user", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	err = o.service.Logout(r.Context(), service.LogoutParams{
		AccessToken: accessTokenData.AccessToken,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	accessTokenCookie := createCookie(accessTokenCookieName, "", accessTokenCookiePath, -1)
	refreshTokenCookie := createCookie(refreshTokenCookieName, "", refreshTokenCookiePath, -1)

	http.SetCookie(w, &accessTokenCookie)
	http.SetCookie(w, &refreshTokenCookie)

	w.WriteHeader(http.StatusOK)
}

func (o *Service) setTokenPairCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	accessTokenCookie := createCookie(accessTokenCookieName, accessToken, accessTokenCookiePath, authrepo.AccessTokenMaxAge)
	refreshTokenCookie := createCookie(refreshTokenCookieName, refreshToken, refreshTokenCookiePath, authrepo.RefreshTokenMaxAge)

	http.SetCookie(w, &accessTokenCookie)
	http.SetCookie(w, &refreshTokenCookie)
}
