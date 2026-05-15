package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/auth"
)

// ---- helpers ----------------------------------------------------------------

func newTestSvc(t *testing.T) *auth.Service {
	t.Helper()

	db := setupTestDB(t)
	t.Cleanup(func() { _ = db.Close() })

	return &auth.Service{
		Users:    auth.NewUserRepo(db),
		Sessions: auth.NewSessionRepo(db),
		Secret:   []byte("test-secret-key-32-bytes-long!00"),
		Lifetime: 24 * time.Hour,
	}
}

func handlerOK(c *echo.Context) error { return c.String(http.StatusOK, "ok") }

func applySessionOrBearer(e *echo.Echo, svc *auth.Service, apiToken string, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	mw := auth.SessionOrBearer(apiToken, svc)(handlerOK)
	_ = mw(c)

	return rec
}

// ---- SessionOrBearer: Bearer path -------------------------------------------

func TestSessionOrBearer_ValidBearer_Returns200(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/v1/test", nil)
	req.Header.Set("Authorization", "Bearer my-api-token")

	rec := applySessionOrBearer(e, nil, "my-api-token", req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestSessionOrBearer_WrongBearer_Returns401(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/v1/test", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")

	rec := applySessionOrBearer(e, nil, "my-api-token", req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

// ---- SessionOrBearer: cookie path -------------------------------------------

func TestSessionOrBearer_ValidCookie_Returns200(t *testing.T) {
	svc := newTestSvc(t)

	// Seed a user so Login works.
	ctx := context.Background()
	hash, _ := auth.HashPassword("pw")
	_ = svc.Users.UpsertUser(ctx, hash)

	token, err := svc.Login(ctx, "pw", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/v1/test", nil)
	req.AddCookie(&http.Cookie{Name: "kith_session", Value: token})

	rec := applySessionOrBearer(e, svc, "", req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestSessionOrBearer_InvalidCookie_Returns401(t *testing.T) {
	svc := newTestSvc(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/v1/test", nil)
	req.AddCookie(&http.Cookie{Name: "kith_session", Value: "bad-token"})

	rec := applySessionOrBearer(e, svc, "", req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestSessionOrBearer_NoAuth_Returns401(t *testing.T) {
	svc := newTestSvc(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/v1/test", nil)

	rec := applySessionOrBearer(e, svc, "token", req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

// ---- SpaCSRF ----------------------------------------------------------------

func applySpaCSRF(e *echo.Echo, method, xrw string, cookieAuthed bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/v1/test", nil)
	if xrw != "" {
		req.Header.Set("X-Requested-With", xrw)
	}

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if cookieAuthed {
		c.Set("cookie_authed", true)
	}

	mw := auth.SpaCSRF()(handlerOK)
	_ = mw(c)

	return rec
}

func TestSpaCSRF_GET_NoHeader_CookieAuthed_Passes(t *testing.T) {
	rec := applySpaCSRF(echo.New(), http.MethodGet, "", true)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestSpaCSRF_POST_CookieAuthed_NoHeader_Returns403(t *testing.T) {
	rec := applySpaCSRF(echo.New(), http.MethodPost, "", true)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestSpaCSRF_POST_CookieAuthed_WithHeader_Passes(t *testing.T) {
	rec := applySpaCSRF(echo.New(), http.MethodPost, "kith-spa", true)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestSpaCSRF_POST_BearerAuthed_NoHeader_Passes(t *testing.T) {
	// cookieAuthed=false means Bearer — CSRF gate is skipped.
	rec := applySpaCSRF(echo.New(), http.MethodPost, "", false)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestSpaCSRF_DELETE_CookieAuthed_NoHeader_Returns403(t *testing.T) {
	rec := applySpaCSRF(echo.New(), http.MethodDelete, "", true)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestSpaCSRF_PUT_CookieAuthed_WithHeader_Passes(t *testing.T) {
	rec := applySpaCSRF(echo.New(), http.MethodPut, "kith-spa", true)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestSpaCSRF_PATCH_CookieAuthed_NoHeader_Returns403(t *testing.T) {
	rec := applySpaCSRF(echo.New(), http.MethodPatch, "", true)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}
