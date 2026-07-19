package evidence

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type signalValidationCorpus struct {
	SchemaVersion int                    `json:"schema_version"`
	Corpus        string                 `json:"corpus"`
	Revision      string                 `json:"revision"`
	Cases         []signalValidationCase `json:"cases"`
}

type signalValidationCase struct {
	Name             string            `json:"name"`
	Kind             Kind              `json:"kind"`
	Document         string            `json:"document"`
	Target           string            `json:"target"`
	EvidenceBasis    string            `json:"evidence_basis"`
	DocumentText     string            `json:"document_text"`
	RepositoryFiles  []string          `json:"repository_files"`
	ExistingTargets  []string          `json:"existing_targets"`
	DependencyEdges  []DependencyEdge  `json:"dependency_edges"`
	Commits          []Commit          `json:"commits"`
	RelatedDocuments []RelatedDocument `json:"related_documents"`
}

type trustedLinkCorpus struct {
	Corpus struct {
		Revision string `json:"revision"`
	} `json:"corpus"`
	Documents []struct {
		Document string `json:"document"`
		Links    []struct {
			Target string `json:"target"`
		} `json:"links"`
	} `json:"documents"`
}

func TestSpaceRocksTrustedLinksExerciseEveryEvidenceSignal(t *testing.T) {
	fixture := loadSignalValidationCorpus(t)
	trusted := loadTrustedLinkCorpus(t)
	if trusted.Corpus.Revision != fixture.Revision {
		t.Fatalf("validation revision %q does not match trusted corpus revision %q", fixture.Revision, trusted.Corpus.Revision)
	}

	trustedLinks := make(map[string]struct{})
	for _, document := range trusted.Documents {
		for _, link := range document.Links {
			trustedLinks[document.Document+"\x00"+link.Target] = struct{}{}
		}
	}

	expectedKinds := map[Kind]bool{
		KindExactPathMention:      false,
		KindUniqueBasenameMention: false,
		KindSiblingTarget:         false,
		KindTestCounterpart:       false,
		KindDependencyNeighbor:    false,
		KindGitDocumentCoChange:   false,
		KindGitTargetCoChange:     false,
		KindRelatedDocumentTarget: false,
	}

	for _, validationCase := range fixture.Cases {
		validationCase := validationCase
		t.Run(validationCase.Name, func(t *testing.T) {
			if _, ok := expectedKinds[validationCase.Kind]; !ok {
				t.Fatalf("fixture uses unknown evidence kind %q", validationCase.Kind)
			}
			if _, ok := trustedLinks[validationCase.Document+"\x00"+validationCase.Target]; !ok {
				t.Fatalf("%s -> %s is not in the trusted Space Rocks corpus", validationCase.Document, validationCase.Target)
			}
			if validationCase.EvidenceBasis == "" {
				t.Fatal("fixture case does not state whether its evidence was observed or isolated")
			}

			candidates := Collect(Input{
				DocumentPath:     validationCase.Document,
				DocumentText:     validationCase.DocumentText,
				RepositoryFiles:  validationCase.RepositoryFiles,
				ExistingTargets:  validationCase.ExistingTargets,
				DependencyEdges:  validationCase.DependencyEdges,
				Commits:          validationCase.Commits,
				RelatedDocuments: validationCase.RelatedDocuments,
			})
			candidate := findCandidate(t, candidates, validationCase.Target)
			assertKind(t, candidate, validationCase.Kind)
			for _, existing := range validationCase.ExistingTargets {
				for _, item := range candidates {
					if item.Path == normalizePath(existing) {
						t.Fatalf("existing target %q was returned as a missing-link candidate", existing)
					}
				}
			}
			expectedKinds[validationCase.Kind] = true
		})
	}

	for kind, covered := range expectedKinds {
		if !covered {
			t.Errorf("trusted-link validation has no case for evidence kind %q", kind)
		}
	}
}

func loadSignalValidationCorpus(t *testing.T) signalValidationCorpus {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join("testdata", "space-rocks-signal-cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	var corpus signalValidationCorpus
	if err := json.Unmarshal(contents, &corpus); err != nil {
		t.Fatal(err)
	}
	if corpus.SchemaVersion != 1 || corpus.Corpus != "space-rocks" || corpus.Revision == "" {
		t.Fatalf("invalid validation corpus metadata: %#v", corpus)
	}
	return corpus
}

func loadTrustedLinkCorpus(t *testing.T) trustedLinkCorpus {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join("..", "..", "research", "codemap-review", "space-rocks-trusted-links.json"))
	if err != nil {
		t.Fatal(err)
	}
	var corpus trustedLinkCorpus
	if err := json.Unmarshal(contents, &corpus); err != nil {
		t.Fatal(err)
	}
	return corpus
}
