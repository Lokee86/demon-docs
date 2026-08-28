package codemaprecommend

import "testing"

func TestCoverageAwareSelectionPreservesRoleAndDirectoryWithinScoreBand(t *testing.T) {
	var ranked []rankedSuggestion
	for index := 0; index < DefaultSuggestionLimitPerDocument; index++ {
		ranked = append(ranked, rankedSuggestion{suggestion: Suggestion{
			Link:  Link{Document: "docs/runtime.md", Target: targetName("src/runtime", index)},
			Score: 15.9 - float64(index)/100,
			Role:  SuggestionRolePrimaryImplementation,
		}})
	}
	ranked = append(ranked,
		rankedSuggestion{suggestion: Suggestion{Link: Link{Document: "docs/runtime.md", Target: "tests/runtime_test.go"}, Score: 8.2, Role: SuggestionRoleVerification}},
		rankedSuggestion{suggestion: Suggestion{Link: Link{Document: "docs/runtime.md", Target: "api/runtime.go"}, Score: 8.1, Role: SuggestionRoleInterfaceBoundary}},
	)

	selected := selectRankedSuggestions(ranked)
	if len(selected) != DefaultSuggestionLimitPerDocument {
		t.Fatalf("selected %d suggestions, want %d", len(selected), DefaultSuggestionLimitPerDocument)
	}
	seen := map[string]bool{}
	for _, item := range selected {
		seen[item.Target] = true
	}
	for _, target := range []string{"tests/runtime_test.go", "api/runtime.go"} {
		if !seen[target] {
			t.Fatalf("coverage candidate %s was crowded out: %#v", target, selected)
		}
	}
}

func TestCoverageAwareSelectionDoesNotCrossScoreBands(t *testing.T) {
	var ranked []rankedSuggestion
	for index := 0; index < DefaultSuggestionLimitPerDocument; index++ {
		ranked = append(ranked, rankedSuggestion{suggestion: Suggestion{
			Link:  Link{Document: "docs/runtime.md", Target: targetName("src/runtime", index)},
			Score: 16 + float64(DefaultSuggestionLimitPerDocument-index),
			Role:  SuggestionRolePrimaryImplementation,
		}})
	}
	ranked = append(ranked, rankedSuggestion{suggestion: Suggestion{
		Link:  Link{Document: "docs/runtime.md", Target: "tests/lower_band_test.go"},
		Score: 15.9,
		Role:  SuggestionRoleVerification,
	}})

	selected := selectRankedSuggestions(ranked)
	for _, item := range selected {
		if item.Target == "tests/lower_band_test.go" {
			t.Fatalf("lower score band displaced stronger candidates: %#v", selected)
		}
	}
}

func TestHardLinkSelectionCapsRoleAndDirectoryCoverage(t *testing.T) {
	ranked := []rankedSuggestion{
		hardTest("pkg/runtime_a_test.go", 24),
		hardTest("pkg/runtime_b_test.go", 23),
		hardTest("pkg/runtime_c_test.go", 22),
		hardSymbol("pkg/runtime.go", 21),
		hardDependency("api/contract.go", 20, SuggestionRoleInterfaceBoundary),
		hardDependency("worker/worker.go", 19, SuggestionRoleSupportingImplementation),
	}

	selected := selectRankedSuggestions(ranked)
	hardByRole := map[SuggestionRole]int{}
	hardByDirectory := map[string]int{}
	hardTotal := 0
	for _, item := range selected {
		if item.Tier != SuggestionTierHardLink {
			continue
		}
		hardTotal++
		hardByRole[item.Role]++
		hardByDirectory[coverageDirectory(item.Target)]++
	}
	if hardTotal != HardLinkSuggestionLimitPerDocument {
		t.Fatalf("hard links = %d, want %d: %#v", hardTotal, HardLinkSuggestionLimitPerDocument, selected)
	}
	if hardByRole[SuggestionRoleVerification] != hardLinkRoleLimitPerDocument {
		t.Fatalf("verification hard links = %d", hardByRole[SuggestionRoleVerification])
	}
	if hardByDirectory["pkg"] != hardLinkDirectoryLimitPerDocument {
		t.Fatalf("pkg hard links = %d", hardByDirectory["pkg"])
	}
}

func hardTest(target string, score float64) rankedSuggestion {
	return rankedSuggestion{
		suggestion:         Suggestion{Link: Link{Document: "docs/runtime.md", Target: target}, Score: score, Role: SuggestionRoleVerification},
		hasTestCounterpart: true,
		hasSiblingTarget:   true,
		targetIsTest:       true,
	}
}

func hardSymbol(target string, score float64) rankedSuggestion {
	return rankedSuggestion{
		suggestion:               Suggestion{Link: Link{Document: "docs/runtime.md", Target: target}, Score: score, Role: SuggestionRolePrimaryImplementation},
		hasDeclaredSymbolMention: true,
	}
}

func hardDependency(target string, score float64, role SuggestionRole) rankedSuggestion {
	return rankedSuggestion{
		suggestion:            Suggestion{Link: Link{Document: "docs/runtime.md", Target: target}, Score: score, Role: role},
		hardLinkScore:         score,
		hasDependencyNeighbor: true,
	}
}

func targetName(directory string, index int) string {
	const digits = "0123456789abcdefghijklmnopqrstuvwxyz"
	return directory + "/file_" + string(digits[index/36]) + string(digits[index%36]) + ".go"
}
