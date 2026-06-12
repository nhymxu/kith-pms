package handler

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/files"
	"github.com/nhymxu/kith-pms/internal/people"
)

// sniff512 reads the first 512 bytes from r to detect its content type, then
// seeks back to the start. Returns http.DetectContentType result.
func sniff512(r io.ReadSeeker) (string, error) {
	buf := make([]byte, 512)

	n, err := r.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}

	detected := http.DetectContentType(buf[:n])

	if _, seekErr := r.Seek(0, io.SeekStart); seekErr != nil {
		return "", seekErr
	}

	return detected, nil
}

const maxAvatarBytes = 5 * 1024 * 1024 // 5 MB

var avatarAllowedTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// AvatarsAPI handles /v1/people/:id/avatar endpoints.
type AvatarsAPI struct {
	PeopleSvc      *people.Service
	FileSvc        files.FileService
	AvatarBasePath string
}

// Upload handles POST /v1/people/:id/avatar.
//
// @Summary      Upload avatar
// @Tags         people
// @Accept       multipart/form-data
// @Produce      json
// @Param        id      path      int   true  "Person ID"
// @Param        avatar  formData  file  true  "Avatar image (jpeg/png/gif/webp, max 5MB)"
// @Success      200     {object}  envelope{data=object{uploaded=bool}}
// @Failure      400     {object}  envelope
// @Failure      413     {object}  envelope
// @Failure      422     {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /people/{id}/avatar [post]
func (h *AvatarsAPI) Upload(c *echo.Context) error {
	personID, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	// Cap request body to 5MB + 1MB multipart overhead.
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, maxAvatarBytes+1024*1024)

	file, err := c.FormFile("avatar")
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			return apiErr(c, http.StatusRequestEntityTooLarge, "file too large (max 5MB)")
		}

		return apiErr(c, http.StatusBadRequest, "no file uploaded")
	}

	if file.Size > maxAvatarBytes {
		return apiErr(c, http.StatusRequestEntityTooLarge, "file too large (max 5MB)")
	}

	src, err := file.Open()
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "failed to open file")
	}
	defer func() { _ = src.Close() }()

	// Detect content type from actual bytes — don't trust client-supplied header.
	detected, err := sniff512(src)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "failed to read file")
	}
	// DetectContentType may append params (e.g. "image/png; charset=...") — strip them.
	detected = strings.SplitN(detected, ";", 2)[0]

	detected = strings.TrimSpace(detected)
	if !avatarAllowedTypes[detected] {
		return apiErr(c, http.StatusUnprocessableEntity, "unsupported file type: use jpeg, png, gif, or webp")
	}

	if err := h.PeopleSvc.UploadAvatar(c.Request().Context(), personID, src, file); err != nil {
		if strings.Contains(err.Error(), "person not found") {
			return apiErr(c, http.StatusNotFound, "person not found")
		}

		return apiErr(c, http.StatusBadRequest, err.Error())
	}

	return ok(c, map[string]any{"uploaded": true})
}

// Delete handles DELETE /v1/people/:id/avatar.
//
// @Summary      Delete avatar
// @Tags         people
// @Produce      json
// @Param        id   path  int  true  "Person ID"
// @Success      204
// @Failure      400  {object}  envelope
// @Failure      404  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /people/{id}/avatar [delete]
func (h *AvatarsAPI) Delete(c *echo.Context) error {
	personID, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	if err := h.PeopleSvc.DeleteAvatar(c.Request().Context(), personID); err != nil {
		if strings.Contains(err.Error(), "person not found") {
			return apiErr(c, http.StatusNotFound, "person not found")
		}

		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return noContent(c)
}

// Get handles GET /v1/people/:id/avatar.
//
// @Summary      Get avatar image
// @Tags         people
// @Produce      image/jpeg
// @Param        id   path  int  true  "Person ID"
// @Success      200
// @Failure      400  {object}  envelope
// @Failure      404  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /people/{id}/avatar [get]
func (h *AvatarsAPI) Get(c *echo.Context) error {
	personID, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	p, err := h.PeopleSvc.Get(c.Request().Context(), personID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	if p == nil || p.AvatarPath == "" {
		return apiErr(c, http.StatusNotFound, "no avatar")
	}

	fullPath := filepath.Join(h.AvatarBasePath, p.AvatarPath)
	cleanPath := filepath.Clean(fullPath)

	if !strings.HasPrefix(cleanPath, filepath.Clean(h.AvatarBasePath)) {
		return apiErr(c, http.StatusNotFound, "not found")
	}

	f, err := os.Open(cleanPath)
	if err != nil {
		return apiErr(c, http.StatusNotFound, "avatar file not found")
	}
	defer func() { _ = f.Close() }()

	mt := mime.TypeByExtension(filepath.Ext(cleanPath))
	if mt == "" {
		mt = "application/octet-stream"
	}

	c.Response().Header().Set("Content-Type", mt)
	c.Response().Header().Set("Cache-Control", "public, max-age=86400")
	_, err = io.Copy(c.Response(), f)

	return err
}
