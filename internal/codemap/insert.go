package codemap

import (
	"fmt"
	"strings"
)

// InsertTarget appends one authored bullet to the first configured codemap
// section. Offsets are byte positions in normalized LF text and describe the
// inserted repair for selective undo.
func InsertTarget(source string, headings []string, target string) (string, int, int, string, error) {
	format := DefaultFormat()
	if len(headings) > 0 {
		format.SectionHeadings = headings
	}
	for _, entry := range Extract("document.md", source, format).Entries {
		if entry.Target == normalizeTarget(target) {
			return "", 0, 0, "", fmt.Errorf("codemap target is already authored: %s", target)
		}
	}
	normalized := strings.ReplaceAll(strings.ReplaceAll(source, "\r\n", "\n"), "\r", "\n")
	accepted := normalizedHeadings(format)
	lines := strings.SplitAfter(normalized, "\n")
	offset := 0
	active := false
	level := 0
	insertAt := -1
	for _, line := range lines {
		lineText := strings.TrimSuffix(line, "\n")
		if match := headingPattern.FindStringSubmatch(lineText); match != nil {
			currentLevel := len(match[1])
			title := strings.ToLower(cleanHeading(match[2]))
			if active && currentLevel <= level {
				insertAt = offset
				break
			}
			if !active && accepted[title] {
				active = true
				level = currentLevel
				insertAt = offset + len(line)
			}
		}
		if active {
			insertAt = offset + len(line)
		}
		offset += len(line)
	}
	if !active || insertAt < 0 {
		return "", 0, 0, "", errorsNoCodemapSection()
	}
	prefix := normalized[:insertAt]
	suffix := normalized[insertAt:]
	insertion := ""
	if prefix != "" && !strings.HasSuffix(prefix, "\n") {
		insertion += "\n"
	}
	if prefix != "" && !strings.HasSuffix(prefix, "\n\n") {
		insertion += "\n"
	}
	insertion += "- `" + target + "`\n"
	if suffix != "" && !strings.HasPrefix(suffix, "\n") {
		insertion += "\n"
	}
	return prefix + insertion + suffix, insertAt, insertAt, insertion, nil
}

func errorsNoCodemapSection() error {
	return fmt.Errorf("document has no configured codemap section")
}
