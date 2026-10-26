# 规范化 Grok 选项证据标签

Type: task
Status: claimed
Blocked by: 08

## Contract

[Spec](../spec.md) R23、E04/E18；[SOP](../../../docs/agents/verification.md)。当前状态仍只由主工作树在 spec 维护。

## Diagnosis

主工作树真实只读发布消费者拒绝 `record-v1.json`：source=true/e2e=false，问题为 native offline bundle invalid。旧冻结 ordinary receipt `20261026T031723.005911Z-syyrwbay` 实际 PASS，但 `TestXAIReasoningAndBudget` 将语义空 effort 直接放入中间标签，生产 `...-full//0` 等非规范目录。原生 importer 依法拒绝；旧原件与原 record 不修改。

沿用已有实际 CLI 失败与已完成的最小路径诊断，不重复宽范围假设调查。复现闭环为新公开 E2E 子进程：真实执行该选项测试，经现有 collector 收集，再调用实际编译的 releasegate；精确断言 source/e2e=true 且未执行 live/race/压力继续阻止 release。

## Failure inventory

- 空档位位于中间路径段，物理归一化掩盖 manifest 原始非规范目录。
- 只断言总 CLI exit1 会误将 e2e 拒绝与正常待验收混淆。
- 标签修复不能改变实际 wire effort 省略语义。

## Evidence

全部原件仍在 `.evidence/barness-ai-grok/e2e/<attempt>/`，未改旧 attempt、旧 record 或原 manifest。

- Red：`20261026T035321.936184Z-6z9pqyan`，外层FAIL/exit1。子选项 producer 实际 PASS；真实编译的原生 releasegate 为 source=true/e2e=false，问题原文为 `native offline bundle invalid`。live/race/pressure继续false，CLI exit1。
- 首次 green：`20261026T035401.944696Z-6_7gbix_`，外层PASS/exit0；原生消费者 source/e2e均true，live/race/pressure均false，CLI仍exit1。该回执为迭代来源，最终提交后另跑定向冻结范围。

## Change

只将选项 producer 标签中的语义空 effort 显示为 `default`，实际请求仍用原空值并断言 reasoning 省略。原生 importer及生产 xAI 实现均未改。

增加公开回归 `TestXAIOptionEvidenceConsumption` / case `XAI-E04-option-evidence-consumer`。它调用实际选项测试、现有 collector 和编译后的原生 releasegate，不使用合成 offline 包冒充真实 producer。有限必交case为full/default/0与simple/default/128000；同时检查source/e2e通过以及全部未执行阶段继续拒绝。新增case和十个 `/default/` 标签重命名须由集成维护覆盖清单。

