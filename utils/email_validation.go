package utils

import "net/mail"

func ValidateEmailFormat(email string) string {
	if email == "" {
		return ""
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return "This is not a valid email: " + email
	}

	return ""
}
