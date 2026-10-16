# Independent multimodal candidate probes

Run from the repository root. No built-in catalog, formal support matrix or Pi
snapshot is changed. The two JSON catalogs are explicit Host candidates, pinned
by file hash in the harness and verifier. Jev uses `jev-1.14.0` and only PNG/JPEG;
Decisions uses its own `gpt-6-luna` identity and context-only images.

Candidate facts come from the existing Jev 1.14 research/native snapshots under
`.scratch/jev-1.14-multimodal-box/` and the pinned Decisions snapshots in
`../decisions/`. Jev's 0.042 input-only rate is a snapshot, not observed billing.
These candidates are deliberately independent of formal inclusion (tickets 13/14).

```sh
# Offline process E2E; each child gets one synthetic key and no inherited keys.
go test -tags live ./ai/live -run '^TestMultimodalCandidateProcess$' -count=1
go test -race -tags live ./ai/live -run '^TestMultimodalCandidateProcess$' -count=1

# Build once; execute each route in a separate, sanitized Host process.
go test -tags live -c -o /tmp/barness-multimodal-probe ./ai/live
```

For real service verification, the trusted Host supplies only these variables
to a fresh process (never source another task's `.env` or copy its credentials):

- `BARNESS_MM_ROUTE=jev` or `decisions`; `BARNESS_MM_LIVE=1`.
- `BARNESS_MM_KEY`: the explicitly approved credential for that route only.
- `BARNESS_MM_AUTHORIZATION`: path to a bounded JSON Host authorization record.
- Optional `BARNESS_AI_LIVE_PROXY`: the existing smoke's loopback proxy format;
  no endpoint override, TLS bypass or automatic environment proxy is supported.
- Optional absolute `BARNESS_AI_EVIDENCE_DIR`: parent of new timestamped packages.

Execute the prepared binary with working directory `ai/live` and flags
`-test.run='^TestMultimodalCandidate$' -test.timeout=180s` (or use `go test` from
the repository root). Missing switch/key/authorization record is NOT_RUN with zero
spend. A supplied invalid authorization is FAIL before spend. Vendor failures
stop the remaining pipeline, preserve the raw failure, and never retry.

Example authorization (aliases and reference identify the Host's approval,
not a provider permission-introspection result):

```json
{
  "route": "jev",
  "model": "jev-1.14.0",
  "service": "https://api.typesafe.ai/v1/systemone",
  "accountAlias": "dedicated-test-account",
  "regionAlias": "host-verified-region",
  "authorizationReference": "approved-ticket-11-run-1",
  "dedicatedCredential": true,
  "maxCalls": 12,
  "maxAttempts": 12
}
```

For Decisions change route/model/service to `decisions`, `gpt-6-luna`,
`https://api.openai.com/v1/decisions`, and use that route's own credential.
Account/region/privileges remain Host attestations; billing, image tokenization,
service capacity, visual quality, IoU and absence false positives remain unknown.

Both public entry paths run default context, explicit context and `[1,0,1]`
ordered reuse. Jev also runs State + own-question instructions, Choice/Bool/Score
visual criteria + public Score legend, and mixed Box + independent Bool existence.
The Hook exchanges the fixture files, so final identity is checked against the
actual request. Real Box confidence can be numeric or null; visual correctness
is outside this protocol probe (ticket 12). No invalid large service requests
are made to force error codes.

Per route/process: at most 12 Logical Calls / 12 HTTP Attempts, zero retries,
32 questions, 24 sources / 32 occurrences, 8 MiB decoded source / 16 MiB expanded /
4 MiB State cumulative; 4 MiB native body per call, 30 s per call, 180 s process
including 10 s finalization reserve. Calls reserve before invoking Classify;
the transport checks final Hook resources before network and counts failed sends.
The larger fixture file is reserved for each source/occurrence; State reserves
bounded projection overhead. Actual counters independently count final native
State and images. Per-call image policy is conservatively 4 MiB under body policy.

The bundle includes source hashes + Git commit, catalog/file hashes, image
fixtures, public inputs, final redacted requests/responses, results/errors,
Attempt/Observer, assertions, reservations/actuals, configuration, resource
release, `manifest.json`, `audit.json`, `multimodal-report.json` and a final
`multimodal-integrity.json`. Never edit old bundles to accept a new source hash.

```sh
python3 .scratch/jev-1.14-multimodal-box/verify-11-evidence.py BUNDLE
# Required for real service acceptance; controlled HTTP is always rejected.
python3 .scratch/jev-1.14-multimodal-box/verify-11-evidence.py --require-live BUNDLE
```

Python 3 is required for independent verification and the process E2E. The
verifier checks every final checksum, fixture, audit, manifest, assertion,
catalog/route/model, exact scenario set, input mapping, final image identities,
visual legend references, Box, usage, Observer and cumulative counters. NOT_RUN,
FAIL and controlled packages cannot substitute for a route's real protocol PASS.
Candidates cannot be imported by the formal support-matrix consumer.
