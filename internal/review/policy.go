package review

import (
	"errors"
	"sort"

	git "github.com/go-git/go-git/v5"
)

type MatchState string

const (
	MatchNone   MatchState = ""
	MatchActive MatchState = "active"
	MatchStale  MatchState = "stale"
)

type Policy struct {
	issues     map[string]Decision
	candidates map[string]Decision
	repairs    map[string]Decision
}

func LoadPolicy(repositoryRoot string) (Policy, error) {
	policy := Policy{
		issues:     make(map[string]Decision),
		candidates: make(map[string]Decision),
		repairs:    make(map[string]Decision),
	}
	store, err := Open(repositoryRoot)
	if errors.Is(err, git.ErrRepositoryNotExists) {
		return policy, nil
	}
	if err != nil {
		return policy, err
	}
	history, err := store.History(0)
	if err != nil {
		return policy, err
	}
	for index := len(history) - 1; index >= 0; index-- {
		decision := history[index].Decision
		if decision == nil {
			continue
		}
		switch decision.Action {
		case DecisionDeclineIssue:
			policy.issues[decision.RelationKey] = *decision
		case DecisionDeclineCandidate:
			policy.candidates[candidateKey(decision.RelationKey, decision.CandidateTarget)] = *decision
		case DecisionReconsider:
			delete(policy.issues, decision.RelationKey)
			prefix := decision.RelationKey + "\x00"
			for key := range policy.candidates {
				if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
					delete(policy.candidates, key)
				}
			}
		case DecisionBlockRepair:
			policy.repairs[decision.RelationKey] = *decision
		case DecisionUnblockRepair:
			delete(policy.repairs, decision.RelationKey)
		}
	}
	return policy, nil
}

func (p Policy) ApplySuggestion(suggestion Suggestion) Suggestion {
	if decision, ok := p.issues[suggestion.RelationKey]; ok {
		suggestion.Reason = decision.Reason
		suggestion.DecisionTime = decision.DecidedAt
		if decision.Fingerprint == suggestion.Fingerprint {
			suggestion.Status = StatusDeclined
		} else {
			suggestion.Status = StatusStale
		}
	}
	for index := range suggestion.Candidates {
		candidate := &suggestion.Candidates[index]
		decision, ok := p.candidates[candidateKey(suggestion.RelationKey, candidate.Target)]
		if !ok {
			continue
		}
		if decision.CandidateFingerprint == candidate.Fingerprint {
			candidate.Declined = true
		} else {
			candidate.Stale = true
		}
	}
	if suggestion.Status == "" {
		suggestion.Status = StatusPending
	}
	return suggestion
}

func (p Policy) Repair(relationKey, fingerprint string) (MatchState, Decision) {
	decision, ok := p.repairs[relationKey]
	if !ok {
		return MatchNone, Decision{}
	}
	if decision.Fingerprint == fingerprint {
		return MatchActive, decision
	}
	return MatchStale, decision
}

func (p Policy) DeclinedSuggestions() []Decision {
	result := make([]Decision, 0, len(p.issues)+len(p.candidates))
	for _, decision := range p.issues {
		result = append(result, decision)
	}
	for _, decision := range p.candidates {
		result = append(result, decision)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].DecidedAt.After(result[j].DecidedAt) })
	return result
}

func candidateKey(relationKey, target string) string {
	return relationKey + "\x00" + target
}
