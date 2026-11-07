package channels

import (
	"context"
	"encoding/json"
	"net/url"
	"regexp"

	"github.com/nhymxu/kith-pms/internal/notify"
)

const discordContentLimit = 2000

// optional /v10-style version segment, then numeric webhook id and token
var discordPathRe = regexp.MustCompile(`^/api(/v[0-9]+)?/webhooks/[0-9]{5,25}/[A-Za-z0-9_-]{20,128}$`)

// Discord posts to an incoming webhook. The whole URL is the credential. Any
// host is accepted so Discord-compatible receivers work; the path shape is
// enforced and notify.HTTPClient vets the resolved address.
type Discord struct{}

type discordConfig struct {
	WebhookURL string `json:"webhook_url"`
}

func (Discord) Kind() string           { return "discord" }
func (Discord) SecretFields() []string { return []string{"webhook_url"} }
func (Discord) PublicFields() []string { return nil }

func (d Discord) Validate(cfg json.RawMessage) error {
	_, err := d.parse(cfg)

	return err
}

func (Discord) parse(cfg json.RawMessage) (string, error) {
	c, err := decode[discordConfig](cfg)
	if err != nil {
		return "", err
	}

	raw, err := parseBaseURL(c.WebhookURL)
	if err != nil {
		return "", err
	}

	if u, perr := url.Parse(raw); perr != nil || !discordPathRe.MatchString(u.Path) {
		return "", errInvalidConfig
	}

	return raw, nil
}

func (d Discord) Send(ctx context.Context, c *notify.HTTPClient, cfg json.RawMessage, m notify.Message) error {
	endpoint, err := d.parse(cfg)
	if err != nil {
		return badConfig()
	}

	text := "**" + m.Title + "**"
	if b := body(m); b != "" {
		text += "\n" + b
	}

	// parse: [] stops reminder text from pinging @everyone or roles.
	return postJSON(ctx, c, endpoint, map[string]any{
		"content":          truncate(text, discordContentLimit),
		"allowed_mentions": map[string]any{"parse": []string{}},
	}, nil)
}
