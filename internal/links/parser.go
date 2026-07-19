package links

import (
	"sort"
	"strings"
)

type occurrence struct {
	Start, End   int
	Line, Column int
	RawPath      string
	Suffix       string
	Syntax       string
	Angle        bool
}

type byteRange struct{ Start, End int }

func parseMarkdownLinks(source string) []occurrence {
	protected := protectedMarkdownRanges(source)
	result := parseReferenceDefinitions(source, protected)
	for i := 0; i < len(source); {
		if end, ok := rangeEnd(i, protected); ok {
			i = end
			continue
		}
		if source[i] == '`' && !escaped(source, i) {
			run := repeated(source, i, '`')
			needle := strings.Repeat("`", run)
			if close := strings.Index(source[i+run:], needle); close >= 0 {
				i += run + close + run
				continue
			}
		}
		if source[i] == ']' && i+1 < len(source) && source[i+1] == '(' && !escaped(source, i) {
			start, end, angle, ok := inlineDestination(source, i+1)
			if ok {
				path, suffix := splitSuffix(source[start:end])
				line, column := sourcePosition(source, start)
				result = append(result, occurrence{Start: start, End: start + len(path), Line: line, Column: column, RawPath: path, Suffix: suffix, Syntax: "inline", Angle: angle})
				i = end
				continue
			}
		}
		i++
	}
	result = append(result, parseHTMLLinks(source, protected)...)
	result = append(result, parseWikiLinks(source, protected)...)
	sort.Slice(result, func(i, j int) bool { return result[i].Start < result[j].Start })
	return result
}

func inlineDestination(source string, open int) (int, int, bool, bool) {
	i := open + 1
	for i < len(source) && (source[i] == ' ' || source[i] == '\t' || source[i] == '\n' || source[i] == '\r') {
		i++
	}
	if i >= len(source) {
		return 0, 0, false, false
	}
	if source[i] == '<' {
		start := i + 1
		for j := start; j < len(source); j++ {
			if source[j] == '>' && !escaped(source, j) {
				if closingParenAfter(source, j+1) {
					return start, j, true, true
				}
				return 0, 0, false, false
			}
			if source[j] == '\n' || source[j] == '\r' {
				return 0, 0, false, false
			}
		}
		return 0, 0, false, false
	}
	start := i
	depth := 0
	for j := i; j < len(source); j++ {
		switch source[j] {
		case '\\':
			j++
		case '(':
			depth++
		case ')':
			if depth == 0 {
				return start, j, false, true
			}
			depth--
		case ' ', '\t', '\n', '\r':
			if depth == 0 && closingParenAfter(source, j) {
				return start, j, false, true
			}
		}
	}
	return 0, 0, false, false
}

func closingParenAfter(source string, start int) bool {
	quote := byte(0)
	for i := start; i < len(source); i++ {
		c := source[i]
		if quote != 0 {
			if c == quote && !escaped(source, i) {
				quote = 0
			}
			continue
		}
		if c == '\'' || c == '"' {
			quote = c
			continue
		}
		if c == ')' {
			return true
		}
		if c == '\n' || c == '\r' {
			return false
		}
		if c != ' ' && c != '\t' {
			return false
		}
	}
	return false
}

func parseReferenceDefinitions(source string, protected []byteRange) []occurrence {
	var result []occurrence
	for lineStart := 0; lineStart <= len(source); {
		lineEnd := strings.IndexByte(source[lineStart:], '\n')
		if lineEnd < 0 {
			lineEnd = len(source)
		} else {
			lineEnd += lineStart
		}
		if !inRanges(lineStart, protected) {
			line := source[lineStart:lineEnd]
			indent := 0
			for indent < len(line) && indent < 4 && line[indent] == ' ' {
				indent++
			}
			if indent <= 3 && indent < len(line) && line[indent] == '[' {
				marker := referenceMarker(line, indent)
				if marker >= 0 {
					i := marker
					for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
						i++
					}
					angle := i < len(line) && line[i] == '<'
					if angle {
						i++
					}
					start := i
					for i < len(line) {
						if angle && line[i] == '>' && !escaped(line, i) {
							break
						}
						if !angle && (line[i] == ' ' || line[i] == '\t') {
							break
						}
						i++
					}
					if i >= start && (!angle || i < len(line)) {
						absoluteStart := lineStart + start
						path, suffix := splitSuffix(line[start:i])
						lineNumber, column := sourcePosition(source, absoluteStart)
						result = append(result, occurrence{Start: absoluteStart, End: absoluteStart + len(path), Line: lineNumber, Column: column, RawPath: path, Suffix: suffix, Syntax: "reference", Angle: angle})
					}
				}
			}
		}
		if lineEnd == len(source) {
			break
		}
		lineStart = lineEnd + 1
	}
	return result
}

func referenceMarker(line string, start int) int {
	close := -1
	for i := start + 1; i < len(line); i++ {
		if line[i] == ']' && !escaped(line, i) {
			close = i
			break
		}
	}
	if close < 0 || close+1 >= len(line) || line[close+1] != ':' {
		return -1
	}
	return close + 2
}

func splitSuffix(target string) (string, string) {
	index := len(target)
	if hash := strings.IndexByte(target, '#'); hash >= 0 && hash < index {
		index = hash
	}
	if query := strings.IndexByte(target, '?'); query >= 0 && query < index {
		index = query
	}
	return target[:index], target[index:]
}

func fencedRanges(source string) []byteRange {
	var result []byteRange
	openStart, marker, markerCount := -1, byte(0), 0
	for lineStart := 0; lineStart <= len(source); {
		lineEnd := strings.IndexByte(source[lineStart:], '\n')
		if lineEnd < 0 {
			lineEnd = len(source)
		} else {
			lineEnd += lineStart
		}
		line := strings.TrimSuffix(source[lineStart:lineEnd], "\r")
		trimmed := strings.TrimLeft(line, " ")
		indent := len(line) - len(trimmed)
		if indent <= 3 && len(trimmed) >= 3 && (trimmed[0] == '`' || trimmed[0] == '~') {
			count := repeated(trimmed, 0, trimmed[0])
			if count >= 3 {
				if openStart < 0 {
					openStart, marker, markerCount = lineStart, trimmed[0], count
				} else if trimmed[0] == marker && count >= markerCount && strings.TrimSpace(trimmed[count:]) == "" {
					end := lineEnd
					if end < len(source) {
						end++
					}
					result = append(result, byteRange{Start: openStart, End: end})
					openStart = -1
				}
			}
		}
		if lineEnd == len(source) {
			break
		}
		lineStart = lineEnd + 1
	}
	if openStart >= 0 {
		result = append(result, byteRange{Start: openStart, End: len(source)})
	}
	return result
}

func repeated(source string, start int, value byte) int {
	count := 0
	for start+count < len(source) && source[start+count] == value {
		count++
	}
	return count
}

func escaped(source string, position int) bool {
	count := 0
	for position-count-1 >= 0 && source[position-count-1] == '\\' {
		count++
	}
	return count%2 == 1
}

func inRanges(position int, ranges []byteRange) bool {
	_, ok := rangeEnd(position, ranges)
	return ok
}

func rangeEnd(position int, ranges []byteRange) (int, bool) {
	for _, r := range ranges {
		if position >= r.Start && position < r.End {
			return r.End, true
		}
	}
	return 0, false
}

func sourcePosition(source string, offset int) (int, int) {
	line := 1 + strings.Count(source[:offset], "\n")
	last := strings.LastIndex(source[:offset], "\n")
	return line, offset - last
}
