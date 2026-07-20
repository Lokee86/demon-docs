package links

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"

	"github.com/Lokee86/demon-docs/internal/filetxn"
	"github.com/Lokee86/demon-docs/internal/review"
	"github.com/Lokee86/demon-docs/internal/textio"
)

// LinkTransformation describes one generated Markdown destination change.
// Start and End are byte offsets in the normalized text held by Document.
type LinkTransformation struct {
	LinkID         string `json:"link_id"`
	Start          int    `json:"start"`
	End            int    `json:"end"`
	OldDestination string `json:"old_destination"`
	NewDestination string `json:"new_destination"`
	TargetFileID   string `json:"target_file_id,omitempty"`
	TargetPath     string `json:"target_path,omitempty"`
}

// GeneratedRewrite is a complete, content-addressed rewrite for one source
// Markdown file. Its unexported transaction is populated by the constructors so
// callers cannot apply text with a different line-ending style.
type GeneratedRewrite struct {
	SourceFileID       string                `json:"source_file_id"`
	Path               string                `json:"path"`
	ExpectedOldSHA256  string                `json:"expected_old_sha256"`
	ExpectedNewSHA256  string                `json:"expected_new_sha256"`
	Transformations    []LinkTransformation  `json:"transformations"`
	Kind               review.SuggestionKind `json:"kind,omitempty"`
	Selection          review.SelectionMode  `json:"selection,omitempty"`
	OriginSuggestionID string                `json:"origin_suggestion_id,omitempty"`

	transaction filetxn.Rewrite
}

// Suppression describes a generated write that a watcher can suppress when it
// observes the resulting source-file event. It contains only stable data and
// does not retain in-memory watcher state.
type Suppression struct {
	SourceFileID      string   `json:"source_file_id"`
	Path              string   `json:"path"`
	ExpectedOldSHA256 string   `json:"expected_old_sha256"`
	ExpectedNewSHA256 string   `json:"expected_new_sha256"`
	AffectedLinkIDs   []string `json:"affected_link_ids"`
	OldDestinations   []string `json:"old_destinations"`
	NewDestinations   []string `json:"new_destinations"`
}

// NewGeneratedRewrite builds a rewrite from normalized Markdown text while
// retaining the source document's original newline encoding for both hashes
// and the eventual write.
func NewGeneratedRewrite(sourceFileID, path string, document textio.Document, transformations []LinkTransformation) (GeneratedRewrite, error) {
	if sourceFileID == "" {
		return GeneratedRewrite{}, errors.New("generated rewrite source file ID is empty")
	}
	if path == "" {
		return GeneratedRewrite{}, errors.New("generated rewrite path is empty")
	}

	oldData := document.Encode(document.Text)
	newText, err := rewriteText(document.Text, transformations)
	if err != nil {
		return GeneratedRewrite{}, fmt.Errorf("build generated rewrite for %s: %w", path, err)
	}
	newData := document.Encode(newText)
	return newGeneratedRewriteBytes(sourceFileID, path, oldData, newData, transformations), nil
}

// NewGeneratedRewriteBytes constructs a content-addressed rewrite from exact
// before and after bytes. It is used by review-driven repairs and undo while
// retaining the same hash guards and atomic replacement path as link repair.
func NewGeneratedRewriteBytes(sourceFileID, path string, oldData, newData []byte, transformations []LinkTransformation) (GeneratedRewrite, error) {
	if sourceFileID == "" {
		return GeneratedRewrite{}, errors.New("generated rewrite source file ID is empty")
	}
	if path == "" {
		return GeneratedRewrite{}, errors.New("generated rewrite path is empty")
	}
	return newGeneratedRewriteBytes(sourceFileID, path, oldData, newData, transformations), nil
}

func newGeneratedRewriteBytes(sourceFileID, path string, oldData, newData []byte, transformations []LinkTransformation) GeneratedRewrite {
	transaction := filetxn.New(path, oldData, newData)
	return GeneratedRewrite{
		SourceFileID:      sourceFileID,
		Path:              transaction.Path(),
		ExpectedOldSHA256: transaction.ExpectedOldSHA256(),
		ExpectedNewSHA256: transaction.ExpectedNewSHA256(),
		Transformations:   append([]LinkTransformation(nil), transformations...),
		Kind:              review.SuggestionLinkRepair,
		Selection:         review.SelectionAutomatic,
		transaction:       transaction,
	}
}

