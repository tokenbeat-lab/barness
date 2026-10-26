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
