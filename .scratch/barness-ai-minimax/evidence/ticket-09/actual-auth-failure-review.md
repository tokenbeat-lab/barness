# Actual MiniMax authentication-failure audit

Verdict: PASS for failure accounting and retention; actual model support remains FAIL/incomplete.

Source: `9481a6ad8028025a7f35da19a3e488935c6cd897`. Session: `20261017T235416.418778000Z`.

Scope: safe report/Attempt attribution, launcher metadata, source approval hashes, ledger chain, seal hashes and opaque byte-for-byte retained-file hashes. Authentication headers, raw provider error bodies and credential values were not inspected.

Both actual China Chat research calls returned HTTP 401 (`upstream_auth`, request phase), each with one actual Attempt, no retry and unreported usage. The dependent native continuation is explicitly NOT_RUN and has no Call/reservation. Token reservations are conservative limits, not observed billed consumption.

Consumption: 2 calls, 2 Attempts, requested output 5120, reserved input 4096. Candidate phase remains 2/120 calls and 2/144 Attempts; formal phase remains 0/60 and 0/72. No successful actual origin or formal support projection was created.

The retained canonical research archive contains the same 34 files and hashes as the original. Report SHA-256: `9eda3db2343c9ef75b06f6410ea1081323108e2627d46bc8c0406eecb222b286`; seal SHA-256: `4dddbc1f1819fb6dadab5a1b2aaa882fa254a78f2390662bad868139ef060abb`.

## Checks

- PASS: exact reviewed source commit
- PASS: approved exact source/binary/plan receipts
- PASS: all 256 captured execution-source hashes match the frozen committed source; all 254 overlapping launcher-manifest entries agree
- PASS: only selected credential injected
- PASS: actual research remains failed incomplete
- PASS: two logical calls two actual attempts
- PASS: dependent continuation explicitly not run
- PASS: 401 one request attempt research-full-stream
- PASS: correct attribution research-full-stream
- PASS: 401 one request attempt research-tool-call
- PASS: correct attribution research-tool-call
- PASS: four exact linked reserve/attempt receipts
- PASS: original sealed complete file hashes
- PASS: canonical retained file set exact
- PASS: external research pins match original
- PASS: redaction audits clean
- PASS: no success origin or support projection
- PASS: no candidate or formal launch in this collection
- PASS: shared phase caps preserved
- PASS: identity output/input reservations bounded

This narrow audit does not replace final whole-spec review or establish candidate/formal support.

Applicability detail: the capture source list additionally includes `ai/internal/testkit/minimax/README.md` and `pricing-source.md`, outside the launcher's Go/JSON selection. Both exact hashes were verified against the frozen commit, along with every other captured execution-source entry. No source content was displayed or changed.
