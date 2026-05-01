package auth

import "github.com/labstack/echo/v5"

// csrfContextKey is the key Echo's CSRF middleware uses to store the token.
// Echo v5 middleware.CSRF() stores the token under the key "csrf".
const csrfContextKey = "csrf"

// CSRFToken returns the CSRF token string stored by Echo's CSRF middleware.
// Use this in templ components to populate the hidden _csrf input field.
// Returns an empty string when the CSRF middleware is not active on the route.
func CSRFToken(c *echo.Context) string {
	val, _ := c.Get(csrfContextKey).(string)
	return val
}
