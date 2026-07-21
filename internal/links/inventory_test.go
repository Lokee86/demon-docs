package links

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func TestInventoryRebuildPrefersPresentDuplicatePathRecord(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "docs", "target.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("# Target\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	stored := storePath(root, path)
	inventory := &inventory{
		root: root,
		manifest: FilesManifest{Files: []FileRecord{
			{ID: "present", Path: stored, Scope: "repository", Kind: "file", Present: true},
			{ID: "historical", Path: stored, Scope: "repository", Kind: "file", Present: false},
		}},
	}
	inventory.rebuild()

	record, actual := inventory.exact(path)
	if record == nil {
		t.Fatal("present record was hidden by historical duplicate")
	}
	if record.ID != "present" {
		t.Fatalf("record ID = %q, want present", record.ID)
	}
	if filepath.Clean(actual) != filepath.Clean(path) {
		t.Fatalf("actual path = %q, want %q", actual, path)
	}

	resolved, actual, err := inventory.ensureTarget(path, "historical")
	if err != nil {
		t.Fatal(err)
	}
	if resolved == nil || resolved.ID != "present" {
		t.Fatalf("resolved record = %#v, want present record", resolved)
	}
	if filepath.Clean(actual) != filepath.Clean(path) {
		t.Fatalf("actual path = %q, want %q", actual, path)
	}
}

func TestBuildInventoryRefreshesChangedAndNewContentButReusesUnchangedMetadata(t *testing.T) {
	root := t.TempDir()
	unchangedPath := filepath.Join(root, "unchanged.md")
	changedPath := filepath.Join(root, "changed.md")
	newPath := filepath.Join(root, "new.md")
	unchangedID := "019f8337-225e-7150-8cd3-7863f843c719"
	oldChangedID := "019f8337-225e-7150-8cd3-7863f843c720"
	newChangedID := "019f8337-225e-7150-8cd3-7863f843c721"
	newFileID := "019f8337-225e-7150-8cd3-7863f843c722"
	writeInventoryTestFile(t, unchangedPath, "---\ndocument_id: "+unchangedID+"\n---\n# Same\n")
	writeInventoryTestFile(t, changedPath, "---\ndocument_id: "+oldChangedID+"\n---\n# Old\n")

	first, err := buildInventory(root, FilesManifest{})
	if err != nil {
		t.Fatal(err)
	}
	firstUnchanged := inventoryRecordByPath(t, first.manifest, "unchanged.md")
	firstChanged := inventoryRecordByPath(t, first.manifest, "changed.md")

	writeInventoryTestFile(t, changedPath, "---\ndocument_id: "+newChangedID+"\n---\n# New\n")
	changedTime := time.Unix(0, firstChanged.ModifiedUnixNano).Add(2 * time.Second)
	if err := os.Chtimes(changedPath, changedTime, changedTime); err != nil {
		t.Fatal(err)
	}
	writeInventoryTestFile(t, newPath, "---\ndocument_id: "+newFileID+"\n---\n# Added\n")

	second, err := buildInventory(root, first.manifest)
	if err != nil {
		t.Fatal(err)
	}
	secondUnchanged := inventoryRecordByPath(t, second.manifest, "unchanged.md")
	secondChanged := inventoryRecordByPath(t, second.manifest, "changed.md")
	secondNew := inventoryRecordByPath(t, second.manifest, "new.md")
	if secondUnchanged.ID != firstUnchanged.ID || secondUnchanged.Fingerprint != firstUnchanged.Fingerprint || secondUnchanged.DocumentID != unchangedID {
		t.Fatalf("unchanged metadata changed: first=%#v second=%#v", firstUnchanged, secondUnchanged)
	}
	if secondChanged.DocumentID != newChangedID || secondChanged.Fingerprint == firstChanged.Fingerprint {
		t.Fatalf("changed metadata was not refreshed: first=%#v second=%#v", firstChanged, secondChanged)
	}
	if secondNew.DocumentID != newFileID || secondNew.Fingerprint == "" {
		t.Fatalf("new content metadata was not recorded: %#v", secondNew)
	}

	third, err := buildInventory(root, second.manifest)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(second.manifest, third.manifest) {
		t.Fatalf("unchanged inventory output is not deterministic:\nsecond=%#v\nthird=%#v", second.manifest, third.manifest)
	}
}

func TestInventoryContentWorkersAreBoundedAndCollectErrorsByIndex(t *testing.T) {
	var active, maximum int32
	release := make(chan struct{})
	started := make(chan struct{}, linkUpdateWorkerLimit*2)
	done := make(chan []error, 1)
	go func() {
		done <- runLinkWorkers(linkUpdateWorkerLimit*2, func(index int) error {
			current := atomic.AddInt32(&active, 1)
			for {
				old := atomic.LoadInt32(&maximum)
				if current <= old || atomic.CompareAndSwapInt32(&maximum, old, current) {
					break
				}
			}
			started <- struct{}{}
			<-release
			atomic.AddInt32(&active, -1)
			if index == 0 {
				return errors.New("first content error")
			}
			return nil
		})
	}()

	for index := 0; index < linkUpdateWorkerLimit; index++ {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("bounded content workers did not start")
		}
	}
	if got := atomic.LoadInt32(&maximum); got > linkUpdateWorkerLimit {
		t.Fatalf("maximum concurrent workers = %d, want at most %d", got, linkUpdateWorkerLimit)
	}
	close(release)
	errorsByIndex := <-done
	if errorsByIndex[0] == nil || errorsByIndex[0].Error() != "first content error" {
		t.Fatalf("error at index 0 = %v, want first content error", errorsByIndex[0])
	}
}

func inventoryRecordByPath(t *testing.T, manifest FilesManifest, path string) *FileRecord {
	t.Helper()
	for index := range manifest.Files {
		if manifest.Files[index].Path == path {
			return &manifest.Files[index]
		}
	}
	t.Fatalf("inventory does not contain %q", path)
	return nil
}

func writeInventoryTestFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
