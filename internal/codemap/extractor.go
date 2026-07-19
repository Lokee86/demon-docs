package codemap

import (
	"path"
	"regexp"
	"strings"
)

var (
	headingPattern = regexp.MustCompile(`^\s{0,3}(#{1,6})[ \t]+(.+?)[ \t]*#*[ \t]*$`)
	listPattern    = regexp.MustCompile(`^(\s*[-*+][ \t]+)(.*)$`)
	codePattern    = regexp.MustCompile("`([^`]+)`")
	fencePattern   = regexp.MustCompile(`^\s{0,3}(` + "`{3,}|~{3,}" + `)`)
)

// Extract parses authored code-map entries without resolving them against the
// repository. It preserves unsupported list entries as diagnostics rather than
// guessing what they reference.
func Extract(documentPath, source string, format Format) Result {
	documentPath = normalizePath(documentPath)
	headings := normalizedHeadings(format)
	lines := strings.Split(strings.ReplaceAll(source, "\r\n", "\n"), "\n")

	var result Result
	active := false
	mapLevel := 0
	context := ""
	fenceChar := byte(0)
	fenceSize := 0

	for index, line := range lines {
		if marker := fencePattern.FindStringSubmatch(line); marker != nil {
			char := marker[1][0]
			size := len(marker[1])
			if fenceChar == 0 {
				fenceChar, fenceSize = char, size
			} else if char == fenceChar && size >= fenceSize {
				fenceChar, fenceSize = 0, 0
			}
			continue
		}
		if fenceChar != 0 {
			continue
		}

		if match := headingPattern.FindStringSubmatch(line); match != nil {
			level := len(match[1])
			title := cleanHeading(match[2])
			if active {
				if level <= mapLevel {
					active, context = false, ""
				} else {
					context = title
				}
			} else if headings[strings.ToLower(title)] {
				active, mapLevel, context = true, level, ""
			}
			continue
		}
		if !active {
			continue
		}

		list := listPattern.FindStringSubmatch(line)
		if list == nil {
			continue
		}
		prefix, body := list[1], list[2]
		matches := codePattern.FindAllStringSubmatchIndex(body, -1)
		accepted := make([][]int, 0, len(matches))
		for matchIndex, match := range matches {
			value := body[match[2]:match[3]]
			if matchIndex == 0 && looksLikePrimaryTarget(value) || matchIndex > 0 && looksLikeTarget(value) {
				accepted = append(accepted, match)
			}
		}
		if len(accepted) == 0 {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Code:         "unparsed_entry",
				DocumentPath: documentPath,
				Message:      "code-map list entry contains no recognizable path or symbol target",
				Source:       lineSpan(index+1, len(prefix)+1, len(line)),
				RawLine:      line,
			})
			continue
		}

		description := trailingDescription(body, accepted[len(accepted)-1][1])
		for _, match := range accepted {
			rawTarget := body[match[2]:match[3]]
			target := normalizeTarget(rawTarget)
			result.Entries = append(result.Entries, Entry{
				DocumentPath: documentPath,
				Target:       target,
				Kind:         classifyTarget(target),
				Context:      context,
				Description:  description,
				Source:       lineSpan(index+1, len(prefix)+match[2]+1, len(prefix)+match[3]),
				RawLine:      line,
			})
		}
	}
	return result
}

func ExtractDefault(documentPath, source string) Result {
	return Extract(documentPath, source, DefaultFormat())
}

func normalizedHeadings(format Format) map[string]bool {
	if len(format.SectionHeadings) == 0 {
		format = DefaultFormat()
	}
	result := make(map[string]bool, len(format.SectionHeadings))
	for _, heading := range format.SectionHeadings {
		result[strings.ToLower(cleanHeading(heading))] = true
	}
	return result
}

func cleanHeading(value string) string {
	return strings.TrimSpace(strings.TrimRight(strings.TrimSpace(value), "#"))
}

func looksLikePrimaryTarget(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && !strings.Contains(value, "://")
}

func looksLikeTarget(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || strings.Contains(value, "://") {
		return false
	}
	return strings.ContainsAny(value, `/\\`) || strings.HasPrefix(value, "symbol:") ||
		strings.Contains(value, "::") || strings.Contains(value, "#")
}

func normalizeTarget(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, `\`, "/"))
	if strings.HasPrefix(value, "symbol:") {
		return value
	}
	trailingSlash := strings.HasSuffix(value, "/")
	value = path.Clean(value)
	if trailingSlash && value != "." {
		value += "/"
	}
	return value
}

func normalizePath(value string) string {
	return path.Clean(strings.ReplaceAll(value, `\`, "/"))
}

func classifyTarget(target string) TargetKind {
	if strings.HasPrefix(target, "symbol:") || strings.Contains(target, "::") || strings.Contains(target, "#") || strings.ContainsAny(target, " \t") {
		return TargetSymbol
	}
	if strings.HasSuffix(target, "/") {
		return TargetDirectory
	}
	if !strings.Contains(target, "/") && path.Ext(target) == "" {
		return TargetSymbol
	}
	if path.Ext(target) != "" {
		return TargetFile
	}
	return TargetUnknown
}

func trailingDescription(body string, end int) string {
	description := strings.TrimSpace(body[end:])
	description = strings.TrimLeft(description, " \t:;-–—")
	return strings.TrimSpace(description)
}

func lineSpan(line, column, endColumn int) SourceSpan {
	if endColumn < column {
		endColumn = column
	}
	return SourceSpan{Line: line, Column: column, EndLine: line, EndColumn: endColumn}
}
