package api_password_restore

import "context"

// Dependencies defines the dependencies required for the password
// restore flow (issuing a password reset link).
type Dependencies struct {
	UserFindByUsername func(ctx context.Context, email, firstName, lastName string) (userID string, err error)

	TemporaryKeySet func(key string, value string, expiresSeconds int) error
	ExpiresSeconds  int

	EmailTemplate func(ctx context.Context, userID, token string) string
	EmailSend     func(ctx context.Context, userID, subject, body string) error
}
