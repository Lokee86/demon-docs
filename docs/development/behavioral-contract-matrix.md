---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7b71-a12d-60a0efab1898
document_type: general
policy_exempt: false
summary: This document maps Demon Docs' critical behavioral contracts to their canonical documentation owners, focused tests, integration coverage, and release gates.
---
# Behavioral Contract Matrix

Parent index: [Development](./INDEX.md)

## Purpose

This document maps Demon Docs' critical behavioral contracts to their canonical documentation owners, focused tests, integration coverage, and release gates.

## Overview

A package inventory proves that code has a documentation pointer. It does not prove that the system's safety properties are protected or that a maintainer knows which test must change with a behavioral decision.

This matrix uses the more useful unit of coverage:

```text
stateful flow
mutation boundary
persistent model
concurrency boundary
public contract
extension seam
```

Each row names a durable contract rather than every individual test. Test files listed here are the primary evidence; related tests may exist elsewhere.

## How to use this matrix

Before changing a listed behavior:

1. read the canonical owner;
2. identify whether the change is a contract change or an implementation-preserving refactor;
3. update or add the focused test before weakening an invariant;
4. run the broader package or integration gate;
5. update this matrix when ownership or verification changes; and
6. update public reference, limits, or migration documentation when users can observe the change.

A test passing does not authorize undocumented contract changes. A document claim without a protecting test should be treated as an identified coverage gap.

## Managed documentation contracts

| Contract | Canonical owner | Focused tests | Broader gate |
| --- | --- | --- | --- |
| Managed replacements preserve bytes outside the owned block | [Managed Markdown Transformation](../architecture/managed-markdown-transformation.md) | `internal/markdown/source_preservation_test.go`, `markdown_test.go` | `go test ./internal/markdown ./internal/reconcile -count=1` |
| Fence-contained headings, markers, and parent lines remain examples | [Managed Markdown Transformation](../architecture/managed-markdown-transformation.md) | `TestGoldmarkIgnoresHeadingsInsideCodeFences`, `TestMarkerLikeFenceContentIsNeverManaged`, `TestUpdateParentIgnoresFencedParentLinkCandidates` | regression fixtures and full Go suite |
| Missing and legacy managed sections migrate within bounded structural ranges | [Managed Markdown Transformation](../architecture/managed-markdown-transformation.md) | `TestManagedSectionMigrationAndPlacement`, `TestMalformedBlockRepairStopsBeforeFollowingSection` | `make regression` |
| Stable descriptions survive normal reconciliation | [Managed Markdown Transformation](../architecture/managed-markdown-transformation.md) | `TestExistingDescriptionsAndRootDisplayTitleRemainStable` | fixture matrix |
| Direct/stub and unique cross-folder transitions preserve descriptions without ambiguous guessing | [Managed Markdown Transformation](../architecture/managed-markdown-transformation.md) | `transitions_moves_test.go` | fixture matrix |
| LF, CRLF, mixed endings, trailing spaces, and final-newline state are preserved as documented | [Managed Markdown Transformation](../architecture/managed-markdown-transformation.md) | `line_endings_test.go`, `source_preservation_test.go`, `textio_test.go` | Linux and Windows CI |
| Planning is deterministic, idempotent, and non-mutating | [Reconciliation Model](../architecture/reconciliation-pipeline.md) | `determinism_test.go`, `TestCheckPlanningDoesNotMutate` | `ddocs fix --indexes` followed by clean `ddocs check --indexes` |
| Missing generated indexes expose their complete planned content to later policy stages, and generated index defaults do not depend on format enforcement | [Managed Markdown Transformation](../architecture/managed-markdown-transformation.md), [Front Matter Schemas](../reference/frontmatter.md) | `TestFreshRepositoryGeneratedIndexConverges`, `TestGeneratedIndexesReceiveFrontmatterInSameFix`, `TestBuildUsesGeneratedIndexDefaultsWithoutFormat` | `go test ./internal/app ./internal/frontmatter -count=1` |
| No forward-index write escapes the managed docs root | [Managed Markdown Transformation](../architecture/managed-markdown-transformation.md) | `TestApplyWithinRejectsOutsideWriteBeforeMutation` | full Go suite |

## Frontmatter and document-format contracts

