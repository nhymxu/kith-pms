package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/uptrace/bun"

	"github.com/nhymxu/kith-pms/internal/api/handler"
	"github.com/nhymxu/kith-pms/internal/auth"
	"github.com/nhymxu/kith-pms/internal/notify"
	"github.com/nhymxu/kith-pms/internal/notify/notifytest"
)

type hookChannel struct{}

func (hookChannel) Kind() string           { return "hook" }
func (hookChannel) SecretFields() []string { return []string{"token"} }
func (hookChannel) PublicFields() []string { return []string{"url", "token"} }

func (hookChannel) Validate(cfg json.RawMessage) error {
	var v struct{ URL string }
	if err := json.Unmarshal(cfg, &v); err != nil || !strings.HasPrefix(v.URL, "http") {
		return errors.New("bad url")
	}

	return nil
}

func (hookChannel) Send(ctx context.Context, c *notify.HTTPClient, cfg json.RawMessage, _ notify.Message) error {
	var v struct{ URL, Token string }

	_ = json.Unmarshal(cfg, &v)

	req, _ := http.NewRequest(http.MethodPost, v.URL, strings.NewReader("x"))
	req.Header.Set("Authorization", "Bearer "+v.Token)

	return c.Do(ctx, req)
}

type notifyFixture struct {
	db     *bun.DB
	e      *echo.Echo
	cookie string
	target *httptest.Server
	status int
	auth   string
}

func newNotifyFixture(t *testing.T) *notifyFixture {
	t.Helper()

	db := openTestDB(t)
	authSvc := newTestAuthSvc(t, db, "pw-123456789")

	box, err := notify.NewSecretBox("", testSecret)
	if err != nil {
		t.Fatal(err)
	}

	reg := notify.NewRegistry()
	reg.Register(hookChannel{})

	f := &notifyFixture{
		db:     db,
		e:      echo.New(),
		cookie: loginAndGetCookie(t, authSvc, "pw-123456789"),
		status: http.StatusOK,
	}
	f.target = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.auth = r.Header.Get("Authorization")
		w.WriteHeader(f.status)
	}))
	t.Cleanup(f.target.Close)

	h := &handler.NotificationsAPI{
		Repo: notify.NewChannelRepo(db, box), Registry: reg, Client: notify.NewHTTPClient(),
	}

	g := f.e.Group("/v1", auth.SessionOrBearer(testAPIToken, authSvc), auth.SpaCSRF())
	g.GET("/notifications/channels", h.List)
	g.POST("/notifications/channels", h.Create)
	g.PUT("/notifications/channels/:id", h.Update)
	g.PATCH("/notifications/channels/:id", h.Patch)
	g.DELETE("/notifications/channels/:id", h.Delete)
	g.POST("/notifications/channels/:id/test", h.Test)

	return f
}

// do sends a cookie-authenticated request with the CSRF header unless overridden.
func (f *notifyFixture) do(method, path, body string, mut ...func(*http.Request)) *httptest.ResponseRecorder {
	req := jsonRequest(method, path, body)
	req.AddCookie(&http.Cookie{Name: "kith_session", Value: f.cookie})
	req.Header.Set("X-Requested-With", "kith-spa")

	for _, m := range mut {
		m(req)
	}

	rec := httptest.NewRecorder()
	f.e.ServeHTTP(rec, req)

	return rec
}

func (f *notifyFixture) create(t *testing.T, token string) (int64, string) {
	t.Helper()

	body, _ := json.Marshal(map[string]any{
		"kind": "hook", "name": "mine", "config": map[string]string{"url": f.target.URL, "token": token},
	})

	rec := f.do(http.MethodPost, "/v1/notifications/channels", string(body))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", rec.Code, rec.Body.String())
	}

	var env struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}

	_ = json.Unmarshal(rec.Body.Bytes(), &env)

	return env.Data.ID, rec.Body.String()
}

func idPath(id int64) string {
	b, _ := json.Marshal(id)

	return "/v1/notifications/channels/" + string(b)
}

