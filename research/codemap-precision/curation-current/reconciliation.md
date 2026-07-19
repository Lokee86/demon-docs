# Duplicate-review reconciliation

The duplicate reviews for inputs 01–03 covered 54 document/target pairs.

| Shard | Label disagreements | Resolution |
|---|---:|---|
| 01 | 0 / 18 | Labels agreed. The `labeled-01` text was retained for all rows because its section-aware references, line ranges, and paraphrased excerpts were more specific and better audited than `output-01`. |
| 02 | 0 / 18 | Labels and audited row text agreed; the canonical `labeled-02` rows were retained. |
| 03 | 0 / 18 | Labels and audited row text agreed; the canonical `labeled-03` rows were retained. |
| **Total** | **0 / 54** | No label required independent adjudication. |

The canonical shards use the strict definitions in the benchmark README. All immutable suggestion fields were checked against the current sample, and all labels, rationales, source references/excerpts, target kinds, and target fingerprints were revalidated against Space Rocks revision `3387c94d10fdb94008f27b404098f3e0c32d911c`.
