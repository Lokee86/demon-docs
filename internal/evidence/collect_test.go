package evidence

import (
	"reflect"
	"strings"
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

func TestCollectPatternCoverageDoesNotExpandOutward(t *testing.T) {
	input := Input{
		DocumentPath: "docs/runtime.md",
		RepositoryFiles: []string{
			"internal/app/codemap_execute.go",
			"internal/app/codemap_precision.go",
			"internal/app/help_test.go",
			"cmd/ddocs/main.go",
		},
		ExistingTargets: []string{
			"internal/app/codemap_execute.go",
			"internal/app/codemap_precision.go",
		},
		AuthoredTargets: []AuthoredTarget{{
			Target: "internal/app/codemap_*.go",
			Kind:   AuthoredTargetPattern,
			ResolvedTargets: []string{
				"internal/app/codemap_execute.go",
				"internal/app/codemap_precision.go",
			},
		}},
		DependencyEdges: []DependencyEdge{
			{Source: "cmd/ddocs/main.go", Target: "internal/app/codemap_execute.go", Relation: "imports"},
		},
		Commits: []Commit{{
			ID:    "a",
			Paths: []string{"internal/app/codemap_execute.go", "internal/app/help_test.go"},
		}},
	}

	candidates := Collect(input)
	for _, candidate := range candidates {
		if candidate.Path == "internal/app/help_test.go" || candidate.Path == "cmd/ddocs/main.go" {
			t.Fatalf("pattern coverage expanded outward to %q: %#v", candidate.Path, candidate.Evidence)
		}
	}
}

func TestCollectDirectoryTargetCoversDescendants(t *testing.T) {
	input := Input{
		DocumentPath: "docs/runtime.md",
		DocumentText: "The codemap runtime is described here.",
		RepositoryFiles: []string{
			"internal/codemap/",
			"internal/codemap/managed.go",
			"internal/codemap/managed_render.go",
		},
		ExistingTargets: []string{"internal/codemap"},
		AuthoredTargets: []AuthoredTarget{{
			Target:          "internal/codemap/",
			Kind:            AuthoredTargetDirectory,
			ResolvedTargets: []string{"internal/codemap"},
		}},
		RelatedDocuments: []RelatedDocument{{
			Path:    "docs/managed.md",
			Targets: []string{"internal/codemap/managed.go"},
		}},
	}

	for _, candidate := range Collect(input) {
		if strings.HasPrefix(candidate.Path, "internal/codemap/") {
			t.Fatalf("directory-covered descendant was suggested: %#v", candidate)
		}
	}
}

func TestCollectPatternBoundaryBlocksInferredSiblingButAllowsDirectMention(t *testing.T) {
	base := Input{
		DocumentPath: "docs/runtime.md",
		RepositoryFiles: []string{
			"internal/app/codemap_execute.go",
			"internal/app/app.go",
		},
		ExistingTargets: []string{"internal/app/codemap_execute.go"},
		AuthoredTargets: []AuthoredTarget{{
			Target:          "internal/app/codemap_*.go",
			Kind:            AuthoredTargetPattern,
			ResolvedTargets: []string{"internal/app/codemap_execute.go"},
		}},
		RelatedDocuments: []RelatedDocument{{Path: "docs/app.md", Targets: []string{"internal/app/app.go"}}},
		Commits:          []Commit{{ID: "a", Paths: []string{"docs/runtime.md", "internal/app/app.go"}}},
	}
	for _, candidate := range Collect(base) {
		if candidate.Path == "internal/app/app.go" {
			t.Fatalf("pattern boundary admitted inferred sibling: %#v", candidate)
		}
	}

	base.DocumentText = "The explicit exception is `internal/app/app.go`."
	candidate := findCandidate(t, Collect(base), "internal/app/app.go")
	assertKind(t, candidate, KindExactPathMention)
}

func TestCollectExplicitFileStillExpandsOutward(t *testing.T) {
	input := Input{
		DocumentPath: "docs/runtime.md",
		RepositoryFiles: []string{
			"internal/app/codemap_execute.go",
			"internal/app/help_test.go",
			"cmd/ddocs/main.go",
		},
		ExistingTargets: []string{"internal/app/codemap_execute.go"},
		AuthoredTargets: []AuthoredTarget{{
			Target:          "internal/app/codemap_execute.go",
			Kind:            AuthoredTargetFile,
			ResolvedTargets: []string{"internal/app/codemap_execute.go"},
		}},
		DependencyEdges: []DependencyEdge{
			{Source: "cmd/ddocs/main.go", Target: "internal/app/codemap_execute.go", Relation: "imports"},
		},
	}

	candidates := Collect(input)
	assertKind(t, findCandidate(t, candidates, "internal/app/help_test.go"), KindSiblingTarget)
	assertKind(t, findCandidate(t, candidates, "cmd/ddocs/main.go"), KindDependencyNeighbor)
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
