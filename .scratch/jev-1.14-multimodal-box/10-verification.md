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

## Validation status

Focused public E2E and actual executable race run passed; `go vet ./...` passed.
Full-suite validation and mandatory independent two-axis review are in progress.
Final run paths, counts, audits and review findings will be recorded below.
