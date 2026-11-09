package channels

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/nhymxu/kith-pms/internal/notify"
)

const (
	webhookTimestampHeader = "X-Kith-Timestamp"
	webhookSignatureHeader = "X-Kith-Signature"
)

var webhookSecretRe = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)

// Webhook posts a JSON body to a user-owned URL, signed with HMAC-SHA256 over
// "<timestamp>.<raw body>" so the receiver can check origin, integrity and age.
type Webhook struct{}

type webhookConfig struct {
	URL    string `json:"url"`
	Secret string `json:"secret"`
}

func (Webhook) Kind() string           { return "webhook" }
func (Webhook) SecretFields() []string { return []string{"secret"} }
func (Webhook) PublicFields() []string { return []string{"url"} }

func (w Webhook) Validate(cfg json.RawMessage) error {
	_, err := w.parse(cfg)

	return err
}

func (Webhook) parse(cfg json.RawMessage) (webhookConfig, error) {
	c, err := decode[webhookConfig](cfg)
	if err != nil {
		return c, err
	}

	// the URL is returned by the API, so it must not carry credentials
	u, err := url.Parse(c.URL)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" ||
		u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Opaque != "" {
		return c, errInvalidConfig
	}

	if !webhookSecretRe.MatchString(c.Secret) {
		return c, errInvalidConfig
	}

	return c, nil
}

func webhookSignature(secret string, ts int64, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strconv.FormatInt(ts, 10) + "."))
	mac.Write(body)

	return "v1=" + hex.EncodeToString(mac.Sum(nil))
}

func (w Webhook) Send(ctx context.Context, c *notify.HTTPClient, cfg json.RawMessage, m notify.Message) error {
	v, err := w.parse(cfg)
	if err != nil {
		return badConfig()
	}

	loc := m.Location
	if loc == nil {
		loc = time.UTC
	}

	payload, err := json.Marshal(map[string]any{
		"id":       m.ReminderID,
		"title":    m.Title,
		"notes":    m.Notes,
		"person":   m.PersonName,
		"due_at":   m.DueAt.UTC().Format(time.RFC3339),
		"timezone": loc.String(),
	})
	if err != nil {
		return &notify.SendError{Class: notify.ClassInternal, Permanent: true}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.URL, bytes.NewReader(payload))
	if err != nil {
		return badConfig()
	}

	ts := time.Now().Unix()

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(webhookTimestampHeader, strconv.FormatInt(ts, 10))
	req.Header.Set(webhookSignatureHeader, webhookSignature(v.Secret, ts, payload))

	return c.Do(ctx, req)
}
