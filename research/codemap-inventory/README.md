# Space Rocks Codemap Format Inventory

This inventory records the codemap shapes currently present in the Space Rocks documentation corpus. It exists to guide Demon Docs' first codemap adapter and extractor fixtures. It does not define one mandatory Markdown format and does not require Space Rocks documents to be rewritten.

## Corpus

The inventory scanned every Markdown file under `space-rocks/docs/`, excluding nested worktrees.

Observed authored map sections:

| Heading | Sections |
| --- | ---: |
| `Code map` | 142 |
| `Code or source map` | 8 |
| `Code and test map` | 1 |
| **Total** | **151** |

All observed map headings are level-two headings, but that is a repository convention rather than a portable requirement.

Body forms:

| Form | Sections |
| --- | ---: |
| Fenced blocks only | 91 |
| Markdown bullets only | 57 |
| Mixed fenced blocks and bullets | 1 |
| Legacy indented entries | 1 |
| Unpopulated TODO placeholder | 1 |

Additional structure is common:

- 48 sections use nested headings to group entries.
- 91 sections use plain prose labels ending in a colon.
- Both `*` and `-` bullet markers occur.
- Fences may be tagged `text` or left untagged.
- No table-based codemap was found.
- No Markdown-link-based codemap was found.

Broad heading matching is unsafe. `Code Map Policy`, `Code Maps`, and `Implementation references` also occur in the corpus, but they are policy, procedure, or explicitly non-codemap sections. An adapter should match configured heading aliases exactly, case-insensitively.

## Observed Entry Shapes

### Fenced path lists

The most common form is one or more fenced blocks containing repository paths. Prose labels or nested headings group the blocks.

````md
## Code map

Primary files:

```text
services/example/runtime.go
services/example/config/
```
````

### Bullet targets

A bullet may begin with an inline-code target. The first inline-code span is the target; later inline-code spans belong to the description.

```md
* `client/scripts/session/controller.gd` - Creates `SessionState` and owns session transitions.
```

The first target may be a file, directory, glob, or symbol. A symbol-first bullet is explicit enough to preserve as a symbol candidate even when it cannot yet be resolved.

### Fenced target-description pairs

Some fenced blocks alternate a target line with a description line.

```text
web/src/content.ts
-> owns the content schema
```

Another observed form uses indentation:

```text
internal/spatial/index.go
    Defines the spatial index contract.
```

The description is not a second target.

### Leading path plus prose

Some fenced boundary blocks put the target and description on the same line.

```text
services/example/runtime/ owns runtime state.
```

The leading path is the target. The remaining text is its description.

### Legacy indented entries

One document uses an indented inline-code target followed by indented prose.

```md
  `services/example/session.go`

    Defines the session state.
    Owns the respawn cooldown.
```

The same style also permits a target and description on one indented line.

### Mixed sections

A section may combine fenced path inventories with boundary bullets. The extractor must preserve both forms instead of selecting one parser for the whole section.

### Placeholders

A codemap section may exist but contain only a TODO with no target.

```md
## Code map

- TODO: add paths when implementation exists.
```

This is an empty map, not an unresolved target.

## Target Forms

The Space Rocks corpus contains:

- repository-relative files;
- repository-relative directories;
- component-relative paths such as `internal/game/spatial/index.go` inside game-server documentation;
- glob selectors such as `services/game-server/internal/game/control_*.go`;
- symbol-first bullets such as `DevtoolsWindowController`; and
- descriptions containing additional code spans that are not independent targets.

A reusable resolver therefore cannot assume every path is repository-relative. Resolution should try the repository root and configured component roots, then report unresolved or ambiguous candidates without guessing.

## Initial Adapter Contract

The first Space Rocks-compatible adapter should:

1. Match configured heading aliases exactly and case-insensitively.
2. Accept the configured heading at any Markdown level.
3. Treat the section as ending at the next heading of the same or higher level.
4. Recognize fenced, bullet, legacy-indented, and mixed entry shapes.
5. Preserve group labels, descriptions, source document, and source span.
6. Preserve file, directory, glob, and symbol target text before resolution.
7. Treat only the first structured target position as a target; later code spans remain description text.
8. Resolve against the repository root and configured component roots.
9. Emit diagnostics for missing, ambiguous, or unsupported targets.
10. Treat TODO-only sections as empty maps.
11. Never rewrite an existing codemap merely to normalize its style.

The adapter should not ingest `Code Map Policy`, procedural `Code Maps` sections, or `Implementation references` unless a repository explicitly configures those headings as map aliases.

## Normalized Record

The extractor should be able to produce a record equivalent to:

```text
document: docs/services/example.md
heading: Code map
group: Primary implementation
target_text: services/example/runtime.go
target_kind: file
description: Owns runtime state.
syntax_kind: bullet
source_span: document line range
resolution_base: repository
```

This is an internal representation. It does not prescribe how the source document must be written.

## Representative Space Rocks Sources

- `docs/data/observability-contract.md`: simple fenced file list and alternate heading.
- `docs/data/packet-schemas.md`: grouped fenced lists plus leading-path ownership statements.
- `docs/services/client/app-shell-and-session/room-session-state.md`: grouped bullets with file descriptions and symbol mentions.
- `docs/services/web/devlog-static-site.md`: fenced path/arrow-description pairs.
- `docs/services/game-server/simulation/world/spatial-query-index.md`: component-relative paths with indented descriptions.
- `docs/services/game-server/simulation/players/player-respawn.md`: legacy indented entries.
- `docs/devtools/client/devtools-window.md`: mixed fenced and bullet forms.
- `docs/planning/web/stubs/interactive-website.md`: TODO-only placeholder.

## Fixtures

The `fixtures/` directory contains reduced examples of each supported shape. `fixtures/expected.json` records the normalized entries expected from each fixture. These fixtures are intentionally repository-neutral while retaining the syntax found in Space Rocks.
