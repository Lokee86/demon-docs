package links

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

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
// Markdown file. Its unexported data is populated by NewGeneratedRewrite so
// callers cannot accidentally apply text with a different line-ending style.
type GeneratedRewrite struct {
	SourceFileID       string                `json:"source_file_id"`
	Path               string                `json:"path"`
	ExpectedOldSHA256  string                `json:"expected_old_sha256"`
	ExpectedNewSHA256  string                `json:"expected_new_sha256"`
	Transformations    []LinkTransformation  `json:"transformations"`
	Kind               review.SuggestionKind `json:"kind,omitempty"`
	Selection          review.SelectionMode  `json:"selection,omitempty"`
	OriginSuggestionID string                `json:"origin_suggestion_id,omitempty"`

	oldData []byte
	newData []byte
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
	return GeneratedRewrite{
		SourceFileID:      sourceFileID,
		Path:              filepath.Clean(path),
		ExpectedOldSHA256: sha256Digest(oldData),
		ExpectedNewSHA256: sha256Digest(newData),
		Transformations:   append([]LinkTransformation(nil), transformations...),
		Kind:              review.SuggestionLinkRepair,
		Selection:         review.SelectionAutomatic,
		oldData:           append([]byte(nil), oldData...),
		newData:           append([]byte(nil), newData...),
	}
}

func (r GeneratedRewrite) OldData() []byte { return append([]byte(nil), r.oldData...) }
func (r GeneratedRewrite) NewData() []byte { return append([]byte(nil), r.newData...) }

// ApplyGenerated applies a batch only after every source file's expected old
// hash has been verified. Each write uses a same-directory temporary file and
// an OS-specific atomic replacement, then verifies the resulting new hash.
func ApplyGenerated(rewrites []GeneratedRewrite) ([]Suppression, error) {
	pending := make([]pendingRewrite, 0, len(rewrites))
	seen := make(map[string]struct{}, len(rewrites))
	for index, rewrite := range rewrites {
		if rewrite.Path == "" {
			return nil, fmt.Errorf("generated rewrite %d has an empty path", index)
		}
		path := filepath.Clean(rewrite.Path)
		key := pathKey(path)
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("generated rewrite batch contains duplicate path: %s", path)
		}
		seen[key] = struct{}{}
		if rewrite.ExpectedOldSHA256 == "" || rewrite.ExpectedNewSHA256 == "" {
			return nil, fmt.Errorf("generated rewrite %s is missing an expected hash", path)
		}
		if rewrite.oldData == nil || rewrite.newData == nil {
			return nil, fmt.Errorf("generated rewrite %s was not created by NewGeneratedRewrite", path)
		}
		if actual := sha256Digest(rewrite.oldData); actual != rewrite.ExpectedOldSHA256 {
			return nil, fmt.Errorf("generated rewrite %s has inconsistent old hash: expected %s, got %s", path, rewrite.ExpectedOldSHA256, actual)
		}
		if actual := sha256Digest(rewrite.newData); actual != rewrite.ExpectedNewSHA256 {
			return nil, fmt.Errorf("generated rewrite %s has inconsistent new hash: expected %s, got %s", path, rewrite.ExpectedNewSHA256, actual)
		}
		pending = append(pending, pendingRewrite{rewrite: rewrite, path: path})
	}

	// Preflight every source before allowing any worker to replace a file.
	// This preserves the batch's all-files concurrency check as a barrier.
	preflightErrors := runLinkWorkers(len(pending), func(index int) error {
		return preflightGeneratedRewrite(&pending[index])
	})
	for _, err := range preflightErrors {
		if err != nil {
			return nil, err
		}
	}

	suppressions := make([]Suppression, len(pending))
	applyErrors := runLinkWorkers(len(pending), func(index int) error {
		item := &pending[index]
		if err := replaceGenerated(item.path, item.rewrite.newData, item.mode); err != nil {
			return fmt.Errorf("apply generated rewrite %s: %w", item.path, err)
		}
		current, err := os.ReadFile(item.path)
		if err != nil {
			return fmt.Errorf("verify generated rewrite %s: %w", item.path, err)
		}
		if actual := sha256Digest(current); actual != item.rewrite.ExpectedNewSHA256 {
			return fmt.Errorf("generated rewrite new hash mismatch %s: expected %s, got %s", item.path, item.rewrite.ExpectedNewSHA256, actual)
		}
		suppressions[index] = suppressionFor(item.rewrite)
		return nil
	})
	for _, err := range applyErrors {
		if err != nil {
			return nil, err
		}
	}
	return suppressions, nil
}

type pendingRewrite struct {
	rewrite GeneratedRewrite
	path    string
	mode    fs.FileMode
}

func preflightGeneratedRewrite(item *pendingRewrite) error {
	info, err := os.Stat(item.path)
	if err != nil {
		return fmt.Errorf("stat generated rewrite source %s: %w", item.path, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("generated rewrite source is not a regular file: %s", item.path)
	}
	current, err := os.ReadFile(item.path)
	if err != nil {
		return fmt.Errorf("read generated rewrite source %s: %w", item.path, err)
	}
	if actual := sha256Digest(current); actual != item.rewrite.ExpectedOldSHA256 {
		return fmt.Errorf("generated rewrite source changed before apply %s: expected %s, got %s", item.path, item.rewrite.ExpectedOldSHA256, actual)
	}
	item.mode = info.Mode()
	return nil
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
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func replaceGenerated(path string, data []byte, mode fs.FileMode) (err error) {
	temporary, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".ddocs-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() {
		if temporaryPath != "" {
			_ = os.Remove(temporaryPath)
		}
	}()
	if err := temporary.Chmod(mode.Perm()); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("preserve permissions: %w", err)
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := atomicReplace(temporaryPath, path); err != nil {
		return err
	}
	temporaryPath = ""
	return nil
}
