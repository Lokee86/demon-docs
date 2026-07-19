package review

import (
	"fmt"
	"sort"
	"time"

	"github.com/Lokee86/demon-docs/internal/textio"
)

func UndoEligible(history []StoredEvent, changeID string, depth, maxAgeDays int, now time.Time) error {
	if depth == 0 {
		return fmt.Errorf("undo is disabled by review.undo_depth")
	}
	position := 0
	var target *Change
	for _, event := range history {
		if event.Change == nil || event.Change.UndoOf != "" {
			continue
		}
		position++
		if event.Change.ID == changeID {
			target = event.Change
			break
		}
	}
	if target == nil {
		return fmt.Errorf("change not found: %s", changeID)
	}
	if depth > 0 && position > depth {
		return fmt.Errorf("change %s is outside the configured undo depth of %d", changeID, depth)
	}
	if maxAgeDays > 0 && now.Sub(target.AppliedAt) > time.Duration(maxAgeDays)*24*time.Hour {
		return fmt.Errorf("change %s is older than the configured undo age of %d days", changeID, maxAgeDays)
	}
	return nil
}

func BuildUndoData(change Change, before, after []byte, repairID string) ([]byte, error) {
	if repairID == "" {
		if before == nil {
			return nil, fmt.Errorf("change %s does not retain undo data", change.ID)
		}
		return append([]byte(nil), before...), nil
	}
	if after == nil {
		return nil, fmt.Errorf("change %s does not retain after data", change.ID)
	}
	document := textio.Decode(after)
	ordered := append([]Transformation(nil), change.Transformations...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Start < ordered[j].Start })
	offset := 0
	for _, transformation := range ordered {
		start := transformation.Start + offset
		end := start + len(transformation.NewText)
		if transformation.ID == repairID {
			if start < 0 || end < start || end > len(document.Text) {
				return nil, fmt.Errorf("repair %s has invalid current range", repairID)
			}
			if document.Text[start:end] != transformation.NewText {
				return nil, fmt.Errorf("repair %s no longer matches the recorded after state", repairID)
			}
			updated := document.Text[:start] + transformation.OldText + document.Text[end:]
			return document.Encode(updated), nil
		}
		offset += len(transformation.NewText) - (transformation.End - transformation.Start)
	}
	return nil, fmt.Errorf("repair not found in change %s: %s", change.ID, repairID)
}
