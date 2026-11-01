# 受控公共协议合约与候选检查应用

Type: task
Status: ready-for-agent
Blocked by: 01, 02

## Context

[Spec](../spec.md)、[设计](../design.md)、[当前状态](../status.md)、[验证 SOP](../../../docs/agents/verification.md)。

## Ownership and outcome

负责候选 check 的真实进程执行和严格证据消费、CLI check/apply 闭环及每身份公共协议合约。每条最终四维身份都必须有适用合约和实际必需子场景，验证请求/响应、Usage/价格以及声明能力。保留实际代码/冻结来源规则/候选/基准/版本、命令、退出和不可变原件；apply 重新核对来源集合与证据并写原受测快照，不重新取源或重映射。不依赖 ai/internal/testkit 的生产 API，不调用真实厂商。缺能力时明确 code_gap，不放宽 Kimi/xAI 固定约束。

## Acceptance and verification

T05、T06、T11、T12、T13；CLI → 真实受控检查 → apply → 加载 → Client，含失败、取消、超时、篡改和旧证据拒绝。
先在已确认公共边界建立失败场景，再实现。遵循 TDD 技能及仓库 SOP，失败原件保留。单文件不超过 500 行。不得执行未获单独授权的 race、压力或真实厂商请求。

## Comments

2026-11-01：从已确认 spec 建立实施依赖图。

2026-11-01：候选/存储接口已明确，可与 03 并行实施 runner/证据消费；本工单完成前仍须合并 03 并验证完整 check/apply CLI 闭环。设计笔记：`/tmp/barness-catalog-check-notes.md`。
