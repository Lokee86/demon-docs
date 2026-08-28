package reverseindex

import (
	"path/filepath"

	"github.com/Lokee86/demon-docs/internal/codemap"
)

func semanticResolutionInScope(repositoryRoot string, roots []string, record codemap.TargetRecord) bool {
	if record.SemanticNode != nil && semanticNodeInScope(repositoryRoot, roots, *record.SemanticNode) {
		return true
	}
	for _, node := range record.SemanticCandidates {
		if semanticNodeInScope(repositoryRoot, roots, node) {
			return true
		}
	}
	return false
}

func semanticNodeInScope(repositoryRoot string, roots []string, node codemap.SemanticNode) bool {
	if node.Path == "" {
		return false
	}
	full := filepath.Join(repositoryRoot, filepath.FromSlash(node.Path))
	return insideAny(filepath.Clean(full), roots)
}
