package links

import (
	"path/filepath"
	"strings"
)

// rankPathAwareCandidates narrows repeated-basename matches only when the
// missing path carries additional directory evidence. A basename-only match
// remains ambiguous, and equal path scores remain ambiguous.
func rankPathAwareCandidates(root, missingPath string, candidates []string) []string {
	if len(candidates) <= 1 {
		return candidates
	}

	missing, ok := repositoryPathSegments(root, missingPath)
	if !ok {
		return candidates
	}

	type scoredCandidate struct {
		path     string
		segments []string
		suffix   int
		distance int
	}

	scored := make([]scoredCandidate, 0, len(candidates))
	maxSuffix := 0
	for _, candidate := range candidates {
		segments, candidateOK := repositoryPathSegments(root, candidate)
		if !candidateOK {
			continue
		}
		suffix := commonPathSuffix(missing, segments)
		if suffix > maxSuffix {
			maxSuffix = suffix
		}
		scored = append(scored, scoredCandidate{
			path:     candidate,
			segments: segments,
			suffix:   suffix,
		})
	}

	// The basename is always one matching segment. Require at least one
	// matching parent directory before using path shape as repair evidence.
	if maxSuffix < 2 {
		return candidates
	}

	strongest := scored[:0]
	for _, candidate := range scored {
		if candidate.suffix == maxSuffix {
			candidate.distance = pathSegmentDistance(missing, candidate.segments)
			strongest = append(strongest, candidate)
		}
	}
	if len(strongest) == 1 {
		return []string{strongest[0].path}
	}

	minimumDistance := strongest[0].distance
	for _, candidate := range strongest[1:] {
		if candidate.distance < minimumDistance {
			minimumDistance = candidate.distance
		}
	}

	result := make([]string, 0, len(strongest))
	for _, candidate := range strongest {
		if candidate.distance == minimumDistance {
			result = append(result, candidate.path)
		}
	}
	return result
}

func repositoryPathSegments(root, path string) ([]string, bool) {
	relative, err := filepath.Rel(root, path)
	if err != nil || relative == "." || filepath.IsAbs(relative) || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return nil, false
	}
	return strings.Split(filepath.Clean(relative), string(filepath.Separator)), true
}

func commonPathSuffix(left, right []string) int {
	matched := 0
	for leftIndex, rightIndex := len(left)-1, len(right)-1; leftIndex >= 0 && rightIndex >= 0; leftIndex, rightIndex = leftIndex-1, rightIndex-1 {
		if !strings.EqualFold(left[leftIndex], right[rightIndex]) {
			break
		}
		matched++
	}
	return matched
}

func pathSegmentDistance(left, right []string) int {
	previous := make([]int, len(right)+1)
	for index := range previous {
		previous[index] = index
	}

	for leftIndex, leftSegment := range left {
		current := make([]int, len(right)+1)
		current[0] = leftIndex + 1
		for rightIndex, rightSegment := range right {
			cost := 1
			if strings.EqualFold(leftSegment, rightSegment) {
				cost = 0
			}
			deletion := previous[rightIndex+1] + 1
			insertion := current[rightIndex] + 1
			substitution := previous[rightIndex] + cost
			current[rightIndex+1] = minInt(deletion, insertion, substitution)
		}
		previous = current
	}
	return previous[len(right)]
}

func minInt(values ...int) int {
	minimum := values[0]
	for _, value := range values[1:] {
		if value < minimum {
			minimum = value
		}
	}
	return minimum
}
