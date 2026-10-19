package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
)

// envelope is the standard JSON response wrapper.
type envelope struct {
	Data any `json:"data,omitempty"`
	// Error is the English fallback text; clients localize by Code.
	Error  string         `json:"error,omitempty"`
	Code   string         `json:"code,omitempty"`
	Params map[string]any `json:"params,omitempty"`
}

func ok(c *echo.Context, data any) error {
	return c.JSON(http.StatusOK, envelope{Data: data})
}

func created(c *echo.Context, data any) error {
	return c.JSON(http.StatusCreated, envelope{Data: data})
}

func noContent(c *echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

// apiErr writes an error envelope. code is a stable snake_case identifier the
// SPA maps to a localized message; msg stays as the English fallback.
func apiErr(c *echo.Context, status int, code, msg string) error {
	return c.JSON(status, envelope{Error: msg, Code: code})
}

// apiErrParams is apiErr with values the client interpolates into the message.
func apiErrParams(c *echo.Context, status int, code, msg string, params map[string]any) error {
	return c.JSON(status, envelope{Error: msg, Code: code, Params: params})
}

// nullableString returns nil for empty strings, otherwise a pointer to the string.
func nullableString(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}

// parseDateOrDatetime accepts "YYYY-MM-DD" or "YYYY-MM-DDTHH:MM" (datetime-local format).
// Frontend always sends UTC values, so parse in UTC.
func parseDateOrDatetime(s string) (time.Time, error) {
	if len(s) > 10 {
		return time.ParseInLocation("2006-01-02T15:04", s, time.UTC)
	}

	return time.ParseInLocation("2006-01-02", s, time.UTC)
}
