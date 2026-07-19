package review

import "time"

const SchemaVersion = 1

type SuggestionKind string

const (
	SuggestionLinkRepair SuggestionKind = "link_repair"
	SuggestionCodemap    SuggestionKind = "codemap_link"
)

type SuggestionStatus string

const (
	StatusPending  SuggestionStatus = "pending"
	StatusDeclined SuggestionStatus = "declined"
	StatusBlocked  SuggestionStatus = "blocked"
	StatusStale    SuggestionStatus = "stale"
)

type Candidate struct {
	Index       int      `json:"index"`
	Target      string   `json:"target"`
	Fingerprint string   `json:"fingerprint"`
	Score       float64  `json:"score,omitempty"`
	Tier        string   `json:"tier,omitempty"`
	Evidence    []string `json:"evidence,omitempty"`
	Declined    bool     `json:"declined,omitempty"`
	Stale       bool     `json:"stale,omitempty"`
}

type Suggestion struct {
	ID           string           `json:"id"`
	Kind         SuggestionKind   `json:"kind"`
	RelationKey  string           `json:"relation_key"`
	Fingerprint  string           `json:"fingerprint"`
	SourceFileID string           `json:"source_file_id,omitempty"`
	SourcePath   string           `json:"source_path"`
	LinkID       string           `json:"link_id,omitempty"`
	Line         int              `json:"line,omitempty"`
	Column       int              `json:"column,omitempty"`
	BrokenTarget string           `json:"broken_target,omitempty"`
	Candidates   []Candidate      `json:"candidates"`
	Status       SuggestionStatus `json:"status,omitempty"`
	Reason       string           `json:"reason,omitempty"`
	DecisionTime time.Time        `json:"decision_time,omitempty"`
}

type SelectionMode string

const (
	SelectionAutomatic SelectionMode = "deterministic"
	SelectionUser      SelectionMode = "user"
	SelectionUndo      SelectionMode = "undo"
)

type RelatedFile struct {
	FileID string `json:"file_id,omitempty"`
	Path   string `json:"path"`
}

type Transformation struct {
	ID            string `json:"id"`
	LinkID        string `json:"link_id,omitempty"`
	Start         int    `json:"start"`
	End           int    `json:"end"`
	OldText       string `json:"old_text"`
	NewText       string `json:"new_text"`
	RelationKey   string `json:"relation_key"`
	RelationToken string `json:"relation_token,omitempty"`
	Fingerprint   string `json:"fingerprint"`
	TargetFileID  string `json:"target_file_id,omitempty"`
	TargetPath    string `json:"target_path,omitempty"`
}

type Change struct {
	ID                 string           `json:"id"`
	RunID              string           `json:"run_id"`
	Kind               SuggestionKind   `json:"kind"`
	Selection          SelectionMode    `json:"selection"`
	OriginSuggestionID string           `json:"origin_suggestion_id,omitempty"`
	SourceFileID       string           `json:"source_file_id"`
	SourcePath         string           `json:"source_path"`
	BeforeSHA256       string           `json:"before_sha256"`
	AfterSHA256        string           `json:"after_sha256"`
	Transformations    []Transformation `json:"transformations"`
	Related            []RelatedFile    `json:"related,omitempty"`
	UndoOf             string           `json:"undo_of,omitempty"`
	UndoRepairID       string           `json:"undo_repair_id,omitempty"`
	AppliedAt          time.Time        `json:"applied_at"`
}

type DecisionAction string

const (
	DecisionDeclineIssue     DecisionAction = "decline_issue"
	DecisionDeclineCandidate DecisionAction = "decline_candidate"
	DecisionReconsider       DecisionAction = "reconsider"
	DecisionBlockRepair      DecisionAction = "block_repair"
	DecisionUnblockRepair    DecisionAction = "unblock_repair"
)

type Decision struct {
	ID                   string         `json:"id"`
	Action               DecisionAction `json:"action"`
	RelationKey          string         `json:"relation_key"`
	Fingerprint          string         `json:"fingerprint,omitempty"`
	CandidateTarget      string         `json:"candidate_target,omitempty"`
	CandidateFingerprint string         `json:"candidate_fingerprint,omitempty"`
	SuggestionID         string         `json:"suggestion_id,omitempty"`
	ChangeID             string         `json:"change_id,omitempty"`
	Reason               string         `json:"reason,omitempty"`
	Suggestion           *Suggestion    `json:"suggestion,omitempty"`
	DecidedAt            time.Time      `json:"decided_at"`
}

type EventType string

const (
	EventChange   EventType = "change"
	EventDecision EventType = "decision"
)

type Event struct {
	SchemaVersion int       `json:"schema_version"`
	ID            string    `json:"id"`
	Type          EventType `json:"type"`
	Time          time.Time `json:"time"`
	Change        *Change   `json:"change,omitempty"`
	Decision      *Decision `json:"decision,omitempty"`
}

type StoredEvent struct {
	Event
	CommitHash string `json:"commit_hash"`
	Before     []byte `json:"-"`
	After      []byte `json:"-"`
}
