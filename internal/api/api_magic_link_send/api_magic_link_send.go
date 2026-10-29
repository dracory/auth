// Package api_magic_link_send provides the JSON endpoint that sends a
// single-use magic login link to an email address.
package api_magic_link_send

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/dracory/api"
	"github.com/dracory/auth/internal/rules"
	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
	"github.com/dracory/req"
)

// Dependencies defines dependencies for sending a magic login link.
type Dependencies struct {
	// TemporaryKeyGet/Set give access to the short-lived key store; Get is
	// used for the per-email send throttle, Set for the token payload.
	TemporaryKeyGet    func(key string) (string, error)
	TemporaryKeySet    func(key string, value string, expiresSeconds int) error
	TemporaryKeyDelete func(key string) error

	// ExpiresSeconds is the link TTL; 0 defaults to utils.MagicLinkExpiresSeconds.
	ExpiresSeconds int

	// BindIP records the requester's IP in the payload; the callback then
	// requires the same IP. Default off — proxies, corporate NATs and
	// mobile network switches would otherwise break legitimate links.
	BindIP bool

	// EmailAllowed optionally gates which addresses may receive a link
	// (allowlist hook). Nil means all valid emails are allowed.
	EmailAllowed func(email string) bool

	// LinkMagicLinkCallback builds the absolute or relative callback URL
	// for a token, e.g. "/auth/magiclink-callback?token=" + token.
	LinkMagicLinkCallback func(token string) string

	// EmailTemplate builds the email body; receives the full callback link.
	// Optional — a minimal default body is used when nil.
	EmailTemplate func(ctx context.Context, email string, magicLink string) string

	// EmailSend delivers the email. Optional — when nil the link is only
	// stored (useful for queue-based delivery, task 008).
	EmailSend func(ctx context.Context, email string, subject string, body string) error

	Logger *slog.Logger
}

// MagicLinkSendErrorCode categorizes possible error sources.
type MagicLinkSendErrorCode string

const (
	MagicLinkSendErrorCodeNone           MagicLinkSendErrorCode = ""
	MagicLinkSendErrorCodeValidation     MagicLinkSendErrorCode = "validation"
	MagicLinkSendErrorCodeTokenGeneration MagicLinkSendErrorCode = "token_generation"
	MagicLinkSendErrorCodeTokenStore     MagicLinkSendErrorCode = "token_store"
	MagicLinkSendErrorCodeTooManySends   MagicLinkSendErrorCode = "too_many_sends"
	MagicLinkSendErrorCodeEmailSend      MagicLinkSendErrorCode = "email_send"
)

// MagicLinkSendError is a structured error for the send flow.
type MagicLinkSendError struct {
	Code    MagicLinkSendErrorCode
	Message string
	Err     error
}

func (e *MagicLinkSendError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// ApiMagicLinkSend is the HTTP handler: POST email (& optional return).
func ApiMagicLinkSend(w http.ResponseWriter, r *http.Request, deps Dependencies) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, perr := MagicLinkSend(r.Context(), r, deps)
	if perr != nil {
		switch perr.Code {
		case MagicLinkSendErrorCodeValidation,
			MagicLinkSendErrorCodeTooManySends:
			api.Respond(w, r, api.Error(perr.Message))
		case MagicLinkSendErrorCodeEmailSend:
			api.Respond(w, r, api.Error(types.MsgFailedToSendEmail))
		default:
			api.Respond(w, r, api.Error(types.MsgInternalServer))
		}
		return
	}

	api.Respond(w, r, api.Success("Magic link sent"))
}

// MagicLinkSend validates the request, stores the token payload, and
// delivers the link via EmailSend.
func MagicLinkSend(ctx context.Context, r *http.Request, deps Dependencies) (bool, *MagicLinkSendError) {
	email := req.GetStringTrimmed(r, "email")
	returnPath := req.GetStringTrimmed(r, "return")

	if email == "" {
		return false, &MagicLinkSendError{Code: MagicLinkSendErrorCodeValidation, Message: types.MsgEmailRequired}
	}
	if msg := utils.ValidateEmailFormat(email); msg != "" {
		return false, &MagicLinkSendError{Code: MagicLinkSendErrorCodeValidation, Message: msg}
	}
	// `return` must be a safe relative path or dropped entirely — never
	// propagate an external URL into the emailed link.
	if returnPath != "" && !utils.IsSafeReturnPath(returnPath) {
		returnPath = ""
	}

	// Allowlist check runs before any token is created or email sent.
	if deps.EmailAllowed != nil && !deps.EmailAllowed(email) {
		return false, &MagicLinkSendError{Code: MagicLinkSendErrorCodeValidation, Message: types.MsgInvalidCredentials}
	}

	// Per-email send throttle (shared with the code-send paths).
	if check := rules.NewEmailSendThrottleRule(deps.TemporaryKeyGet, email); check.Fails() {
		return false, &MagicLinkSendError{Code: MagicLinkSendErrorCodeTooManySends, Message: check.FailMessageFirst()}
	}

	if deps.TemporaryKeySet == nil {
		return false, &MagicLinkSendError{Code: MagicLinkSendErrorCodeTokenStore, Message: types.MsgFailedToProcess}
	}

	token, err := utils.GenerateNonce()
	if err != nil {
		return false, &MagicLinkSendError{Code: MagicLinkSendErrorCodeTokenGeneration, Err: err}
	}

	expires := deps.ExpiresSeconds
	if expires <= 0 {
		expires = utils.MagicLinkExpiresSeconds
	}

	ip := ""
	if deps.BindIP {
		ip = req.GetIP(r)
	}

	payload, err := json.Marshal(utils.MagicLinkPayload{
		Email:    email,
		IP:       ip,
		Token:    token,
		Return:   returnPath,
		Remember: req.GetStringTrimmed(r, "remember") == "true",
	})
	if err != nil {
		return false, &MagicLinkSendError{Code: MagicLinkSendErrorCodeTokenStore, Err: err}
	}

	if err := deps.TemporaryKeySet(utils.MagicLinkKeyPrefix+token, string(payload), expires); err != nil {
		return false, &MagicLinkSendError{Code: MagicLinkSendErrorCodeTokenStore, Err: err}
	}

	if deps.EmailSend == nil {
		return true, nil // stored only — a queue picks it up (task 008)
	}

	link := token
	if deps.LinkMagicLinkCallback != nil {
		link = deps.LinkMagicLinkCallback(token)
	}

	body := "Click to log in: " + link
	if deps.EmailTemplate != nil {
		body = deps.EmailTemplate(ctx, email, link)
	}
	if err := deps.EmailSend(ctx, email, "Login Link", body); err != nil {
		// The link was never delivered — remove the orphaned token so it
		// cannot be redeemed without the email reaching the user.
		_ = utils.InvalidateTempKey(deps.TemporaryKeyDelete, deps.TemporaryKeySet, utils.MagicLinkKeyPrefix+token)
		return false, &MagicLinkSendError{Code: MagicLinkSendErrorCodeEmailSend, Err: err}
	}

	// Consume send quota only after the email was handed off successfully.
	if deps.TemporaryKeyGet != nil {
		_ = utils.RecordEmailSend(deps.TemporaryKeyGet, deps.TemporaryKeySet, email, expires)
	}

	return true, nil
}
