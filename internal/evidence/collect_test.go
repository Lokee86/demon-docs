package evidence

import (
	"reflect"
	"testing"
)

func TestCollectProducesDeterministicEvidence(t *testing.T) {
	input := Input{
		DocumentPath: "docs/respawn.md",
		DocumentText: "Respawn is implemented in `server/respawn.go`; see manager.go for coordination.",
		RepositoryFiles: []string{
			"server/manager_test.go",
			"server/respawn.go",
			"server/manager.go",
			"server/state.go",
			"docs/respawn.md",
		},
		ExistingTargets: []string{"server/manager.go"},
		DependencyEdges: []DependencyEdge{
			{Source: "server/manager.go", Target: "server/state.go", Relation: "imports"},
		},
		Commits: []Commit{
			{ID: "a", Paths: []string{"docs/respawn.md", "server/respawn.go"}},
			{ID: "b", Paths: []string{"server/manager.go", "server/state.go"}},
		},
		RelatedDocuments: []RelatedDocument{
			{Path: "docs/state.md", Targets: []string{"server/state.go"}},
		},
	}

	first := Collect(input)
	second := Collect(input)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("repeated collection differed:\nfirst: %#v\nsecond: %#v", first, second)
	}

	respawn := findCandidate(t, first, "server/respawn.go")
	assertKind(t, respawn, KindExactPathMention)
	assertKind(t, respawn, KindGitDocumentCoChange)

	testFile := findCandidate(t, first, "server/manager_test.go")
	assertKind(t, testFile, KindTestCounterpart)

	state := findCandidate(t, first, "server/state.go")
	assertKind(t, state, KindDependencyNeighbor)
	assertKind(t, state, KindGitTargetCoChange)
	assertKind(t, state, KindRelatedDocumentTarget)

	for _, candidate := range first {
		if candidate.Path == "server/manager.go" {
			t.Fatal("existing target was returned as a missing-link candidate")
		}
		if candidate.Fingerprint == "" {
			t.Fatalf("candidate %q has no fingerprint", candidate.Path)
		}
	}
}

func TestCollectPreservesDirectoryCandidates(t *testing.T) {
	input := Input{
		DocumentPath:    "docs/runtime.md",
		DocumentText:    "The manual producer remains at `services/diagnostic/cmd/submit/`.",
		RepositoryFiles: []string{"services/diagnostic/cmd/submit/"},
	}
	candidate := findCandidate(t, Collect(input), "services/diagnostic/cmd/submit/")
	assertKind(t, candidate, KindExactPathMention)
}

func TestUniqueBasenameMentionRequiresUniqueRepositoryPath(t *testing.T) {
	input := Input{
		DocumentPath: "docs/example.md",
		DocumentText: "The behavior lives in `worker.go` and `unique.go`.",
		RepositoryFiles: []string{
			"api/worker.go",
			"jobs/worker.go",
			"jobs/unique.go",
		},
	}
	candidates := Collect(input)
	findCandidate(t, candidates, "jobs/unique.go")
	for _, candidate := range candidates {
		if candidate.Path == "api/worker.go" || candidate.Path == "jobs/worker.go" {
			t.Fatalf("ambiguous basename produced candidate %q", candidate.Path)
		}
	}
}

func TestFingerprintChangesWhenEvidenceChanges(t *testing.T) {
	base := Input{
		DocumentPath:    "docs/example.md",
		DocumentText:    "See `server/example.go`.",
		RepositoryFiles: []string{"server/example.go"},
	}
	first := findCandidate(t, Collect(base), "server/example.go")
	base.Commits = []Commit{{ID: "a", Paths: []string{"docs/example.md", "server/example.go"}}}
	second := findCandidate(t, Collect(base), "server/example.go")
	if first.Fingerprint == second.Fingerprint {
		t.Fatal("fingerprint did not change when material evidence changed")
	}
}

func findCandidate(t *testing.T, candidates []Candidate, path string) Candidate {
	t.Helper()
	for _, candidate := range candidates {
		if candidate.Path == path {
			return candidate
		}
	}
	t.Fatalf("candidate %q not found in %#v", path, candidates)
	return Candidate{}
}

func assertKind(t *testing.T, candidate Candidate, kind Kind) {
	t.Helper()
	for _, item := range candidate.Evidence {
		if item.Kind == kind {
			return
		}
	}
	t.Fatalf("candidate %q lacks evidence kind %q: %#v", candidate.Path, kind, candidate.Evidence)
}
