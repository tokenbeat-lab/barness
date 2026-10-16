# Ticket 13 — fixed Jev 1.14 formal support delivered

Verified 2026-10-17 (Asia/Shanghai), branch `main`. Implementation commits:
`527f633`, `e633b3f`, `1721024`. Runtime/source proof commit:
`17210241b7d2c35cd9918d3ee2499c4031504939`. Ticket 13 is complete for the fixed Jev route;
whole-release acceptance, other current-Hash route evidence and parent completion
remain ticket 14. ADR-0030 and the parent specification were not marked complete.

## Admission before implementation

The initial prerequisite check honestly recorded BLOCKED/NOT_RUN, zero calls and
zero Attempts in `d344ea0` / `75ae590`; its immutable
[receipt](13-prerequisite-evidence.json) and original ticket input are retained.
The original preparation-only note/review remain in those commits.

The user then explicitly supplied root `.env:TYPESAFE_KEY`. Only this selected
key was passed to isolated processes for the fixed official Jev endpoint; no
other key or route was used. [Admission receipt](13-admission.json) seals fresh
real candidate protocol PASS (12 calls/12 Attempts) and independent effect
COMPLETE (12 fixed samples/12 calls/12 Attempts), zero retries, before catalog
implementation. Candidate Hash remains
`sha256:2cec32a720f2a8ebea657ae59ba3b47b6c088f3025965862dee454d7f1bfa49d`.
Those packages and earlier controlled/NOT_RUN packages were not rewritten.

Account, region and provider permissions remain explicitly unverified.
`dedicatedCredential` attests the isolated selected-key process, not provider
account exclusivity or privilege introspection. The selected key is absent from committed task
files and all accepted raw packages passed their own redaction audits.

## Delivered behavior

Catalog `2026-10-17.1` adds only fixed `jev-1.14.0` and retains 1.13 unchanged.
Formal Hash: `sha256:ae3315ce504af2019ce6476f3b021093e13428a5992b3e5fc7d3208576eb49c7`.
Four image locations (context, State, own-question Instructions, Criteria),
static PNG/JPEG and Box use the existing runtime, explicit experimental selection,
Binding whitelist/permission and finite policies. The catalog's 24 sources is a
bounded barness subset, not a vendor count claim. Moving aliases are excluded.
Pinned primary-source limits/rates and distinctions between source/expanded/
processed capacity live in [sources.json](../../ai/live/testdata/jev-formal/sources.json).
Input is estimated at $0.042/M reported tokens, output free; this is not billing.

`typesafe-classifier-jev-1.14.0` has its own twelve mandatory scenarios, six each
through Client.Classify and HookedClient.Classify. The old `typesafe-classifier`
validator remains fixed to 1.13. Schema 3 adds an optional route-specific `jev`
member with all producers/consumers updated together; no runtime compatibility
shim, dual write or historical schema rewrite was introduced. Formal protocol,
independent effects and current source/catalog seals are required together.

Public file-based producer/verifier/merge reject identity, capability, budget,
capture, source, audit and inventory corruption. Review corrections add exact
JSON member spelling, pre-open regular-file checks, public Classify replay of
all captured protocol responses, normalized Box checks and independent mandatory
scope/kind validation. An accepting report cannot be established by saved PASS
assertions or refreshed checksums alone. Controlled evidence verifies only as
controlled and cannot merge as live support; a rejected batch never partly writes.

The executable classifier image example now uses the actual built-in model.
Current Chinese text evaluation configs and their release consumer pins advance
together without changing their fixed task, 1.13/Luna identities or prices.
Historical raw text evaluations still verify with their own recorded snapshots;
current configs do not promote those old results to a new-Hash PASS.

## Fresh real formal-Hash evidence

Self-contained strict-verified package:
`.evidence/ticket13-live-formal-20261016T230251.995980Z/support`.
Protocol original: `.evidence/ticket13-live-formal-20261016T230251.995980Z/protocol/20261016T230253.926005000Z`.
Independent effects original:
`.evidence/ticket13-live-formal-20261016T230251.995980Z/effects`.
Copies in the support package preserve the original subpackage bytes.

| Run | Outcome | Calls / Attempts | Questions | Image resources |
| --- | --- | --- | --- | --- |
| Formal protocol, both public entries | PASS, complete/audit clean | 12 / 12 | 20 | 18 sources, 26 occurrences; 44,784 decoded / 64,688 expanded bytes; 1,670 State bytes |
| Independent formal visual evaluation | COMPLETE, realQuality COMPLETE | 12 / 12 | 28 | 16 occurrences; 27,959 decoded / 41,603 expanded bytes |

