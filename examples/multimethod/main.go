// Multi-method example: mounts username/password, email OTP, and
// magic-link login on a single mux via auth.NewMultiMethodAuth. Also
// demonstrates branding (BrandName/Tagline), sibling-method cross-links
// (LoginAlternatives), and a host FuncLayout that embeds the built-in
// pages inside a site chrome.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dracory/auth"
	"github.com/dracory/auth/types"
)

// user is the in-memory account record shared by all methods.
type user struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
	Password  string
}

// memoryStore is a demo store backing both auth implementations. Real
// apps would adapt their DB here.
type memoryStore struct {
	mu        sync.Mutex
	users     map[string]user    // email -> user
	tokenUser map[string]string  // auth token -> userID
	tempKeys  map[string]tempKey // temp key -> {value, expiry}
	nextID    int
}

type tempKey struct {
	value   string
	expires time.Time
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		users:     map[string]user{},
		tokenUser: map[string]string{},
		tempKeys:  map[string]tempKey{},
	}
}

func (s *memoryStore) userLogin(_ context.Context, username, password string, _ types.UserAuthOptions) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[username]
	if !ok || u.Password != password {
		return "", errors.New("invalid credentials")
	}
	return u.ID, nil
}

func (s *memoryStore) userRegister(_ context.Context, username, _, firstName, lastName string, _ types.UserAuthOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.users[username]; exists {
		return errors.New("user already exists")
	}
	s.nextID++
	s.users[username] = user{ID: fmt.Sprint(s.nextID), Email: username, FirstName: firstName, LastName: lastName}
	return nil
}

func (s *memoryStore) userFindByUsername(_ context.Context, username, _, _ string, _ types.UserAuthOptions) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if u, ok := s.users[username]; ok {
		return u.ID, nil
	}
	return "", nil
}

func (s *memoryStore) userFindByEmail(_ context.Context, email string, _ types.UserAuthOptions) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if u, ok := s.users[email]; ok {
		return u.ID, nil
	}
	return "", nil
}

func (s *memoryStore) userPasswordChange(_ context.Context, username, newPassword string, _ types.UserAuthOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[username]
	if !ok {
		return errors.New("user not found")
	}
	u.Password = newPassword
	s.users[username] = u
	return nil
}

func (s *memoryStore) storeAuthToken(_ context.Context, token, userID string, _ types.UserAuthOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokenUser[token] = userID
	return nil
}

func (s *memoryStore) findByAuthToken(_ context.Context, token string, _ types.UserAuthOptions) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tokenUser[token], nil
}

func (s *memoryStore) logout(_ context.Context, userID string, _ types.UserAuthOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for t, id := range s.tokenUser {
		if id == userID {
			delete(s.tokenUser, t)
		}
	}
	return nil
}

func (s *memoryStore) tempKeySet(key, value string, expiresSeconds int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tempKeys[key] = tempKey{value: value, expires: time.Now().Add(time.Duration(expiresSeconds) * time.Second)}
	return nil
}

func (s *memoryStore) tempKeyGet(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	k, ok := s.tempKeys[key]
	if !ok || time.Now().After(k.expires) {
		return "", errors.New("key not found or expired")
	}
	return k.value, nil
}

func (s *memoryStore) displayName(userID string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, u := range s.users {
		if u.ID == userID {
			return strings.TrimSpace(u.FirstName + " " + u.LastName)
		}
	}
	return "User"
}

// emailSend logs emails to the console — run a dev SMTP on :1025
// (mailhog/mailpit) to receive real OTPs and magic links.
func emailSend(_ context.Context, email, subject, body string) error {
	log.Printf("[email] to=%s subject=%s\n%s", email, subject, body)
	return nil
}

