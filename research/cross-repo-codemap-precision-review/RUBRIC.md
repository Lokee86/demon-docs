# Cross-repository precision review rubric

Review the document and target code without consulting the suggestion tier or score in `sample-manifest.json` until the label is recorded.

Use exactly one label:

- `valid` — the target is directly useful for understanding, maintaining, testing, or implementing the documented subject and should reasonably be added to the document's codemap.
- `plausible` — the relationship is real and may help bounded context, but adding it to the document's direct codemap is optional, indirect, redundant, or broader than the document's main responsibility.
- `incorrect` — the target is unrelated, connected only through generic repository structure, or would mislead a maintainer about the document's implementation surface.

Rules:

- Existing authored links are positive context, not proof that an unlinked suggestion is wrong.
- Do not infer irrelevance merely because the document omits the target.
- Judge the semantic relationship, not whether the current algorithm score seems high enough.
- Directory-wide and generated-index documents are outside this review.
- Record a short concrete rationale naming the relevant responsibility, call path, data flow, or mismatch.
