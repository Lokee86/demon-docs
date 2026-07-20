package documentpolicy

import (
	"regexp"
	"strings"
)

type listKind string

const (
	listUnordered listKind = "unordered"
	listOrdered   listKind = "ordered"
	listTask      listKind = "task"
)

var listMarker = regexp.MustCompile(`^(\s*)([-+*]|\d+[.)])\s+(.*)$`)
var taskMarker = regexp.MustCompile(`^\[[ xX]\]\s+`)

func mergeNodes(first, second *markdownSection, newline string) {
	firstKind, firstItems, firstOK := parseWholeList(first)
	secondKind, secondItems, secondOK := parseWholeList(second)
	if firstOK && secondOK && firstKind == secondKind {
		seen := map[string]bool{}
		merged := make([]string, 0, len(firstItems)+len(secondItems))
		for _, item := range append(firstItems, secondItems...) {
			key := normalizeListItem(item, firstKind)
			if seen[key] {
				continue
			}
			seen[key] = true
			merged = append(merged, strings.Trim(item, "\r\n"))
		}
		first.Lead = newline + strings.Join(merged, newline) + newline + newline
		return
	}
	content := renderSectionContent(second)
	if strings.TrimSpace(content) == "" {
		return
	}
	if first.Tail == "" && !strings.HasSuffix(first.Lead, newline+newline) {
		first.Tail = newline
	}
	if first.Tail != "" && !strings.HasSuffix(first.Tail, newline+newline) {
		first.Tail += newline
	}
	first.Tail += strings.TrimLeft(content, "\r\n")
}

func renderSectionContent(section *markdownSection) string {
	var out strings.Builder
	out.WriteString(section.Lead)
	for _, child := range section.Children {
		renderSection(&out, child)
	}
	out.WriteString(section.Tail)
	return out.String()
}

func parseWholeList(section *markdownSection) (listKind, []string, bool) {
	if len(section.Children) > 0 || strings.TrimSpace(section.Tail) != "" {
		return "", nil, false
	}
	lines := strings.Split(strings.ReplaceAll(section.Lead, "\r\n", "\n"), "\n")
	var items []string
	var current []string
	kind := listKind("")
	baseIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			if len(current) > 0 {
				current = append(current, line)
			}
			continue
		}
		match := listMarker.FindStringSubmatch(line)
		if match != nil {
			indent := len(strings.ReplaceAll(match[1], "\t", "    "))
			itemKind := markerKind(match[2], match[3])
			if baseIndent < 0 {
				baseIndent = indent
				kind = itemKind
			}
			if indent == baseIndent {
				if itemKind != kind {
					return "", nil, false
				}
				if len(current) > 0 {
					items = append(items, strings.TrimRight(strings.Join(current, "\n"), "\n"))
				}
				current = []string{line}
				continue
			}
			if indent < baseIndent {
				return "", nil, false
			}
		}
		if len(current) == 0 {
			return "", nil, false
		}
		current = append(current, line)
	}
	if len(current) > 0 {
		items = append(items, strings.TrimRight(strings.Join(current, "\n"), "\n"))
	}
	return kind, items, kind != "" && len(items) > 0
}

func markerKind(marker, text string) listKind {
	if marker == "-" || marker == "+" || marker == "*" {
		if taskMarker.MatchString(text) {
			return listTask
		}
		return listUnordered
	}
	return listOrdered
}

func normalizeListItem(item string, kind listKind) string {
	item = strings.ReplaceAll(item, "\r\n", "\n")
	lines := strings.Split(strings.TrimSpace(item), "\n")
	for i, line := range lines {
		line = strings.TrimRight(line, " \t")
		if i == 0 {
			if match := listMarker.FindStringSubmatch(line); match != nil {
				marker := "-"
				if kind == listOrdered {
					marker = "1."
				}
				line = marker + " " + strings.TrimSpace(match[3])
			}
		}
		lines[i] = strings.Join(strings.Fields(line), " ")
	}
	return strings.Join(lines, "\n")
}
