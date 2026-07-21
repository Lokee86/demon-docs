package frontmatter

import "testing"

func TestIdentityHashIgnoresBodyOnlyChanges(t *testing.T) {
	first := []byte("---\ndocument_id: doc-1\ndocument_type: guide\n---\n# First\nBody one.\n")
	second := []byte("---\ndocument_id: doc-1\ndocument_type: guide\n---\n# Second\nBody two.\n")
	if IdentityHash(first) != IdentityHash(second) {
		t.Fatal("body-only change altered frontmatter identity")
	}
}

func TestIdentityHashChangesWithRawFrontmatter(t *testing.T) {
	lf := []byte("---\ndocument_id: doc-1\n---\n# Guide\n")
	crlf := []byte("---\r\ndocument_id: doc-1\r\n---\r\n# Guide\r\n")
	changed := []byte("---\ndocument_id: doc-2\n---\n# Guide\n")
	if IdentityHash(lf) == IdentityHash(crlf) {
		t.Fatal("frontmatter line-ending change did not alter identity")
	}
	if IdentityHash(lf) == IdentityHash(changed) {
		t.Fatal("frontmatter value change did not alter identity")
	}
}

func TestIdentityHashDetectsSecondLeadingBlock(t *testing.T) {
	clean := []byte("---\ndocument_id: doc-1\n---\n# Guide\n")
	multiple := []byte("---\ndocument_id: doc-1\n---\n---\nsummary: second\n---\n# Guide\n")
	if IdentityHash(clean) == IdentityHash(multiple) {
		t.Fatal("second leading frontmatter block did not alter identity")
	}
}

func TestIdentityHashTracksUnterminatedBlockRemainder(t *testing.T) {
	first := []byte("---\ndocument_id: doc-1\n")
	second := []byte("---\ndocument_id: doc-1\nsummary: changed\n")
	if IdentityHash(first) == IdentityHash(second) {
		t.Fatal("unterminated frontmatter change did not alter identity")
	}
}
