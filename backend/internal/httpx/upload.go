package httpx

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// MaxUpload is the maximum accepted image size.
const MaxUpload = 5 << 20

var errTooLarge = E(http.StatusRequestEntityTooLarge, "too_large", "File is larger than 5 MB")

// SaveImage reads the multipart field "file", validates it is a JPEG/PNG/WebP image
// (by content sniffing, never by client filename) and stores it in dir under a random name.
// It returns the stored file name.
func SaveImage(w http.ResponseWriter, r *http.Request, dir string) (string, error) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxUpload+64<<10)
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			return "", errTooLarge
		}
		return "", BadRequest("Expected multipart/form-data with a file field")
	}
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()
	f, fh, err := r.FormFile("file")
	if err != nil {
		return "", Validation(map[string]string{"file": "required"})
	}
	defer f.Close()
	if fh.Size > MaxUpload {
		return "", errTooLarge
	}
	data, err := io.ReadAll(io.LimitReader(f, MaxUpload+1))
	if err != nil {
		return "", err
	}
	if len(data) > MaxUpload {
		return "", errTooLarge
	}
	ext := ImageExt(data)
	if ext == "" {
		return "", Validation(map[string]string{"file": "must be a JPG, PNG or WebP image"})
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	name := RandomHex(16) + ext
	if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
		return "", err
	}
	return name, nil
}

// ImageExt returns the file extension for a supported image or "" if unsupported.
func ImageExt(data []byte) string {
	switch http.DetectContentType(data) {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	}
	if len(data) >= 12 && bytes.Equal(data[0:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")) {
		return ".webp"
	}
	return ""
}

// ContentTypeForExt maps stored image extensions to MIME types.
func ContentTypeForExt(name string) string {
	switch filepath.Ext(name) {
	case ".jpg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	}
	return "application/octet-stream"
}

// RandomHex returns n random bytes hex-encoded.
func RandomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
