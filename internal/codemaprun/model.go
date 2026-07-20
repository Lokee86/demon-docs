package codemaprun

import (
	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/codemaprecommend"
	"github.com/Lokee86/demon-docs/internal/filetxn"
)

type Options struct {
	RepositoryRoot          string
	DocsRoot                string
	TargetFiles             []string
	Headings                []string
	MarkerPrefix            string
	RemoveUndiscoveredLinks bool
	RemoveLowScoreLinks     bool
	Schema                  codemap.SectionSchema
}

type Recommendation struct {
	codemaprecommend.Suggestion
	Declined bool
}

type DocumentPlan struct {
	Path            string
	SectionFound    bool
	SectionCreated  bool
	Changed         bool
	Existing        []string
	Recommendations []Recommendation
	Added           []string
	Removed         []string
	Suppressed      []string
	Before          []byte
	After           []byte
}

type Plan struct {
	Documents []DocumentPlan
	Rewrites  []filetxn.Rewrite
}

func (plan Plan) ChangedCount() int {
	return len(plan.Rewrites)
}
