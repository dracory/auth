package utils

import (
	"context"
	"testing"
	"time"
)

func TestMemSessionStore_CreateFindRoundTrip(t *testing.T) {
	store := NewMemSessionStore()
	s := NewMemSession()
	s.SetUserID("user1")
	s.SetExpiresAt(time.Now().Add(time.Hour))
	s.SetValue("ip", "1.2.3.4")

	if err := store.SessionCreate(context.Background(), s); err != nil {
		t.Fatalf("create: %v", err)
	}
	if s.GetKey() == "" || s.GetID() == "" {
		t.Fatal("expected generated key and id")
	}

	found, err := store.SessionFindByKey(context.Background(), s.GetKey())
	if err != nil || found == nil {
		t.Fatalf("find: %v", err)
	}
	if found.GetUserID() != "user1" || found.GetValue("ip") != "1.2.3.4" {
		t.Fatalf("round-trip mismatch: %v", found)
	}
}

func TestMemSessionStore_FindUnknownReturnsNil(t *testing.T) {
	store := NewMemSessionStore()
	found, err := store.SessionFindByKey(context.Background(), "nope")
	if err != nil || found != nil {
		t.Fatalf("expected nil session, nil error; got %v, %v", found, err)
	}
}

func TestMemSessionStore_DeleteByKey(t *testing.T) {
	store := NewMemSessionStore()
	s := NewMemSession()
	s.SetUserID("user1")
	_ = store.SessionCreate(context.Background(), s)

	if err := store.SessionDeleteByKey(context.Background(), s.GetKey()); err != nil {
		t.Fatalf("delete: %v", err)
	}
	found, _ := store.SessionFindByKey(context.Background(), s.GetKey())
	if found != nil {
		t.Fatal("expected deleted session to be gone")
	}
}

func TestMemSessionStore_DeleteByUserID(t *testing.T) {
	store := NewMemSessionStore()
	s1, s2, other := NewMemSession(), NewMemSession(), NewMemSession()
	s1.SetUserID("user1")
	s2.SetUserID("user1")
	other.SetUserID("user2")
	_ = store.SessionCreate(context.Background(), s1)
	_ = store.SessionCreate(context.Background(), s2)
	_ = store.SessionCreate(context.Background(), other)

	_ = store.SessionDeleteByUserID(context.Background(), "user1")

	for _, key := range []string{s1.GetKey(), s2.GetKey()} {
		found, _ := store.SessionFindByKey(context.Background(), key)
		if found != nil {
			t.Fatalf("expected user1 session %q deleted", key)
		}
	}
	found, _ := store.SessionFindByKey(context.Background(), other.GetKey())
	if found == nil {
		t.Fatal("expected other user's session intact")
	}
}

func TestMemSessionStore_ExpiredSessionNotFound(t *testing.T) {
	store := NewMemSessionStore()
	s := NewMemSession()
	s.SetUserID("user1")
	s.SetExpiresAt(time.Now().Add(-time.Minute)) // already expired
	_ = store.SessionCreate(context.Background(), s)

	found, _ := store.SessionFindByKey(context.Background(), s.GetKey())
	if found != nil {
		t.Fatal("expected expired session to be dropped on lookup")
	}
}