| Contract | Canonical owner | Focused tests | Broader gate |
| --- | --- | --- | --- |
| YAML and TOML frontmatter parse and render deterministically while preserving the body | [Front Matter Schemas](../reference/frontmatter.md) | `parse_test.go`, `TestParseSupportsYAMLAndTOMLAndPreservesBody`, `TestRenderPreservesSelectedFormatAndSortsFields` | `go test ./internal/frontmatter -count=1` |
| Unknown-field policy, configured sources, immutable values, and conditional requirements remain explicit | [Front Matter Schemas](../reference/frontmatter.md) | `evaluate_test.go`, `TestUnknownFieldModes`, `TestEvaluateRestoresRecordedImmutableValue`, `TestConditionalRuleCanRemainUnresolvedOrUseConfiguredSource` | `go test ./internal/frontmatter -count=1` |
| Frontmatter plans stay inside the docs root, detect duplicate document IDs, preserve source encoding, and converge on repeat apply | [Front Matter Schemas](../reference/frontmatter.md) | `plan_test.go`, `TestBuildApplyIsIdempotentPreservesCRLFAndKeepsIDAcrossMove`, `TestBuildStaysInsideDocsRootAndDetectsDuplicateIDs` | `go test ./internal/frontmatter ./internal/filetxn -count=1` |
| Unchanged clean frontmatter and document-format validation skips document parsing/evaluation while content, policy, schema, engine, immutable, and duplicate-identity changes invalidate reuse | [Validation Cache](../architecture/validation-cache.md) | `validation_cache_test.go` in `internal/validationcache`, `internal/frontmatter`, and `internal/documentpolicy` | `go test ./internal/validationcache ./internal/frontmatter ./internal/documentpolicy -count=1` |
| Metadata selects a document schema before path fallback and schema validation rejects unsafe hierarchy or ambiguous aliases | [Document Schemas And Format Enforcement](../reference/document-schemas.md) | `TestSchemaSelectionUsesMetadataBeforePathFallback`, `TestSchemaValidationRejectsCyclesAndAmbiguousSiblingHeadings` | `go test ./internal/documentpolicy -count=1` |
| Unknown or duplicate human-authored sections block automatic body mutation until an explicit resolution is recorded | [Document Schemas And Format Enforcement](../reference/document-schemas.md) | `TestUnknownSectionBlocksFix`, app integration tests for ignore/merge/delete | `go test ./internal/documentpolicy ./internal/app -count=1` |
| Stable section IDs drive deterministic heading renames and cumulative document-specific schema invalidation | [Document Schemas And Format Enforcement](../reference/document-schemas.md) | `TestSchemaRenameUsesStableSectionID`, migration tests in `document_policy_migration_test.go` | `go test ./internal/documentpolicy ./internal/app -count=1` |
| A required codemap section is created only through validated effective-schema placement; schemas without one remain unchanged | [Codemap Managed Execution](../architecture/codemap-managed-execution.md) | `TestCodemapSchemaProviderPlacesRequiredServiceSection`, `TestCodemapSchemaProviderLeavesSchemaWithoutCodemapUnchanged`, `TestReconcileManagedSkipsMissingSectionWithoutSchema` | codemap execution CLI test and focused package suites |

## Link and authored-file mutation contracts

