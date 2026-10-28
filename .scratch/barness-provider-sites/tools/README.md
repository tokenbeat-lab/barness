> 固定轮次控制器已按 [ADR-0055](../../../docs/adr/0055-route-scoped-evidence-applicability.md) 从活动实现退役。原暂停、授权、账本和失败事实保留；本页旧接口仅按固定 Git 版本追溯，不再作为执行入口。

# 地区 Provider 验收工具入口

唯一当前状态见[地区 Provider spec 的当前状态](../spec.md#当前状态)，对应[恢复7当前索引](../../../.evidence/provider-sites/final-acceptance-20261020/recovery7-preparation/index.json)与[暂停交接](../../../.evidence/provider-sites/final-acceptance-20261020/recovery7-preparation/pause-handoff.md)。本页只导航，不另维护轮次、预算、授权或发布结果。

截至本次文档修订（2026-10-24），recovery7 在完整 native 普通门禁失败后按用户要求暂停；没有新的执行入口获准启动。恢复6保持关闭，新7授权、后续构建、付费和正式 race 均不因这次入口修订而继续。具体已执行与未执行事实以 spec/index 为准，release.pass 仍为 false。

## 当前操作合同

- 测试与发布步骤、选择边界及完成条件统一见[测试 SOP](../../../docs/agents/verification.md)和[ADR-0051](../../../docs/adr/0051-scoped-verification-receipts.md)。新工作不续建代际恢复控制器。
- 已有恢复7实现及待办见[工单24](../issues/24-minimax-native-authorization-timeout.md)、[工单25](../issues/25-recovery7-native-authority.md)、[工单26](../issues/26-recovery7-fixed-continuation.md)。这些链接说明当前工作，不是继续执行的授权。
- 当前没有可从本页复制执行的恢复命令。任何后续工作须先结合唯一状态、暂停要求、具体来源和各原生合同确定范围；prepare/check 也可能构建或运行测试，不能因名称推断为只读。
- 历史 grant、claim、nonce、账本和失败/UNKNOWN 原件保持；关闭轮次和未用余额不转为新执行权限。普通验证、候选、正式资格及业务效果各自满足原生合同，离线成功不产生真实资格。

## R2–R4 历史合同与命令

[修订前完整 README](../../scratch-artifact-remediation/history/phase2-before-20261024/provider-tools-README.md)逐字节保存 R2、单次 Responses retry、恢复3与恢复4的合同、金额和命令。其 Git、哈希及原相对链接基址见[历史索引](../../scratch-artifact-remediation/history/phase2-before-20261024/index.json)。

历史原文里的 current/当前/批准/准备入口描述只适用于当时固定来源和已记录授权，不是当前操作指引。阅读时按原基址解析引用，不执行旧 prepare/run-selected/finalize 或整轮验证来修补文档。旧控制器源码及原件不在本次修订中改写、迁移或重编号；recovery7 暂停状态也不因历史材料可读而改变。
