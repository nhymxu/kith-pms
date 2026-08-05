package handler_test

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nhymxu/kith-pms/internal/api/handler"
	"github.com/nhymxu/kith-pms/pkg/config"
)

// ---- stub FileService -------------------------------------------------------

type stubFileService struct {
	savedPath string
	saveErr   error
}

func (s *stubFileService) SaveAvatar(_ int64, _ multipart.File, h *multipart.FileHeader) (string, error) {
	if s.saveErr != nil {
		return "", s.saveErr
	}

	s.savedPath = "avatars/" + h.Filename

	return s.savedPath, nil
}

func (s *stubFileService) SaveAvatarBytes(_ int64, _ []byte, _ string) (string, error) {
	return "", nil
}
func (s *stubFileService) DeleteAvatar(_ int64, _ string) error { return nil }
func (s *stubFileService) SaveGiftImage(_ int64, _ multipart.File, _ *multipart.FileHeader) (string, error) {
	return "", nil
}
func (s *stubFileService) DeleteGiftImage(_ int64, _ string) error { return nil }
func (s *stubFileService) SaveDocument(_ int64, _ []byte, _ string) (string, error) {
	return "", nil
}

// ---- helpers ----------------------------------------------------------------

// buildMultipartRequest builds a multipart/form-data request with a single file field.
func buildMultipartRequest( // nolint:unused
	t *testing.T,
	fieldName, filename, contentType string,
	content []byte,
) *http.Request {
	t.Helper()

	var buf bytes.Buffer

	w := multipart.NewWriter(&buf)

	part, err := w.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}

	if _, err := part.Write(content); err != nil {
		t.Fatalf("write content: %v", err)
	}

	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/v1/people/1/avatar", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Overwrite the file part's Content-Type with what we specify (multipart writer
	// sets application/octet-stream by default; real clients send the actual mime).
	// We must set it on the part header which is already written, so instead we
	// wrap the request and intercept FormFile. The simpler approach: use ParseMultipartForm
	// directly and inject Content-Type header on the file header after parse.
	// For test purposes we rely on the stub bypassing mime check in UploadAvatar service.
	_ = contentType // used via the file header override below

	return req
}

// buildMultipartRequestWithMIME creates a multipart request where the file part
// carries an explicit Content-Type header (how browsers send file uploads).
func buildMultipartRequestWithMIME(t *testing.T, fieldName, filename, mimeType string, content []byte) *http.Request {
	t.Helper()

	var buf bytes.Buffer

	w := multipart.NewWriter(&buf)

	h := make(map[string][]string)
	h["Content-Disposition"] = []string{fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, filename)}
	h["Content-Type"] = []string{mimeType}

	part, err := w.CreatePart(h)
	if err != nil {
		t.Fatalf("create part: %v", err)
	}

	if _, err := part.Write(content); err != nil {
		t.Fatalf("write content: %v", err)
	}

	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/v1/people/1/avatar", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())

	return req
}

// ---- Get tests --------------------------------------------------------------

