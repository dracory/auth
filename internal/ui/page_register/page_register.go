package page_register

import (
	"context"
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/rules"
	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/auth/types"
	authutils "github.com/dracory/auth/utils"
	"github.com/dracory/req"
)

// PageRegister renders the register page using the provided dependencies and
// writes the result to the ResponseWriter.

func PageRegister(w http.ResponseWriter, r *http.Request, a types.AuthSharedInterface) {
	// With a verified-email key (verified_key issued by an OTP/magic-link
	// verification or the AuthKnight callback): render the verified-email
	// register form. Checked first — before the AuthKnight redirect guard.
	verifiedKey := r.URL.Query().Get("verified_key")
	if verifiedKey != "" {
		email := getVerifiedEmail(a, verifiedKey)
		if email != "" {
			// The URL carries a bearer credential — do not leak it via
			// Referer to any subresource on this page.
			w.Header().Set("Referrer-Policy", "no-referrer")
			content := RegisterVerifiedEmailContent(links.Login(a.GetEndpoint()), email, verifiedKey)
			scripts := RegisterVerifiedEmailScripts(
				links.ApiRegister(a.GetEndpoint()),
				a.LinkRedirectOnSuccess(),
			)
			shared.PageRender(w, r, shared.PageOptions{
				Title:      "Register",
				Auth:       a,
				Layout:     a.GetLayout(),
				Content:    content,
				Scripts:    scripts,
				Logger:     a.GetLogger(),
				LogMessage: "failed to write register page response",
			})
			return
		}
	}

	// Business rule: an authenticated session turns registration into
	// profile completion — the email is already verified, shown
	// read-only, and never read from the form.
	if completion := rules.NewProfileCompletionRule(func(r *http.Request) string {
		return authutils.ResolveCurrentUserID(r, a)
	}, r); !completion.Fails() {
		userID := completion.UserID()
		email := ""
		if provider, ok := a.(interface {
			GetFuncUserEmailGet() func(ctx context.Context, userID string, options types.UserAuthOptions) (string, error)
		}); ok {
			if fn := provider.GetFuncUserEmailGet(); fn != nil {
				email, _ = fn(r.Context(), userID, types.UserAuthOptions{
					UserIp:    req.GetIP(r),
					UserAgent: r.UserAgent(),
				})
			}
		}
		content := RegisterCompletionContent(links.Login(a.GetEndpoint()), email)
		scripts := RegisterCompletionScripts(
			links.ApiRegister(a.GetEndpoint()),
			a.LinkRedirectOnSuccess(),
		)
		shared.PageRender(w, r, shared.PageOptions{
			Title:      "Complete Registration",
			Auth:       a,
			Layout:     a.GetLayout(),
			Content:    content,
			Scripts:    scripts,
			Logger:     a.GetLogger(),
			LogMessage: "failed to write register page response",
		})
		return
	}

	if shared.RedirectAuthKnightIfActive(w, r, a) {
		return
	}

	content := ""
	scripts := ""

	if a.IsOtp() {
		content = RegisterOtpContent(links.Login(a.GetEndpoint()))
		scripts = RegisterOtpScripts(
			links.ApiRegister(a.GetEndpoint()),
			links.RegisterCodeVerify(a.GetEndpoint()),
		)
	} else {
		content = RegisterUsernameAndPasswordContent(
			links.Login(a.GetEndpoint()),
			links.PasswordRestore(a.GetEndpoint()),
		)
		urlSuccess := links.Login(a.GetEndpoint())
		if a.IsVerificationEnabled() {
			urlSuccess = links.RegisterCodeVerify(a.GetEndpoint())
		}
		scripts = RegisterUsernameAndPasswordScripts(
			links.ApiRegister(a.GetEndpoint()),
			urlSuccess,
		)
	}

	shared.PageRender(w, r, shared.PageOptions{
		Title:      "Register",
		Auth:       a,
		Layout:     a.GetLayout(),
		Content:    content,
		Scripts:    scripts,
		Logger:     a.GetLogger(),
		LogMessage: "failed to write register page response",
	})
}

// getVerifiedEmail resolves the verified email from the temporary key
// store via the shared VerifiedEmailKeyRule (empty string on failure).
func getVerifiedEmail(a types.AuthSharedInterface, verifiedKey string) string {
	if verifiedKey == "" {
		return ""
	}
	keyRule := rules.NewVerifiedEmailKeyRule(a.GetFuncTemporaryKeyGet(), verifiedKey)
	if keyRule.Fails() {
		return ""
	}
	return keyRule.Email()
}
