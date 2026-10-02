# 25: 失败响应携带的用量记录到尝试上

**What to build:** Provider 以 `response.failed`（及其他协议的同类失败终态）结束、但随之报告了用量时，把这份用量按同一价格快照计价后记在该尝试的 `Attempt.Usage` 上，`UsageReporting` 按实际报告程度标注；兼容消息的 Usage 仍与 pi 一致保持零值。这样 Provider 已计费的失败请求在元数据与观测中可见，不再只能依赖厂商账单对账（spec I10"不能从零值推断失败请求免费"）。

**Blocked by:** 15

**Status:** needs-triage

**Context:** 工单 15 的初稿已实现过这一行为（`responsesParser.failed` 调用 `responsesUsage.of` 并记到尝试上），维护者于 2026-10-02 决定首期先与 pi 一致丢弃，见 ADR-0010 决策三。恢复时需同步修改 `ai/e2e/testdata/responses/usage.json` 中 `failed-usage-zero`、`failed-usage-reported` 的预期，以及 ADR-0010 决策三的 Consequences。

- [ ] Responses `response.failed` 的用量记在尝试上并计价，消息 Usage 仍为 pi 的零值
- [ ] 后续协议（Anthropic、Gemini、Chat Completions）的失败终态同样处理
- [ ] 差分仍无待处理差异；尝试上的用量作为扩展单独断言
