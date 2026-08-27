package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"sort"
	"strings"
)

type collector struct {
	files             map[string]struct{}
	excluded          map[string]struct{}
	patternBoundaries map[string]map[string]struct{}
	items             map[string]map[string]*Evidence
}

func Collect(input Input) []Candidate {
	document := normalizePath(input.DocumentPath)
	files := normalizedSet(input.RepositoryFiles)
	excluded := normalizedSet(input.ExistingTargets)
	for target := range authoredCoverage(input.AuthoredTargets, files) {
		excluded[target] = struct{}{}
	}
	if document != "" {
		excluded[document] = struct{}{}
	}

	c := &collector{
		files:             files,
		excluded:          excluded,
		patternBoundaries: authoredPatternBoundaries(input.AuthoredTargets),
		items:             map[string]map[string]*Evidence{},
	}
	expansionTargets := evidenceExpansionTargets(input)
	semanticTargets := semanticRelationshipTargets(input)
	c.collectMentions(input.DocumentText)
	c.collectDeclaredSymbols(input.DocumentText, input.SymbolDeclarations)
	c.collectStructure(expansionTargets)
	c.collectDependencies(expansionTargets, input.DependencyEdges)
	c.collectSemanticRelationships(semanticTargets, input.SemanticRelationships)
	c.collectHistory(document, expansionTargets, input.Commits)
	c.collectRelatedDocuments(input.RelatedDocuments)
	return c.result()
}

func authoredCoverage(targets []AuthoredTarget, files map[string]struct{}) map[string]struct{} {
	covered := map[string]struct{}{}
	for _, authored := range targets {
		for _, resolved := range authored.ResolvedTargets {
			resolved = normalizePath(resolved)
			if resolved == "" {
				continue
			}
			if authored.Kind != AuthoredTargetDirectory {
				covered[resolved] = struct{}{}
				continue
			}
			directory := strings.TrimSuffix(resolved, "/")
			for candidate := range files {
				candidateBase := strings.TrimSuffix(candidate, "/")
				if candidateBase == directory || strings.HasPrefix(candidateBase, directory+"/") {
					covered[candidate] = struct{}{}
				}
			}
		}
	}
	return covered
}

func authoredPatternBoundaries(targets []AuthoredTarget) map[string]map[string]struct{} {
	result := map[string]map[string]struct{}{}
	for _, authored := range targets {
		if authored.Kind != AuthoredTargetPattern {
			continue
		}
		target := normalizePath(authored.Target)
		if target == "" {
			continue
		}
		directory := path.Dir(target)
		if strings.ContainsAny(directory, "*?") {
			continue
		}
		if result[directory] == nil {
			result[directory] = map[string]struct{}{}
		}
		for _, resolved := range authored.ResolvedTargets {
			if normalized := normalizePath(resolved); normalized != "" {
				result[directory][normalized] = struct{}{}
			}
		}
	}
	return result
}

func (c *collector) add(candidate string, kind Kind, source, detail string, count int) {
	candidate = normalizePath(candidate)
	source = normalizePath(source)
	if candidate == "" || count < 1 {
		return
	}
	if c.blockedByPatternBoundary(candidate, kind) {
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

func (c *collector) blockedByPatternBoundary(candidate string, kind Kind) bool {
	switch kind {
	case KindExactPathMention, KindUniqueBasenameMention, KindDeclaredSymbolMention:
		return false
	}
	boundary := c.patternBoundaries[path.Dir(strings.TrimSuffix(candidate, "/"))]
	if boundary == nil {
		return false
	}
	_, explicitlyCovered := boundary[candidate]
	return !explicitlyCovered
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
