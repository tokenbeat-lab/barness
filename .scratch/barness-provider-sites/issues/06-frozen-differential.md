# 06 冻结基线地区差分与缓存 v2 账本

Status: claimed
Type: task
Blocked by: 01, 02, 03

## Context

[规格](../spec.md) R05 R06 R11 R18 R20，RP06/RP18。冻结 pi 数据实际含 zai Chat glm-5.3/5.2 与 minimax-cn Messages M3。当前 wrapper 没列出该数据不能作为“不在冻结基线”的依据。

## Ownership

冻结 pi wrapper 的已有 Provider export 选择、新地区实际差分 E2E/fixture、精确缓存 v2 账本条目与相关 wrapper 测试；不变更上游版本、模型数据、SDK 或完整来源 pins。与 04 的 supportmatrix/release/live 并行，各自提交，05 等待两者。

## Acceptance

公共 Client 对冻结 pi 的适用两个新路线及三个型号实际比较；其余四路线登记扩展并独立 fixture。所有旧8差分保留；精确记录 v2 字段变化，不免除其他 wire/归属/usage 比较。先失败再实现，保留实际 argv/exit/时间/来源/原生工件。无付费、无 race。

## Comments

2026-10-20：从票04拆出可独立并行的真实差分工作；额外测试/工具工作不降低原验收合同。
