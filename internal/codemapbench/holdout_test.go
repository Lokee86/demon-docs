package codemapbench

import (
	"reflect"
	"testing"
)

func TestSplitHoldoutIsDeterministicAndOrderIndependent(t *testing.T) {
	links := fixtureLinks()
	config := Config{Seed: "stable-seed", HoldoutCount: 2}

	visibleA, hiddenA, seedA, err := splitHoldout(links, config)
	if err != nil {
		t.Fatal(err)
	}

	reversed := append([]Link(nil), links...)
	for left, right := 0, len(reversed)-1; left < right; left, right = left+1, right-1 {
		reversed[left], reversed[right] = reversed[right], reversed[left]
	}
	visibleB, hiddenB, seedB, err := splitHoldout(reversed, config)
	if err != nil {
		t.Fatal(err)
	}

	if seedA != seedB || seedA != "stable-seed" {
		t.Fatalf("unexpected seeds: %q %q", seedA, seedB)
	}
	if !reflect.DeepEqual(visibleA, visibleB) {
		t.Fatalf("visible links changed with input order:\n%#v\n%#v", visibleA, visibleB)
	}
	if !reflect.DeepEqual(hiddenA, hiddenB) {
		t.Fatalf("hidden links changed with input order:\n%#v\n%#v", hiddenA, hiddenB)
	}
	if len(visibleA) != 3 || len(hiddenA) != 2 {
		t.Fatalf("unexpected split sizes: visible=%d hidden=%d", len(visibleA), len(hiddenA))
	}
}

func TestSplitHoldoutNormalizesAndDeduplicatesKnownLinks(t *testing.T) {
	links := []Link{
		{Document: `docs\runtime.md`, Target: `.\server\runtime.go`},
		{Document: "docs/runtime.md", Target: "server/runtime.go"},
		{Document: "docs/runtime.md", Target: "server/other.go"},
	}

	visible, hidden, _, err := splitHoldout(links, Config{HoldoutCount: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(visible)+len(hidden) != 2 {
		t.Fatalf("got %d normalized links, want 2", len(visible)+len(hidden))
	}
}

func TestSplitHoldoutRejectsConflictingSelectors(t *testing.T) {
	_, _, _, err := splitHoldout(fixtureLinks(), Config{
		HoldoutCount:    1,
		HoldoutFraction: 0.2,
	})
	if err == nil {
		t.Fatal("expected conflicting holdout selectors to fail")
	}
}

func TestSplitHoldoutDefaultSelectsAtLeastOneLink(t *testing.T) {
	visible, hidden, seed, err := splitHoldout(fixtureLinks()[:2], Config{})
	if err != nil {
		t.Fatal(err)
	}
	if seed != DefaultSeed {
		t.Fatalf("got seed %q, want %q", seed, DefaultSeed)
	}
	if len(visible) != 1 || len(hidden) != 1 {
		t.Fatalf("unexpected default split: visible=%d hidden=%d", len(visible), len(hidden))
	}
}

func fixtureLinks() []Link {
	return []Link{
		{Document: "docs/a.md", Target: "src/a.go"},
		{Document: "docs/a.md", Target: "src/a_test.go"},
		{Document: "docs/b.md", Target: "src/b.go"},
		{Document: "docs/b.md", Target: "src/b_test.go"},
		{Document: "docs/c.md", Target: "src/c.go"},
	}
}
