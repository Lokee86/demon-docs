package reverseindex

import "github.com/Lokee86/demon-docs/internal/model"

const section = "reverse-index"

type Plan struct {
	Updates        []model.FileUpdate
	Diagnostics    []string
	IndexCount     int
	ReferenceCount int
}

func (p Plan) Failed() bool { return len(p.Updates) > 0 }

type facts struct {
	fileDocs       map[string]map[string]struct{}
	folderDocs     map[string]map[string]struct{}
	exactFiles     map[string]struct{}
	eligibleFolder map[string]struct{}
	titles         map[string]string
}

func newFacts() facts {
	return facts{
		fileDocs:       map[string]map[string]struct{}{},
		folderDocs:     map[string]map[string]struct{}{},
		exactFiles:     map[string]struct{}{},
		eligibleFolder: map[string]struct{}{},
		titles:         map[string]string{},
	}
}
