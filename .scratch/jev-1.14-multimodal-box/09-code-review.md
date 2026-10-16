# Ticket 09 two-axis review

Fixed baseline: `58ec73c03a3d86c29f1d12ab82bed14184d77e89`.
Reviewed implementation: `c6bdfcd`, then corrections through `e429932`.
Diff: `git diff 58ec73c...e429932`; commit list:
`git log 58ec73c..e429932 --oneline`.

Two independent read-only sub-agents reviewed Standards and Spec in parallel,
following `/Users/cyber/.agents/skills/code-review/SKILL.md`. They used the ticket,
spec/design, AGENTS.md, glossary, relevant ADRs and the skill's complete smell
baseline. No user README/glossary/ADR-0030 work was included. Each confirmed
resource or encoding finding received a public failing E2E before its repair;
reviewer reproductions used external Go overlays without changing source files.

## Standards

Initial review found three documented-standard violations:

- Valid ambiguous promoted business fields were rejected, while encoding/json
  omits them (AGENTS §3/§9; ADR-0021 arbitrary JSON State).
- Repeated field lookup made wide struct inspection quadratic and allocated
  large index metadata (AGENTS §5/§9).
- Integer-key asset maps were omitted from early decoded-size accounting
  (AGENTS §3/§5; ADR-0021 known-size checks).

Incremental reviews found four additional instances of the same boundary:
quoted primitive fields; declared interface/multiple-pointer quoted applicability;
ordinary byte slices encoded as Base64 content; decoded byte-slice asset data.
Each violated boundary conversion/resource rules in AGENTS §3/§5 and ADR-0021.

The repairs resolve JSON fields once, cache content field selection per type,
honor encoder promotion/ambiguity/omission/quoting, account predictable map-key
and Base64 representations, and preserve custom-encoder handling. No separate
smell-baseline findings were reported.

Final reviewer report at `e429932`: **0 remaining findings**. The byte-slice
asset fix uses decoded length before encoding. All prior findings are resolved
with public E2E regressions retained. No source edits by the reviewer.

## Spec

Initial review found four gaps:

- Typed asset containers bypassed early reference budgets (R08: known excess
  must be checked before copying/serialization).
- Valid ambiguous business fields were rejected (R10/legal final Hooks; R08).
- Untagged anonymous Go type names double-counted promoted JSON fields (R08:
  count each actual occurrence).
- Batch TotalAlloc stopped at blocked EOF, excluding response decoding (issue
  09: repeatable memory/resource records).

Incremental reviews found predictable quoted primitive fields, numeric map keys
with a JSON-only marshaler, declared interface quoting and UTF-8 coercion. These
likewise violated R08. Verified red reproductions allocated tens of MiB before
an eight-reference rejection; the largest interface case allocated 143 MiB.
The investigated empty json.Number/key `0` corner was explicitly withdrawn as a
release finding because that key conflicts with the supported Jev model name.

All confirmed findings were repaired. The final allocation sample is after call
completion, and the EOF snapshot remains separately recorded. Exact wire
fixtures and per-call capture histories preserve controls, retries and responses;
the independent verifier checks their inventory and public results. The known
100ms timeout-ordering failure remains with its original needs-triage issue.

Final reviewer report at `e429932`: **0 remaining findings**, no scope creep.
Focused borrowed-reference/accounting/quoted-field E2Es pass, including byte-data.
Final full/race execution and sealing are owned by the parent task. Controlled
HTTP evidence retains real-service/visual-quality status NOT_RUN. No source edits
by the reviewer.

The last full run of `e429932` exposed the same separately tracked timer symptom
in the new deadline fixture: `fitCallTimeout` had unintentionally set both call
and read-idle limits to 250ms. The raw failed bundle remains
`20261016T093522.829653000Z`. Test-only `f97648f` uses a 250ms Host deadline and
the original finite policy. Both reviewers independently confirmed **0 findings**:
same payloads/concurrency/case count, same aborted terminal assertion, unchanged
production code; coincident internal timers are outside this ticket. This is
consistent with the tracked symptom, whose cause remains unresolved.

Final totals: Standards **0 remaining** (7 repaired); Spec **0 remaining**
(8 repaired). The axes retain their independent classifications.
