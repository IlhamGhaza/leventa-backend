package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// FileStorage defines the interface for file storage operations.
type FileStorage interface {
	Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error)
	Delete(ctx context.Context, path string) error
	GetURL(path string) string
}

// LocalStorage implements FileStorage for local filesystem.
type LocalStorage struct {
	basePath string
	maxSize  int64
	subDir   string
	allowed  map[string]string // sniffed MIME type -> stored extension
}

// NewLocalStorage creates a new local storage instance.
func NewLocalStorage(basePath string, maxSize int64) *LocalStorage {
	// Ensure the upload directory exists
	_ = os.MkdirAll(basePath, 0o750)
	return &LocalStorage{basePath: basePath, maxSize: maxSize, subDir: "properties", allowed: extByMime}
}

// documentMimes are the file types accepted for ownership documents.
var documentMimes = map[string]string{
	"application/pdf": ".pdf",
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
}

// NewDocumentStorage creates a private store for ownership documents (PDF, JPEG,
// PNG). Its base path must NOT be served by any static route: files are
// returned only through the authorized download endpoint.
func NewDocumentStorage(basePath string, maxSize int64) *LocalStorage {
	_ = os.MkdirAll(basePath, 0o750)
	return &LocalStorage{basePath: basePath, maxSize: maxSize, subDir: "property-documents", allowed: documentMimes}
}

// Path returns the absolute, traversal-safe filesystem path of a stored file.
func (s *LocalStorage) Path(rel string) (string, error) {
	full := filepath.Clean(filepath.Join(s.basePath, filepath.FromSlash(rel)))
	base := filepath.Clean(s.basePath) + string(filepath.Separator)
	if !strings.HasPrefix(full, base) {
		return "", fmt.Errorf("invalid file path")
	}
	return full, nil
}

// extByMime maps a sniffed MIME type to the only extension we store.
var extByMime = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

// AllowedMimeTypes returns the list of allowed image MIME types.
var AllowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// Upload saves a file to the local filesystem.
func (s *LocalStorage) Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error) {
	if s.maxSize > 0 && header.Size > s.maxSize {
		return "", fmt.Errorf("file too large (max %d bytes)", s.maxSize)
	}

	// Validate the real content type by sniffing, never the client-supplied header.
	head := make([]byte, 512)
	n, err := io.ReadFull(file, head)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	contentType := http.DetectContentType(head[:n])
	ext, ok := s.allowed[contentType]
	if !ok {
		return "", fmt.Errorf("unsupported file type: %s", contentType)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Safe generated filename: extension derived from sniffed type, not user input.
	safeFilename := uuid.New().String() + ext

	// Create subdirectory by date
	subDir := s.subDir
	dirPath := filepath.Join(s.basePath, subDir)
	if err := os.MkdirAll(dirPath, 0o750); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Prevent path traversal
	fullPath := filepath.Join(dirPath, safeFilename)
	fullPath = filepath.Clean(fullPath)
	if !strings.HasPrefix(fullPath, filepath.Clean(s.basePath)+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid file path")
	}

	// Create the file
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy the content
	var written int64
	limited := io.Reader(file)
	if s.maxSize > 0 {
		limited = io.LimitReader(file, s.maxSize+1)
	}
	written, err = io.Copy(dst, limited)
	if err == nil && s.maxSize > 0 && written > s.maxSize {
		err = fmt.Errorf("file too large (max %d bytes)", s.maxSize)
	}
	if err != nil {
		_ = os.Remove(fullPath)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Return relative path
	relativePath := filepath.Join(subDir, safeFilename)
	return filepath.ToSlash(relativePath), nil
}

// Delete removes a file from the local filesystem.
func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	path = strings.TrimPrefix(path, "/uploads/")
	fullPath := filepath.Join(s.basePath, path)
	fullPath = filepath.Clean(fullPath)
	if !strings.HasPrefix(fullPath, filepath.Clean(s.basePath)+string(filepath.Separator)) {
		return fmt.Errorf("invalid file path")
	}
	return os.Remove(fullPath)
}

// GetURL returns the URL for a stored file.
func (s *LocalStorage) GetURL(path string) string {
	return "/uploads/" + path
}
