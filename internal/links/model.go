package links

import (
	"github.com/Lokee86/demon-docs/internal/model"
	"github.com/Lokee86/demon-docs/internal/review"
)

const schemaVersion = 2

type FileRecord struct {
	ID                string   `json:"id"`
	Path              string   `json:"path"`
	Scope             string   `json:"scope"`
	Kind              string   `json:"kind"`
	Present           bool     `json:"present"`
	Fingerprint       string   `json:"fingerprint,omitempty"`
	Size              int64    `json:"size,omitempty"`
	ModifiedUnixNano  int64    `json:"modified_unix_nano,omitempty"`
	PathHistory       []string `json:"path_history,omitempty"`
	LinkParserVersion int      `json:"link_parser_version,omitempty"`
}

type FilesManifest struct {
	SchemaVersion int          `json:"schema_version"`
	Files         []FileRecord `json:"files"`
}

type LinkRecord struct {
	ID            string   `json:"id"`
	SourceFileID  string   `json:"source_file_id"`
	SourcePath    string   `json:"source_path"`
	Ordinal       int      `json:"ordinal"`
	Start         int      `json:"start"`
	End           int      `json:"end"`
	Line          int      `json:"line"`
	Column        int      `json:"column"`
	Syntax        string   `json:"syntax"`
	RawPath       string   `json:"raw_path"`
	Suffix        string   `json:"suffix,omitempty"`
	Angle         bool     `json:"angle,omitempty"`
	Target        string   `json:"target"`
	ResolvedPath  string   `json:"resolved_path,omitempty"`
	TargetFileID  string   `json:"target_file_id,omitempty"`
	Status        string   `json:"status"`
	Candidates    []string `json:"candidates,omitempty"`
	ParserVersion int      `json:"parser_version,omitempty"`
}

type LinksManifest struct {
	SchemaVersion int          `json:"schema_version"`
	Links         []LinkRecord `json:"links"`
}

type Plan struct {
	Updates             []model.FileUpdate
	Rewrites            []GeneratedRewrite
	Suppressions        []Suppression
	Messages            []string
	Files               FilesManifest
	Links               LinksManifest
	RepositoryRoot      string
	Initialized         bool
	NeedsInitialization bool
	Unresolved          int
	AppliedChanges      []review.Change
}

func (p Plan) Failed() bool {
	return p.NeedsInitialization || p.Unresolved > 0 || len(p.Updates) > 0 || len(p.Rewrites) > 0
}
