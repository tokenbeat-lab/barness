# Kimi integration review 01

Frozen base `68b7da0a3400f909dc535d7f585c93771df1569c`; reviewed head `ed6d4aba84b340efe99a7b909974537e4779c954`. Two parallel, read-only reviews; no tests run by reviewers. Findings below are the original review outcome, not the current resolution state; follow-up fixes and acceptance are recorded separately.

## Standards

Reviewed `68b7da0a3400f909dc535d7f585c93771df1569c...ed6d4aba84b340efe99a7b909974537e4779c954` (executable source identical to `9c3409d15dea4e6df52cd3a1a0d1db48ece45ef1`). Read-only review; no tests run.

Documented-standard breaches:

- **[P2] Narrowed output limits disappear when the budget is omitted.** [chat_kimi.go:102](/Users/cyber/RestoX/barness/ai/chat_kimi.go:102) and [responses_kimi_request.go:73](/Users/cyber/RestoX/barness/ai/responses_kimi_request.go:73) use `if ..., present := ...; present`, while [kimi_catalog.go:14](/Users/cyber/RestoX/barness/ai/kimi_catalog.go:14) accepts smaller `Model.MaxTokens`. A Host catalog capped at4096 therefore permits Full Stream/Complete to omit the output field, requesting the documented server default131072; a payload hook can likewise remove an explicit budget. Messages correctly derives its default from the model. Enforce the narrowed cap when the field is absent, including after hooks. **Rule:** AGENTS.md principle3, “数据必须在边界处完成校验”; principle9 requires behavior to match its contract. ADR-0052 fixes131072 as the service default while permitting the existing model-boundary contract; ADR-0018 requires honoring model hard constraints.

- **[P2] Validate raw tool indexes before float conversion.** [chat_kimi_stream.go:28](/Users/cyber/RestoX/barness/ai/chat_kimi_stream.go:28) checks `!indexed || index < 0 || index != float64(int64(index))`, after [chat_stream.go:361](/Users/cyber/RestoX/barness/ai/chat_stream.go:361) unmarshals into float64. JSON `index:null` successfully decodes to0; `1.00000000000000001` rounds to1. Both satisfy the new guard and can produce a successful executable tool call from an invalid index. Validate presence/type/exact integer syntax from `tc.Index` first. **Rule:** AGENTS.md principle3, “数据必须在边界处完成校验和独立转换”; principle4, “任何外部输入均视为不可信”.

Baseline judgement:

- **Possible Mysterious Name:** [glm_history_ids.go:17](/Users/cyber/RestoX/barness/ai/glm_history_ids.go:17), `func glmHistoryToolIDRules(...)`, now allocates IDs for Kimi across three protocols as well as GLM; generated replacements still use `"glm_"`. Rename the shared boundary for its opaque-history responsibility when addressing related changes. This is a heuristic, not a hard violation or a request for a new abstraction.

Two documented-standard findings; one naming judgement. Pending real qualification and the inactive formal catalog are intentional and are not findings.


## Spec


Reviewed `git diff 68b7da0a3400f909dc535d7f585c93771df1569c...ed6d4aba84b340efe99a7b909974537e4779c954` read-only; no tests executed.

- **[P1] Enforce Messages `any` through the terminal.** Spec K05 (spec.md:184): “强制模式不被静默改为 auto”; design.md:142 explicitly requires `any` plus ordinary `end_turn` without a function call to be protocol failure. `ai/anthropic_kimi.go:82–87` accepts `any` even with no tools, while `ai/anthropic_sequence.go:38–45` accepts `end_turn` with zero calls and never receives the final request’s choice. Consequently a declared-tool request with `any` can succeed as plain text; a hook can also remove all tools without rejection. Validate nonempty tools and carry final choice into terminal validation.

- **[P2] Reject Responses text/summary completion contradictions.** Spec story 36 (spec.md:62): “contradictory terminal data reported”; design.md:164: “矛盾 item/终态为 protocol”. `ai/responses_kimi_decode.go:67–75` compares an item against preceding completed arguments only for function calls. Text and summary completions stored by `ai/responses_kimi_lifecycle.go:109–124` are never compared. Trigger: `output_item.added`, `output_text.done` with `A`, then `output_item.done` with `B` and matching completed output `B`. The call succeeds and substitutes `B`; the same gap affects reasoning summaries. Compare completed parts with item/terminal content before accepting it.

- **[P1] Bind qualification to actual scenario request controls.** Spec K16 (spec.md:195): “六路线各自完整有限场景”; design.md:301–305 fixes high/low/clamped-high and required/any/auto. `ai/internal/testkit/supportmatrix/regional_capture.go:71–74` validates actual wire only for model/stream. `regional_replay.go:65–106` generates its own request from saved options; `RoundTrip` (35–40) checks only URL. A resealed small L02 capture whose actual effort changes high→low, retaining high in options.json, passes these semantic checks: response replay and the 2048 minimum reservation remain identical. These accepted proofs feed `kimi_promotion.go:36–43`. Independently check captured controls against every scenario, including effort presence/defaults and tool choice.

No additional scope-creep finding. Pending paid qualification, formal activation, race and pressure correctly remain authorization-gated; this review does not classify their NOT_RUN status as a defect.


Standards: 2 documented-standard findings (highest P2) and 1 naming judgement. Spec: 3 findings (highest P1).
