# Kimi review follow-up 01

Original findings: [review-01](review-01.md). Fix code commit `6757fb2`; reviewed freeze `0300710ad027a7f86fb8a55d4c81367c9da39620`. Both reviewers inspected source without executing tests. Ordinary execution facts, pending live and later phases remain separate.

## Standards

Reviewed fix range `ed6d4aba84b340efe99a7b909974537e4779c954...0300710ad027a7f86fb8a55d4c81367c9da39620` in `/private/tmp/barness-kimi/review-fixes`. Read-only source review; no tests run or repository changes.

**Original Standards findings: closed in source.**

- **Narrowed output budgets:** `ai/chat_options.go` and `ai/responses_options.go` now emit the Host's smaller `Model.MaxTokens` when full options omit the budget. Their final Kimi guards reject absent budgets whenever the model cap is below131072, so a payload hook cannot delete the enforced cap. Explicit overflow remains rejected. The131072 candidate retains its existing omitted-field wire contract; the new defaults are restricted to Kimi, leaving old Provider branches unchanged. `TestKimiNarrowOutputLimit` covers both markets, both protocols, all four public entries, hook removal and hook overflow.

- **Exact Chat tool indexes:** `ai/chat_kimi_stream.go` now validates `tc.Index` through `parseKimiCount` before accepting the shared float64 index. Missing/null, negative, fractional and overflow values fail; `n.n > 1<<53-1` also rejects values outside the shared representation's safe integer range. Every accepted integer is exactly representable, including9007199254740991. The existing shared parser and other Providers retain their behavior. `TestKimiChatExactToolIndex` includes malformed values, split deltas, ordinary nonzero indexes and the maximum accepted integer through Stream/Complete.

- **Naming judgement:** the file and helpers now use `opaque_history_*` names, with direct call-site replacement and no compatibility shim. Generated `glm_` spellings remain unchanged, with an explanation that they carry no Provider attribution or replay authority. This addresses the name while preserving existing history bytes.

Also inspected the fix's shared Messages sequencing, Responses completed-part consistency and Kimi native request replay changes. Their new behavior is gated to Kimi; no additional concrete Standards breach or regression was identified in this diff. `git diff --check` passed. This review does not claim the pending ordinary run passed or grant live/race/pressure qualification.

**Status:** two hard findings and one naming judgement closed; zero new Standards findings.


## Spec

## Spec follow-up

Reviewed frozen diff `ed6d4aba84b340efe99a7b909974537e4779c954...0300710ad027a7f86fb8a55d4c81367c9da39620` in `/private/tmp/barness-kimi/review-fixes`. Read-only inspection; no tests executed.

**All three original findings are closed by code inspection. No new actionable defect found in these fixes.**

- **Messages `any`: CLOSED.** `ai/anthropic_kimi.go:82–87` now rejects forced choice without declared tools. `kimiMessagesRequiresTool` reads the final serialized payload, including hook changes; `ai/anthropic.go:229–230` passes this flag to the parser. `ai/anthropic_sequence.go:43` rejects ordinary `end_turn` under that requirement, while retaining length/refusal handling. This meets spec K05 and design.md:142.
- **Responses completed text/summary consistency: CLOSED.** `ai/responses_kimi_completed.go:8–56` compares completed parts by index and text, including missing parts. `ai/responses_kimi_decode.go` invokes these checks before accepting item completion, part completion and terminal output. Both visible text and reasoning summaries are covered, satisfying spec story 36 and design.md:164.
- **Actual-wire qualification controls: CLOSED.** `ai/internal/testkit/supportmatrix/regional_kimi_request.go:26–75` independently validates each declared scenario’s effort presence/value, output budget and tool choice. The replay hook compares captured controls/tools with saved options, replaces the synthetic payload with the captured body, and subjects it to public Client authorization/final guards. `regional_replay.go:35–45` additionally verifies the transmitted replay body equals the capture. Changing both wire and saved effort cannot evade the independent scenario check. New resealed-tamper cases cover all six routes through the existing qualification pipeline, meeting K16/design.md:301–316.

The added public E2Es were inspected, not run by this reviewer. This conclusion does not claim ordinary-test completion, real qualification, promotion, race, pressure or release readiness.


Standards: 0 open documented-standard findings and 0 open naming judgements. Spec: 0 open findings from this review.
