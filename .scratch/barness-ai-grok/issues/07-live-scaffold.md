# 独立 xAI 候选 live 入口与 Host 说明

Type: task
Status: resolved
Blocked by: 01

## Contract

[spec](../spec.md) R21/R22/R24、E17；[design](../design.md) 16；[SOP](../../../docs/agents/verification.md)。

在普通阶段准备明确、有限的自身 live 场景与装配说明，沿用现有 Client、resolver、recorder 与证据。精确入口隔离、无 Key/未授权不触发 HTTP、预算有界、candidate 不产生 formal 资格。尚未获得付费授权，本工单只交付入口与无厂商公共进程检查。

## Ownership

`ai/live/xai*_test.go`、live TestMain 的最小接线、普通 `ai/e2e/xai_live_process_test.go`、Host 装配文档及必需候选证据形状。实现说明见[live 接入研究笔记](../live-implementation-notes.md)（研究阶段原件位于 `/tmp/barness-grok-live-notes.md`）。不改请求/history/stream/usage 实现，不激活正式目录，不运行任何厂商/race/压力。

## Verification

先建立无 Key、身份/预算/入口错误、混合凭据、取消/最终写入失败等公开子进程失败场景，再实现。仅运行普通公共 E2E/受控子进程，经 `tools/verification/run.py` 留存。

## Answer

已交付独立候选 runner 与完整 Host 装配说明；本票普通受控进程范围完成。

- 精确入口 `^TestXAICandidate$`，显式 `internal/testkit/xai.CandidateCatalog()`，固定全球目标和 grok-4.7；复用公共 Client、现有 resolver、recorder/Observer、evidence Run。
- 有限完整计划为 18 场景、24 调用、零自动重试：四入口、默认/low/medium/high/xhigh、非零 reasoning 算术、JSON 持久化后完整密文续接和显式 Host 信任恢复、auto/required/none/并行工具、用户和工具结果 PNG/JPEG、缓存/usage/返回型号。
- 严格一次性 grant 沿用通用 `BARNESS_AI_LIVE_PERMISSIONS_FILE`，绑定来源/二进制/目录/身份/别名/完整计划/有限预算；common Git 排他 claim 在任何请求前消费，失败和取消也不能重用。没有扩大环境转发白名单，controlled 开关不进入 paid 合同。
- [Host 装配](../../../docs/barness-ai/xai.md) 使用已有公共 Catalog/resolver，包含白名单、显式凭据、资源/超时、持久化、工具、费用和禁用/版本回滚；外部 Host 不导入 internal testkit。`docs/barness-ai/README.md` 提供 candidate/NOT_RUN 发现入口。
- 报告始终为 candidate，controlled 的 vendorHTTPAttempts=0，不写正式 support report 或支持矩阵，不激活内置目录。正式消费者已用合法仓库内完整候选包、全新输出目录实际执行，并保存绑定该候选报告哈希的拒绝回执、exit 1。
- 最终报告先完整写入并同步 pending，再替换原非通过报告；普通子进程真实制造 pending 路径 EISDIR 写入失败，测试 PASS 标签不改变该进程 exit 1 与原 RUNNING/complete=false 状态。

冻结普通命令：

```sh
python3 tools/verification/run.py --feature barness-ai-grok --stage e2e -- \
  go test -json ./ai/e2e -run '^TestXAILiveHarnessProcess$' -count=1 -timeout=3m
```

实际受测提交 `c42067f39f457d70742985eb94db6d204b67b2ff`；[不可变版本索引](../evidence/ticket07-controlled-process-v1.json) 记录完整 28 个主 case ID、argv、数字退出、manifest/audit 哈希和原件位置。主包 28/28 PASS、audit 0；完整受控计划 18 场景、24 次合成 HTTP、0 次厂商 HTTP。所有 red 与失败原件另留 attempt。

原件保留于工作树 `/tmp/barness-grok-live-scaffold` 的 `.evidence/barness-ai-grok/e2e/20261026T023907.909775Z-mipzyz5h/`；路径均相对此保留工作树，未复制原生胶囊到 scratch。收尾只修改 Markdown/该索引，完整执行来源/选定输入相等性另留 `.evidence/barness-ai-grok/ticket07-source-equivalence/v1.json`。

本票没有执行完整受影响普通包，统一集成验收由04与主工作树处理。paid live、race、压力均 NOT_RUN；未取得独立授权。地区/Kimi/GLM 原生来源库存因新增源码而变化，当前资格不能继承历史 PASS，由05/06和最终影响审查处理；release.pass=false。回滚为禁用/更新 Binding，保留历史、目录、授权/claim 和证据。

## Comments

- 2026-10-26：从04拆出独立准备工作，允许01后与02/03并行；05仍负责真实资格。

- 2026-10-26：受控公共子进程验收完成；来源 c42067f3，28 主 case 全部 PASS，完整候选计划 24 合成调用/0 vendor；实际正式拒绝及最终写入失败证据已保留。
