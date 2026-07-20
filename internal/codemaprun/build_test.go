package codemaprun

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Lokee86/demon-docs/internal/ddrepo"
	"github.com/Lokee86/demon-docs/internal/review"
)

func TestBuildAdoptsExistingSectionAndAddsRecommendation(t *testing.T) {
	root := t.TempDir()
	docs := filepath.Join(root, "docs")
	writeFile(t, filepath.Join(docs, "runtime.md"), "# Runtime\n\nThe implementation is in `src/runtime.go`.\n\n## Code Map\n")
	writeFile(t, filepath.Join(root, "src", "runtime.go"), "package runtime\n")

	plan, err := Build(context.Background(), Options{
		RepositoryRoot: root,
		DocsRoot:       docs,
		TargetFiles:    []string{filepath.Join(docs, "runtime.md")},
		Headings:       []string{"Code Map"},
		MarkerPrefix:   "ddocs",
	})
	if err != nil {
		t.Fatal(err)
	}
	if plan.ChangedCount() != 1 || len(plan.Documents) != 1 {
		t.Fatalf("unexpected plan: %#v", plan)
	}
	document := plan.Documents[0]
	if len(document.Added) != 1 || document.Added[0] != "src/runtime.go" {
		t.Fatalf("unexpected additions: %#v", document)
	}
	text := string(document.After)
	for _, want := range []string{"<!-- ddocs:codemap:start -->", "- `src/runtime.go`", "<!-- ddocs:codemap:end -->"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q:\n%s", want, text)
		}
	}
	if err := Apply(plan); err != nil {
		t.Fatal(err)
	}
	written, err := os.ReadFile(filepath.Join(docs, "runtime.md"))
	if err != nil || string(written) != text {
		t.Fatalf("written=%q err=%v", written, err)
	}
}

func TestBuildPreservesUndiscoveredLinksByDefault(t *testing.T) {
	root := t.TempDir()
	docs := filepath.Join(root, "docs")
	writeFile(t, filepath.Join(docs, "runtime.md"), "# Runtime\n\n## Code Map\n\n- `src/manual.go`\n")
	writeFile(t, filepath.Join(root, "src", "manual.go"), "package runtime\n")

	plan, err := Build(context.Background(), Options{
		RepositoryRoot: root,
		DocsRoot:       docs,
		TargetFiles:    []string{filepath.Join(docs, "runtime.md")},
		Headings:       []string{"Code Map"},
		MarkerPrefix:   "ddocs",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Documents[0].Removed) != 0 || !strings.Contains(string(plan.Documents[0].After), "src/manual.go") {
		t.Fatalf("manual link was not preserved: %#v", plan.Documents[0])
	}
}

func TestBuildCanRemoveUndiscoveredLinksWhenConfigured(t *testing.T) {
	root := t.TempDir()
	docs := filepath.Join(root, "docs")
	writeFile(t, filepath.Join(docs, "runtime.md"), "# Runtime\n\n## Code Map\n\n- `src/manual.go`\n")
	writeFile(t, filepath.Join(root, "src", "manual.go"), "package runtime\n")

	plan, err := Build(context.Background(), Options{
		RepositoryRoot:          root,
		DocsRoot:                docs,
		TargetFiles:             []string{filepath.Join(docs, "runtime.md")},
		Headings:                []string{"Code Map"},
		MarkerPrefix:            "ddocs",
		RemoveUndiscoveredLinks: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Documents[0].Removed) != 1 || strings.Contains(string(plan.Documents[0].After), "src/manual.go") {
		t.Fatalf("undiscovered link was not removed: %#v", plan.Documents[0])
	}
}

func TestBuildHonorsSharedDeclinedSuggestionPolicy(t *testing.T) {
	root := t.TempDir()
	docs := filepath.Join(root, "docs")
	writeFile(t, filepath.Join(docs, "runtime.md"), "# Runtime\n\nThe implementation is in `src/runtime.go`.\n\n## Code Map\n")
	writeFile(t, filepath.Join(root, "src", "runtime.go"), "package runtime\n")
	options := Options{
		RepositoryRoot: root,
		DocsRoot:       docs,
		TargetFiles:    []string{filepath.Join(docs, "runtime.md")},
		Headings:       []string{"Code Map"},
		MarkerPrefix:   "ddocs",
	}
	initial, err := Build(context.Background(), options)
	if err != nil || len(initial.Documents[0].Recommendations) == 0 {
		t.Fatalf("initial recommendations=%#v err=%v", initial.Documents, err)
	}
	item := initial.Documents[0].Recommendations[0].Suggestion
	if _, err := ddrepo.Init(root); err != nil {
		t.Fatal(err)
	}
	suggestion := review.CodemapSuggestion(item.Document, item.Target, item.Score, string(item.Tier), item.Evidence)
	store, err := review.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	decision := review.Decision{
		ID:           review.NewID("dc"),
		Action:       review.DecisionDeclineIssue,
		RelationKey:  suggestion.RelationKey,
		Fingerprint:  suggestion.Fingerprint,
		SuggestionID: suggestion.ID,
		Suggestion:   &suggestion,
		DecidedAt:    now,
	}
	if _, err := store.Append(review.Event{Type: review.EventDecision, Time: now, Decision: &decision}, nil, nil); err != nil {
		t.Fatal(err)
	}
	plan, err := Build(context.Background(), options)
	if err != nil {
		t.Fatal(err)
	}
	document := plan.Documents[0]
	if len(document.Added) != 0 || len(document.Suppressed) != 1 || document.Suppressed[0] != "src/runtime.go" {
		t.Fatalf("decline was not applied: %#v", document)
	}
}

func TestBuildDoesNotCreateMissingSectionWithoutSchema(t *testing.T) {
	root := t.TempDir()
	docs := filepath.Join(root, "docs")
	writeFile(t, filepath.Join(docs, "runtime.md"), "# Runtime\n\nThe implementation is in `src/runtime.go`.\n")
	writeFile(t, filepath.Join(root, "src", "runtime.go"), "package runtime\n")

	plan, err := Build(context.Background(), Options{
		RepositoryRoot: root,
		DocsRoot:       docs,
		TargetFiles:    []string{filepath.Join(docs, "runtime.md")},
		Headings:       []string{"Code Map"},
		MarkerPrefix:   "ddocs",
	})
	if err != nil {
		t.Fatal(err)
	}
	if plan.ChangedCount() != 0 || plan.Documents[0].SectionFound {
		t.Fatalf("missing section was unexpectedly created: %#v", plan.Documents[0])
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
