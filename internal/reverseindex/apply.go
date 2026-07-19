package reverseindex

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Lokee86/demon-docs/internal/repository"
	"github.com/Lokee86/demon-docs/internal/textio"
)

func Apply(repositoryRoot string, plan Plan) (int, error) {
	changed := 0
	for _, update := range plan.Updates {
		if !repository.Contains(repositoryRoot, update.Path) {
			return changed, fmt.Errorf("refusing to write outside repository root: %s", update.Path)
		}
		if err := os.MkdirAll(filepath.Dir(update.Path), 0o755); err != nil {
			return changed, err
		}
		data := textio.EncodeNew(update.NewText)
		if update.OldText != nil {
			doc, err := textio.Read(update.Path)
			if err != nil {
				return changed, err
			}
			data = doc.Encode(update.NewText)
		}
		if err := os.WriteFile(update.Path, data, 0o644); err != nil {
			return changed, err
		}
		changed++
	}
	return changed, nil
}
