package http

import (
	"net/http"
)

func createCookie(name, value, path string, maxAge int) http.Cookie {
	return http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
}
