# Ticket 13 prerequisite check — implementation not started

Checked on 2026-10-17 (Asia/Shanghai), baseline
`9c25b5583621fa18cc286726ed115298b0ad1fa8`, current branch `main`.

**Admission is BLOCKED; ticket-13 implementation and real verification are
NOT_RUN.** This is preparation and a repeatable check of existing evidence,
not completed implementation, a new E2E red/green cycle or a formal support report.

The ticket explicitly requires: “只有候选协议与独立效果证据完整且可审计后才考虑纳入固定
Jev 1.14”. Tickets 11/12 are resolved for their own implementation scope; both
explicitly preserve real-service NOT_RUN. That does not satisfy ticket 13's
additional admission prerequisite. No approved dedicated route credential or
authorization configuration was supplied in this task. The pending request asks
for that configuration or existing authentic real evidence; no other task's key
or authorization was read or used.

## Rechecked evidence

The [receipt pointer](13-prerequisite-evidence.json) seals the new local command
transcript, baseline input/report hashes and verifier binary hash. Its sibling
`ticket-input.md` preserves the original ticket bytes matching the recorded input
hash before this task changed its triage status. All historical bundles remain
unchanged. The new transcript contains 9 successful expectation checks:

| Existing package/check | Observed result |
| --- | --- |
| Build the unchanged visual verifier | Exit 0 |
| Jev controlled candidate protocol, ordinary historical verification | Exit 0; 152 hashes verified |
| Same candidate with `--require-live` | Exit 1: controlled package is not live evidence |
| Jev real candidate protocol NOT_RUN package with `--require-live` | Exit 1: no accepting verdict |
| Decisions real candidate protocol NOT_RUN package with `--require-live` | Exit 1: no accepting verdict |
| Jev controlled quality package, ordinary historical verification | Exit 0; captured-response replay |
| Same quality package with `-require-live` | Exit 1: complete real visual evaluation has not run |
| Jev real quality NOT_RUN package with `-require-live` | Same explicit rejection |
| Decisions real quality NOT_RUN package with `-require-live` | Same explicit rejection |

These are expected prerequisite rejections, not new feature tests or evidence of
a provider failure. There were **zero new real calls and zero new real Attempts**.
The processes inherited only ordinary build variables, no route credentials.

The Jev packages share candidate Hash
`sha256:2cec32a720f2a8ebea657ae59ba3b47b6c088f3025965862dee454d7f1bfa49d`.
That candidate identity is not a formal catalog identity. Formal support remains
bound to Jev 1.13; `ai/internal/testkit/supportmatrix/classifier.go` explicitly
rejects another model under its existing report identity.

## Preparation and boundary

The [failure inventory](13-failure-modes.md) was saved before implementation or
new tests. It records the already approved Classify seams, public command seams,
catalog/authorization/Hook behavior, model-specific report identity, corruption
rejections, budget ceilings, historical preservation and fresh formal-Hash runs.
All those new feature tests remain NOT_RUN; no acceptance checkbox was checked.

No runtime, catalog/version/Hash, support matrix/schema, frozen Pi, dependency,
public support claim, proposed ADR status or parent completion state changed.
Existing user edits and other untracked specs/ADRs are excluded from this task's
commit. Only the input ticket's triage/comment and new ticket-13 preparation
records are saved. No implementation exists to typecheck, run under race, or
exercise in a full-suite completion run; those checks remain pending for resumed
implementation. The existing verifier build and recorded command checks above
did run.

Preparation commit: `d344ea0`. The independent parallel
[Standards/Spec review](13-code-review.md) found zero actionable preparation
findings. The receipt checksum, all 16 recorded input hashes, 9 command outcomes
and local links were rechecked; this does not approve the pending feature work.

## Resume

Supply authentic candidate protocol and independent effect evidence for the
relevant route, or dedicated approved configuration to collect them. The guides
specify exact authorization records and isolated processes:
[candidate probe](../../ai/live/testdata/multimodal/README.md),
[visual evaluation](../../ai/examples/visualeval/README.md).
Credentials belong in secure local configuration, never chat or committed files.

Keep each run within its existing ceiling: protocol ≤12 calls/12 Attempts,
32 questions, 24 sources/32 occurrences, 8 MiB source/16 MiB expanded, 30 seconds
per call/180 seconds per process; effects use 12 fixed samples, ≤12 calls/12
Attempts, 48 questions/24 occurrences, the same image-byte ceilings, 30 seconds
per call/240 seconds per process. Both have zero retries. Do not rerun to improve
scores. The formal catalog still needs fresh separately authorized evidence at
its own Hash after admission and implementation; old candidate bundles stay intact.

Recheck the blocker from the repository root without credentials or network:

```sh
python3 .scratch/jev-1.14-multimodal-box/verify-11-evidence.py --require-live \
  .evidence/ticket11-sealed/20261016T115409.762160000Z/jev
go run ./ai/examples/visualeval -verify \
  .evidence/ticket12-live-not-run-e6548ae/jev -require-live
```

Both commands must currently exit 1 for the reasons above. Ordinary controlled
verification remains available for integrity checks; it cannot unlock admission.
