package app

import (
	"fmt"
	"time"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/documentpolicy"
	"github.com/Lokee86/demon-docs/internal/frontmatter"
	"github.com/Lokee86/demon-docs/internal/links"
	"github.com/Lokee86/demon-docs/internal/model"
	"github.com/Lokee86/demon-docs/internal/reconcile"
	"github.com/Lokee86/demon-docs/internal/repository"
	"github.com/Lokee86/demon-docs/internal/reverseindex"
	"github.com/Lokee86/demon-docs/internal/validationcache"
	"github.com/Lokee86/demon-docs/internal/validationworkers"
	"github.com/Lokee86/demon-docs/internal/watch"
)

type checkPlanningResult struct {
	index           model.ReconcileResult
	links           links.Plan
	frontmatter     frontmatter.Plan
	format          documentpolicy.Plan
	reverse         reverseindex.Plan
	validationCache *validationcache.Store
}

func buildCheckPlans(scope repository.Scope, cfg config.Config, features watch.Features, reverse reverseOptions) (checkPlanningResult, error) {
	result := checkPlanningResult{}
	if features.Frontmatter || features.Format {
		cache, err := validationcache.Open(scope.RepositoryRoot)
		if err != nil {
			return result, fmt.Errorf("open validation cache: %w", err)
		}
		result.validationCache = cache
	}

	planners := make([]func() error, 0, 5)
	if features.Indexes {
		planners = append(planners, func() error {
			planned, err := reconcile.TreeWithIgnoreRoot(scope.DocsRoot, scope.RepositoryRoot, cfg)
			result.index = planned
			return err
		})
	}
	if features.Reverse {
		planners = append(planners, func() error {
			planned, err := reverseindex.Build(scope.RepositoryRoot, scope.DocsRoot, reverse.roots, cfg, reverse.format)
			result.reverse = planned
			return err
		})
	}
	validationTime := time.Now()
	if features.Frontmatter {
		planners = append(planners, func() error {
			planned, err := frontmatter.BuildWithValidationCache(scope.RepositoryRoot, scope.DocsRoot, cfg, false, validationTime, result.validationCache)
			result.frontmatter = planned
			return err
		})
	}
	if features.Format {
		planners = append(planners, func() error {
			planned, err := documentpolicy.BuildWithValidationCache(scope.RepositoryRoot, scope.DocsRoot, cfg, false, result.validationCache)
			result.format = planned
			return err
		})
	}
	if features.TrackLinks {
		planners = append(planners, func() error {
			var planned links.Plan
			var err error
			if features.Links {
				planned, err = links.Reconcile(scope.RepositoryRoot)
			} else {
				planned, err = links.Track(scope.RepositoryRoot)
			}
			result.links = planned
			return err
		})
	}

	plannerErrors := validationworkers.Run(len(planners), func(index int) error {
		return planners[index]()
	})
	for _, err := range plannerErrors {
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

func (result checkPlanningResult) savePrivateState(features watch.Features) error {
	if result.validationCache != nil {
		if err := result.validationCache.Save(); err != nil {
			return err
		}
	}
	if features.TrackLinks && !features.Links {
		if err := links.Save(result.links); err != nil {
			return err
		}
	}
	return nil
}
