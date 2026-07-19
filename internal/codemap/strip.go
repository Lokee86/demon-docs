package codemap

import "strings"

// StripAuthoredSections removes configured code-map sections while preserving
// all non-map text and line positions. Benchmark evidence must not contain the
// authored links it is expected to rediscover.
func StripAuthoredSections(source string, format Format) string {
	headings := normalizedHeadings(format)
	source = strings.ReplaceAll(source, "\r\n", "\n")
	source = strings.ReplaceAll(source, "\r", "\n")
	lines := strings.Split(source, "\n")

	active := false
	mapLevel := 0
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
			if active {
				lines[index] = ""
			}
			continue
		}
		if fenceChar != 0 {
			if active {
				lines[index] = ""
			}
			continue
		}

		match := headingPattern.FindStringSubmatch(line)
		if match != nil {
			level := len(match[1])
			title := cleanHeading(match[2])
			if active && level <= mapLevel {
				active = false
			}
			if !active && headings[strings.ToLower(title)] {
				active, mapLevel = true, level
			}
		}
		if active {
			lines[index] = ""
		}
	}
	return strings.Join(lines, "\n")
}
