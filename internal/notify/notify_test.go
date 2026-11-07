package notify_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nhymxu/kith-pms/internal/notify"
	"github.com/nhymxu/kith-pms/internal/notify/notifytest"
	"github.com/nhymxu/kith-pms/internal/testutil"
)

type fakeChannel struct{}

func (fakeChannel) Kind() string           { return "fake" }
func (fakeChannel) SecretFields() []string { return []string{"token"} }

// token listed on purpose: PublicView must still drop it
func (fakeChannel) PublicFields() []string { return []string{"url", "token"} }

func (fakeChannel) Validate(cfg json.RawMessage) error {
	var c struct{ URL, Token string }
	if err := json.Unmarshal(cfg, &c); err != nil || c.URL == "" {
		return errors.New("bad config")
	}

	return nil
}

func (fakeChannel) Send(ctx context.Context, c *notify.HTTPClient, cfg json.RawMessage, _ notify.Message) error {
	var v struct{ URL, Token string }

	_ = json.Unmarshal(cfg, &v)

	req, _ := http.NewRequest(http.MethodPost, v.URL, strings.NewReader("x"))
	req.Header.Set("Authorization", "Bearer "+v.Token)

	return c.Do(ctx, req)
}

func cfg(url, token string) json.RawMessage {
	b, _ := json.Marshal(map[string]string{"url": url, "token": token})

	return b
}

func newBox(t *testing.T) *notify.SecretBox {
	t.Helper()

	k := make([]byte, 32)
	_, _ = rand.Read(k)

	b, err := notify.NewSecretBox(base64.StdEncoding.EncodeToString(k), "session-secret-session-secret-xx")
	if err != nil {
		t.Fatal(err)
	}

	return b
}

func TestSecretBox(t *testing.T) {
	box := newBox(t)
	blob, _ := box.Seal("fake|1", []byte("hello"))

	if pt, err := box.Open("fake|1", blob); err != nil || string(pt) != "hello" {
		t.Fatalf("round trip: %q %v", pt, err)
	}

	if _, err := box.Open("fake|2", blob); !errors.Is(err, notify.ErrSecretUnreadable) {
		t.Fatal("row swap must fail")
	}

	tampered := append([]byte{}, blob...)
	tampered[len(tampered)-1] ^= 1

	if _, err := box.Open("fake|1", tampered); !errors.Is(err, notify.ErrSecretUnreadable) {
		t.Fatal("tamper must fail")
	}

	if _, err := newBox(t).Open("fake|1", blob); !errors.Is(err, notify.ErrSecretUnreadable) {
		t.Fatal("wrong key must fail")
	}

	derived, _ := notify.NewSecretBox("", "session-secret-session-secret-xx")
	dblob, _ := derived.Seal("fake|1", []byte("hi"))

	both, _ := notify.NewSecretBox(
		base64.StdEncoding.EncodeToString(make([]byte, 32)),
		"session-secret-session-secret-xx",
	)
	if _, err := both.Open("fake|1", dblob); err != nil {
		t.Fatalf("derived-key blob must stay readable after adding explicit key: %v", err)
	}

	if _, err := notify.NewSecretBox("not-base64!", "x"); err == nil {
		t.Fatal("bad key must be rejected")
	}
}

func TestHTTPClient(t *testing.T) {
	var hits atomic.Int32

	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { hits.Add(1) }))
	defer target.Close()

	redir := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusFound)
	}))
	defer redir.Close()

	c := notify.NewHTTPClient()
	do := func(u string) *notify.SendError {
		req, _ := http.NewRequest(http.MethodGet, u, nil)

		err := c.Do(context.Background(), req)
		if err == nil {
			return nil
		}

		var se *notify.SendError
		if !errors.As(err, &se) {
			t.Fatalf("non SendError: %v", err)
		}

		return se
	}

	if se := do(target.URL); se != nil {
		t.Fatalf("loopback http should work: %v", se)
	}

	hits.Store(0)

	if se := do(redir.URL); se == nil || se.Class != notify.ClassInvalidConfig || hits.Load() != 0 {
		t.Fatalf("redirect must not be followed: %v hits=%d", se, hits.Load())
	}

	blocked := []string{
		"http://169.254.169.254/", "http://0.0.0.0:1/", "http://[fe80::1]:1/",
		"https://169.254.169.254/", "http://224.0.0.1:1/",
		"http://[::ffff:169.254.169.254]/", "http://[fd00:ec2::254]/", "https://[64:ff9b::a9fe:a9fe]/",
	}

	for _, u := range blocked {
		if se := do(u); se == nil || se.Class != notify.ClassInvalidConfig || !se.Permanent {
			t.Errorf("%s must be blocked, got %v", u, se)
		}
	}

	if se := do("http://8.8.8.8/"); se == nil || se.Class != notify.ClassInvalidConfig {
		t.Errorf("plain http to public IP must be rejected, got %v", se)
	}
}

