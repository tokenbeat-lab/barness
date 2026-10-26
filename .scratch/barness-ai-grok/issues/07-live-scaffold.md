# 独立 xAI 候选 live 入口与 Host 说明

Type: task
Status: claimed
Blocked by: 01

## Contract

[spec](../spec.md) R21/R22/R24、E17；[design](../design.md) 16；[SOP](../../../docs/agents/verification.md)。

在普通阶段准备明确、有限的自身 live 场景与装配说明，沿用现有 Client、resolver、recorder 与证据。精确入口隔离、无 Key/未授权不触发 HTTP、预算有界、candidate 不产生 formal 资格。尚未获得付费授权，本工单只交付入口与无厂商公共进程检查。

## Ownership

`ai/live/xai*_test.go`、live TestMain 的最小接线、普通 `ai/e2e/xai_live_process_test.go`、Host 装配文档及必需候选证据形状。实现说明见外部研究笔记 `/tmp/barness-grok-live-notes.md`（后续交付须保留到功能目录）。不改请求/history/stream/usage 实现，不激活正式目录，不运行任何厂商/race/压力。

## Verification

先建立无 Key、身份/预算/入口错误、混合凭据、取消/最终写入失败等公开子进程失败场景，再实现。仅运行普通公共 E2E/受控子进程，经 `tools/verification/run.py` 留存。

## Answer

独立候选入口与 Host 文档已实现，等待最新请求/历史/流/用量运行时合入后完成 controlled 全场景验证。

- 入口 `^TestXAICandidate$`，显式 `internal/testkit/xai.CandidateCatalog()`，全球固定 endpoint 与 grok-4.7；复用公共 Client、现有 resolver、recorder/Observer、evidence Run。
- 完整计划为 18 场景、24 调用、零自动重试；四入口、默认和四档 effort、非零 reasoning 算术、完整加密续接与 Host 信任恢复、auto/required/none/并行工具、用户及工具结果 PNG/JPEG、缓存/usage/实际型号。
- 原生严格 grant 沿用 `BARNESS_AI_LIVE_PERMISSIONS_FILE`，固定来源/二进制/目录/身份/别名/完整计划/有限预算；common Git 排他 claim 在任何请求前消费，失败/取消不可重用。未扩大全局环境转发白名单，controlled 字段不进入 paid 转发。
- [Host 装配](../../../../docs/barness-ai/xai.md) 覆盖现有公共 Catalog/resolver、白名单、凭据快照、资源与超时、持久化、工具、费用及禁用/版本回滚；外部 Host 不导入 internal testkit。
- 所有报告明确 candidate，controlled 的 vendorHTTPAttempts=0；不写正式 support report/支持矩阵、不激活内置目录。正式消费者拒绝检查已纳入普通进程验收。

本工作树 `/tmp/barness-grok-live-scaffold` 保留普通原件。首个 red：`.evidence/barness-ai-grok/e2e/20261026T020919.278441Z-j12llsmn`（缺自身候选报告，exit 1）；配置/取消门禁 green：`20261026T021340.425058Z-pfas5r5r`；后续严格字段/代理/路线边界 red 与 green 均另留 attempt。最新 selected-route/no-switch green 为 `.evidence/barness-ai-grok/e2e/20261026T022350.363064Z-3t5pnyn3`（exit 0）。完整 controlled 计划目前在旧运行时用量转换处保持 red：`.evidence/barness-ai-grok/e2e/20261026T021614.466774Z-wzufh6vy`，只触发一次合成 HTTP，未触发厂商。

paid live、race、压力均 NOT_RUN；未取得本次独立授权，不运行后续阶段。旧地区/Kimi/GLM 来源库存受新增源码影响，当前资格不能继承历史 PASS；由05/06与最终影响审查处理。

## Comments

- 2026-10-26：从04拆出独立准备工作，允许01后与02/03并行；05仍负责真实资格。
