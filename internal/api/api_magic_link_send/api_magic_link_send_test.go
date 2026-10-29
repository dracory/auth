package api_magic_link_send

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/auth/utils"
)

func makePost(t *testing.T, path string, values url.Values) (*httptest.ResponseRecorder, *http.Request) {
	body := strings.NewReader(values.Encode())
	req, err := http.NewRequest("POST", path, body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return httptest.NewRecorder(), req
}

func TestMagicLinkSendSuccess(t *testing.T) {
	stored := map[string]string{}
	var sentBody string
	deps := Dependencies{
		TemporaryKeyGet: func(k string) (string, error) { return "", errors.New("not found") },
		TemporaryKeySet: func(k, v string, e int) error { stored[k] = v; return nil },
		LinkMagicLinkCallback: func(token string) string {
			return "/auth/magiclink-callback?token=" + token
		},
		EmailSend: func(ctx context.Context, email, subject, body string) error {
			sentBody = body
			return nil
		},
	}

	recorder, req := makePost(t, "/api/magic-link-send", url.Values{"email": {"u@x.com"}})
	ApiMagicLinkSend(recorder, req, deps)

	if !strings.Contains(recorder.Body.String(), `"status":"success"`) {
		t.Fatalf("expected success, got %q", recorder.Body.String())
	}
	if len(stored) == 0 {
		t.Fatal("expected token stored")
	}
	for k, v := range stored {
		if !strings.HasPrefix(k, utils.MagicLinkKeyPrefix) {
			continue // throttle counter writes share the store
		}
		var p utils.MagicLinkPayload
		if json.Unmarshal([]byte(v), &p) != nil || p.Email != "u@x.com" || p.Token == "" {
			t.Fatalf("bad payload %q", v)
		}
	}
	if !strings.Contains(sentBody, "magiclink-callback?token=") {
		t.Fatalf("expected callback URL in email, got %q", sentBody)
	}
}

func TestMagicLinkSendInvalidEmail(t *testing.T) {
	stored := map[string]string{}
	deps := Dependencies{
		TemporaryKeySet: func(k, v string, e int) error { stored[k] = v; return nil },
		EmailSend: func(ctx context.Context, email, subject, body string) error {
			t.Fatal("must not send")
			return nil
		},
	}
	recorder, req := makePost(t, "/api/magic-link-send", url.Values{"email": {"nope"}})
	ApiMagicLinkSend(recorder, req, deps)
	if strings.Contains(recorder.Body.String(), `"status":"success"`) {
		t.Fatalf("expected error, got %q", recorder.Body.String())
	}
	if len(stored) != 0 {
		t.Fatal("nothing should be stored")
	}
}

func TestMagicLinkSendThrottle(t *testing.T) {
	deps := Dependencies{
		TemporaryKeyGet: func(k string) (string, error) { return "3", nil },
		TemporaryKeySet: func(k, v string, e int) error { return nil },
	}
	recorder, req := makePost(t, "/api/magic-link-send", url.Values{"email": {"u@x.com"}})
	ApiMagicLinkSend(recorder, req, deps)
	if !strings.Contains(recorder.Body.String(), "Too many codes requested") {
		t.Fatalf("expected throttle message, got %q", recorder.Body.String())
	}
}

func TestMagicLinkSendDropsExternalReturn(t *testing.T) {
	var payload utils.MagicLinkPayload
	deps := Dependencies{
		TemporaryKeyGet: func(k string) (string, error) { return "", errors.New("nf") },
		TemporaryKeySet: func(k, v string, e int) error { _ = json.Unmarshal([]byte(v), &payload); return nil },
	}
	recorder, req := makePost(t, "/api/magic-link-send", url.Values{
		"email":  {"u@x.com"},
		"return": {"//evil.com"},
	})
	ApiMagicLinkSend(recorder, req, deps)
	if payload.Return != "" {
		t.Fatalf("expected external return dropped, got %q", payload.Return)
	}
}
