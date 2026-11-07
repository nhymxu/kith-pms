package channels_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"unicode/utf16"

	"github.com/nhymxu/kith-pms/internal/notify"
	"github.com/nhymxu/kith-pms/internal/notify/channels"
	"github.com/nhymxu/kith-pms/internal/notify/notifytest"
)

// Alphabet every validator accepts, so the contract check exercises real configs.
const secret = "CANARYs3cr3tv4lue0123456789"

func j(v map[string]string) json.RawMessage {
	b, _ := json.Marshal(v)

	return b
}

func registry() *notify.Registry {
	reg := notify.NewRegistry()
	channels.Register(reg)

	return reg
}

func TestContract(t *testing.T) {
	notifytest.RunContractWith(t, registry(), map[string]notifytest.Fixture{
		"ntfy": func(base, s string) json.RawMessage {
			return j(map[string]string{"server_url": base, "topic": s, "token": s})
		},
		"telegram": func(base, s string) json.RawMessage {
			return j(map[string]string{"api_url": base, "bot_token": "123456:" + s, "chat_id": "-100123"})
		},
		"discord": func(base, s string) json.RawMessage {
			return j(map[string]string{"webhook_url": base + "/api/webhooks/123456789/" + s})
		},
	}, secret)
}

func TestValidate(t *testing.T) {
	tok := "123456:" + secret
	hook := "https://discord.com/api/webhooks/123456789/" + secret

	cases := []struct {
		kind string
		cfg  map[string]string
		ok   bool
	}{
		{"ntfy", map[string]string{"topic": "my-topic_1"}, true},
		{
			"ntfy",
			map[string]string{"topic": "my-topic", "server_url": "https://ntfy.example.com/", "token": "tk_abc"},
			true,
		},
		{"ntfy", map[string]string{"topic": ""}, false},
		{"ntfy", map[string]string{"topic": "a/b"}, false},
		{"ntfy", map[string]string{"topic": "a b"}, false},
		{"ntfy", map[string]string{"topic": "t", "server_url": "ftp://x"}, false},
		{"ntfy", map[string]string{"topic": "t", "server_url": "https://u:p@x"}, false},
		{"ntfy", map[string]string{"topic": "t", "token": "bad token"}, false},
		{"ntfy", map[string]string{"topic": "t", "server_url": "https://ntfy.sh/mytopic"}, false},
		{"telegram", map[string]string{"bot_token": tok, "chat_id": "12345"}, true},
		{"telegram", map[string]string{"bot_token": tok, "chat_id": "-1001234567890"}, true},
		{"telegram", map[string]string{"bot_token": tok, "chat_id": "@mychannel"}, true},
		{"telegram", map[string]string{"bot_token": tok, "chat_id": "@ab"}, false},
		{"telegram", map[string]string{"bot_token": tok, "chat_id": "abc"}, false},
		{"telegram", map[string]string{"bot_token": "nocolon", "chat_id": "1"}, false},
		{"telegram", map[string]string{"bot_token": tok + "/../x", "chat_id": "1"}, false},
		{"discord", map[string]string{"webhook_url": hook}, true},
		{"discord", map[string]string{"webhook_url": "https://discord.com/api/v10/webhooks/123456789/" + secret}, true},
		{"discord", map[string]string{"webhook_url": hook + "?wait=true"}, false},
		{"discord", map[string]string{"webhook_url": "https://discord.com/api/webhooks/abc/" + secret}, false},
		{"discord", map[string]string{"webhook_url": "https://discord.com/other"}, false},
		{"discord", map[string]string{"webhook_url": ""}, false},
	}

	reg := registry()

	for _, tc := range cases {
		ch, _ := reg.Get(tc.kind)
		if err := ch.Validate(j(tc.cfg)); (err == nil) != tc.ok {
			t.Errorf("%s %v: ok=%v, err=%v", tc.kind, tc.cfg, tc.ok, err)
		}
	}

	ch, _ := reg.Get("ntfy")
	if ch.Validate(json.RawMessage(`not json`)) == nil {
		t.Error("malformed JSON must be rejected")
	}
}

type captured struct {
	path, auth, ctype string
	body              map[string]any
}

func send(t *testing.T, kind string, cfg func(base string) map[string]string, m notify.Message) captured {
	t.Helper()

	var got captured

	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &got.body)
		got.path, got.auth, got.ctype = r.URL.Path, r.Header.Get("Authorization"), r.Header.Get("Content-Type")
	}))
	defer srv.Close()

	ch, _ := registry().Get(kind)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ch.Send(ctx, notify.NewHTTPClient(), j(cfg(srv.URL)), m); err != nil {
		t.Fatalf("send: %v", err)
	}

	return got
}

