# 01 — Provider 身份、六个候选型号与思考元数据

Status: resolved
Type: task
Updated: 2026-10-18
Blocked by: None

## Context

[交付规格](../spec.md)、[研究](../research.md)、[任务图](../map.md)。本票继承完整 spec，不以摘要缩减合同。

## Scope and acceptance

负责共享 Provider、型号元数据及候选目录。建立零值兼容、序列化/复制/Hash、思考默认标签与 mandatory/binary 语义；不要提前发布未经自身 live 验收的 BuiltinCatalog。三协议条目独立，候选合成费率明确标记。公共 Client E2E 先列失败方式、保留 red/green 证据；为后续协议实现写明共享接口。

追溯：R01–R08、R18；A01–A05、A15；US01、US09–US26、US50–US53。

## Execution

在自己的分支/worktree 从 `codex/minimax-cn` 当前 tip 开始；使用 tdd 技能，测试接缝已由 spec T01 确认。先登记失败方式，再逐垂直切片 red→green，产出可重放合成请求/SSE/事件/Result/Attempt/Observer 与脱敏/来源记录。实现前阅读相关 ADR。不能回退其他工作；单文件 ≤500 行。完成后合入集成分支 tip，在 Answer 附提交、验证及证据指针；真实缺口保持可见。

## Answer

2026-10-18：共享 ProviderMiniMaxCN、零值兼容的 Model.DefaultThinkingLevel 与配置校验已实现；默认标签只影响省略 reasoning 的采样选择，进入 JSON/Hash/Client 快照，不派生 wire effort。复用 RequiresThinking 和档位 map 表达 mandatory/binary，不引入冗余开关。

六个独立文本候选及三组 Catalog/Binding 装配位于 `ai/internal/testkit/minimax/`。所有费率明确标为 synthetic，未进入 BuiltinCatalog；官方来源 Hash、容量上限及候选限制已记录。关键决策：ADR-0037。三协议最终校验、授权/认证路由、真实价格和 live 证据属于后续票，不由共享 E2E 冒充证明。

验证：30 个自身公共元数据/Client 接缝 cases 普通与 race PASS；既有采样 options/snapshots/configuration 和 GLM mandatory metadata 回归 PASS；`go vet ./ai ./ai/internal/testkit/minimax` PASS。red 及测试装配失误的失败历史保留于 [证据索引](../evidence/ticket-01/README.md)，完整包以可校验压缩件保留。
