package links

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Lokee86/demon-docs/internal/repository"
)

func fileFingerprint(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}

func bytesFingerprint(data []byte) string {
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func recordAbsolute(root string, record FileRecord) string {
	if record.Scope == "repository" {
		return filepath.Clean(filepath.Join(root, filepath.FromSlash(record.Path)))
	}
	return filepath.Clean(filepath.FromSlash(record.Path))
}

func storePath(root, path string) string {
	clean := filepath.Clean(path)
	if repository.Contains(root, clean) {
		relative, _ := filepath.Rel(root, clean)
		return filepath.ToSlash(relative)
	}
	return filepath.ToSlash(clean)
}

func scopeFor(root, path string) string {
	if repository.Contains(root, path) {
		return "repository"
	}
	return "external"
}

func pathKey(path string) string {
	clean := filepath.Clean(path)
	if runtime.GOOS == "windows" {
		return strings.ToLower(clean)
	}
	return clean
}

func kindFromInfo(info os.FileInfo) string {
	if info.IsDir() {
		return "directory"
	}
	return "file"
}

func appendUnique(values []string, value string) []string {
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}
