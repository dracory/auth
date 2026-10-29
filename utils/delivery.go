package utils

import (
	"context"
	"encoding/json"
	"errors"
)

// DeliveryKeyPrefix prefixes the temporary keys under which queued email
// deliveries are stored. Queue payloads contain only the opaque key — the
// secret (code or link) lives solely in the TTL-bound temporary-key store.
const DeliveryKeyPrefix = "delivery:"

// Delivery is the stored payload behind a delivery key. It carries the
// full rendered message — subject and body included — so a queue worker
// needs nothing beyond resolve + send; it does not have to know which
// delivery type or template produced the email.
type Delivery struct {
	Email   string `json:"email"`
	Secret  string `json:"secret"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// ResolveDeliveryKey resolves an enqueued delivery back into its full
// Delivery payload. The queue worker calls this at execution time — the
// module provides the resolve half of the pattern so apps don't reimplement
// it. ok is false for unknown, expired, or malformed keys. Resolution does
// NOT consume the key — use ConsumeDeliveryKey when a retried queue job
// must not re-send.
func ResolveDeliveryKey(get func(key string) (string, error), deliveryKey string) (Delivery, bool) {
	var d Delivery
	if get == nil {
		return d, false
	}
	raw, err := get(DeliveryKeyPrefix + deliveryKey)
	if err != nil || raw == "" {
		return d, false
	}
	if json.Unmarshal([]byte(raw), &d) != nil || d.Email == "" {
		return Delivery{}, false
	}
	return d, true
}

// ConsumeDeliveryKey resolves the delivery and deletes the key in one step,
// preventing duplicate sends when a queue retries jobs (at-least-once
// queues). deleteFn may be nil, falling back to the "empty value, 0 expiry"
// convention via InvalidateTempKey.
func ConsumeDeliveryKey(
	get func(key string) (string, error),
	deleteFn func(key string) error,
	setFn func(key string, value string, expiresSeconds int) error,
	deliveryKey string,
) (Delivery, bool) {
	d, ok := ResolveDeliveryKey(get, deliveryKey)
	if !ok {
		return d, false
	}
	_ = InvalidateTempKey(deleteFn, setFn, DeliveryKeyPrefix+deliveryKey)
	return d, true
}

// DeliverSecret sends a secret (verification code, magic link) via either
// the synchronous send hook or the queue-safe enqueue hook.
//
// Enqueue takes precedence when set: the function stores a JSON Delivery
// (email, secret, subject, body) under a fresh delivery key and calls
// enqueueFn with ONLY that key — the queue payload contains no secret or
// email. The worker later resolves it with ResolveDeliveryKey or
// ConsumeDeliveryKey. When enqueueFn is nil, sendFn delivers directly;
// subject/body are only used in that path. An error is returned when
// neither hook is configured — callers surface it as a delivery failure.
func DeliverSecret(
	ctx context.Context,
	get func(key string) (string, error),
	set func(key string, value string, expiresSeconds int) error,
	enqueueFn func(ctx context.Context, deliveryKey string) error,
	sendFn func(ctx context.Context, email string, subject string, body string) error,
	email string,
	secret string,
	subject string,
	body string,
	ttlSeconds int,
) error {
	if enqueueFn != nil {
		if set == nil {
			return errors.New("temporary key store required for queued delivery")
		}
		if ttlSeconds <= 0 {
			ttlSeconds = 3600
		}
		deliveryKey, err := GenerateNonce()
		if err != nil {
			return err
		}
		payload, err := json.Marshal(Delivery{
			Email:   email,
			Secret:  secret,
			Subject: subject,
			Body:    body,
		})
		if err != nil {
			return err
		}
		if err := set(DeliveryKeyPrefix+deliveryKey, string(payload), ttlSeconds); err != nil {
			return err
		}
		return enqueueFn(ctx, deliveryKey)
	}
	if sendFn != nil {
		return sendFn(ctx, email, subject, body)
	}
	return errors.New("no email delivery configured")
}