| Contract | Canonical owner | Focused tests | Broader gate |
| --- | --- | --- | --- |
| Supported link syntax is parsed without treating protected code as links | [Supported Link Syntax](../reference/supported-link-syntax.md), [Markdown Link Reconciliation](../architecture/markdown-link-reconciliation.md) | `internal/links/parser_test.go` and syntax-specific tests | `go test ./internal/links -count=1` |
| First link pass records a baseline before identity-based repair | [Markdown Link Reconciliation](../architecture/markdown-link-reconciliation.md) | `TestFirstScanRecordsOnlyThenRepairsMovedNonMarkdownTarget`, `TestBrokenLinkGuessWaitsUntilAfterInitialScan` | repository link fixture checks |
| One live `document_id` collapses stale absent private aliases, remaps links, merges history, and restores historical-path repair without weakening ambiguity refusal | [Link Reconciliation State Machine](../architecture/link-reconciliation-state-machine.md) | `TestCollapseDocumentIdentityAliasesRemapsLinksAndMergesHistory`, document-alias reconciliation tests, `TestDocumentIDsPreserveLinkIdentityAcrossModifiedMassMoves` | `go test ./internal/links -count=1` |
| Ambiguous targets are reported rather than selected automatically | [Markdown Link Reconciliation](../architecture/markdown-link-reconciliation.md) | `TestAmbiguousGuessIsLeftForTheUser`, wiki ambiguity tests | `ddocs check --links` |
| Stale stored link offsets abandon the internal move fast path and rebuild from current source parsing | [Link Reconciliation State Machine](../architecture/link-reconciliation-state-machine.md) | `TestStaleStoredOffsetsFallBackToCurrentSourceParsing` | `go test ./internal/links -count=1` |
| Batch preflight prevents every generated source write when any source changed | [Repository State and Transactions](../architecture/repository-state-and-transactions.md) | `TestApplyGeneratedPreflightFailurePreventsAllWrites` | `go test ./internal/links -count=1` |
| Generated replacement preserves suppression order and deterministic plans | [Watcher and Automation](../operations/watcher-and-automation.md) | `TestApplyGeneratedPreservesSuppressionOrder`, concurrency tests | link integration suite |
| Watch startup retries recognized stale move plans, and event-buffer overflow requests a complete reconciliation without terminating observation | [Watch Scheduler and Reconciliation Serialization](../architecture/watch-scheduler.md), [Repository Demon](../operations/repository-demon.md) | `TestInitialReconciliationRetriesTransientFilesystemRaces`, `TestWatcherRecoversFromEventOverflowWithFullReconciliation` | `go test ./internal/watch -count=1` |
| New generated rewrites and scoped tracking preserve unrelated pending watcher suppressions | [Generated Rewrite Publication](../architecture/generated-rewrite-publication.md), [Watcher and Automation](../operations/watcher-and-automation.md) | suppression merge tests, `TestTrackSourcesRefreshesOnlySelectedSourceRecords` | `go test ./internal/links -count=1` |
| Policy and index fixes skip clean-run link tracking and refresh only changed source paths | [Reconciliation Command Lifecycle](../architecture/reconciliation-command-lifecycle.md) | `TestFrontmatterOnlyCleanFixDoesNotRefreshLinkState`, `TestTrackSourcesRefreshesOnlySelectedSourceRecords` | `go test ./internal/app ./internal/links -count=1` |
| Rollback never overwrites content created after Demon Docs' write | [Repository State and Transactions](../architecture/repository-state-and-transactions.md) | `TestRollbackGeneratedRefusesToOverwriteNewerContent` | full Go suite |
| Review-publication failure restores generated source content | [Review Ledger](../architecture/review-ledger.md) | `TestApplyAndSaveRestoresSourcesWhenReviewBatchFails`, `TestRollbackAfterReviewFailureRestoresUndoSource` | review CLI integration suite |
| Stateless move preflights hashes and repository containment | [Stateless Document Refactoring](../guides/document-refactoring.md) | `internal/links/move_test.go`, `internal/app/move_test.go` | `go test ./internal/links ./internal/app -count=1` |
| Shared rewrite batches reject duplicate or inconsistent inputs, preflight every source before writes, verify new hashes, and refuse rollback over newer content | [Generated Rewrite Publication](../architecture/generated-rewrite-publication.md) | `TestApplyAndRollbackBatch`, `TestPreflightFailurePreventsEveryWrite`, `TestRollbackRefusesNewerContent` | `go test ./internal/filetxn -count=1` |

## Private state and review contracts

| Contract | Canonical owner | Focused tests | Broader gate |
| --- | --- | --- | --- |
| Private state transactions reject stale bases | [Repository State and Transactions](../architecture/repository-state-and-transactions.md) | `TestRepositoryRejectsStaleTransaction` | `go test ./internal/ddrepo -count=1` |
| One record update rewrites only its deterministic shard and root | [Repository State and Transactions](../architecture/repository-state-and-transactions.md) | `TestSingleRecordUpdateOnlyChangesItsShard`, codec tests | private-state package suite |
| Review event batches publish one complete `batch.json` commit or nothing, while legacy per-event commits remain readable | [Review Ledger](../architecture/review-ledger.md), [Generated Rewrite Publication](../architecture/generated-rewrite-publication.md) | `store_batch_test.go` batch object-count, nil/empty snapshot, legacy history, and compaction cases | `go test ./internal/review ./internal/links ./internal/app -count=1` |
| Private object compaction runs after durable publication, preserves every referenced state/review/undo object, and is non-fatal to the logical write | [Private Object Repository](../architecture/private-object-repository.md) | `internal/ddrepo/compaction_test.go`, review compaction retention tests | `go test ./internal/ddrepo ./internal/review -count=1` |
| Declines remain effective until their evidence fingerprint changes | [Review Ledger](../architecture/review-ledger.md) | `TestPolicyKeepsDeclineUntilFingerprintChanges` | review CLI integration suite |
| User selection applies only the chosen candidate after preflight | [Reviewing Suggestions and Changes](../guides/reviewing-suggestions-and-changes.md) | `TestSuggestionsSelectPreflightsAndAppliesOnlyChosenRepair`, `TestPrepareSelectionPlanRemovesAutomaticWritesAndRestoresRecords` | review CLI integration suite |
| Undo can target one repair while preserving unrelated transformations | [Review Ledger](../architecture/review-ledger.md) | `TestBuildUndoDataSupportsOneRepairWithinFileChange`, review CLI tests | full Go suite |
| Undo-created repair blocks prevent immediate deterministic reapplication | [Review Ledger](../architecture/review-ledger.md) | `TestReviewCLIRecordsUndoAndBlocksDeterministicRepair`, `TestBlockedDeterministicRepairIsNotReapplied` | review CLI integration suite |

