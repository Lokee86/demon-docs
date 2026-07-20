package frontmatter

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Lokee86/demon-docs/internal/config"
)

func Evaluate(path string, document Document, schema config.Frontmatter, repair bool, recorded map[string]any, now time.Time) Outcome {
	values := cloneValues(document.Values)
	out := Outcome{Values: values, Immutable: map[string]any{}}
	unknownMode := strings.ToLower(strings.TrimSpace(schema.UnknownFields))

	unknown := make([]string, 0)
	for key := range values {
		if _, ok := schema.Fields[key]; !ok {
			unknown = append(unknown, key)
		}
	}
	sort.Strings(unknown)
	for _, key := range unknown {
		switch unknownMode {
		case "ignore":
		case "warn":
			out.add(path, key, "unknown frontmatter field", true, false)
		case "remove":
			out.add(path, key, "unknown frontmatter field; fix removes it", false, repair)
			if repair {
				delete(values, key)
				out.Changed = true
			}
		}
	}

	fieldNames := make([]string, 0, len(schema.Fields))
	for name := range schema.Fields {
		fieldNames = append(fieldNames, name)
	}
	sort.Strings(fieldNames)
	for _, name := range fieldNames {
		definition := schema.Fields[name]
		kind := normalizedType(definition.Type)
		current, present := values[name]
		prior, hasPrior := recorded[name]
		prior, priorErr := validateValue(kind, prior)
		hasPrior = hasPrior && priorErr == nil

		if definition.Immutable && present && hasPrior && !equalValue(current, prior) {
			out.add(path, name, "immutable field differs from its recorded value", false, repair)
			if repair {
				values[name] = prior
				current, present = prior, true
				out.Changed = true
			}
		}

		if !present || emptyValue(current) {
			available := definition.Immutable && hasPrior || hasConfiguredSource(definition, schema)
			if repair && available {
				replacement, ok, err := replacementValue(definition, schema, prior, hasPrior, now)
				if err != nil {
					out.add(path, name, err.Error(), false, false)
				} else if ok {
					values[name] = replacement
					current, present = replacement, true
					out.Changed = true
				}
			}
			if present && !emptyValue(current) {
				out.add(path, name, "frontmatter field was missing; fix added it", false, repair)
			} else if available {
				out.add(path, name, "frontmatter field is missing; fix can add it", false, false)
			} else if definition.Required {
				out.add(path, name, "required frontmatter field is missing or empty", false, false)
			}
		}

		if !present || emptyValue(current) {
			continue
		}
		normalized, err := validateValue(kind, current)
		if err != nil {
			resolved := false
			if definition.Immutable && repair {
				replacement, ok, replacementErr := replacementValue(definition, schema, prior, hasPrior, now)
				if replacementErr != nil {
					out.add(path, name, replacementErr.Error(), false, false)
				} else if ok {
					values[name] = replacement
					current = replacement
					out.Changed = true
					resolved = true
				}
			}
			out.add(path, name, err.Error(), false, resolved)
			if !resolved {
				continue
			}
			normalized = current
		}
		if definition.Immutable {
			out.Immutable[name] = normalized
		}
	}

	for _, rule := range schema.Rules {
		if !equalValue(values[rule.WhenField], rule.Equals) {
			continue
		}
		value, present := values[rule.Require]
		if present && !emptyValue(value) {
			continue
		}
		definition := schema.Fields[rule.Require]
		resolved := false
		if repair && hasConfiguredSource(definition, schema) {
			replacement, ok, err := replacementValue(definition, schema, nil, false, now)
			if err != nil {
				out.add(path, rule.Require, err.Error(), false, false)
			} else if ok {
				values[rule.Require] = replacement
				out.Changed = true
				resolved = true
			}
		}
		out.add(path, rule.Require, fmt.Sprintf("field is required when %s equals %v", rule.WhenField, rule.Equals), false, resolved)
	}
	return out
}
