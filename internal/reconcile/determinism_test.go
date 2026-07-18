package reconcile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/model"
)

func TestPlanningAndStaleMessagesAreDeterministic(t *testing.T) {
	root := filepath.Join(t.TempDir(), "docs")
	write(t, filepath.Join(root, "b", "README.md"), staleIndex("B", "z.md", "Z stale."))
	write(t, filepath.Join(root, "a", "README.md"), staleIndex("A", "y.md", "Y stale."))
	write(t, filepath.Join(root, "page.md"), "# Page\n")
	c := config.Default()

	var want string
	for iteration := 0; iteration < 20; iteration++ {
		result, err := Tree(root, c)
		if err != nil {
			t.Fatal(err)
		}
		got := resultSignature(result)
		if iteration == 0 {
			want = got
		} else if got != want {
			t.Fatalf("planning order changed on iteration %d\nwant:\n%s\ngot:\n%s", iteration, want, got)
		}
	}
	if strings.Index(want, filepath.Join("a", "README.md")) > strings.Index(want, filepath.Join("b", "README.md")) {
		t.Fatalf("stale messages not path-sorted:\n%s", want)
	}
}

func TestRepeatedApplyProducesIdenticalTreeAndNoFurtherUpdates(t *testing.T) {
	root := filepath.Join(t.TempDir(), "docs")
	write(t, filepath.Join(root, "z.md"), "# Z\n")
	write(t, filepath.Join(root, "A.md"), "# A\n")
	c := config.Default()
	first, err := Tree(root, c)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Apply(first); err != nil {
		t.Fatal(err)
	}
	before := snapshotTree(t, root)
	second, err := Tree(root, c)
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Updates) != 0 || len(second.Messages) != 0 {
		t.Fatalf("second plan was not stable: %+v", second)
	}
	if _, err := Apply(second); err != nil {
		t.Fatal(err)
	}
	after := snapshotTree(t, root)
	if before != after {
		t.Fatalf("second apply changed bytes\nbefore=%q\nafter=%q", before, after)
	}
}

func resultSignature(result model.ReconcileResult) string {
	var out strings.Builder
	for _, update := range result.Updates {
		fmt.Fprintf(&out, "update:%s:%q\n", update.Path, update.NewText)
	}
	for _, message := range result.Messages {
		fmt.Fprintf(&out, "message:%s\n", message)
	}
	return out.String()
}

func staleIndex(title, target, description string) string {
	return "# " + title + "\n\n## Direct Files\n<!-- doc-ledger:files:start -->\n\n- [" + target + "](" + target + ") - " + description + "\n<!-- doc-ledger:files:end -->\n\n## Stub Files\n<!-- doc-ledger:stubs:start -->\n<!-- doc-ledger:stubs:end -->\n\n## Direct Folders\n<!-- doc-ledger:folders:start -->\n<!-- doc-ledger:folders:end -->"
}

func snapshotTree(t *testing.T, root string) string {
	t.Helper()
	var out strings.Builder
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		fmt.Fprintf(&out, "%s:%q\n", filepath.ToSlash(relative), data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out.String()
}
