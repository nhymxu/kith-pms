package api

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func testFunc(c *echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func testFuncRequestID(c *echo.Context) error {
	return c.String(http.StatusOK, c.Response().Header().Get(echo.HeaderXRequestID))
}

func testFuncPrivate(c *echo.Context) error {
	return c.String(http.StatusOK, "Private area")
}
