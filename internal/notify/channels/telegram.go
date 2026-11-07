package channels

import (
	"context"
	"encoding/json"
	"regexp"

	"github.com/nhymxu/kith-pms/internal/notify"
)

const (
	defaultTelegramAPI = "https://api.telegram.org"
	telegramTextLimit  = 4096
)

var (
	telegramTokenRe = regexp.MustCompile(`^[0-9]{3,20}:[A-Za-z0-9_-]{20,64}$`)
	// numeric user/group/supergroup id, or a public channel @username
	telegramChatRe = regexp.MustCompile(`^(-?[0-9]{1,20}|@[A-Za-z][A-Za-z0-9_]{4,31})$`)
)

// Telegram sends through the Bot API. The token travels in the request path,
// so it is never logged: errors reduce to fixed classes before leaving Send.
type Telegram struct{}

type telegramConfig struct {
	APIURL   string `json:"api_url"`
	BotToken string `json:"bot_token"`
	ChatID   string `json:"chat_id"`
}

func (Telegram) Kind() string           { return "telegram" }
func (Telegram) SecretFields() []string { return []string{"bot_token"} }
func (Telegram) PublicFields() []string { return []string{"api_url", "chat_id"} }

func (t Telegram) Validate(cfg json.RawMessage) error {
	_, err := t.parse(cfg)

	return err
}

func (Telegram) parse(cfg json.RawMessage) (telegramConfig, error) {
	c, err := decode[telegramConfig](cfg)
	if err != nil {
		return c, err
	}

	if c.APIURL == "" {
		c.APIURL = defaultTelegramAPI
	}

	if c.APIURL, err = parseBaseURL(c.APIURL); err != nil {
		return c, err
	}

	if !telegramTokenRe.MatchString(c.BotToken) || !telegramChatRe.MatchString(c.ChatID) {
		return c, errInvalidConfig
	}

	return c, nil
}

func (t Telegram) Send(ctx context.Context, c *notify.HTTPClient, cfg json.RawMessage, m notify.Message) error {
	v, err := t.parse(cfg)
	if err != nil {
		return badConfig()
	}

	text := m.Title
	if b := body(m); b != "" {
		text += "\n" + b
	}

	return postJSON(ctx, c, v.APIURL+"/bot"+v.BotToken+"/sendMessage", map[string]any{
		"chat_id":                  v.ChatID,
		"text":                     truncateUTF16(text, telegramTextLimit),
		"disable_web_page_preview": true,
	}, nil)
}
