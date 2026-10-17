# Ticket 01 evidence

Mode: offline synthetic configuration/shared-sampling acceptance. No real
credentials or vendor calls. MiniMax protocol/live acceptance remains with
tickets 02–07. Candidate rates are mechanism-test values, not real estimates.

`failure-modes.md` predates implementation. Archives retain complete original
bundle manifests, assertion reports, fixtures, final requests, SSE reply
scripts, events, Result/Attempt/Usage and redaction audit inputs. `SHA256SUMS`
binds all preserved archives and logs. The shared sampling test intentionally
uses established OpenAI routes: it proves default-label consumption, not any
unimplemented MiniMax wire behavior. Protocol-specific observation/isolation
is covered by follow-on tickets.

| Run | Actual result |
| --- | --- |
| red-metadata | FAIL: label lost and invalid configuration accepted |
| green-metadata | PASS after metadata/validation implementation |
| red-sampling | FAIL: high/max profiles select off; Responses fixture lookup also exposed a test assembly mistake |
| green-sampling | FAIL: missing Responses text fixture, preserved rather than overwritten |
| green-final | FAIL: shared test wrongly assumed native OpenAI Responses omitted its existing none field; corrected to require unchanged native wire |
| green-candidate | PASS: six explicit candidates and public configuration/Hash isolation |
| green-accepted | PASS: all 30 own metadata/shared-sampling/candidate cases |
| regression | FAIL: same earlier fixture assembly error; no production regression claimed |
| regression-final | PASS: existing sampling options/snapshots/configuration and GLM mandatory metadata |
| race | PASS: all own MiniMax foundation cases with race detector |
| vet | PASS: `go vet ./ai ./ai/internal/testkit/minimax` |

The fixture and native-wire corrections changed only test assembly/expectations;
the production sampler remained the one-line omitted-label selection. Initial
failures and the final corrected cases remain attributable in separate bundles.
Generic Client worlds use the repository's existing controlled loopback HTTP
assembly; new MiniMax protocol worlds use their own verified transport seam.

Replay from the repository root without a key or network vendor access:

```sh
go test ./ai/e2e -run '^TestMiniMax' -count=1
go test -race ./ai/e2e -run '^TestMiniMax' -count=1
go test ./ai/e2e -run '^Test(ThinkingSamplingOptions|ThinkingSamplingSnapshot|CatalogThinkingSampling|GLMChatMandatoryThinkingMetadata)' -count=1
go vet ./ai ./ai/internal/testkit/minimax
```

Extract any archive with `tar -xzf <archive>` to inspect its full immutable
historical bundle. The run's test replay command is in manifest.json.
