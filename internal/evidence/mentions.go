package evidence

import (
	"path"
	"sort"
	"strings"
	"unicode"
)

func (c *collector) collectMentions(documentText string) {
	text := strings.ReplaceAll(documentText, `\`, "/")
	files := sortedKeys(c.files)
	basenameCounts := map[string]int{}
	for _, file := range files {
		basenameCounts[path.Base(file)]++
	}

	for _, file := range files {
		if containsToken(text, file) {
			c.add(file, KindExactPathMention, "", file, 1)
			continue
		}
		base := path.Base(file)
		if basenameCounts[base] == 1 && containsToken(text, base) {
			c.add(file, KindUniqueBasenameMention, "", base, 1)
		}
	}
}

func containsToken(text, token string) bool {
	if token == "" {
		return false
	}
	for offset := 0; offset <= len(text)-len(token); {
		index := strings.Index(text[offset:], token)
		if index < 0 {
			return false
		}
		index += offset
		beforeOK := index == 0 || !isPathRune(rune(text[index-1]))
		after := index + len(token)
		afterOK := after == len(text) || !isPathRune(rune(text[after]))
		if beforeOK && afterOK {
			return true
		}
		offset = index + 1
	}
	return false
}

func isPathRune(value rune) bool {
	return unicode.IsLetter(value) || unicode.IsDigit(value) || strings.ContainsRune("._-/\\", value)
}

func sortedFileList(files map[string]struct{}) []string {
	result := make([]string, 0, len(files))
	for file := range files {
		result = append(result, file)
	}
	sort.Strings(result)
	return result
}
