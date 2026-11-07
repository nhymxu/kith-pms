package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/auth"
	"github.com/nhymxu/kith-pms/internal/notify"
)

const (
	maxChannelNameLen = 100
	testSendInterval  = 5 * time.Second
	maxChannelBody    = 16 << 10
)

// NotificationsAPI manages notification channels. Mutations and test-send need a
// cookie session: the machine API_TOKEN bearer must not be able to point secrets
// at a new destination or make the server call arbitrary URLs.
type NotificationsAPI struct {
	Repo     *notify.ChannelRepo
	Registry *notify.Registry
	Client   *notify.HTTPClient

	mu       sync.Mutex
	lastTest time.Time
}

type channelView struct {
	ID        int64          `json:"id"`
	Kind      string         `json:"kind"`
	Name      string         `json:"name"`
	Enabled   bool           `json:"enabled"`
	Status    string         `json:"status"` // ok | secret_unreadable
	Config    map[string]any `json:"config"`
	CreatedAt time.Time      `json:"created_at"`
}

type channelList struct {
	Channels []channelView `json:"channels"`
	Kinds    []string      `json:"kinds"`
}

func (h *NotificationsAPI) view(rec *notify.Record) channelView {
	v := channelView{
		ID: rec.ID, Kind: rec.Kind, Name: rec.Name, Enabled: rec.Enabled,
		Status: "ok", Config: map[string]any{}, CreatedAt: rec.CreatedAt,
	}

	if rec.Unreadable {
		v.Status = "secret_unreadable"
		return v
	}

	if ch, found := h.Registry.Get(rec.Kind); found {
		v.Config = notify.PublicView(ch, rec.Config)
	}

	return v
}

func requireSession(c *echo.Context) error {
	if auth.UserFromContext(c) == nil {
		return apiErr(c, http.StatusForbidden, "session_required", "this action needs a signed-in session")
	}

	return nil
}

func channelID(c *echo.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	return id, err == nil && id > 0
}

// loadChannel returns the record, or nil plus the already-written error response.
func (h *NotificationsAPI) loadChannel(c *echo.Context) (*notify.Record, error) {
	id, valid := channelID(c)
	if !valid {
		return nil, apiErr(c, http.StatusBadRequest, "invalid_id", "invalid id")
	}

	rec, err := h.Repo.Get(c.Request().Context(), id)
	if errors.Is(err, notify.ErrChannelNotFound) {
		return nil, apiErr(c, http.StatusNotFound, "channel_not_found", "channel not found")
	}

	if err != nil {
		return nil, apiErr(c, http.StatusInternalServerError, "internal", "internal server error")
	}

	return rec, nil
}

// repoFailure maps a repo error to a response; a channel deleted mid-request is a 404, not a 500.
func repoFailure(c *echo.Context, err error) error {
	if errors.Is(err, notify.ErrChannelNotFound) {
		return apiErr(c, http.StatusNotFound, "channel_not_found", "channel not found")
	}

	return apiErr(c, http.StatusInternalServerError, "internal", "internal server error")
}

func cleanName(s string) (string, bool) {
	s = strings.TrimSpace(s)

	return s, len([]rune(s)) <= maxChannelNameLen
}

// List godoc
//
// @Summary      List notification channels
// @Tags         notifications
// @Produce      json
// @Success      200  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /notifications/channels [get]
func (h *NotificationsAPI) List(c *echo.Context) error {
	recs, err := h.Repo.List(c.Request().Context())
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal", "internal server error")
	}

	out := channelList{Channels: make([]channelView, 0, len(recs)), Kinds: h.Registry.Kinds()}
	for _, rec := range recs {
		out.Channels = append(out.Channels, h.view(rec))
	}

	return ok(c, out)
}

type createChannelReq struct {
	Kind    string          `json:"kind"`
	Name    string          `json:"name"`
	Enabled *bool           `json:"enabled"`
	Config  json.RawMessage `json:"config"`
}

// Create godoc
//
// @Summary      Create a notification channel
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        body  body      createChannelReq  true  "Channel"
// @Success      201   {object}  envelope
// @Failure      403   {object}  envelope
// @Failure      422   {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /notifications/channels [post]
func (h *NotificationsAPI) Create(c *echo.Context) error {
	if err := requireSession(c); err != nil {
		return err
	}

	var req createChannelReq
	if err := json.NewDecoder(http.MaxBytesReader(c.Response(), c.Request().Body, maxChannelBody)).
		Decode(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid_request_body", "invalid request body")
	}

	ch, found := h.Registry.Get(req.Kind)
	name, nameOK := cleanName(req.Name)

	if !found || !nameOK || ch.Validate(req.Config) != nil {
		return apiErr(c, http.StatusUnprocessableEntity, "invalid_channel_config", "invalid channel configuration")
	}

	if notify.HasMaskedSecret(ch, req.Config) {
		return apiErr(c, http.StatusUnprocessableEntity, "secret_required", "enter the real secret values")
	}

	rec, err := h.Repo.Create(c.Request().Context(), req.Kind, name, req.Enabled == nil || *req.Enabled, req.Config)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal", "internal server error")
	}

	return created(c, h.view(rec))
}

