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

func TestResolveLocalTargetRejectsWebURLsButAcceptsAbsolutePaths(t *testing.T) {
	if _, _, local := resolveLocalTarget("https://example.com/a.md", "C:/repo/README.md", false); local {
		t.Fatal("web URL was treated as a local target")
	}
	if _, _, local := resolveLocalTarget("C:/outside/a.md", "C:/repo/README.md", false); !local {
		t.Fatal("absolute filesystem path was not treated as local")
	}
}
