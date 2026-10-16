# Ticket 11 two-axis review

Fixed point: `80e8d36e4c838be1ffee23b66b499823a94476df`.
First implementation: `cf55cbd`. Standards and Spec ran in independent parallel
agents as required by the code-review skill; unrelated user work was excluded.

## Standards

First review: 2 findings, both P2.

- `multimodal_harness_test.go`: the source glob omitted nested testkit evidence,
  audit and provider code; a dirty nested change could borrow earlier provenance.
  Conflicts with AGENTS.md principle 9. Fixed by recursively sealing all local Go
  sources plus module and verifier files; the verifier requires key nested seals.
- `multimodal_report_test.go`: the finish timestamp/deadline check preceded the
  final audit and inventory. Slow finalization could exceed its declared budget.
  Conflicts with principles 5/9. Fixed by keeping the report non-accepting through
  expensive work, timestamping after it, checking the deadline during inventory
  and again after final publication; a late write revokes the accepting status.

No additional smell findings. Source-seal semantic corruption and a controlled
deadline shortened at final publication were first red, then green.

## Spec

First review: 1 finding, P1.

- `main_test.go`: legacy live configuration could execute before the multimodal
  test rejected mixed credentials when running without the exact filter. This
  violates “每进程只持有/请求一个路线的获准凭据”. Fixed by selecting isolated
  mode before loading legacy configuration, and rejecting a non-exclusive entry
  before `m.Run`. The unfiltered executable regression was first red, then green;
  it requires zero calls and attempts.

Other reviewed requirements matched the ticket, including compliant NOT_RUN
when this job has no explicitly approved real credentials.

## Final follow-up

Both independent reviewers reinspected `cf55cbd...59e2dbc` and found no remaining
issues or new gaps. Standards confirms nested source sealing and publication
deadline checks are fixed; Spec confirms isolated configuration and pre-`m.Run`
rejection close the cross-route path. Full validation used `59e2dbc` unchanged.

Standards: 2 initial P2 findings, fixed, 0 remaining. Spec: 1 initial P1 finding,
fixed, 0 remaining.

## Artifact-only follow-up

The final receipt/sealer/document delta was independently reviewed on both axes.
Spec reported 0 findings. Standards found one P2: top-level package metadata
could bypass the sealer's symlink check. An identical-byte NOT_RUN integrity
symlink reproduced acceptance before the fix, then explicit rejection after all
metadata reads used `safe_path`. The original file and bytes were restored.
Independent Standards reinspection: 0 remaining. No Go or primary verifier
source changed after the full/race runs. The sealer runs identically under normal
Python and `python3 -O`, and refuses to replace a differing existing receipt.

Final totals: Standards 3 findings across both review stages, all fixed (worst
initial P2), 0 remaining; Spec 1 initial P1, fixed, 0 remaining.