## Watcher and daemon contracts

| Contract | Canonical owner | Focused tests | Broader gate |
| --- | --- | --- | --- |
| Debounced events produce one run plus one follow-up when events arrive during execution | [Watcher and Automation](../operations/watcher-and-automation.md) | `TestSchedulerDebouncesAndRunsFollowup` | `go test ./internal/watch -count=1` |
| Selected systems share one reconciliation run lock | [Watcher and Automation](../operations/watcher-and-automation.md) | `TestRootSelectedWithRunLockSerializesReconciliation`, reverse watch serialization test | watch and reverse-index suites |
| Initial reconciliation completes before observer creation | [Watcher and Automation](../operations/watcher-and-automation.md) | `TestInitialFixCompletesBeforeObserverCreation` | watch integration suite |
| Generated writes do not create an infinite self-write loop | [Watcher and Automation](../operations/watcher-and-automation.md) | `TestWatchConvergesWithoutSelfWriteLoop` | watch integration suite |
| Watch scope follows new directories and `.docignore` changes | [Watcher and Automation](../operations/watcher-and-automation.md) | watcher contract tests for nested directories, deletion, and ignore reload | Linux and Windows CI |
| Exactly one fresh repository-demon owner may hold the lease | [Repository Demon](../operations/repository-demon.md) | `runtime_test.go`, `ownership_stress_test.go` | repeated focused stress plus full suite |
| Stale or abandoned demon ownership can be recovered token-safely | [Repository Demon](../operations/repository-demon.md) | stale recovery and contention stress tests | Linux and Windows CI |
| Status is read-only and does not delete expired feeder records | [Repository Demon](../operations/repository-demon.md) | `TestStatusSnapshotDoesNotDeleteExpiredFeeder`, app status test | demon package and CLI suite |

## Reverse-index contracts

| Contract | Canonical owner | Focused tests | Broader gate |
| --- | --- | --- | --- |
| Reverse roots require explicit safe scope and cannot contain the docs root | [Reverse Index Architecture](../architecture/reverse-indexes.md) | `root_scope_test.go` | `go test ./internal/reverseindex ./internal/app -count=1` |
| Reverse traversal honors nested `.docignore` domains | [Reverse Index Architecture](../architecture/reverse-indexes.md) | `TestBuildHonorsNestedDocignoreFiles`, watch reload test | reverse-index integration suite |
| Missing or unresolved scoped codemap targets fail check deterministically | [Reverse Index Architecture](../architecture/reverse-indexes.md) | reverse-index build and app tests | `ddocs check --reverse` on fixtures |
| Reverse watch participates in shared serialization | [Watcher and Automation](../operations/watcher-and-automation.md) | `TestWatchWithRunLockSerializesReconciliation` | combined watch suite |

## Codemap analysis contracts

