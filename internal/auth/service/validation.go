package service

import (
	"fmt"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
)

var usernameRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9_.]{0,30}[a-z0-9]$`)

const (
	minPasswordLength    = 8
	maxPasswordLength    = 128
	maxDisplayNameLength = 32
)

type Email string

var EmailValidationError = Error{ExternalMessage: "Invalid email", HTTPCode: http.StatusBadRequest}

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

type Password string

var PasswordValidationError = Error{ExternalMessage: "Invalid password", HTTPCode: http.StatusBadRequest}
var PasswordLengthValidationError = Error{ExternalMessage: fmt.Sprintf("Invalid password length, must be %d to %d characters", minPasswordLength, maxPasswordLength), HTTPCode: http.StatusBadRequest}

func NewPassword(password string) (Password, error) {
	if password == "" {
		return "", fmt.Errorf("password is empty: %w", PasswordValidationError)
	}

	if len(password) < minPasswordLength {
		return "", fmt.Errorf("password too short: %w", PasswordLengthValidationError)
	}

	if len(password) > maxPasswordLength {
		return "", fmt.Errorf("password too long: %w", PasswordLengthValidationError)
	}

	return Password(password), nil
}

type Username string

var UsernameValidationError = Error{ExternalMessage: "Invalid username", HTTPCode: http.StatusBadRequest}

// NewUsername validates that the username is between 2-32 characters long, starts and ends with a letter or digit,
// only contains letters, digits, period and underscore.
func NewUsername(username string) (Username, error) {
	if username == "" {
		return "", fmt.Errorf("username is empty: %w", UsernameValidationError)
	}

	username = strings.ToLower(username)

	isValid := usernameRegex.MatchString(username)
	if !isValid {
		return "", fmt.Errorf("username does not match regex: %w", UsernameValidationError)
	}

	return Username(username), nil
}

type Login string

var LoginValidationError = Error{ExternalMessage: "Invalid login", HTTPCode: http.StatusBadRequest}

// NewLogin validates that the value is a valid email or username
func NewLogin(login string) (Login, error) {
	username, err := NewUsername(login)
	if err == nil {
		return Login(username), nil
	}

	email, err := NewEmail(login)
	if err == nil {
		return Login(email), nil
	}

	return "", fmt.Errorf("invalid login: %w", LoginValidationError)
}

type DisplayName string

var DisplayNameValidationError = Error{ExternalMessage: "Invalid display name", HTTPCode: http.StatusBadRequest}
var DisplayNameLengthValidationError = Error{ExternalMessage: fmt.Sprintf("Invalid display name length, must not be longer than %d characters", maxDisplayNameLength), HTTPCode: http.StatusBadRequest}

func NewDisplayName(displayName string) (DisplayName, error) {
	if displayName == "" {
		return "", fmt.Errorf("display name is empty: %w", DisplayNameValidationError)
	}

	if len(displayName) > maxDisplayNameLength {
		return "", fmt.Errorf("display name too long: %w", DisplayNameLengthValidationError)
	}

	return DisplayName(displayName), nil
}