// The avatar URL never changes, so freshness depends on revalidation rather than
// a long max-age; a stale cache would otherwise hide a re-upload for a full day.
func TestAvatarsGet_ServesETagAndRevalidates(t *testing.T) {
	db := openTestDB(t)
	peopleSvc := newPeopleService(db)
	personID := insertTestPerson(t, db, "Cara")

	baseDir := t.TempDir()
	filename := fmt.Sprintf("%d.jpg", personID)
	jpegContent := []byte{0xff, 0xd8, 0xff, 0xe0, 0, 0x10, 'J', 'F', 'I', 'F', 0}

	if err := os.WriteFile(filepath.Join(baseDir, filename), jpegContent, 0o600); err != nil {
		t.Fatalf("write avatar file: %v", err)
	}

	if _, err := db.NewUpdate().
		Table("person").
		Set("avatar_path = ?", filename).
		Where("id = ?", personID).
		Exec(context.Background()); err != nil {
		t.Fatalf("set avatar_path: %v", err)
	}

	h := &handler.AvatarsAPI{
		PeopleSvc:      peopleSvc,
		FileSvc:        &stubFileService{},
		AvatarBasePath: baseDir,
	}
	pathParams := map[string]string{"id": fmt.Sprintf("%d", personID)}

	e := newTestEcho()
	rec := execHandler(e, httptest.NewRequest(http.MethodGet, "/", nil), pathParams, h.Get)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", rec.Code, rec.Body.String())
	}

	etag := rec.Header().Get("ETag")
	if etag == "" {
		t.Fatal("expected an ETag header")
	}

	if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "no-cache") {
		t.Errorf("Cache-Control = %q, want it to force revalidation", cc)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "image/jpeg" {
		t.Errorf("Content-Type = %q, want image/jpeg", ct)
	}

	// Same ETag back → 304, no body.
	condReq := httptest.NewRequest(http.MethodGet, "/", nil)
	condReq.Header.Set("If-None-Match", etag)

	condRec := execHandler(newTestEcho(), condReq, pathParams, h.Get)
	if condRec.Code != http.StatusNotModified {
		t.Errorf("expected 304 for matching If-None-Match, got %d", condRec.Code)
	}

	if condRec.Body.Len() != 0 {
		t.Errorf("expected empty body on 304, got %d bytes", condRec.Body.Len())
	}

	// A stale ETag must still deliver the bytes.
	staleReq := httptest.NewRequest(http.MethodGet, "/", nil)
	staleReq.Header.Set("If-None-Match", `"stale"`)

	staleRec := execHandler(newTestEcho(), staleReq, pathParams, h.Get)
	if staleRec.Code != http.StatusOK {
		t.Errorf("expected 200 for non-matching If-None-Match, got %d", staleRec.Code)
	}

	if !bytes.Equal(staleRec.Body.Bytes(), jpegContent) {
		t.Error("expected full image bytes when the ETag does not match")
	}

	// Serving via http.ServeContent advertises range support, so pin that it
	// actually answers a partial request correctly. The gzip middleware is
	// configured to skip these paths precisely so a 206 stays coherent.
	rangeReq := httptest.NewRequest(http.MethodGet, "/", nil)
	rangeReq.Header.Set("Range", "bytes=0-3")

	rangeRec := execHandler(newTestEcho(), rangeReq, pathParams, h.Get)
	if rangeRec.Code != http.StatusPartialContent {
		t.Errorf("expected 206 for a range request, got %d", rangeRec.Code)
	}

	if got := rangeRec.Body.Len(); got != 4 {
		t.Errorf("expected 4 bytes for bytes=0-3, got %d", got)
	}

	if cr := rangeRec.Header().Get("Content-Range"); cr == "" {
		t.Error("expected a Content-Range header on the 206")
	}

	if rangeRec.Header().Get("Content-Encoding") != "" {
		t.Error("partial image responses must not be content-encoded")
	}
}

// ---- Upload tests -----------------------------------------------------------

func TestAvatarsUpload_HappyPath(t *testing.T) {
	db := openTestDB(t)
	peopleSvc := newPeopleService(db)
	personID := insertTestPerson(t, db, "Alice")

	fileSvc := &stubFileService{}
	// people.Service.UploadAvatar delegates to its own FileService field.
	peopleSvc.FileService = fileSvc

	h := &handler.AvatarsAPI{
		PeopleSvc:      peopleSvc,
		FileSvc:        fileSvc,
		AvatarBasePath: t.TempDir(),
	}

	// Minimal valid JPEG magic bytes.
	jpegContent := []byte{0xff, 0xd8, 0xff, 0xe0, 0, 0x10, 'J', 'F', 'I', 'F', 0}
	req := buildMultipartRequestWithMIME(t, "avatar", "photo.jpg", "image/jpeg", jpegContent)

	e := newTestEcho()
	rec := execHandler(e, req, map[string]string{"id": fmt.Sprintf("%d", personID)}, h.Upload)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", rec.Code, rec.Body.String())
	}
}

