package review

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"sort"
	"strings"
)

func NewID(prefix string) string {
	var raw [8]byte
	if _, err := rand.Read(raw[:]); err == nil {
		return prefix + "-" + hex.EncodeToString(raw[:])
	}
	return prefix + "-" + stableDigest(prefix)[:16]
}

func LinkSuggestion(sourceFileID, sourcePath, linkID, brokenTarget string, targets []string) Suggestion {
	ordered := append([]string(nil), targets...)
	sort.Strings(ordered)
	relation := stableDigest("link-suggestion", sourceFileID, linkID, brokenTarget)
	fingerprint := stableDigest("link-suggestion-evidence", relation, strings.Join(ordered, "\x00"))
	suggestion := Suggestion{
		ID:           "sg-" + fingerprint[:16],
		Kind:         SuggestionLinkRepair,
		RelationKey:  relation,
		Fingerprint:  fingerprint,
		SourceFileID: sourceFileID,
		SourcePath:   sourcePath,
		LinkID:       linkID,
		BrokenTarget: brokenTarget,
	}
	for index, target := range ordered {
		suggestion.Candidates = append(suggestion.Candidates, Candidate{
			Index:       index + 1,
			Target:      target,
			Fingerprint: stableDigest("link-candidate", relation, target, fingerprint),
		})
	}
	return suggestion
}

func CodemapSuggestion(document, target string, score float64, tier string, evidence []string) Suggestion {
	ordered := append([]string(nil), evidence...)
	sort.Strings(ordered)
	relation := stableDigest("codemap-suggestion", document, target)
	fingerprint := stableDigest("codemap-evidence", relation, tier, strings.Join(ordered, "\x00"))
	return Suggestion{
		ID:          "sg-" + fingerprint[:16],
		Kind:        SuggestionCodemap,
		RelationKey: relation,
		Fingerprint: fingerprint,
		SourcePath:  document,
		Candidates: []Candidate{{
			Index:       1,
			Target:      target,
			Fingerprint: stableDigest("codemap-candidate", relation, fingerprint),
			Score:       score,
			Tier:        tier,
			Evidence:    ordered,
		}},
	}
}

func PathIdentity(path string) string {
	return "path:" + filepath.ToSlash(filepath.Clean(path))
}

func RepairIdentity(sourceFileID, linkID, oldText, newText, targetFileID string) (string, string) {
	relation := stableDigest("repair-relation", sourceFileID, linkID, oldText)
	fingerprint := stableDigest("repair-evidence", relation, newText, targetFileID)
	return relation, fingerprint
}

func TransformationID(relation, fingerprint string) string {
	return "rp-" + stableDigest("repair", relation, fingerprint)[:16]
}

func Digest(data []byte) string {
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func stableDigest(parts ...string) string {
	hash := sha256.New()
	for _, part := range parts {
		_, _ = hash.Write([]byte(part))
		_, _ = hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil))
}
