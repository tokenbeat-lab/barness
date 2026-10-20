Spec-axis preflight delta review: PASS; 0 actionable findings.

Reviewed de051931b8d4e30fb6adf5cc5b2a3672f243ba35 against reviewed 8d564103, including the source delta from integration 09d0c68, with the fixed-base ecfd4b4 review context retained. Confidence: high for this delta.

(a) Missing/partial requirements: none newly identified. R27 requires “全部既有十三组合、十四自身支持身份…继续有效” and “不得删除旧映射” (spec.md:300). Historical GLM playback now selects the original catalog from the indexed immutable Git commit using a bounded read, recomputes its hash, and checks the original report's exact commit/hash/version. Original payload, Result and Observer comparisons remain unchanged. Missing/invalid/oversized snapshots fail closed; no normalization or old evidence rewrite was added.

(b) Scope creep: none found. ADR0044 limits historical catalog authority to replay. Current actual formal selection still uses BuiltinCatalog; publishing requires live/formal mode; the support importer independently requires the current complete catalog and execution source. The added child checks explicitly reject current certification and publication of historical replay. This preserves R26's “candidate/controlled/capture replay 不发布正式 PASS” (spec.md:294).

(c) Wrong-looking implementation: none confirmed. MiniMax replay chooses only candidate or current built-in snapshots and verifies both hash and version, resolving the shared-label collision without accepting a label as authority. Historical catalog comparisons exclude only independently added provider subsets; unchanged old models still require exact equality/hash. Current metadata tests use the checked-in release snapshot while independent usage/hash pins remain. The controlled Host fixture replaces only its owned MiniMax subset, preserving other providers and all six synthetic candidates.

Catalog renewal updates the generated snapshot and four current pin documents, verifies every noncatalog field is unchanged, preserves fixed budgets/retries/instructions/datasets and legacy 180-second versus route 600-second limits, and restores original bytes on failure.

Actual candidate/formal evidence, activation, legacy/effect renewal and final gates remain tracked in tasks 10/09. R27 still requires current verifiable reports (spec.md:302); this review is not formal-support certification. No repository edits, tests, credentials, environment files or network/vendor calls were used.
