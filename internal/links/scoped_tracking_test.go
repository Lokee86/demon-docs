package links

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTrackSourcesDoesNotInitializeMissingLinkState(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source.md")
	writeTestFile(t, source, "# Source\n")
	plan, err := TrackSources(root, []string{source})
	if err != nil {
		t.Fatal(err)
	}
	if !plan.NeedsInitialization || plan.Initialized {
		t.Fatalf("unexpected uninitialized scoped plan: %#v", plan)
	}
	if _, err := os.Stat(filepath.Join(root, ".ddocs")); !os.IsNotExist(err) {
		t.Fatalf("scoped tracking initialized private link state: %v", err)
	}
}

func TestTrackSourcesRefreshesOnlySelectedSourceRecords(t *testing.T) {
	root := t.TempDir()
	firstPath := filepath.Join(root, "first.md")
	secondPath := filepath.Join(root, "second.md")
	writeTestFile(t, firstPath, "# First\n\n[Target](target.md)\n")
	writeTestFile(t, secondPath, "# Second\n\n[Target](target.md)\n")
	writeTestFile(t, filepath.Join(root, "target.md"), "# Target\n")

	baseline, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := Save(baseline); err != nil {
		t.Fatal(err)
	}
	baseline.Suppressions = []Suppression{{
		SourceFileID: baseline.Files.Files[0].ID,
		Path:         firstPath,
	}}
	if err := Save(baseline); err != nil {
		t.Fatal(err)
	}
	baselineBySource := linksBySource(baseline.Links.Links)
	if err := os.WriteFile(firstPath, []byte("# First\n\nAdded context.\n\n[Target](target.md)\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	refresh, err := TrackSources(root, []string{firstPath})
	if err != nil {
		t.Fatal(err)
	}
	refreshedBySource := linksBySource(refresh.Links.Links)
	if len(refreshedBySource) != 2 {
		t.Fatalf("scoped refresh dropped untouched source records: %#v", refreshedBySource)
	}
	if refreshedBySource["first.md"][0].Start == baselineBySource["first.md"][0].Start {
		t.Fatalf("selected source link offsets were not refreshed: before=%d after=%d", baselineBySource["first.md"][0].Start, refreshedBySource["first.md"][0].Start)
	}
	if refreshedBySource["second.md"][0].Start != baselineBySource["second.md"][0].Start {
		t.Fatalf("untouched source record changed: before=%d after=%d", baselineBySource["second.md"][0].Start, refreshedBySource["second.md"][0].Start)
	}
	if len(refresh.Suppressions) != 1 || refresh.Suppressions[0].Path != firstPath {
		t.Fatalf("scoped refresh dropped pending watcher suppression: %#v", refresh.Suppressions)
	}
}

func linksBySource(records []LinkRecord) map[string][]LinkRecord {
	result := map[string][]LinkRecord{}
	for _, record := range records {
		result[record.SourcePath] = append(result[record.SourcePath], record)
	}
	return result
}
