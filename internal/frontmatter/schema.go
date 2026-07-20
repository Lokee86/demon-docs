package frontmatter

type Diagnostic struct {
	Path     string
	Field    string
	Message  string
	Warning  bool
	Resolved bool
}

type Outcome struct {
	Values      map[string]any
	Diagnostics []Diagnostic
	Changed     bool
	Immutable   map[string]any
}
