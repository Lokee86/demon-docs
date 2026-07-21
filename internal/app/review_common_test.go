package app

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestSuggestionsDoesNotRunCodemapGeneration(t *testing.T) {
	withWorkingDirectory(t, t.TempDir(), func(root string) {
		writeTestFile(t, filepath.Join(root, "docs", "source.md"), "# Source\n\n## Code map\n\n- [thing](../code/thing.go)\n")
		writeTestFile(t, filepath.Join(root, "code", "thing.go"), "package code\n")

		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"init", "--root", "docs"}, &stdout, &stderr); code != 0 {
			t.Fatalf("init code=%d stderr=%q", code, stderr.String())
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		stdout.Reset()
		stderr.Reset()
		if code := Run(ctx, []string{"suggestions"}, &stdout, &stderr); code != 0 {
			t.Fatalf("suggestions code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		if strings.Contains(stdout.String(), "codemap_link") || strings.Contains(stderr.String(), "codemap") {
			t.Fatalf("suggestions invoked codemap generation: stdout=%q stderr=%q", stdout.String(), stderr.String())
		}
	})
}

func TestSuggestionsHelpDescribesLinkSuggestionsOnly(t *testing.T) {
	var out bytes.Buffer
	suggestionsHelp(&out)
	if strings.Contains(out.String(), "codemap missing-link suggestions") {
		t.Fatalf("help still advertises codemap suggestions: %q", out.String())
	}
	if !strings.Contains(out.String(), "Codemap recommendations are owned by `ddocs codemaps`") {
		t.Fatalf("help does not state codemap ownership: %q", out.String())
	}
}
