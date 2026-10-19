package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/api"
)

func TestJSONErrorHandler_EmitsHTTPStatusCode(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/x", nil), rec)

	api.JSONErrorHandler(c, echo.NewHTTPError(http.StatusNotFound, "nope"))

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if rec.Code != http.StatusNotFound || body["code"] != "http_404" || body["error"] != "Not Found" {
		t.Errorf("got %d %v", rec.Code, body)
	}
}
