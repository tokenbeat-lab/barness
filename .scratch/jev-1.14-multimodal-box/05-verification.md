# Ticket 05 — Jev State value images

Baseline: `641fc8475dbede0bff59cb7c04b04982cfdddb13`, current branch `main`.
Runtime/test source: `6793760`, following `3a6fb7b` and `a79cbe3`.
Ticket 04 is resolved with its own sealed evidence. This ticket uses the agreed
`Client.Classify` / `HookedClient.Classify` seam and controlled loopback HTTP.
Real-service protocol execution and visual quality are **NOT_RUN**.

## Delivered boundary

An explicit fixed Host candidate `jev-1.14.0` can declare `state`, with or without
`context`, for PNG/JPEG images. Public RFC 6901 value pointers select existing
null slots in State, including its root, nested objects and arrays, escaped,
numeric and empty object keys. Only those slots receive private whole-value
asset references. Ordinary nulls, text, template-looking objects, business
state/images fields, JSON number spellings and question/candidate names survive.
One source uploads once even when reused; equal files in distinct assets remain
distinct final records. Borrowed State/references/sources/options are copied
before external resolution.

No-context requests keep the State tree at its original native location. Mixed
requests resolve public pointers before wrapping, preserving the complete
business State beside the ordered context image array. Questions share this
single State; their count does not multiply image occurrences. Source limits,
occurrence/expanded decoded-byte limits, projected State and final native body
limits apply independently. The projected State reserves generated escaped keys
and optional wrapper framing before expanding occurrences.

Hook final validation scans actual whole-value references, not the old pointer
table. It rebuilds authorized usage, occurrence budgets and final ImageInputs
in deterministic native State order (sorted object keys, original array order;
context envelope images before state). Reordered/replaced/added/deleted assets
and moved values are validated independently. Substrings, object/identity keys,
question-local private references and unused assets fail before Attempt.
Builder envelope topology stays separate from user content. For text-origin
Hooks, a State-capable model scans the entire business tree; a context-only
model requires a strict explicit native envelope referencing real assets.

All image policy, Binding, catalog location/MIME/count and experiment authority
still apply. Decisions rejects an entire valid State value request at capability,
before credentials; context-only Jev models reject State references. Metadata
does not advertise Instructions, Criteria or Box. Builtin models/catalog hash,
formal support, support matrices and frozen Pi remain unchanged. No dependency,
public abstraction, persistent state or migration was added; rollback reverts
these code commits. The existing retry, cancellation, usage and Observer runtime
remains shared, with fixed final bytes/headers/private keys across retries.

## Test-first and review evidence

[Failure inventory](05-failure-modes.md) predates implementation. Initial public
projection E2Es failed because State capability was unsupported, then passed
after the native projection/guard was implemented. The next red/green slice
fixed Hook addition of a first State image. A final independent Spec review
found business-envelope ambiguity; two public regressions reproduced it before
the fix. No private-helper unit tests were added. Intermediate logs retain
fixture/config/assertion corrections and are not final PASS evidence.

The final public cases cover both entries for projection/root/mixed requests,
bad pointers/plans, unsupported protocols/positions, independent budgets,
literal paths, snapshot mutation and retry/atomicity. Hook-specific cases cover
legal mutations, unsupported/private-key uses, text-origin State/context and
raw payloads. Catalog boundary cases keep future locations rejected. Existing
Jev context/text, TypeSafe native, Decisions reference/identity and Pi regressions
are included in final validation. All cases use the established recorder for
input, file fixtures/hash, final request/response script, result/error,
Attempt/Observer, finite configuration, credentials and released-resource gauges.

[Two-axis review](05-code-review.md): Standards 0 findings; Spec 1 P2 corrected,
0 remaining. Shared observations exclude image bytes, content hashes, private
keys and reference paths. Results publish Answers/ImageInputs atomically;
validated usage survives answer failure, and 2xx failure does not replay.

## Reproduce

```sh
go test ./... -count=1
go test -race ./ai/e2e -run '^Test(JevState|JevContext|TypeSafeNative|JevRegression|DecisionsReferences|DecisionsImageInputs)' -count=1
go vet ./...
python3 .scratch/jev-1.14-multimodal-box/verify-05-evidence.py
```

Each recorded case includes its replay command. The independent verifier checks
all artifact hashes, assertions and released-resource gauges, recomputes decoded
file identities in final wire traversal order, rejects unused/partial identity
records and checks private keys/hashes are absent from shared observations.
It seals Go source/module hashes, logs, this report, review, failure inventory,
ADR and issue alongside full/race manifests and zero-finding redaction audits.
After an intentional fresh run, use `--record FULL_MANIFEST RACE_MANIFEST` to
seal that bundle. Evidence files stay local under ignored `.evidence/`.

## Final sealed validation

- Full `go test ./... -count=1`: exit 0; **7,067 PASS / 1,050 explicit opt-in
  NOT_RUN**. All **129 ticket-05 cases PASS**.
- Focused race: exit 0; **592/592 PASS**, including all **129 ticket cases** and
  the prior context/native/text/Decisions reference and identity regressions.
- `go vet ./...`: exit 0 after the final correction.
- Full and race redaction audits: **zero findings**.
- Integrity inventory covers **49,614 full + 7,677 race = 57,291 artifact hashes**,
  decoded final-file identities, atomic errors, private observations, assertions,
  released-resource gauges and the complete Go/module source inventory.

[Full manifest](../../.evidence/barness-ai/20261016T053519.766947000Z/manifest.json),
[race manifest](../../.evidence/barness-ai/20261016T053738.830029000Z/manifest.json),
[integrity inventory](05-evidence.json), [review](05-code-review.md).
Final logs are `05-logs/full.txt`, `05-logs/race.txt` and
`05-logs/vet-final.txt`. The earlier full run passed before the Spec correction;
`full-before-review.txt` and all red/intermediate logs remain intact. The final
two manifests both identify source commit `6793760`; source hashes verify the
same Go/module files after the documentation-only seal commit.

Offline PASS applies to the controlled HTTP protocol/runtime boundary only.
Real-service protocol, visual quality and formal support remain **NOT_RUN** /
unclaimed, with no paid request or catalog promotion.
