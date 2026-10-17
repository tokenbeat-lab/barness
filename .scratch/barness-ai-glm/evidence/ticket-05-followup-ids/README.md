# Ticket 05 scoped follow-up: opaque IDs and downgrade pairing

[Failure plan](failure-plan.md), [hashed bundles](index.json), [ADR-0034](../../../../docs/adr/0034-barness-ai-zhipu-responses-boundary.md). No credentials or paid calls were used. Implementation commits: `709e6d0`, `c29cb4a`, `0f33a39`; final committed replay source: `0f33a39` (includes integration `8839138`). Raw bundles are durable at root `.evidence/barness-ai-glm/ticket-05-followup-ids/`.

The official Responses schema snapshot SHA256 `0914af477ff4c1fae0372f9c44ace23033d0bed0cc486796eaff2c2024e5f18c` defines call_id and output item id as plain strings, without excluding percent, Unicode or `|`. These are independent synthetic boundary cases from that contract, not newly observed service IDs or a live claim.

Three public Client reds preceded their corresponding changes:

- `red-public`: 48 FAIL / 10 PASS across 58 cases. Two native call IDs `same|a` / `same|b` both replayed as `same`; distinct native tuples could also yield the same public ToolCall.ID. Both models, percent/delimiter/Unicode, trusted/Host-restored/untrusted/cross-model/cross-account, and independent Chat-origin→GLM Responses histories are included. `green-public` passes all 58.
- `red-downgrade-unique`: 22 FAIL / 6 PASS across 28 cases. Distinct valid GLM Responses call/item IDs collapsed after punctuation replacement or 64-character truncation on GLM Chat/Messages and native OpenAI/DeepSeek Responses/Chat/Messages targets. `green-downgrade` passes these plus the first 58 cases (86 PASS).
- `red-empty-origin`: 4 FAIL. One public GLM Chat Unicode call ID became an empty call_id when targeting legacy OpenAI/DeepSeek Responses. `green-all-ids` passes all 90 new edge cases.

GLM Responses now escapes percent and delimiter characters independently in both opaque public ID components; its request boundary decodes only the call component. Ordinary captured IDs preserve their public spelling and exact own-route wire IDs. Native-state downgrades continue independently, output item IDs never enter GLM input, and Host results retain their specific call pairings.

Other target protocols keep their existing normalization. A complete-history prepass repairs only GLM-origin wire IDs that collide or become empty, reserving existing spellings, choosing deterministic short replacements, and using a bounded free-ID allocation if a hash spelling is already occupied. The existing result renaming pairs the same replacements. Every history without GLM and non-colliding ordinary GLM conversions retain the original rules; no public DTO or general ID registry was added.

Committed-source verification: affected public history/tool/native/cross-route regression **288 PASS**, selected race **124 PASS**, and `go vet ./ai/...` passes. The frozen native Responses differential passes **335** cases with zero pending; Zhipu remains an extension route.

Public replay selectors:

```sh
BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-glm/ticket-05-followup-ids/replay" go test ./ai/e2e -run '^Test(GLMResponses(OpaqueNativeIDs|CrossOriginOpaqueIDs|OriginIDCollision)|GLMChatOriginNonemptyResponseID)$' -count=1
BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-glm/ticket-05-followup-ids/replay-race" go test -race ./ai/e2e -run '^Test(GLMResponses(OpaqueNativeIDs|CrossOriginOpaqueIDs|OriginIDCollision)|GLMChatOriginNonemptyResponseID)$' -count=1
BARNESS_AI_PIDIFF=1 BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-glm/ticket-05-followup-ids/replay-diff" go test ./ai/e2e -run '^TestPiDifferential/openai-responses$' -count=1
```

Case prefixes are `GLM05-opaque-ids-*`, `GLM05-cross-origin-ids-*`, `GLM05-origin-id-collision-*`, and `GLM05-empty-origin-id-*`. Frozen differential uses the unchanged pinned pi 1.1.0 oracle; the temporary node_modules worktree link is removed before handoff. Every indexed completed bundle has a passing redaction audit.

One early compilation attempt had no running test/bundle. Initial `red-downgrade` used duplicate item IDs and is retained only as diagnostic; it is superseded by `red-downgrade-unique`, whose native call and item IDs are independently unique. None of these edge-case fixtures claims real vendor observation.
