package evidence

import "strings"

func (c *collector) collectDeclaredSymbols(documentText string, declarations []SymbolDeclaration) {
	pathsBySymbol := make(map[string]map[string]struct{})
	for _, declaration := range declarations {
		symbol := strings.TrimSpace(declaration.Symbol)
		candidate := normalizePath(declaration.Path)
		if symbol == "" || candidate == "" {
			continue
		}
		if pathsBySymbol[symbol] == nil {
			pathsBySymbol[symbol] = map[string]struct{}{}
		}
		pathsBySymbol[symbol][candidate] = struct{}{}
	}

	for symbol, paths := range pathsBySymbol {
		if len(paths) != 1 {
			continue
		}
		count := tokenCount(documentText, symbol)
		if count == 0 {
			continue
		}
		for candidate := range paths {
			c.add(candidate, KindDeclaredSymbolMention, "", symbol, count)
		}
	}
}
