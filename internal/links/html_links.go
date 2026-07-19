package links

import "strings"

func parseHTMLLinks(source string, protected []byteRange) []occurrence {
	var result []occurrence
	for i := 0; i < len(source); {
		if end, ok := rangeEnd(i, protected); ok {
			i = end
			continue
		}
		if source[i] != '<' {
			i++
			continue
		}
		end := htmlTagEnd(source, i)
		if end < 0 {
			i++
			continue
		}
		result = append(result, htmlTagLinks(source, i, end)...)
		i = end + 1
	}
	return result
}

func htmlTagEnd(source string, start int) int {
	quote := byte(0)
	for i := start + 1; i < len(source); i++ {
		if quote != 0 {
			if source[i] == quote {
				quote = 0
			}
			continue
		}
		if source[i] == '\'' || source[i] == '"' {
			quote = source[i]
			continue
		}
		if source[i] == '>' {
			return i
		}
	}
	return -1
}

func htmlTagLinks(source string, start, end int) []occurrence {
	i := start + 1
	for i < end && isHTMLSpace(source[i]) {
		i++
	}
	if i >= end || source[i] == '/' || source[i] == '!' || source[i] == '?' {
		return nil
	}
	tagStart := i
	for i < end && isHTMLName(source[i]) {
		i++
	}
	tag := strings.ToLower(source[tagStart:i])
	if tag == "" {
		return nil
	}
	var result []occurrence
	for i < end {
		for i < end && isHTMLSpace(source[i]) {
			i++
		}
		if i >= end || source[i] == '/' {
			break
		}
		nameStart := i
		for i < end && isHTMLName(source[i]) {
			i++
		}
		if nameStart == i {
			i++
			continue
		}
		name := strings.ToLower(source[nameStart:i])
		for i < end && isHTMLSpace(source[i]) {
			i++
		}
		if i >= end || source[i] != '=' {
			continue
		}
		i++
		for i < end && isHTMLSpace(source[i]) {
			i++
		}
		if i >= end {
			break
		}
		valueStart, valueEnd := htmlAttributeValue(source, i, end)
		i = valueEnd
		if i < end && (source[i] == '\'' || source[i] == '"') {
			i++
		}
		if valueEnd <= valueStart || !isHTMLTargetAttribute(tag, name) {
			continue
		}
		rawPath, suffix := splitSuffix(source[valueStart:valueEnd])
		line, column := sourcePosition(source, valueStart)
		result = append(result, occurrence{
			Start: valueStart, End: valueStart + len(rawPath), Line: line, Column: column,
			RawPath: rawPath, Suffix: suffix, Syntax: "html",
		})
	}
	return result
}

func htmlAttributeValue(source string, start, end int) (int, int) {
	if source[start] == '\'' || source[start] == '"' {
		quote := source[start]
		start++
		i := start
		for i < end && source[i] != quote {
			i++
		}
		return start, i
	}
	i := start
	for i < end && !isHTMLSpace(source[i]) && source[i] != '>' {
		i++
	}
	return start, i
}

func isHTMLTargetAttribute(tag, name string) bool {
	switch name {
	case "href":
		return tag == "a" || tag == "link"
	case "src":
		switch tag {
		case "img", "script", "source", "video", "audio", "iframe":
			return true
		}
	case "poster":
		return tag == "video"
	}
	return false
}

func isHTMLSpace(value byte) bool {
	return value == ' ' || value == '\t' || value == '\n' || value == '\r'
}

func isHTMLName(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' ||
		value >= '0' && value <= '9' || value == '-' || value == ':' || value == '_'
}
