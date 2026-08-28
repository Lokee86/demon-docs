package reverseindex

import (
	"fmt"
	"sort"
	"strings"
)

func sortedSymbolReferences(values map[string]*symbolReference) []*symbolReference {
	result := make([]*symbolReference, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool {
		left, right := symbolDisplayName(result[i]), symbolDisplayName(result[j])
		if left != right {
			return left < right
		}
		leftLine, rightLine := symbolStartLine(result[i]), symbolStartLine(result[j])
		if leftLine != rightLine {
			return leftLine < rightLine
		}
		if result[i].Key != result[j].Key {
			return result[i].Key < result[j].Key
		}
		return result[i].Identity < result[j].Identity
	})
	return result
}

func symbolDisplayName(symbol *symbolReference) string {
	if symbol.QualifiedName != "" {
		return symbol.QualifiedName
	}
	return symbol.Name
}

func symbolStartLine(symbol *symbolReference) uint32 {
	if symbol.Span == nil {
		return 0
	}
	return symbol.Span.StartLine
}

func symbolLabel(symbol *symbolReference) string {
	name := strings.ReplaceAll(symbolDisplayName(symbol), "`", "'")
	details := make([]string, 0, 2)
	if symbol.Kind != "" {
		details = append(details, symbol.Kind)
	}
	if symbol.Span != nil && symbol.Span.StartLine > 0 {
		if symbol.Span.EndLine > symbol.Span.StartLine {
			details = append(details, fmt.Sprintf("lines %d-%d", symbol.Span.StartLine, symbol.Span.EndLine))
		} else {
			details = append(details, fmt.Sprintf("line %d", symbol.Span.StartLine))
		}
	}
	if len(details) == 0 {
		return fmt.Sprintf("Symbol `%s`", name)
	}
	return fmt.Sprintf("Symbol `%s` (%s)", name, strings.Join(details, ", "))
}
