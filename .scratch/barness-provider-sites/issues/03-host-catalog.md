# 03 Host 装配与独立候选目录

Status: claimed
Type: task
Blocked by: none

## Context

[规格](../spec.md)、[设计](../design.md)、[验证策略](../../../docs/agents/verification.md)。需求：R02 R03 R04 R07 R08 R09 R15 R16 R19。

## Ownership

ai/examples/glmhost、ai/examples/minimaxhost、各调用者迁移；独立候选与官方来源/价格记录；公共 Host/Catalog E2E。未获自身 live 证明不得晋升 BuiltinCatalog。

## Acceptance

先在规格指定的公共 Client / Host / CLI 接缝建立失败场景，再实现；保留实际命令、退出码、时间、执行来源、工件及 Hash。先读取相关 ADR 和验证策略。不读取或输出真实秘密；不调用付费服务，不运行 race。按实际证明关闭实现，不用普通 replay 代替自身 live。

## Comments

2026-10-20：由 implement-spec 根据已提供规格建立；无 GitHub PR 关闭义务。各票完成后在 Answer 留实现/验证指针，整体状态唯一维护于 spec。
