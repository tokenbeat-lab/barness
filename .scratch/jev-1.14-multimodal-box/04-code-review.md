# Ticket 04 — two-axis code review

Fixed point: task-start `490e2f901d711ded59df97f4156fa6ffb43d47f5`.
Reviewed scope: `git diff 490e2f901d711ded59df97f4156fa6ffb43d47f5...724a0a7`.
Commits: `5615b02` implementation, `7bfec2e` approved-design corrections,
`724a0a7` projected-State allocation guard.

The implement skill explicitly requests code-review; that skill requests two
parallel, independent read-only subagents. Both axes were reviewed again after
corrections. Preexisting user readmes, glossary and unrelated planning changes
were outside the committed diff. Final evidence sealing was assessed separately.

## Standards

Initial and final review: **0 findings**. No documented-standard violations or
actionable baseline smells found. The diff preserves native protocol ownership,
explicit authority, finite budgets, immutable snapshots, cancellation and atomic
publication. Public E2E seams and ADR-0021 retain the rationale and rollback path.
Final projected-State reservation uses remaining-budget subtraction before
expansion; no file exceeds the repository's 500-line limit.

## Spec

Initial review: **3 P2 findings**; follow-up review: **1 P2 finding**. All fixed
with public failing E2Es before the relevant implementation correction:

1. Native asset `type` may be omitted; when present it must be `image`. The
   builder still always emits it. Missing type now succeeds, explicit null and
   other invalid values reject.
2. Jev model image count applies to independent sources/assets. Explicit
   reference budgets govern repeated uses; Decisions retains its expanded cap.
3. `MaxStateBytes` includes the complete projected native State, including
   generated keys and wrapper framing, at pre-pin and final Hook boundaries.
4. Known projected-State oversize must reject before expansion/serialization.
   A 10,000-reference, 128-byte State case exposed 2,106,664 bytes allocated after
   resolution. The corrected guard reserves encoded State, key sizes, framing
   and commas before appending; the same E2E now observes <128 KiB allocation.

Final independent review of `490e2f9...724a0a7`: **0 residual findings**, no scope
creep. Red/green logs are preserved under `04-logs/`; final test artifacts are
linked by the verification record and integrity inventory.

Final total: Standards 0 unresolved; Spec 0 unresolved (4 P2 findings resolved).
