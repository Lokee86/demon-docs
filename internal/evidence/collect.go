package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
)

type collector struct {
	files    map[string]struct{}
	excluded map[string]struct{}
	items    map[string]map[string]*Evidence
}

func Collect(input Input) []Candidate {
	document := normalizePath(input.DocumentPath)
	files := normalizedSet(input.RepositoryFiles)
	excluded := normalizedSet(input.ExistingTargets)
	if document != "" {
		excluded[document] = struct{}{}
	}

	c := &collector{
		files:    files,
		excluded: excluded,
		items:    map[string]map[string]*Evidence{},
	}
	c.collectMentions(input.DocumentText)
	c.collectStructure(input.ExistingTargets)
	c.collectDependencies(input.ExistingTargets, input.DependencyEdges)
	c.collectHistory(document, input.ExistingTargets, input.Commits)
	c.collectRelatedDocuments(input.RelatedDocuments)
	return c.result()
}

func (c *collector) add(candidate string, kind Kind, source, detail string, count int) {
	candidate = normalizePath(candidate)
	source = normalizePath(source)
	if candidate == "" || count < 1 {
		return
	}
	if _, ok := c.files[candidate]; !ok {
		return
	}
	if _, ok := c.excluded[candidate]; ok {
		return
	}
	if c.items[candidate] == nil {
		c.items[candidate] = map[string]*Evidence{}
	}
	key := fmt.Sprintf("%s\x00%s\x00%s", kind, source, detail)
	if current := c.items[candidate][key]; current != nil {
		current.Count += count
		return
	}
	c.items[candidate][key] = &Evidence{Kind: kind, Source: source, Detail: detail, Count: count}
}

func (c *collector) result() []Candidate {
	paths := make([]string, 0, len(c.items))
	for candidate := range c.items {
		paths = append(paths, candidate)
	}
	sort.Strings(paths)

	result := make([]Candidate, 0, len(paths))
	for _, candidatePath := range paths {
		evidence := make([]Evidence, 0, len(c.items[candidatePath]))
		for _, item := range c.items[candidatePath] {
			evidence = append(evidence, *item)
		}
		sort.Slice(evidence, func(i, j int) bool {
			left := fmt.Sprintf("%s\x00%s\x00%s", evidence[i].Kind, evidence[i].Source, evidence[i].Detail)
			right := fmt.Sprintf("%s\x00%s\x00%s", evidence[j].Kind, evidence[j].Source, evidence[j].Detail)
			return left < right
		})
		result = append(result, Candidate{
			Path:        candidatePath,
			Evidence:    evidence,
			Fingerprint: fingerprint(candidatePath, evidence),
		})
	}
	return result
}

func fingerprint(candidate string, evidence []Evidence) string {
	hash := sha256.New()
	fmt.Fprintf(hash, "%s\n", candidate)
	for _, item := range evidence {
		fmt.Fprintf(hash, "%s\x00%s\x00%s\x00%d\n", item.Kind, item.Source, item.Detail, item.Count)
	}
	return hex.EncodeToString(hash.Sum(nil))
}
