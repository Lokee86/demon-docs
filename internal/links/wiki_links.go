package links

import "strings"

func parseWikiLinks(source string, protected []byteRange) []occurrence {
	var result []occurrence
	for i := 0; i+1 < len(source); {
		if end, ok := rangeEnd(i, protected); ok {
			i = end
			continue
		}
		if source[i] != '[' || source[i+1] != '[' || escaped(source, i) {
			i++
			continue
		}
		close := strings.Index(source[i+2:], "]]")
		if close < 0 {
			i += 2
			continue
		}
		contentStart := i + 2
		contentEnd := contentStart + close
		if strings.ContainsAny(source[contentStart:contentEnd], "\r\n") {
			i += 2
			continue
		}
		targetEnd := contentEnd
		if pipe := strings.IndexByte(source[contentStart:contentEnd], '|'); pipe >= 0 {
			targetEnd = contentStart + pipe
		}
		for contentStart < targetEnd && isHTMLSpace(source[contentStart]) {
			contentStart++
		}
		for targetEnd > contentStart && isHTMLSpace(source[targetEnd-1]) {
			targetEnd--
		}
		if targetEnd > contentStart {
			rawPath, suffix := splitSuffix(source[contentStart:targetEnd])
			line, column := sourcePosition(source, contentStart)
			result = append(result, occurrence{
				Start: contentStart, End: contentStart + len(rawPath), Line: line, Column: column,
				RawPath: rawPath, Suffix: suffix, Syntax: "wiki",
			})
		}
		i = contentEnd + 2
	}
	return result
}
