package codemap

// TargetKind is the lexical kind of an authored code-map target.
type TargetKind string

const (
	TargetUnknown   TargetKind = "unknown"
	TargetFile      TargetKind = "file"
	TargetDirectory TargetKind = "directory"
	TargetGlob      TargetKind = "glob"
	TargetSymbol    TargetKind = "symbol"
)

// SyntaxKind records the authored Markdown shape without prescribing one
// canonical source format.
type SyntaxKind string

const (
	SyntaxBullet            SyntaxKind = "bullet"
	SyntaxFenced            SyntaxKind = "fenced"
	SyntaxFencedArrow       SyntaxKind = "fenced-arrow"
	SyntaxFencedEquals      SyntaxKind = "fenced-equals"
	SyntaxFencedIndented    SyntaxKind = "fenced-indented"
	SyntaxFencedLeadingPath SyntaxKind = "fenced-leading-path"
	SyntaxLegacyIndented    SyntaxKind = "legacy-indented"
	SyntaxLegacyInline      SyntaxKind = "legacy-inline"
)

// SourceSpan identifies a target in its source document. Lines and columns are
// one-based UTF-8 byte positions; EndColumn is inclusive.
type SourceSpan struct {
	Line      int `json:"line"`
	Column    int `json:"column"`
	EndLine   int `json:"end_line"`
	EndColumn int `json:"end_column"`
}

// Entry is one normalized authored relationship from a document to a code
// target. Context is the nearest nested heading or prose group label inside the
// code-map section.
type Entry struct {
	DocumentPath string     `json:"document_path"`
	Heading      string     `json:"heading,omitempty"`
	Target       string     `json:"target"`
	Kind         TargetKind `json:"kind"`
	Syntax       SyntaxKind `json:"syntax,omitempty"`
	Context      string     `json:"context,omitempty"`
	Description  string     `json:"description,omitempty"`
	Source       SourceSpan `json:"source"`
	RawLine      string     `json:"raw_line"`
}

// Diagnostic records an authored code-map entry that the extractor could not
// normalize without guessing.
type Diagnostic struct {
	Code         string     `json:"code"`
	DocumentPath string     `json:"document_path"`
	Message      string     `json:"message"`
	Source       SourceSpan `json:"source"`
	RawLine      string     `json:"raw_line"`
}

// Result contains every extracted entry and every unsupported authored entry.
type Result struct {
	Entries      []Entry
	Diagnostics  []Diagnostic
	SectionCount int
}

type TargetBase string

const (
	TargetBaseRepository TargetBase = "repository"
	TargetBaseDocument   TargetBase = "document"
)

// Format describes how a repository labels and resolves code-map sections.
type Format struct {
	SectionHeadings []string
	TargetBase      TargetBase
	// TargetRoots are repository-relative component roots tried after the
	// primary target base. They allow component-relative authored paths without
	// silently guessing a component from document location.
	TargetRoots []string
}

func DefaultFormat() Format {
	return Format{
		SectionHeadings: []string{"code map", "codemap", "code or source map", "code and test map"},
		TargetBase:      TargetBaseRepository,
	}
}
