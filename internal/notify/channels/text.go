// Package channels holds the concrete notification adapters (ntfy, Telegram,
// Discord, signed webhook). Each implements notify.Channel and sends only through the injected
// notify.HTTPClient.
package channels

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/nhymxu/kith-pms/internal/notify"
)

var errInvalidConfig = errors.New("channels: invalid config")

// Register adds every adapter in this package to reg.
func Register(reg *notify.Registry, vapid VAPIDKeys) {
	reg.Register(WebPush{Keys: vapid})
	reg.Register(NTFY{})
	reg.Register(Telegram{})
	reg.Register(Discord{})
	reg.Register(Webhook{})
}

// body renders the reminder as plain text: person, due time, notes.
// A midnight-UTC due time marks a date-only reminder, so no clock is shown.
func body(m notify.Message) string {
	var parts []string

	if m.PersonName != "" {
		parts = append(parts, m.PersonName)
	}

	loc := m.Location
	if loc == nil {
		loc = time.UTC
	}

	if !m.DueAt.IsZero() {
		layout := "2006-01-02 15:04"
		if u := m.DueAt.UTC(); u.Hour() == 0 && u.Minute() == 0 && u.Second() == 0 && u.Nanosecond() == 0 {
			layout, loc = "2006-01-02", time.UTC
		}

		parts = append(parts, m.DueAt.In(loc).Format(layout))
	}

	if m.Notes != "" {
		parts = append(parts, m.Notes)
	}

	return strings.Join(parts, "\n")
}

// truncate caps s at max runes, ending in an ellipsis when cut.
func truncate(s string, limit int) string {
	r := []rune(s)
	if len(r) <= limit {
		return s
	}

	return string(r[:limit-1]) + "…"
}

// truncateBytes caps s at max UTF-8 bytes on a rune boundary, adding an ellipsis when cut.
func truncateBytes(s string, limit int) string {
	if len(s) <= limit {
		return s
	}

	const ellipsis = "…"

	cut := limit - len(ellipsis)
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}

	return s[:cut] + ellipsis
}

// truncateUTF16 caps s at max UTF-16 code units, which is how Telegram counts text length.
func truncateUTF16(s string, limit int) string {
	n := 0

	for i, r := range s {
		w := utf16.RuneLen(r)
		if n+w > limit-1 {
			if n+w <= limit && i+len(string(r)) == len(s) {
				return s
			}

			return s[:i] + "…"
		}

		n += w
	}

	return s
}

// parseBaseURL accepts an http(s) URL with a host and no credentials, query or
// fragment, and returns it without a trailing slash. Reachability of the host
// is enforced at dial time by notify.HTTPClient.
func parseBaseURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" ||
		u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Opaque != "" {
		return "", errInvalidConfig
	}

	return strings.TrimRight(u.String(), "/"), nil
}

func postJSON(
	ctx context.Context, c *notify.HTTPClient, endpoint string, payload any, header http.Header,
) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return &notify.SendError{Class: notify.ClassInternal, Permanent: true}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(b))
	if err != nil {
		return &notify.SendError{Class: notify.ClassInvalidConfig, Permanent: true}
	}

	req.Header.Set("Content-Type", "application/json")

	for k, v := range header {
		req.Header[k] = v
	}

	return c.Do(ctx, req)
}

func asSendError(err error, target **notify.SendError) bool { return errors.As(err, target) }

func decode[T any](cfg json.RawMessage) (T, error) {
	var v T

	if err := json.Unmarshal(cfg, &v); err != nil {
		return v, errInvalidConfig
	}

	return v, nil
}

func badConfig() error { return &notify.SendError{Class: notify.ClassInvalidConfig, Permanent: true} }
