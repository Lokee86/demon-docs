package review

import (
	"strings"
	"testing"
)

func TestBuildUndoDataSupportsOneRepairWithinFileChange(t *testing.T) {
	before := []byte("[one](old-one.md)\r\n[two](old-two.md)\r\n")
	after := []byte("[one](new-one.md)\r\n[two](new-two-long.md)\r\n")
	change := Change{ID: "ch-1", Transformations: []Transformation{
		{ID: "rp-1", Start: strings.Index(string(before), "old-one.md"), End: strings.Index(string(before), "old-one.md") + len("old-one.md"), OldText: "old-one.md", NewText: "new-one.md"},
		{ID: "rp-2", Start: strings.Index(strings.ReplaceAll(string(before), "\r\n", "\n"), "old-two.md"), End: strings.Index(strings.ReplaceAll(string(before), "\r\n", "\n"), "old-two.md") + len("old-two.md"), OldText: "old-two.md", NewText: "new-two-long.md"},
	}}
	updated, err := BuildUndoData(change, before, after, "rp-2")
	if err != nil {
		t.Fatal(err)
	}
	want := "[one](new-one.md)\r\n[two](old-two.md)\r\n"
	if string(updated) != want {
		t.Fatalf("updated = %q, want %q", updated, want)
	}
}
