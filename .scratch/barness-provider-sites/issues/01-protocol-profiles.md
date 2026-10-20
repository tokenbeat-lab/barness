# 01 协议配置与地区身份

Status: ready-for-agent
Type: task
Blocked by: none

## Context

[规格](../spec.md)、[设计](../design.md)、[验证策略](../../../docs/agents/verification.md)。需求：R01 R05 R06 R10 R12。

## Ownership

ai/binding.go；三套协议配置及请求、解析、历史、用量与最终校验路径；新增 regional protocol E2E。真实 Provider 不改写，逐分支区分协议形状与身份语义。

## Acceptance

先在规格指定的公共 Client / Host / CLI 接缝建立失败场景，再实现；保留实际命令、退出码、时间、执行来源、工件及 Hash。先读取相关 ADR 和验证策略。不读取或输出真实秘密；不调用付费服务，不运行 race。按实际证明关闭实现，不用普通 replay 代替自身 live。

## Comments

2026-10-20：由 implement-spec 根据已提供规格建立；无 GitHub PR 关闭义务。各票完成后在 Answer 留实现/验证指针，整体状态唯一维护于 spec。
