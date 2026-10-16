# Ticket 08 — Box and independent existence

Baseline `8ecfdecb8a653e59199fa22378fa75a63939a38b`, branch `main`.
Dependency 06 is resolved. Public seams: Client.Classify / HookedClient.Classify
against the existing controlled loopback HTTP Provider. Real-service protocol,
visual quality and formal support remain NOT_RUN. Existing user README/glossary,
ADR-0030 and other issue/spec changes are excluded from this implementation.

The [failure inventory](08-failure-modes.md) preceded implementation. The first
public JSON tracer failed through both entry paths with unsupported Box. Typed
E2E scenarios were then written before implementation; the typed compiler red
is additional history, not a runtime evidence claim. Subsequent build attempts
caught unused imports and a case-ID index typo in the tests; those were corrected
without altering expected behavior. Final verification results follow below.

Tests cover all four kinds in a batch, null/zero/low/full confidence and
independent zero/low/full existence, missing/null/type/range/nonfinite/degenerate
edges, sealed fields, per-Box visibility, shared State/context and own local
images, escaped paths, independent input/catalog/callback snapshots, capability
and zero-pin denial, image authority/options/headers/budgets, legal four-kind
Hook changes, final image replacement and atomic failures with usage preserved.
Existing Jev and Decisions/refusal regressions are included in the focused race
and final full suite. No implementation-afterward unit tests were added.

The [public contract](../../docs/barness-ai/jev-box.md) and
[ADR-0021](../../docs/adr/0021-barness-ai-typesafe-unary-classification.md) explain
normalized edges, nullable confidence, independent existence, Host image
association/transforms, final visibility, authority and rollback. Revoking
Binding/model/image authorization stops new calls; pinned calls keep snapshots.
No database migration or compatibility path was introduced.

Reproduce:

```sh
go test ./ai/e2e -run '^TestJevBox' -count=1
go test -race ./ai/e2e -run '^Test(Jev|TypeSafeNative|Classifier|Decisions)' -count=1
go test ./... -count=1
go vet ./...
python3 .scratch/jev-1.14-multimodal-box/verify-08-evidence.py
```

The final integrity inventory seals artifact hashes, fixture files, inputs,
final requests/responses, outcomes/errors, Attempts/Observer, assertions,
configuration, Host reads, resources, source files and recorder redaction audits.
It independently checks decoded image identities and native/public Box equality,
visibility, confidence presence, final kinds/names, failure atomicity and secrecy.
Controlled HTTP PASS does not prove real-service protocol or visual accuracy.
