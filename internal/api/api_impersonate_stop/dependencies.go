package api_impersonate_stop

import (
	"context"
	"net/http"

	"github.com/dracory/auth/types"
)

type Dependencies struct {
	UserFindByAuthToken func(ctx context.Context, token string, options types.UserAuthOptions) (string, error)
	UserLogout          func(ctx context.Context, userID string, options types.UserAuthOptions) error
	TemporaryKeyGet     func(key string) (string, error)
	TemporaryKeySet     func(key string, value string, expiresSeconds int) error
	UseCookies          bool
	SetAuthCookie       func(w http.ResponseWriter, r *http.Request, token string)
	ObservabilityHooks  types.ObservabilityHooks
	ImpersonationStop   func(ctx context.Context, adminUserID string, targetUserID string) error
}
