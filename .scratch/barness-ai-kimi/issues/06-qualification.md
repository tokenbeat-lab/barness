# 06 — 六路线资格入口与证据消费

Type: task
Status: resolved
Blocked by: 01

## Context

- [Spec](../spec.md)
- [设计](../design.md)
- [研究](../research.md)
- [唯一当前状态](../spec.md#唯一当前状态)
- [测试 SOP](../../../docs/agents/verification.md)

## Scope

R01、R12；K16。复用原生证据设施增加六路线有限 live 入口、只读资格消费和受控进程验证；未获授权前不得请求厂商或晋升正式目录。

## Acceptance

在已确认的公共 Client E2E seam 先留存失败、再实现并转绿；使用 SOP run.py 保留命令、退出与原生 bundle。文件低于 500 行；保留其他 Provider 合同。完成时记录提交、回执指针及未执行项。

## Comments

2026-10-22：从已批准实施的 spec/design 派生；任务依赖是集成顺序，独立协议可在同一基线并行开发，合并前同步集成分支。

2026-10-22：按实际模块依赖细化：资格基础设施直接依赖候选与共享运行合同，可与协议/Host 并行开发；完整六路线的受控验收和源码冻结仍由工单07阻塞，不提前声称路线完成。

2026-10-22：codex/kimi-qualification 基线 d15f642，公共 Client/真实原生子进程受控 seam；实现路线化已有 regional 合同，不运行真实 live/race/压力或目录晋升。

## Answer

2026-10-22：完成独立六路线 9 场景原生入口、有限 USD/一用授权及账本、source/binary 绑定、真实 L01 continuation、只读 replay/CostEstimate 消费与独立 proof-gated 晋升入口。最终受控 Kimi+旧 Regional 普通范围 814 叶测试通过，六路线各 9 场景及消费者通过；[受测提交、完整原件/失败与收尾索引](../evidence/ticket-06.md)。当前正式 Kimi payload inactive；live/race/压力、真实资格与实际晋升未执行，最终整合验收由07负责。
