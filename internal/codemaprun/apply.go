package codemaprun

import "github.com/Lokee86/demon-docs/internal/filetxn"

func Apply(plan Plan) error {
	_, err := filetxn.Apply(plan.Rewrites)
	return err
}
