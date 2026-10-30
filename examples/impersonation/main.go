// Impersonation example: demonstrates administrator "login as" feature.
// Admin users can impersonate other users with security checks and session swapping.
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

type userAccount struct {
	ID        string
	Username  string
	Password  string
	FirstName string
	LastName  string
	IsAdmin   bool
}

type impersonationStore struct {
	mu            sync.Mutex
	users         map[string]*userAccount // username -> user
	usersByID     map[string]*userAccount // id -> user
	sessions      map[string]string       // token -> userID
	impersonating map[string]string       // adminID -> targetID
}

var store = &impersonationStore{
	users:         make(map[string]*userAccount),
	usersByID:     make(map[string]*userAccount),
	sessions:      make(map[string]string),
	impersonating: make(map[string]string),
}

func (s *impersonationStore) userLogin(_ context.Context, username, password string, _ types.UserAuthOptions) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.users[username]
	if !ok || u.Password != password {
		return "", errors.New("invalid username or password")
	}
	return u.ID, nil
}

func (s *impersonationStore) userFindByUsername(_ context.Context, username, _, _ string, _ types.UserAuthOptions) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if u, ok := s.users[username]; ok {
		return u.ID, nil
	}
	return "", errors.New("user not found")
}

func (s *impersonationStore) storeAuthToken(_ context.Context, token, userID string, _ types.UserAuthOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[token] = userID
	return nil
}

func (s *impersonationStore) findByAuthToken(_ context.Context, token string, _ types.UserAuthOptions) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	userID, ok := s.sessions[token]
	if !ok {
		return "", errors.New("invalid session token")
	}
	return userID, nil
}

func (s *impersonationStore) logout(_ context.Context, userID string, _ types.UserAuthOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for token, id := range s.sessions {
		if id == userID {
			delete(s.sessions, token)
		}
	}
	delete(s.impersonating, userID)
	return nil
}

func (s *impersonationStore) tempKeySet(key, value string, _ int) error {
	return nil
}

func (s *impersonationStore) tempKeyGet(key string) (string, error) {
	return "", errors.New("not found")
}

func (s *impersonationStore) canImpersonate(_ context.Context, adminUserID, targetUserID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	admin, ok := s.usersByID[adminUserID]
	if !ok || !admin.IsAdmin {
		return false, nil
	}

	_, targetExists := s.usersByID[targetUserID]
	return targetExists, nil
}

func (s *impersonationStore) impersonationStart(_ context.Context, adminUserID, targetUserID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.impersonating[adminUserID] = targetUserID
	return nil
}

func (s *impersonationStore) impersonationStop(_ context.Context, adminUserID, targetUserID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.impersonating, adminUserID)
	return nil
}

func (s *impersonationStore) displayName(userID string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if u, ok := s.usersByID[userID]; ok {
		full := strings.TrimSpace(u.FirstName + " " + u.LastName)
		if full != "" {
			return full
		}
		return u.Username
	}
	return userID
}

func emailSend(_ context.Context, userID, subject, body string) error {
	log.Printf("[impersonation email] to=%s subject=%s\n", userID, subject)
	return nil
}

func main() {
	// Seed admin and normal user
	adminUser := &userAccount{ID: "admin-1", Username: "admin@example.com", Password: "adminpassword", FirstName: "Admin", LastName: "User", IsAdmin: true}
	targetUser := &userAccount{ID: "user-2", Username: "john@example.com", Password: "userpassword", FirstName: "John", LastName: "Doe", IsAdmin: false}

	store.users[adminUser.Username] = adminUser
	store.usersByID[adminUser.ID] = adminUser
	store.users[targetUser.Username] = targetUser
	store.usersByID[targetUser.ID] = targetUser

	authInstance, err := auth.NewUsernameAndPasswordAuth(types.ConfigUsernameAndPassword{
		ConfigShared: types.ConfigShared{
			Endpoint:             "/auth",
			UrlRedirectOnSuccess: "/dashboard",
			UseCookies:           true,
			EnableRegistration:   false,
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
		ConfigImpersonation: types.ConfigImpersonation{
			EnableImpersonation:    true,
			FuncCanImpersonate:     store.canImpersonate,
			FuncImpersonationStart: store.impersonationStart,
			FuncImpersonationStop:  store.impersonationStop,
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
		fmt.Fprint(w, `<h1>Impersonation Demo</h1><p><a href="/auth/login">Login</a> (admin@example.com / adminpassword)</p>`)
	})

	mux.Handle("/dashboard", authInstance.WebAuthOrRedirectMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := authInstance.GetCurrentUserID(r)
		isImpersonating := auth.IsImpersonating(r)
		realAdminID := auth.GetImpersonatorUserID(r)

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<h1>Dashboard</h1><p>Currently logged in as: <strong>%s</strong> (ID: %s)</p>`, store.displayName(userID), userID)

		if isImpersonating {
			fmt.Fprintf(w, `<p style="color: red;">⚠️ You are impersonating this user! Real Admin ID: %s</p>`, realAdminID)
			fmt.Fprintf(w, `<form action="%s" method="POST"><button type="submit">Stop Impersonating</button></form>`, authInstance.LinkApiImpersonateStop())
		} else if userID == "admin-1" {
			fmt.Fprintf(w, `<p>Admin Control Panel</p>`)
			fmt.Fprintf(w, `<form action="%s" method="POST"><input type="hidden" name="user_id" value="user-2"><button type="submit">Impersonate John Doe (user-2)</button></form>`, authInstance.LinkApiImpersonateStart())
		}

		fmt.Fprintf(w, `<p><a href="%s">Logout</a></p>`, authInstance.LinkLogout())
	})))

	fmt.Println("Impersonation auth example running on http://localhost:8086")
	if err := http.ListenAndServe(":8086", mux); err != nil {
		log.Fatal(err)
	}
}
