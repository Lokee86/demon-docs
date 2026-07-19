# Space Rocks Codemap Dataset Audit

## Automated snapshot

- Markdown documents scanned: 340
- Documents with codemaps: 150 (44.12%)
- Implementation-facing heuristic coverage: 150/182 (82.42%)
- Extracted links: 4516
- Resolved exact links: 4446
- Missing links: 24
- Ambiguous links: 0
- Pattern-resolved links: 18
- Unsupported/template links: 28
- Entries with descriptions: 21.86%
- Entries with group context: 96.41%
- Map size median / p90 / maximum: 28.5 / 47.9 / 93

## Trusted subset representation

- Reviewed documents: 10 (6.67% of mapped documents)
- Trusted links: 51 (1.13% of extracted links)
- Categories represented: data, devtools, protocol, services
- All top-level categories currently producing parsed codemaps are represented, but service coverage is shallow: the sample omits API-server, web, and most gameplay/runtime subsystems.

## Initial interpretation

- The corpus is large enough to benchmark extraction and ranking, but the manually trusted subset is too small to establish whole-corpus quality.
- Existing-link resolution is a freshness check, not a semantic completeness check. A second review must inspect whether major implementation/test/contract links are missing from sampled documents.
- Very large maps and directory/glob entries should be scored separately from selective exact-file maps because they are much easier to recover and may inflate recall.

## Next audit slice

1. Review a stratified sample from every mapped top-level category.
2. For each sample, confirm existing-link usefulness and record only potentially missing links.
3. Compare the trusted subset distribution with the full corpus by map size, syntax, category, and target kind.
4. Build a benchmark-eligible subset that excludes stale, ambiguous, pattern, directory-only, and overbroad inventory maps unless evaluated separately.

## Policy-aware coverage

- Required service/data/devtools-client-server documents mapped: 138/142 (97.18%).
- Services: 104/108.
- Data: 9/9.
- Devtools client/server: 25/25.
- Protocol is conditional by policy: 6/10 currently have maps.
- Domains correctly have 0 direct codemaps.

The earlier broad implementation-facing heuristic was too aggressive because Space Rocks policy explicitly forbids direct domain codemaps and does not require planning maps.

## Structured links outside codemaps

- Structured path references found outside codemap sections: 1302 across 133 mapped documents.
- Test/verification candidates outside codemaps: 924.
- Source/generated/consumer candidates outside codemaps: 338.

These are candidates, not automatically valid missing links. They show that benchmark ground truth based only on parsed codemap sections can omit relationships that the same document explicitly records elsewhere.

The clearest recurring case is a `## Tests` section placed after `## Code map`. The extractor correctly stops at the peer heading, so those test paths never enter the generated codemap dataset even though documentation policy says code maps should include related tests.

## Manual sample findings

A small cross-area review found that existing maps are generally semantically useful:

- `damage-resolution.md`, `runtime-processing.md`, `auth-and-oauth.md`, and `drop-tables.md` connect the document to ownership, integration, contract, generated, and test boundaries that a repository graph alone would not identify.
- `runtime-processing.md` and `auth-and-oauth.md` place substantial test inventories under peer `## Tests` sections. Those links are useful, but the current codemap extractor does not include them.
- `realtime-compact-wire-mapping.md` contains a clear map under `## Code Paths`. The current default heading set does not recognize it, so this is an adapter/compatibility miss rather than absent documentation.
- `presentation-bridge.md` and `inbound-packet-routing.md` are implementation-facing service docs with no recognized map and appear to be genuine documentation coverage gaps.
- `hosted-webrtc-connectivity.md` is primarily deployment policy and may reasonably be exempt despite living under `services/`.

The four unmapped service documents therefore should not be treated as one homogeneous failure category.

## Revised conclusion

- Structural policy coverage is strong: all required data and devtools client/server docs are mapped, and service coverage is at least 96.3% before classifying policy exceptions and aliases.
- Existing maps are generally rich and current, but the extracted dataset is not yet a complete representation of the semantic links authored in the docs.
- The current 51-link trusted subset is suitable as an initial positive set, but it is too small and too shallow within the service category for final tuning.
- Before aggressive precision tuning, add a second trusted fold from API-server, client runtime, game-server combat/networking, and web docs.
- Reconcile `Tests`/`Verification` sections and repository-specific aliases such as `Code Paths` before treating parsed codemap entries as complete ground truth.
