package channels

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"

	"github.com/nhymxu/kith-pms/internal/notify"
)

const defaultNTFYServer = "https://ntfy.sh"

var (
	ntfyTopicRe = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)
	ntfyTokenRe = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,256}$`)
)

// NTFY publishes to an ntfy topic. The topic is a secret: on a public server
// it is the only access control.
type NTFY struct{}

type ntfyConfig struct {
	ServerURL string `json:"server_url"`
	Topic     string `json:"topic"`
	Token     string `json:"token"`
}

func (NTFY) Kind() string           { return "ntfy" }
func (NTFY) SecretFields() []string { return []string{"topic", "token"} }
func (NTFY) PublicFields() []string { return []string{"server_url"} }

func (n NTFY) Validate(cfg json.RawMessage) error {
	_, err := n.parse(cfg)

	return err
}

func (NTFY) parse(cfg json.RawMessage) (ntfyConfig, error) {
	c, err := decode[ntfyConfig](cfg)
	if err != nil {
		return c, err
	}

	if c.ServerURL == "" {
		c.ServerURL = defaultNTFYServer
	}

	if c.ServerURL, err = parseBaseURL(c.ServerURL); err != nil {
		return c, err
	}

	// a pasted topic URL would publish the envelope as text to that topic
	if u, perr := url.Parse(c.ServerURL); perr != nil || u.Path != "" {
		return c, errInvalidConfig
	}

	if !ntfyTopicRe.MatchString(c.Topic) || (c.Token != "" && !ntfyTokenRe.MatchString(c.Token)) {
		return c, errInvalidConfig
	}

	return c, nil
}

func (n NTFY) Send(ctx context.Context, c *notify.HTTPClient, cfg json.RawMessage, m notify.Message) error {
	v, err := n.parse(cfg)
	if err != nil {
		return badConfig()
	}

	header := http.Header{}
	if v.Token != "" {
		header.Set("Authorization", "Bearer "+v.Token)
	}

	// JSON publish keeps non-ASCII titles intact, which a Title header would not.
	return postJSON(ctx, c, v.ServerURL, map[string]string{
		"topic":   v.Topic,
		"title":   truncate(m.Title, 250),
		"message": truncateBytes(body(m), 3900),
	}, header)
}
