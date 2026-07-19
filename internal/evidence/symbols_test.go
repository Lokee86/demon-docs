package evidence

import "testing"

func TestCollectDeclaredSymbolsResolvesUniqueQualifiedAndClassMentions(t *testing.T) {
	input := Input{
		DocumentPath:    "docs/runtime.md",
		DocumentText:    "Runtime.LoadStats delegates to the store. ToolingPacketRouter handles results.",
		RepositoryFiles: []string{"server/runtime.go", "client/tooling_packet_router.gd"},
		SymbolDeclarations: []SymbolDeclaration{
			{Path: "server/runtime.go", Symbol: "Runtime.LoadStats"},
			{Path: "client/tooling_packet_router.gd", Symbol: "ToolingPacketRouter"},
		},
	}

	assertKind(t, findCandidate(t, Collect(input), "server/runtime.go"), KindDeclaredSymbolMention)
	assertKind(t, findCandidate(t, Collect(input), "client/tooling_packet_router.gd"), KindDeclaredSymbolMention)
}

func TestCollectDeclaredSymbolsRejectsAmbiguousSymbols(t *testing.T) {
	input := Input{
		DocumentPath:    "docs/runtime.md",
		DocumentText:    "Runtime owns the operation.",
		RepositoryFiles: []string{"server/a/runtime.go", "server/b/runtime.go"},
		SymbolDeclarations: []SymbolDeclaration{
			{Path: "server/a/runtime.go", Symbol: "Runtime"},
			{Path: "server/b/runtime.go", Symbol: "Runtime"},
		},
	}

	if candidates := Collect(input); len(candidates) != 0 {
		t.Fatalf("ambiguous symbol produced candidates: %#v", candidates)
	}
}
