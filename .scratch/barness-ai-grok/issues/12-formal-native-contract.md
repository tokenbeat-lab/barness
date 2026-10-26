# xAI 自身正式资格合同与冻结版本

Type: task
Status: ready-for-agent
Blocked by: 11

## Contract

[spec](../spec.md) R21/R24；[SOP](../../../docs/agents/verification.md)；[冻结顺序选择](../formal-qualification-plan.md)；[代码路径研究](../post-candidate-qualification-notes.md)。本工单承接05的软件部分，不能把现有candidate改写为formal。

## Boundary to resolve

维护者已明确选择推荐的未发布分支/正式发布边界；公共回放已合入。现在一次完成最终目录、原生自身合同和正式harness后冻结；保留release.pass=false，正式live与人工阶段各自另行授权。

## Ownership and flow

- supportmatrix/xai专属报告/固定18能力、身份、24调用预算和独立捕获检查；不注册为regional，不放宽generic输出预算假设。
- 现有 `releasegate -live-bundle` 接入xAI正式原件；保持candidate/controlled/replay拒绝正式资格，不增加只为让旧candidate通过的新CLI开关。
- live xAI formal入口以全Builtin版本/hash、新source/binary、正式stage和一次性grant/claim执行；复用场景及失败收尾，产出自身schema和live-report.json。
- runtime最终目录事实不能从internal/testkit导入；如需独立候选与正式快照，注明其版本/资格独立演进原因。
- 新原生消费在实际paid HEAD捕获；后续只读Record可按完整文件/输入相等沿用已捕获schema2，不能在qualifiedLive内新加当前HEAD比较破坏文档收尾。

## Failure scenarios before implementation

公共CLI/进程先保留失败，再实现；已知当前真实candidate因缺live-report.json不能正式消费。必须覆盖：candidate/controlled拒绝、缺失/重复/替换场景、错误stage/身份/SDK/目录/source/binary/grant、被改/截断/缺失capture-result-observer三元组、usage/observer不一致、late写入/同步/audit失败、历史source或commit不一致。受控/合成正例不产生实际live资格。

## Remaining authorization

代码审查/完整普通/冻结与构建完成后，才为最终正式版本提出具体新grant；另外22路线保持各自未授权，race/压力保持未选择。禁止再次运行已消耗的v3候选grant。

## Answer

边界已由维护者确认，待实施。当前候选真实PASS事实见[索引](../evidence/candidate-live-v3.json)。

## Comments

- 2026-10-26：由完整真实候选和原生消费者代码研究细化；没有执行新的厂商请求、race或压力。
