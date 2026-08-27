package codemaprecommend

// Link is one authored relationship from a document to a code target.
type Link struct {
	Document string `json:"document"`
	Target   string `json:"target"`
}

type SuggestionTier string

const (
	SuggestionTierHardLink SuggestionTier = "hard_link"
	SuggestionTierContext  SuggestionTier = "context"
)

type SuggestionRole string

const (
	SuggestionRolePrimaryImplementation    SuggestionRole = "primary_implementation"
	SuggestionRoleSupportingImplementation SuggestionRole = "supporting_implementation"
	SuggestionRoleVerification             SuggestionRole = "verification_test"
	SuggestionRoleInterfaceBoundary        SuggestionRole = "interface_boundary"
	SuggestionRoleContext                  SuggestionRole = "context_only"
)

// Valid accepts current roles and the empty legacy value used by reports
// produced before role classification was added.
func (role SuggestionRole) Valid() bool {
	switch role {
	case "", SuggestionRolePrimaryImplementation, SuggestionRoleSupportingImplementation,
		SuggestionRoleVerification, SuggestionRoleInterfaceBoundary, SuggestionRoleContext:
		return true
	default:
		return false
	}
}

// Valid accepts the two current tiers and the empty legacy value. Consumers
// normalize an empty tier to context when evaluating older schema-1 reports.
func (tier SuggestionTier) Valid() bool {
	return tier == "" || tier == SuggestionTierHardLink || tier == SuggestionTierContext
}

// Suggestion is one candidate missing link produced by an evidence source.
// Tier separates link-worthy recommendations from weaker relationships that
// remain useful when assembling bounded agent context.
type Suggestion struct {
	Link
	Score    float64        `json:"score,omitempty"`
	Evidence []string       `json:"evidence,omitempty"`
	Tier     SuggestionTier `json:"tier,omitempty"`
	Role     SuggestionRole `json:"role,omitempty"`
}
