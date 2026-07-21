package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/documentpolicy"
	"github.com/Lokee86/demon-docs/internal/frontmatter"
)

const documentCount = 1000

func main() {
	root, err := os.MkdirTemp("", "ddocs-validation-bench-")
	must(err)
	defer os.RemoveAll(root)

	docs := filepath.Join(root, "docs")
	schemas := filepath.Join(root, "schemas")
	must(os.MkdirAll(docs, 0o755))
	must(os.MkdirAll(schemas, 0o755))
	must(os.WriteFile(filepath.Join(schemas, "general.toml"), []byte("version = 1\nname = 'general'\nunknown_sections = 'allow'\nduplicate_sections = 'allow'\n"), 0o644))

	body := strings.Repeat("## Section\n\nA paragraph with **formatting**, `code`, and a [link](target.md).\n\n", 12)
	for index := 0; index < documentCount; index++ {
		id := fmt.Sprintf("%08x-0000-4000-8000-%012x", index+1, index+1)
		content := fmt.Sprintf("---\nauthor: benchmark\ncreated: '2026-07-21'\ndocument_id: %s\ndocument_type: general\nsummary: Benchmark document %d.\n---\n# Document %d\n\n%s", id, index, index, body)
		must(os.WriteFile(filepath.Join(docs, fmt.Sprintf("document-%04d.md", index)), []byte(content), 0o644))
	}

	frontmatterConfig := config.Default()
	frontmatterConfig.Root = "docs"
	frontmatterConfig.Frontmatter.Enabled = true
	frontmatterConfig.Frontmatter.UnknownFields = "ignore"
	frontmatterConfig.Format.Enabled = false

	formatConfig := config.Default()
	formatConfig.Root = "docs"
	formatConfig.Frontmatter.Enabled = true
	formatConfig.Format = config.Format{
		Enabled:           true,
		SchemaDir:         "schemas",
		DocumentSchemaDir: "document-schemas",
		DefaultSchema:     "general",
	}

	fmt.Printf("documents=%d gomaxprocs=%d\n", documentCount, runtime.GOMAXPROCS(0))
	measure("frontmatter-cold", 5, func() error {
		plan, err := frontmatter.Build(root, docs, frontmatterConfig, false, time.Now())
		if err == nil && len(plan.Diagnostics) != 0 {
			return fmt.Errorf("frontmatter diagnostics: %d", len(plan.Diagnostics))
		}
		return err
	})
	measure("format-cold", 5, func() error {
		plan, err := documentpolicy.Build(root, docs, formatConfig, false)
		if err == nil && len(plan.Diagnostics) != 0 {
			return fmt.Errorf("format diagnostics: %d", len(plan.Diagnostics))
		}
		return err
	})
}

func measure(name string, runs int, operation func() error) {
	times := make([]time.Duration, 0, runs)
	for run := 0; run < runs; run++ {
		started := time.Now()
		must(operation())
		times = append(times, time.Since(started))
	}
	var total time.Duration
	best := times[0]
	for _, elapsed := range times {
		total += elapsed
		if elapsed < best {
			best = elapsed
		}
	}
	fmt.Printf("%s mean=%s best=%s runs=%v\n", name, total/time.Duration(len(times)), best, times)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
