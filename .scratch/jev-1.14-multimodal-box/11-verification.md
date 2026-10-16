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

Implementation commits: `cf55cbd`, corrected by `59e2dbc`. The
[independent two-axis review](11-code-review.md) initially found Standards 2 / Spec
1 issues; all three were fixed with red/green regressions. Final independent
reinspection: Standards 0 remaining / Spec 0 remaining.

Additional red/green slices covered an ambient OAuth token (zero-spend rejection)
and in-flight process deadline with terminal evidence. The initial Jev timeout
fixture exposed an httptest cleanup hang while its request body was unread;
the fixture now drains that bounded body and has an explicit cleanup stop channel.
These are harness failures, not evidence of vendor failures or runtime changes.

## Completed validation

`go test -tags live ./... -count=1`: exit 0. Recorded E2E + live cases total
**8,068 PASS / 1,112 explicit opt-in NOT_RUN**, including all **67 ticket-11
cases PASS**. Other Go package tests also pass. Full E2E manifest:
`.evidence/barness-ai/20261016T115408.429237000Z/manifest.json`; live manifest:
`.evidence/barness-ai/20261016T115409.762160000Z/manifest.json`.

`go test -race -tags live ./ai/live -run '^TestMultimodalCandidateProcess$'
-count=1`: exit 0, **67/67 PASS**, manifest
`.evidence/barness-ai/20261016T115411.218521000Z/manifest.json`.
`go vet -tags live ./...`: exit 0. All three recorder audits have zero findings.

The [integrity receipt](11-evidence.json) verifies **68,491 final artifact hashes**
plus original fixture and first-red assertion hashes. The existing recorder's
275 legacy failure-fixture instances intentionally hash their input before
redaction; their four original repository files were independently matched,
while all final redacted copies match their separate artifact checksums. No
fixture tolerance or assertion was relaxed.

For each full/race parent bundle, child Jev and Decisions success packages were
reconstructed from the retained raw bytes under
`.evidence/ticket11-sealed/<parent-run-id>/{jev,decisions}`. Each original report,
timestamp and inventory is preserved; existing differing files cannot be replaced.
The independent verifier accepts each controlled package (152 Jev / 78 Decisions
hashes each), and `--require-live` rejects all four. Parent bundles also preserve
every intentional configuration, budget, vendor, audit and manifest failure.

Actual probe use per batch: Jev 12 calls / 12 HTTP / 20 questions / 18 final
sources / 26 occurrences; Decisions 6 / 6 / 6 / 8 / 10. Reservations cover the
larger fixture file and State projection overhead; actual decoded bytes and
native State are separately reconciled. Every final Hook MIME/hash, Score legend,
Box, reported usage, Attempt/Observer and resource-release record is checked.

## Real routes: NOT_RUN

Two new sanitized processes under `59e2dbc` recorded real mode independently:

- Jev: `.evidence/ticket11-live-not-run/20261016T115446.169508000Z`.
- Decisions: `.evidence/ticket11-live-not-run/20261016T115446.169967000Z`.

Both reports say **NOT_RUN**, incomplete, zero Logical Calls / zero HTTP Attempts:
this job has no explicitly approved route credential/authorization record. Earlier
same-input packages remain intact. No other task's credentials were read or used.
Account/region/privileges remain unconfirmed Host attestations; billing, capacity,
visual quality/IoU/absence false positives and formal support remain unclaimed.

Reproduce from repository root:

```sh
go test -tags live ./ai/live -run '^TestMultimodalCandidateProcess$' -count=1
go test -race -tags live ./ai/live -run '^TestMultimodalCandidateProcess$' -count=1
go test -tags live ./... -count=1
go vet -tags live ./...
PYTHONDONTWRITEBYTECODE=1 python3 .scratch/jev-1.14-multimodal-box/seal-11-evidence.py
python3 .scratch/jev-1.14-multimodal-box/verify-11-evidence.py \
  .evidence/ticket11-sealed/20261016T115409.762160000Z/jev
```

The seal command rechecks these completed historical IDs; new test runs produce
new packages and must retain their own IDs/hashes. A real probe requires the
guide's isolated authorization and a separate `--require-live` verification for
each route. NOT_RUN and controlled PASS do not authorize catalog publication.

The final artifact-only review also closed a metadata symlink rejection gap in
the receipt sealer. Its public process accepted an identical-byte NOT_RUN
integrity symlink before the fix and rejected it afterward; history is retained
in `.evidence/ticket11-metadata-inventory-symlink-{red,green}.json`. The original
package file was restored exactly. The final sealer produces identical output
with normal Python and `python3 -O`; differing existing receipt bytes are refused.
Both final review axes report zero remaining issues. This artifact-only closure
does not change the Go/primary verifier sources tested under `59e2dbc`.
