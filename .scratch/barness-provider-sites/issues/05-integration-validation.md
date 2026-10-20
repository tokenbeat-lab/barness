# 05 整合验证与迁移交付

Status: claimed
Type: task
Blocked by: 01, 02, 03, 04

## Context

[规格](../spec.md)、[设计](../design.md)、[验证策略](../../../docs/agents/verification.md)。需求：R01–R20，RP01–RP20。

## Ownership

普通 E2E、适用旧服务回归、双 vet、关键边界审查和双轴 code-review；补齐迁移/回滚与当前状态。付费/race 无授权保持 NOT_RUN，正式验收保持开放。

## Acceptance

先在规格指定的公共 Client / Host / CLI 接缝建立失败场景，再实现；保留实际命令、退出码、时间、执行来源、工件及 Hash。先读取相关 ADR 和验证策略。不读取或输出真实秘密；不调用付费服务，不运行 race。按实际证明关闭实现，不用普通 replay 代替自身 live。

## Comments

2026-10-20：由 implement-spec 根据已提供规格建立；无 GitHub PR 关闭义务。各票完成后在 Answer 留实现/验证指针，整体状态唯一维护于 spec。

2026-10-20：主代理已留存旧服务普通基线：`.evidence/provider-sites/baseline/receipt.json`（exit 0；3376 场景：3367 PASS / 9 NOT_RUN；未启用差分/压力）。此票待前置实现合并后完成最终整合，不以该基线代替新功能结果。
