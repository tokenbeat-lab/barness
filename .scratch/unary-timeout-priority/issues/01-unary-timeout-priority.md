# Investigate unary call-deadline / idle-timeout terminal race

Status: needs-triage

## Observed

During ticket 04's first full offline run, existing
`TestUnaryTransportTimeouts/body-call-deadline` returned
`deadline_exceeded/response`, `StopReason:error` with
`ReadIdleTimeout (100ms)`, while its assertion expected `StopReason:aborted`.
The fixture configures CallTimeout and ReadIdleTimeout to the same 100ms value.
The full run was concurrent with a focused race run; scheduling is a hypothesis,
not a proven production defect. Resource-release checks still passed.

Ten consecutive isolated reruns then passed without changing timeout runtime
or that existing fixture. The next complete run also passed. This is currently
an intermittent verification concern, not a confirmed cause or a reason to
broaden the Jev-image change into the shared transport.

## Evidence and next investigation

- Original full log: ../../jev-1.14-multimodal-box/04-logs/full.txt
- Repeated isolated log: ../../jev-1.14-multimodal-box/04-logs/timeout-recheck.txt
- Exact case: .evidence/barness-ai/20261016T050322.809983000Z/P07-E08-unary-timeout-body-call-deadline
- Fixture: ai/e2e/unary_lifecycle_test.go, body-call-deadline.
- Starting source: 490e2f901d711ded59df97f4156fa6ffb43d47f5.

Reproduce under scheduling pressure, determine the required winner when equal
absolute deadlines expire, then distinguish a fixture synchronization issue from
shared runtime classification. Add a public failing E2E before any correction.
Resolve this ticket only after deterministic evidence covers both timeout kinds
and confirms body/permit release. Do not simply weaken the terminal assertion.
