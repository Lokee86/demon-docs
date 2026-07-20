package frontmatter

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

const (
	FormatYAML = "yaml"
	FormatTOML = "toml"
)

type Document struct {
	Format   string
	Values   map[string]any
	Body     string
	HasBlock bool
}

func Parse(source string, allowed []string) (Document, error) {
	format, delimiter := detectFormat(source)
	if format == "" {
		return Document{Values: map[string]any{}, Body: source}, nil
	}
	if !contains(allowed, format) {
		return Document{}, fmt.Errorf("unsupported front matter format %q", format)
	}
	lines := strings.SplitAfter(source, "\n")
	if len(lines) < 2 {
		return Document{}, fmt.Errorf("unterminated %s front matter", format)
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if trimLine(lines[i]) == delimiter {
			end = i
			break
		}
	}
	if end < 0 {
		return Document{}, fmt.Errorf("unterminated %s front matter", format)
	}
	block := strings.Join(lines[1:end], "")
	body := strings.Join(lines[end+1:], "")
	if next, _ := detectFormat(body); next != "" {
		return Document{}, fmt.Errorf("multiple leading front matter blocks")
	}
	values, err := decode(format, block)
	if err != nil {
		return Document{}, fmt.Errorf("parse %s front matter: %w", format, err)
	}
	return Document{Format: format, Values: values, Body: body, HasBlock: true}, nil
}

func Render(format string, values map[string]any, body string) (string, error) {
	var block string
	var err error
	switch format {
	case FormatYAML:
		block, err = renderYAML(values)
	case FormatTOML:
		block, err = renderTOML(values)
	default:
		return "", fmt.Errorf("unsupported front matter format %q", format)
	}
	if err != nil {
		return "", err
	}
	delimiter := "---"
	if format == FormatTOML {
		delimiter = "+++"
	}
	return delimiter + "\n" + block + delimiter + "\n" + body, nil
}

func detectFormat(source string) (string, string) {
	if strings.HasPrefix(source, "---\n") || strings.HasPrefix(source, "---\r\n") {
		return FormatYAML, "---"
	}
	if strings.HasPrefix(source, "+++\n") || strings.HasPrefix(source, "+++\r\n") {
		return FormatTOML, "+++"
	}
	return "", ""
}

// LeadingBlockEnd returns the byte offset immediately after a leading YAML or
// TOML front matter block. An unterminated leading block protects the remainder
// of the source from Markdown scanners and is diagnosed separately by Parse.
func LeadingBlockEnd(source string) int {
	_, delimiter := detectFormat(source)
	if delimiter == "" {
		return 0
	}
	lineEnd := strings.IndexByte(source, '\n')
	if lineEnd < 0 {
		return len(source)
	}
	position := lineEnd + 1
	for position < len(source) {
		relativeEnd := strings.IndexByte(source[position:], '\n')
		end := len(source)
		if relativeEnd >= 0 {
			end = position + relativeEnd + 1
		}
		if trimLine(source[position:end]) == delimiter {
			return end
		}
		position = end
	}
	return len(source)
}

func decode(format, block string) (map[string]any, error) {
	values := map[string]any{}
	if strings.TrimSpace(block) == "" {
		return values, nil
	}
	if format == FormatTOML {
		if _, err := toml.Decode(block, &values); err != nil {
			return nil, err
		}
		return normalizeMap(values), nil
	}
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(block), &node); err != nil {
		return nil, err
	}
	if err := rejectDuplicateYAMLKeys(&node); err != nil {
		return nil, err
	}
	if err := node.Decode(&values); err != nil {
		return nil, err
	}
	return normalizeMap(values), nil
}

func rejectDuplicateYAMLKeys(node *yaml.Node) error {
	if node.Kind == yaml.MappingNode {
		seen := map[string]bool{}
		for i := 0; i+1 < len(node.Content); i += 2 {
			key := node.Content[i].Value
			if seen[key] {
				return fmt.Errorf("duplicate key %q", key)
			}
			seen[key] = true
			if err := rejectDuplicateYAMLKeys(node.Content[i+1]); err != nil {
				return err
			}
		}
	}
	for _, child := range node.Content {
		if node.Kind != yaml.MappingNode {
			if err := rejectDuplicateYAMLKeys(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func renderYAML(values map[string]any) (string, error) {
	root := &yaml.Node{Kind: yaml.MappingNode}
	for _, key := range sortedKeys(values) {
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}
		valueNode := &yaml.Node{}
		if err := valueNode.Encode(values[key]); err != nil {
			return "", err
		}
		root.Content = append(root.Content, keyNode, valueNode)
	}
	var buffer bytes.Buffer
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	if err := encoder.Encode(root); err != nil {
		return "", err
	}
	_ = encoder.Close()
	return buffer.String(), nil
}

func renderTOML(values map[string]any) (string, error) {
	var output bytes.Buffer
	if err := toml.NewEncoder(&output).Encode(values); err != nil {
		return "", err
	}
	return output.String(), nil
}

func normalizeMap(values map[string]any) map[string]any {
	result := make(map[string]any, len(values))
	for key, value := range values {
		result[key] = normalize(value)
	}
	return result
}

func normalize(value any) any {
	switch typed := value.(type) {
	case time.Time:
		return typed.Format("2006-01-02")
	case fmt.Stringer:
		text := typed.String()
		if _, err := time.Parse("2006-01-02", text); err == nil {
			return text
		}
		return value
	case int:
		return int64(typed)
	case int32:
		return int64(typed)
	case []string:
		result := make([]any, len(typed))
		for i, item := range typed {
			result[i] = item
		}
		return result
	case []interface{}:
		result := make([]any, len(typed))
		for i, item := range typed {
			result[i] = normalize(item)
		}
		return result
	case map[string]interface{}:
		return normalizeMap(typed)
	default:
		return value
	}
}

func sortedKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func trimLine(line string) string { return strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r") }

func contains(values []string, want string) bool {
	for _, value := range values {
		if strings.EqualFold(value, want) {
			return true
		}
	}
	return false
}
