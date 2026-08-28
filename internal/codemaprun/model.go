package codemaprun

import (
	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/codemapcorpus"
	"github.com/Lokee86/demon-docs/internal/codemaprecommend"
	"github.com/Lokee86/demon-docs/internal/codemapsemantic"
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
	CodeIntelligence        codemapcorpus.CodeIntelligenceProvider
	TargetResolver          codemap.TargetResolver
	RelationshipProvider    codemapcorpus.RelationshipProvider
	SemanticStaleness       codemap.SemanticStalenessProvider
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
	SemanticChanges []codemap.SemanticChange
	Before          []byte
	After           []byte
}

type Plan struct {
	Documents       []DocumentPlan
	Rewrites        []filetxn.Rewrite
	BaselineUpdates []codemapsemantic.Baseline
	RepositoryRoot  string
}

func (plan Plan) ChangedCount() int {
	return len(plan.Rewrites)
}

func (plan Plan) StaleCount() int {
	count := 0
	for _, document := range plan.Documents {
		if len(document.SemanticChanges) > 0 {
			count++
		}
	}
	return count
}
