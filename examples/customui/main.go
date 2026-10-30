// Custom UI example: demonstrates branding, custom CSS, and site layout embedding.
// Auth UI pages are rendered seamlessly inside the application's navbar, layout, and styling.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/dracory/auth"
	"github.com/dracory/auth/types"
)

type customUIUser struct {
	ID        string
	Username  string
	Password  string
	FirstName string
	LastName  string
}

type customUIMemoryStore struct {
	mu        sync.Mutex
	users     map[string]*customUIUser
	sessions  map[string]string
	tempKeys  map[string]string
	nextIndex int
}

var store = &customUIMemoryStore{
	users:     make(map[string]*customUIUser),
	sessions:  make(map[string]string),
	tempKeys:  make(map[string]string),
	nextIndex: 1,
}

func (s *customUIMemoryStore) userLogin(_ context.Context, username, password string, _ types.UserAuthOptions) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.users[username]
	if !ok || u.Password != password {
		return "", errors.New("invalid credentials")
	}
	return u.ID, nil
}

func (s *customUIMemoryStore) userFindByUsername(_ context.Context, username, _, _ string, _ types.UserAuthOptions) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if u, ok := s.users[username]; ok {
		return u.ID, nil
	}
	return "", errors.New("user not found")
}

func (s *customUIMemoryStore) storeAuthToken(_ context.Context, token, userID string, _ types.UserAuthOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[token] = userID
	return nil
}

func (s *customUIMemoryStore) findByAuthToken(_ context.Context, token string, _ types.UserAuthOptions) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, ok := s.sessions[token]
	if !ok {
		return "", errors.New("invalid token")
	}
	return id, nil
}

func (s *customUIMemoryStore) logout(_ context.Context, userID string, _ types.UserAuthOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for t, id := range s.sessions {
		if id == userID {
			delete(s.sessions, t)
		}
	}
	return nil
}

func (s *customUIMemoryStore) tempKeySet(key, value string, _ int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tempKeys[key] = value
	return nil
}

func (s *customUIMemoryStore) tempKeyGet(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.tempKeys[key]
	if !ok {
		return "", errors.New("key not found")
	}
	return v, nil
}

func (s *customUIMemoryStore) displayName(userID string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range s.users {
		if u.ID == userID {
			full := strings.TrimSpace(u.FirstName + " " + u.LastName)
			if full != "" {
				return full
			}
			return u.Username
		}
	}
	return userID
}

func emailSend(_ context.Context, userID, subject, body string) error {
	log.Printf("[customui email] to=%s subject=%s\n", userID, subject)
	return nil
}

// customLayout renders the entire page document with custom navbar, header, and footer.
func customLayout(_ http.ResponseWriter, _ *http.Request, title, body string, options types.LayoutOptions) string {
	var styles strings.Builder
	for _, u := range options.StyleURLs {
		styles.WriteString(fmt.Sprintf(`<link rel="stylesheet" href="%s">`, u))
	}
	for _, s := range options.Styles {
		styles.WriteString(fmt.Sprintf(`<style>%s</style>`, s))
	}

	var scripts strings.Builder
	for _, u := range options.ScriptURLs {
		scripts.WriteString(fmt.Sprintf(`<script src="%s"></script>`, u))
	}
	for _, s := range options.Scripts {
		scripts.WriteString(s)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - Aura Platform</title>
    %s
    <style>
        body { background-color: #0f172a; color: #f8fafc; font-family: system-ui, -apple-system, sans-serif; }
        .navbar { background-color: #1e293b; border-bottom: 1px solid #334155; padding: 1rem 2rem; display: flex; justify-content: space-between; align-items: center; }
        .brand { font-size: 1.25rem; font-weight: 700; color: #38bdf8; text-decoration: none; }
        .card { background-color: #1e293b !important; border: 1px solid #334155 !important; color: #f8fafc !important; }
        .btn-primary { background-color: #0284c7 !important; border-color: #0284c7 !important; }
        .btn-primary:hover { background-color: #0369a1 !important; }
        .form-control { background-color: #0f172a !important; border-color: #334155 !important; color: #f8fafc !important; }
        footer { text-align: center; padding: 2rem; color: #64748b; font-size: 0.875rem; margin-top: 3rem; }
    </style>
</head>
<body>
    <nav class="navbar">
        <a class="brand" href="/">✨ AURA PLATFORM</a>
        <div><a href="/auth/login" style="color: #94a3b8; text-decoration: none;">Login</a></div>
    </nav>
    <main class="container py-5">
        %s
    </main>
    <footer>&copy; 2025 Aura Security Systems. All rights reserved.</footer>
    %s
</body>
</html>`, title, styles.String(), body, scripts.String())
}

func main() {
	// Seed demo user
	demoUser := &customUIUser{ID: "usr-1", Username: "user@aura.local", Password: "aurapassword", FirstName: "Aura", LastName: "Member"}
	store.users[demoUser.Username] = demoUser

	authInstance, err := auth.NewUsernameAndPasswordAuth(types.ConfigUsernameAndPassword{
		ConfigShared: types.ConfigShared{
			Endpoint:             "/auth",
			UrlRedirectOnSuccess: "/dashboard",
			UseCookies:           true,
			EnableRegistration:   false,
			BrandName:            "AURA PLATFORM",
			BrandTagline:         "Next-Generation Identity Infrastructure",
			FuncLayout:           customLayout,
			CustomCSS:            "body { background: #0f172a !important; }",
			CookieConfig: &types.CookieConfig{
				Secure:   types.CookieInsecure,
				HttpOnly: types.CookieHttpOnly,
				SameSite: http.SameSiteLaxMode,
				Path:     "/",
				MaxAge:   2 * 60 * 60,
			},

			FuncUserFindByAuthToken: store.findByAuthToken,
			FuncUserLogout:          store.logout,
			FuncUserStoreAuthToken:  store.storeAuthToken,
			FuncTemporaryKeyGet:     store.tempKeyGet,
			FuncTemporaryKeySet:     store.tempKeySet,
		},

		FuncEmailSend:          emailSend,
		FuncUserFindByUsername: store.userFindByUsername,
		FuncUserLogin:          store.userLogin,
	})
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/auth/", authInstance.Router().ServeHTTP)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<h1>Aura Platform</h1><p><a href="/auth/login">Login with Custom UI</a> (user@aura.local / aurapassword)</p>`)
	})

	mux.Handle("/dashboard", authInstance.WebAuthOrRedirectMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := authInstance.GetCurrentUserID(r)
		fmt.Fprintf(w, `<h1>Aura Dashboard</h1><p>Welcome, %s (ID: %s)</p><p><a href="/auth/logout">Logout</a></p>`, store.displayName(userID), userID)
	})))

	fmt.Println("Custom UI example running on http://localhost:8087")
	if err := http.ListenAndServe(":8087", mux); err != nil {
		log.Fatal(err)
	}
}