Both retain zero retries, 30-second per-call bounds and the original per-round
ceilings: protocol 180 seconds including 10 seconds finalization, ≤12 calls/
Attempts, 32 questions, 24 sources/32 occurrences, 8 MiB source/16 MiB expanded/
4 MiB State; effects 240 seconds including 15 seconds finalization, 12 fixed
samples, ≤12 calls/Attempts, 48 questions/24 occurrences, same image-byte ceilings.
No sample was rerun for a higher score. The protocol recorded 9,875 input / 372
output tokens; effects 10,024 input / 716 output, estimated $0.000421008.
Across admission and formal rounds: **48 real calls, 48 Attempts, zero retries**.

Common classification/existence: 8/8 each. Local-scope classification/existence:
4/4 each. Box: 2 present / 2 absent samples; present mean IoU
`0.701974024586969`; 2 raw boxes on absent samples, 0 consumed absent false
positives after the independent existence gate. These are fixed synthetic task
results, not production accuracy, existence inferred from Box confidence, or a
cross-model quality ranking. Raw boxes and all denominators remain retained.

The strict protocol, effect, outer producer, outer verifier and public matrix
merge commands all exited 0. Only the new 1.14 row advanced. Removing that row
leaves the exact original ten rows as JSON values, including states and timestamps.
Original raw reports retain their own catalog Hashes.
Their historical PASS does not become current-catalog PASS. Decisions real image
protocol/effects remain NOT_RUN; no Decisions credential or real call was used.

## Red/green, regression and independent review

[Failure inventory](13-failure-modes.md) predates implementation. Public E2E red
bundles are sealed in [final receipt](13-evidence.json): catalog/discovery,
formal effects entry, matrix identity, manifest/duplicate-field handling,
resource reconciliation, alias/FIFO/correlated Box and missing mandatory State
scope. The first full run exposed stale examples/config/pins; its failed manifest
and log remain sealed. Focused public E2Es then passed before the final full run.
No post-implementation unit tests were added.

| Final check at `1721024` | Result |
| --- | --- |
| `go test -tags live ./... -count=1` | Exit 0; primary E2E 8,011 PASS / 1,050 NOT_RUN; live harness 204 PASS / 62 NOT_RUN |
| `go test -race -tags live ./ai/live -run '^TestJevFormal' -count=1` | Exit 0; 26 ticket-13 PASS |
| `BARNESS_EXAMPLE_RACE=1 go test -race ./ai/e2e -run '^TestJevFormal' -count=1` | Exit 0; 21 ticket-13 PASS, including race-built effect child |
| `go vet -tags live ./...` | Exit 0 |
| Independent Standards review | 0 remaining findings |
| Independent Spec review | 0 remaining findings |

Normal full execution has 47 ticket-13 PASS; both focused race packages repeat
those 47 cases. NOT_RUN covers opt-in Pi comparisons, pressure imports/runs,
scaling/baselines and unconfigured other live routes. It is not counted as PASS.
No whole-release gate PASS is claimed. [Review report](13-code-review.md) retains
the initial findings, their public red/green proof and separate follow-ups.

## Artifact receipt and replay

[13-evidence.json](13-evidence.json) SHA-256:
`c083d2c9ea20f02b21732f3c5be1e209aad25add72b3afaceda4ef8c3d89c214`.
It seals four final audited runs, 67,431 stored-byte case artifacts, immutable
admission receipt, red manifests, command transcripts, fixture pins and the fresh
real support package's full integrity chain. Public requests, final native
requests/responses, images and fixture hashes, atomic results/errors,
Attempt/Observer, assertions, host configuration, inventory and audits are retained.

The existing generic recorder fingerprints fixture input before secret redaction;
275 legacy fixture fingerprints differ from their retained redacted bytes. The
post-redaction `artifacts` inventory is verified for every stored file. All new
UC13 image/fixture fingerprints match both identities. No historical recorder or
bundle was rewritten. Raw process console logs are retained separately from the
accepted package audits; the receipt distinguishes their hashes.

From the repository root at the recorded source snapshot, without route keys:

```sh
python3 .scratch/jev-1.14-multimodal-box/seal-13-evidence.py \
  --verify .scratch/jev-1.14-multimodal-box/13-evidence.json
go run ./ai/live/cmd/jevformal -verify \
  .evidence/ticket13-live-formal-20261016T230251.995980Z/support -require-live
```

These perform retained-byte verification and captured-response replay only.
Changing runtime/verifier source requires fresh applicable evidence; historical
packages remain verifiable in their recorded checkout. Integrity receipts are
not origin signatures: preserve trusted raw live packages. Rollback disables the
explicit 1.14 Binding/model authorization; it does not fall back to 1.13 or erase
usage/history. See [formal guide](../../ai/live/testdata/jev-formal/README.md) and
[ADR-0031](../../docs/adr/0031-jev-model-specific-formal-evidence.md).
