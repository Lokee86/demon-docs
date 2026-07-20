package filetxn

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"runtime"
	"strings"
)

type Rewrite struct {
	path              string
	expectedOldSHA256 string
	expectedNewSHA256 string
	oldData           []byte
	newData           []byte
	prepared          bool
}

type Suppression struct {
	Path              string
	ExpectedOldSHA256 string
	ExpectedNewSHA256 string
}

func New(path string, oldData, newData []byte) Rewrite {
	return Rewrite{
		path:              filepath.Clean(path),
		expectedOldSHA256: Digest(oldData),
		expectedNewSHA256: Digest(newData),
		oldData:           append([]byte{}, oldData...),
		newData:           append([]byte{}, newData...),
		prepared:          true,
	}
}

func (rewrite Rewrite) Path() string { return rewrite.path }

func (rewrite Rewrite) ExpectedOldSHA256() string { return rewrite.expectedOldSHA256 }

func (rewrite Rewrite) ExpectedNewSHA256() string { return rewrite.expectedNewSHA256 }

func (rewrite Rewrite) OldData() []byte {
	return append([]byte(nil), rewrite.oldData...)
}

func (rewrite Rewrite) NewData() []byte {
	return append([]byte(nil), rewrite.newData...)
}

func (rewrite Rewrite) Prepared() bool {
	return rewrite.prepared
}

func Digest(data []byte) string {
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func pathKey(path string) string {
	clean := filepath.Clean(path)
	if runtime.GOOS == "windows" {
		return strings.ToLower(clean)
	}
	return clean
}
