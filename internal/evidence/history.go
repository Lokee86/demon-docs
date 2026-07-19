package evidence

func (c *collector) collectHistory(document string, existingTargets []string, commits []Commit) {
	seeds := normalizedSet(existingTargets)
	for _, commit := range commits {
		paths := normalizedSet(commit.Paths)
		_, documentChanged := paths[document]
		if documentChanged {
			for candidate := range paths {
				c.add(candidate, KindGitDocumentCoChange, document, "", 1)
			}
		}
		for seed := range seeds {
			if _, changed := paths[seed]; !changed {
				continue
			}
			for candidate := range paths {
				c.add(candidate, KindGitTargetCoChange, seed, "", 1)
			}
		}
	}
}
