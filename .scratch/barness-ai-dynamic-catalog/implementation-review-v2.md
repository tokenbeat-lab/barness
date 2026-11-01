# Integration review resolution

Original review: [v1](implementation-review-v1.md). Frozen reviewed code: `853279be371ab6887b183d85dd147ecd31c5debd`. These static conclusions precede the independent full integration run.

## Standards

**Cleared — no unresolved Standards findings or concrete regressions found in frozen `853279be371ab6887b183d85dd147ecd31c5debd`.**

- Revised price tiers still require every explicit threshold/rate before typed canonicalization. [revise.go:37](/Users/cyber/.codex/worktrees/catalog-review-fixes/barness/catalog/maintenance/revise.go:37) now uses shared cost-only validation, so it no longer imposes addition-only reasoning-map requirements on the nine accepted legacy definitions.
- Post-rename failures remain correctly distinguished from pre-install failures. [install.go:12](/Users/cyber/.codex/worktrees/catalog-review-fixes/barness/catalog/maintenance/install.go:12) performs bounded restoration of the pointer/output and reports `storage_failed` with repair requirements when rollback or its durability cannot complete.
- [diff.go:64](/Users/cyber/.codex/worktrees/catalog-review-fixes/barness/catalog/maintenance/diff.go:64) ignores object-key ordering inside otherwise equal arrays while retaining whole-array changes for genuine edits.
- [catalog_check.py:28](/Users/cyber/.codex/worktrees/catalog-review-fixes/barness/tools/verification/catalog_check.py:28) now inventories and archives all embedded baseline assets, including the compiler-required README sentinel.

No substantive baseline smell findings. Static re-review only; no tests executed or repository files changed.


## Spec

Reviewed frozen `853279be` in `/Users/cyber/.codex/worktrees/catalog-review-fixes/barness`, including `47e0835c...853279be` and the retained original fixes. Static, read-only review; no tests executed by this reviewer.

All four original Spec findings remain resolved:

1. **Revised tier presence:** `revise.go:37–42` calls `requiredCostFields` before typed canonicalization. This still checks all base rates and every tier threshold/rate, so missing prices cannot become zero merely through decoding (spec.md:154–156).
2. **Source operation isolation:** shared operation validation remains in both association and price-refresh paths; absent type maps only to chat and decision to classifier. Same wire ID cannot borrow another operation’s prices (spec.md:109,121).
3. **Adopted provenance:** complete baseline validation, seed-lineage retention and subsequent export consumption remain intact. `catalog_check.py` now includes every embedded baseline asset, including the initial README sentinel, in the actual source inventory/archive. The new public scenario rebuilds the maintenance CLI from that retained source (spec.md:108,138, verification requirements).
4. **Generator failures:** `flag.ContinueOnError` still routes invalid flags through JSON error output and nonzero exit (spec.md:129–130).

**The follow-up legacy-reasoning P2 is resolved.** Revision validation now shares only raw price-presence rules; unrelated revisions no longer impose add-time thinking-map requirements on the nine accepted legacy definitions. Public scenarios explicitly cover name, price and empty overlays for each definition. This preserves existing omitted-map behavior (spec.md:120).

The array comparison change preserves element order and whole-array differences while ignoring object-key ordering within equal elements. It prevents serialization-only changes from manufacturing a candidate without hiding real array edits. No new concrete Spec defect or semantic regression found within the agreed review scope.

Spec final re-review: 0 open findings. This conclusion is a code-review result, not a claim that the independently running combined verification has completed or that vendor/release qualification changed.


## Evidence and remaining integration work

All original findings and follow-up regressions have public RED/GREEN evidence in [review-fixes-verification-v1.json](review-fixes-verification-v1.json). The subsequent full ordinary attempt found Chinese evaluation consumer drift and was stopped; see [integration-interrupted-v1.json](integration-interrupted-v1.json). That regression requires repair and focused re-review before final ordinary verification can pass. No vendor, race, pressure or release qualification is asserted.
