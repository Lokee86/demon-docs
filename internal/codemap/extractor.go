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

type authoredLine struct {
	text string
	line int
}

// Extract parses authored code-map entries without resolving them against the
// repository. It accepts repository-configured heading aliases and preserves
// the source syntax instead of requiring documents to be normalized first.
func Extract(documentPath, source string, format Format) Result {
	documentPath = normalizePath(documentPath)
	headings := normalizedHeadings(format)
	source = strings.ReplaceAll(source, "\r\n", "\n")
	source = strings.ReplaceAll(source, "\r", "\n")
	lines := strings.Split(source, "\n")

	var result Result
	active := false
	mapLevel := 0
	heading := ""
	context := ""
	pendingLegacy := -1

	fenceChar := byte(0)
	fenceSize := 0
	fenceIsMap := false
	fenceContext := ""
	fenceHeading := ""
	var fenceLines []authoredLine

	flushFence := func() {
		if fenceIsMap {
			parseFenced(documentPath, fenceHeading, fenceContext, fenceLines, &result)
		}
		fenceChar, fenceSize = 0, 0
		fenceIsMap = false
		fenceContext, fenceHeading = "", ""
		fenceLines = nil
	}

	for index, line := range lines {
		lineNumber := index + 1
		if marker := fencePattern.FindStringSubmatch(line); marker != nil {
			char := marker[1][0]
			size := len(marker[1])
			if fenceChar == 0 {
				fenceChar, fenceSize = char, size
				fenceIsMap = active
				fenceContext, fenceHeading = context, heading
				fenceLines = nil
			} else if char == fenceChar && size >= fenceSize {
				flushFence()
			} else if fenceIsMap {
				fenceLines = append(fenceLines, authoredLine{text: line, line: lineNumber})
			}
			continue
		}
		if fenceChar != 0 {
			if fenceIsMap {
				fenceLines = append(fenceLines, authoredLine{text: line, line: lineNumber})
			}
			continue
		}

		if match := headingPattern.FindStringSubmatch(line); match != nil {
			level := len(match[1])
			title := cleanHeading(match[2])
			pendingLegacy = -1
			if active {
				if level <= mapLevel {
					active, heading, context = false, "", ""
				} else {
					context = title
				}
			} else if headings[strings.ToLower(title)] {
				active, mapLevel, heading, context = true, level, title, ""
				result.SectionCount++
			}
			continue
		}
		if !active {
			continue
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if isGroupLabel(trimmed) {
			context = strings.TrimSpace(strings.TrimSuffix(trimmed, ":"))
			pendingLegacy = -1
			continue
		}

		if list := listPattern.FindStringSubmatch(line); list != nil {
			pendingLegacy = -1
			prefix, body := list[1], list[2]
			match := codePattern.FindStringSubmatchIndex(body)
			if match == nil || !looksLikePrimaryTarget(body[match[2]:match[3]]) {
				continue
			}
			rawTarget := body[match[2]:match[3]]
			target := normalizeTarget(rawTarget)
			result.Entries = append(result.Entries, Entry{
				DocumentPath: documentPath,
				Heading:      heading,
				Target:       target,
				Kind:         classifyTarget(target),
				Syntax:       SyntaxBullet,
				Context:      context,
				Description:  trailingDescription(body, match[1]),
				Source:       lineSpan(lineNumber, len(prefix)+match[2]+1, len(prefix)+match[3]),
				RawLine:      line,
			})
			continue
		}

		if match := codePattern.FindStringSubmatchIndex(line); match != nil &&
			strings.TrimSpace(line[:match[0]]) == "" && len(line[:match[0]]) >= 2 &&
			looksLikePrimaryTarget(line[match[2]:match[3]]) {
			rawTarget := line[match[2]:match[3]]
			target := normalizeTarget(rawTarget)
			description := trailingDescription(line, match[1])
			syntax := SyntaxLegacyIndented
			if description != "" {
				syntax = SyntaxLegacyInline
			}
			result.Entries = append(result.Entries, Entry{
				DocumentPath: documentPath,
				Heading:      heading,
				Target:       target,
				Kind:         classifyTarget(target),
				Syntax:       syntax,
				Context:      context,
				Description:  description,
				Source:       lineSpan(lineNumber, match[2]+1, match[3]),
				RawLine:      line,
			})
			pendingLegacy = len(result.Entries) - 1
			continue
		}

		if pendingLegacy >= 0 && leadingIndent(line) >= 4 {
			appendDescription(&result.Entries[pendingLegacy], trimmed)
			continue
		}
		pendingLegacy = -1
	}
	if fenceChar != 0 {
		flushFence()
	}
	return result
}

func parseFenced(documentPath, heading, context string, lines []authoredLine, result *Result) {
	previous := -1
	for _, authored := range lines {
		trimmed := strings.TrimSpace(authored.text)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "->") || strings.HasPrefix(trimmed, "=") {
			if previous >= 0 {
				prefix := "->"
				syntax := SyntaxFencedArrow
				if strings.HasPrefix(trimmed, "=") {
					prefix = "="
					syntax = SyntaxFencedEquals
				}
				appendDescription(&result.Entries[previous], strings.TrimSpace(strings.TrimPrefix(trimmed, prefix)))
				result.Entries[previous].Syntax = syntax
			}
			continue
		}
		if previous >= 0 && leadingIndent(authored.text) > 0 && !looksLikeFencedTarget(trimmed) {
			appendDescription(&result.Entries[previous], trimmed)
			result.Entries[previous].Syntax = SyntaxFencedIndented
			continue
		}

		rawTarget, description, ok := splitFencedTarget(trimmed)
		if !ok {
			previous = -1
			continue
		}
		target := normalizeTarget(rawTarget)
		syntax := SyntaxFenced
		if description != "" {
			syntax = SyntaxFencedLeadingPath
		}
		column := strings.Index(authored.text, rawTarget) + 1
		result.Entries = append(result.Entries, Entry{
			DocumentPath: documentPath,
			Heading:      heading,
			Target:       target,
			Kind:         classifyTarget(target),
			Syntax:       syntax,
			Context:      context,
			Description:  description,
			Source:       lineSpan(authored.line, column, column+len(rawTarget)-1),
			RawLine:      authored.text,
		})
		previous = len(result.Entries) - 1
	}
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

func isGroupLabel(value string) bool {
	if !strings.HasSuffix(value, ":") || strings.Contains(value, "`") {
		return false
	}
	return !strings.HasPrefix(value, "-") && !strings.HasPrefix(value, "*") && !strings.HasPrefix(value, "+")
}

func isPlaceholder(value string) bool {
	return strings.Contains(strings.ToLower(value), "todo")
}

func looksLikePrimaryTarget(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || strings.Contains(value, "://") {
		return false
	}
	for _, char := range value {
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' {
			return true
		}
	}
	return false
}

