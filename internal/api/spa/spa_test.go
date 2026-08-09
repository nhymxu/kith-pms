package spa

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
)

// The frontend crop dialog (react-advanced-cropper) renders to a <canvas>, so
// it fetches its blob: source. fetch() on a blob: URL is governed by
// connect-src; without blob: here the "Adjust image" popup renders black in
// production (the Vite dev server sends no CSP, hiding the break).
func TestIndexCSPAllowsConnectSrcBlob(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	setIndexHeaders(c)

	csp := rec.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Fatal("Content-Security-Policy header not set on index responses")
	}

	for _, want := range []string{
		"connect-src 'self' blob:",
		"worker-src 'self' blob:",
		"img-src 'self' data: blob:",
	} {
		if !strings.Contains(csp, want) {
			t.Errorf("CSP missing %q; got:\n%s", want, csp)
		}
	}
}
