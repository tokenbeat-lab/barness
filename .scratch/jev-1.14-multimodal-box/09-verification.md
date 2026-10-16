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

[Independent review](09-code-review.md) found additional ordinary JSON encodings
at the same boundary. Typed/numeric-key asset maps and byte-slice data now retain
known decoded sizes; json.Number remains numeric unless its declared field uses
`,string`. Quoting follows declared field types, including pointer/interface
rules. Ordinary byte slices become Base64 content, and invalid UTF-8 is coerced
per byte exactly like the JSON encoder. Bounded raw strings use the final
decoder's Unicode behavior, including unpaired surrogates. Field resolution is
once per type per content pass, rather than quadratic per-field lookup.

Review red histories include `20261016T091926.968515000Z` (typed assets, integer
keys, wide fields, ambiguity and numeric content), `20261016T092616.635786000Z`
(quoted primitives/JSON-only numeric keys), `20261016T093156.841686000Z`
(Base64 content/UTF-8), `20261016T093259.189528000Z` (raw surrogate) and
`20261016T093429.413111000Z` (decoded byte-slice asset data). The interface
finding first failed in the reviewer's public overlay at
`20261016T092756.694222000Z`; the parent independently replayed `259ed61` through
an explicit Go overlay after repair at `20261016T092927.278425000Z`. This latter
run is regression reproduction, not a backdated first-red claim.

The first full run at `20261016T091701.294713000Z` failed the pre-existing
`P10-OD42-runtime-timeout-body-call-deadline`: CallTimeout and ReadIdleTimeout
both equal 100ms, and the observed terminal was idle/error instead of the
expected call/aborted. The original [timeout-ordering issue](../barness-ai-timeout-ordering/issues/01-call-and-read-idle-deadline-order.md)
remains needs-triage, as does the related [unary priority investigation](../unary-timeout-priority/issues/01-unary-timeout-priority.md).
Neither test nor shared timeout runtime was changed. The failed bundle and its
audit are sealed; later PASS does not establish a cause or repair.

The full `e429932` run at `20261016T093522.829653000Z` reproduced that tracked
symptom in `UC09-cancel-luna-plain-body-deadline`: the new fixture had also used
`fitCallTimeout`, making both limits 250ms. Its original failure and audit are
sealed. Test-only `f97648f` specifies an explicit 250ms Host deadline with the
original finite policy (30s call, 10s read idle); inputs, concurrency, case count
and required aborted terminal remain the same. Both reviewers approved this
separation from the unresolved coincident-timer question. This correction does
not establish a cause or repair for either original timeout issue.

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
go test -race ./ai/e2e -run '^Test(MultimodalRuntime.*|JevContextBorrowedImageBudget|DecisionsImagesTypedRequestBudget|DecisionsImagesRawText|UnaryCallbackAllocationBound)$' -count=1
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

Implementation commits: `c6bdfcd`, `92dc2f9`, `259ed61`, `41dbae3`, `e71ec88`,
`e429932`; test-only fixture correction `f97648f`.
Final independent review: Standards 0 remaining, Spec 0 remaining.

Final frozen-source `go test ./... -count=1`: exit 0, **7,838 PASS / 1,050
explicit opt-in NOT_RUN**, including **191/191 ticket cases PASS**. Manifest:
`.evidence/barness-ai/20261016T093917.320707000Z/manifest.json` (`f97648f`).
Final focused related race: exit 0, **207 PASS**, including **191/191 ticket
cases PASS**. Manifest:
`.evidence/barness-ai/20261016T093921.964366000Z/manifest.json` (`f97648f`).
Final `go vet ./...`: exit 0. Both recorder audits contain zero findings.
The verifier checks **66,732 final artifact hashes** (62,163 full + 4,569 race).
Each final bundle records 333 public results, with usage, Attempts/Observer,
fixtures, exact mixed-runtime requests/responses, assertions and zero resources.

The [integrity inventory](09-evidence.json) seals all Go/module sources,
validation documents/logs, both final manifests and written artifacts, and 12
failed historical bundles. Legacy source-fixture hashes may precede credential
redaction; their actual written bytes are verified by the manifest's artifact
hashes. Every ticket fixture additionally matches its stored fixture hash.
Image identities and fixed prices are recomputed independently from the final
wire/result evidence. Mutated/missing artifacts and a nonempty audit were
independently confirmed to make the verifier fail; original bundles were not
modified. The [check log](09-logs/verifier-negative-check.log) preserves this.

Seal the completed runs with:

```sh
PYTHONDONTWRITEBYTECODE=1 python3 .scratch/jev-1.14-multimodal-box/verify-09-evidence.py --record \
  .evidence/barness-ai/20261016T093917.320707000Z/manifest.json \
  .evidence/barness-ai/20261016T093921.964366000Z/manifest.json
```

Bundles remain local in ignored `.evidence`, with per-case replay commands.
Controlled HTTP PASS does not establish real-service protocol, visual quality or
formal support. Catalogs, support matrix, frozen Pi and prior defect statuses
remain unchanged; final PASS does not resolve the preserved timer competition.
