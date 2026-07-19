package links

import (
	"fmt"
	"strings"

	"github.com/Lokee86/demon-docs/internal/review"
)

func reviewRelationToken(record LinkRecord) string {
	return fmt.Sprintf("%d\x00%s", record.Ordinal, record.Syntax)
}

func reviewRepairPolicy(policy review.Policy, sourceFileID, sourcePath string, record LinkRecord, oldText, newText, targetFileID string) (review.MatchState, review.Decision) {
	token := reviewRelationToken(record)
	relation, fingerprint := review.RepairIdentity(sourceFileID, token, oldText, newText, targetFileID)
	if state, decision := policy.Repair(relation, fingerprint); state != review.MatchNone {
		return state, decision
	}
	pathRelation, pathFingerprint := review.RepairIdentity(review.PathIdentity(sourcePath), token, oldText, newText, targetFileID)
	return policy.Repair(pathRelation, pathFingerprint)
}

func reviewReason(reason string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return ""
	}
	return "; reason: " + reason
}
