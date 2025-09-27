package auth

import (
	"net/mail"
	"regexp"
)

var usernameRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9_.]{0,30}[a-z0-9]$`)

// validateUsername validates that the username is between 2-32 characters long, starts and ends with a letter or digit,
// only contains letters, digits, period and underscore.
func validateUsername(username string) (bool, string) {
	if username == "" {
		return false, "Username is required"
	}

	isValid := usernameRegex.MatchString(username)
	if !isValid {
		return false, "Username is not valid"
	}

	return true, ""
}

func validateEmail(email string) (bool, string) {
	if email == "" {
		return false, "Email is required"
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return false, "Email is not valid"
	}

	return true, ""
}

func validatePassword(password string) (bool, string) {
	if password == "" {
		return false, "Password is required"
	}

	const minPasswordLength = 8
	if len(password) < minPasswordLength {
		return false, "Password is too short"
	}

	const maxPasswordLength = 128
	if len(password) > maxPasswordLength {
		return false, "Password is too long"
	}

	return true, ""
}
