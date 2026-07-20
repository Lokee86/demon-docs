package links

import "testing"

func TestParseMarkdownLinksFindsInlineImagesAndReferences(t *testing.T) {
	source := "[doc](docs/a.md#part)\n![image](<assets/a b.png>)\n[asset]: files/data.pdf \"Data\"\n`[code](ignored.md)`\n```md\n[fenced](ignored.md)\n```\n"
	found := parseMarkdownLinks(source)
	if len(found) != 3 {
		t.Fatalf("found %d links, want 3: %#v", len(found), found)
	}
	if found[0].RawPath != "docs/a.md" || found[0].Suffix != "#part" || found[0].Syntax != "inline" {
		t.Fatalf("unexpected inline link: %#v", found[0])
	}
	if found[1].RawPath != "assets/a b.png" || !found[1].Angle {
		t.Fatalf("unexpected image link: %#v", found[1])
	}
	if found[2].RawPath != "files/data.pdf" || found[2].Syntax != "reference" {
		t.Fatalf("unexpected reference definition: %#v", found[2])
	}
}

func TestParseMarkdownLinksIgnoresFrontmatterTargets(t *testing.T) {
	source := "---\nsummary: See [metadata](ignored.md).\nrelated: [[also-ignored]]\n---\n[body](kept.md)\n"
	found := parseMarkdownLinks(source)
	if len(found) != 1 || found[0].RawPath != "kept.md" {
		t.Fatalf("frontmatter links leaked into inventory: %#v", found)
	}
}

func TestResolveLocalTargetRejectsWebURLsButAcceptsAbsolutePaths(t *testing.T) {
	if _, _, local := resolveLocalTarget("https://example.com/a.md", "C:/repo/README.md", false); local {
		t.Fatal("web URL was treated as a local target")
	}
	if _, _, local := resolveLocalTarget("C:/outside/a.md", "C:/repo/README.md", false); !local {
		t.Fatal("absolute filesystem path was not treated as local")
	}
}

func TestParseMarkdownLinksFindsHTMLAndWikiTargets(t *testing.T) {
	source := "<a href=\"docs/guide.md#part\">Guide</a>\n<img src='assets/image.png'>\n[[notes/design|Design]]\n![[assets/diagram.svg]]\n`[[ignored]]`\n"
	found := parseMarkdownLinks(source)
	if len(found) != 4 {
		t.Fatalf("found %d links, want 4: %#v", len(found), found)
	}
	want := []struct {
		path, suffix, syntax string
	}{
		{"docs/guide.md", "#part", "html"},
		{"assets/image.png", "", "html"},
		{"notes/design", "", "wiki"},
		{"assets/diagram.svg", "", "wiki"},
	}
	for i := range want {
		if found[i].RawPath != want[i].path || found[i].Suffix != want[i].suffix || found[i].Syntax != want[i].syntax {
			t.Fatalf("link %d = %#v, want %#v", i, found[i], want[i])
		}
	}
}

func TestParseMarkdownLinksFindsUnquotedHTMLPathWithSlash(t *testing.T) {
	source := "<a href=docs/guide.md>Guide</a>\n"
	found := parseMarkdownLinks(source)
	if len(found) != 1 {
		t.Fatalf("found %d links, want 1: %#v", len(found), found)
	}
	if found[0].RawPath != "docs/guide.md" || found[0].Syntax != "html" {
		t.Fatalf("unexpected HTML link: %#v", found[0])
	}
}

func TestParseMarkdownDocumentReportsUndefinedExplicitReferences(t *testing.T) {
	source := "[known][guide]\n[missing][nope]\n[collapsed][]\n[guide]: docs/guide.md\n`[ignored][missing]`\n"
	parsed := parseMarkdownDocument(source)
	if len(parsed.UndefinedReferences) != 2 {
		t.Fatalf("undefined references = %#v", parsed.UndefinedReferences)
	}
	if parsed.UndefinedReferences[0].Label != "nope" || parsed.UndefinedReferences[1].Label != "collapsed" {
		t.Fatalf("unexpected undefined labels: %#v", parsed.UndefinedReferences)
	}
}
