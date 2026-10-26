# 已声明不支持的原生能力资格

Type: task
Status: claimed
Blocked by: 16

## Contract

[Spec](../spec.md) R23/E18；[SOP](../../../docs/agents/verification.md)；[旧目标原生合同](../../../docs/barness-ai/legacy-history-qualification.md)。保留完整场景集合、已声明能力语义与所有原始状态；不得把UNSUPPORTED改成PASS，不得将任意未支持、失败或未运行放行为正式资格。xai-origin-history始终必须真实成功。

## Failure

v3正式批次第4条OpenAI Chat：10场景PASS，reasoning-history依ai/live/combos_test.go既有noReplay合同为UNSUPPORTED、0call/0attempt。Go原生执行PASS/15HTTP，正式消费者exit1/pass=false；此前3条正式资格通过，后13条NOT_RUN，合计65calls/65HTTP。源码ai/internal/testkit/release/chat_qualification.go统一要求每scene正calls/attempts，record_live.go统一要求PASS，遗漏既有无原生推理内容可回放的能力声明。

## Ownership

同一implementer负责最小原生合同/正式消费者修复、公开CLI回放回归及必要合同说明；root负责真实停止批次与唯一当前状态。不得改旧真实原件、续用旧grant或发起厂商调用。

## Verification

先通过公共CLI建立完整OpenAI Chat native fixture的RED（保留reasoning-history为真实UNSUPPORTED、零请求/零捕获，完整其余场景且预算一致），再实现。只允许原生既有声明明确界定的不支持状态；审计Generic六条的所有unsupported分支，对照当前固定目录/型号，避免再遗漏。反例覆盖伪造状态/原因、型号/身份、请求/attempt/重试/HTTP/捕获与预算伪造，其他route/scene与mandatory xAI来源历史不得借例外放行。验证支持矩阵保留该capability为UNSUPPORTED/无last-passed，不赋予回放能力。测试先列失败方式；优先复用现有公开E2E/CLI及固定scope source helpers，不建新框架或扩大真实权限。源码变更重新冻结，后续真实调用另取授权。

## Answer

最小生产修复268faeba、完整publicCLI回归b77859c0及独立v12两轴0/0完成；等待本来源完整四包普通验收后关闭软件票。旧批次停止不续跑，后续新source付费仍另取选择。

- v11各1 P2已通过完整原生包公共CLI接线补齐：原strict规则、报告接线、原件接线三份实际RED，61项定向GREEN；[完整证据](../evidence/generic-unsupported-cli-v1.json)/[v12](../review/review-v12.md)。当前完整普通由集成统一执行，不把合成材料计为正式资格。