| Contract | Canonical owner | Focused tests | Broader gate |
| --- | --- | --- | --- |
| Extraction reads only configured codemap sections and preserves source spans | [Codemap Extraction and Dataset](../architecture/codemap-extraction-and-dataset.md) | extractor and inventory fixture tests | `go test ./internal/codemap -count=1` |
| Target resolution keeps missing, ambiguous, unsupported, and pattern states explicit | [Codemap Extraction and Dataset](../architecture/codemap-extraction-and-dataset.md) | dataset target-base/root/ambiguity tests | codemap export command tests |
| Corpus adapters emit only recognized local deterministic facts | [Codemap Corpus and Adapters](../architecture/codemap-corpus-adapters.md) | corpus dependency, symbol, path, history, and related-document tests | `go test ./internal/codemapcorpus -count=1` |
| Existing authored targets and the document itself are excluded from candidates | [Codemap Evidence and Ranking](../architecture/codemap-evidence-and-ranking.md) | evidence collector tests | evidence and benchmark suites |
| Recommendation ranking, admission, caps, fan-out discounting, negative-evidence filters, and tiers are deterministic | [Codemap Evidence and Ranking](../architecture/codemap-evidence-and-ranking.md) | `internal/codemaprecommend/suggestions_test.go`, evidence symbol tests, benchmark compatibility tests | pinned source-report comparison plus `go test ./internal/codemaprecommend ./internal/codemapbench -count=1` |
| Controlled holdouts do not leak through document text, visible targets, or related documents | [Codemap Benchmark Methodology](../research/codemap-benchmark-methodology.md) | run/orchestrator tests and app benchmark engine isolation test | pinned repository holdout |
| Holdout selection is deterministic and input-order independent | [Codemap Benchmark Methodology](../research/codemap-benchmark-methodology.md) | `holdout_test.go` | repeated benchmark with fixed seed |
| Report JSON is canonical and schema-versioned | [Codemap Report Formats](../reference/codemap-report-formats.md) | `report_export_test.go`, dataset stable JSON test | artifact diff in research validation |
| Precision samples are deterministic, stratified, auditable, and fully labeled before evaluation | [Codemap Precision Governance](../research/codemap-precision-governance.md) | `codemapprecision/precision_test.go` | pinned labeled evaluation |
| Existing links are retained by default; confidence pruning occurs only under explicit independent settings | [Codemap Managed Execution](../architecture/codemap-managed-execution.md), [Configuration Reference](../reference/configuration.md) | `TestReconcileManagedRemovesOnlySelectedEntry`, `internal/codemaprun/build_test.go` pruning cases | `codemap inspect` and dry-run against representative repository docs plus full suite |
| Complete existing codemap sections are adopted into one managed region without authored/generated provenance splitting | [Codemap Managed Execution](../architecture/codemap-managed-execution.md), [Managed Files and State](../reference/managed-files-and-state.md) | `TestReconcileManagedAdoptsWholeExistingSection`, `TestReconcileManagedUnifiesExistingPartialManagedRegion` | focused codemap package suite and idempotent second fix |
| Fenced Space Rocks-style maps remain fenced and do not receive redundant bullet lists | [Codemap Managed Execution](../architecture/codemap-managed-execution.md) | `TestReconcileManagedPreservesFencedCodemapStyle` | representative Space Rocks dry-run plus focused codemap tests |
| A missing section is skipped without a schema and created only from an explicit validated schema placement | [Codemap Managed Execution](../architecture/codemap-managed-execution.md) | `TestReconcileManagedSkipsMissingSectionWithoutSchema`, `TestReconcileManagedCreatesOnlySchemaRequiredSection`, `TestCodemapSchemaProviderPlacesRequiredServiceSection` | codemap execution CLI test and focused codemap/document-policy suites |
| Unchanged declined recommendations are suppressed before section reconciliation | [Codemap Managed Execution](../architecture/codemap-managed-execution.md), [Review Lifecycles](../architecture/review-lifecycles.md) | `internal/codemaprun/build_test.go`, `TestPolicyKeepsDeclineUntilFingerprintChanges` | codemap inspect plus review integration suite |
| Codemap check, inspect, and dry-run do not write; fix converges and uses content-addressed publication | [Codemap Managed Execution](../architecture/codemap-managed-execution.md) | `TestCodemapFixDryRunCheckAndApplySingleFile`, `internal/filetxn/apply_test.go` | check/fix/check smoke plus full Go suite |
| Normal reconciliation, watch, and repository-demon paths never invoke codemap generation | [Codemap Managed Execution](../architecture/codemap-managed-execution.md), [Application Orchestration](../architecture/application-orchestration.md) | codemap command dispatch tests and absence from reconciliation/watch call graphs | full application, watch, and demon suites |

## CLI and configuration contracts

