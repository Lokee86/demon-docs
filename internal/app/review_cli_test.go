package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReviewCLIRecordsUndoAndBlocksDeterministicRepair(t *testing.T) {
	withWorkingDirectory(t, t.TempDir(), func(root string) {
		writeTestFile(t, filepath.Join(root, "docs", "source.md"), "# Source\n\n[manual](old/manual.pdf)\n")
		writeTestFile(t, filepath.Join(root, "references", "manual.pdf"), "manual")

		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"init", "--root", "docs"}, &stdout, &stderr); code != 0 {
			t.Fatalf("init code=%d stderr=%q", code, stderr.String())
		}

		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"fix", "-l"}, &stdout, &stderr); code != 1 {
			t.Fatalf("baseline code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}

		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"fix", "-l"}, &stdout, &stderr); code != 0 {
			t.Fatalf("repair code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		repaired := readTestFile(t, filepath.Join(root, "docs", "source.md"))
		if !strings.Contains(repaired, "../references/manual.pdf") {
			t.Fatalf("link was not repaired: %q", repaired)
		}

		if err := os.Rename(filepath.Join(root, "docs", "source.md"), filepath.Join(root, "docs", "moved-source.md")); err != nil {
			t.Fatal(err)
		}

		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"changes", "docs/moved-source.md"}, &stdout, &stderr); code != 0 {
			t.Fatalf("changes code=%d stderr=%q", code, stderr.String())
		}
		fields := strings.Fields(stdout.String())
		if len(fields) == 0 || !strings.HasPrefix(fields[0], "ch-") {
			t.Fatalf("missing change ID: %q", stdout.String())
		}
		changeID := fields[0]

		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"changes", "undo", changeID, "--block", "--reason", "keep legacy target"}, &stdout, &stderr); code != 0 {
			t.Fatalf("undo code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		undone := readTestFile(t, filepath.Join(root, "docs", "moved-source.md"))
		if !strings.Contains(undone, "old/manual.pdf") {
			t.Fatalf("link was not undone: %q", undone)
		}

		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"fix", "-l"}, &stdout, &stderr); code != 1 {
			t.Fatalf("blocked fix code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		if !strings.Contains(stdout.String(), "Blocked link repair") {
			t.Fatalf("blocked repair was not reported: %q", stdout.String())
		}
		if got := readTestFile(t, filepath.Join(root, "docs", "moved-source.md")); got != undone {
			t.Fatalf("blocked repair was reapplied:\n%s", got)
		}

		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"suggestions", "docs/moved-source.md"}, &stdout, &stderr); code != 0 {
			t.Fatalf("suggestions code=%d stderr=%q", code, stderr.String())
		}
		if !strings.Contains(stdout.String(), "BLOCKED") || !strings.Contains(stdout.String(), "references/manual.pdf") {
			t.Fatalf("blocked repair is not inspectable: %q", stdout.String())
		}
	})
}

func TestSuggestionsSelectPreflightsAndAppliesOnlyChosenRepair(t *testing.T) {
	withWorkingDirectory(t, t.TempDir(), func(root string) {
		writeTestFile(t, filepath.Join(root, "docs", "source.md"), "# Source\n\n[manual](missing/manual.pdf)\n[guide](guide.md)\n")
		writeTestFile(t, filepath.Join(root, "references", "manual.pdf"), "manual")
		writeTestFile(t, filepath.Join(root, "docs", "a", "guide.md"), "# Guide A\n")
		writeTestFile(t, filepath.Join(root, "docs", "b", "guide.md"), "# Guide B\n")

		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"init", "--root", "docs"}, &stdout, &stderr); code != 0 {
			t.Fatalf("init code=%d stderr=%q", code, stderr.String())
		}

		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"fix", "-l"}, &stdout, &stderr); code != 1 {
			t.Fatalf("baseline code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}

		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"suggestions", "docs/source.md"}, &stdout, &stderr); code != 0 {
			t.Fatalf("suggestions code=%d stderr=%q", code, stderr.String())
		}
		fields := strings.Fields(stdout.String())
		if len(fields) == 0 || !strings.HasPrefix(fields[0], "sg-") {
			t.Fatalf("missing suggestion ID: %q", stdout.String())
		}
		suggestionID := fields[0]
		original := readTestFile(t, filepath.Join(root, "docs", "source.md"))

		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"suggestions", "select", suggestionID, "99"}, &stdout, &stderr); code != 2 {
			t.Fatalf("invalid selection code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		if got := readTestFile(t, filepath.Join(root, "docs", "source.md")); got != original {
			t.Fatalf("failed selection modified the source:\n%s", got)
		}

		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"suggestions", "select", suggestionID, "1"}, &stdout, &stderr); code != 0 {
			t.Fatalf("selection code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		selected := readTestFile(t, filepath.Join(root, "docs", "source.md"))
		if !strings.Contains(selected, "[manual](missing/manual.pdf)") {
			t.Fatalf("selection applied an unrelated deterministic repair: %q", selected)
		}
		if strings.Contains(selected, "[guide](guide.md)") || !(strings.Contains(selected, "[guide](a/guide.md)") || strings.Contains(selected, "[guide](b/guide.md)")) {
			t.Fatalf("selected repair was not applied: %q", selected)
		}
		if count := strings.Count(stdout.String(), "applied "); count != 1 {
			t.Fatalf("selection reported %d applied changes: %q", count, stdout.String())
		}

		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"fix", "-l"}, &stdout, &stderr); code != 0 {
			t.Fatalf("deferred repair code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		if got := readTestFile(t, filepath.Join(root, "docs", "source.md")); !strings.Contains(got, "../references/manual.pdf") {
			t.Fatalf("deferred deterministic repair was lost: %q", got)
		}
	})
}
