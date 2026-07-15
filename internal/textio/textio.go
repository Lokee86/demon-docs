package textio

import (
	"os"
	"runtime"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
)

type Document struct {
	Text, Newline string
	raw           string
	mixed         bool
}

func Decode(data []byte) Document {
	raw := string(data)
	nl := "\n"
	hasCRLF := strings.Contains(raw, "\r\n")
	withoutCRLF := strings.ReplaceAll(raw, "\r\n", "")
	hasLF := strings.Contains(withoutCRLF, "\n")
	if hasCRLF {
		nl = "\r\n"
	}
	return Document{Text: strings.ReplaceAll(raw, "\r\n", "\n"), Newline: nl, raw: raw, mixed: hasCRLF && hasLF}
}

func Read(path string) (Document, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Document{}, err
	}
	return Decode(b), nil
}

func (d Document) Encode(text string) []byte {
	if d.mixed {
		return []byte(d.encodeMixed(text))
	}
	if d.Newline == "\r\n" {
		text = strings.ReplaceAll(text, "\n", "\r\n")
	}
	return []byte(text)
}

func EncodeNew(text string) []byte {
	if runtime.GOOS == "windows" {
		text = strings.ReplaceAll(text, "\n", "\r\n")
	}
	return []byte(text)
}

// encodeMixed keeps every unchanged source line, including its original line
// ending, and uses the document's first-observed line ending for inserted lines.
func (d Document) encodeMixed(text string) string {
	oldLines := normalizedLines(d.Text)
	newLines := normalizedLines(text)
	rawLines := rawLines(d.raw)
	if len(oldLines) != len(rawLines) {
		return strings.ReplaceAll(text, "\n", d.Newline)
	}
	blocks := difflib.NewMatcher(oldLines, newLines).GetMatchingBlocks()
	matchedNew := make(map[int]int)
	for _, block := range blocks {
		for offset := 0; offset < block.Size; offset++ {
			matchedNew[block.B+offset] = block.A + offset
		}
	}
	var out strings.Builder
	for i, line := range newLines {
		if old, ok := matchedNew[i]; ok {
			out.WriteString(rawLines[old])
			continue
		}
		out.WriteString(strings.TrimSuffix(line, "\n"))
		if strings.HasSuffix(line, "\n") {
			out.WriteString(d.Newline)
		}
	}
	return out.String()
}

func normalizedLines(text string) []string {
	if text == "" {
		return nil
	}
	parts := strings.SplitAfter(text, "\n")
	if parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

func rawLines(text string) []string {
	if text == "" {
		return nil
	}
	var lines []string
	start := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			lines = append(lines, text[start:i+1])
			start = i + 1
		}
	}
	if start < len(text) {
		lines = append(lines, text[start:])
	}
	return lines
}
