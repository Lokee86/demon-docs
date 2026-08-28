package codemaprecommend

import (
	"math"
	"path"
	"sort"
)

type coverageState struct {
	roles       map[SuggestionRole]int
	directories map[string]int
	pairs       map[string]int
}

func newCoverageState() coverageState {
	return coverageState{
		roles:       map[SuggestionRole]int{},
		directories: map[string]int{},
		pairs:       map[string]int{},
	}
}

func selectRankedSuggestions(ranked []rankedSuggestion) []Suggestion {
	sortRankedSuggestions(ranked)
	selected := selectCoverageAware(ranked, min(DefaultSuggestionLimitPerDocument, len(ranked)))
	addRepeatedMentionReserve(selected, ranked)

	ordered := make([]rankedSuggestion, 0, len(selected))
	for _, item := range selected {
		ordered = append(ordered, item)
	}
	sortRankedSuggestions(ordered)

	hardLinks := selectHardLinkTargets(ordered)
	result := make([]Suggestion, 0, len(ordered))
	for _, item := range ordered {
		suggestion := item.suggestion
		suggestion.Tier = SuggestionTierContext
		if _, ok := hardLinks[suggestion.Target]; ok {
			suggestion.Tier = SuggestionTierHardLink
		}
		result = append(result, suggestion)
	}
	return result
}

func selectCoverageAware(ranked []rankedSuggestion, limit int) map[string]rankedSuggestion {
	selected := make(map[string]rankedSuggestion, limit)
	state := newCoverageState()
	for start := 0; start < len(ranked) && len(selected) < limit; {
		band := scoreBand(ranked[start].suggestion.Score)
		end := start + 1
		for end < len(ranked) && scoreBand(ranked[end].suggestion.Score) == band {
			end++
		}
		remaining := append([]rankedSuggestion(nil), ranked[start:end]...)
		for len(remaining) > 0 && len(selected) < limit {
			index := bestCoverageCandidate(remaining, state)
			item := remaining[index]
			selected[item.suggestion.Target] = item
			state.add(item)
			remaining = append(remaining[:index], remaining[index+1:]...)
		}
		start = end
	}
	return selected
}

func bestCoverageCandidate(items []rankedSuggestion, state coverageState) int {
	best := 0
	for index := 1; index < len(items); index++ {
		if coverageLess(items[index], items[best], state) {
			best = index
		}
	}
	return best
}

func coverageLess(left, right rankedSuggestion, state coverageState) bool {
	leftClass := state.class(left)
	rightClass := state.class(right)
	if leftClass != rightClass {
		return leftClass < rightClass
	}
	if left.suggestion.Score != right.suggestion.Score {
		return left.suggestion.Score > right.suggestion.Score
	}
	return left.suggestion.Target < right.suggestion.Target
}

func (state coverageState) class(item rankedSuggestion) int {
	role := item.suggestion.Role
	directory := coverageDirectory(item.suggestion.Target)
	meaningful := role != "" && role != SuggestionRoleContext
	roleNew := meaningful && state.roles[role] == 0
	directoryNew := state.directories[directory] == 0
	pairNew := meaningful && state.pairs[coveragePair(role, directory)] == 0
	switch {
	case roleNew && directoryNew:
		return 0
	case roleNew:
		return 1
	case meaningful && directoryNew:
		return 2
	case pairNew:
		return 3
	case !meaningful && directoryNew:
		return 4
	default:
		return 5
	}
}

func (state coverageState) add(item rankedSuggestion) {
	role := item.suggestion.Role
	directory := coverageDirectory(item.suggestion.Target)
	state.directories[directory]++
	if role == "" || role == SuggestionRoleContext {
		return
	}
	state.roles[role]++
	state.pairs[coveragePair(role, directory)]++
}

func addRepeatedMentionReserve(selected map[string]rankedSuggestion, ranked []rankedSuggestion) {
	repeated := append([]rankedSuggestion(nil), ranked...)
	sort.Slice(repeated, func(i, j int) bool {
		if repeated[i].repeatedMentionCount != repeated[j].repeatedMentionCount {
			return repeated[i].repeatedMentionCount > repeated[j].repeatedMentionCount
		}
		if repeated[i].suggestion.Score != repeated[j].suggestion.Score {
			return repeated[i].suggestion.Score > repeated[j].suggestion.Score
		}
		return repeated[i].suggestion.Target < repeated[j].suggestion.Target
	})
	reserved := 0
	for _, item := range repeated {
		if item.repeatedMentionCount < RepeatedMentionMinimumCount || reserved >= RepeatedMentionReservePerDocument {
			break
		}
		if _, exists := selected[item.suggestion.Target]; exists {
			continue
		}
		selected[item.suggestion.Target] = item
		reserved++
	}
}

func sortRankedSuggestions(items []rankedSuggestion) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].suggestion.Score != items[j].suggestion.Score {
			return items[i].suggestion.Score > items[j].suggestion.Score
		}
		return items[i].suggestion.Target < items[j].suggestion.Target
	})
}

func scoreBand(score float64) int {
	if score <= 0 {
		return -1 << 30
	}
	return int(math.Floor(math.Log2(score)))
}

func coverageDirectory(target string) string {
	directory := path.Dir(target)
	if directory == "." || directory == "/" {
		return "."
	}
	return directory
}

func coveragePair(role SuggestionRole, directory string) string {
	return string(role) + "\x00" + directory
}