var msg = notify.Message{
	Title: "Call Anna", Notes: "ask about trip", PersonName: "Anna",
	DueAt: time.Date(2026, 11, 7, 14, 30, 0, 0, time.UTC), Location: time.UTC,
}

func TestNTFYPayload(t *testing.T) {
	got := send(t, "ntfy", func(b string) map[string]string {
		return map[string]string{"server_url": b, "topic": "tp", "token": "tk_1"}
	}, msg)

	if got.auth != "Bearer tk_1" || got.body["topic"] != "tp" || got.body["title"] != "Call Anna" {
		t.Fatalf("unexpected request: %+v", got)
	}

	if want := "Anna\n2026-11-07 14:30\nask about trip"; got.body["message"] != want {
		t.Fatalf("message %q, want %q", got.body["message"], want)
	}
}

func TestNTFYNoTokenNoAuthHeader(t *testing.T) {
	got := send(t, "ntfy", func(b string) map[string]string {
		return map[string]string{"server_url": b, "topic": "tp"}
	}, msg)

	if got.auth != "" {
		t.Fatalf("auth header %q without a token", got.auth)
	}
}

func TestTelegramPayload(t *testing.T) {
	got := send(t, "telegram", func(b string) map[string]string {
		return map[string]string{"api_url": b, "bot_token": "123456:" + secret, "chat_id": "42"}
	}, msg)

	if got.path != "/bot123456:"+secret+"/sendMessage" || got.body["chat_id"] != "42" {
		t.Fatalf("unexpected request: %+v", got)
	}

	if !strings.HasPrefix(got.body["text"].(string), "Call Anna\nAnna\n") {
		t.Fatalf("text %q", got.body["text"])
	}
}

func TestDiscordPayloadBlocksMentions(t *testing.T) {
	got := send(t, "discord", func(b string) map[string]string {
		return map[string]string{"webhook_url": b + "/api/webhooks/123456789/" + secret}
	}, notify.Message{Title: "@everyone", DueAt: time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC)})

	am, _ := got.body["allowed_mentions"].(map[string]any)
	if parse, ok := am["parse"].([]any); !ok || len(parse) != 0 {
		t.Fatalf("allowed_mentions.parse must be an empty list: %v", got.body)
	}

	// date-only reminder: no clock time
	if got.body["content"] != "**@everyone**\n2026-11-07" {
		t.Fatalf("content %q", got.body["content"])
	}
}

func TestTruncateKeepsRunes(t *testing.T) {
	long := notify.Message{Title: strings.Repeat("é", 3000)}
	got := send(t, "discord", func(b string) map[string]string {
		return map[string]string{"webhook_url": b + "/api/webhooks/123456789/" + secret}
	}, long)

	if n := len([]rune(got.body["content"].(string))); n != 2000 {
		t.Fatalf("content length %d runes, want 2000", n)
	}
}

func TestRateLimitHonorsRetryAfter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "7")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	ch, _ := registry().Get("telegram")
	err := ch.Send(context.Background(), notify.NewHTTPClient(),
		j(map[string]string{"api_url": srv.URL, "bot_token": "123456:" + secret, "chat_id": "1"}), msg)

	se := notify.SanitizeErr(err)
	if se == nil || se.Class != notify.ClassRateLimited || se.RetryAfter != 7*time.Second {
		t.Fatalf("got %+v", se)
	}
}

func TestTruncateLimits(t *testing.T) {
	// 4000 Vietnamese letters are 3 bytes each: must stay under ntfy's 4096-byte cap
	got := send(t, "ntfy", func(b string) map[string]string {
		return map[string]string{"server_url": b, "topic": "tp"}
	}, notify.Message{Title: "t", Notes: strings.Repeat("ạ", 4000)})

	if n := len(got.body["message"].(string)); n > 3900 {
		t.Fatalf("ntfy message %d bytes", n)
	}

	// emoji are 2 UTF-16 units: 3000 of them exceed 4096 units
	got = send(t, "telegram", func(b string) map[string]string {
		return map[string]string{"api_url": b, "bot_token": "123456:" + secret, "chat_id": "1"}
	}, notify.Message{Title: strings.Repeat("😀", 3000)})

	if n := len(utf16.Encode([]rune(got.body["text"].(string)))); n > 4096 {
		t.Fatalf("telegram text %d UTF-16 units", n)
	}
}
