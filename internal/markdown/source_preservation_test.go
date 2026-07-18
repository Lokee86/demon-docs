package markdown

import (
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/config"
)

func TestEachManagedReplacementPreservesBytesOutsideItsBlock(t *testing.T) {
	c := config.Default()
	source := "# Docs\n\n<!-- author comment -->\n\n## Direct Files\n<!-- doc-ledger:files:start -->\n- [old.md](old.md) - Old.\n<!-- doc-ledger:files:end -->\n\n```markdown\n## Direct Folders\n<!-- doc-ledger:folders:start -->\n- [fake](fake/README.md) - Fake.\n<!-- doc-ledger:folders:end -->\n```\n\n## Stub Files\n<!-- doc-ledger:stubs:start -->\n- [old.md](stubs/old.md) - Old.\n<!-- doc-ledger:stubs:end -->\n\n## Direct Folders\n<!-- doc-ledger:folders:start -->\n- [old](old/README.md) - Old.\n<!-- doc-ledger:folders:end -->\n\n## Notes\nHand-authored tail  \n"

	for _, section := range []string{"files", "stubs", "folders"} {
		t.Run(section, func(t *testing.T) {
			startMarker := MarkerStart(c.Markers.Prefix, section)
			endMarker := MarkerEnd(c.Markers.Prefix, section)
			start := structuralIndex(source, startMarker, 0, fencedCodeRanges(source))
			endStart := structuralIndex(source, endMarker, start+len(startMarker), fencedCodeRanges(source))
			end := endStart + len(endMarker)
			got, err := ReplaceManaged(source, section, []string{"- [new](new) - New."}, c)
			if err != nil {
				t.Fatal(err)
			}
			newEndStart := structuralIndex(got, endMarker, start+len(startMarker), fencedCodeRanges(got))
			newEnd := newEndStart + len(endMarker)
			if got[:start] != source[:start] {
				t.Fatalf("prefix changed\nwant=%q\ngot=%q", source[:start], got[:start])
			}
			if got[newEnd:] != source[end:] {
				t.Fatalf("suffix changed\nwant=%q\ngot=%q", source[end:], got[newEnd:])
			}
			for _, unmanaged := range []string{"<!-- author comment -->", "```markdown\n## Direct Folders", "[fake](fake/README.md)", "Hand-authored tail  \n"} {
				if !strings.Contains(got, unmanaged) {
					t.Errorf("unmanaged bytes disappeared: %q", unmanaged)
				}
			}
		})
	}
}

func TestCustomHeadingAndMarkerLikeFenceContentStayUntouched(t *testing.T) {
	c := config.Default()
	c.Markers.Prefix = "nav"
	c.Sections.FilesHeading = "Pages"
	source := "# Docs\n\n```markdown\n## Pages\n<!-- nav:files:start -->\n- [fake](fake.md) - Fake.\n<!-- nav:files:end -->\n```\n\nComment after fence."
	got, err := ReplaceManaged(source, "files", []string{"- [real](real.md) - Real."}, c)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, source) {
		t.Fatalf("fenced source was rewritten:\n%s", got)
	}
	if strings.Count(got, "<!-- nav:files:start -->") != 2 || strings.Count(got, "<!-- nav:files:end -->") != 2 {
		t.Fatalf("real managed block was not added independently:\n%s", got)
	}
	if strings.Count(got, "## Pages") != 2 || !strings.Contains(got, "- [real](real.md) - Real.") {
		t.Fatalf("custom managed block missing:\n%s", got)
	}
}
