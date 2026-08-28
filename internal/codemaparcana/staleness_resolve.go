package codemaparcana

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Lokee86/demon-docs/internal/codemap"
)

func resolveHistoricalEntry(ctx context.Context, client queryClient, entry codemap.DatasetEntry) (*protocolNode, error) {
	path, name, qualified := authoredSemanticHint(entry)
	request := map[string]any{"limit": 10000}
	switch {
	case name != "":
		request["op"] = "resolve_symbol"
		request["name"] = name
		if path != "" {
			request["path"] = path
		}
	case path != "":
		request["op"] = "resolve_file"
		request["path"] = path
	default:
		return nil, nil
	}
	var result nodeList
	if err := client.query(ctx, request, &result); err != nil {
		return nil, err
	}
	if result.Truncated || result.Returned != len(result.Nodes) {
		return nil, fmt.Errorf("historical target resolution was truncated")
	}
	var match *protocolNode
	for _, node := range result.Nodes {
		if path != "" && normalizeRepositoryPath(node.Path) != path {
			continue
		}
		if name != "" && node.Name != name {
			continue
		}
		if qualified != "" && node.QualifiedName != qualified {
			continue
		}
		if match != nil {
			return nil, nil
		}
		copy := node
		match = &copy
	}
	return match, nil
}

func authoredSemanticHint(entry codemap.DatasetEntry) (path, name, qualified string) {
	target := strings.TrimSpace(entry.Entry.Target)
	if strings.HasPrefix(target, "symbol:") {
		value := strings.TrimSpace(strings.TrimPrefix(target, "symbol:"))
		if index := strings.LastIndex(value, "::"); index >= 0 {
			return "", strings.TrimSpace(value[index+2:]), value
		}
		return "", value, ""
	}
	if index := strings.Index(target, "#"); index > 0 {
		return normalizeRepositoryPath(target[:index]), strings.TrimSpace(target[index+1:]), ""
	}
	if index := strings.LastIndex(target, "::"); index > 0 {
		left := target[:index]
		if strings.Contains(left, "/") || filepath.Ext(left) != "" {
			return normalizeRepositoryPath(left), strings.TrimSpace(target[index+2:]), ""
		}
		return "", strings.TrimSpace(target[index+2:]), target
	}
	if entry.Resolution.ResolvedPath != "" {
		return normalizeRepositoryPath(entry.Resolution.ResolvedPath), "", ""
	}
	return normalizeRepositoryPath(target), "", ""
}

func qualifiedTail(value string) string {
	if index := strings.LastIndex(value, "::"); index >= 0 {
		return value[index+2:]
	}
	return value
}
