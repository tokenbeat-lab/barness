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

## Completed review and validation

[Independent two-axis review](08-code-review.md) of implementation `443fc30`:
Standards 0 findings; Spec 0 findings. No implementation corrections required.
Verifier/review commit `d3466a9` does not change Go sources.

Final `go test ./... -count=1`: exit 0; **7,647 PASS / 1,050 explicit opt-in
NOT_RUN**, including **246/246 ticket cases PASS**. Full manifest:
`.evidence/barness-ai/20261016T085037.852408000Z/manifest.json` (Git `d3466a9`).
Final related race: exit 0; **2,436 PASS / 225 explicit opt-in NOT_RUN**, including
**246/246 ticket cases PASS**. Race manifest:
`.evidence/barness-ai/20261016T084314.101695000Z/manifest.json` (Git `443fc30`).
Final `go vet ./...`: exit 0. Both recorder redaction audits have zero findings.

[Integrity inventory](08-evidence.json) independently verifies **57,733 full +
25,701 race = 83,434 artifact hashes**, and the two first-red assertion artifacts.
Each final bundle includes 234 recorded public calls plus 12 catalog/closed-JSON
cases. Box values and explicit confidence exactly match the controlled response;
final native questions determine kinds, names, visibility and image identity.
Illegal responses publish no answers/image metadata; valid usage survives.
Shared observations omit private image keys/hashes, and all recorded resources
are released. All Go/module sources and ticket validation files are sealed.
Capability-denial configurations for the legacy text Jev/Decisions fixtures also
remain reproducible from the sealed existing Host fixture/configuration sources.

Seal completed runs with:

```sh
PYTHONDONTWRITEBYTECODE=1 python3 .scratch/jev-1.14-multimodal-box/verify-08-evidence.py --record \
  .evidence/barness-ai/20261016T085037.852408000Z/manifest.json \
  .evidence/barness-ai/20261016T084314.101695000Z/manifest.json
```

Bundles remain local in ignored `.evidence`; each manifest contains single-case
replay commands. Offline PASS is solely controlled HTTP behavior. Real-service
protocol, visual quality and formal support remain NOT_RUN/unclaimed. Built-in
Jev/Decisions catalogs, release snapshot, support matrix and frozen Pi are unchanged.

The final incremental two-axis review covered all added verifier/evidence/ticket
changes through `b7bbd45`: Standards 0, Spec 0. The Spec reviewer independently
reran the sealed verifier successfully; no implementation/test changes were
needed. Subsequent documentation-only review records leave verified Go sources
unchanged. The inventory was resealed after recording this final review.
