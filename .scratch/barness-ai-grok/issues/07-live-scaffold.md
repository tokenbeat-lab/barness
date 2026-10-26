# 独立 xAI 候选 live 入口与 Host 说明

Type: task
Status: ready-for-agent
Blocked by: 01

## Contract

[spec](../spec.md) R21/R22/R24、E17；[design](../design.md) 16；[SOP](../../../docs/agents/verification.md)。

在普通阶段准备明确、有限的自身 live 场景与装配说明，沿用现有 Client、resolver、recorder 与证据。精确入口隔离、无 Key/未授权不触发 HTTP、预算有界、candidate 不产生 formal 资格。尚未获得付费授权，本工单只交付入口与无厂商公共进程检查。

## Ownership

`ai/live/xai*_test.go`、live TestMain 的最小接线、普通 `ai/e2e/xai_live_process_test.go`、Host 装配文档及必需候选证据形状。实现说明见[live 接入研究笔记](../live-implementation-notes.md)（研究阶段原件位于 `/tmp/barness-grok-live-notes.md`）。不改请求/history/stream/usage 实现，不激活正式目录，不运行任何厂商/race/压力。

## Verification

先建立无 Key、身份/预算/入口错误、混合凭据、取消/最终写入失败等公开子进程失败场景，再实现。仅运行普通公共 E2E/受控子进程，经 `tools/verification/run.py` 留存。

## Answer

待实施。

## Comments

- 2026-10-26：从04拆出独立准备工作，允许01后与02/03并行；05仍负责真实资格。
