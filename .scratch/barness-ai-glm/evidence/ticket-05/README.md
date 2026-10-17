# Ticket 05: stateless Responses history, tools and cache routing

[Failure plan](failure-modes.md), [hashed runs](index.json), [ADR-0034](../../../../docs/adr/0034-barness-ai-zhipu-responses-boundary.md).

Implementation `688847c`, committed replay source after merging integration through ticket 08: `d0d935f`. All new source was staged before final inventory/regression. Durable raw bundles are at root `.evidence/barness-ai-glm/ticket-05/`; each completed bundle contains scripted SSE, captured final requests, public events/Result/Attempt/usage, fixtures, assertions and redaction audit. No credential was read and no paid request was sent.

Final committed post-merge verification: **3,716 PASS** across GLM and affected public regressions, **471 PASS** under Responses race, **335 PASS** frozen OpenAI Responses differential with zero pending, and `go vet ./ai/...` passes. The GLM route remains an extension.

Both models complete Host-driven function loops, including independent call_id/item_id, argument deltas and authoritative done correction. Native reasoning keeps its original output object/array signature; the request boundary derives the documented InputReasoning content object without id/status/summary/encrypted fields. All GLM InputMessage items have type/role/content, without output IDs or synthetic OpenAI IDs. Trusted in-process and Host-restored state, untrusted/cross-model/provider/API/account downgrades and unchanged source messages are exercised. Visible-only thinking survives as ordinary assistant text without fabricated native reasoning. Malformed nonempty trusted state fails explicitly.

Only default/none/auto tool choice is integrated. Named unsupported controls refuse before credentials; final sampling/hook controls refuse before any Attempt, with model/cache/account authorization taking precedence. Host Catalog cannot enable strict; function declarations, namespace/hosted/custom tools, summary, encrypted include, tier, TTL/retention, strict JSON Schema, saved-state references and final text-only modalities are checked. Public simple choices remain unchanged. Legal text/json_object callback edits preserve header→payload→response order and frozen bytes across explicit retry, even after caller-owned maps/raw JSON are changed.

Cache routing uses the existing tenant/account/session-derived key for default/short/long when session exists; missing session or none omits it. Both models, identical tenant-local Binding/credential/account/session names, distinct tenants/accounts, key authorization and absence of affinity/retention/TTL fields are covered. A routing hint does not promise cache lifetime or a hit.

`testdata/glm-responses/tools.json` contains independent official-schema object/array synthetic scenes and each model's own raw observed tool-loop frames. The observed 5.3 tool turn has no reasoning. Its own basic reasoning array is replayed offline separately; neither that synthetic continuation nor the successful 5.2 experimental object conversion establishes a production 5.3 native-history live PASS. Ticket 11 still owns both models' final production live verification. This ticket does not change the built-in catalog or Support Matrix and does not claim a native frozen-pi Zhipu route.

Replay from repository root without credentials:

```sh
BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-glm/ticket-05/replay" go test ./ai/e2e -run '^TestGLMResponses' -count=1
BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-glm/ticket-05/replay-race" go test -race ./ai/e2e -run '^TestGLMResponses' -count=1
go vet ./ai/...
BARNESS_AI_PIDIFF=1 BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-glm/ticket-05/replay-diff" go test ./ai/e2e -run '^TestPiDifferential/openai-responses$' -count=1
```

The differential requires the already locked oracle (`npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node`). Its temporary worktree node_modules symlink is not committed. All indexed completed runs pass their redaction audit. `red-history` initially failed because fixture generation had not completed; `cache` exposed a test setup assumption about reserved catalog cache keys; early tools-lifecycle/regression diagnostics assumed a numeric string was not coercible and treated the shared live PartialView as a fixed pending snapshot. Those are retained as diagnostics, not implementation red or acceptance PASS. Corrected tests respect existing public coercion/live-view contracts. Code reds are separately named and retain their failing public assertions.
