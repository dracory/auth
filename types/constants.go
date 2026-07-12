package types

type AuthenticatedUserID struct{}

type ImpersonatorUserID struct{}

type IsImpersonating struct{}

const CookieName = "authtoken"

// ImpersonationKeyPrefix is the prefix used for temporary keys that store
// the original admin auth token during an impersonation session.
const ImpersonationKeyPrefix = "imp:"
