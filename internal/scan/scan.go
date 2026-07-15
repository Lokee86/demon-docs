package scan

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/Lokee86/doc-ledger/internal/config"
	"github.com/Lokee86/doc-ledger/internal/model"
	"github.com/Lokee86/doc-ledger/internal/pathutil"
)

func Tree(root string, c config.Config) (model.DocsTree, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return model.DocsTree{}, err
	}
	tree := model.DocsTree{Root: filepath.Clean(abs), Folders: map[string]*model.FolderInfo{}}
	var walk func(string) error
	walk = func(folder string) error {
		entries, err := os.ReadDir(folder)
		if err != nil {
			return err
		}
		stub := filepath.Base(folder) == c.Draft.Folder
		info := &model.FolderInfo{Path: folder, IsStubs: stub}
		if !stub {
			info.IndexPath = filepath.Join(folder, c.IndexFile)
		}
		for _, entry := range entries {
			p := filepath.Join(folder, entry.Name())
			if entry.IsDir() {
				if !stub && entry.Name() == c.Draft.Folder {
					continue
				}
				info.Subfolders = append(info.Subfolders, p)
				continue
			}
			ok, err := IsIndexable(tree.Root, p, c)
			if err != nil {
				return err
			}
			if ok {
				info.DirectFiles = append(info.DirectFiles, p)
			}
		}
		if !stub {
			stubDir := filepath.Join(folder, c.Draft.Folder)
			if st, err := os.Stat(stubDir); err == nil && st.IsDir() {
				stubEntries, err := os.ReadDir(stubDir)
				if err != nil {
					return err
				}
				for _, entry := range stubEntries {
					if entry.IsDir() {
						continue
					}
					p := filepath.Join(stubDir, entry.Name())
					ok, err := IsIndexable(tree.Root, p, c)
					if err != nil {
						return err
					}
					if ok {
						info.StubFiles = append(info.StubFiles, p)
					}
				}
			}
		}
		sortPaths(info.DirectFiles)
		sortPaths(info.StubFiles)
		sortPaths(info.Subfolders)
		tree.Folders[folder] = info
		for _, child := range info.Subfolders {
			if err := walk(child); err != nil {
				return err
			}
		}
		if !stub {
			stubDir := filepath.Join(folder, c.Draft.Folder)
			if st, err := os.Stat(stubDir); err == nil && st.IsDir() {
				if err := walk(stubDir); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := walk(tree.Root); err != nil {
		return model.DocsTree{}, err
	}
	return tree, nil
}

func sortPaths(paths []string) {
	sort.Slice(paths, func(i, j int) bool {
		left, right := paths[i], paths[j]
		if runtime.GOOS == "windows" {
			left, right = strings.ToLower(left), strings.ToLower(right)
		}
		return left < right
	})
}

func IsIndexable(root, path string, c config.Config) (bool, error) {
	if filepath.Base(path) == c.IndexFile {
		return false, nil
	}
	rel, err := pathutil.Relative(path, root)
	if err != nil {
		return false, err
	}
	for _, p := range c.Files.ExcludePatterns {
		if matches(rel, p) {
			return false, nil
		}
	}
	for _, p := range c.Files.IncludePatterns {
		if matches(rel, p) {
			return true, nil
		}
	}
	return false, nil
}
func matches(path, pattern string) bool {
	if glob(pattern, path) {
		return true
	}
	return strings.HasPrefix(pattern, "**/") && glob(strings.TrimPrefix(pattern, "**/"), path)
}
func glob(pattern, value string) bool {
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '*':
			b.WriteString(".*")
		case '?':
			b.WriteByte('.')
		case '[':
			j := strings.IndexByte(pattern[i:], ']')
			if j > 0 {
				b.WriteString(pattern[i : i+j+1])
				i += j
			} else {
				b.WriteString(`\[`)
			}
		default:
			b.WriteString(regexp.QuoteMeta(string(pattern[i])))
		}
	}
	b.WriteString("$")
	return regexp.MustCompile(b.String()).MatchString(value)
}
