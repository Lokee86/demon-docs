package codemaprecommend

import (
	"strings"

	"github.com/Lokee86/demon-docs/internal/evidence"
)

// classifySuggestionRole assigns one deterministic documentation role without
// changing candidate score, ordering, or mutation eligibility. Step 7 may use
// these roles for coverage-aware selection; this phase only publishes them.
func classifySuggestionRole(target string, items []evidence.Evidence) SuggestionRole {
	if IsTestTarget(target) || hasVerificationRelationship(items) {
		return SuggestionRoleVerification
	}
	if hasBoundaryRelationship(items) {
		return SuggestionRoleInterfaceBoundary
	}
	if hasDirectImplementationEvidence(items) {
		return SuggestionRolePrimaryImplementation
	}
	if hasSupportingImplementationEvidence(items) {
		return SuggestionRoleSupportingImplementation
	}
	return SuggestionRoleContext
}

func hasVerificationRelationship(items []evidence.Evidence) bool {
	for _, item := range items {
		if item.Kind == evidence.KindSemanticRelationship && item.Detail == "inbound:tests" {
			return true
		}
	}
	return false
}

func hasBoundaryRelationship(items []evidence.Evidence) bool {
	for _, item := range items {
		if item.Kind != evidence.KindSemanticRelationship {
			continue
		}
		direction, relation, ok := semanticRelationship(item.Detail)
		if !ok || direction != "outbound" {
			continue
		}
		switch relation {
		case "implements", "extends", "overrides", "uses-trait", "includes":
			return true
		}
	}
	return false
}

func hasDirectImplementationEvidence(items []evidence.Evidence) bool {
	for _, item := range items {
		switch item.Kind {
		case evidence.KindDeclaredSymbolMention,
			evidence.KindExactPathMention,
			evidence.KindUniqueBasenameMention:
			return true
		}
	}
	return false
}

func hasSupportingImplementationEvidence(items []evidence.Evidence) bool {
	for _, item := range items {
		switch item.Kind {
		case evidence.KindTestCounterpart,
			evidence.KindDependencyNeighbor,
			evidence.KindSemanticRelationship,
			evidence.KindSiblingTarget,
			evidence.KindRelatedDocumentTarget:
			return true
		}
	}
	return false
}

func semanticRelationship(detail string) (direction, relation string, ok bool) {
	direction, relation, ok = strings.Cut(strings.TrimSpace(detail), ":")
	if !ok || relation == "" || (direction != "incoming" && direction != "outbound") {
		return "", "", false
	}
	return direction, relation, true
}
