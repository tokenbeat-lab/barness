# Ticket 12 independent visual quality evaluation

Baseline `ba68f0cf962941a26307dfd334b80887026dbbdc`; current branch `main`.
Dependencies 10/11 were resolved before implementation. Existing user changes to
GLOSSARY, READMEs, draft spec/design and proposed ADR were preserved and excluded
from this implementation's commits.

Implementation: `d4cd265`; initial review corrections: `f7c9bcf`; complete source
set verification: `e2c8287`; official-path replay: `e6548ae`. No protocol runtime, built-in catalog, support matrix,
frozen Pi or dependency version changed. The opt-in Host example is the rollback
boundary; no migrations or compatibility paths were introduced.

[Executable guide](../../ai/examples/visualeval/README.md),
[failure inventory](12-failure-modes.md),
[independent two-axis review](12-code-review.md),
[integrity receipt](12-evidence.json).

The Host owns fixed Chinese annotations/questions, independent predictions,
classification/existence metrics and Jev local/Box metrics. The first eight
public requests are identical across routes except Target/Options; Jev's last
four local/Box tasks are separate from Decisions' supplemental context tasks.
Every route retains all 12 samples and every failure/refusal/unexecuted result.
Raw present-target IoU is retained even below the Host's independent Bool gate;
absent raw boxes and consumed absent false positives are reported separately.
No universal accuracy threshold or label changes/repeated sample selection were
used. Scripted predictions intentionally include errors and prove scoring only.

The 24 original/processed image files were regenerated independently in a fresh
temporary directory using `scenes.json` + `generate.go`; every byte hash matched
`pins.json`. Annotation coordinates remain an independent dataset file; drawing,
request construction and the controlled Provider never read those labels.
Every package includes candidate/model/catalog, time, source, fixture/config
pins, raw public inputs, final requests/responses, result/error, predictions,
labels, metrics, Attempt/Observer, release assertions, audit and inventory.

## Red/green and review evidence

The [failure inventory](12-failure-modes.md) preceded implementation. The initial
command tracer failed because no visual evaluation executable existed:
`.evidence/barness-ai/20261016T120613.074112000Z`. That initial missing-command
diagnostic included the checkout path and therefore failed its redaction audit;
it is diagnostic history, not an accepted evidence package. Subsequent semantic
red tracers and all final packages use sanitized diagnostics.

Public process red/green slices use both Client.Classify and
HookedClient.Classify, never implementation-coupled unit tests. The short-timeout
tracer exposed AdmissionWait > CallTimeout; the Host now bounds both together.
A Jev native `refusal` remains an honest protocol failure (that protocol does not
support that response variant); Decisions' valid per-question refusal is retained.

Precise, resealed semantic reds:

- `.evidence/barness-ai/20261016T121526.292313000Z`: controlled-to-live promotion
  without authorization and sample time beyond run end.
- `.evidence/barness-ai/20261016T121647.173955000Z`: hiding captured calls as NOT_RUN
  and replacing required assertion names.
- `.evidence/barness-ai/20261016T121730.048400000Z`: credential echo survives binary
  capture encoding; capture now redacts consumed bytes before Base64 serialization.
- `.evidence/barness-ai/20261016T122208.091449000Z`: output overflow must preserve
  the extra consumed byte needed to replay resource_limit.
- `.evidence/barness-ai/20261016T122209.044144000Z`: contradictory Observer terminal
  error/usage must not pass a refreshed inventory.
- `.evidence/barness-ai/20261016T122235.198781000Z`: both cancelled routes cannot
  invent 900 tokens/$123 from an empty response.
- `.evidence/barness-ai/20261016T122811.258504000Z`: replacing a required source with
  an unknown empty file must fail even with identical map size and resealed hashes.

The final official-path tracer additionally preserved `/v1/systemone` and
`/v1/decisions` through both entries, first exposing the replay mismatch and then
passing with the fixed official base path.

