package links

import "testing"

func TestApiImpersonateStart(t *testing.T) {
	got := ApiImpersonateStart("http://localhost/auth")
	expected := "http://localhost/auth/api/impersonate/start"
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestApiImpersonateStartTrailingSlash(t *testing.T) {
	got := ApiImpersonateStart("http://localhost/auth/")
	expected := "http://localhost/auth/api/impersonate/start"
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestApiImpersonateStop(t *testing.T) {
	got := ApiImpersonateStop("http://localhost/auth")
	expected := "http://localhost/auth/api/impersonate/stop"
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestApiImpersonateStopTrailingSlash(t *testing.T) {
	got := ApiImpersonateStop("http://localhost/auth/")
	expected := "http://localhost/auth/api/impersonate/stop"
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestJoin(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		uri      string
		expected string
	}{
		{"no trailing slash", "http://localhost/auth", "login", "http://localhost/auth/login"},
		{"trailing slash", "http://localhost/auth/", "login", "http://localhost/auth/login"},
		{"empty endpoint", "", "login", "/login"},
		{"empty uri", "http://localhost/auth", "", "http://localhost/auth/"},
		{"both empty", "", "", "/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Join(tt.endpoint, tt.uri)
			if got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestApiLinks(t *testing.T) {
	endpoint := "http://localhost/auth"
	tests := []struct {
		name     string
		fn       func(string) string
		expected string
	}{
		{"ApiLogin", ApiLogin, "http://localhost/auth/api/login"},
		{"ApiLoginCodeVerify", ApiLoginCodeVerify, "http://localhost/auth/api/login-code-verify"},
		{"ApiLogout", ApiLogout, "http://localhost/auth/api/logout"},
		{"ApiRegister", ApiRegister, "http://localhost/auth/api/register"},
		{"ApiRegisterCodeVerify", ApiRegisterCodeVerify, "http://localhost/auth/api/register-code-verify"},
		{"ApiPasswordRestore", ApiPasswordRestore, "http://localhost/auth/api/restore-password"},
		{"ApiPasswordReset", ApiPasswordReset, "http://localhost/auth/api/reset-password"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn(endpoint)
			if got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestApiLinksTrailingSlash(t *testing.T) {
	endpoint := "http://localhost/auth/"
	tests := []struct {
		name     string
		fn       func(string) string
		expected string
	}{
		{"ApiLogin", ApiLogin, "http://localhost/auth/api/login"},
		{"ApiLoginCodeVerify", ApiLoginCodeVerify, "http://localhost/auth/api/login-code-verify"},
		{"ApiLogout", ApiLogout, "http://localhost/auth/api/logout"},
		{"ApiRegister", ApiRegister, "http://localhost/auth/api/register"},
		{"ApiRegisterCodeVerify", ApiRegisterCodeVerify, "http://localhost/auth/api/register-code-verify"},
		{"ApiPasswordRestore", ApiPasswordRestore, "http://localhost/auth/api/restore-password"},
		{"ApiPasswordReset", ApiPasswordReset, "http://localhost/auth/api/reset-password"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn(endpoint)
			if got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestWebLinks(t *testing.T) {
	endpoint := "http://localhost/auth"
	tests := []struct {
		name     string
		fn       func(string) string
		expected string
	}{
		{"Login", Login, "http://localhost/auth/login"},
		{"LoginCodeVerify", LoginCodeVerify, "http://localhost/auth/login-code-verify"},
		{"Logout", Logout, "http://localhost/auth/logout"},
		{"PasswordRestore", PasswordRestore, "http://localhost/auth/password-restore"},
		{"PasswordReset", PasswordReset, "http://localhost/auth/password-reset"},
		{"Register", Register, "http://localhost/auth/register"},
		{"RegisterCodeVerify", RegisterCodeVerify, "http://localhost/auth/register-code-verify"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn(endpoint)
			if got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestWebLinksTrailingSlash(t *testing.T) {
	endpoint := "http://localhost/auth/"
	tests := []struct {
		name     string
		fn       func(string) string
		expected string
	}{
		{"Login", Login, "http://localhost/auth/login"},
		{"LoginCodeVerify", LoginCodeVerify, "http://localhost/auth/login-code-verify"},
		{"Logout", Logout, "http://localhost/auth/logout"},
		{"PasswordRestore", PasswordRestore, "http://localhost/auth/password-restore"},
		{"PasswordReset", PasswordReset, "http://localhost/auth/password-reset"},
		{"Register", Register, "http://localhost/auth/register"},
		{"RegisterCodeVerify", RegisterCodeVerify, "http://localhost/auth/register-code-verify"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn(endpoint)
			if got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
