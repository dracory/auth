package types

import (
	"context"
	"time"
)

// SessionInterface is the minimal session abstraction the module needs —
// deliberately free of any concrete store dependency (no sessionstore
// import) so each host project adapts its own store.
type SessionInterface interface {
	GetID() string
	SetID(id string)

	GetUserID() string
	SetUserID(userID string)

	// GetKey returns the opaque session key placed in the auth cookie.
	GetKey() string
	SetKey(key string)

	GetExpiresAt() time.Time
	SetExpiresAt(t time.Time)

	// GetValue/SetValue expose optional free-form attributes (e.g.
	// "remember" markers, IP, user-agent) without widening the interface.
	GetValue(name string) string
	SetValue(name string, value string)
}

// SessionStoreInterface is the minimal persistence contract for sessions.
// Implementations adapt the host project's session store; an in-memory
// implementation ships in utils for tests and zero-dependency setups.
type SessionStoreInterface interface {
	// NewSession returns a blank session of the store's concrete type so
	// the module can mint sessions without knowing it.
	NewSession() SessionInterface

	SessionCreate(ctx context.Context, session SessionInterface) error
	SessionFindByKey(ctx context.Context, key string) (SessionInterface, error)
	SessionDeleteByKey(ctx context.Context, key string) error
	SessionDeleteByUserID(ctx context.Context, userID string) error
}
