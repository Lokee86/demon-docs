package codemapcorpus

import (
	"bufio"
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"unicode"
)

func goDeclaredSymbols(file string, contents []byte) []string {
	parsed, err := parser.ParseFile(token.NewFileSet(), file, contents, parser.SkipObjectResolution)
	if err != nil || parsed == nil {
		return nil
	}
	set := map[string]struct{}{}
	for _, declaration := range parsed.Decls {
		switch value := declaration.(type) {
		case *ast.GenDecl:
			if value.Tok != token.TYPE {
				continue
			}
			for _, spec := range value.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok && typeSpec.Name != nil && specificBareSymbol(typeSpec.Name.Name) {
					set[typeSpec.Name.Name] = struct{}{}
				}
			}
		case *ast.FuncDecl:
			if value.Name == nil || !ast.IsExported(value.Name.Name) {
				continue
			}
			if value.Recv == nil || len(value.Recv.List) == 0 {
				if specificBareSymbol(value.Name.Name) {
					set[value.Name.Name] = struct{}{}
				}
				continue
			}
			if receiver := goReceiverName(value.Recv.List[0].Type); ast.IsExported(receiver) {
				set[receiver+"."+value.Name.Name] = struct{}{}
			}
		}
	}
	return sortedSet(set)
}

func goReceiverName(expression ast.Expr) string {
	switch value := expression.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.StarExpr:
		return goReceiverName(value.X)
	case *ast.IndexExpr:
		return goReceiverName(value.X)
	case *ast.IndexListExpr:
		return goReceiverName(value.X)
	default:
		return ""
	}
}

func gdscriptDeclaredSymbols(contents []byte) []string {
	set := map[string]struct{}{}
	className := ""
	scanner := bufio.NewScanner(bytes.NewReader(contents))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "class_name ") {
			className = leadingIdentifier(strings.TrimSpace(strings.TrimPrefix(line, "class_name ")))
			if specificBareSymbol(className) {
				set[className] = struct{}{}
			} else {
				className = ""
			}
			continue
		}
		line = strings.TrimPrefix(line, "static ")
		if !strings.HasPrefix(line, "func ") {
			continue
		}
		name := leadingIdentifier(strings.TrimSpace(strings.TrimPrefix(line, "func ")))
		if className != "" && name != "" {
			set[className+"."+name] = struct{}{}
		}
	}
	return sortedSet(set)
}

func specificBareSymbol(value string) bool {
	if len(value) < 8 {
		return false
	}
	uppercase := 0
	for _, character := range value {
		if unicode.IsUpper(character) {
			uppercase++
		}
	}
	return uppercase >= 2
}

func leadingIdentifier(value string) string {
	for index, character := range value {
		if index == 0 {
			if character != '_' && !unicode.IsLetter(character) {
				return ""
			}
			continue
		}
		if character != '_' && !unicode.IsLetter(character) && !unicode.IsDigit(character) {
			return value[:index]
		}
	}
	return value
}
