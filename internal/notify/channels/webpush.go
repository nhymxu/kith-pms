package channels

import (
	"bytes"
	"context"
	"crypto/ecdh"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	webpush "github.com/SherClockHolmes/webpush-go"

	"github.com/nhymxu/kith-pms/internal/notify"
)

// vapidSubscriber is the VAPID "sub" claim; push services use it as a contact for the sender.
const (
	vapidSubscriber = "https://github.com/nhymxu/kith-pms"
	pushTTLSeconds  = 24 * 60 * 60
	pushTitleRunes  = 100
	pushBodyBytes   = 1500
	pushPayloadMax  = 3800
)

// VAPIDKeys supplies the server's VAPID key pair.
type VAPIDKeys interface {
	Keys(ctx context.Context) (priv, pub string, err error)
}

// WebPush delivers to one browser subscription. Every config field is secret:
// the endpoint is a capability URL for that browser.
type WebPush struct{ Keys VAPIDKeys }

type webPushConfig struct {
	Endpoint string `json:"endpoint"`
	P256dh   string `json:"p256dh"`
	Auth     string `json:"auth"`
}

func (WebPush) Kind() string           { return notify.KindWebPush }
func (WebPush) SecretFields() []string { return []string{"endpoint", "p256dh", "auth"} }
func (WebPush) PublicFields() []string { return nil }

func (w WebPush) Validate(cfg json.RawMessage) error {
	_, err := w.parse(cfg)

	return err
}

func decodeB64URL(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(strings.TrimRight(s, "="))
}

func (WebPush) parse(cfg json.RawMessage) (webPushConfig, error) {
	c, err := decode[webPushConfig](cfg)
	if err != nil {
		return c, err
	}

	// push services may carry a token in the query (Windows), so only credentials and fragments are refused
	u, err := url.Parse(c.Endpoint)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" ||
		u.User != nil || u.Fragment != "" || u.Opaque != "" {
		return c, errInvalidConfig
	}

	p, err := decodeB64URL(c.P256dh)
	if err != nil {
		return c, errInvalidConfig
	}

	if _, err := ecdh.P256().NewPublicKey(p); err != nil {
		return c, errInvalidConfig
	}

	if a, err := decodeB64URL(c.Auth); err != nil || len(a) != 16 {
		return c, errInvalidConfig
	}

	return c, nil
}

// pushClient lets webpush-go build and encrypt the request while the send
// itself still goes through notify.HTTPClient (address guard, error classes).
type pushClient struct {
	c   *notify.HTTPClient
	err error
}

func (p *pushClient) Do(req *http.Request) (*http.Response, error) {
	p.err = p.c.Do(req.Context(), req)

	return &http.Response{StatusCode: http.StatusCreated, Body: io.NopCloser(bytes.NewReader(nil))}, nil
}

func (w WebPush) Send(ctx context.Context, c *notify.HTTPClient, cfg json.RawMessage, m notify.Message) error {
	v, err := w.parse(cfg)
	if err != nil {
		return badConfig()
	}

	priv, pub, err := w.Keys.Keys(ctx)
	if err != nil {
		return &notify.SendError{Class: notify.ClassInternal, Permanent: true}
	}

	payload, err := pushPayload(m)
	if err != nil {
		return &notify.SendError{Class: notify.ClassInternal, Permanent: true}
	}

	pc := &pushClient{c: c}

	_, err = webpush.SendNotificationWithContext(ctx, payload,
		&webpush.Subscription{Endpoint: v.Endpoint, Keys: webpush.Keys{P256dh: v.P256dh, Auth: v.Auth}},
		&webpush.Options{
			HTTPClient: pc, Subscriber: vapidSubscriber, VAPIDPublicKey: pub, VAPIDPrivateKey: priv,
			TTL: pushTTLSeconds, Urgency: webpush.UrgencyHigh,
		})
	if err != nil {
		return &notify.SendError{Class: notify.ClassInternal, Permanent: true}
	}

	return pc.result()
}

// pushPayload keeps the encoded JSON under the push record limit (webpush-go
// rejects ~4 KB); escaped characters can expand the body, so it shrinks until it fits.
func pushPayload(m notify.Message) ([]byte, error) {
	text := truncateBytes(body(m), pushBodyBytes)

	for {
		var buf bytes.Buffer

		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)

		if err := enc.Encode(
			map[string]string{"title": truncate(m.Title, pushTitleRunes), "body": text, "url": "/"},
		); err != nil {
			return nil, err
		}

		if buf.Len() <= pushPayloadMax || len(text) < 16 {
			return buf.Bytes(), nil
		}

		text = truncateBytes(text, len(text)/2)
	}
}

// result maps a 404/410 from the push service to ClassGone: the subscription is expired for good.
func (p *pushClient) result() error {
	var se *notify.SendError

	if p.err == nil {
		return nil
	}

	if asSendError(p.err, &se) && (se.Status == http.StatusNotFound || se.Status == http.StatusGone) {
		return &notify.SendError{Class: notify.ClassGone, Status: se.Status, Permanent: true}
	}

	return p.err
}
