# Integration two-axis code review

Base: 9a43ec0827b0ab794501682a7e66037a592a2f8d
Tip: d2bb122efb7588218661fb4e26d01349e9dd7f50
Diff: `git diff 9a43ec0827b0ab794501682a7e66037a592a2f8d...d2bb122efb7588218661fb4e26d01349e9dd7f50`

Read-only review; no new executions or vendor qualifications.

## Standards

- **[P2] Documented-standard breach — revised tiers silently acquire zero prices.** [revise.go:25](/Users/cyber/RestoX/barness/catalog/maintenance/revise.go:25) overlays the definition and returns without the `resolveRequiredFields`/`missingTierFields` checks used by additions. For an existing chat/classifier, `{"cost":{"tiers":[{"inputTokensAbove":100,"input":1}]}}` can become a frozen candidate: shape validation checks only present fields, and typed generation fills omitted output/cache rates with zero. Controlled checks then compare against those same invented prices. Validate the final revised raw definition before canonicalization. Violates **AGENTS.md §3** (validate at boundaries) and **ADR-0058, Decisions** (“缺价不补零”); see also `docs/barness-ai/catalog-maintenance.md:111` (complete tier rates).

- **[P2] Documented-standard breach — a failed apply can still replace runtime output or split it from current.** [check_apply.go:134](/Users/cyber/RestoX/barness/catalog/maintenance/check_apply.go:134) treats every `atomicReplace` error as if replacement did not occur, but [export.go:50](/Users/cyber/RestoX/barness/catalog/maintenance/export.go:50) renames first and may fail afterward in directory sync. An output-directory sync failure leaves the new output with the old current pointer. A current-directory sync failure leaves the new pointer while lines 139–145 restore the old output. Distinguish pre-install failures from post-install durability failures and keep the two artifacts on the same revision. Violates **AGENTS.md §5** (transaction/failure boundaries) and **§9** (verifiable rollback).

No additional substantive baseline smell findings. Static review only; no tests executed.


## Spec

1. **[P1] Validate tier field presence on `revise`.** `catalog/maintenance/revise.go:25–38` skips `resolveRequiredFields`/`missingTierFields`. Revising GPT-4 with `{"cost":{"tiers":[{"inputTokensAbove":100,"input":2}]}}` therefore freezes omitted output/cache rates as typed zeroes. `ai/usage.go:94–107` accepts those zeroes, and candidate contracts calculate expectations from them, so check/apply can accept unknown rates as free. Spec: “维护边界检查字段是否实际提供” (spec.md:154), “阶梯需确认完整费率” (:156). Apply presence validation before canonicalization for every revision path.

2. **[P2] Reject cross-operation serving associations and refreshes.** `catalog/maintenance/revise.go:62–74` validates only Provider and wire ID; `cost_update.go:41–50` likewise ignores offering type. A classifier `openai/openai-decisions/gpt-6-luna` can acquire a same-ID `type:"chat"` association and source-owned cost fields, then import chat prices. Existing associations also remain refreshable after upstream type changes. Spec: “不能跨服务方、地区或操作借用路由与价格” (:109), “6 条图片和分类记录…不伪造关联或借价” (:121). Validate source operation whenever establishing or consuming the association.

3. **[P2] Preserve usable provenance across default-source adoption.** `catalog/maintenance/store.go:31–42` creates an absent-store seed without applied candidates. After the documented Provider/snapshot adoption, `review_export.go:147–149,173–195,207–209` requires every nonhistorical source digest but searches only applied History candidates. Thus the next review export fails `source_unavailable`, even with the prior review bundle retained in the checkout; no bundle/import lookup exists. Spec: “实际取回的数据必须固定为可重放输入” (:108), “唯一维护源按 Provider 管理来源…与历史” (:99). Make adopted provenance resolvable from the new baseline and cover a second maintenance/export cycle.

4. **[P2] Route generator parse errors through JSON.** `cmd/cataloggen/main.go:31` uses global `flag.Parse()` with ExitOnError. Unknown options or missing flag values exit 2 with plain stderr, bypassing the structured response at :16–18. Spec: “每个变更入口提供…JSON 输出” (:129), “错误有稳定结构化类别和明确非零退出” (:130). Use a ContinueOnError FlagSet and encode parsing failures consistently.

Read-only review; no tests executed. Spec findings: 4; worst severity: P1.


## Resolution

PENDING: all findings assigned to one implementer; final source freeze/ordinary verification follows fixes and re-review. The tier finding appears independently in both axes; axis counts and severities remain independent.
