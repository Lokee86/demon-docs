package frontmatter

import (
	"bytes"

	"github.com/Lokee86/demon-docs/internal/validationcache"
)

const frontmatterIdentityVersion = "frontmatter-identity-v1"

// IdentityHash fingerprints only the leading frontmatter surface that Parse
// evaluates. Ordinary Markdown body edits therefore do not invalidate a clean
// frontmatter result. An unterminated block includes the remaining source, and
// a newly introduced second leading block changes the identity marker so the
// multiple-block diagnostic cannot be hidden by a prior clean cache entry.
func IdentityHash(data []byte) string {
	delimiter, firstEnd := leadingDelimiter(data)
	if delimiter == nil {
		return validationcache.ContentHash([]byte(frontmatterIdentityVersion + "\x00none"))
	}
	position := firstEnd
	for position < len(data) {
		lineEnd := nextLineEnd(data, position)
		if bytes.Equal(trimLineEnding(data[position:lineEnd]), delimiter) {
			blockEnd := lineEnd
			marker := byte(0)
			if next, _ := leadingDelimiter(data[blockEnd:]); next != nil {
				marker = 1
			}
			identity := make([]byte, 0, len(frontmatterIdentityVersion)+1+blockEnd+1)
			identity = append(identity, frontmatterIdentityVersion...)
			identity = append(identity, 0)
			identity = append(identity, data[:blockEnd]...)
			identity = append(identity, marker)
			return validationcache.ContentHash(identity)
		}
		position = lineEnd
	}
	identity := make([]byte, 0, len(frontmatterIdentityVersion)+14+len(data))
	identity = append(identity, frontmatterIdentityVersion...)
	identity = append(identity, "\x00unterminated\x00"...)
	identity = append(identity, data...)
	return validationcache.ContentHash(identity)
}

func leadingDelimiter(data []byte) ([]byte, int) {
	if len(data) < 4 {
		return nil, 0
	}
	firstEnd := nextLineEnd(data, 0)
	line := trimLineEnding(data[:firstEnd])
	if bytes.Equal(line, []byte("---")) || bytes.Equal(line, []byte("+++")) {
		return line, firstEnd
	}
	return nil, 0
}

func nextLineEnd(data []byte, start int) int {
	if relative := bytes.IndexByte(data[start:], '\n'); relative >= 0 {
		return start + relative + 1
	}
	return len(data)
}

func trimLineEnding(line []byte) []byte {
	line = bytes.TrimSuffix(line, []byte("\n"))
	return bytes.TrimSuffix(line, []byte("\r"))
}
