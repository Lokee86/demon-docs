package links

import "strings"

func parseUndefinedReferences(source string, protected []byteRange) []undefinedReference {
	definitions := referenceDefinitionLabels(source, protected)
	var result []undefinedReference
	for i := 0; i < len(source); {
		if end, ok := rangeEnd(i, protected); ok {
			i = end
			continue
		}
		if source[i] != '[' || escaped(source, i) || i+1 < len(source) && source[i+1] == '[' {
			i++
			continue
		}
		firstEnd := referenceBracketEnd(source, i+1)
		if firstEnd < 0 || firstEnd+1 >= len(source) || source[firstEnd+1] != '[' {
			i++
			continue
		}
		secondEnd := referenceBracketEnd(source, firstEnd+2)
		if secondEnd < 0 {
			i++
			continue
		}
		labelStart, labelEnd := firstEnd+2, secondEnd
		label := source[labelStart:labelEnd]
		if label == "" {
			labelStart, labelEnd = i+1, firstEnd
			label = source[labelStart:labelEnd]
		}
		normalized := normalizeReferenceLabel(label)
		if normalized != "" && !definitions[normalized] {
			line, column := sourcePosition(source, labelStart)
			result = append(result, undefinedReference{
				Start: labelStart, End: labelEnd, Line: line, Column: column, Label: label,
			})
		}
		i = secondEnd + 1
	}
	return result
}

func referenceDefinitionLabels(source string, protected []byteRange) map[string]bool {
	result := map[string]bool{}
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
				close := referenceBracketEnd(line, indent+1)
				if close >= 0 && close+1 < len(line) && line[close+1] == ':' {
					label := normalizeReferenceLabel(line[indent+1 : close])
					if label != "" {
						result[label] = true
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

func referenceBracketEnd(source string, start int) int {
	for i := start; i < len(source); i++ {
		if source[i] == '\n' || source[i] == '\r' {
			return -1
		}
		if source[i] == ']' && !escaped(source, i) {
			return i
		}
	}
	return -1
}

func normalizeReferenceLabel(label string) string {
	return strings.ToLower(strings.Join(strings.Fields(label), " "))
}
