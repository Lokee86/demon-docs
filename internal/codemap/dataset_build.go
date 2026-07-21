package codemap

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	ignorepolicy "github.com/Lokee86/demon-docs/internal/ignore"
)

type datasetDocumentJob struct {
	filePath     string
	documentPath string
}

type datasetDocumentResult struct {
	document    DocumentRecord
	entries     []DatasetEntry
	diagnostics []Diagnostic
}

// BuildDataset scans Markdown documents under docsRoot, extracts authored code
// maps, and resolves their targets against repositoryRoot. Output ordering and
// hashes depend only on repository content and the supplied format.
func BuildDataset(repositoryRoot, docsRoot string, format Format) (Dataset, error) {
	repositoryRoot, err := filepath.Abs(repositoryRoot)
	if err != nil {
		return Dataset{}, err
	}
	docsRoot, err = filepath.Abs(docsRoot)
	if err != nil {
		return Dataset{}, err
	}
	if !within(repositoryRoot, docsRoot) {
		return Dataset{}, fmt.Errorf("docs root %s is outside repository root %s", docsRoot, repositoryRoot)
	}
	if format.TargetBase == "" {
		format.TargetBase = TargetBaseRepository
	}
	policy, err := ignorepolicy.Load(repositoryRoot)
	if err != nil {
		return Dataset{}, err
	}
	jobs, err := discoverDatasetDocuments(repositoryRoot, docsRoot, policy)
	if err != nil {
		return Dataset{}, err
	}

	results := make([]datasetDocumentResult, len(jobs))
	contentCache := newTargetContentCache(os.ReadFile)
	errors := runDatasetWorkers(len(jobs), func(index int) error {
		result, err := prepareDatasetDocument(repositoryRoot, jobs[index], format, contentCache)
		results[index] = result
		return err
	})
	for _, err := range errors {
		if err != nil {
			return Dataset{}, err
		}
	}

	dataset := Dataset{SchemaVersion: DatasetSchemaVersion}
	for _, result := range results {
		dataset.Documents = append(dataset.Documents, result.document)
		dataset.Entries = append(dataset.Entries, result.entries...)
		dataset.Diagnostics = append(dataset.Diagnostics, result.diagnostics...)
	}
	sortDataset(&dataset)
	return dataset, nil
}

func discoverDatasetDocuments(repositoryRoot, docsRoot string, policy ignorepolicy.Policy) ([]datasetDocumentJob, error) {
	var jobs []datasetDocumentJob
	err := filepath.WalkDir(docsRoot, func(filePath string, item os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if filePath != docsRoot {
			ignored, err := policy.Ignored(filePath, item.IsDir())
			if err != nil {
				return err
			}
			if ignored {
				if item.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if item.IsDir() || item.Type()&os.ModeSymlink != 0 || !strings.EqualFold(filepath.Ext(filePath), ".md") {
			return nil
		}
		documentPath, err := repositoryRelative(repositoryRoot, filePath)
		if err != nil {
			return err
		}
		jobs = append(jobs, datasetDocumentJob{filePath: filePath, documentPath: documentPath})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(jobs, func(i, j int) bool { return jobs[i].documentPath < jobs[j].documentPath })
	return jobs, nil
}

func prepareDatasetDocument(repositoryRoot string, job datasetDocumentJob, format Format, contentCache *targetContentCache) (datasetDocumentResult, error) {
	source, err := os.ReadFile(job.filePath)
	if err != nil {
		return datasetDocumentResult{}, err
	}
	extracted := Extract(job.documentPath, string(source), format)
	result := datasetDocumentResult{
		document: DocumentRecord{
			Path:            job.documentPath,
			Size:            int64(len(source)),
			SHA256:          digest(source),
			SectionCount:    extracted.SectionCount,
			EntryCount:      len(extracted.Entries),
			DiagnosticCount: len(extracted.Diagnostics),
		},
		diagnostics: extracted.Diagnostics,
		entries:     make([]DatasetEntry, 0, len(extracted.Entries)),
	}
	for _, entry := range extracted.Entries {
		resolution, err := resolveTargetWithCache(repositoryRoot, job.documentPath, entry, format, contentCache)
		if err != nil {
			return datasetDocumentResult{}, err
		}
		result.entries = append(result.entries, DatasetEntry{Entry: entry, Resolution: resolution})
	}
	return result, nil
}
