# Ticket 07 — Jev visual criteria and public legends

Baseline: `6c51ee83a270017ad0f1bcd54fb6dc9f98935b8d`, current branch `main`.
Ticket 05 is resolved. Seams: public Client.Classify / HookedClient.Classify
against controlled loopback HTTP. Real-service protocol, visual quality and
formal support are NOT_RUN. Built-in catalog, frozen Pi and support matrices
are unchanged. Existing unrelated user README/GLOSSARY/spec edits are excluded.

[Failure inventory](07-failure-modes.md) preceded implementation. The first 24
projection E2Es failed at the closed candidate boundary. After projection,
public legend E2Es failed on leaked native legend values and absent public
references, before implementing restoration. Subsequent boundary/Hook/lifecycle
cases cover all three kinds, root/nested/ordered/mixed positions, escaped names,
ordinary null/text, single Choice, independent source/occurrence/question/body
budgets, authority, restricted locations/formats and zero-pin Decisions denial.
Hook cases rename/retype/change levels/replace/reorder images, including distinct
assets with identical files and State-first indices; bad final echoes fail
atomically with usage preserved and no 2xx replay. Snapshot, retry, cancellation,
Actor/tenant and existing no-image/wide-JSON/rounding/refusal regressions remain.

Intermediate logs are history, not final PASS seals. During projection the
private question decoder had to invoke its own strict UnmarshalJSON instead of
applying generic tagged-struct validation to its private fields. The nested
legend fixture comparison was corrected to UseNumber because the general E2E
helper cannot represent huge exponents; the expanded-byte fixture was corrected
to actually exceed its configured limit. No implementation-afterward unit tests
were added.

[Public contract](../../docs/barness-ai/jev-criteria-images.md) and
[ADR-0021](../../docs/adr/0021-barness-ai-typesafe-unary-classification.md) document
scope, deterministic traversal, final image identity, result-only hashes, public
Legend pointers, authority and rollback. Code rollback has no data migration;
revoking the Binding/model/image allowance stops new calls, while pinned calls
finish from their snapshots.

Reproduce:

```sh
go test ./... -count=1
go test -race ./ai/e2e -run '^Test(JevCriteria|JevInstructions|JevState|JevContext|TypeSafeNative|JevRegression|DecisionsReferences|DecisionsImageInputs|ClassifierNativeLegend)' -count=1
go vet ./...
python3 .scratch/jev-1.14-multimodal-box/verify-07-evidence.py
```

The verifier seals every artifact hash in full/race manifests, all Go/module
sources and this ticket's verification files. It independently rebuilds final
image MIME/file hashes and native-to-public legend nulls/pointers/indices,
checks atomic failures, final question names/kinds, shared-observation secrecy,
resource release and zero-finding recorder redaction audits. To seal completed
runs: `--record FULL_MANIFEST RACE_MANIFEST`. Bundles stay in ignored `.evidence`.
Final validation and independent review results will be recorded below.
