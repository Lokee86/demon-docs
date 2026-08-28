package reverseindex

import (
	"fmt"
	"path/filepath"

	"github.com/Lokee86/demon-docs/internal/codemap"
)

func (f facts) addSymbolReference(record codemap.TargetRecord, document string) error {
	node := record.SemanticNode
	if record.Status != codemap.ResolutionResolved || node == nil {
		return fmt.Errorf("symbol projection requires one verified semantic node")
	}
	path := filepath.ToSlash(filepath.Clean(record.ResolvedPath))
	nodePath := filepath.ToSlash(filepath.Clean(node.Path))
	if path == "." || nodePath == "." || path == "" || nodePath == "" || path != nodePath {
		return fmt.Errorf("symbol projection path mismatch: target=%s node=%s", path, nodePath)
	}
	if node.Span != nil && node.Span.Path != "" {
		spanPath := filepath.ToSlash(filepath.Clean(node.Span.Path))
		if spanPath != nodePath {
			return fmt.Errorf("symbol projection span path mismatch: node=%s span=%s", nodePath, spanPath)
		}
	}
	identity := symbolProjectionIdentity(*node)
	if identity == "" {
		return fmt.Errorf("symbol projection has no stable identity")
	}
	if f.symbolDocs[path] == nil {
		f.symbolDocs[path] = map[string]*symbolReference{}
	}
	if existing := f.symbolDocs[path][identity]; existing != nil {
		if !sameSymbolReference(existing, *node) {
			return fmt.Errorf("symbol projection identity %s has conflicting metadata", identity)
		}
		existing.Documents[document] = struct{}{}
		return nil
	}
	f.symbolDocs[path][identity] = &symbolReference{
		Key: node.Key, Identity: node.Identity, Kind: node.Kind, Name: node.Name,
		QualifiedName: node.QualifiedName, Span: cloneSemanticSpan(node.Span),
		Documents: map[string]struct{}{document: {}},
	}
	return nil
}

func symbolProjectionIdentity(node codemap.SemanticNode) string {
	if node.Key != "" {
		return "key:" + node.Key
	}
	if node.Identity != "" {
		return "identity:" + node.Identity
	}
	return ""
}

func sameSymbolReference(existing *symbolReference, node codemap.SemanticNode) bool {
	if existing.Key != node.Key || existing.Identity != node.Identity || existing.Kind != node.Kind ||
		existing.Name != node.Name || existing.QualifiedName != node.QualifiedName {
		return false
	}
	return sameSemanticSpan(existing.Span, node.Span)
}

func sameSemanticSpan(left, right *codemap.SemanticSpan) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func cloneSemanticSpan(span *codemap.SemanticSpan) *codemap.SemanticSpan {
	if span == nil {
		return nil
	}
	copy := *span
	return &copy
}
