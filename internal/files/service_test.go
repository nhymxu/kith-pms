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

	if !strings.HasPrefix(path, "123/") {
		t.Errorf("path = %q, want prefix '123/'", path)
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

	personDir := filepath.Join(tempDir, "123")
	if err := os.MkdirAll(personDir, 0755); err != nil {
		t.Fatalf("create person dir: %v", err)
	}

	testFile := filepath.Join(personDir, "test-avatar.jpg")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	err := svc.DeleteAvatar(123, "123/test-avatar.jpg")
	if err != nil {
		t.Fatalf("DeleteAvatar: %v", err)
	}

	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("file should be deleted")
	}

	if _, err := os.Stat(personDir); !os.IsNotExist(err) {
		t.Error("empty person directory should be removed")
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
