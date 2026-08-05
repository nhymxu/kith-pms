package handler

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v5"
)

// serveImageFile writes a stored image with revalidation-based caching.
//
// The URL for a person's avatar or a gift's image never changes, so a long
// max-age served stale bytes after a re-upload — up to a full day on every page
// except the uploader itself. An ETag derived from size+mtime lets the browser
// keep the bytes and revalidate with a cheap 304 instead.
//
// Cache-Control is private because these images sit behind session auth and must
// never be retained by a shared proxy cache. "no-cache" means "revalidate before
// reuse", not "do not store".
func serveImageFile(c *echo.Context, f *os.File, path string) {
	mt := mime.TypeByExtension(filepath.Ext(path))
	if mt == "" {
		mt = "application/octet-stream"
	}

	header := c.Response().Header()
	header.Set("Content-Type", mt)
	header.Set("Cache-Control", "private, no-cache")

	// A stat failure is not fatal: fall back to serving without validators.
	var modTime time.Time

	if st, err := f.Stat(); err == nil {
		modTime = st.ModTime()
		header.Set("ETag", fmt.Sprintf(`"%x-%x"`, st.Size(), modTime.UnixNano()))
	}

	// ServeContent honours If-None-Match against the ETag set above and replies
	// 304 itself; it also gives us Range support for free.
	http.ServeContent(c.Response(), c.Request(), filepath.Base(path), modTime, f)
}
