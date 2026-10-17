# Ticket 04 Responses content and request boundary evidence

[Failure plan](plan.md), [hashed run index](index.json), [ADR-0034](../../../../docs/adr/0034-barness-ai-zhipu-responses-boundary.md). Durable bundles are at root `.evidence/barness-ai-glm/ticket-04/`; each completed bundle contains replay commands, request/response scripts, public events/Result/Attempt/usage, fixture hashes, assertions and a passing redaction audit. All new source files were staged before the final regression run.

The first public Client run reproduced the silent SDK loss in both models: no error and completed, but object reasoning/message content yielded empty strings. Structural normalization fixes the parsing copy; original native reasoning remains separately preserved. The object fixtures derive from the official Responses schema and are synthetic, never claimed as live object observations. Both actual basic streams in `protocol-exploration/responses/01-basic` were arrays; tracked `testdata/glm-responses/text.json` also replays their raw frames and original item.done reasoning independently by model.

Final staged regression: 1,978 PASS plus one expected NOT_RUN (the separately opt-in frozen-oracle guard), no FAIL. Post-merge GLM plus thinking/sampling baseline with the frozen-oracle opt-in: all 212 PASS, no NOT_RUN. Exact final budget/preflight rerun after switching authorization decode to UseNumber: 115 PASS. Affected OpenAI Responses frozen pi differential: 335 PASS, zero pending; GLM remains an extension and is not counted as native pi parity. `go vet ./ai/...` passes. Source budget precision red and green evidence uses the same public Client boundary; Chat/Responses share the exact integer helper while retaining separate request field encoding.

Coverage: both models; four public entries, nil/full and WithHooks; schema objects and arrays; done-only, terminal-only, delta/done correction and out-of-order output indexes; malformed structures, text types/indexes and terminal statuses; incomplete length/error, failed/error and missing terminal EOF; independent Binding/path/Bearer/attribution; store=false and omission of OpenAI automatic fields; 5.3 full effort rejection/simple clamping; exact output budget1..min(model cap,131072), including all sampling layers and final hook, fractional/huge-exponent rejection, no OpenAI16 uplift; authorization/credential/zero-attempt ordering.

Replay from a clean checkout (no credentials required):

```sh
BARNESS_AI_EVIDENCE_DIR=.evidence/barness-ai-glm/ticket-04/replay go test ./ai/e2e -run '^TestGLMResponses' -count=1
go vet ./ai/...
BARNESS_AI_PIDIFF=1 BARNESS_AI_EVIDENCE_DIR=.evidence/barness-ai-glm/ticket-04/replay-diff go test ./ai/e2e -run '^TestPiDifferential/openai-responses$' -count=1
```

The frozen differential requires the already locked oracle `node_modules` (`npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node`). Its temporary worktree symlink is never committed. No GLM credential was read and no paid request was sent by this ticket. Tool/history input conversion and enhanced guards remain ticket05; formal rates/catalog/live support remain their own tickets.

Two early diagnostics under `red-options` / `green-options` were interrupted when a test helper called Fatal inside a callback goroutine while decoding a deliberately huge numeric literal. The test was corrected to pass raw JSON through the public callback, avoiding helper goroutine termination. Those incomplete runs are diagnostic only, not acceptance evidence; all indexed completed runs were audited.