| Contract | Canonical owner | Focused tests | Broader gate |
| --- | --- | --- | --- |
| `check` reports drift without writing | [CLI Reference](../reference/cli.md), [Application Orchestration](../architecture/application-orchestration.md) | `TestCheckReportsDriftWithoutWriting` and feature-selection tests | fixture matrix |
| `check` and `fix` select and order indexes, frontmatter, body format, links, and reverse indexes as documented | [Reconciliation Command Lifecycle](../architecture/reconciliation-command-lifecycle.md) | `app_test.go`, `feature_flags_test.go`, frontmatter/format integration tests | `go test ./internal/app ./internal/frontmatter ./internal/documentpolicy -count=1` |
| Feature selectors run only requested systems; `--indexes` is index-only and `--docs` is the indexes/frontmatter/format umbrella | [CLI Reference](../reference/cli.md), [Compatibility and Migrations](../reference/compatibility-and-migrations.md) | `feature_flags_test.go`, help tests | CLI regression suite |
| Every public and nested command has scoped side-effect-free help | [Testing and Fixtures](testing-and-fixtures.md) | `help_test.go`, `help_nested_test.go`, `cmd/demon/main_test.go` | smoke gate |
| The documented PowerShell shell-hook command emits one native-output object, installs under Windows PowerShell, and preserves complete repository paths and shell counts | [Repository Demon](../operations/repository-demon.md) | `TestShellHookUsesTokenLeaveAndValidPowerShellInstallation`, `TestPowerShellHookBootstrapInstallsFunctions` | Windows app tests |
| Codemap fix permits an optional root; check and inspect require one; singular and plural command names route to the same implementation | [CLI Reference](../reference/cli.md), [Codemap Managed Execution](../architecture/codemap-managed-execution.md) | `TestCodemapExecutionHelpAndRequiredRoots` | executable help smoke and app suite |
| Configuration selection and compatibility aliases retain documented precedence | [Configuration Reference](../reference/configuration.md), [Compatibility and Migrations](../reference/compatibility-and-migrations.md) | config behavior and alias tests | Linux and Windows CI |
| Config mutation preserves unrelated comments, keys, and formatting | [Configuration Reference](../reference/configuration.md) | demon-run atomic edit tests | config package suite |

## Release-gate mapping

The preferred complete gate remains:

```bash
make release-check
```

It combines package tests, fixture regression, vet, builds, and executable smoke tests. Additional research gates are required only when codemap ranking, evidence, sampling, or report semantics change.

Documentation-only changes should still run:

```bash
go run ./cmd/ddocs fix --indexes
go run ./cmd/ddocs check --indexes
go run ./cmd/ddocs check --frontmatter
go run ./cmd/ddocs check --format
go run ./cmd/ddocs check --links
go test ./... -count=1
go vet ./...
```

## Adding a contract

Add a row when a change introduces or reveals a durable invariant whose violation could:

- alter authored files unexpectedly;
- corrupt or fork persistent state;
- weaken ambiguity or containment refusal;
- change concurrency ownership;
- change a public command, format, or diagnostic;
- invalidate a benchmark or precision comparison; or
- make a future refactor unsafe without knowing the behavior.

Do not add rows for private helper implementation details that are fully covered by a broader contract.

## Failure modes

This matrix becomes misleading when:

- test names move and the row is not updated;
- a broad package suite is listed without naming the contract-focused test;
- research artifacts are treated as universal release guarantees;
- a changed contract updates tests but not canonical documentation;
- a document claims a guarantee that no test protects; or
- a package row is mistaken for flow-level behavioral coverage.

## Code map

- `internal/**/*_test.go` — focused package contracts.
- `tests/` — repository-level fixture regression.
- `.github/workflows/ci.yml` — cross-platform gates.
- `Makefile` — local release-check composition.
- `research/` — pinned codemap and performance validation artifacts.

## Related docs

- [Testing and Fixtures](testing-and-fixtures.md)
- [Documentation Coverage Map](documentation-coverage.md)
- [Documentation Policy](../documentation-policy.md)
- [Safe Extension Procedures](safe-extension-procedures.md)
- [Managed Markdown Transformation](../architecture/managed-markdown-transformation.md)
- [Codemap Pipeline](../architecture/codemap-pipeline.md)
- [Codemap Managed Execution](../architecture/codemap-managed-execution.md)
- [Managing Codemaps](../guides/managing-codemaps.md)

## Notes

This matrix documents the tests that protect intended behavior. It is not a generated code-coverage report and does not replace reading the focused tests before changing their contract.
