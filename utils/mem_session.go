package utils

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/dracory/auth/types"
	"github.com/dracory/str"
)

// MemSession is a minimal in-memory SessionInterface implementation.
type MemSession struct {
	id        string
	userID    string
	key       string
	expiresAt time.Time
	values    map[string]string
}

var _ types.SessionInterface = (*MemSession)(nil)

// NewMemSession creates an empty MemSession with an initialized value map.
func NewMemSession() *MemSession {
	return &MemSession{values: map[string]string{}}
}

func (s *MemSession) GetID() string               { return s.id }
func (s *MemSession) SetID(id string)             { s.id = id }
func (s *MemSession) GetUserID() string           { return s.userID }
func (s *MemSession) SetUserID(userID string)     { s.userID = userID }
func (s *MemSession) GetKey() string              { return s.key }
func (s *MemSession) SetKey(key string)           { s.key = key }
func (s *MemSession) GetExpiresAt() time.Time     { return s.expiresAt }
func (s *MemSession) SetExpiresAt(t time.Time)    { s.expiresAt = t }
func (s *MemSession) GetValue(name string) string { return s.values[name] }
func (s *MemSession) SetValue(name, value string) { s.values[name] = value }

// MemSessionStore is an in-memory SessionStoreInterface for tests and
// zero-dependency setups. Expired sessions are filtered out on lookup and
// lazily dropped.
type MemSessionStore struct {
	mu       sync.Mutex
	byKey    map[string]*MemSession
	byUserID map[string]map[string]struct{}
}

var _ types.SessionStoreInterface = (*MemSessionStore)(nil)

// NewMemSessionStore creates an empty MemSessionStore.
func NewMemSessionStore() *MemSessionStore {
	return &MemSessionStore{
		byKey:    map[string]*MemSession{},
		byUserID: map[string]map[string]struct{}{},
	}
}

// NewSession returns a blank *MemSession for the store's concrete type.
func (m *MemSessionStore) NewSession() types.SessionInterface {
	return NewMemSession()
}

// SessionCreate stores the session, generating ID and key when unset.
func (m *MemSessionStore) SessionCreate(ctx context.Context, session types.SessionInterface) error {
	if session == nil {
		return errors.New("session is nil")
	}
	ms, ok := session.(*MemSession)
	if !ok {
		return errors.New("MemSessionStore only accepts *MemSession values")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if ms.id == "" {
		id, err := str.RandomFromGamma(32, "0123456789abcdef")
		if err != nil {
			return err
		}
		ms.id = id
	}
	if ms.key == "" {
		key, err := str.RandomFromGamma(32, "0123456789abcdef")
		if err != nil {
			return err
		}
		ms.key = key
	}

	m.byKey[ms.key] = ms
	if ms.userID != "" {
		if m.byUserID[ms.userID] == nil {
			m.byUserID[ms.userID] = map[string]struct{}{}
		}
		m.byUserID[ms.userID][ms.key] = struct{}{}
	}
	return nil
}

// SessionFindByKey returns the session for key, or nil (with nil error)
// when the key is unknown or the session has expired. Expired sessions are
// lazily removed.
func (m *MemSessionStore) SessionFindByKey(ctx context.Context, key string) (types.SessionInterface, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ms, ok := m.byKey[key]
	if !ok {
		return nil, nil
	}
	if !ms.expiresAt.IsZero() && time.Now().After(ms.expiresAt) {
		m.deleteLocked(ms.key)
		return nil, nil
	}
	return ms, nil
}

// SessionUpdate stores the session under its existing key (middlewares
// use it for grace-window shortening). Unknown keys are inserted.
func (m *MemSessionStore) SessionUpdate(ctx context.Context, session types.SessionInterface) error {
	if session == nil {
		return errors.New("session is nil")
	}
	ms, ok := session.(*MemSession)
	if !ok {
		return errors.New("MemSessionStore only accepts *MemSession values")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.byKey[ms.key] = ms
	if ms.userID != "" {
		if m.byUserID[ms.userID] == nil {
			m.byUserID[ms.userID] = map[string]struct{}{}
		}
		m.byUserID[ms.userID][ms.key] = struct{}{}
	}
	return nil
}

// SessionDeleteByKey removes the session for key; unknown keys are a no-op.
func (m *MemSessionStore) SessionDeleteByKey(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleteLocked(key)
	return nil
}

// SessionDeleteByUserID removes all sessions belonging to the user.
func (m *MemSessionStore) SessionDeleteByUserID(ctx context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for key := range m.byUserID[userID] {
		m.deleteLocked(key)
	}
	delete(m.byUserID, userID)
	return nil
}

func (m *MemSessionStore) deleteLocked(key string) {
	ms, ok := m.byKey[key]
	if !ok {
		return
	}
	delete(m.byKey, key)
	if keys, ok := m.byUserID[ms.userID]; ok {
		delete(keys, key)
	}
}
