package evidence

import "testing"

func TestCollectMentionsCountsRepeatedExplicitPaths(t *testing.T) {
	input := Input{
		DocumentPath:    "docs/runtime.md",
		DocumentText:    "`shared/packets/realtime.toml` owns the contract. Update `shared/packets/realtime.toml` first.",
		RepositoryFiles: []string{"shared/packets/realtime.toml"},
	}
	candidate := findCandidate(t, Collect(input), "shared/packets/realtime.toml")
	for _, item := range candidate.Evidence {
		if item.Kind == KindExactPathMention && item.Count == 2 {
			return
		}
	}
	t.Fatalf("unexpected repeated mention evidence: %#v", candidate.Evidence)
}
