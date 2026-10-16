# Ticket 11 candidate protocol verification

Baseline `80e8d36e4c838be1ffee23b66b499823a94476df`, current branch `main`.
Dependencies 07/08 are resolved. Existing user documentation/spec/ADR changes
are preserved and excluded from this implementation's commits.

The [failure inventory](11-failure-modes.md) was written before the probe.
The first process E2E failed because no probe report existed:
`.evidence/ticket11-first-red/ticket11-red/20261016T114005.681943000Z/manifest.json`.
Public probe execution then exposed two invalid per-call policy sizes; these
were fixed while preserving the separately defined cumulative budgets.
The independently read bundle tracer subsequently rejected the initial accounting
and exposed Go HTML escaping in the native State byte count. Final counters use
the same explicit wire-independent counting convention, documented in the guide.

The semantic corruption tracer failed before its verifier fix: even after
refreshing artifact and manifest checksums, a changed Score legend index passed.
Its preserved red assertion is
`.evidence/barness-ai/20261016T114522.551049000Z/MM11-tamper-jev-legend-index/assertions.json`.
The verifier now independently restores the final Criteria and checks its public
legend null slots and references into final MIME/hash metadata. No unit tests
were added after implementation; validation stays at the public process, both
Classify entries, HTTP and package seams.

Implementation/replay/authorization/budget documentation:
[probe guide](../../ai/live/testdata/multimodal/README.md).
Each route keeps an explicit pinned Host candidate and independent process/key.
No production runtime API/dependency, built-in catalog, formal matrix, frozen Pi,
ADR status or visual-quality conclusion is changed. Rollback removes the opt-in
probe; no migrations or compatibility execution paths are introduced.

Validation and independent review results will be appended after completion.

Additional red/green slices covered an ambient OAuth token (zero-spend rejection)
and in-flight process deadline with terminal evidence. The initial Jev timeout
fixture exposed an httptest cleanup hang while its request body was unread;
the fixture now drains that bounded body and has an explicit cleanup stop channel.
These are harness failures, not evidence of vendor failures or runtime changes.
