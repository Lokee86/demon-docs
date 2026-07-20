package links

import (
	"sort"
	"strings"

	"github.com/Lokee86/demon-docs/internal/frontmatter"
)

const linkParserVersion = 2

type undefinedReference struct {
	Start, End   int
	Line, Column int
	Label        string
}

type parsedMarkdown struct {
	Links               []occurrence
	UndefinedReferences []undefinedReference
}

func parseMarkdownDocument(source string) parsedMarkdown {
	protected := protectedMarkdownRanges(source)
	return parsedMarkdown{
		Links:               parseMarkdownLinks(source),
		UndefinedReferences: parseUndefinedReferences(source, protected),
	}
}

func protectedMarkdownRanges(source string) []byteRange {
	ranges := fencedRanges(source)
	if end := frontmatter.LeadingBlockEnd(source); end > 0 {
		ranges = append(ranges, byteRange{Start: 0, End: end})
	}
	for i := 0; i < len(source); {
		if end, ok := rangeEnd(i, ranges); ok {
			i = end
			continue
		}
		if source[i] != '`' || escaped(source, i) {
			i++
			continue
		}
		run := repeated(source, i, '`')
		needle := strings.Repeat("`", run)
		close := strings.Index(source[i+run:], needle)
		if close < 0 {
			i += run
			continue
		}
		end := i + run + close + run
		ranges = append(ranges, byteRange{Start: i, End: end})
		i = end
	}
	sort.Slice(ranges, func(i, j int) bool { return ranges[i].Start < ranges[j].Start })
	return ranges
}
