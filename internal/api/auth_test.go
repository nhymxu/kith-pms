package api_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/api"
	"github.com/nhymxu/kith-pms/internal/auth"
)

// ---- Login ------------------------------------------------------------------

func TestAuthLogin_HappyPath(t *testing.T) {
	db := openTestDB(t)
	svc := newTestAuthSvc(t, db, "correct-pw")
	h := &api.AuthAPI{Svc: svc}

	e := newTestEcho()
	req := jsonRequest(http.MethodPost, "/v1/auth/login", `{"password":"correct-pw"}`)
	rec := execHandler(e, req, nil, h.Login)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", rec.Code, rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), `"logged_in"`) {
		t.Fatalf("expected logged_in in response, got: %s", rec.Body.String())
	}

	// Session cookie must be set.
	found := false
	for _, c := range rec.Result().Cookies() {
		if c.Name == "kith_session" && c.Value != "" {
			found = true
		}
	}

	if !found {
		t.Fatal("expected kith_session cookie to be set")
	}
}

func TestAuthLogin_WrongPassword_Returns401(t *testing.T) {
	db := openTestDB(t)
	svc := newTestAuthSvc(t, db, "correct-pw")
	h := &api.AuthAPI{Svc: svc}

	e := newTestEcho()
	req := jsonRequest(http.MethodPost, "/v1/auth/login", `{"password":"wrong-pw"}`)
	rec := execHandler(e, req, nil, h.Login)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthLogin_MissingPassword_Returns400(t *testing.T) {
	db := openTestDB(t)
	svc := newTestAuthSvc(t, db, "correct-pw")
	h := &api.AuthAPI{Svc: svc}

	e := newTestEcho()
	req := jsonRequest(http.MethodPost, "/v1/auth/login", `{}`)
	rec := execHandler(e, req, nil, h.Login)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// ---- Logout -----------------------------------------------------------------

func TestAuthLogout_HappyPath(t *testing.T) {
	db := openTestDB(t)
	svc := newTestAuthSvc(t, db, "pw")
	h := &api.AuthAPI{Svc: svc}

	session := loginAndGetCookie(t, svc, "pw")

	e := newTestEcho()
	req := jsonRequest(http.MethodPost, "/v1/auth/logout", "")
	req.AddCookie(&http.Cookie{Name: "kith_session", Value: session})
	rec := execHandler(e, req, nil, h.Logout)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// Cookie must be cleared (MaxAge == -1).
	cleared := false
	for _, c := range rec.Result().Cookies() {
		if c.Name == "kith_session" && c.MaxAge == -1 {
			cleared = true
		}
	}

	if !cleared {
		t.Fatal("expected kith_session cookie to be cleared (MaxAge=-1)")
	}
}

// ---- Me ---------------------------------------------------------------------

func TestAuthMe_AuthenticatedUser_Returns200(t *testing.T) {
	db := openTestDB(t)
	svc := newTestAuthSvc(t, db, "pw")
	h := &api.AuthAPI{Svc: svc}

	// Fetch the seeded user to use as context value.
	user, err := svc.Users.GetUser(context.Background()) //nolint:staticcheck
	if err != nil || user == nil {
		t.Fatalf("get test user: %v", err)
	}

	e := newTestEcho()
	req := jsonRequest(http.MethodGet, "/v1/auth/me", "")
	rec := execHandlerWithUser(e, req, nil, user, h.Me)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", rec.Code, rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), `"id"`) {
		t.Fatalf("expected id in response, got: %s", rec.Body.String())
	}
}

func TestAuthMe_Unauthenticated_Returns401(t *testing.T) {
	db := openTestDB(t)
	svc := newTestAuthSvc(t, db, "pw")
	h := &api.AuthAPI{Svc: svc}

	e := newTestEcho()
	req := jsonRequest(http.MethodGet, "/v1/auth/me", "")
	// No user set on context — simulates unauthenticated request.
	rec := execHandler(e, req, nil, h.Me)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

// ---- ChangePassword ---------------------------------------------------------

func TestAuthChangePassword_HappyPath(t *testing.T) {
	db := openTestDB(t)
	svc := newTestAuthSvc(t, db, "old-pw-long")
	h := &api.AuthAPI{Svc: svc}

	e := newTestEcho()
	body := `{"current_password":"old-pw-long","new_password":"new-pw-long","confirm_password":"new-pw-long"}`
	req := jsonRequest(http.MethodPost, "/v1/auth/password", body)
	rec := execHandler(e, req, nil, h.ChangePassword)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", rec.Code, rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), `"password_changed"`) {
		t.Fatalf("expected password_changed in response, got: %s", rec.Body.String())
	}
}

func TestAuthChangePassword_WrongCurrentPassword_Returns422(t *testing.T) {
	db := openTestDB(t)
	svc := newTestAuthSvc(t, db, "correct-pw")
	h := &api.AuthAPI{Svc: svc}

	e := newTestEcho()
	body := `{"current_password":"wrong-pw","new_password":"new-pw-long","confirm_password":"new-pw-long"}`
	req := jsonRequest(http.MethodPost, "/v1/auth/password", body)
	rec := execHandler(e, req, nil, h.ChangePassword)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

func TestAuthChangePassword_MismatchConfirm_Returns422(t *testing.T) {
	db := openTestDB(t)
	svc := newTestAuthSvc(t, db, "correct-pw")
	h := &api.AuthAPI{Svc: svc}

	e := newTestEcho()
	body := `{"current_password":"correct-pw","new_password":"new-pw-long","confirm_password":"different-pw"}`
	req := jsonRequest(http.MethodPost, "/v1/auth/password", body)
	rec := execHandler(e, req, nil, h.ChangePassword)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

func TestAuthChangePassword_TooShort_Returns422(t *testing.T) {
	db := openTestDB(t)
	svc := newTestAuthSvc(t, db, "correct-pw")
	h := &api.AuthAPI{Svc: svc}

	e := newTestEcho()
	body := `{"current_password":"correct-pw","new_password":"short","confirm_password":"short"}`
	req := jsonRequest(http.MethodPost, "/v1/auth/password", body)
	rec := execHandler(e, req, nil, h.ChangePassword)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

// ---- LogoutAll --------------------------------------------------------------

func TestAuthLogoutAll_HappyPath(t *testing.T) {
	db := openTestDB(t)
	svc := newTestAuthSvc(t, db, "pw")
	h := &api.AuthAPI{Svc: svc}

	// User is already seeded by newTestAuthSvc — LogoutAll resolves via Users.GetUser.

	e := newTestEcho()
	req := jsonRequest(http.MethodPost, "/v1/auth/logout-all", "")
	rec := execHandler(e, req, nil, h.LogoutAll)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// ---- auth failure via SessionOrBearer (integration smoke) -------------------

func TestSessionOrBearer_AuthFailure_Returns401JSON(t *testing.T) {
	db := openTestDB(t)
	svc := newTestAuthSvc(t, db, "pw")

	e := newTestEcho()
	req := jsonRequest(http.MethodGet, "/v1/people", "")
	// No Bearer, no cookie — must get JSON 401.
	rec := execHandler(e, req, nil, func(c *echo.Context) error {
		mw := auth.SessionOrBearer(testAPIToken, svc)
		return mw(func(_ *echo.Context) error { return nil })(c)
	})

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `"error"`) {
		t.Fatalf("expected JSON error body, got: %s", rec.Body.String())
	}
}
