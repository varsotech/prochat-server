package http

import (
	"errors"
	"github.com/varsotech/prochat-server/internal/auth/service"
	prochatv1 "github.com/varsotech/prochat-server/internal/models/gen/prochat/v1"
	"github.com/varsotech/prochat-server/internal/pkg/redis/authrepo"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"log/slog"
	"net/http"
)

const (
	accessTokenCookieName  = "prochat_accesstoken"
	refreshTokenCookieName = "prochat_refreshtoken"
)

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

func (o *Routes) Refresh(w http.ResponseWriter, r *http.Request) {
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

func (o *Routes) Login(w http.ResponseWriter, r *http.Request) {
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

	login, err := service.NewLogin(req.Login)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	password, err := service.NewPassword(req.Password)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	loginResult, err := o.service.Login(r.Context(), service.LoginParams{
		Login:    login,
		Password: password,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	o.setTokenPairCookies(w, loginResult.AccessToken, loginResult.RefreshToken)
	w.WriteHeader(http.StatusOK)
}

func (o *Routes) Register(w http.ResponseWriter, r *http.Request) {
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

func (o *Routes) setTokenPairCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	accessTokenCookie := createCookie(accessTokenCookieName, accessToken, "/", authrepo.AccessTokenMaxAge)
	refreshTokenCookie := createCookie(refreshTokenCookieName, refreshToken, "/api/v1/auth/refresh", authrepo.RefreshTokenMaxAge)

	http.SetCookie(w, &accessTokenCookie)
	http.SetCookie(w, &refreshTokenCookie)
}