// Update godoc
//
// @Summary      Replace a channel's config
// @Description  Omitted or masked secrets keep their stored value; a destination change needs every secret re-sent.
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        id    path      int              true  "Channel ID"
// @Param        body  body      object           true  "Config"
// @Success      200   {object}  envelope
// @Failure      403   {object}  envelope
// @Failure      404   {object}  envelope
// @Failure      422   {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /notifications/channels/{id} [put]
func (h *NotificationsAPI) Update(c *echo.Context) error {
	if err := requireSession(c); err != nil {
		return err
	}

	rec, resp := h.loadChannel(c)
	if rec == nil {
		return resp
	}

	var incoming json.RawMessage
	if err := json.NewDecoder(http.MaxBytesReader(c.Response(), c.Request().Body, maxChannelBody)).
		Decode(&incoming); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid_request_body", "invalid request body")
	}

	ch, found := h.Registry.Get(rec.Kind)
	if !found {
		return apiErr(c, http.StatusUnprocessableEntity, "invalid_channel_config", "invalid channel configuration")
	}

	merged := incoming

	var mergeErr error

	if rec.Unreadable {
		// Nothing to merge with: the user re-enters every secret.
		if ch.Validate(incoming) != nil {
			return apiErr(c, http.StatusUnprocessableEntity, "invalid_channel_config", "invalid channel configuration")
		}

		if notify.HasMaskedSecret(ch, incoming) {
			return apiErr(c, http.StatusUnprocessableEntity, "secret_required", "enter the real secret values")
		}
	} else if merged, mergeErr = notify.MergeConfig(ch, rec.Config, incoming); mergeErr != nil {
		if errors.Is(mergeErr, notify.ErrSecretRequired) {
			return apiErr(
				c,
				http.StatusUnprocessableEntity,
				"secret_required",
				"re-enter the secrets to change the destination",
			)
		}

		return apiErr(c, http.StatusUnprocessableEntity, "invalid_channel_config", "invalid channel configuration")
	}

	if err := h.Repo.UpdateConfig(c.Request().Context(), rec.ID, merged); err != nil {
		return repoFailure(c, err)
	}

	updated, err := h.Repo.Get(c.Request().Context(), rec.ID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal", "internal server error")
	}

	return ok(c, h.view(updated))
}

type patchChannelReq struct {
	Name    *string `json:"name"`
	Enabled *bool   `json:"enabled"`
}

// Patch godoc
//
// @Summary      Rename or enable/disable a channel
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        id    path      int               true  "Channel ID"
// @Param        body  body      patchChannelReq   true  "Fields"
// @Success      200   {object}  envelope
// @Failure      403   {object}  envelope
// @Failure      404   {object}  envelope
// @Failure      422   {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /notifications/channels/{id} [patch]
func (h *NotificationsAPI) Patch(c *echo.Context) error {
	if err := requireSession(c); err != nil {
		return err
	}

	rec, resp := h.loadChannel(c)
	if rec == nil {
		return resp
	}

	var req patchChannelReq
	if err := json.NewDecoder(http.MaxBytesReader(c.Response(), c.Request().Body, maxChannelBody)).
		Decode(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid_request_body", "invalid request body")
	}

	ctx := c.Request().Context()

	if req.Name != nil {
		name, nameOK := cleanName(*req.Name)
		if !nameOK {
			return apiErr(c, http.StatusUnprocessableEntity, "invalid_channel_config", "invalid channel configuration")
		}

		if err := h.Repo.Rename(ctx, rec.ID, name); err != nil {
			return repoFailure(c, err)
		}
	}

	if req.Enabled != nil {
		if err := h.Repo.SetEnabled(ctx, rec.ID, *req.Enabled); err != nil {
			return repoFailure(c, err)
		}
	}

	updated, err := h.Repo.Get(ctx, rec.ID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal", "internal server error")
	}

	return ok(c, h.view(updated))
}

// Delete godoc
//
// @Summary      Delete a channel
// @Tags         notifications
// @Param        id   path  int  true  "Channel ID"
// @Success      204
// @Failure      403  {object}  envelope
// @Failure      404  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /notifications/channels/{id} [delete]
func (h *NotificationsAPI) Delete(c *echo.Context) error {
	if err := requireSession(c); err != nil {
		return err
	}

	rec, resp := h.loadChannel(c)
	if rec == nil {
		return resp
	}

	if err := h.Repo.Delete(c.Request().Context(), rec.ID); err != nil {
		return repoFailure(c, err)
	}

	return noContent(c)
}

type testResult struct {
	OK     bool   `json:"ok"`
	Class  string `json:"class,omitempty"`
	Status int    `json:"status,omitempty"`
}

// allowTest is one global limiter: test-send is an outbound request to a user-chosen URL.
func (h *NotificationsAPI) allowTest() bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if time.Since(h.lastTest) < testSendInterval {
		return false
	}

	h.lastTest = time.Now()

	return true
}

// Test godoc
//
// @Summary      Send a test notification through a channel
// @Tags         notifications
// @Produce      json
// @Param        id   path      int  true  "Channel ID"
// @Success      200  {object}  envelope
// @Failure      403  {object}  envelope
// @Failure      404  {object}  envelope
// @Failure      422  {object}  envelope
// @Failure      429  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /notifications/channels/{id}/test [post]
func (h *NotificationsAPI) Test(c *echo.Context) error {
	if err := requireSession(c); err != nil {
		return err
	}

	rec, resp := h.loadChannel(c)
	if rec == nil {
		return resp
	}

	if rec.Unreadable {
		return apiErr(c, http.StatusUnprocessableEntity, "channel_secret_unreadable",
			"stored secrets cannot be read; re-enter them")
	}

	if !h.allowTest() {
		return apiErr(c, http.StatusTooManyRequests, "rate_limited", "wait a few seconds before testing again")
	}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(c.Request().Context()), 10*time.Second)
	defer cancel()

	se := notify.SendNow(ctx, h.Registry, h.Client, rec, notify.Message{
		Title:    "kith test notification",
		Notes:    "This is a test from kith.",
		DueAt:    time.Now().UTC(),
		Location: time.UTC,
	})
	if se == nil {
		return ok(c, testResult{OK: true})
	}

	return ok(c, testResult{Class: se.Class, Status: se.Status})
}
