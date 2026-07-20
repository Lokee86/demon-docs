package documentpolicy

import (
	"regexp"
	"strings"
)

type markdownDocument struct {
	Prefix  string
	Roots   []*markdownSection
	Newline string
}

type markdownSection struct {
	Heading     string
	Level       int
	HeadingText string
	Lead        string
	Children    []*markdownSection
	Tail        string
}

type headingRecord struct {
	start, end int
	level      int
	heading    string
	text       string
}

var atxHeading = regexp.MustCompile(`^( {0,3})(#{1,6})[ \t]+(.+?)(?:[ \t]+#+[ \t]*)?$`)
var setextHeading = regexp.MustCompile(`^ {0,3}(=+|-+)[ \t]*$`)

func parseMarkdown(body string) markdownDocument {
	newline := "\n"
	if strings.Contains(body, "\r\n") {
		newline = "\r\n"
	}
	records := scanHeadings(body)
	sections := make([]headingRecord, 0, len(records))
	prefixEnd := len(body)
	for _, record := range records {
		if record.level < 2 {
			continue
		}
		if len(sections) == 0 {
			prefixEnd = record.start
		}
		sections = append(sections, record)
	}
	if len(sections) == 0 {
		return markdownDocument{Prefix: body, Newline: newline}
	}
	doc := markdownDocument{Prefix: body[:prefixEnd], Newline: newline}
	var stack []*markdownSection
	for i, record := range sections {
		next := len(body)
		if i+1 < len(sections) {
			next = sections[i+1].start
		}
		node := &markdownSection{
			Heading:     record.heading,
			Level:       record.level,
			HeadingText: record.text,
			Lead:        body[record.end:next],
		}
		for len(stack) > 0 && stack[len(stack)-1].Level >= node.Level {
			stack = stack[:len(stack)-1]
		}
		if len(stack) == 0 {
			doc.Roots = append(doc.Roots, node)
		} else {
			parent := stack[len(stack)-1]
			parent.Children = append(parent.Children, node)
		}
		stack = append(stack, node)
	}
	return doc
}

func (doc markdownDocument) render() string {
	var out strings.Builder
	out.WriteString(doc.Prefix)
	for _, section := range doc.Roots {
		renderSection(&out, section)
	}
	return out.String()
}

func renderSection(out *strings.Builder, section *markdownSection) {
	out.WriteString(section.HeadingText)
	out.WriteString(section.Lead)
	for _, child := range section.Children {
		renderSection(out, child)
	}
	out.WriteString(section.Tail)
}

func scanHeadings(source string) []headingRecord {
	lines := splitLines(source)
	var records []headingRecord
	fenceChar := byte(0)
	fenceLength := 0
	htmlEnd := ""
	htmlUntilBlank := false
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		plain := trimEnding(line.text)
		trimmed := strings.TrimSpace(plain)
		if fenceChar != 0 {
			if closesFence(trimmed, fenceChar, fenceLength) {
				fenceChar = 0
				fenceLength = 0
			}
			continue
		}
		if htmlEnd != "" {
			if strings.Contains(strings.ToLower(trimmed), htmlEnd) {
				htmlEnd = ""
			}
			continue
		}
		if htmlUntilBlank {
			if trimmed == "" {
				htmlUntilBlank = false
			}
			continue
		}
		if char, length, ok := opensFence(trimmed); ok {
			fenceChar, fenceLength = char, length
			continue
		}
		if end, untilBlank, ok := opensHTML(trimmed); ok {
			htmlEnd, htmlUntilBlank = end, untilBlank
			if end != "" && strings.Contains(strings.ToLower(trimmed), end) {
				htmlEnd = ""
			}
			continue
		}
		if strings.HasPrefix(strings.TrimLeft(plain, " \t"), ">") {
			continue
		}
		if match := atxHeading.FindStringSubmatch(plain); match != nil {
			records = append(records, headingRecord{
				start:   line.start,
				end:     line.end,
				level:   len(match[2]),
				heading: strings.TrimSpace(match[3]),
				text:    line.text,
			})
			continue
		}
		if i == 0 || trimmed == "" {
			continue
		}
		match := setextHeading.FindStringSubmatch(plain)
		if match == nil {
			continue
		}
		previous := lines[i-1]
		previousPlain := trimEnding(previous.text)
		if strings.TrimSpace(previousPlain) == "" || strings.HasPrefix(strings.TrimLeft(previousPlain, " \t"), ">") {
			continue
		}
		level := 2
		if strings.HasPrefix(match[1], "=") {
			level = 1
		}
		records = append(records, headingRecord{
			start:   previous.start,
			end:     line.end,
			level:   level,
			heading: strings.TrimSpace(previousPlain),
			text:    previous.text + line.text,
		})
	}
	return records
}

type sourceLine struct {
	start, end int
	text       string
}

func splitLines(source string) []sourceLine {
	if source == "" {
		return nil
	}
	var lines []sourceLine
	start := 0
	for start < len(source) {
		relative := strings.IndexByte(source[start:], '\n')
		end := len(source)
		if relative >= 0 {
			end = start + relative + 1
		}
		lines = append(lines, sourceLine{start: start, end: end, text: source[start:end]})
		start = end
	}
	return lines
}

func trimEnding(line string) string {
	line = strings.TrimSuffix(line, "\n")
	return strings.TrimSuffix(line, "\r")
}

func opensFence(trimmed string) (byte, int, bool) {
	if len(trimmed) < 3 || trimmed[0] != '`' && trimmed[0] != '~' {
		return 0, 0, false
	}
	char := trimmed[0]
	length := 0
	for length < len(trimmed) && trimmed[length] == char {
		length++
	}
	return char, length, length >= 3
}

func closesFence(trimmed string, char byte, minimum int) bool {
	if len(trimmed) < minimum || trimmed[0] != char {
		return false
	}
	count := 0
	for count < len(trimmed) && trimmed[count] == char {
		count++
	}
	return count >= minimum && strings.TrimSpace(trimmed[count:]) == ""
}

func opensHTML(trimmed string) (string, bool, bool) {
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "<!--") {
		return "-->", false, true
	}
	if strings.HasPrefix(lower, "<?") {
		return "?>", false, true
	}
	if strings.HasPrefix(lower, "<![cdata[") {
		return "]]>", false, true
	}
	for _, tag := range []string{"script", "pre", "style", "textarea"} {
		if strings.HasPrefix(lower, "<"+tag) {
			return "</" + tag + ">", false, true
		}
	}
	if strings.HasPrefix(lower, "<!") || strings.HasPrefix(lower, "<") && len(lower) > 1 && lower[1] >= 'a' && lower[1] <= 'z' {
		return "", true, true
	}
	return "", false, false
}

func replaceHeading(section *markdownSection, heading string, newline string) {
	if section.Heading == heading {
		return
	}
	trimmed := trimEnding(section.HeadingText)
	if match := atxHeading.FindStringSubmatch(trimmed); match != nil {
		ending := newline
		if strings.HasSuffix(section.HeadingText, "\r\n") {
			ending = "\r\n"
		} else if strings.HasSuffix(section.HeadingText, "\n") {
			ending = "\n"
		} else {
			ending = ""
		}
		section.HeadingText = match[1] + match[2] + " " + heading + ending
	} else {
		underline := strings.Repeat("-", len([]rune(heading)))
		if section.Level == 1 {
			underline = strings.Repeat("=", len([]rune(heading)))
		}
		section.HeadingText = heading + newline + underline + newline
	}
	section.Heading = heading
}
