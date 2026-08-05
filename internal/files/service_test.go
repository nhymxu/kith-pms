package files

import (
	"bytes"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalFileService_SaveAvatar(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewLocalFileService(tempDir)

	content := []byte{
		0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01,
		0x01, 0x01, 0x00, 0x48, 0x00, 0x48, 0x00, 0x00, 0xff, 0xd9,
	}
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)

	part, err := writer.CreateFormFile("avatar", "test-photo.jpg")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}

	part.Write(content)
	writer.Close()

	reader := multipart.NewReader(buf, writer.Boundary())

	form, err := reader.ReadForm(maxAvatarSize)
	if err != nil {
		t.Fatalf("read form: %v", err)
	}
	defer form.RemoveAll()

	files := form.File["avatar"]
	if len(files) == 0 {
		t.Fatal("no files in form")
	}

	fileHeader := files[0]
	fileHeader.Header.Set("Content-Type", "image/jpeg")

	file, err := fileHeader.Open()
	if err != nil {
		t.Fatalf("open file: %v", err)
	}
	defer file.Close()

	path, err := svc.SaveAvatar(123, file, fileHeader)
	if err != nil {
		t.Fatalf("SaveAvatar: %v", err)
	}

	if path == "" {
		t.Fatal("expected non-empty path")
	}

	if path != "123.jpg" {
		t.Errorf("path = %q, want %q", path, "123.jpg")
	}

	fullPath := filepath.Join(tempDir, path)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Errorf("file not created at %s", fullPath)
	}

	savedContent, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}

	if !bytes.Equal(savedContent, content) {
		t.Error("saved content does not match original")
	}
}

func TestLocalFileService_SaveAvatar_SizeLimit(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewLocalFileService(tempDir)

	content := make([]byte, maxAvatarSize+1)
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)

	part, err := writer.CreateFormFile("avatar", "large.jpg")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}

	part.Write(content)
	writer.Close()

	reader := multipart.NewReader(buf, writer.Boundary())

	form, err := reader.ReadForm(maxAvatarSize + 1024)
	if err != nil {
		t.Fatalf("read form: %v", err)
	}
	defer form.RemoveAll()

	fileHeader := form.File["avatar"][0]
	fileHeader.Header.Set("Content-Type", "image/jpeg")

	file, err := fileHeader.Open()
	if err != nil {
		t.Fatalf("open file: %v", err)
	}
	defer file.Close()

	_, err = svc.SaveAvatar(123, file, fileHeader)
	if err == nil {
		t.Error("expected error for oversized file")
	}

	if !strings.Contains(err.Error(), "exceeds maximum") {
		t.Errorf("error = %v, want 'exceeds maximum'", err)
	}
}

func TestLocalFileService_SaveAvatar_InvalidMimeType(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewLocalFileService(tempDir)

	content := []byte{
		0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01,
		0x01, 0x01, 0x00, 0x48, 0x00, 0x48, 0x00, 0x00, 0xff, 0xd9,
	}
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)

	part, err := writer.CreateFormFile("avatar", "test.txt")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}

	part.Write(content)
	writer.Close()

	reader := multipart.NewReader(buf, writer.Boundary())

	form, err := reader.ReadForm(maxAvatarSize)
	if err != nil {
		t.Fatalf("read form: %v", err)
	}
	defer form.RemoveAll()

	fileHeader := form.File["avatar"][0]
	fileHeader.Header.Set("Content-Type", "text/plain")

	file, err := fileHeader.Open()
	if err != nil {
		t.Fatalf("open file: %v", err)
	}
	defer file.Close()

	_, err = svc.SaveAvatar(123, file, fileHeader)
	if err == nil {
		t.Error("expected error for invalid MIME type")
	}

	if !strings.Contains(err.Error(), "unsupported MIME type") {
		t.Errorf("error = %v, want 'unsupported MIME type'", err)
	}
}

func TestLocalFileService_DeleteAvatar(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewLocalFileService(tempDir)

	testFile := filepath.Join(tempDir, "123.jpg")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	err := svc.DeleteAvatar(123, "123.jpg")
	if err != nil {
		t.Fatalf("DeleteAvatar: %v", err)
	}

	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("file should be deleted")
	}
}

func TestLocalFileService_SaveAvatar_SameExtOverwrite(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewLocalFileService(tempDir)

	firstContent := []byte{
		0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01,
		0x01, 0x01, 0x00, 0x48, 0x00, 0x48, 0x00, 0x00, 0xff, 0xd9,
	}
	secondContent := append([]byte{}, firstContent...)
	secondContent = append(secondContent, 0x00)

	firstPath := saveJPEG(t, svc, 5, firstContent)
	secondPath := saveJPEG(t, svc, 5, secondContent)

	if firstPath != secondPath {
		t.Errorf("expected same path across re-uploads, got %q then %q", firstPath, secondPath)
	}

	if firstPath != "5.jpg" {
		t.Errorf("path = %q, want %q", firstPath, "5.jpg")
	}

	saved, err := os.ReadFile(filepath.Join(tempDir, secondPath))
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}

	if !bytes.Equal(saved, secondContent) {
		t.Error("expected second upload's content to overwrite the first")
	}
}

// jpegBytes and pngBytes are minimal headers that http.DetectContentType
// recognises; the tests never decode them.
var (
	jpegBytes = []byte{
		0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01,
		0x01, 0x01, 0x00, 0x48, 0x00, 0x48, 0x00, 0x00, 0xff, 0xd9,
	}
	pngBytes = []byte{
		0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 'I', 'H', 'D', 'R',
	}
)