// siteLayout embeds the built-in auth body inside a demo site chrome.
// Returning "" would fall back to the package's own document.
func siteLayout(_ http.ResponseWriter, _ *http.Request, title, body string, options types.LayoutOptions) string {
	links := ""
	for _, u := range options.StyleURLs {
		links += "<link rel=\"stylesheet\" href=\"" + u + "\" />"
	}
	styles := ""
	for _, s := range options.Styles {
		styles += "<style>" + s + "</style>"
	}
	scripts := ""
	for _, u := range options.ScriptURLs {
		scripts += "<script src=\"" + u + "\"></script>"
	}
	for _, s := range options.Scripts {
		scripts += "<script>" + s + "</script>"
	}
	return `<!DOCTYPE html><html><head><meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<title>` + title + ` | Acme Corp</title>` + links + styles + `</head>
<body><nav class="navbar navbar-dark bg-dark px-3">
	<a class="navbar-brand fw-bold" href="/">Acme Corp</a>
</nav>` + body + scripts + `</body></html>`
}

func main() {
	store := newMemoryStore()

	// Seed a demo account.
	_ = store.userRegister(context.Background(), "demo@example.com", "", "Demo", "User", types.UserAuthOptions{})
	_ = store.userPasswordChange(context.Background(), "demo@example.com", "password123", types.UserAuthOptions{})

	shared := types.ConfigShared{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/dashboard",
		UseCookies:           true,
		EnableRegistration:   true,
		BrandName:            "ACME",
		BrandTagline:         "EXAMPLE PRODUCTS",

		// Host layout — the built-in pages render inside this chrome.
		// Comment out to see the package's standalone document instead.
		FuncLayout: siteLayout,

		FuncTemporaryKeyGet:     store.tempKeyGet,
		FuncTemporaryKeySet:     store.tempKeySet,
		FuncUserStoreAuthToken:  store.storeAuthToken,
		FuncUserFindByAuthToken: store.findByAuthToken,
		FuncUserLogout:          store.logout,
	}

	// Password method — owns login, register, restore/reset routes.
	passwordAuth, err := auth.NewUsernameAndPasswordAuth(types.ConfigUsernameAndPassword{
		ConfigShared:           shared,
		FuncUserFindByUsername: store.userFindByUsername,
		FuncUserLogin:          store.userLogin,
		FuncEmailSend:          emailSend,
		FuncUserRegister:       store.userRegister,
		FuncUserPasswordChange: store.userPasswordChange,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Passwordless impl — shared by the OTP and magic-link methods.
	passwordlessAuth, err := auth.NewPasswordlessAuth(types.ConfigPasswordless{
		ConfigShared:        shared,
		FuncUserFindByEmail: store.userFindByEmail,
		FuncEmailSend:       emailSend,
		FuncUserRegister: func(ctx context.Context, email, firstName, lastName string, opts types.UserAuthOptions) error {
			return store.userRegister(ctx, email, "", firstName, lastName, opts)
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	methods := []types.Method{
		auth.NewPasswordMethod(passwordAuth, nil),
		auth.NewOTPMethod(passwordlessAuth),
		auth.NewMagicLinkMethodFromAuth(passwordlessAuth, nil,
			func(token string) string {
				return "/auth/magiclink-callback?token=" + token
			}),
	}

	multi, err := auth.NewMultiMethodAuth("/auth", methods)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/auth/", multi.Handler())

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<h1>Home</h1><p><a href="/auth/login">Login</a> (demo@example.com / password123)</p>`)
	})

	mux.Handle("/dashboard", passwordAuth.WebAuthOrRedirectMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := passwordAuth.GetCurrentUserID(r)
		fmt.Fprintf(w, `<h1>Dashboard</h1><p>Welcome, %s (id: %s)!</p><p><a href="/auth/logout">Logout</a></p>`, store.displayName(userID), userID)
	})))

	fmt.Println("Multi-method example running on http://localhost:8084")
	fmt.Println("  /auth/login            - password")
	fmt.Println("  /auth/login-otp        - email code")
	fmt.Println("  /auth/login-magiclink  - magic link")
	fmt.Println("(emails are logged to this console; send a real SMTP server on :1025 to receive them)")
	if err := http.ListenAndServe(":8084", mux); err != nil {
		log.Fatal(err)
	}
}
