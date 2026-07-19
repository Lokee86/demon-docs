package links

import (
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"unicode"
)

type targetStyle struct {
	absolute   bool
	fileURL    bool
	backslash  bool
	dotSlash   bool
	urlEscaped bool
	angle      bool
}

func resolveLocalTarget(rawPath, sourcePath string, angle bool) (string, targetStyle, bool) {
	style := targetStyle{
		backslash:  strings.Contains(rawPath, "\\") && !strings.Contains(rawPath, "/"),
		dotSlash:   strings.HasPrefix(rawPath, "./") || strings.HasPrefix(rawPath, ".\\"),
		urlEscaped: strings.Contains(rawPath, "%"),
		angle:      angle,
	}
	if rawPath == "" {
		return filepath.Clean(sourcePath), style, true
	}
	decoded, err := url.PathUnescape(rawPath)
	if err != nil {
		decoded = rawPath
	}
	lower := strings.ToLower(decoded)
	if strings.HasPrefix(lower, "file://") {
		parsed, err := url.Parse(decoded)
		if err != nil {
			return "", style, false
		}
		path := parsed.Path
		if parsed.Host != "" {
			path = "//" + parsed.Host + path
		}
		if runtime.GOOS == "windows" && len(path) >= 3 && path[0] == '/' && isDrivePath(path[1:]) {
			path = path[1:]
		}
		style.absolute = true
		style.fileURL = true
		return filepath.Clean(filepath.FromSlash(path)), style, true
	}
	if hasScheme(decoded) && !isDrivePath(decoded) {
		return "", style, false
	}
	filesystemPath := filepath.FromSlash(decoded)
	style.absolute = filepath.IsAbs(filesystemPath) || isDrivePath(decoded) || strings.HasPrefix(decoded, `\\`)
	if style.absolute {
		return filepath.Clean(filesystemPath), style, true
	}
	return filepath.Clean(filepath.Join(filepath.Dir(sourcePath), filesystemPath)), style, true
}

func renderTargetPath(style targetStyle, originalPath, sourcePath, targetPath string) string {
	var rendered string
	if style.absolute {
		rendered = filepath.Clean(targetPath)
	} else {
		relative, err := filepath.Rel(filepath.Dir(sourcePath), targetPath)
		if err != nil {
			rendered = filepath.Clean(targetPath)
		} else {
			rendered = relative
			if style.dotSlash && rendered != "." && !strings.HasPrefix(rendered, "."+string(filepath.Separator)) && !strings.HasPrefix(rendered, ".."+string(filepath.Separator)) {
				rendered = "." + string(filepath.Separator) + rendered
			}
		}
	}
	if style.fileURL {
		slashed := filepath.ToSlash(rendered)
		if runtime.GOOS == "windows" && isDrivePath(slashed) {
			slashed = "/" + slashed
		}
		rendered = "file://" + slashed
	} else if style.backslash {
		rendered = strings.ReplaceAll(rendered, "/", "\\")
	} else {
		rendered = filepath.ToSlash(rendered)
	}
	if style.urlEscaped || (!style.angle && strings.ContainsAny(rendered, " #?")) {
		rendered = escapeFilesystemPath(rendered)
	}
	if originalPath == "." && filepath.Clean(targetPath) == filepath.Clean(sourcePath) {
		return "."
	}
	return rendered
}

func hasScheme(value string) bool {
	colon := strings.IndexByte(value, ':')
	if colon <= 0 {
		return false
	}
	for index, r := range value[:colon] {
		if index == 0 && !unicode.IsLetter(r) {
			return false
		}
		if index > 0 && !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '+' && r != '-' && r != '.' {
			return false
		}
	}
	return true
}

func isDrivePath(value string) bool {
	return len(value) >= 3 && unicode.IsLetter(rune(value[0])) && value[1] == ':' && (value[2] == '/' || value[2] == '\\')
}

func escapeFilesystemPath(value string) string {
	replacer := strings.NewReplacer("%", "%25", " ", "%20", "#", "%23", "?", "%3F")
	return replacer.Replace(value)
}

func discoverExternalCandidates(oldPath, base, kind, fingerprint string) []string {
	if oldPath == "" || base == "" {
		return nil
	}
	root := filepath.Dir(oldPath)
	for {
		info, err := os.Stat(root)
		if err == nil && info.IsDir() {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			return nil
		}
		root = parent
	}
	if filepath.Dir(root) == root {
		return nil
	}
	const maximumEntries = 10000
	entries := 0
	var named, exact []string
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entries >= maximumEntries {
			return filepath.SkipDir
		}
		entries++
		if entry.Type()&os.ModeSymlink != 0 {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if path == oldPath || !strings.EqualFold(filepath.Base(path), base) {
			return nil
		}
		info, err := entry.Info()
		if err != nil || kindFromInfo(info) != kind {
			return nil
		}
		clean := filepath.Clean(path)
		named = append(named, clean)
		if fingerprint != "" && info.Mode().IsRegular() {
			if current, err := fileFingerprint(clean); err == nil && current == fingerprint {
				exact = append(exact, clean)
			}
		}
		return nil
	})
	if len(exact) > 0 {
		named = exact
	}
	sort.Slice(named, func(i, j int) bool { return pathKey(named[i]) < pathKey(named[j]) })
	return uniquePaths(named)
}

func uniquePaths(paths []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		key := pathKey(path)
		if !seen[key] {
			seen[key] = true
			result = append(result, path)
		}
	}
	return result
}
