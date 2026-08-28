package diagnostics

const SchemaVersion = 1

const (
	SeverityError   = "error"
	SeverityWarning = "warning"
	SeverityInfo    = "info"
)

type Diagnostic struct {
	Code        string   `json:"code"`
	Severity    string   `json:"severity"`
	Subsystem   string   `json:"subsystem"`
	Message     string   `json:"message"`
	Path        string   `json:"path,omitempty"`
	Line        int      `json:"line,omitempty"`
	Column      int      `json:"column,omitempty"`
	Target      string   `json:"target,omitempty"`
	Replacement string   `json:"replacement,omitempty"`
	Candidates  []string `json:"candidates,omitempty"`
}

type Report struct {
	SchemaVersion int          `json:"schema_version"`
	Command       string       `json:"command"`
	Status        string       `json:"status"`
	ExitCode      int          `json:"exit_code"`
	Diagnostics   []Diagnostic `json:"diagnostics"`
}
