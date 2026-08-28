package codemaparcana

import (
	"sort"

	"github.com/Lokee86/demon-docs/internal/codemap"
)

func currentCounterpart(previous protocolNode, diff semanticDiffResult) *protocolNode {
	current := uniqueChangedNodes(diff)
	if previous.Key != "" {
		for _, node := range current {
			if node.Key == previous.Key {
				copy := node
				return &copy
			}
		}
	}
	if previous.Kind == "file" && previous.ContentID != "" {
		if match := uniqueProtocolMatch(current, func(node protocolNode) bool {
			return node.Kind == "file" && node.ContentID == previous.ContentID
		}); match != nil {
			return match
		}
	}
	return uniqueProtocolMatch(current, func(node protocolNode) bool {
		return sameLogicalProtocolNode(node, previous)
	})
}

func uniqueChangedNodes(diff semanticDiffResult) []protocolNode {
	seen := map[string]struct{}{}
	var result []protocolNode
	for _, group := range [][]protocolNode{diff.Nodes.Added, diff.Nodes.MetadataChanged, diff.Nodes.RelationshipChanged} {
		for _, node := range group {
			key := node.Key
			if key == "" {
				key = node.Identity
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			result = append(result, node)
		}
	}
	return result
}

func uniqueProtocolMatch(nodes []protocolNode, matches func(protocolNode) bool) *protocolNode {
	var match *protocolNode
	for _, node := range nodes {
		if !matches(node) {
			continue
		}
		if match != nil {
			return nil
		}
		copy := node
		match = &copy
	}
	return match
}

func sameLogicalProtocolNode(node, previous protocolNode) bool {
	if node.Kind != previous.Kind || node.Name == "" || node.Name != previous.Name {
		return false
	}
	if node.QualifiedName != "" && previous.QualifiedName != "" {
		return node.QualifiedName == previous.QualifiedName || qualifiedTail(node.QualifiedName) == qualifiedTail(previous.QualifiedName)
	}
	return true
}

func changedNodeSet(nodes []protocolNode) map[string]struct{} {
	result := make(map[string]struct{}, len(nodes)*2)
	for _, node := range nodes {
		if node.Key != "" {
			result["key:"+node.Key] = struct{}{}
		}
		if node.Identity != "" {
			result["identity:"+node.Identity] = struct{}{}
		}
	}
	return result
}

func appendChangedNodeSignals(changes *[]codemap.SemanticChange, target string, current codemap.SemanticNode, metadata, relationships map[string]struct{}) {
	if semanticNodeInSet(current, metadata) {
		*changes = append(*changes, semanticChange(target, codemap.SemanticChangeMetadata, nil, &current))
	}
	if semanticNodeInSet(current, relationships) {
		*changes = append(*changes, semanticChange(target, codemap.SemanticChangeRelationships, nil, &current))
	}
}

func semanticNodeInSet(node codemap.SemanticNode, set map[string]struct{}) bool {
	if node.Key != "" {
		if _, ok := set["key:"+node.Key]; ok {
			return true
		}
	}
	_, ok := set["identity:"+node.Identity]
	return ok
}

func semanticIdentityChanged(previous, current codemap.SemanticNode) bool {
	if normalizeRepositoryPath(previous.Path) != normalizeRepositoryPath(current.Path) || previous.QualifiedName != current.QualifiedName {
		return true
	}
	if previous.Key != "" && current.Key != "" {
		return previous.Key != current.Key
	}
	return previous.Identity != current.Identity
}

func changedIdentityKind(previous, current codemap.SemanticNode) codemap.SemanticChangeKind {
	if normalizeRepositoryPath(previous.Path) != normalizeRepositoryPath(current.Path) {
		return codemap.SemanticChangeMoved
	}
	if previous.QualifiedName != "" && current.QualifiedName != "" && previous.QualifiedName != current.QualifiedName {
		return codemap.SemanticChangeOwnership
	}
	return codemap.SemanticChangeIdentity
}

func semanticChange(target string, kind codemap.SemanticChangeKind, previous, current *codemap.SemanticNode) codemap.SemanticChange {
	change := codemap.SemanticChange{Target: target, Kind: kind}
	if previous != nil {
		change.PreviousPath = previous.Path
		change.Identity = previous.Identity
		change.QualifiedName = previous.QualifiedName
	}
	if current != nil {
		change.CurrentPath = current.Path
		change.Identity = current.Identity
		change.QualifiedName = current.QualifiedName
	}
	return change
}

func protocolSemanticNode(node protocolNode) codemap.SemanticNode {
	return codemap.SemanticNode{Key: node.Key, Identity: node.Identity, Kind: node.Kind, Path: normalizeRepositoryPath(node.Path), Name: node.Name, QualifiedName: node.QualifiedName}
}

func dedupeSemanticChanges(changes []codemap.SemanticChange) []codemap.SemanticChange {
	seen := map[string]struct{}{}
	result := make([]codemap.SemanticChange, 0, len(changes))
	for _, change := range changes {
		key := change.Target + "\x00" + string(change.Kind) + "\x00" + change.PreviousPath + "\x00" + change.CurrentPath
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, change)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Target != result[j].Target {
			return result[i].Target < result[j].Target
		}
		return result[i].Kind < result[j].Kind
	})
	return result
}
