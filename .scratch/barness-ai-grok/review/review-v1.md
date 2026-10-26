# Grok 两轴审查 v1

Base: `2f58c68d8fbb30a07e71b9666feed9979c791d01`
Head: `a3c3a269d15c35563d2ec3c364595a7f179e0849`
Diff: `git diff 2f58c68d8fbb30a07e71b9666feed9979c791d01...a3c3a269d15c35563d2ec3c364595a7f179e0849`

只读审查，未执行测试。单一修复代理在 `codex/grok-review-fixes` 补公共失败场景并修复；原审查不覆盖为新来源通过。

## Standards

Hard violations:

- **[P1] Authorize assistant IDs when `type` is omitted.** [responses_xai_input.go:132](/Users/cyber/RestoX/barness/ai/responses_xai_input.go:132) only fingerprints native message references for `case "message"`, but lines 55–66 accept an absent/null type and allow an assistant `id`. An `OnPayload` replacement such as `{role:"assistant",id:"foreign",content:"history"}` therefore produces no authorization key, passes final validation, and reaches HTTP with an ID absent from the trusted baseline. This breaches **ai/README.md, invariant 9** (hooks cannot widen native references), **AGENTS.md §4** (tenant isolation/minimum authority), and **ADR-0005** (final native-reference authorization). Normalize the omitted message discriminator before reference checking, or reject IDs on shorthand messages; verify omitted/null-type injections through the public hook boundary.

- **[P2] Keep the retained report non-PASS until finalization succeeds.** [xai_finish_test.go:66](/Users/cyber/RestoX/barness/ai/live/xai_finish_test.go:66) promotes the report containing `Outcome=PASS`, `Complete=true`, and `AuditPassed=true` before opening/syncing its directory and running the final audit. Any failure at lines 69–80 returns exit 1 while leaving those success claims in `xai-candidate-report.json`. This breaches **AGENTS.md §9** (verifiable, observable failures) and **docs/agents/verification.md §§7,9** (actual outcome and retained evidence agree). Audit staged contents before promotion and fail-close the retained status on post-promotion errors; add a public process failure case beyond the existing pre-rename write fault.

Judgement calls: no additional actionable baseline smell identified. Provider-private state/codecs and deliberately independent persistence contracts follow the documented design.

Read-only review; no checks executed. Final frozen ordinary, live, race and pressure runs remain explicitly pending, which is correctly disclosed rather than a standards violation.


## Spec

- **[P2] xAI 的第二套缓存路由头仍可绕过守卫。** R16 要求“SessionID仅通过已有tenant/provider/account/session派生键发送”；设计第316行明确要求拒绝 Hook 补入 `x-grok-conv-id`。`ai/responses.go:24–34` 没有 xAI 专用检查，公共 `guardedHeaders`（`ai/hooks_run.go:29–34`）也未保护此头。`TransformHeaders` 写入任意 `X-Grok-Conv-Id` 后会直接发送，允许原始或其他作用域的会话标识替代已派生的缓存路由。应在 xAI 的 HTTP 前边界拒绝，并先补公共失败场景。

- **[P2] 最终守卫接受重复协议控制键。** R07 要求“最终wire子集由设计中的字段表固定”，设计第318行要求“控制字段重复键拒绝”。`ai/responses_xai_request.go:46–59` 直接解码为 map，重复键会被折叠。例如无 Hook 时，SamplingParams 的 reasoning 原文 `{"effort":"off","effort":"high"}` 被第28–39行原样合并；守卫只看到最后的 `high`，实际 HTTP 仍携带两个 effort。Hook 放入同类 RawMessage 也可触发。需在折叠前检查控制对象的唯一性，避免守卫和厂商选择不同值。

- **[P2] Grok 工具历史向 OpenAI Responses 降级会合并关联。** R13 要求降级“保留可见文本与工具配对”，E18 包含共享 ID 回归。新增 `ai/responses_xai_ids.go:15–17` 的所有 ID 都以 `xar1|` 开头，但 `opaqueHistoryIDSource`（`ai/opaque_history_ids.go:139–151`）未识别 xAI。Host 明确选择 OpenAI 继续含两个 Grok 工具调用的历史时，`ai/responses_history.go:43–54` 将两者都归一成 `call_id=xar1`；结果也映射至同一调用，造成配对歧义或上游拒绝。应加入保留完整来源 ID 的无碰撞降级。

未发现额外范围扩张。以上均为静态路径核对，未运行测试。最终冻结来源的完整普通回执仍待执行；真实 live、正式晋升及人工 race/压力分别待授权/资格，候选目录未激活属于正确的当前状态。


Standards：2项，最高P1；Spec：3项，最高P2。两轴分别保留，不合并排序。
