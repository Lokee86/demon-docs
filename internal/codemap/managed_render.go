package codemap

import (
	"fmt"
	"sort"
	"strings"
)

func replaceSectionBody(source string, span sectionSpan, startMarker, endMarker, content string) string {
	block := startMarker
	if content != "" {
		block += "\n\n" + content + "\n"
	} else {
		block += "\n"
	}
	block += endMarker

	leading := "\n"
	if span.bodyStart == 0 || source[span.bodyStart-1] != '\n' {
		leading = "\n\n"
	}
	replacement := leading + block
	if span.bodyEnd < len(source) {
		replacement += "\n\n"
	} else {
		replacement += "\n"
	}
	return source[:span.bodyStart] + replacement + source[span.bodyEnd:]
}

func appendTargetsInSection(content string, targets []string, entries []Entry) string {
	if closing, ok := firstCompleteFence(content); ok {
		insertion := strings.Join(targets, "\n")
		prefix := content[:closing]
		if prefix != "" && !strings.HasSuffix(prefix, "\n") {
			insertion = "\n" + insertion
		}
		if insertion != "" && !strings.HasSuffix(insertion, "\n") {
			insertion += "\n"
		}
		return prefix + insertion + content[closing:]
	}

	prefix := "- "
	for _, entry := range entries {
		if entry.Syntax != SyntaxBullet {
			continue
		}
		if match := listPattern.FindStringSubmatch(entry.RawLine); match != nil {
			prefix = match[1]
			break
		}
	}
	lines := make([]string, len(targets))
	for index, target := range targets {
		lines[index] = prefix + "`" + target + "`"
	}
	addition := strings.Join(lines, "\n")
	if content == "" {
		return addition
	}
	return strings.TrimRight(content, "\n") + "\n" + addition
}

func firstCompleteFence(content string) (closing int, ok bool) {
	offset := 0
	fenceChar := byte(0)
	fenceSize := 0
	for _, line := range strings.SplitAfter(content, "\n") {
		text := strings.TrimSuffix(line, "\n")
		marker := fencePattern.FindStringSubmatch(text)
		if marker == nil {
			offset += len(line)
			continue
		}
		char, size := marker[1][0], len(marker[1])
		if fenceChar == 0 {
			fenceChar, fenceSize = char, size
		} else if char == fenceChar && size >= fenceSize {
			return offset, true
		}
		offset += len(line)
	}
	return -1, false
}

func removeManagedMarkers(body, startMarker, endMarker string) (string, error) {
	startCount, endCount := 0, 0
	kept := []string{}
	for _, line := range strings.SplitAfter(body, "\n") {
		trimmed := strings.TrimSpace(strings.TrimSuffix(line, "\n"))
		switch trimmed {
		case startMarker:
			startCount++
		case endMarker:
			endCount++
		default:
			kept = append(kept, line)
		}
	}
	if startCount != endCount || startCount > 1 {
		return "", fmt.Errorf("codemap managed markers are malformed or duplicated")
	}
	return strings.Join(kept, ""), nil
}

func removeEntryLines(documentPath, source string, format Format, targets map[string]struct{}) (string, []string) {
	entries := Extract(documentPath, source, format).Entries
	lines := strings.SplitAfter(source, "\n")
	removeLines := map[int]string{}
	for _, entry := range entries {
		if _, remove := targets[normalizeTarget(entry.Target)]; remove {
			removeLines[entry.Source.Line-1] = normalizeTarget(entry.Target)
		}
	}
	removedSet := map[string]struct{}{}
	kept := make([]string, 0, len(lines))
	for index, line := range lines {
		if target, remove := removeLines[index]; remove {
			removedSet[target] = struct{}{}
			continue
		}
		kept = append(kept, line)
	}
	removed := make([]string, 0, len(removedSet))
	for target := range removedSet {
		removed = append(removed, target)
	}
	sort.Strings(removed)
	return strings.Join(kept, ""), removed
}

func normalizedTargetSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		if normalized := normalizeTarget(value); normalized != "" && normalized != "." {
			result[normalized] = struct{}{}
		}
	}
	return result
}
