# Codemap Ground-Truth Review

This directory contains the manually reviewed Space Rocks codemap subset for the missing-link suggestion benchmark.

The subset is deliberately conservative:

- every trusted entry is an existing document-to-code link already present in Space Rocks documentation;
- links are classified by their role in the document;
- stale, wildcard, parser-artifact, or otherwise ambiguous entries are excluded rather than treated as negative examples;
- directory links remain marked as directories so benchmarks can evaluate them separately from exact-file links.

Files:

- `space-rocks-trusted-links.json` — machine-readable trusted links and excluded entries.
- `space-rocks-review-findings.md` — review method, corpus findings, and benchmark guidance.
