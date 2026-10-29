package auth

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/dracory/auth/internal/api/api_authknight_callback"
	"github.com/dracory/auth/internal/api/api_login"
	"github.com/dracory/auth/internal/api/api_login_code_verify"
	"github.com/dracory/auth/internal/api/api_magic_link_callback"
	"github.com/dracory/auth/internal/api/api_magic_link_send"
	"github.com/dracory/auth/internal/api/api_password_reset"
	"github.com/dracory/auth/internal/api/api_password_restore"
	"github.com/dracory/auth/internal/ui/page_login"
	"github.com/dracory/auth/internal/ui/page_login_code_verify"
	"github.com/dracory/auth/internal/ui/page_password_reset"
	"github.com/dracory/auth/internal/ui/page_password_restore"
	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
)

// MultiMethodAuth composes N login methods behind a single mux:
// the first (primary) method mounts at <endpoint>/login, secondary methods
// at <endpoint>/login-<name>, method callbacks mount once each regardless
// of position, and method-provided extra routes (e.g. password recovery)
// mount only when the providing method is in the list.
//
// Custom methods implement types.Method — no library change needed.
type MultiMethodAuth struct {
	endpoint string
	methods  []types.Method
	mux      *http.ServeMux
}

// NewMultiMethodAuth builds the composed mux. Returns an error for an
// empty method list or duplicate method names (no silent shadowing).
func NewMultiMethodAuth(endpoint string, methods []types.Method) (*MultiMethodAuth, error) {
	if len(methods) == 0 {
		return nil, fmt.Errorf("multi-method auth requires at least one method")
	}
	endpoint = strings.TrimSuffix(endpoint, "/")
	if endpoint == "" {
		endpoint = "/auth"
	}

	slugRe := regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
	seen := map[string]bool{}
	for _, m := range methods {
		if m == nil || m.Name() == "" {
			return nil, fmt.Errorf("method with empty name")
		}
		if !slugRe.MatchString(m.Name()) {
			return nil, fmt.Errorf("method name %q must be a route slug (lowercase, digits, hyphens)", m.Name())
		}
		if seen[m.Name()] {
			return nil, fmt.Errorf("duplicate method name %q", m.Name())
		}
		seen[m.Name()] = true
	}

	mma := &MultiMethodAuth{endpoint: endpoint, methods: methods, mux: http.NewServeMux()}
	mounted := map[string]bool{}
	mount := func(pattern string, handler http.HandlerFunc) error {
		if mounted[pattern] {
			return fmt.Errorf("duplicate route %q", pattern)
		}
		mounted[pattern] = true
		mma.mux.HandleFunc(pattern, handler)
		return nil
	}

	for i, m := range methods {
		base := endpoint + "/login"
		if i > 0 {
			base = endpoint + "/login-" + m.Name()
		}
		if page := m.PageHandler(); page != nil {
			if err := mount("GET "+base, page); err != nil {
				return nil, err
			}
		}
		if ajax := m.AjaxHandler(); ajax != nil {
			if err := mount("POST "+base, ajax); err != nil {
				return nil, err
			}
		}
		if path, handler, ok := m.CallbackRoute(); ok {
			// The path may embed a method ("POST /x-callback") — same
			// convention as ExtraRoutes; defaults to GET.
			pattern := "GET " + endpoint + path
			if strings.Contains(path, " ") {
				parts := strings.SplitN(path, " ", 2)
				pattern = parts[0] + " " + endpoint + parts[1]
			}
			if err := mount(pattern, handler); err != nil {
				return nil, err
			}
		}
		if rp, ok := m.(types.RouteProvider); ok {
			for pattern, handler := range rp.ExtraRoutes() {
				if handler == nil {
					continue
				}
				// Keys may embed a method ("GET /path") or be path-only.
				full := pattern
				if strings.Contains(pattern, " ") {
					parts := strings.SplitN(pattern, " ", 2)
					full = parts[0] + " " + endpoint + parts[1]
				} else {
					full = endpoint + pattern
				}
				if err := mount(full, handler); err != nil {
					return nil, err
				}
			}
		}
	}
	return mma, nil
}

// Handler returns the composed http.Handler (the underlying ServeMux).
func (m *MultiMethodAuth) Handler() http.Handler { return m.mux }

// Router returns the composed *http.ServeMux.
func (m *MultiMethodAuth) Router() *http.ServeMux { return m.mux }

// Methods returns the configured method list (primary first).
func (m *MultiMethodAuth) Methods() []types.Method { return m.methods }

// LoginAlternativesHTML renders the "OR — use another method" fragment for
// the given current method name. Pages (built-in or app-supplied) can embed
// it so users can switch between mounted methods.
func (m *MultiMethodAuth) LoginAlternativesHTML(currentMethod string) string {
	return utils.LoginAlternatives(m.endpoint, currentMethod, m.methods)
}

// ============================================================================
// Built-in method adapters
// ============================================================================

// otpMethod wraps the passwordless (email code) flow.
type otpMethod struct {
	a types.AuthSharedInterface
}

// NewOTPMethod adapts the passwordless email-code login to types.Method.
func NewOTPMethod(a types.AuthSharedInterface) types.Method {
	return &otpMethod{a: a}
}

