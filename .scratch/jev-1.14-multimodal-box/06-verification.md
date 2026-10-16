# Ticket 06 — Jev question-local instruction images

Baseline: `4adf8c2b4786a083e719cef57032f5b8e3cff9b3`, branch `main`.
Implementation: `60f33a8`; review correction: `5449e59`. Ticket 05 is resolved. Agreed seams are public
`Client.Classify` and `HookedClient.Classify`, against controlled loopback HTTP.
Real-service protocol, visual quality and formal-support claims are NOT_RUN.

Choice, Score and Bool Instructions accept root images, nested object/array
values and ordered mixed text/image fragments through explicit null-slot RFC
6901 references. Their scope stays local to that question; without context,
State keeps its business position. Ordinary text/nulls/objects/keys and exact
JSON number spellings remain intact. Root images are governed independently of
text DescriptionFormats; ordinary containers still obey those formats. Criteria
images and Box remain unimplemented/rejected. Decisions rejects any valid local
image request at capability before credentials, without splitting/conversion.

A source uploads once; each occurrence, including reuse across questions,
consumes finite reference/expanded-byte budgets. Source, single-file, projected
full-question, State and final native request budgets stay independent. The
Builder reserves escaped private-key growth before replacing instruction slots.
Final Hook validation freezes independent native questions and checks actual
whole-value image uses, authority, format and capacity again. Images are indexed
by first use in State, then sorted question names and their Instructions, with
sorted object keys and original array order. Hook renames/retyping/replacement
and additions from text are reflected in final answer and ImageInputs validation.
Failures clear Answers/ImageInputs together, retaining validated usage. Shared
observations carry no file bytes, hashes, private keys or reference paths.

[Failure inventory](06-failure-modes.md) was written before implementation. The
first 24 public projection cases failed on the unsupported location. Subsequent
red/green Hook tests corrected image-bearing per-question capacity errors to
resource_limit/request, keeping existing text-only error behavior. The Score
fixture was corrected to the established native map/legend/confidence schema;
a multipart-part rejection fixture was corrected to place the unsupported field
on the native asset rather than on an otherwise-valid business content object.
All intermediate red logs are retained as history, not final PASS evidence.
The pre-review race run was stopped after the Spec correction; its unsealed
bundle is superseded by the final corrected-source race run.
Independent Spec review then found that removing the last context occurrence
wrongly blocked still-used local images. A public Hook E2E reproduced that
failure before the correction; explicit empty arrays now preserve the builder
envelope while missing/null collections remain invalid. No private-helper unit
tests were added after implementation.

[Independent two-axis review](06-code-review.md): Standards 0 findings;
Spec 1 P2 corrected, 0 remaining after independent follow-up.

The 149 ticket cases cover both entry points for projection, invalid pointers,
nonempty content, formats/capabilities/auth, independent budgets, deterministic
ordering and source reuse, tenant/Actor denial, snapshot mutation, retry and
cancellation. Hook-specific cases cover question rename/retyping, local image
replacement/order, text-origin images, raw JSON, restricted locations, private
key misuse, unsupported resources, final capacity, stale/bad answers and atomic
usage preservation. Existing State/context/native/Decisions regressions remain
part of the final runs. All cases use the existing recorder: file fixtures and
SHA-256, input/configuration, response scripts, captured final requests, results,
errors, Attempt/Observer records, assertions and released-resource gauges.

Reproduce:

```sh
go test ./... -count=1
go test -race ./ai/e2e -run '^Test(JevInstructions|JevState|JevContext|TypeSafeNative|JevRegression|DecisionsReferences|DecisionsImageInputs)' -count=1
go vet ./...
python3 .scratch/jev-1.14-multimodal-box/verify-06-evidence.py
```

The independent verifier checks every full/race artifact hash and ticket
assertion, recomputes final decoded-image identities in deterministic order,
compares final question names/kinds, verifies error atomicity and resource
release, rejects private identity leaks into shared metadata, and seals Go/module
source hashes, verification docs/logs, issue and ADR. Both recorder audits must
have zero findings. To seal fresh completed runs use `--record FULL_MANIFEST
RACE_MANIFEST`; bundles remain local in ignored `.evidence/`.

Public usage and rollback are documented in
[the instruction images guide](../../docs/barness-ai/jev-instruction-images.md)
and [ADR-0021](../../docs/adr/0021-barness-ai-typesafe-unary-classification.md).
Disable the Binding/model or revoke classifier image permission to stop new
calls; already pinned calls complete with their snapshot. Code rollback has no
migration. Built-in catalog, support matrices, frozen Pi and unrelated user
README/GLOSSARY/specification edits are unchanged by this ticket.

## Initial full-run timing failure

The initial full run had 7,215 PASS, 1,050 opt-in NOT_RUN and one failure in
the pre-existing `TestUnaryTransportTimeouts/body-call-deadline`: CallTimeout
and ReadIdleTimeout are both 100ms, and it observed the idle terminal before
the expected aborted call terminal under concurrent test/audit load. All 149
ticket cases passed and that run's redaction audit had zero findings. No library
or test change was made for this unrelated timing result. The exact existing
case passed three consecutive isolated replays; the final full suite then passed
serially after race completion. `06-logs/full-first-failure.txt` and
`timeout-replay.txt` preserve the failure and replay. The failed bundle is
`.evidence/barness-ai/20261016T075630.934181000Z/manifest.json`; it is historical
evidence, not the final full PASS seal.

## Final sealed validation

- Full `go test ./... -count=1`: exit 0; **7,216 PASS / 1,050 explicit opt-in
  NOT_RUN** recorded acceptance cases. All **149 ticket cases PASS**.
- Focused race: exit 0; **741/741 PASS**, including all **149 ticket cases**.
- Final `go vet ./...`: exit 0.
- Full and race redaction audits: **zero findings**.
- The integrity inventory verifies **51,735 full + 9,798 race = 61,533 artifact
  hashes**, final decoded-file identities, final question names/kinds, atomic
  failures, private observations, assertions, released-resource gauges and
  complete Go/module source hashes.

[Full manifest](../../.evidence/barness-ai/20261016T080309.647611000Z/manifest.json),
[race manifest](../../.evidence/barness-ai/20261016T075732.063665000Z/manifest.json),
[integrity inventory](06-evidence.json), [review](06-code-review.md).
Final logs are `06-logs/full.txt`, `race.txt` and `vet-final.txt`. Both final
bundles identify source commit `5449e59`; source hashes continue to verify after
the documentation-only seal commit. The interrupted pre-review race bundle and
the first full timing failure are retained as history, outside the final seal.

Offline PASS proves only the controlled HTTP protocol/runtime boundary.
Real-service protocol, visual quality and formal support remain NOT_RUN /
unclaimed; built-in catalog, frozen Pi and support matrices are unchanged.
