package service

import (
	"fmt"
	"net/http"
	"net/mail"
)

var EmailValidationError = Error{ExternalMessage: "Invalid email", HTTPCode: http.StatusBadRequest}

type Email string

func NewEmail(email string) (Email, error) {
	if email == "" {
		return "", fmt.Errorf("email is empty: %w", EmailValidationError)
	}

	e, err := mail.ParseAddress(email)
	if err != nil {
		return "", fmt.Errorf("email is invalid: %w: %w", EmailValidationError, err)
	}

	return Email(e.String()), nil
}
