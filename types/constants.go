package types

type AuthenticatedUserID struct{}

type ImpersonatorUserID struct{}

type IsImpersonating struct{}

// CookieName is the default cookie name used by auth token helpers.
// It is a variable for backward compatibility and global override scenarios,
// but it should be set at init-time before any auth instances are used.
// For per-instance customization, use CookieConfig.Name or SetCookieName.
var CookieName = "authtoken"

// ImpersonationKeyPrefix is the prefix used for temporary keys that store
// the original admin auth token during an impersonation session.
const ImpersonationKeyPrefix = "imp:"
