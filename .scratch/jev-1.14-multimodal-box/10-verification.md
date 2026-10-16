# 10 — Host unified image example verification

Scope: standalone Host example only. Built-in catalogs, formal support matrix,
release snapshot, frozen Pi, parent spec and proposed ADR statuses are unchanged.
Controlled HTTP proves protocol/Host behavior; real-service protocol and visual
quality remain NOT_RUN. No real credentials or remote service were used.

## Implementation and replay

```sh
go run ./ai/examples/classifierimages
go test ./ai/e2e -run '^TestClassifierImagesExample' -count=1
BARNESS_EXAMPLE_RACE=1 go test -race ./ai/e2e -run '^TestClassifierImagesExample' -count=1
PYTHONDONTWRITEBYTECODE=1 python3 .scratch/jev-1.14-multimodal-box/verify-10-evidence.py
```

[Failure analysis](10-failure-modes.md) and external E2E were written before the
example implementation. First red `20261016T112243.588000000Z` was honestly the
missing executable, not a library regression. Behavioral red
`20261016T112642.767633000Z` then executed the public calls and detected lost
legend pointer/index fields in the Host view (a promoted embedded JSON encoder).
The Host view now uses independent explicit fields. No unit tests were added.

The 48-scenario executable crosses both public Classify entries and both actual
HTTP adapters, including expected capability/budget/authorization/cancel
rejections. E2E independently checks hashes, final indices, relative legend
pointers, typed atomic errors, Box null/zero confidence, zero independent
existence and the worked rotation/crop inverse. Three negative live-option
configurations must not manufacture a live manifest. Total: 52 ticket cases.

Each persistent nested `classifier-images-host` bundle contains source and
processed PNG/JPEG files, provenance, full public input/config/options/scope,
final native requests/responses, results/errors/views, Attempts, Observer,
resource release assertions, response metadata, per-case replay commands,
complete SHA-256 inventory and redaction audit. The enclosing E2E recorder
adds independent assertions, audit and reproducible case commands.

[Integrity receipt](10-evidence.json) seals the exact manifests and ticket source
hashes; its verifier checks every recorded E2E artifact and every standalone
Host file, final image identities, legend references, coordinate conversion,
common-input equivalence and resource release. The two historical red manifests
are retained, never relabeled PASS. Dynamic RequestIDs/times differ on replay;
synthetic fixture hashes are deterministic.

## Final validation

- `go test ./... -count=1`: exit 0, **7,890 PASS / 1,050 explicit opt-in
  NOT_RUN**, including **52/52 ticket cases PASS**. Final E2E manifest:
  `.evidence/barness-ai/20261016T113347.968132000Z/manifest.json`.
- `BARNESS_EXAMPLE_RACE=1 go test -race ./ai/e2e -run
  '^TestClassifierImagesExample' -count=1`: exit 0, **52/52 PASS**; the actual
  executable is also compiled/run with `-race`, not only its outer test process.
  Manifest: `.evidence/barness-ai/20261016T113347.035722000Z/manifest.json`.
- `go vet ./...`: exit 0 on the final Go sources.
- Both final parent recorder audits and both nested Host audits: zero findings.
  Every recorded resource gauge returns to zero in all 48 executable scenarios.
- Independent integrity verification: **63,239 artifact hashes**, **104 ticket
  cases across the two final runs**, and both historical red manifests/children.

[Independent two-axis review](10-code-review.md): Standards found one P3 mutable
legend alias, fixed in `5d2d6ac` and independently confirmed resolved. Spec found
no actionable findings. The final public E2E/race/full runs cover the corrected
source; no Go implementation changes followed those runs.

Seal these completed runs with:

```sh
PYTHONDONTWRITEBYTECODE=1 python3 .scratch/jev-1.14-multimodal-box/verify-10-evidence.py --record \
  .evidence/barness-ai/20261016T113347.968132000Z/manifest.json \
  .evidence/barness-ai/20261016T113347.035722000Z/manifest.json
```

Bundles stay local under ignored `.evidence`; the committed receipt preserves
manifest hashes, source hashes, audit hashes, counts and historical red pointers.
Single-case E2E replay may regenerate the complete small Host matrix before
selecting the requested assertion case. Standalone per-scenario commands are
also stored in each config. Reruns require new output directories.

The pre-existing uncommitted glossary/README/planning edits were preserved;
commits include only ticket changes. Parent spec/design, proposed ADR statuses,
built-in catalog, release snapshot, support matrix and frozen Pi remain unchanged.
Real-service protocol, independent visual quality and production support are
**NOT_RUN / unclaimed** by this ticket; controlled HTTP PASS never replaces them.
