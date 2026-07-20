package documentpolicy

import (
	"fmt"
	"strings"
)

func ValidateSchema(schema Schema) error {
	if !safeSchemaName(strings.TrimSpace(schema.Name)) {
		return fmt.Errorf("name must be a safe document-type identifier")
	}
	if schema.Version < 0 || schema.Version > 1 {
		return fmt.Errorf("unsupported version %d", schema.Version)
	}
	if !allowedPolicy(schema.UnknownSections, "manual", "delete", "keep", "allow") {
		return fmt.Errorf("unknown_sections must be manual, delete, keep, or allow")
	}
	if !allowedPolicy(schema.DuplicateSections, "manual", "merge", "delete-first", "delete-last", "keep", "allow") {
		return fmt.Errorf("duplicate_sections must be manual, merge, delete-first, delete-last, keep, or allow")
	}
	if format := strings.ToLower(strings.TrimSpace(schema.Frontmatter.Format)); format != "" && format != "yaml" && format != "toml" {
		return fmt.Errorf("frontmatter.format must be yaml or toml")
	}

	sections := make(map[string]Section, len(schema.Sections))
	ordered := make([]Section, 0, len(schema.Sections))
	for index, section := range schema.Sections {
		section.ID = strings.TrimSpace(section.ID)
		section.Heading = strings.TrimSpace(section.Heading)
		section.Parent = strings.TrimSpace(section.Parent)
		section.After = strings.TrimSpace(section.After)
		if section.ID == "" || section.Heading == "" {
			return fmt.Errorf("sections[%d] requires id and heading", index)
		}
		if _, exists := sections[section.ID]; exists {
			return fmt.Errorf("duplicate section id %q", section.ID)
		}
		sections[section.ID] = section
		ordered = append(ordered, section)
	}

	for _, section := range ordered {
		if section.Parent == section.ID {
			return fmt.Errorf("section %q cannot be its own parent", section.ID)
		}
		if section.Parent != "" {
			if _, exists := sections[section.Parent]; !exists {
				return fmt.Errorf("section %q has unknown parent %q", section.ID, section.Parent)
			}
		}
		if section.After != "" {
			if section.After == section.ID {
				return fmt.Errorf("section %q cannot be positioned after itself", section.ID)
			}
			after, exists := sections[section.After]
			if !exists {
				return fmt.Errorf("section %q has unknown after target %q", section.ID, section.After)
			}
			if after.Parent != section.Parent {
				return fmt.Errorf("section %q after target %q is not a sibling", section.ID, section.After)
			}
		}
	}

	for _, section := range ordered {
		visited := map[string]bool{}
		current := section.ID
		for current != "" {
			if visited[current] {
				return fmt.Errorf("section parent cycle includes %q", current)
			}
			visited[current] = true
			current = sections[current].Parent
		}
	}

	headingsByParent := map[string]map[string]string{}
	for _, section := range ordered {
		if headingsByParent[section.Parent] == nil {
			headingsByParent[section.Parent] = map[string]string{}
		}
		for _, heading := range append([]string{section.Heading}, section.Aliases...) {
			key := strings.ToLower(strings.TrimSpace(heading))
			if key == "" {
				return fmt.Errorf("section %q contains an empty heading alias", section.ID)
			}
			if owner, exists := headingsByParent[section.Parent][key]; exists && owner != section.ID {
				return fmt.Errorf("sibling sections %q and %q share heading or alias %q", owner, section.ID, heading)
			}
			headingsByParent[section.Parent][key] = section.ID
		}
	}
	return nil
}

func allowedPolicy(value string, allowed ...string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}
