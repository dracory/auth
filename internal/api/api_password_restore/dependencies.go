package api_password_restore

import (
	"context"
	"errors"
	"log/slog"
)

// dependencies is the internal, fully-validated dependency set used by
// the password-restore business logic.
type dependencies struct {
	UserFindByUsername func(ctx context.Context, email, firstName, lastName string) (userID string, err error)

	TemporaryKeySet func(key string, value string, expiresSeconds int) error
	ExpiresSeconds  int

	EmailTemplate func(ctx context.Context, userID, token string) string
	EmailSend     func(ctx context.Context, userID, subject, body string) error

	Logger *slog.Logger
}

// NewDependencies validates that all required dependencies are provided
// and returns a fully-initialized private dependencies value or an error.
func NewDependencies(
	userFindByUsername func(ctx context.Context, email, firstName, lastName string) (userID string, err error),
	temporaryKeySet func(key string, value string, expiresSeconds int) error,
	expiresSeconds int,
	emailTemplate func(ctx context.Context, userID, token string) string,
	emailSend func(ctx context.Context, userID, subject, body string) error,
	logger *slog.Logger,
) (dependencies, error) {
	if logger == nil {
		return dependencies{}, errors.New("logger is required")
	}
	if userFindByUsername == nil {
		return dependencies{}, errors.New("UserFindByUsername is required")
	}
	if temporaryKeySet == nil {
		return dependencies{}, errors.New("TemporaryKeySet is required")
	}
	if emailTemplate == nil {
		return dependencies{}, errors.New("EmailTemplate is required")
	}
	if emailSend == nil {
		return dependencies{}, errors.New("EmailSend is required")
	}
	if expiresSeconds <= 0 {
		// default expiration: one hour
		expiresSeconds = 3600
	}
	return dependencies{
		UserFindByUsername: userFindByUsername,
		TemporaryKeySet:    temporaryKeySet,
		ExpiresSeconds:     expiresSeconds,
		EmailTemplate:      emailTemplate,
		EmailSend:          emailSend,
		Logger:             logger,
	}, nil
}