The independent Standards and Spec agents confirmed fixes with their own fresh
controlled probes. Final independent reports: Standards 0 / Spec 0 remaining.
Only captured-response replay and loopback HTTP were used; no real key/network.

## Final validation

`go test -tags live ./... -count=1`: exit 0 under `e6548ae`. Recorder totals:
**8,168 PASS / 1,112 explicit opt-in NOT_RUN**, including **100/100 ticket-12
cases PASS**; other Go package tests also passed. E2E manifest:
`.evidence/barness-ai/20261016T123641.786370000Z/manifest.json`; live manifest:
`.evidence/barness-ai/20261016T123643.659440000Z/manifest.json`.

The focused actual-process race command below exited 0: **70/70 PASS**, manifest
`.evidence/barness-ai/20261016T123644.632854000Z/manifest.json`. It instruments the
Host executable as well as the parent E2E and covers both routes/entries,
failures/refusals, budgets, authorization, deadlines, release and accounting.
The complete sequential filesystem-corruption matrix is included in the normal
100-case run. All three final parent redaction audits have **zero findings**.
`go vet -tags live ./...`: exit 0 on the final implementation.

The final receipt verifies **144 Host packages plus 2 independent real NOT_RUN
packages**, including **8 deliberately corrupted packages rejected for their
specific expected reason**, and **47,704 artifact hashes**. Accepted packages replay
through the public runtime with no network. All 24 regenerated image hashes and
all fixture pins match.

Earlier implementation/test runs remain diagnostic history; final acceptance
uses the exact source snapshot under `e6548ae`, including official base paths.

## Real routes: NOT_RUN

New isolated real-mode processes under `e6548ae`:

- Jev: `.evidence/ticket12-live-not-run-e6548ae/jev`.
- Decisions: `.evidence/ticket12-live-not-run-e6548ae/decisions`.

Both contain 12 preserved samples, NOT_RUN/null accuracy and zero Logical Calls /
zero Attempts. This task has no explicit approved route credential/authorization.
Only ordinary build environment variables were inherited; no other task's key
was read or used. Both packages pass ordinary verification and cannot satisfy
`-require-live`. Neither controlled COMPLETE nor candidate protocol evidence is
real visual quality. The ticket-11 independent receipt is copied/hash-linked to
its exact route/candidate catalog; neither old candidate nor the other protocol
is promoted to a new catalog Hash's effect evidence.

## Repeatability

From repository root:

```sh
go test ./ai/e2e -run '^TestVisualEvaluation' -count=1
BARNESS_EXAMPLE_RACE=1 go test -race ./ai/e2e -run '^TestVisualEvaluation(FixedWorkflow|FailureRetention|CancelledAccounting|SourceSet|Authorization|NotRun)$' -count=1
go test -tags live ./... -count=1
go vet -tags live ./...
go run ./ai/examples/visualeval -verify <new-run>/visual-jev-plain
go run ./ai/examples/visualeval -verify <new-run>/visual-decisions-hooked
```

Use a checkout matching the recorded source snapshot for historical verification.
The receipt sealer invokes the command's public verifier without network for each
Host package; intentional cancellation-accounting/source-set corruption packages
must fail for their specific expected reason. It verifies every selected ticket
artifact and each complete Host inventory, audit receipt, fixture pin and source
snapshot. Output receipts are append-only: differing existing bytes are refused.
This is integrity evidence, not a digital origin signature.

Seal/recheck these completed final runs (new test runs get new identifiers):

```sh
python3 .scratch/jev-1.14-multimodal-box/seal-12-evidence.py \
  --run .evidence/barness-ai/20261016T123641.786370000Z \
  --run .evidence/barness-ai/20261016T123643.659440000Z \
  --run .evidence/barness-ai/20261016T123644.632854000Z \
  --live .evidence/ticket12-live-not-run-e6548ae/jev \
  --live .evidence/ticket12-live-not-run-e6548ae/decisions \
  --out .scratch/jev-1.14-multimodal-box/12-evidence.json
```
