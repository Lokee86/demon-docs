package reverseindex

import (
	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/model"
)

const section = "reverse-index"

type Plan struct {
	Updates        []model.FileUpdate
	Diagnostics    []string
	Orphans        []string
	IndexCount     int
	ReferenceCount int
}

func (p Plan) Failed() bool { return len(p.Updates) > 0 || len(p.Diagnostics) > 0 }

func (p Plan) CheckFailed() bool { return p.Failed() || len(p.Orphans) > 0 }

type symbolReference struct {
	Key           string
	Identity      string
	Kind          string
	Name          string
	QualifiedName string
	Span          *codemap.SemanticSpan
	Documents     map[string]struct{}
}

type facts struct {
	fileDocs   map[string]map[string]struct{}
	folderDocs map[string]map[string]struct{}
	symbolDocs map[string]map[string]*symbolReference
	exactFiles map[string]struct{}
	titles     map[string]string
}

func newFacts() facts {
	return facts{
		fileDocs:   map[string]map[string]struct{}{},
		folderDocs: map[string]map[string]struct{}{},
		symbolDocs: map[string]map[string]*symbolReference{},
		exactFiles: map[string]struct{}{},
		titles:     map[string]string{},
	}
}
