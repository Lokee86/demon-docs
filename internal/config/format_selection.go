package config

import (
	"fmt"
	pathpkg "path"
	"path/filepath"
	"strings"
)

// SelectFormatSchema resolves the effective shared document schema from explicit
// document_type metadata, then configured path rules, then the configured default.
func SelectFormatSchema(relative string, values map[string]any, cfg Format) (string, error) {
	if raw, ok := values["document_type"]; ok {
		name, ok := raw.(string)
		if !ok || strings.TrimSpace(name) == "" {
			return "", fmt.Errorf("document_type metadata must be a non-empty string")
		}
		return strings.TrimSpace(name), nil
	}
	for _, rule := range cfg.PathRules {
		matched, err := MatchFormatPath(rule.Pattern, relative)
		if err != nil {
			return "", fmt.Errorf("invalid format path pattern %q: %w", rule.Pattern, err)
		}
		if matched {
			return rule.Schema, nil
		}
	}
	return cfg.DefaultSchema, nil
}

func ValidateFormatPathPattern(pattern string) error {
	pattern = strings.TrimPrefix(filepath.ToSlash(strings.TrimSpace(pattern)), "./")
	for _, segment := range strings.Split(pattern, "/") {
		if segment == "**" {
			continue
		}
		if _, err := pathpkg.Match(segment, ""); err != nil {
			return err
		}
	}
	return nil
}

func MatchFormatPath(pattern, name string) (bool, error) {
	pattern = strings.TrimPrefix(filepath.ToSlash(strings.TrimSpace(pattern)), "./")
	name = strings.TrimPrefix(filepath.ToSlash(name), "./")
	return matchFormatPathSegments(strings.Split(pattern, "/"), strings.Split(name, "/"))
}

func matchFormatPathSegments(pattern, name []string) (bool, error) {
	if len(pattern) == 0 {
		return len(name) == 0, nil
	}
	if pattern[0] == "**" {
		for len(pattern) > 1 && pattern[1] == "**" {
			pattern = pattern[1:]
		}
		if len(pattern) == 1 {
			return true, nil
		}
		for consumed := 0; consumed <= len(name); consumed++ {
			matched, err := matchFormatPathSegments(pattern[1:], name[consumed:])
			if err != nil || matched {
				return matched, err
			}
		}
		return false, nil
	}
	if len(name) == 0 {
		return false, nil
	}
	matched, err := pathpkg.Match(pattern[0], name[0])
	if err != nil || !matched {
		return matched, err
	}
	return matchFormatPathSegments(pattern[1:], name[1:])
}
