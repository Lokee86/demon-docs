package links

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	markdownutil "github.com/Lokee86/demon-docs/internal/markdown"
)

type headingFragmentValidator struct {
	anchors map[string]map[string]struct{}
}

func newHeadingFragmentValidator() *headingFragmentValidator {
	return &headingFragmentValidator{anchors: map[string]map[string]struct{}{}}
}

func fragmentFromSuffix(suffix string) (string, bool) {
	index := strings.IndexByte(suffix, '#')
	if index < 0 || index+1 >= len(suffix) {
		return "", false
	}
	fragment := suffix[index+1:]
	decoded, err := url.PathUnescape(fragment)
	if err == nil {
		fragment = decoded
	}
	return fragment, fragment != ""
}

func (v *headingFragmentValidator) validate(targetPath, sourcePath, sourceText, suffix string) (bool, bool, error) {
	fragment, present := fragmentFromSuffix(suffix)
	if !present || !isMarkdown(targetPath) {
		return true, false, nil
	}
	path := filepath.Clean(targetPath)
	key := pathKey(path)
	anchors, ok := v.anchors[key]
	if !ok {
		text := sourceText
		if pathKey(sourcePath) != key {
			contents, err := os.ReadFile(path)
			if err != nil {
				return false, true, err
			}
			text = string(contents)
		}
		anchors = map[string]struct{}{}
		for _, anchor := range markdownutil.HeadingAnchors(text) {
			anchors[anchor] = struct{}{}
		}
		v.anchors[key] = anchors
	}
	_, exists := anchors[fragment]
	return exists, true, nil
}