func TestNotificationChannelsCRUDAndSecretHandling(t *testing.T) {
	f := newNotifyFixture(t)
	id, createBody := f.create(t, notifytest.Canary)
	notifytest.AssertNoCanary(t, "create response", createBody)

	list := f.do(http.MethodGet, "/v1/notifications/channels", "")
	notifytest.AssertNoCanary(t, "list response", list.Body.String())

	if !strings.Contains(list.Body.String(), `"kinds":["hook"]`) ||
		!strings.Contains(list.Body.String(), f.target.URL) {
		t.Fatalf("list should expose kinds and public fields: %s", list.Body.String())
	}

	// rename keeps the secret: a masked-secret PUT with the same destination still tests OK
	same, _ := json.Marshal(map[string]string{"url": f.target.URL, "token": notify.MaskSentinel})
	if rec := f.do(http.MethodPut, idPath(id), string(same)); rec.Code != http.StatusOK {
		t.Fatalf("masked PUT: %d %s", rec.Code, rec.Body.String())
	}

	if rec := f.do(http.MethodPatch, idPath(id), `{"name":"renamed","enabled":false}`); rec.Code != http.StatusOK ||
		!strings.Contains(
			rec.Body.String(),
			`"name":"renamed"`,
		) || !strings.Contains(rec.Body.String(), `"enabled":false`) {
		t.Fatalf("patch: %d %s", rec.Code, rec.Body.String())
	}

	if rec := f.do(http.MethodPost, idPath(id)+"/test", ""); !strings.Contains(rec.Body.String(), `"ok":true`) ||
		f.auth != "Bearer "+notifytest.Canary {
		t.Fatalf("stored secret must survive masked PUT and rename; target saw %q", f.auth)
	}

	// destination change without re-sent secret is refused
	moved, _ := json.Marshal(map[string]string{"url": "https://evil.example/hook"})
	rec := f.do(http.MethodPut, idPath(id), string(moved))

	if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), `"code":"secret_required"`) {
		t.Fatalf("destination change: %d %s", rec.Code, rec.Body.String())
	}

	if rec := f.do(
		http.MethodPut,
		idPath(id),
		`{"url":"nope","token":"x"}`,
	); rec.Code != http.StatusUnprocessableEntity ||
		!strings.Contains(rec.Body.String(), `"code":"invalid_channel_config"`) {
		t.Fatalf("validation: %d %s", rec.Code, rec.Body.String())
	}

	if rec := f.do(http.MethodPost, "/v1/notifications/channels", `{"kind":"nope","name":"x","config":{}}`); rec.Code !=
		http.StatusUnprocessableEntity {
		t.Fatalf("unknown kind: %d", rec.Code)
	}

	if rec := f.do(http.MethodDelete, idPath(id), ""); rec.Code != http.StatusNoContent {
		t.Fatalf("delete: %d", rec.Code)
	}

	if rec := f.do(http.MethodDelete, idPath(id), ""); rec.Code != http.StatusNotFound {
		t.Fatalf("second delete: %d", rec.Code)
	}
}

