# 06 — 固定 MiniMax pi 差分和扩展证据

Status: resolved
Type: task
Updated: 2026-10-18
Blocked by: 02, 03, 04

## Context

[交付规格](../spec.md)、[研究](../research.md)、[任务图](../map.md)。本票继承完整 spec，不以摘要缩减合同。

## Scope and acceptance

冻结 pi 1.1.0 pins 不变；装配 MiniMax Messages M3 catalog 并运行适用文本/工具差分，逐 case/path 解释差异；Chat/Responses 和 M3.1 登记扩展而非伪差分。旧差分覆盖不减少。

追溯：R05、R24；A19；US71–US75。

## Execution

在自己的分支/worktree 从 `codex/minimax-cn` 当前 tip 开始；使用 tdd 技能，测试接缝已由 spec T01 确认。先登记失败方式，再逐垂直切片 red→green，产出可重放合成请求/SSE/事件/Result/Attempt/Observer 与脱敏/来源记录。实现前阅读相关 ADR。不能回退其他工作；单文件 ≤500 行。完成后合入集成分支 tip，在 Answer 附提交、验证及证据指针；真实缺口保持可见。

## Answer

2026-10-18：wrapper 装配已有冻结 MINIMAX_CN_MODELS export，pi1.1.0/commit/完整数据/源码/SDK lock pins 均未改变。M3 Messages full/simple 文本、签名思考+双工具调用及 Host 完整结果续接实际产生6条成对差分；2条完整目录/缺失身份核对、1条分类边界及20条五个路线/型号扩展的四入口验收，共29个自身case普通与race全部PASS，pending0。

中央账本新增41条精确case/path/kind说明和MiniMax Chat/Responses两条Extension Route，317条旧决策与6条旧路线逐对象保持原样。M3.1三协议fixture明确pi skip/model extension，不借用M3身份；Messages整条路线不豁免。现有70条Anthropic原生差分、pioracle tests、9条runner guards和targeted vet通过。

[有限差异与来源](../../../ai/e2e/testdata/minimax-pidiff/README.md)、[失败先行与不可变证据索引](../evidence/ticket-06/README.md)记录完整请求/SSE/事件/Result/Attempt/Observer、来源/目录/SDK哈希、red及测试装配修正历史。所有运行离线合成，未读取真实Key或调用厂商；合成候选费率不证明当前RMB→USD价格，不声称live或正式目录支持。
