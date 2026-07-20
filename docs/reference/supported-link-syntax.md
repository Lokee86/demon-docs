---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7852-ae71-3a30d63b7dfa
document_type: general
policy_exempt: false
summary: This document defines the current link forms that Demon Docs recognizes while reconciling repository Markdown. It is a parser and rewrite contract, not a statement that every link format supported by a Markdown renderer is validated.
---
# Supported Link Syntax

Parent index: [Reference](./INDEX.md)

## Purpose

This document defines the current link forms that Demon Docs recognizes while reconciling repository Markdown. It is a parser and rewrite contract, not a statement that every link format supported by a Markdown renderer is validated.

## Overview

Link scanning runs on Markdown-like source files with these extensions: `.md`, `.markdown`, `.mdown`, `.mkd`, and `.mdx`. Recognized local destinations are resolved relative to the source file, recorded in the link manifest, and may be rewritten when one deterministic target is known. The resolved path is the only rewritten portion.

The parser records the syntax family (`inline`, `reference`, `html`, or `wiki`), the raw path, any query or fragment suffix, source position, and whether a Markdown destination was angle-wrapped. Undefined reference uses are recorded separately as `reference_use` diagnostics.

## Exact contract

### Inline Markdown links and images

The parser recognizes an inline destination after a closing bracket followed by an opening parenthesis:

```markdown
[Guide](../docs/guide.md)
![Diagram](../assets/diagram.svg)
[Guide](../docs/guide.md "Read the guide")
```

Images use the same destination parser as links. Labels and image alt text are not link targets and are not changed.

The destination may have whitespace after `(`, balanced parentheses, and an angle-wrapped path:

```markdown
[Guide]( docs/guide.md )
![Diagram](<assets/diagram with spaces.svg>)
```

A Markdown title following the destination is not part of the stored path. Titles and surrounding punctuation are preserved during a rewrite. Backslash escaping affects delimiter recognition; an escaped closing bracket or closing angle bracket does not close the corresponding construct.

### Reference definitions

A reference definition is recognized when, outside protected code, a line has zero to three leading spaces and this shape:

```markdown
[guide]: ../docs/guide.md
[diagram]: <../assets/diagram with spaces.svg>
[guide-with-state]: ../docs/guide.md?view=print#intro "Guide"
```

The definition label is not a destination. The first destination token is recorded and can be repaired. Definitions may be angle-wrapped; a non-angle destination ends at whitespace. The definition may be tracked even when no reference use appears in the same document.

Reference labels are normalized for undefined-use checks by lowercasing and collapsing whitespace. The definition and use therefore match case-insensitively with normalized internal whitespace:

```markdown
[Guide]: docs/guide.md
[Read this][ guide ]
```

A definition does not cause every use to become an individual destination record. Demon Docs records the definition's destination and checks explicit or collapsed uses for an undefined label.

### Reference uses

The parser checks these use forms:

```markdown
[Read this][guide]
[guide][]
```

The collapsed form uses the first bracketed label as its reference label. If the normalized label has no definition, the use is recorded with status `undefined_reference`. A defined use is not independently resolved or rewritten; its definition is the recognized destination surface.

Shortcut reference uses such as `[guide]` are not resolved by this link parser. A malformed, multiline, or otherwise unclosed reference use is not converted into a link diagnostic.

### Wiki links and embeds

Wiki targets use double brackets, with an optional alias after the first pipe:

```markdown
[[notes/design]]
[[notes/design|Design notes]]
![[assets/diagram.svg]]
[[notes/design#overview|Overview]]
```

The same parser handles links and embeds. Leading and trailing HTML whitespace around the target is ignored for resolution. The alias is presentation text and is preserved; only the target before the first `|` is considered. Wiki markup must close on the same line and must not contain a newline.

For a wiki target without an extension, exact resolution first tries the `.md` form. A bare target such as `[[design]]` is therefore normally resolved as `design.md`, while a target that includes a path or extension keeps that explicit form. Rewrites preserve extensionless wiki style when it was unambiguous:

