# 正式资格软件两轴审查修复

Type: task
Status: claimed
Blocked by: 13

## Contract

[审查v5](../review/review-v5.md)、[Standards](../review/standards-v5.md)、[Spec](../review/spec-v5.md)；[SOP](../../../docs/agents/verification.md)。同一实现者处理全部发现，公共CLI先失败后修复，保留原件。12软件首次交付已合入，工单整体仍等集成/冻结；不把此修复反向阻塞为环。

## Ownership

ai/internal/testkit/supportmatrix/xai相关独立原生验证、必要公开CLI场景、ai/release/catalog-snapshot.json当前快照及测试。不得更改原v3真实捕获或消耗新的付费授权。

## Answer

待统一修复/复核；两轴发现分别保留，最终完整普通另行冻结。

## v6 增量修复

v5修复已合入 `51eb2a24`，定向普通1010测试/884case/audit0，完整原件见 `.evidence/barness-ai-grok/formal-review-fixes-v1.json`。对合并旧目标软件的 `9ea3b9d3` 独立[v6审查](../review/review-v6.md)新增Standards2项/Spec2项P2；仍由原同一实现者统一修复，扩大15所有权至 `livegrant` 文件读取/claim、`grokhistory` 调用内容证明及对应公共E2E/CLI。14实现者源码已稳定，不再共同修改这些文件。

先保留公开RED及原件，再做最小修复并运行受影响普通范围。历史grant/claim、closed ledger、原始candidate和旧价格/评测pins保持原件事实；禁止借修复重买真实调用。冻结完整普通等待四项修复及独立复核。

## 集成软件验收完成

2026-10-26：源码已统一合入 `f89e59a6`，v7独立两轴复核0/0；冻结后的完整受影响普通实际受测 `97697bb2c2937974beb362aefc7a41bfaad3ee15`，16,184个Go测试PASS、1456/1456必交case与119/119入口PASS、audit0、exit0，墙钟1283.188秒。原15m/10m历史矩阵失败/超时与所有RED原件保留；本次完整minimaxlive包实际PASS。

[独立完整核验](../../../.evidence/barness-ai-grok/final-ordinary-validation-v5.json)、[最终受测回执](../../../.evidence/barness-ai-grok/e2e/20261026T093152.246258Z-ua2zgcqn/receipt.json)、[复核](../review/review-v7.md)、[正式计划](../formal-live-plan-v1.md)。本工单仅关闭软件实现；05正式live、06人工阶段/完整发布尚未完成或授权，不构成正式支持。

## 原生离线消费实际收尾失败

完整Go执行97697bb2通过，但随后编译后releasegate实际拒绝主离线bundle，source=true/e2e=false。唯一已确定库存违约为新增`GROK-HISTORY-GLM-CONTROLLED-FULL`4350文件，超过既有每case512限制，原始[诊断](../../../.evidence/barness-ai-grok/offline-inventory-diagnosis-v1.json)与[实际报告](../../../.evidence/barness-ai-grok/release/readonly-v5/native/release-report.json)保留。重新打开15，仍由同一实现者修复该测试生成/归属边界；禁止放宽导入上限、删除原件或修改旧pin。修复后需公开producer→真实CLI消费GREEN和影响范围适用复验，再更新当前状态与新索引。