func TestNotificationTestSendResultAndLimiter(t *testing.T) {
	f := newNotifyFixture(t)
	id, _ := f.create(t, notifytest.Canary)

	rec := f.do(http.MethodPost, idPath(id)+"/test", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"ok":true`) {
		t.Fatalf("test send: %d %s", rec.Code, rec.Body.String())
	}

	if rec := f.do(http.MethodPost, idPath(id)+"/test", ""); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("limiter: %d", rec.Code)
	}
}

func TestNotificationTestSendReportsClassWithoutLeaks(t *testing.T) {
	f := newNotifyFixture(t)
	f.status = http.StatusBadGateway
	id, _ := f.create(t, notifytest.Canary)

	rec := f.do(http.MethodPost, idPath(id)+"/test", "")
	if !strings.Contains(rec.Body.String(), `"class":"http_5xx"`) ||
		!strings.Contains(rec.Body.String(), `"status":502`) {
		t.Fatalf("expected class+status, got %s", rec.Body.String())
	}

	notifytest.AssertNoCanary(t, "test response", rec.Body.String())
}

func TestNotificationMutationsRejectBearerAndMissingCSRF(t *testing.T) {
	f := newNotifyFixture(t)
	id, _ := f.create(t, "tok")

	bearer := func(r *http.Request) {
		r.Header.Del("Cookie")
		r.Header.Set("Authorization", "Bearer "+testAPIToken)
	}

	for _, c := range []struct{ method, path, body string }{
		{http.MethodPost, "/v1/notifications/channels", `{"kind":"hook","name":"x","config":{"url":"http://a"}}`},
		{http.MethodPut, idPath(id), `{"url":"http://a"}`},
		{http.MethodPatch, idPath(id), `{"enabled":false}`},
		{http.MethodDelete, idPath(id), ""},
		{http.MethodPost, idPath(id) + "/test", ""},
	} {
		rec := f.do(c.method, c.path, c.body, bearer)
		if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "session_required") {
			t.Errorf("bearer %s %s: %d %s", c.method, c.path, rec.Code, rec.Body.String())
		}

		noCSRF := func(r *http.Request) { r.Header.Del("X-Requested-With") }
		if rec := f.do(c.method, c.path, c.body, noCSRF); rec.Code != http.StatusForbidden {
			t.Errorf("missing CSRF header %s %s: %d", c.method, c.path, rec.Code)
		}
	}

	if rec := f.do(http.MethodGet, "/v1/notifications/channels", "", bearer); rec.Code != http.StatusOK {
		t.Errorf("bearer may still read the list: %d", rec.Code)
	}
}

func TestNotificationUnreadableChannelCanBeReentered(t *testing.T) {
	f := newNotifyFixture(t)
	id, _ := f.create(t, "tok")

	if _, err := f.db.ExecContext(
		context.Background(),
		`UPDATE notification_channel SET config_enc = x'0102'`,
	); err != nil {
		t.Fatal(err)
	}

	list := f.do(http.MethodGet, "/v1/notifications/channels", "")
	if !strings.Contains(list.Body.String(), `"status":"secret_unreadable"`) {
		t.Fatalf("list must flag the channel: %s", list.Body.String())
	}

	if rec := f.do(http.MethodPost, idPath(id)+"/test", ""); rec.Code != http.StatusUnprocessableEntity ||
		!strings.Contains(rec.Body.String(), "channel_secret_unreadable") {
		t.Fatalf("test-send on unreadable: %d %s", rec.Code, rec.Body.String())
	}

	fresh, _ := json.Marshal(map[string]string{"url": f.target.URL, "token": "new"})
	if rec := f.do(http.MethodPut, idPath(id), string(fresh)); rec.Code != http.StatusOK ||
		!strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("re-entering secrets: %d %s", rec.Code, rec.Body.String())
	}
}

func TestNotificationMaskedSecretIsNeverStored(t *testing.T) {
	f := newNotifyFixture(t)

	body, _ := json.Marshal(map[string]any{
		"kind": "hook", "name": "x", "config": map[string]string{"url": f.target.URL, "token": notify.MaskSentinel},
	})
	if rec := f.do(
		http.MethodPost,
		"/v1/notifications/channels",
		string(body),
	); rec.Code != http.StatusUnprocessableEntity ||
		!strings.Contains(rec.Body.String(), "secret_required") {
		t.Fatalf("create with mask: %d %s", rec.Code, rec.Body.String())
	}

	id, _ := f.create(t, "tok")
	if _, err := f.db.ExecContext(
		context.Background(),
		`UPDATE notification_channel SET config_enc = x'0102'`,
	); err != nil {
		t.Fatal(err)
	}

	masked, _ := json.Marshal(map[string]string{"url": f.target.URL, "token": notify.MaskSentinel})
	if rec := f.do(http.MethodPut, idPath(id), string(masked)); rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("re-enter with mask must be refused: %d %s", rec.Code, rec.Body.String())
	}
}