```text
[[design|Design]]  ->  [[archive/design|Design]]
```

If a bare wiki name would become ambiguous after a move, the move is refused or the link remains unresolved rather than silently selecting a document.

### Supported raw HTML attributes

HTML scanning is limited to these tag/attribute combinations:

| Element | Attribute |
| --- | --- |
| `a`, `link` | `href` |
| `img`, `script`, `source`, `video`, `audio`, `iframe` | `src` |
| `video` | `poster` |

Examples:

```html
<a href="docs/guide.md#intro">Guide</a>
<img src='assets/diagram.svg'>
<video poster="media/poster.png" src="media/video.mp4"></video>
```

Quoted and unquoted attribute values are accepted. Tag and attribute names are matched case-insensitively. Other attributes and elements are not link targets for reconciliation. In particular, `srcset`, `data-*`, `action`, `content`, CSS URLs, and arbitrary HTML attributes are not parsed by this surface.

### Paths, queries, and fragments

The parser splits the first literal `?` or `#` from the raw destination. The part before it is the path; the complete suffix, including the delimiter and anything after it, is preserved:

```markdown
[Guide](docs/guide.md?view=print#intro)
![Image](assets/image.png#preview)
```

The query and fragment are not used to locate the filesystem target. They remain attached to the destination when a path is repaired. A destination such as `(#intro)` has an empty path and resolves to the source document itself.

Percent-encoded path characters are decoded for local resolution. Encode a literal `?`, `#`, or space that belongs to a filename rather than a suffix, for example `file%23name.md`. Malformed percent escapes fall back to the raw path rather than producing a separate validation diagnostic.

Demon Docs does not validate whether a fragment names an existing heading or anchor. Fragment text is preserved as opaque suffix data; `#missing-heading` is not evidence that the file target is broken.

### Local targets and external targets

Relative paths are resolved from the directory containing the Markdown source:

```markdown
[Sibling](./guide.md)
[Parent](../README.md)
[Directory](assets/)
```

Existing repository files and directories can be recorded as local targets. Absolute filesystem paths are also local to the resolver, including Windows drive paths, UNC-style paths, and `file://` URLs. A filesystem path outside the repository is tracked as an external filesystem target when it can be inspected; it is distinct from an external web URL.

A URI with a non-file scheme, such as `https:`, `http:`, `mailto:`, or `ftp:`, is external to link reconciliation. It is not resolved, recorded, validated, or rewritten:

```markdown
[Website](https://example.com/guide)
[Email](mailto:docs@example.com)
```

A path that is inside a `.docignore` exclusion is skipped by link reconciliation rather than recorded as a broken target.

### Escaping and angle wrapping

Angle wrapping is supported for Markdown inline destinations and reference definitions when a path contains spaces or other characters that need a delimited destination:

```markdown
[File](<docs/guide with spaces.md>)
[File]: <docs/guide with spaces.md>
```

The angle wrapper is retained when the destination is rewritten. Non-angle paths retain their original slash/backslash and relative/absolute style where the rewrite renderer can do so. URL-escaped paths retain URL escaping, and paths containing spaces, `#`, or `?` are escaped when rendered outside an angle wrapper.

Backslashes used to escape Markdown delimiters are honored while finding the destination. Wiki targets do not have an angle-wrapped destination form; HTML values use their normal quoted or unquoted attribute syntax.

### Preservation and rewrite scope

When one target is deterministic, reconciliation or `ddocs mv` may replace only the raw path byte range. It preserves:

- link labels and image alt text;
- Markdown titles;
- query strings and fragments;
- wiki aliases and embed markers;
- HTML tag and attribute text outside the value;
- angle wrapping and supported path style;
- surrounding prose and unrelated links;
- the source's newline encoding; and
- whether the source ended with a final newline.

For example:

```markdown
![Guide](<old/guide file.png?raw=1#preview> "Keep this title")
```

may become:

```markdown
![Guide](<new/guide file.png?raw=1#preview> "Keep this title")
```

