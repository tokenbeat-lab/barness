# Ticket 03 — Decisions final image identity

Implemented on `main`, starting from `1319e1f48d3b01d4f9e63890376dc381a8a5e2fa`.
Runtime/test source commit: `b6175d705591f4477263e42f313b5811c62c4704`.

`ClassifierResult.ImageInputs` contains only canonical MIME and lowercase SHA-256
of each final decoded file. The final validated native Decisions content owns
the occurrence order, including duplicates and Hook changes. Hashing reads one
file at a time with a fixed 4 KiB buffer, bounded by the existing final image and
request budgets, and checks cancellation at each chunk. Answers and image
metadata publish together after complete response validation; errors clear both.
Original usage/ResponseModel rules and metadata-only Observer remain intact.

This implements ticket 03 only. Jev images, local positions, image legends and
Box remain assigned to later tickets. Catalog/support claims are unchanged.
Real-service execution is **NOT_RUN**; all new protocol evidence uses controlled
loopback HTTP and synthetic image fixtures.

## Test-first record

The [failure inventory](03-failure-modes.md) was written before implementation.
The ticket explicitly agreed the public `Client.Classify` and
`HookedClient.Classify` seams. No unit tests were added.

- [Identity red](03-logs/order-red.txt): six image scenarios failed because the
  returned result omitted final identities; the text-only control passed.
- [Identity green](03-logs/order-green.txt): all seven scenarios then passed.
- [Codec red](03-logs/codec-red.txt): invalid metadata could encode/decode and
  failed decoding could mutate a previous record. The closed codec fixed this.
- Final E2Es also exercise Hook replacement/addition/removal/reorder, escaped,
  typed and custom-encoded content, whole-file hashes, invalid final inputs with
  zero Attempt, bad answers/usage, cancellation, refusal, missing/null usage,
  returned-record ownership and private Observer/diagnostics. Existing snapshot,
  retry and simultaneous two-tenant mixed-operation E2Es verify the new metadata.
- An intentionally invalid JSON codec fixture initially used a `.json` filename;
  the audit correctly rejected that evidence run. It is now a `.txt` input
  fixture. That intermediate run is not a final PASS bundle.

## Reproduce and verify

```sh
go test ./... -count=1
go test -race ./ai/e2e -run '^TestDecisions(ImageInputs|ImagesSnapshot|ReferencesFrozenCall|RuntimeMixedIsolation)' -count=1
go vet ./...
python3 .scratch/jev-1.14-multimodal-box/verify-03-evidence.py
```

Each scenario's `assertions.json` and manifest retain its exact single-case
replay command. The [inventory](03-evidence.json) fixes source hashes, manifest
and audit hashes, case counts and all artifact counts. The verifier independently
recomputes decoded file identities from final captured requests and compares
them with successful results, checks failure atomicity and shared diagnostics,
then verifies every artifact and both audits.

Bundles under ignored `.evidence/barness-ai/` contain PNG/JPEG/WebP fixtures and
file hashes, public input, Hook input, final request and response scripts,
results/errors, Attempt/Observer, assertions, finite policy/catalog/Binding and
released-resource gauges. [Two-axis review](03-code-review.md) records the
independent Standards and Spec reviews and resolution of any findings.

Final validation passed on 2026-10-16:

- `go test ./... -count=1`: exit 0; 6,801 PASS and 1,050 explicit opt-in
  NOT_RUN cases. All 84 ticket-03 scenarios PASS.
- Focused race: exit 0; 89/89 PASS (84 new plus five snapshot/retry/isolation
  regression cases).
- `go vet ./...`: exit 0.
- Both final redaction audits contain zero findings. The independent verifier
  checks all 45,670 full-suite and 1,762 race artifacts, the complete tracked Go
  source inventory, 84 final-identity/atomicity cases in each bundle, and linked
  source/config/fixture/verification files.

Final manifests: [full](../../.evidence/barness-ai/20261016T044522.215757000Z/manifest.json),
[race](../../.evidence/barness-ai/20261016T044456.014744000Z/manifest.json).
Logs: [full](03-logs/full.txt), [race](03-logs/race.txt), [vet](03-logs/vet.txt).
All implementation and verification files are committed; preexisting unrelated
user documentation and planning changes are left in the working tree.