func saveJPEG(t *testing.T, svc *LocalFileService, personID int64, content []byte) string {
	t.Helper()

	return saveAvatarAs(t, svc, personID, content, "image/jpeg", "photo.jpg")
}

func saveAvatarAs(
	t *testing.T,
	svc *LocalFileService,
	personID int64,
	content []byte,
	mimeType string,
	filename string,
) string {
	t.Helper()

	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)

	part, err := writer.CreateFormFile("avatar", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}

	part.Write(content)
	writer.Close()

	reader := multipart.NewReader(buf, writer.Boundary())

	form, err := reader.ReadForm(maxAvatarSize)
	if err != nil {
		t.Fatalf("read form: %v", err)
	}
	defer form.RemoveAll()

	fileHeader := form.File["avatar"][0]
	fileHeader.Header.Set("Content-Type", mimeType)

	file, err := fileHeader.Open()
	if err != nil {
		t.Fatalf("open file: %v", err)
	}
	defer file.Close()

	path, err := svc.SaveAvatar(personID, file, fileHeader)
	if err != nil {
		t.Fatalf("SaveAvatar: %v", err)
	}

	return path
}

// A format change writes a new deterministic name. Pruning the previous file is
// the caller's job, post-commit (people.Service.UploadAvatar), so that a failed
// avatar transaction cannot destroy the avatar that is still referenced by the
// DB. SaveAvatar itself must leave the old file alone.
func TestLocalFileService_SaveAvatar_FormatChangeLeavesPruningToCaller(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewLocalFileService(tempDir)

	if got := saveAvatarAs(t, svc, 7, pngBytes, "image/png", "photo.png"); got != "7.png" {
		t.Fatalf("first save path = %q, want %q", got, "7.png")
	}

	if got := saveJPEG(t, svc, 7, jpegBytes); got != "7.jpg" {
		t.Fatalf("re-upload path = %q, want %q", got, "7.jpg")
	}

	if _, err := os.Stat(filepath.Join(tempDir, "7.jpg")); err != nil {
		t.Errorf("expected 7.jpg to exist: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tempDir, "7.png")); err != nil {
		t.Errorf("expected 7.png to survive the write so the caller can prune it: %v", err)
	}
}

func TestLocalFileService_DeleteAvatar_PathTraversal(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewLocalFileService(tempDir)

	err := svc.DeleteAvatar(123, "../../../etc/passwd")
	if err == nil {
		t.Error("expected error for path traversal attempt")
	}

	if !strings.Contains(err.Error(), "outside base directory") {
		t.Errorf("error = %v, want 'outside base directory'", err)
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"normal-file", "normal-file"},
		{"file with spaces", "file-with-spaces"},
		{"file@#$%name", "file----name"},
		{"verylongfilenamethatshouldbetruncatedtopreventissues", "verylongfilenamethatshouldbetruncatedtopreventissu"},
		{"---leading-dashes", "leading-dashes"},
		{"trailing-dashes---", "trailing-dashes"},
	}

	for _, tt := range tests {
		got := sanitizeFilename(tt.input)
		if got != tt.want {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestLocalFileService_SaveDocument(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewLocalFileService(tempDir)

	data := []byte("fake-pdf-content")

	path, err := svc.SaveDocument(42, data, "report.pdf")
	if err != nil {
		t.Fatalf("SaveDocument: %v", err)
	}

	if path == "" {
		t.Fatal("expected non-empty path")
	}

	if !strings.HasPrefix(path, "documents/42/") {
		t.Errorf("path = %q, want prefix 'documents/42/'", path)
	}

	if !strings.HasSuffix(path, ".pdf") {
		t.Errorf("path = %q, want .pdf extension", path)
	}

	fullPath := filepath.Join(tempDir, path)

	saved, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}

	if !bytes.Equal(saved, data) {
		t.Error("saved content does not match")
	}
}

func TestLocalFileService_SaveDocument_SizeLimit(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewLocalFileService(tempDir)

	data := make([]byte, maxDocumentSize+1)

	_, err := svc.SaveDocument(1, data, "big.bin")
	if err == nil {
		t.Error("expected error for oversized document")
	}

	if !strings.Contains(err.Error(), "exceeds maximum") {
		t.Errorf("error = %v, want 'exceeds maximum'", err)
	}
}

func TestLocalFileService_SaveDocument_AnyMimeType(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewLocalFileService(tempDir)

	// Any file type must be accepted (no mime allowlist for documents).
	for _, name := range []string{"doc.pdf", "spreadsheet.xlsx", "archive.zip", "image.png"} {
		if _, err := svc.SaveDocument(1, []byte("data"), name); err != nil {
			t.Errorf("SaveDocument(%q) unexpected error: %v", name, err)
		}
	}
}

func TestMimeTypeToExt(t *testing.T) {
	tests := []struct {
		mimeType string
		want     string
	}{
		{"image/jpeg", ".jpg"},
		{"image/png", ".png"},
		{"image/gif", ".gif"},
		{"image/webp", ".webp"},
		{"application/octet-stream", ".bin"},
	}

	for _, tt := range tests {
		got := mimeTypeToExt(tt.mimeType)
		if got != tt.want {
			t.Errorf("mimeTypeToExt(%q) = %q, want %q", tt.mimeType, got, tt.want)
		}
	}
}
