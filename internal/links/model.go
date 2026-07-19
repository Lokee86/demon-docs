package links

import "github.com/Lokee86/demon-docs/internal/model"

const schemaVersion = 1

type FileRecord struct {
	ID          string   `json:"id"`
	Path        string   `json:"path"`
	Scope       string   `json:"scope"`
	Kind        string   `json:"kind"`
	Present     bool     `json:"present"`
	Fingerprint string   `json:"fingerprint,omitempty"`
	Size        int64    `json:"size,omitempty"`
	PathHistory []string `json:"path_history,omitempty"`
}

type FilesManifest struct {
	SchemaVersion int          `json:"schema_version"`
	Files         []FileRecord `json:"files"`
}

type LinkRecord struct {
	SourceFileID string   `json:"source_file_id"`
	SourcePath   string   `json:"source_path"`
	Ordinal      int      `json:"ordinal"`
	Line         int      `json:"line"`
	Column       int      `json:"column"`
	Syntax       string   `json:"syntax"`
	Target       string   `json:"target"`
	ResolvedPath string   `json:"resolved_path,omitempty"`
	TargetFileID string   `json:"target_file_id,omitempty"`
	Status       string   `json:"status"`
	Candidates   []string `json:"candidates,omitempty"`
}

type LinksManifest struct {
	SchemaVersion int          `json:"schema_version"`
	Links         []LinkRecord `json:"links"`
}

type Plan struct {
	Updates             []model.FileUpdate
	Messages            []string
	Files               FilesManifest
	Links               LinksManifest
	RepositoryRoot      string
	Initialized         bool
	NeedsInitialization bool
	Unresolved          int
}

func (p Plan) Failed() bool {
	return p.NeedsInitialization || p.Unresolved > 0 || len(p.Updates) > 0
}