func looksLikeFencedTarget(value string) bool {
	if !looksLikePrimaryTarget(value) {
		return false
	}
	if !strings.ContainsAny(value, " \t") {
		return true
	}
	first, _, ok := splitFirstField(value)
	return ok && looksLikeStructuredTarget(first)
}

func splitFencedTarget(value string) (string, string, bool) {
	value = strings.TrimSpace(value)
	if !looksLikePrimaryTarget(value) {
		return "", "", false
	}
	first, rest, split := splitFirstField(value)
	if !split {
		return value, "", true
	}
	if !looksLikeStructuredTarget(first) {
		return "", "", false
	}
	return first, strings.TrimSpace(rest), true
}

func splitFirstField(value string) (string, string, bool) {
	index := strings.IndexAny(value, " \t")
	if index < 0 {
		return value, "", false
	}
	return value[:index], value[index:], true
}

func looksLikeStructuredTarget(value string) bool {
	return strings.ContainsAny(value, `/\\*?[`) || strings.HasPrefix(value, "symbol:") ||
		strings.Contains(value, "::") || strings.Contains(value, "#") || path.Ext(value) != ""
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
	if hasPattern(target) {
		return TargetGlob
	}
	if strings.HasPrefix(target, "symbol:") || strings.Contains(target, "::") || strings.Contains(target, "#") ||
		strings.ContainsAny(target, " \t()") {
		return TargetSymbol
	}
	if strings.HasSuffix(target, "/") {
		return TargetDirectory
	}
	extension := path.Ext(target)
	if extension != "" {
		if isLikelySymbolExtension(extension) || !strings.Contains(target, "/") && !isKnownFileExtension(extension) {
			return TargetSymbol
		}
		return TargetFile
	}
	if !strings.Contains(target, "/") {
		return TargetSymbol
	}
	return TargetUnknown
}

func isLikelySymbolExtension(extension string) bool {
	trimmed := strings.TrimPrefix(extension, ".")
	return trimmed != "" && trimmed[0] >= 'A' && trimmed[0] <= 'Z'
}

func isKnownFileExtension(extension string) bool {
	switch strings.ToLower(extension) {
	case ".go", ".gd", ".md", ".toml", ".yaml", ".yml", ".json", ".ts", ".tsx", ".js", ".jsx", ".py", ".rb", ".html", ".css", ".scss", ".sql", ".proto", ".txt", ".sh", ".ps1", ".cs", ".rs", ".java", ".kt", ".xml", ".svg", ".png", ".jpg", ".jpeg", ".webp", ".tscn", ".tres", ".cfg", ".ini", ".lock", ".mod", ".sum", ".astro":
		return true
	default:
		return false
	}
}

func trailingDescription(body string, end int) string {
	description := strings.TrimSpace(body[end:])
	description = strings.TrimLeft(description, " \t:;-–—")
	return strings.TrimSpace(description)
}

func appendDescription(entry *Entry, description string) {
	description = strings.TrimSpace(description)
	if description == "" {
		return
	}
	if entry.Description == "" {
		entry.Description = description
		return
	}
	entry.Description += "\n" + description
}

func leadingIndent(value string) int {
	return len(value) - len(strings.TrimLeft(value, " \t"))
}

func lineSpan(line, column, endColumn int) SourceSpan {
	if column < 1 {
		column = 1
	}
	if endColumn < column {
		endColumn = column
	}
	return SourceSpan{Line: line, Column: column, EndLine: line, EndColumn: endColumn}
}
