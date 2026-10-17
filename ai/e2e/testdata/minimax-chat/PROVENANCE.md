# MiniMax China Chat protocol candidate

All inputs and SSE frames in `minimax_chat_*_test.go` are synthetic, authored
against spec R07–R11/R17–R21 and the 2026-10-18 official source snapshot. No
vendor output, real credential or actual-price claim is embedded in this suite.
Each public Client run records the final request, Provider SSE reply, events,
Result/Attempt/Usage and (text lifecycle cases) Observer records. The runner
performs its mandatory redaction audit before a run succeeds.

Official contracts:

- [OpenAI-compatible text API](https://platform.minimax.cn/docs/api-reference/text-openai-api):
  snapshot SHA-256 `2b1b748877fe8f1c5712f9073dfaf3f4b83dfde89bf97cd7219300216fb3f552`.
- [Chat OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-openai.json):
  snapshot SHA-256 `9b01666f1df386776cb4d9f49a8dd2832f94ab7bbb446d1c72fa3e76af4348fb`.
- [M3 tool guide](https://platform.minimax.cn/docs/guides/text-m3-function-call):
  Markdown SHA-256 `42193c24a679a9bc170cf1ee30b15a1ff5adae9a6f5c1a7cd222470f1fecc8a7`.

The adapter preserves omitted thinking and effort, asks for split reasoning and
usage, uses system instructions/max_completion_tokens, and suppresses native
OpenAI/DeepSeek/GLM defaults. M3's full high means adaptive on without a depth
field. M3.1 uses declared depths and refuses disabling. Public auto omits the
native choice; an injected native auto is provisional until its own real
candidate proves acceptance. Other choices/parallel controls remain rejected.

The official M3 split-field conflict is **unresolved by synthetic evidence**.
reasoning_content and existing typed reasoning_details are exercised separately;
distinct native ids/indexes survive replay rather than flattening entries. An
unintegrated or mixed native shape fails explicitly so it cannot be lost while
reporting success. Neither synthetic format proves real M3 continuation. Ticket
07 must capture and replay China standard pay-as-you-go M3, confirm native auto,
and either validate these shapes or make the smallest evidence-backed boundary
conversion. Raw content containing think tags is never stripped or relabelled
as independent Thinking. reasoning_split=false remains outside this subset.

Replay without a key:

```sh
go test ./ai/e2e -run '^TestMiniMaxChat' -count=1
go test -race ./ai/e2e -run '^TestMiniMaxChat' -count=1
go test ./ai/e2e -run '^Test(MiniMaxChat|Chat|DeepSeekChat|GLMChat)' -count=1
go vet ./ai
```

Test rates come exclusively from the labelled candidate catalog. Shared HTTP
business errors, trace id evidence, hooks panic recovery, concurrent isolation
and policy pressure are ticket 05; real protocol acceptance is ticket 07.
