// Package notifytest holds the contract checks every registered channel kind
// must pass. Child channel packages call RunContract from their own tests.
package notifytest

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/nhymxu/kith-pms/internal/notify"
)

// Canary is a secret-looking value with characters that change under encoding.
const Canary = "CANARY-s3cr3t/+=&\"v4lue"

// CanaryForms returns the secret as it may appear after common encodings.
func CanaryForms(secret string) []string {
	j, _ := json.Marshal(secret)

	return []string{
		secret,
		url.QueryEscape(secret),
		url.PathEscape(secret),
		base64.StdEncoding.EncodeToString([]byte(secret)),
		base64.StdEncoding.EncodeToString([]byte("user:" + secret)),
		base64.RawURLEncoding.EncodeToString([]byte(secret)),
		strings.Trim(string(j), `"`),
	}
}

func AssertNoCanary(t *testing.T, where, text string) { assertNoSecret(t, where, text, Canary) }

func assertNoSecret(t *testing.T, where, text, secret string) {
	t.Helper()

	for _, f := range CanaryForms(secret) {
		if strings.Contains(text, f) {
			t.Errorf("%s leaks canary form %q", where, f)
		}
	}
}

// Fixture builds a valid config for a kind pointing at baseURL, with every
// secret field set to secret.
type Fixture func(baseURL, secret string) json.RawMessage

// RunContract checks, for each registered kind: PublicView never carries a
// secret, and a failing send (server echoing the canary in body and headers)
// yields no canary in the sanitized error.
func RunContract(t *testing.T, reg *notify.Registry, fixtures map[string]Fixture) {
	t.Helper()
	RunContractWith(t, reg, fixtures, Canary)
}

// RunContractWith is RunContract with a caller-chosen secret, for kinds whose
// validators only accept a restricted alphabet (tokens, topics).
func RunContractWith(t *testing.T, reg *notify.Registry, fixtures map[string]Fixture, secret string) {
	t.Helper()

	client := notify.NewHTTPClient()

	for _, kind := range reg.Kinds() {
		t.Run(kind, func(t *testing.T) {
			ch, _ := reg.Get(kind)

			fx, ok := fixtures[kind]
			if !ok {
				t.Fatalf("no contract fixture for kind %q", kind)
			}

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("X-Echo", secret)
				http.Error(w, strings.Join(CanaryForms(secret), " "), http.StatusInternalServerError)
			}))
			defer srv.Close()

			cfg := fx(srv.URL, secret)
			if err := ch.Validate(cfg); err != nil {
				t.Fatalf("fixture invalid: %v", err)
			}

			view, _ := json.Marshal(notify.PublicView(ch, cfg))
			assertNoSecret(t, "PublicView", string(view), secret)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := ch.Send(ctx, client, cfg, notify.Message{ReminderID: 1, Title: "t", DueAt: time.Now()})
			if err == nil {
				return
			}

			se := notify.SanitizeErr(err)
			out, _ := json.Marshal(se)
			assertNoSecret(t, "SendError", se.Error()+string(out), secret)
		})
	}
}
