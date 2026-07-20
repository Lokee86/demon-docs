package codemap

import (
	"fmt"
	"strings"
)

type sectionSpan struct {
	heading            string
	level              int
	headingStart       int
	bodyStart, bodyEnd int
}

// HasSection reports whether the document contains exactly one configured
// codemap section. Multiple matching sections are an error.
func HasSection(source string, format Format) (bool, error) {
	_, found, err := locateSection(source, format)
	return found, err
}

func locateSection(source string, format Format) (sectionSpan, bool, error) {
	accepted := normalizedHeadings(format)
	lines := strings.SplitAfter(source, "\n")
	offset := 0
	fenceChar := byte(0)
	fenceSize := 0
	matches := []sectionSpan{}
	for _, line := range lines {
		text := strings.TrimSuffix(line, "\n")
		if marker := fencePattern.FindStringSubmatch(text); marker != nil {
			char, size := marker[1][0], len(marker[1])
			if fenceChar == 0 {
				fenceChar, fenceSize = char, size
			} else if char == fenceChar && size >= fenceSize {
				fenceChar, fenceSize = 0, 0
			}
			offset += len(line)
			continue
		}
		if fenceChar == 0 {
			if match := headingPattern.FindStringSubmatch(text); match != nil {
				level := len(match[1])
				title := cleanHeading(match[2])
				if len(matches) > 0 && matches[len(matches)-1].bodyEnd == 0 && level <= matches[len(matches)-1].level {
					matches[len(matches)-1].bodyEnd = offset
				}
				if accepted[strings.ToLower(title)] {
					matches = append(matches, sectionSpan{heading: title, level: level, headingStart: offset, bodyStart: offset + len(line)})
				}
			}
		}
		offset += len(line)
	}
	for index := range matches {
		if matches[index].bodyEnd == 0 {
			matches[index].bodyEnd = len(source)
		}
	}
	if len(matches) == 0 {
		return sectionSpan{}, false, nil
	}
	if len(matches) > 1 {
		return sectionSpan{}, false, fmt.Errorf("document contains multiple configured codemap sections")
	}
	return matches[0], true, nil
}

func insertSchemaSection(source string, format Format, placement SectionPlacement) (string, Format, error) {
	placement.Heading = cleanHeading(placement.Heading)
	if placement.Heading == "" {
		return "", format, fmt.Errorf("schema codemap heading is empty")
	}
	if placement.Level < 1 || placement.Level > 6 {
		return "", format, fmt.Errorf("schema codemap heading level must be between 1 and 6")
	}
	if placement.Offset < 0 || placement.Offset > len(source) {
		return "", format, fmt.Errorf("schema codemap insertion offset is outside the document")
	}
	prefix, suffix := source[:placement.Offset], source[placement.Offset:]
	before := ""
	if prefix != "" && !strings.HasSuffix(prefix, "\n\n") {
		before = "\n\n"
	}
	after := ""
	if suffix != "" && !strings.HasPrefix(suffix, "\n\n") {
		after = "\n\n"
	}
	heading := strings.Repeat("#", placement.Level) + " " + placement.Heading
	format.SectionHeadings = append(append([]string(nil), format.SectionHeadings...), placement.Heading)
	return prefix + before + heading + after + suffix, format, nil
}
