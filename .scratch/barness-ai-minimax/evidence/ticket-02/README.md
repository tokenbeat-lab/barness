# Ticket 02 Chat evidence

Mode: offline synthetic public Client acceptance. No vendor calls or real keys.
Candidate rates are labelled synthetic. Real split schema and native auto remain
candidate questions for ticket 07; this evidence makes no formal support claim.

`ai/e2e/testdata/minimax-chat/FAILURES.md` was written before implementation.
`PROVENANCE.md` records official snapshot hashes, implemented subset and real
acceptance boundaries. Archives retain full immutable bundle manifests, request
captures, synthetic SSE, event sequences, Result/Attempt/Usage, Observer lifecycle
records where configured and the runner's redaction audit. Extract with
`tar -xzf <archive>`; per-case exact replay commands are in each manifest.
`SHA256SUMS` binds the curated archives and original logs.

| Run | Result and reason |
| --- | --- |
| red-text | FAIL: inherited store/developer defaults and absent reasoning_split |
| red-options | FAIL: unsupported full levels/choices accepted, auto sent explicitly |
| red-guard | FAIL: invalid final thinking, budget, sampling and capability fields sent |
| red-stream | FAIL: tool calls conflicted with stop or tool_calls had no calls |
| red-history | FAIL: two distinct typed reasoning ids flattened into one entry |
| red-containers | FAIL: malformed native containers and rounded range excess accepted |
| red-usage | FAIL: explicit final disabled thinking incorrectly required reasoning usage |
| red-reasoning | FAIL: unknown native formats silently ignored or accepted |
| red-sampling | FAIL: first exact sampler incorrectly rejected valid 1e-100 |
| red-tool-identity | FAIL: different tool indexes/ids merged into one call |
| green-before-observer | PASS: protocol slices before later tool identity/observer additions |
| observer-assembly-failure | FAIL: new Observer test assembly omitted required queue policy |
| observer-attribution-failure | FAIL: test expected CallStarted to be resolved; existing contract starts unresolved |
| regression | PASS: final 579 own cases + 883 existing Chat/DeepSeek/GLM/parity cases |
| race-final | PASS: final 579 own cases with race detector |
| vet | PASS: go vet ./ai |

The Observer test also corrected its slash-containing RequestID before using the
existing evidence filename helper. These test assembly corrections did not alter
production lifecycle behavior. Earlier failed evidence remains visible rather
than being presented as a pass.

Final behavior covers nil/full/simple entries, omitted default sampling label
without wire effort, M3 adaptive switch without depth, mandatory M3.1 depths,
public auto omission, all final sampling/hook paths and authorization precedence,
exact one-token/boundary/fractional budgets, numeric sampling shape/ranges,
text-only function/message containers without recursively rejecting business
schema properties, complete assistant/tool/native history, interleaved calls,
trusted restore and downgrade, legal finish/EOF with/without DONE, usage at finish
or trailing frames, and error/filter/unknown/missing terminal handling.

Shared HTTP business errors/actual trace-id evidence, hook panic recovery,
concurrent isolation and policy-pressure integration remain ticket 05. Actual M3
reasoning wire and continuation acceptance remain ticket 07. Existing synthetic
typed reasoning_details support is deliberately strict; a different real shape
must receive an evidence-backed conversion before formal acceptance.

Replay from the repository root:

```sh
go test ./ai/e2e -run '^TestMiniMaxChat' -count=1
go test -race ./ai/e2e -run '^TestMiniMaxChat' -count=1
go test ./ai/e2e -run '^Test(MiniMaxChat|Chat|DeepSeekChat|GLMChat)' -count=1
go vet ./ai
```
