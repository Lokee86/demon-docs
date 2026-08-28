package codemaprun

import (
	"github.com/Lokee86/demon-docs/internal/codemapsemantic"
	"github.com/Lokee86/demon-docs/internal/filetxn"
)

func Apply(plan Plan) error {
	if _, err := filetxn.Apply(plan.Rewrites); err != nil {
		return err
	}
	return codemapsemantic.SaveAll(plan.RepositoryRoot, plan.BaselineUpdates)
}
