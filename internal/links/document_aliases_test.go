package links

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollapseDocumentIdentityAliasesRemapsLinksAndMergesHistory(t *testing.T) {
	const documentID = "019f7d55-fb2c-71df-8d63-571897b5dd4b"
	previousFiles := FilesManifest{SchemaVersion: schemaVersion, Files: []FileRecord{
		{ID: "source", Path: "README.md", Scope: "repository", Kind: "file", Present: true},
		{ID: "stale-target", DocumentID: documentID, Path: "docs/documentation-policy.md", Scope: "repository", Kind: "file", Present: true},
		{ID: "live-target", DocumentID: documentID, Path: "docs/documentation-policy-reference.md", Scope: "repository", Kind: "file", Present: false},
	}}
	previousLinks := LinksManifest{SchemaVersion: schemaVersion, Links: []LinkRecord{{
		ID: "link", SourceFileID: "source", TargetFileID: "stale-target", Ordinal: 0, Syntax: "inline", Target: "docs/documentation-policy.md", Status: "valid",
	}}}
	current := FilesManifest{SchemaVersion: schemaVersion, Files: []FileRecord{
		{ID: "source", Path: "README.md", Scope: "repository", Kind: "file", Present: true},
		{ID: "stale-target", DocumentID: documentID, Path: "docs/documentation-policy.md", Scope: "repository", Kind: "file", Present: false},
		{ID: "live-target", DocumentID: documentID, Path: "docs/documentation-policy-reference.md", PathHistory: []string{"docs/documentation-policy.md"}, Scope: "repository", Kind: "file", Present: true},
	}}

	previousFiles, previousLinks = collapseDocumentIdentityAliases(previousFiles, previousLinks, &current)

	for _, manifest := range []FilesManifest{previousFiles, current} {
		count := 0
		for _, record := range manifest.Files {
			if record.DocumentID != documentID {
				continue
			}
			count++
			if record.ID != "live-target" {
				t.Fatalf("canonical ID = %q, want live-target", record.ID)
			}
			if !containsString(record.PathHistory, "docs/documentation-policy.md") && record.Path != "docs/documentation-policy.md" {
				t.Fatalf("old path was not retained: %#v", record)
			}
		}
		if count != 1 {
			t.Fatalf("document identity records = %d, want 1: %#v", count, manifest.Files)
		}
	}
	if len(previousLinks.Links) != 1 || previousLinks.Links[0].TargetFileID != "live-target" {
		t.Fatalf("link target alias was not remapped: %#v", previousLinks.Links)
	}
}

func TestReconcileRecoversBrokenLinkFromDocumentPathHistory(t *testing.T) {
	root := t.TempDir()
	const documentID = "019f7d55-fb2c-71df-8d63-571897b5dd4b"
	oldRelative := "docs/documentation-policy.md"
	newRelative := "docs/documentation-policy-reference.md"
	oldPath := filepath.Join(root, filepath.FromSlash(oldRelative))
	newPath := filepath.Join(root, filepath.FromSlash(newRelative))
	writeTestFile(t, filepath.Join(root, "README.md"), "[Documentation Policy](docs/documentation-policy.md)\n")
	writeTestFile(t, oldPath, "---\ndocument_id: "+documentID+"\n---\n# Documentation Policies\n")

	baseline, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if baseline.Unresolved != 0 || len(baseline.Links.Links) != 1 {
		t.Fatalf("baseline state was not valid: %#v", baseline)
	}
	if err := Save(baseline); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		t.Fatal(err)
	}

	targetID := baseline.Links.Links[0].TargetFileID
	corrupted := baseline
	corrupted.RepositoryRoot = root
	for index := range corrupted.Files.Files {
		record := &corrupted.Files.Files[index]
		if record.ID == targetID {
			record.Present = false
			record.Path = oldRelative
		}
	}
	live := fileRecordByID(t, baseline.Files, targetID)
	live.ID = "live-duplicate"
	live.Path = newRelative
	live.PathHistory = []string{oldRelative}
	live.Present = true
	corrupted.Files.Files = append(corrupted.Files.Files, live)
	corrupted.Links.Links[0].TargetFileID = ""
	corrupted.Links.Links[0].ResolvedPath = ""
	corrupted.Links.Links[0].Status = "broken"
	if err := Save(corrupted); err != nil {
		t.Fatal(err)
	}

	plan, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Unresolved != 0 || len(plan.Updates) != 1 {
		t.Fatalf("historical path did not recover the deterministic rename: unresolved=%d updates=%d messages=%v links=%#v", plan.Unresolved, len(plan.Updates), plan.Messages, plan.Links.Links)
	}
	if !strings.Contains(plan.Updates[0].NewText, "docs/documentation-policy-reference.md") {
		t.Fatalf("link was not rewritten to the renamed document: %q", plan.Updates[0].NewText)
	}
	if len(plan.Links.Links) != 1 || plan.Links.Links[0].TargetFileID != "live-duplicate" {
		t.Fatalf("link did not adopt the live private identity: %#v", plan.Links.Links)
	}
	count := 0
	for _, record := range plan.Files.Files {
		if record.DocumentID == documentID {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("duplicate private document identities were not collapsed: %#v", plan.Files.Files)
	}
}

func fileRecordByID(t *testing.T, manifest FilesManifest, id string) FileRecord {
	t.Helper()
	for _, record := range manifest.Files {
		if record.ID == id {
			return record
		}
	}
	t.Fatalf("file record %q not found", id)
	return FileRecord{}
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