func (rewrite GeneratedRewrite) OldData() []byte { return rewrite.transaction.OldData() }
func (rewrite GeneratedRewrite) NewData() []byte { return rewrite.transaction.NewData() }

// ApplyGenerated delegates filesystem safety to the shared content-addressed
// transaction while retaining link-specific suppression metadata.
func ApplyGenerated(rewrites []GeneratedRewrite) ([]Suppression, error) {
	transactions := make([]filetxn.Rewrite, len(rewrites))
	for index, rewrite := range rewrites {
		if rewrite.Path == "" {
			return nil, fmt.Errorf("generated rewrite %d has an empty path", index)
		}
		if !rewrite.transaction.Prepared() {
			return nil, fmt.Errorf("generated rewrite %s was not created by NewGeneratedRewrite", rewrite.Path)
		}
		if filepath.Clean(rewrite.Path) != rewrite.transaction.Path() ||
			rewrite.ExpectedOldSHA256 != rewrite.transaction.ExpectedOldSHA256() ||
			rewrite.ExpectedNewSHA256 != rewrite.transaction.ExpectedNewSHA256() {
			return nil, fmt.Errorf("generated rewrite %s does not match its prepared transaction", rewrite.Path)
		}
		transactions[index] = rewrite.transaction
	}

	applied, err := filetxn.Apply(transactions)
	if err != nil {
		return nil, err
	}
	suppressions := make([]Suppression, len(applied))
	for index := range applied {
		suppressions[index] = suppressionFor(rewrites[index])
	}
	return suppressions, nil
}

func rewriteText(source string, transformations []LinkTransformation) (string, error) {
	ordered := append([]LinkTransformation(nil), transformations...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].Start < ordered[j].Start
	})
	lastEnd := 0
	for index, transformation := range ordered {
		if transformation.Start < 0 || transformation.End < transformation.Start || transformation.End > len(source) {
			return "", fmt.Errorf("transformation %d has invalid range [%d:%d]", index, transformation.Start, transformation.End)
		}
		if transformation.Start < lastEnd {
			return "", fmt.Errorf("transformation %d overlaps a previous transformation", index)
		}
		if got := source[transformation.Start:transformation.End]; got != transformation.OldDestination {
			return "", fmt.Errorf("transformation %d old destination mismatch: source has %q, want %q", index, got, transformation.OldDestination)
		}
		lastEnd = transformation.End
	}

	result := source
	for index := len(ordered) - 1; index >= 0; index-- {
		transformation := ordered[index]
		result = result[:transformation.Start] + transformation.NewDestination + result[transformation.End:]
	}
	return result, nil
}

func suppressionFor(rewrite GeneratedRewrite) Suppression {
	suppression := Suppression{
		SourceFileID:      rewrite.SourceFileID,
		Path:              rewrite.Path,
		ExpectedOldSHA256: rewrite.ExpectedOldSHA256,
		ExpectedNewSHA256: rewrite.ExpectedNewSHA256,
		AffectedLinkIDs:   make([]string, 0, len(rewrite.Transformations)),
		OldDestinations:   make([]string, 0, len(rewrite.Transformations)),
		NewDestinations:   make([]string, 0, len(rewrite.Transformations)),
	}
	for _, transformation := range rewrite.Transformations {
		suppression.AffectedLinkIDs = append(suppression.AffectedLinkIDs, transformation.LinkID)
		suppression.OldDestinations = append(suppression.OldDestinations, transformation.OldDestination)
		suppression.NewDestinations = append(suppression.NewDestinations, transformation.NewDestination)
	}
	return suppression
}

func sha256Digest(data []byte) string {
	return filetxn.Digest(data)
}

func replaceGenerated(path string, data []byte, mode fs.FileMode) error {
	return filetxn.Replace(path, data, mode)
}
