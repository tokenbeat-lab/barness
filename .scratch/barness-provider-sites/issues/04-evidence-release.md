# 04 自身证据与发布追加门禁

Status: claimed
Type: task
Blocked by: 01, 02, 03

## Context

[规格](../spec.md)、[设计](../design.md)、[验证策略](../../../docs/agents/verification.md)。需求：R02 R16 R17 R18 R20。

## Ownership

现有 live、支持矩阵、差分扩展、release 消费者追加新路线/型号与身份/来源校验，保留全部旧门禁和原件；CLI/消费者失败场景优先。

## Acceptance

先在规格指定的公共 Client / Host / CLI 接缝建立失败场景，再实现；保留实际命令、退出码、时间、执行来源、工件及 Hash。先读取相关 ADR 和验证策略。不读取或输出真实秘密；不调用付费服务，不运行 race。按实际证明关闭实现，不用普通 replay 代替自身 live。

## Comments

2026-10-20：由 implement-spec 根据已提供规格建立；无 GitHub PR 关闭义务。各票完成后在 Answer 留实现/验证指针，整体状态唯一维护于 spec。
