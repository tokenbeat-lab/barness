# Ticket 04 — experimental Jev JSON context images

Baseline: `490e2f901d711ded59df97f4156fa6ffb43d47f5`, current branch `main`.
This implements ticket 04 at the already agreed public `Client.Classify` /
`HookedClient.Classify` seam. All Providers here are controlled loopback HTTP;
real service execution and visual task quality are **NOT_RUN**.

## Delivered boundary

Host catalogs may explicitly declare fixed `jev-1.14.0`, text+image input,
`ImageLocations:[context]`, PNG/JPEG and finite image counts. The call also needs
Classifier image policy, `Binding.AllowClassifierImages` and
`TypeSafeOptions{ExperimentalImages:true}`. Defaults remain text only. Builtin
models, catalog hash, support matrices and frozen Pi inputs stay unchanged.

The adapter sends raw Base64 native JSON assets, one per source; context order
is the declared reference order. It wraps State as `{state:<original>,images:[]}`
without overwriting business fields or rewriting natural language paths. Private
keys use bounded collision checking against decoded strings/object keys and
stay fixed across retries. Upload and decoded expansion budgets are independent.
Jev's static PNG/JPEG, APNG, dimensions, decoded 16 MiB file, 32 MiB body and
key limits are independent of Decisions' existing WebP and small-image contract.
The service's processed file size and precise image token count are not predicted.

The State budget includes generated keys and wrapping; model image counts bound
sources, while reference policy bounds occurrences. Final native assets may omit
`type`; any supplied value must be `image`. The builder always emits it.

Final Hook output is independently frozen as native TypeSafe data. Legal source
addition/deletion/replacement/reorder is supported. Only whole keys in the context
array are supported; unused/overlapping keys, substring, part, URL and local uses
fail before Attempt. Final identities follow first actual asset use; independent
assets with identical files remain independent records. Answers and ImageInputs
publish together; errors retain validated usage/ResponseModel under existing rules.
Observer and shared metadata contain no keys, file hashes or image content.

State value images, question-local images, visual legends and Box remain for
later tickets. No public Assets type, dependencies, global asset cache, fallback
or new transport/retry lifecycle was added. Disabling image permission blocks new
calls while pinned calls finish. Code rollback needs no data migration.

## Test-first evidence

[Failure inventory](04-failure-modes.md) predates implementation. Each new behavior
was exercised at the public seam before its fix; no unit tests were added.
Logs under `04-logs/` preserve red/green projection, file/header, Hook, budget,
snapshot and borrowed-value checks. Some intermediate logs include fixture/config
or floating-point assertion corrections; they are not final PASS evidence.
The collision regression replaces only the external randomness source so escaped
text/object-key collisions and bounded duplicate selections are reproducible.

Public E2E coverage includes reference reuse, preserved State and exact numbers,
final asset mutation, absent/invalid authority and experiments, capacity, atomic
bad answer/usage, missing/null usage, credential reads, fixed retries, pre-entry
and in-response cancellation/deadlines, revocation and simultaneous two-tenant
calls. Existing Jev native/text, Decisions reference/identity and Pi regressions
remain part of final validation.

## Reproduce

```sh
go test ./... -count=1
go test -race ./ai/e2e -run '^Test(JevContext|TypeSafeNative|JevRegression|DecisionsReferences|DecisionsImageInputs)' -count=1
go vet ./...
python3 .scratch/jev-1.14-multimodal-box/verify-04-evidence.py
```

Each evidence case also records its exact replay command in assertions and the
manifest. Synthetic images, fixture SHA-256, public input, final requests,
response scripts, results/errors, Attempt/Observer, finite policy/catalog/Binding
and resource-release assertions use the existing evidence recorder. The verifier
checks every artifact hash and independently recomputes decoded-file identities
from actual sent native assets, atomic failures and private observations.

Independent review found and corrected four approved-design details: optional
native asset type; source counts versus reference occurrences; wrapped State
budget; and rejection before projected-State expansion. The final review has no
unresolved findings on either axis; see [two-axis report](04-code-review.md).

The first full run had one existing equal-100ms timeout terminal assertion fail.
Ten isolated repetitions passed unchanged, and the next full run passed. The
observation and investigation conditions are preserved in a [triage ticket](../unary-timeout-priority/issues/01-unary-timeout-priority.md); the Jev change does
not modify shared timeout runtime or weaken that assertion. Full runs after
review corrections preserve the original failure rather than replacing its log.

## Final sealed validation

Final runtime/test source commit: `724a0a7` (with `5615b02` and `7bfec2e`).

- Full `go test ./... -count=1`: exit 0, **6,938 PASS / 1,050 explicit opt-in
  NOT_RUN** evidence cases. All **137 ticket-04 cases PASS**.
- Focused race: exit 0, **463/463 PASS**, including all 137 ticket cases and the
  prior TypeSafe native/text plus Decisions reference/identity regressions.
- `go vet ./...`: exit 0.
- Full and race redaction audits: **zero findings**.
- Independent verifier: **47,764 full + 5,827 race = 53,591 artifact hashes**,
  final decoded-file identities, atomic errors, private observations and complete
  Go source inventory verified. Both bundles contain image fixtures/hashes,
  actual inputs/wire/scripts, results/errors, Attempt/Observer, finite config,
  assertions and released-resource gauges.

[Full manifest](../../.evidence/barness-ai/20261016T051029.530899000Z/manifest.json),
[race manifest](../../.evidence/barness-ai/20261016T051215.828696000Z/manifest.json),
[integrity inventory](04-evidence.json), [review report](04-code-review.md).
Final logs: `04-logs/full-sealed.txt`, `04-logs/race-sealed.txt`,
`04-logs/vet-sealed.txt`; intermediate red/green and prior runs remain preserved.

The inventory also hashes verification documents, logs, source and the audit /
manifest files. Re-run the verifier without arguments to check the sealed local
bundle. After an intentional fresh test run, `--record FULL_MANIFEST RACE_MANIFEST`
seals that new bundle, retaining independently verified identities and all hashes.
Offline PASS never implies live PASS; real-service protocol execution, visual
quality and formal support remain **NOT_RUN** / unclaimed.