func (m *otpMethod) Name() string             { return "otp" }
func (m *otpMethod) AlternativeLabel() string { return "Email me a code" }
func (m *otpMethod) PageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { page_login.PageLogin(w, r, m.a) }
}
func (m *otpMethod) AjaxHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { api_login.ApiLoginWithAuth(w, r, m.a) }
}
func (m *otpMethod) CallbackRoute() (string, http.HandlerFunc, bool) { return "", nil, false }

// ExtraRoutes mounts the code-verify endpoints so OTP login can complete.
func (m *otpMethod) ExtraRoutes() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"GET /login-code-verify": func(w http.ResponseWriter, r *http.Request) { page_login_code_verify.PageLoginCodeVerify(w, r, m.a) },
		"POST /api/login-code-verify": func(w http.ResponseWriter, r *http.Request) {
			api_login_code_verify.ApiLoginCodeVerifyWithAuth(w, r, m.a)
		},
	}
}

// api_magic_link_send plus the independently mountable callback.
type magicLinkMethod struct {
	page         http.HandlerFunc
	sendDeps     api_magic_link_send.Dependencies
	callbackDeps api_magic_link_callback.Dependencies
}

// NewMagicLinkMethod adapts the magic-link login to types.Method. The page
// handler is supplied by the caller (a simple email form posting to the
// send ajax is enough); sendDeps/callbackDeps are the 007 dependency sets.
func NewMagicLinkMethod(
	page http.HandlerFunc,
	sendDeps api_magic_link_send.Dependencies,
	callbackDeps api_magic_link_callback.Dependencies,
) types.Method {
	return &magicLinkMethod{page: page, sendDeps: sendDeps, callbackDeps: callbackDeps}
}

func (m *magicLinkMethod) Name() string                  { return "magiclink" }
func (m *magicLinkMethod) AlternativeLabel() string      { return "Email me a link" }
func (m *magicLinkMethod) PageHandler() http.HandlerFunc { return m.page }
func (m *magicLinkMethod) AjaxHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { api_magic_link_send.ApiMagicLinkSend(w, r, m.sendDeps) }
}
func (m *magicLinkMethod) CallbackRoute() (string, http.HandlerFunc, bool) {
	return "/magiclink-callback", func(w http.ResponseWriter, r *http.Request) {
		api_magic_link_callback.ApiMagicLinkCallback(w, r, m.callbackDeps)
	}, true
}

// passwordMethod wraps the username/password login plus recovery routes.
type passwordMethod struct {
	a    types.AuthSharedInterface
	page http.HandlerFunc
}

// NewPasswordMethod adapts the username/password login to types.Method.
// page is the login page handler; recovery routes mount only when this
// method is in the list (RouteProvider).
func NewPasswordMethod(a types.AuthSharedInterface, page http.HandlerFunc) types.Method {
	return &passwordMethod{a: a, page: page}
}

func (m *passwordMethod) Name() string                  { return "password" }
func (m *passwordMethod) AlternativeLabel() string      { return "Use a password" }
func (m *passwordMethod) PageHandler() http.HandlerFunc { return m.page }
func (m *passwordMethod) AjaxHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { api_login.ApiLoginWithAuth(w, r, m.a) }
}
func (m *passwordMethod) CallbackRoute() (string, http.HandlerFunc, bool) { return "", nil, false }

// ExtraRoutes mounts password recovery only when the method is present.
func (m *passwordMethod) ExtraRoutes() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"GET /password-restore": func(w http.ResponseWriter, r *http.Request) { page_password_restore.PagePasswordRestore(w, r, m.a) },
		"GET /password-reset":   func(w http.ResponseWriter, r *http.Request) { page_password_reset.PagePasswordReset(w, r, m.a) },
		"POST /api/password-restore": func(w http.ResponseWriter, r *http.Request) {
			api_password_restore.ApiPasswordRestoreWithAuth(w, r, m.a)
		},
		"POST /api/password-reset": func(w http.ResponseWriter, r *http.Request) { api_password_reset.ApiPasswordResetWithAuth(w, r, m.a) },
	}
}

// authKnightMethod wraps the external AuthKnight flow.
type authKnightMethod struct {
	a types.AuthAuthKnightInterface
}

// NewAuthKnightMethod adapts AuthKnight login to types.Method.
func NewAuthKnightMethod(a types.AuthAuthKnightInterface) types.Method {
	return &authKnightMethod{a: a}
}

func (m *authKnightMethod) Name() string             { return "authknight" }
func (m *authKnightMethod) AlternativeLabel() string { return "Continue with AuthKnight" }
func (m *authKnightMethod) PageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { page_login.PageLogin(w, r, m.a) }
}
func (m *authKnightMethod) AjaxHandler() http.HandlerFunc { return nil } // none — external flow
func (m *authKnightMethod) CallbackRoute() (string, http.HandlerFunc, bool) {
	return "/authknight-callback", func(w http.ResponseWriter, r *http.Request) {
		api_authknight_callback.ApiAuthKnightCallbackWithAuth(w, r, m.a)
	}, true
}
