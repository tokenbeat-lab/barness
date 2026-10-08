# Code review — issue 11

Fixed point: c0b860c80e2175676984d93dbfa32a063be0cb71; first candidate 5b7b150.
Two independent read-only agents reviewed Standards and Spec with the implement
skill's code-review workflow. Source/evidence fixes are kept in a follow-up commit.

## Standards

One documented P2: per-case generation/mask replay selected only its own subtest;
the edit-first guard therefore made it NOT_RUN. This violated AGENTS.md's repeatable
artifact and verification requirement. Resolved by a pipeline replay override at
existing evidence.Case, generating the complete image combination in one process.
The original manifest/assertions replay metadata is corrected; vendor captures,
actual results, request IDs and accounting were not changed.

Two judgement calls: repeated expected-set validation in classifier/images merger
and transparent background proof without alpha validation. A shared expected-set
check now retains separate route/model/budget semantics. Full raster verification
requires transparent pixels for a requested transparent background; controlled
opaque-output E2E first fails then passes as a refusal. The original WebP has alpha.

## Spec

One P2: the same replay defect, independently reproduced by the reviewer with an
isolated dummy credential: NOT_RUN, zero HTTP attempts. This conflicted with the
spec's repeatable E2E evidence package requirement. Resolved by full-combination
replay and a test of the actual emitted manifest. No other missing/incorrect
requirements or scope creep were found. Input fidelity and account/region limits
are accurately documented.

## Verification of fixes

Review-red/transparency-red logs retain the failure; review-green and final-targeted-
race logs verify fixes with the shared TypeSafe route. Mask input format/size failures
were additionally checked before tightening optional rejection classification;
those errors now remain FAIL rather than becoming unsupported mask capability.
The first full run exposed a stale responses/usage.json directory hash pin; its
failure/audit are retained and the pin is updated to the separately regenerated
catalog snapshot. Final full-suite and follow-up review results are appended below.

Initial findings: Standards 1 documented + 2 judgement calls; Spec 1 documented.
Both documented findings concern reproducibility and are corrected.

## Follow-up review at 53b12e8

Standards: PASS, no actionable documented violations or remaining smell findings.
The reviewer verified the emitted full-combination replay, transparency guard,
shared expected-set validation, optional mask rejection semantics and capture
limits; the targeted live-tag race regression passed.

Spec: PASS, no remaining missing/incorrect requirements or scope creep. The
reviewer independently checked the replay regression and preserved image-first
prerequisite, usage/cost/accounting evidence and conservative catalog scope.

Remaining findings: Standards 0; Spec 0. Full-suite evidence is recorded separately
without changing either review axis.
