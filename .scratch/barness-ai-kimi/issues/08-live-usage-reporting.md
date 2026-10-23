# 08 — 真实中国站 Chat 缺少缓存读取计数

Type: research
Status: needs-info

## Context

- [Spec R09–R10 / 唯一当前状态](../spec.md#唯一当前状态)
- [真实失败及确定性回放](../evidence/candidate-live-result-01.md)
- [ADR-0052](../../../docs/adr/0052-barness-ai-kimi-cost-estimates.md)

## Question

在保持缺失与零、用量与成本可靠性独立的合同下，如何让中国站 Chat 返回满足自身资格的完整计费依据？当前自身真实响应同时缺少 nested 与 legacy top-level `cached_tokens`，只有 cache_write_tokens 显式0；不能推断未上报读取量为0。

## Acceptance

取得可核实的官方遗漏语义或服务端完整字段证据，并明确其适用的市场/型号/协议和成本影响。若需要改变语义，先审阅 spec/ADR 与公共 E2E；不得直接把真实缺项填0、降低资格条件、删路线或伪造完整费率。新来源/输入需冻结普通验收；真实重跑需要独立新 grant/scope 和付费授权，旧失败永久保留。解决前07不能 resolved，六路线不能晋升正式目录。

## Comments

2026-10-23：六路线候选批准后首个真实用例触发此缺项，按首失败停止；1调用/1Attempt，其他53个计划调用未执行。公共 Client 合法保留成功与 partial；确定性无网络回放及最小化复现了同一资格拒绝，显式0/1的合成对照均正确归一化。尚无需要改动运行时的证据，当前缺少厂商报告合同/完整响应依据。完整实际事实与原件指针见 Context；本票不另维护功能当前状态，也不授予任何真实重试。
