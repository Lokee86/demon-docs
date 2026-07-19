package links

import "strings"

// PrepareSelectionPlan removes automatically planned repairs before a user
// selection is added. Reconcile describes automatic repairs as if their link
// metadata were already updated, so restore those records to the current
// on-disk destinations before saving a plan that applies only the selection.
func PrepareSelectionPlan(plan *Plan) {
	transformations := make(map[string]LinkTransformation)
	for _, rewrite := range plan.Rewrites {
		for _, transformation := range rewrite.Transformations {
			transformations[transformation.LinkID] = transformation
		}
	}

	reverted := make(map[string]struct{})
	for index := range plan.Links.Links {
		record := &plan.Links.Links[index]
		transformation, ok := transformations[record.ID]
		if !ok {
			continue
		}
		record.RawPath = transformation.OldDestination
		record.Target = transformation.OldDestination + record.Suffix
		record.Status = "moved"
		if strings.EqualFold(transformation.OldDestination, transformation.NewDestination) {
			record.Status = "case_mismatch"
		}
		if record.ResolvedPath != "" {
			record.Candidates = []string{record.ResolvedPath}
		}
		if _, seen := reverted[record.ID]; !seen {
			plan.Unresolved++
			reverted[record.ID] = struct{}{}
		}
	}

	plan.Updates = nil
	plan.Rewrites = nil
	plan.Suppressions = nil
	plan.AppliedChanges = nil
}
