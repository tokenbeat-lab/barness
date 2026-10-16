# Ticket 12 two-axis code review

Fixed point: `ba68f0cf962941a26307dfd334b80887026dbbdc` (task-start HEAD).
Implementation: `d4cd265`; corrections: `f7c9bcf`, source-set hardening `e2c8287`, official path replay `e6548ae`.
Diff: `git diff ba68f0cf962941a26307dfd334b80887026dbbdc...HEAD`.
Spec: [ticket 12](issues/12-independent-visual-quality-evaluation.md).
Standards: AGENTS.md, GLOSSARY.md, docs/agents and ADR-0024/0029.
The code-review skill ran independent parallel Standards and Spec agents.
Unrelated existing user documentation changes were excluded from both commits.

## Standards

Initial independent finding, P2: cancelled samples bypassed consumption
verification (`verify.go`'s cancellation early return). A fresh deadline package
with empty HTTP response was modified to claim 900 input tokens/$123 complete
usage; matching Observer, totals and inventory were updated, and the verifier
accepted it. This violated AGENTS principle 9 and ADR-0024's actually reported
consumption requirement. No additional actionable baseline smells were found.

Correction: cancelled results must be either entirely zero/unreported with absent
response model before decoding, or exactly supported by captured-response replay
at that candidate price snapshot. Both public route processes first reproduced
this defect, then passed the regression. Independent reinspection repeated the
900/$123 corruption and observed rejection. **No remaining Standards findings.**

## Spec

Initial independent findings, both P2:

1. Bounded response capture discarded the extra byte used to detect the 64 KiB
   output limit. A 70,000-byte padding response correctly failed with
   resource_limit/response, but its untouched package could not replay that error.
2. Final Observer metadata was compared, but terminal Error and Usage were not.
   A successful sample's final observation could be resealed with a contradictory
   tenant_denied/scope error and pass verification.

Corrections: preserve the bounded overflow sentinel; compare terminal usage and
error code/phase to the public result. Each had public process red regressions
before correction. Independent reinspection reproduced the oversize workflow and
verified the honest failure package, then rejected contradictory Observer errors.
The cancellation accounting fix was also rechecked. **No remaining Spec findings.**

Initial totals: Standards 1 (worst P2); Spec 2 (worst P2).
Final totals: Standards 0; Spec 0. No real-network or credential use in review.

## Supplementary source-set closure

Receipt sealing found an additional P2 completeness gap: checking source-map
lengths and hashing a missing lookup as empty bytes allowed an unknown empty
file to replace a required source. Both route public processes reproduced it
before the membership check was added; red:
`.evidence/barness-ai/20261016T122811.258504000Z`.
Commit `e2c8287` now requires each source name to exist in the current source
set before comparing hashes. Both independent axes inspected the one-condition
fix and its fully resealed same-length regression: no remaining findings.

## Supplementary official-path closure

The final real-route read-through found that official services use `/v1` while
controlled execution/replay previously omitted that prefix. A real package would
therefore fail strict path equality. Both routes/entries first failed the
public official-path assertion (`20261016T123506.931140000Z`), then reproduced
replay failure with the prefix preserved (`20261016T123515.863408000Z`).
Commit `e6548ae` derives the official base path from the existing fixed route URL
for both controlled execution and offline replay. All four workflows pass.
Independent Standards and Spec scoped reviews report no remaining findings.
