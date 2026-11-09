package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/notify"
)

// WebPushAPI exposes the VAPID public key the browser needs to subscribe, and rotation.
type WebPushAPI struct {
	Store *notify.VAPIDStore
}

type vapidKeyView struct {
	PublicKey string `json:"public_key"`
}

// Key godoc
//
// @Summary      VAPID public key for push subscriptions
// @Tags         notifications
// @Produce      json
// @Success      200  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /notifications/webpush/key [get]
func (h *WebPushAPI) Key(c *echo.Context) error {
	_, pub, err := h.Store.Keys(c.Request().Context())
	if errors.Is(err, notify.ErrSecretUnreadable) {
		return apiErr(c, http.StatusUnprocessableEntity, "vapid_key_unreadable",
			"push keys cannot be read; rotate them")
	}

	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal", "internal server error")
	}

	return ok(c, vapidKeyView{PublicKey: pub})
}

// Rotate godoc
//
// @Summary      Rotate VAPID keys
// @Description  Deletes every webpush channel: subscriptions are bound to the old key.
// @Tags         notifications
// @Produce      json
// @Success      200  {object}  envelope
// @Failure      403  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /notifications/webpush/rotate [post]
func (h *WebPushAPI) Rotate(c *echo.Context) error {
	if err := requireSession(c); err != nil {
		return err
	}

	pub, err := h.Store.Rotate(c.Request().Context())
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal", "internal server error")
	}

	return ok(c, vapidKeyView{PublicKey: pub})
}
