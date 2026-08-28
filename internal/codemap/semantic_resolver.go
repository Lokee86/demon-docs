package codemap

import "context"

type SemanticResolutionStatus string

const (
	SemanticResolved    SemanticResolutionStatus = "resolved"
	SemanticMissing     SemanticResolutionStatus = "missing"
	SemanticAmbiguous   SemanticResolutionStatus = "ambiguous"
	SemanticUnsupported SemanticResolutionStatus = "unsupported"
)

type SemanticSpan struct {
	Path        string `json:"path"`
	StartLine   uint32 `json:"start_line"`
	StartColumn uint32 `json:"start_column"`
	EndLine     uint32 `json:"end_line"`
	EndColumn   uint32 `json:"end_column"`
}

type SemanticNode struct {
	Key           string        `json:"key,omitempty"`
	Identity      string        `json:"identity,omitempty"`
	Kind          string        `json:"kind"`
	Path          string        `json:"path"`
	Name          string        `json:"name"`
	QualifiedName string        `json:"qualified_name,omitempty"`
	Span          *SemanticSpan `json:"span,omitempty"`
}

type SemanticResolution struct {
	Status SemanticResolutionStatus
	Nodes  []SemanticNode
}

type SymbolQuery struct {
	Name          string
	Path          string
	QualifiedName string
}

type SemanticChangeKind string

const (
	SemanticChangeDisappeared         SemanticChangeKind = "disappeared"
	SemanticChangeMoved               SemanticChangeKind = "moved"
	SemanticChangeIdentity            SemanticChangeKind = "identity_changed"
	SemanticChangeOwnership           SemanticChangeKind = "ownership_changed"
	SemanticChangeMetadata            SemanticChangeKind = "metadata_changed"
	SemanticChangeRelationships       SemanticChangeKind = "relationships_changed"
	SemanticChangeResolutionAmbiguous SemanticChangeKind = "resolution_ambiguous"
)

type SemanticChange struct {
	Target        string             `json:"target"`
	Kind          SemanticChangeKind `json:"kind"`
	PreviousPath  string             `json:"previous_path,omitempty"`
	CurrentPath   string             `json:"current_path,omitempty"`
	Identity      string             `json:"identity,omitempty"`
	QualifiedName string             `json:"qualified_name,omitempty"`
}

type SemanticStalenessProvider interface {
	SnapshotID() string
	AnalyzeStaleness(context.Context, string, []DatasetEntry) ([]SemanticChange, error)
}

// TargetResolver resolves exact repository files and symbols against an
// external deterministic code-intelligence snapshot. Unsupported means the
// resolver cannot safely attest the requested target and callers must degrade
// without guessing.
type TargetResolver interface {
	ResolveFile(context.Context, string) (SemanticResolution, error)
	ResolveSymbol(context.Context, SymbolQuery) (SemanticResolution, error)
}
