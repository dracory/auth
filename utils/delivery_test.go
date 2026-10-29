package utils

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestDeliverSecret_EnqueuePrecedenceAndPrivacy(t *testing.T) {
	store := map[string]string{}
	var enqueuedKey string
	sendCalled := false

	get := func(k string) (string, error) {
		if v, ok := store[k]; ok {
			return v, nil
		}
		return "", errors.New("nf")
	}
	err := DeliverSecret(context.Background(), get,
		func(k, v string, e int) error { store[k] = v; return nil },
		func(ctx context.Context, key string) error { enqueuedKey = key; return nil },
		func(ctx context.Context, email, subject, body string) error { sendCalled = true; return nil },
		"u@x.com", "BCDFGHJK", "Login Code", "code: BCDFGHJK", 900)

	if err != nil {
		t.Fatalf("deliver: %v", err)
	}
	if sendCalled {
		t.Fatal("enqueue must take precedence over send")
	}
	if enqueuedKey == "" {
		t.Fatal("expected delivery key")
	}
	if strings.Contains(enqueuedKey, "u@x.com") || strings.Contains(enqueuedKey, "BCDFGHJK") {
		t.Fatalf("queue payload must contain only the opaque key, got %q", enqueuedKey)
	}

	// Worker side: resolve the opaque key into the full delivery — email,
	// secret, subject AND body are all recoverable.
	d, ok := ResolveDeliveryKey(get, enqueuedKey)
	if !ok || d.Email != "u@x.com" || d.Secret != "BCDFGHJK" ||
		d.Subject != "Login Code" || d.Body != "code: BCDFGHJK" {
		t.Fatalf("resolve mismatch: %+v ok=%v", d, ok)
	}
}

func TestConsumeDeliveryKey(t *testing.T) {
	store := map[string]string{DeliveryKeyPrefix + "k1": `{"email":"u@x.com","secret":"s","subject":"s","body":"b"}`}
	get := func(k string) (string, error) {
		if v, ok := store[k]; ok {
			return v, nil
		}
		return "", errors.New("nf")
	}
	del := func(k string) error { delete(store, k); return nil }

	d, ok := ConsumeDeliveryKey(get, del, nil, "k1")
	if !ok || d.Email != "u@x.com" {
		t.Fatalf("consume: %+v ok=%v", d, ok)
	}
	if _, err := get(DeliveryKeyPrefix + "k1"); err == nil {
		t.Fatal("expected key consumed")
	}

	// Second consume must fail — retried jobs don't re-send.
	if _, ok := ConsumeDeliveryKey(get, del, nil, "k1"); ok {
		t.Fatal("expected second consume to fail")
	}
}

func TestResolveDeliveryKey_Unknown(t *testing.T) {
	_, ok := ResolveDeliveryKey(func(k string) (string, error) {
		return "", errors.New("nf")
	}, "ghost")
	if ok {
		t.Fatal("expected ok=false for unknown key")
	}
}

func TestDeliverSecret_FallsBackToSend(t *testing.T) {
	var sentTo string
	err := DeliverSecret(context.Background(), nil, nil, nil,
		func(ctx context.Context, email, subject, body string) error { sentTo = email; return nil },
		"u@x.com", "BCDFGHJK", "S", "B", 0)
	if err != nil || sentTo != "u@x.com" {
		t.Fatalf("expected send fallback, got %v %q", err, sentTo)
	}
}

func TestDeliverSecret_NeitherHook(t *testing.T) {
	err := DeliverSecret(context.Background(), nil, nil, nil, nil, "u@x.com", "x", "s", "b", 0)
	if err == nil {
		t.Fatal("expected error when no delivery hook configured")
	}
}
