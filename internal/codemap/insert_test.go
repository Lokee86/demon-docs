package codemap

import "testing"

func TestInsertTargetAppendsBulletInsideConfiguredSection(t *testing.T) {
	source := "# Guide\n\n## Code map\n\n- `internal/old.go`\n\n## Notes\nKeep.\n"
	updated, start, end, inserted, err := InsertTarget(source, nil, "internal/new.go")
	if err != nil {
		t.Fatal(err)
	}
	if start != end || inserted != "- `internal/new.go`\n\n" {
		t.Fatalf("unexpected insertion: start=%d end=%d inserted=%q", start, end, inserted)
	}
	want := "# Guide\n\n## Code map\n\n- `internal/old.go`\n\n- `internal/new.go`\n\n## Notes\nKeep.\n"
	if updated != want {
		t.Fatalf("updated:\n%s\nwant:\n%s", updated, want)
	}
}
