package codemaprecommend

const (
	hardLinkRoleLimitPerDocument      = 2
	hardLinkDirectoryLimitPerDocument = 3
)

func selectHardLinkTargets(items []rankedSuggestion) map[string]struct{} {
	eligible := make([]rankedSuggestion, 0, len(items))
	for _, item := range items {
		if item.suggestion.Role == "" || item.suggestion.Role == SuggestionRoleContext || !item.isHardLinkCandidate() {
			continue
		}
		eligible = append(eligible, item)
	}
	sortRankedSuggestions(eligible)

	selected := make(map[string]struct{}, HardLinkSuggestionLimitPerDocument)
	state := newCoverageState()
	for start := 0; start < len(eligible) && len(selected) < HardLinkSuggestionLimitPerDocument; {
		band := scoreBand(eligible[start].suggestion.Score)
		end := start + 1
		for end < len(eligible) && scoreBand(eligible[end].suggestion.Score) == band {
			end++
		}
		remaining := append([]rankedSuggestion(nil), eligible[start:end]...)
		for len(remaining) > 0 && len(selected) < HardLinkSuggestionLimitPerDocument {
			index := bestCoverageCandidate(remaining, state)
			item := remaining[index]
			remaining = append(remaining[:index], remaining[index+1:]...)
			role := item.suggestion.Role
			directory := coverageDirectory(item.suggestion.Target)
			if state.roles[role] >= hardLinkRoleLimitPerDocument || state.directories[directory] >= hardLinkDirectoryLimitPerDocument {
				continue
			}
			selected[item.suggestion.Target] = struct{}{}
			state.add(item)
		}
		start = end
	}
	return selected
}
