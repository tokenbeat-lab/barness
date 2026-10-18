# MiniMax international Chat protocol candidate

All inputs and SSE frames in `minimax_chat_*_test.go` are synthetic, authored
against spec R07–R11/R17–R21 and the 2026-10-18 official source snapshot. No
vendor output, real credential or actual-price claim is embedded in this suite.
Each public Client run records the final request, Provider SSE reply, events,
Result/Attempt/Usage and (text lifecycle cases) Observer records. The runner
performs its mandatory redaction audit before a run succeeds.

Official contracts:

- [OpenAI-compatible text API](https://platform.minimax.io/docs/api-reference/text-openai-api):
  snapshot SHA-256 `ee59b46734b303cce80dc60fd7729abce43cb205ad5914356c16beac95934412`.
- [Chat OpenAPI](https://platform.minimax.io/docs/api-reference/text/api/openapi-chat-openai.json):
  snapshot SHA-256 `8d26ce6ffaa62d95be4ccaa5926e08bbdb6cb0f7d172067373439207f93a65ce`.
- [M3 tool guide](https://platform.minimax.io/docs/guides/text-m3-function-call):
  Markdown SHA-256 `40ab52a616a5321ff841dc6250b0b62a7cdc0384eb0218365f27beef648f25bb`.

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
07 must capture and replay international standard pay-as-you-go M3, confirm native auto,
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

International applicability was rechecked for ticket11 using independently
retrieved2026-10-18 `.io` snapshots. The three OpenAPI structural schemas match
the historical CN shapes after prose metadata is excluded; this is not live
wire/account/pricing proof. Historical CN captures and source hashes remain
immutable in their original evidence archives; no CN native origin qualifies
an international continuation. See ADR-0042 and research-international.md.

Current international observations: [ticket11](../../../../.scratch/barness-ai-minimax/evidence/ticket-11/README.md).
