package links

import (
	"fmt"
	"os"
	"testing"
)

func TestIsTransientFilesystemRace(t *testing.T) {
	for _, err := range []error{
		fmt.Errorf("read internal rewrite source: %w", os.ErrNotExist),
		fmt.Errorf("build generated rewrite: transformation 0 old destination mismatch"),
		fmt.Errorf("rewrite source changed before apply path.md"),
		fmt.Errorf("document changed since format plan creation"),
		fmt.Errorf("document changed since frontmatter plan creation"),
	} {
		if !IsTransientFilesystemRace(err) {
			t.Fatalf("error was not classified as transient: %v", err)
		}
	}
	if IsTransientFilesystemRace(fmt.Errorf("invalid configuration")) {
		t.Fatal("non-filesystem error was classified as transient")
	}
}
