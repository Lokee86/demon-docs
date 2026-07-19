package codemapbench

import (
	"path"
	"strings"

	"github.com/Lokee86/demon-docs/internal/evidence"
)

type suggestionEvidenceProfile struct {
	hasExactPathMention      bool
	hasUniqueBasenameMention bool
	hasDeclaredSymbolMention bool
	hasDependencyNeighbor    bool
	hasRelatedDocumentTarget bool
}

func rejectIncidentalSuggestionCandidate(candidate evidence.Candidate) bool {
	profile := profileSuggestionEvidence(candidate.Evidence)
	if isDependencyLockfile(candidate.Path) && !profile.hasOwnershipEvidence() {
		return true
	}
	if profile.hasUniqueBasenameMention && !profile.hasExplicitOrOwnershipEvidence() {
		return isNestedContentTarget(candidate.Path) || isWorkflowInfrastructureTarget(candidate.Path)
	}
	return false
}

func profileSuggestionEvidence(items []evidence.Evidence) suggestionEvidenceProfile {
	var profile suggestionEvidenceProfile
	for _, item := range items {
		switch item.Kind {
		case evidence.KindExactPathMention:
			profile.hasExactPathMention = true
		case evidence.KindUniqueBasenameMention:
			profile.hasUniqueBasenameMention = true
		case evidence.KindDeclaredSymbolMention:
			profile.hasDeclaredSymbolMention = true
		case evidence.KindDependencyNeighbor:
			profile.hasDependencyNeighbor = true
		case evidence.KindRelatedDocumentTarget:
			profile.hasRelatedDocumentTarget = true
		}
	}
	return profile
}

func (profile suggestionEvidenceProfile) hasOwnershipEvidence() bool {
	return profile.hasDeclaredSymbolMention ||
		profile.hasDependencyNeighbor ||
		profile.hasRelatedDocumentTarget
}

func (profile suggestionEvidenceProfile) hasExplicitOrOwnershipEvidence() bool {
	return profile.hasExactPathMention || profile.hasOwnershipEvidence()
}

func isDependencyLockfile(value string) bool {
	base := strings.ToLower(path.Base(normalizeSuggestionPath(value)))
	switch base {
	case "bun.lock", "bun.lockb", "cargo.lock", "composer.lock", "flake.lock",
		"gemfile.lock", "go.sum", "npm-shrinkwrap.json", "package-lock.json",
		"pipfile.lock", "pnpm-lock.yaml", "poetry.lock", "uv.lock", "yarn.lock":
		return true
	default:
		return strings.HasSuffix(base, ".lock")
	}
}

func isNestedContentTarget(value string) bool {
	segments := suggestionPathSegments(value)
	for index, segment := range segments {
		switch segment {
		case "asset", "assets", "example", "examples", "fixture", "fixtures", "sample", "samples", "testdata":
			if len(segments)-index-1 >= 2 {
				return true
			}
		}
	}
	return false
}

func isWorkflowInfrastructureTarget(value string) bool {
	segments := suggestionPathSegments(value)
	return len(segments) >= 3 && segments[0] == ".github" && segments[1] == "workflows"
}

func suggestionPathSegments(value string) []string {
	cleaned := strings.Trim(normalizeSuggestionPath(value), "/")
	if cleaned == "" || cleaned == "." {
		return nil
	}
	segments := strings.Split(cleaned, "/")
	for index := range segments {
		segments[index] = strings.ToLower(segments[index])
	}
	return segments
}

func normalizeSuggestionPath(value string) string {
	return path.Clean(strings.ReplaceAll(value, "\\", "/"))
}