func TestClassifyStatus(t *testing.T) {
	h := http.Header{"Retry-After": {"99999"}}

	cases := []struct {
		status    int
		class     string
		permanent bool
	}{
		{200, "", false},
		{400, notify.ClassHTTP4xx, true},
		{408, notify.ClassHTTP4xx, false},
		{
			429,
			notify.ClassRateLimited,
			false,
		},
		{503, notify.ClassHTTP5xx, false},
		{302, notify.ClassInvalidConfig, true},
	}

	for _, c := range cases {
		se := notify.ClassifyStatus(c.status, h)
		if c.class == "" {
			if se != nil {
				t.Errorf("%d: want nil", c.status)
			}

			continue
		}

		if se.Class != c.class || se.Permanent != c.permanent {
			t.Errorf("%d: got %+v", c.status, se)
		}
	}

	if se := notify.ClassifyStatus(429, h); se.RetryAfter != time.Hour {
		t.Errorf("Retry-After must cap at 1h, got %v", se.RetryAfter)
	}
}

func TestClassifyTransport(t *testing.T) {
	if se := notify.Classify(context.DeadlineExceeded); se.Class != notify.ClassTimeout {
		t.Error("deadline")
	}

	if se := notify.Classify(
		&net.DNSError{Err: "secret-host"},
	); se.Class != notify.ClassDNS ||
		strings.Contains(se.Error(), "secret-host") {
		t.Error("dns")
	}

	if se := notify.SanitizeErr(
		errors.New(notifytest.Canary),
	); se.Class != notify.ClassInternal ||
		strings.Contains(se.Error(), "CANARY") {
		t.Error("raw error text must be dropped")
	}

	if se := notify.SanitizeErr(&notify.SendError{Class: notifytest.Canary}); se.Class != notify.ClassInternal {
		t.Error("unknown class must collapse to internal")
	}
}

func TestMergeConfig(t *testing.T) {
	ch := fakeChannel{}
	stored := cfg("https://a.example/hook", "S1")

	got, err := notify.MergeConfig(ch, stored, json.RawMessage(`{"url":"https://a.example/hook","token":"********"}`))
	if err != nil || !strings.Contains(string(got), `"S1"`) {
		t.Fatalf("masked secret must keep stored: %s %v", got, err)
	}

	got, err = notify.MergeConfig(ch, stored, json.RawMessage(`{"url":"https://a.example/hook"}`))
	if err != nil || !strings.Contains(string(got), `"S1"`) {
		t.Fatalf("omitted secret must keep stored: %s %v", got, err)
	}

	if _, err = notify.MergeConfig(
		ch,
		stored,
		json.RawMessage(`{"url":"https://evil.example/"}`),
	); !errors.Is(
		err,
		notify.ErrSecretRequired,
	) {
		t.Fatalf("destination change without secret: %v", err)
	}

	if _, err = notify.MergeConfig(
		ch,
		stored,
		json.RawMessage(`{"url":"https://evil.example/","token":"********"}`),
	); !errors.Is(
		err,
		notify.ErrSecretRequired,
	) {
		t.Fatalf("destination change with masked secret: %v", err)
	}

	got, err = notify.MergeConfig(ch, stored, json.RawMessage(`{"url":"https://b.example/","token":"S2"}`))
	if err != nil || !strings.Contains(string(got), `"S2"`) {
		t.Fatalf("destination change with new secret: %s %v", got, err)
	}

	if _, err = notify.MergeConfig(ch, stored, json.RawMessage(`[]`)); err == nil {
		t.Fatal("non-object must fail")
	}
}

