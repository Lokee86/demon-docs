package links

import (
	"fmt"

	"github.com/Lokee86/demon-docs/internal/textio"
)

type preparedMarkdownSource struct {
	document textio.Document
	parsed   parsedMarkdown
}

func prepareMarkdownSource(source markdownSource) (preparedMarkdownSource, error) {
	document, err := textio.Read(source.path)
	if err != nil {
		return preparedMarkdownSource{}, fmt.Errorf("read Markdown source %s: %w", source.path, err)
	}
	return preparedMarkdownSource{
		document: document,
		parsed:   parseMarkdownDocument(document.Text),
	}, nil
}

// prepareMarkdownSources reads and parses independent changed sources through
// the bounded link worker pool. Results stay indexed by deterministic source
// order so target resolution and plan mutation can remain serial.
func prepareMarkdownSources(sources []markdownSource, sourceIndexes []int) ([]preparedMarkdownSource, error) {
	prepared := make([]preparedMarkdownSource, len(sources))
	errors := runLinkWorkers(len(sourceIndexes), func(jobIndex int) error {
		sourceIndex := sourceIndexes[jobIndex]
		result, err := prepareMarkdownSource(sources[sourceIndex])
		if err != nil {
			return err
		}
		prepared[sourceIndex] = result
		return nil
	})
	for _, err := range errors {
		if err != nil {
			return nil, err
		}
	}
	return prepared, nil
}
