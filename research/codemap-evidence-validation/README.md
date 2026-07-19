# Space Rocks Evidence-Signal Validation

This validation set checks every initial missing-link evidence signal against targets from the manually reviewed Space Rocks codemap corpus.

The executable fixture is `internal/evidence/testdata/space-rocks-signal-cases.json`. Each case names a reviewed document-to-target link and supplies only the evidence needed to exercise one signal.

Most cases use evidence observed directly in the repository, documentation, or Git history. The unique-basename case is explicitly marked as synthetic isolated evidence because the ten-document trusted sample contains no reviewed link that is mentioned only by a unique basename outside its codemap.

The test verifies that:

- every case points to a link in `research/codemap-review/space-rocks-trusted-links.json`;
- every evidence kind has at least one case;
- the hidden trusted target is recovered with the expected evidence kind; and
- accepted existing targets are never returned as missing-link candidates.

This is signal validation, not an end-to-end precision or recall benchmark. The benchmark runner owns holdout selection, full repository inputs, scoring, and aggregate reports.