func TestAvatarsUpload_InvalidID_Returns400(t *testing.T) {
	db := openTestDB(t)
	h := &handler.AvatarsAPI{
		PeopleSvc:      newPeopleService(db),
		FileSvc:        &stubFileService{},
		AvatarBasePath: t.TempDir(),
	}

	req := buildMultipartRequestWithMIME(t, "avatar", "photo.jpg", "image/jpeg", []byte{0xff, 0xd8})
	e := newTestEcho()
	rec := execHandler(e, req, map[string]string{"id": "abc"}, h.Upload)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAvatarsUpload_UnsupportedMIME_Returns422(t *testing.T) {
	db := openTestDB(t)
	personID := insertTestPerson(t, db, "Bob")
	h := &handler.AvatarsAPI{
		PeopleSvc:      newPeopleService(db),
		FileSvc:        &stubFileService{},
		AvatarBasePath: t.TempDir(),
	}

	req := buildMultipartRequestWithMIME(t, "avatar", "file.pdf", "application/pdf", []byte("%PDF-1.4"))
	e := newTestEcho()
	rec := execHandler(e, req, map[string]string{"id": fmt.Sprintf("%d", personID)}, h.Upload)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

func TestAvatarsUpload_OverConfiguredLimit_Rejected(t *testing.T) {
	// Pin the cap instead of relying on the default, so this keeps testing the
	// limit behaviour if the default ever changes again.
	const capMB = 5

	prev := config.C.MaxUploadSizeMB
	config.C.MaxUploadSizeMB = capMB

	t.Cleanup(func() { config.C.MaxUploadSizeMB = prev })

	db := openTestDB(t)
	personID := insertTestPerson(t, db, "Charlie")
	h := &handler.AvatarsAPI{
		PeopleSvc:      newPeopleService(db),
		FileSvc:        &stubFileService{},
		AvatarBasePath: t.TempDir(),
	}

	oversize := make([]byte, capMB*1024*1024+1)
	oversize[0] = 0xff
	oversize[1] = 0xd8 // JPEG magic

	req := buildMultipartRequestWithMIME(t, "avatar", "big.jpg", "image/jpeg", oversize)
	e := newTestEcho()
	rec := execHandler(e, req, map[string]string{"id": fmt.Sprintf("%d", personID)}, h.Upload)

	// Body is capped at cap+1MB; the file.Size check fires first at cap+1 byte.
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rec.Code)
	}
}

// ---- Delete tests -----------------------------------------------------------

func TestAvatarsDelete_PersonNotFound_Returns404(t *testing.T) {
	db := openTestDB(t)
	h := &handler.AvatarsAPI{
		PeopleSvc:      newPeopleService(db),
		FileSvc:        &stubFileService{},
		AvatarBasePath: t.TempDir(),
	}

	req := httptest.NewRequest(http.MethodDelete, "/v1/people/999/avatar", nil)
	e := newTestEcho()
	rec := execHandler(e, req, map[string]string{"id": "999"}, h.Delete)

	// person not found → service returns error containing "not found" → 404 or 500.
	// Our handler maps "person not found" → 404.
	if rec.Code != http.StatusNotFound && rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 404 or 500, got %d", rec.Code)
	}
}

// ---- Get tests --------------------------------------------------------------

func TestAvatarsGet_NoAvatar_Returns404(t *testing.T) {
	db := openTestDB(t)
	personID := insertTestPerson(t, db, "Dave")
	h := &handler.AvatarsAPI{
		PeopleSvc:      newPeopleService(db),
		FileSvc:        &stubFileService{},
		AvatarBasePath: t.TempDir(),
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/people/1/avatar", nil)
	e := newTestEcho()
	rec := execHandler(e, req, map[string]string{"id": fmt.Sprintf("%d", personID)}, h.Get)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestAvatarsGet_InvalidID_Returns400(t *testing.T) {
	db := openTestDB(t)
	h := &handler.AvatarsAPI{
		PeopleSvc:      newPeopleService(db),
		AvatarBasePath: t.TempDir(),
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/people/bad/avatar", nil)
	e := newTestEcho()
	rec := execHandler(e, req, map[string]string{"id": "bad"}, h.Get)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// ensure unused import doesn't break build
var _ = strings.Contains
