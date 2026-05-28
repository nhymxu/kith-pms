package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/uptrace/bun"

	"github.com/nhymxu/kith-pms/internal/audit"
	"github.com/nhymxu/kith-pms/internal/auth"
	internaldb "github.com/nhymxu/kith-pms/internal/db"
	"github.com/nhymxu/kith-pms/internal/gifts"
	"github.com/nhymxu/kith-pms/internal/journal"
	"github.com/nhymxu/kith-pms/internal/labels"
	"github.com/nhymxu/kith-pms/internal/people"
	"github.com/nhymxu/kith-pms/internal/relationships"
)

const testAPIToken = "test-api-token-12345"
const testSecret = "test-secret-key-32-bytes-long!00"

// openTestDB opens an in-memory SQLite database with all migrations applied.
func openTestDB(t *testing.T) *bun.DB {
	t.Helper()

	db, err := internaldb.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	if err := internaldb.Up(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	return db
}

// newTestAuthSvc creates an auth.Service backed by db. If password is non-empty, seeds a user.
func newTestAuthSvc(t *testing.T, db *bun.DB, password string) *auth.Service {
	t.Helper()

	svc := &auth.Service{
		Users:    auth.NewUserRepo(db),
		Sessions: auth.NewSessionRepo(db),
		Secret:   []byte(testSecret),
		Lifetime: 24 * time.Hour,
	}

	if password != "" {
		hash, err := auth.HashPassword(password)
		if err != nil {
			t.Fatalf("hash password: %v", err)
		}

		if err := svc.Users.UpsertUser(context.Background(), hash); err != nil {
			t.Fatalf("upsert user: %v", err)
		}
	}

	return svc
}

// loginAndGetCookie logs in and returns the session token for use as cookie value.
func loginAndGetCookie(t *testing.T, svc *auth.Service, password string) string {
	t.Helper()

	token, err := svc.Login(context.Background(), password, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	return token
}

// newTestEcho returns a bare Echo instance for handler tests.
func newTestEcho() *echo.Echo { return echo.New() }

// jsonRequest builds an HTTP request with a JSON body.
func jsonRequest(method, path, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	return req
}

// execHandler runs a handler directly, setting named path params via Echo v5 PathValues.
// params maps param name → value (e.g. {"id": "42"}).
func execHandler(
	e *echo.Echo,
	req *http.Request,
	params map[string]string,
	handler echo.HandlerFunc,
) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if len(params) > 0 {
		pv := make(echo.PathValues, 0, len(params))
		for k, v := range params {
			pv = append(pv, echo.PathValue{Name: k, Value: v})
		}

		c.SetPathValues(pv)
	}

	_ = handler(c)

	return rec
}

// execHandlerWithUser is like execHandler but also sets a *auth.User on the context.
func execHandlerWithUser(
	e *echo.Echo,
	req *http.Request,
	params map[string]string,
	user *auth.User,
	handler echo.HandlerFunc,
) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if len(params) > 0 {
		pv := make(echo.PathValues, 0, len(params))
		for k, v := range params {
			pv = append(pv, echo.PathValue{Name: k, Value: v})
		}

		c.SetPathValues(pv)
	}

	if user != nil {
		c.Set("user", user)
	}

	_ = handler(c)

	return rec
}

// ---- service factories ------------------------------------------------------

func newPeopleService(db *bun.DB) *people.Service {
	svc := people.NewService(db)
	svc.Audit = audit.NewService(db)

	return svc
}

func newLabelsService(db *bun.DB) *labels.Service {
	return labels.NewService(db)
}

func newRelationshipsService(db *bun.DB) *relationships.Service {
	return relationships.NewService(db)
}

func newJournalService(db *bun.DB) *journal.Service {
	return journal.NewService(db)
}

func newGiftsService(db *bun.DB) *gifts.Service {
	return gifts.NewService(db)
}

// insertTestPerson inserts a person row and returns its ID.
func insertTestPerson(t *testing.T, db *bun.DB, name string) int64 {
	t.Helper()

	res, err := db.ExecContext(context.Background(), `INSERT INTO person (name) VALUES (?)`, name)
	if err != nil {
		t.Fatalf("insert person %q: %v", name, err)
	}

	id, _ := res.LastInsertId()

	return id
}
