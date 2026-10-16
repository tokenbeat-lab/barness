# Ticket 02 — Decisions unified context references

Implemented on the existing `main` branch. Baseline: `7c1cd1a533a924dd06c57beaaa03a0977f130de2`.
Runtime/test source commit: `15b66fd6d9ffbbf1572ddc000890bebe3a73861b`.

`ImageReferences` now has a closed codec, complete usage validation, an owned normalized plan, ordered/repeated Decisions context output, local value-slot structure validation with precredential capability refusal, deeply copied catalog locations and finite reference budgets. Final Hook content uses the same permission, model/MIME, occurrence and body bounds. Existing omitted-location catalog hashes and ordinary State/question conversion remain unchanged.

Jev images, local image locations, Box and final image metadata remain assigned to later tickets. Real-service execution is **NOT_RUN**; all results here are controlled loopback HTTP evidence.

## Test-first record

The [failure inventory](02-failure-modes.md) was written before implementation. The public seams were explicitly approved by the originating ticket/spec. No unit tests were added. Public E2Es were made red before implementing each behavioral slice:

- Reference context/round-trip initially failed because `imageReferences` was an unknown request field.
- Value slots initially either sent a request or returned scope errors for legitimate referenced roots; invalid target/duplicate cases exposed missing checks.
- Budget tests exposed credential reads and late native rejection for expanded references; large reference tables/pointers also exercise pre-copy bounds.
- Catalog tests initially failed to compile because the public location capability was absent.
- A document traversal regression first allocated about 931 KB before enforcing the 4 KB State limit; the check now precedes pointer traversal.
- Review's combined-size regression first returned `resource_limit/request`, read credentials and allocated about 10 MB. It now fails at scope before copying/pinning. Whitespace-padded/escaped State regressions were also first red, then passed after replacing mixed raw/native accounting with conservative projected text sizes.

The [two-axis review](02-code-review.md) records both P2 findings and their verified resolutions. No findings remain on either axis.

## Repeat and verify

```sh
go test ./... -count=1
go test -race ./ai/e2e -run '^TestDecisionsReferences' -count=1
go vet ./...
python3 .scratch/jev-1.14-multimodal-box/verify-02-evidence.py
```

Each case's `assertions.json` also records its exact single-case replay command. Default full-suite opt-in cases remain NOT_RUN; no pressure, Pi-diff or real-service result is implied.

The [evidence inventory](02-evidence.json) records the final full/race bundles, manifest and audit hashes, source hashes and verified artifact counts. Cases include synthetic PNG/JPEG/WebP files and decoded hashes, public requests (or bounded deterministic recipes for oversized inputs), final HTTP captures, response scripts, results/errors, Attempts, Observer records, policy/catalog/Binding configuration, assertions and released-resource gauges. No production images or credentials are used.

Final validation passed on 2026-10-16:

- `go test ./... -count=1`: exit 0; E2E manifest has 6,717 PASS and 1,050 explicitly opt-in NOT_RUN cases. All 112 ticket-02 cases PASS.
- `go test -race ./ai/e2e -run '^TestDecisionsReferences' -count=1`: exit 0; 112/112 PASS.
- `go vet ./...`: exit 0.
- Independent SHA-256 verification: 44,095 full-suite artifacts and 964 race artifacts matched their manifests; both audit files contain zero findings. The verifier also checked changed/supporting source files and the aggregate tracked Go source inventory.

Final manifests: [full](../../.evidence/barness-ai/20261016T042652.492335000Z/manifest.json), [race](../../.evidence/barness-ai/20261016T042656.176525000Z/manifest.json). Terminal output is retained in [full.txt](02-logs/full.txt) and [race.txt](02-logs/race.txt).

Evidence remains under the repository's ignored `.evidence/barness-ai/` directory; the verifier checks every case artifact against the bundle manifest. The implementation commits cover only this ticket; preexisting unrelated documentation and planning changes remain in the worktree.
