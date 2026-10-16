# 09 — Mixed image runtime resource isolation

Baseline `58ec73c`, current branch `main`; dependencies 07 and 08 resolved.
The approved public seam is Client.Classify / HookedClient.Classify against the
controlled HTTP Provider. Real-service protocol, visual quality and formal
support remain NOT_RUN. Existing user README/glossary/ADR-0030 changes and other
open release issues are outside this commit.

The [failure inventory](09-failure-modes.md) preceded implementation. The first
three public Hook E2E cases failed at the post-callback allocation ceiling:
map 74,072,520; raw 72,477,424; typed 63,470,672 bytes. Red manifest:
`.evidence/barness-ai/20261016T090521.854551000Z/manifest.json`.
An embedded struct independently reproduced 65,554,584 bytes in
`.evidence/barness-ai/20261016T091346.193090000Z/manifest.json`. A later untagged
business-struct case exposed premature rejection of valid borrowed content:
`.evidence/barness-ai/20261016T091451.100123000Z/manifest.json`.
All E2E cases were written before their corresponding implementation repair;
no unit tests were added after implementation.

The repair accounts known final asset sizes and whole-value occurrences in
borrowed maps, RawMessage and typed/promoted struct fields before serialization.
The existing field selector now honors default JSON field names and tagged-field
precedence. Ordinary business strings, omitted/shadowed fields and byte-slice
encoding retain JSON semantics. Unknown custom encoder output still receives
the existing final native validation. Native schema/authority, image files,
placements, Box visibility and exact encoded bytes are independently rechecked.
No public interface, catalog, retry policy or resource limit was expanded.

The fixed load is one shared Client, two tenants, both protocols in each tenant,
four simultaneous calls, two 280×280 PNG/JPEG fixtures per call, repeated context
[1,0,1] and Jev State/Instructions/Criteria references. Four Jev answer kinds
include a valid Box alongside negative existence; Decisions returns three kinds.
HTTP EOF barriers prove all four bodies and permits are held simultaneously.
Normal/race use identical input, limits and 64 MiB allocation/live-heap ceilings;
measurements include the controlled HTTP recorder. Limits are test values, not
deployment defaults. The same workload is used for cancellation and recovery.

Coverage includes pre-send/body/admission/backoff cancel and deadline; late Host
grants; process/tenant waiting without releasing another owner's resources;
zero-default and finite retry, callbacks once, byte/header/key/credential/catalog
snapshot identity; post-2xx atomic failures retaining legal usage; independent
Binding/model/image revocation; deep ownership of inputs/config/answers/legends;
entry, native framing, final Hook, reference, single/total image and output
budgets; Observer error/panic/slow/congested delivery and content-free metadata.

Reproduce:

```sh
go test ./ai/e2e -run '^TestMultimodalRuntime' -count=1
go test -race ./ai/e2e -run '^Test(MultimodalRuntime|Jev|TypeSafeNative|Classifier|Decisions)' -count=1
go test ./... -count=1
go vet ./...
PYTHONDONTWRITEBYTECODE=1 python3 .scratch/jev-1.14-multimodal-box/verify-09-evidence.py
```

Each case records configuration, actual inputs/fixtures, final requests and
responses, result/error/usage/Attempts/Observer, assertions, resource counts and
memory measurements where relevant. Responses are correlated by final request
SHA-256; error scripts preserve HTTP status. The independent verifier seals
every artifact hash and source, recomputes final image identities, checks result
atomicity, metadata-only observations, resource release and both audits. The
immutable failed runs remain failure history, never substitute final PASS.

## Final validation

Pending final full suite, related race, independent two-axis review and seal.
