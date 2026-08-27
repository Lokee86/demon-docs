package evidence

func evidenceExpansionTargets(input Input) []string {
	return visibleTargetsByKind(input, AuthoredTargetFile)
}

func semanticRelationshipTargets(input Input) []string {
	return visibleTargetsByKind(input, AuthoredTargetFile, AuthoredTargetSymbol)
}

func visibleTargetsByKind(input Input, kinds ...AuthoredTargetKind) []string {
	if len(input.AuthoredTargets) == 0 {
		return append([]string(nil), input.ExistingTargets...)
	}
	allowed := make(map[AuthoredTargetKind]struct{}, len(kinds))
	for _, kind := range kinds {
		allowed[kind] = struct{}{}
	}
	visible := normalizedSet(input.ExistingTargets)
	result := map[string]struct{}{}
	for _, authored := range input.AuthoredTargets {
		if _, ok := allowed[authored.Kind]; !ok {
			continue
		}
		for _, target := range authored.ResolvedTargets {
			target = normalizePath(target)
			if target == "" {
				continue
			}
			if _, ok := visible[target]; ok {
				result[target] = struct{}{}
			}
		}
	}
	return sortedKeys(result)
}
