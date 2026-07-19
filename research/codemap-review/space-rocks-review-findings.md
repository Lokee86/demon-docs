# Space Rocks Codemap Ground-Truth Review

## Purpose

This review creates a small, defensible positive-link corpus for evaluating Demon Docs missing-link suggestions.

It does not judge existing links as irrelevant and does not create negative examples. Entries that cannot safely be treated as ground truth are excluded from scoring.

## Corpus snapshot

Reviewed Space Rocks revision:

```text
3387c94d10fdb94008f27b404098f3e0c32d911c
```

A broad heuristic scan found:

- 151 documents with recognized code-map headings;
- maps under services, protocol, data, devtools, and one planning stub;
- several thousand path-shaped references when prose, tests, boundaries, directories, and repeated references are all counted.

The trusted subset contains 51 links across 10 documents:

- service implementation maps;
- client and server maps;
- protocol maps spanning multiple runtimes;
- data/source/test maps;
- devtools maps;
- exact-file and directory-level links.

## Review rules

A link entered the trusted subset only when:

1. it appeared inside the document's codemap section;
2. its repository target existed at the reviewed revision;
3. the document or codemap grouping gave a clear reason for the relationship;
4. the relationship was useful for recovering implementation context.

An entry was excluded when it was stale, wildcard-based, ambiguous, or merely looked like a path to the heuristic scanner.

Exclusion means **unknown for benchmark purposes**, not irrelevant.

## Relationship vocabulary

The sample uses six intentionally broad relationships:

- `implementation` — directly implements the documented behavior;
- `contract` — defines source data, schema, packets, configuration, or another authoritative contract;
- `generated` — generated artifact consumed by the behavior;
- `test` — verifies the documented behavior;
- `integration` — hosts, routes, consumes, or hands off behavior across an ownership boundary;
- `gate` — controls build/runtime availability.

These are sufficient for the first benchmark. More specific labels can be added later without changing the underlying document-to-target edge.

## Findings

### Space Rocks is useful, but not perfect ground truth

The codemaps contain strong semantic information that a repository graph alone cannot provide. Group headings and per-file descriptions distinguish ownership, source contracts, tests, generated outputs, and non-ownership boundaries.

However, the maps also contain drift and ambiguity. The review found missing paths, wildcard families, directory links, and prose that is easy for a path extractor to misread. Existing maps therefore need validation before becoming benchmark labels.

### Map style varies substantially

The corpus includes:

- annotated file entries with symbol-level explanations;
- grouped fenced path lists;
- flat code-and-test lists;
- directory-level maps;
- large cross-service protocol inventories;
- explicit non-ownership boundary sections.

Demon Docs should normalize these styles internally instead of requiring one Markdown shape.

### Group labels are valuable evidence

Headings such as `Primary implementation files`, `Related tests`, `Contract and schema sources`, and `Important non-ownership boundaries` provide relationship information cheaply and deterministically.

The extractor should retain the nearest group label rather than flattening every path into an untyped edge.

### Directory and wildcard links need separate handling

Directory links are valid semantic links, but they are easier to rediscover and less precise than file links. Benchmarks should report them separately.

Wildcard entries should not be treated as exact edges. They may later become pattern edges or expand to validated concrete targets.

### Stale links are validation findings, not negative labels

A missing target may indicate a rename, deletion, or documentation drift. It does not prove that the original semantic relationship was poor.

The suggestion system should never use a stale entry to infer that an existing link is irrelevant. It should only avoid placing that entry in the trusted positive set until reconciled.

## Benchmark guidance

For the first hidden-link recovery benchmark:

1. use exact-file trusted links only;
2. hide one trusted link at a time from a document;
3. rank candidates without reading the hidden link;
4. measure whether the hidden target appears in the top candidate set;
5. report directory links separately;
6. ignore excluded entries during precision and recall scoring;
7. never convert unselected candidates into negative labels.

Recommended initial metrics:

- top-1 recovery;
- top-5 recovery;
- mean reciprocal rank;
- suggestions per recovered link;
- recovery by relationship type;
- recovery by map style.

## Implication for the algorithm

The first algorithm does not need to generate prose or decide that links should be removed.

It needs to answer one narrow question:

> Given a document and its remaining trusted links, which existing repository targets are most likely to be missing from its codemap?

The reviewed subset provides stable positive examples for that question while keeping uncertain corpus entries out of the score.
