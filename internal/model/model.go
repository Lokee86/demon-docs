package model

import "path/filepath"

type FolderInfo struct {
	Path        string
	IndexPath   string
	DirectFiles []string
	StubFiles   []string
	Subfolders  []string
	IsStubs     bool
}

type DocsTree struct {
	Root    string
	Folders map[string]*FolderInfo
}

type IndexEntry struct {
	IndexPath    string
	Section      string
	LinkText     string
	LinkTarget   string
	Description  string
	OriginalLine string
}

type FileUpdate struct {
	Path    string
	OldText *string
	NewText string
}

type ReconcileResult struct {
	Updates  []FileUpdate
	Messages []string
}

func CleanAbs(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.Clean(absolute), nil
}
