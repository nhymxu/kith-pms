package api

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

// envelope is the standard JSON response wrapper.
type envelope struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
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

func apiErr(c *echo.Context, code int, msg string) error {
	return c.JSON(code, envelope{Error: msg})
}
