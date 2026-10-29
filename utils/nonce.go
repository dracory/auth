package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// LoginCodeKeyPrefix prefixes the temporary keys under which passwordless
// login codes are stored. The nonce — not the code — is the storage key so
// that guessing a code does not reveal or consume it.
const LoginCodeKeyPrefix = "logincode:"

// RegisterCodeKeyPrefix prefixes the temporary keys under which registration
// code payloads are stored.
const RegisterCodeKeyPrefix = "registercode:"

// GenerateNonce generates a cryptographically random 128-bit hex nonce used
// as the temporary-key under which verification secrets are stored. The
// nonce is returned to the client and must be presented alongside the code
// on verification, so brute-forcing a code also requires guessing the nonce.
func GenerateNonce() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// IsValidNonce reports whether s is a well-formed nonce: exactly 32
// lowercase/uppercase hex characters. Verified before use as a temporary-key
// suffix to reject malformed input early.
func IsValidNonce(s string) bool {
	if len(s) != 32 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

// InvalidateTempKey removes a temporary key, preferring a real delete over
// the "empty value, 0 expiry" convention used for authknight keys. Shared by
// the login- and register-code verify paths; a nil delete function falls
// back to set, and both nil is a no-op returning nil.
func InvalidateTempKey(
	deleteFn func(key string) error,
	setFn func(key string, value string, expiresSeconds int) error,
	key string,
) error {
	if deleteFn != nil {
		return deleteFn(key)
	}
	if setFn != nil {
		return setFn(key, "", 0)
	}
	return nil
}
