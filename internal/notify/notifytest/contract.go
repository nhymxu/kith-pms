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

func AssertNoCanary(t *testing.T, where, text string) {
	t.Helper()

	for _, f := range CanaryForms(Canary) {
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

	client := notify.NewHTTPClient()

	for _, kind := range reg.Kinds() {
		t.Run(kind, func(t *testing.T) {
			ch, _ := reg.Get(kind)

			fx, ok := fixtures[kind]
			if !ok {
				t.Fatalf("no contract fixture for kind %q", kind)
			}

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("X-Echo", Canary)
				http.Error(w, strings.Join(CanaryForms(Canary), " "), http.StatusInternalServerError)
			}))
			defer srv.Close()

			cfg := fx(srv.URL, Canary)
			if err := ch.Validate(cfg); err != nil {
				t.Fatalf("fixture invalid: %v", err)
			}

			view, _ := json.Marshal(notify.PublicView(ch, cfg))
			AssertNoCanary(t, "PublicView", string(view))

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := ch.Send(ctx, client, cfg, notify.Message{ReminderID: 1, Title: "t", DueAt: time.Now()})
			if err == nil {
				return
			}

			se := notify.SanitizeErr(err)
			out, _ := json.Marshal(se)
			AssertNoCanary(t, "SendError", se.Error()+string(out))
		})
	}
}
