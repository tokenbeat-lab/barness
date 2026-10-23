# 08 — 真实中国站 Chat 缺少缓存读取计数

Type: task
Status: resolved

## Context

- [Spec R09–R10 / 唯一当前状态](../spec.md#唯一当前状态)
- [真实失败及确定性回放](../evidence/candidate-live-result-01.md)
- [ADR-0052](../../../docs/adr/0052-barness-ai-kimi-cost-estimates.md)

## Scope

用户已选择保守费用区间，并明确 cached_tokens 是可选字段，缺失补0。Kimi Chat/Responses 先使用原有 nested/legacy 值，缺失/null默认0参与完整性及计费；显式非法值仍拒绝。Messages和其他Provider保持原合同。

总输入P、输出O、读取R已知而写量W缺失时，以0≤W≤P−R约束所有允许分配，结合实际响应TTL及固定价格求有限上下界。用量仍partial，成本相等为resolved、不同为bounded，无法证明完整界时partial/unreported。不得伪造未知写量或实际1h观察；点值取下界、预算取上界。真实原始SSE在新公共Client离线回放预期CacheRead0、UsageComplete、CostResolved。

## Acceptance

先在公共Client/CLI seam留存失败，再实现、定向转绿、双轴审查和冻结完整普通验收。覆盖两站Chat/Responses默认/显式/非法读取量、真实捕获、缺写量分区、TTL、相等/反转费率与分档配置继续拒绝、溢出、不可定界及Observer；既有Provider回归保持。旧真实失败及已消耗grant永久保留，新来源资格不能借回放或旧pin取得；真实重跑须另获具体范围批准。完整实际状态只在spec维护。

## Comments

2026-10-23：六路线候选批准后首个真实用例触发此缺项，按首失败停止；1调用/1Attempt，其他53个计划调用未执行。公共 Client 合法保留成功与 partial；确定性无网络回放及最小化复现了同一资格拒绝，显式0/1的合成对照均正确归一化。尚无需要改动运行时的证据，当前缺少厂商报告合同/完整响应依据。完整实际事实与原件指针见 Context；本票不另维护功能当前状态，也不授予任何真实重试。

2026-10-23：用户明确选择第二种解决方案，并补充cached_tokens可选、缺失补0；据此修订spec/design/ADR0052并认领实现，不再等待厂商将缺失字段补齐。实施代码与新来源资格分开判定。

## Answer

已实现用户确认的可选cached_tokens默认0，以及独立于用量完整性的有限费用边界；未知写量不伪造成观察，非法计数/矛盾总量仍拒绝。真实中国站SSE原样回放得到CacheRead0、UsageComplete、CostResolved。两轴审查均PASS、0发现；冻结完整普通验收在2ffa761c437cdb6403cbaf05c7f1846e11965f49实际exit0，14045个必需用例全部PASS、主及五辅助审计零发现。详见[当前状态](../spec.md#唯一当前状态)、[实施证据](../evidence/cache-bounds-implementation-01.md)及[本轮审查](../evidence/cache-bounds-review-01.md)。本票代码和普通验收关闭；新来源真实资格留在07，旧失败/授权不改写。
