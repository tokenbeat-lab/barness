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

Final suite, race, redaction-audit and integrity inventory results are being
sealed after the test processes complete. The same inventory records the exact
Go sources used to reproduce the run; offline PASS never implies live PASS.