No rewrite is performed when the target is ambiguous, when a repair is blocked by review state, or when the source changed after planning. Generated rewrites are content-addressed and applied through the guarded atomic-write path.

### Fenced-code and inline-code exclusions

Markdown links, reference definitions and uses, HTML attributes, and wiki syntax inside fenced code are ignored. Fences may use backticks or tildes, require at least three marker characters, and may be indented by up to three spaces. A closing fence must use the same marker and at least the opening marker count:

````markdown
```markdown
[not-a-link](ignored.md)
[[not-a-wiki-link]]
```
````

An unterminated fence protects the remainder of the source. Inline code spans delimited by matching backtick runs are also protected:

```markdown
`[not-a-link](ignored.md)`
```

This protection is parser behavior; it does not attempt to interpret arbitrary Markdown extensions inside code.

## Defaults and precedence

The current precedence and resolution rules are:

1. Code protection is applied before link syntax scanning.
2. A literal query or fragment delimiter splits the suffix from the path before local resolution.
3. A `file://` URL is treated as a local filesystem path; other URI schemes are external.
4. Relative paths are joined to the source directory, while absolute filesystem paths remain absolute.
5. Wiki paths without an extension use `.md` for exact lookup and candidate matching.
6. An exact target wins; a unique candidate may be used for a missing or moved target. Multiple candidates remain ambiguous.
7. A repository ignore rule suppresses tracking of the target.
8. Only a deterministic local target is eligible for an automatic path-only rewrite.

The first stateful reconciliation pass establishes a baseline and does not infer historical moves. Later passes may use stored identity and fingerprint evidence, or one unique candidate, to repair a local destination.

## Diagnostics and failure behavior

Recognized local destinations produce link records with statuses including:

| Status | Meaning |
| --- | --- |
| `valid` | The resolved target exists and matches the requested path. |
| `case_mismatch` | A target exists with a case difference from the requested path. |
| `moved` | A deterministic repair was planned or applied for a moved target. |
| `broken` | No current target was found. |
| `ambiguous` | Multiple candidates could satisfy the destination. |
| `blocked` | A deterministic repair is held by review policy. |
| `stale_block` | A previous repair block no longer matches current evidence. |
| `undefined_reference` | An explicit or collapsed reference use has no matching definition. |

Broken, ambiguous, undefined, and blocked conditions increment the unresolved count and are reported with source path, line, column where available, and the relevant destination or label. Candidate paths are included for ambiguity. The source remains unchanged unless a safe rewrite is planned and its expected source hash still matches at apply time.

External URI targets, unsupported syntax, ignored targets, and unvalidated heading fragments do not produce broken-link diagnostics from this parser. They are outside this contract rather than confirmed valid.

## Examples

A mixed document containing recognized forms:

```markdown
[Guide](docs/guide.md?print=1#intro)
![Diagram](<assets/diagram with spaces.svg>)
[guide-ref]: docs/guide.md#intro
[Read the guide][guide-ref]
[[docs/guide|Guide]]
![[assets/diagram.svg]]
<a href="docs/guide.md">HTML guide</a>
```

Forms intentionally outside the reconciliation surface:

```markdown
[Web](https://example.com)
[Shortcut reference]
`[Code](docs/guide.md)`
```

## Related docs

- [Reference](./INDEX.md)
- [Diagnostics and Exit Behavior](./diagnostics-and-exit-behavior.md)
- [Managed Files and State](./managed-files-and-state.md)
- [Markdown Link Reconciliation](../architecture/markdown-link-reconciliation.md)
- [Reconciliation Pipeline](../architecture/reconciliation-pipeline.md)
- [Stateless Document Refactoring](../guides/document-refactoring.md)
- [Current Product Limitations](../limits/current-limitations.md)

## Notes

This page describes the current implementation in `internal/links/`. It does not define the full grammar of CommonMark, GitHub-Flavored Markdown, wiki engines, or HTML, and it does not imply heading-fragment validation.