func TestChannelRepo(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	repo := notify.NewChannelRepo(db, newBox(t))

	rec, err := repo.Create(ctx, "fake", "mine", true, cfg("https://a/", "S1"))
	if err != nil {
		t.Fatal(err)
	}

	var raw []byte
	if err := db.NewSelect().
		Table("notification_channel").
		Column("config_enc").
		Where("id = ?", rec.ID).
		Scan(ctx, &raw); err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(raw), "S1") {
		t.Fatal("config stored in plaintext")
	}

	got, err := repo.Get(ctx, rec.ID)
	if err != nil || got.Unreadable || !strings.Contains(string(got.Config), "S1") {
		t.Fatalf("get: %+v %v", got, err)
	}

	if err := repo.UpdateConfig(ctx, rec.ID, cfg("https://b/", "S2")); err != nil {
		t.Fatal(err)
	}

	if err := repo.Rename(ctx, rec.ID, "renamed"); err != nil {
		t.Fatal(err)
	}

	if err := repo.SetEnabled(ctx, rec.ID, false); err != nil {
		t.Fatal(err)
	}

	list, _ := repo.List(ctx)
	if len(list) != 1 || list[0].Name != "renamed" || list[0].Enabled ||
		!strings.Contains(string(list[0].Config), "S2") {
		t.Fatalf("list: %+v", list)
	}

	other := notify.NewChannelRepo(db, newBox(t))
	if g, _ := other.Get(ctx, rec.ID); !g.Unreadable || g.Config != nil {
		t.Fatal("wrong key must flag unreadable")
	}

	if err := repo.Delete(ctx, rec.ID); err != nil {
		t.Fatal(err)
	}

	if _, err := repo.Get(ctx, rec.ID); !errors.Is(err, notify.ErrChannelNotFound) {
		t.Fatalf("after delete: %v", err)
	}

	if err := repo.Delete(ctx, rec.ID); !errors.Is(err, notify.ErrChannelNotFound) {
		t.Fatal("double delete must be not found")
	}
}

func TestContractOnFakeChannel(t *testing.T) {
	reg := notify.NewRegistry()
	reg.Register(fakeChannel{})
	notifytest.RunContract(t, reg, map[string]notifytest.Fixture{
		"fake": func(base, secret string) json.RawMessage { return cfg(base, secret) },
	})
}

func TestRegistryDuplicatePanics(t *testing.T) {
	reg := notify.NewRegistry()
	reg.Register(fakeChannel{})

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	reg.Register(fakeChannel{})
}

func TestRetryAfterParsing(t *testing.T) {
	date := time.Now().Add(90 * time.Second).UTC().Format(http.TimeFormat)

	cases := map[string]func(time.Duration) bool{
		"30":         func(d time.Duration) bool { return d == 30*time.Second },
		date:         func(d time.Duration) bool { return d > 60*time.Second && d <= 90*time.Second },
		"9223372037": func(d time.Duration) bool { return d == time.Hour },
		"-5":         func(d time.Duration) bool { return d == 0 },
		"garbage":    func(d time.Duration) bool { return d == 0 },
		"":           func(d time.Duration) bool { return d == 0 },
	}

	for v, ok := range cases {
		se := notify.ClassifyStatus(http.StatusTooManyRequests, http.Header{"Retry-After": {v}})
		if !ok(se.RetryAfter) {
			t.Errorf("Retry-After %q gave %v", v, se.RetryAfter)
		}
	}
}

func TestSanitizeTypedNilAndNegative(t *testing.T) {
	var nilSE *notify.SendError

	var err error = nilSE
	if notify.SanitizeErr(err) != nil {
		t.Fatal("typed-nil SendError must read as success")
	}

	se := notify.SanitizeErr(&notify.SendError{Class: notify.ClassHTTP5xx, RetryAfter: -time.Hour})
	if se.RetryAfter != 0 {
		t.Fatalf("negative RetryAfter must clamp to 0, got %v", se.RetryAfter)
	}
}

func TestSetEnabledKeepsGateOnIdempotentEnable(t *testing.T) {
	ctx := context.Background()
	repo := notify.NewChannelRepo(testutil.NewDB(t), newBox(t))

	rec, _ := repo.Create(ctx, "fake", "n", true, cfg("https://a/", "S"))

	time.Sleep(5 * time.Millisecond)

	if err := repo.SetEnabled(ctx, rec.ID, true); err != nil {
		t.Fatal(err)
	}

	got, _ := repo.Get(ctx, rec.ID)
	if !got.EnabledAt.Equal(rec.EnabledAt) {
		t.Fatal("enabling an enabled channel must not move enabled_at")
	}

	_ = repo.SetEnabled(ctx, rec.ID, false)

	time.Sleep(5 * time.Millisecond)

	_ = repo.SetEnabled(ctx, rec.ID, true)

	got, _ = repo.Get(ctx, rec.ID)
	if !got.EnabledAt.After(rec.EnabledAt) {
		t.Fatal("re-enabling must advance enabled_at")
	}
}
