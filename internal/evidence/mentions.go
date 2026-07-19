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
		if count := tokenCount(text, file); count > 0 {
			c.add(file, KindExactPathMention, "", file, count)
			continue
		}
		base := path.Base(file)
		if basenameCounts[base] == 1 {
			if count := tokenCount(text, base); count > 0 {
				c.add(file, KindUniqueBasenameMention, "", base, count)
			}
		}
	}
}

func tokenCount(text, token string) int {
	if token == "" {
		return 0
	}
	count := 0
	for offset := 0; offset <= len(text)-len(token); {
		index := strings.Index(text[offset:], token)
		if index < 0 {
			break
		}
		index += offset
		beforeOK := index == 0 || !isPathRune(rune(text[index-1]))
		after := index + len(token)
		afterOK := after == len(text) || !isPathRune(rune(text[after]))
		if beforeOK && afterOK {
			count++
			offset = after
			continue
		}
		offset = index + 1
	}
	return count
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
