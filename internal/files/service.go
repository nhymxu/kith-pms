package files

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxAvatarSize = 5 * 1024 * 1024 // 5MB
)

var allowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

type FileService interface {
	SaveAvatar(personID int64, file multipart.File, header *multipart.FileHeader) (path string, err error)
	DeleteAvatar(personID int64, path string) error
	GetAvatarPath(personID int64) string
}

type LocalFileService struct {
	BaseDir string
}

func NewLocalFileService(baseDir string) *LocalFileService {
	return &LocalFileService{BaseDir: baseDir}
}

func (s *LocalFileService) SaveAvatar(personID int64, file multipart.File, header *multipart.FileHeader) (string, error) {
	if header.Size > maxAvatarSize {
		return "", fmt.Errorf("file size %d exceeds maximum %d bytes", header.Size, maxAvatarSize)
	}

	// Verify actual file content (magic number check)
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("read file header: %w", err)
	}

	detectedType := http.DetectContentType(buffer[:n])
	if !allowedMimeTypes[detectedType] {
		return "", fmt.Errorf("file content type %s does not match allowed types", detectedType)
	}

	// Reset file pointer for subsequent reads
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, 0); err != nil {
			return "", fmt.Errorf("reset file pointer: %w", err)
		}
	}

	mimeType := header.Header.Get("Content-Type")
	if !allowedMimeTypes[mimeType] {
		return "", fmt.Errorf("unsupported MIME type: %s", mimeType)
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = mimeTypeToExt(mimeType)
	}

	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate random filename: %w", err)
	}
	randomStr := hex.EncodeToString(randomBytes)

	sanitizedName := sanitizeFilename(strings.TrimSuffix(header.Filename, ext))
	filename := fmt.Sprintf("%s-%s%s", randomStr, sanitizedName, ext)

	personDir := filepath.Join(s.BaseDir, fmt.Sprintf("%d", personID))
	if err := os.MkdirAll(personDir, 0755); err != nil {
		return "", fmt.Errorf("create directory: %w", err)
	}

	destPath := filepath.Join(personDir, filename)
	tempPath := destPath + ".tmp"

	dest, err := os.Create(tempPath)
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer dest.Close()

	if _, err := io.Copy(dest, file); err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("write file: %w", err)
	}

	if err := dest.Sync(); err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("sync file: %w", err)
	}

	if err := os.Rename(tempPath, destPath); err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("rename file: %w", err)
	}

	relativePath := filepath.Join(fmt.Sprintf("%d", personID), filename)
	return relativePath, nil
}

func (s *LocalFileService) DeleteAvatar(personID int64, path string) error {
	if path == "" {
		return nil
	}

	fullPath := filepath.Join(s.BaseDir, path)

	cleanPath := filepath.Clean(fullPath)
	if !strings.HasPrefix(cleanPath, filepath.Clean(s.BaseDir)) {
		return fmt.Errorf("invalid path: outside base directory")
	}

	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove file: %w", err)
	}

	personDir := filepath.Join(s.BaseDir, fmt.Sprintf("%d", personID))
	entries, err := os.ReadDir(personDir)
	if err == nil && len(entries) == 0 {
		os.Remove(personDir)
	}

	return nil
}

func (s *LocalFileService) GetAvatarPath(personID int64) string {
	return filepath.Join(s.BaseDir, fmt.Sprintf("%d", personID))
}

func sanitizeFilename(name string) string {
	name = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, name)

	if len(name) > 50 {
		name = name[:50]
	}

	return strings.Trim(name, "-_")
}

func mimeTypeToExt(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ".bin"
	}
}
