# 真实非空消息捕获的正式消费者解码

Type: task
Status: claimed
Blocked by: 15

## Contract

[Spec](../spec.md) R23/E18；[验证 SOP](../../../docs/agents/verification.md)。只修复正式证据消费边界，不改变公共Result/Message JSON合同，不放宽归属、预算、usage、库存、audit或正式资格规则，不新增付费调用。

## Ownership

release通用chat捕获DTO与公共CLI回放回归。单一implementer在独立工作树实现；root负责真实批次停止原件与状态。

## Failure

已授权正式批次gf-20261026T112558Z-14ec1226：xAI资格通过；OpenAI Responses9/9场景与15调用通过，但消费者exit1/pass=false，后15条NOT_RUN。实际原件：[receipt](../../../.evidence/barness-ai-grok/live/openai-responses/20261026T114333.867655Z-4st8edcu/receipt.json)。

chat_capture.go直接json.Unmarshal到ai.Result；Message.Content是非空AssistantContent接口切片，实际15结果均非空，读取失败，逐级折为native evidence audit failed。独立read-only诊断确认meta/Observer/target/model/usage/库存均一致；[诊断只读报告](../../../.evidence/barness-ai-grok/live-batches/gf-20261026T112558Z-14ec1226/02-openai-responses/diagnostic-record/native/release-report.json)不代表缩小正式范围或接受。

## Verification

先用真实public producer/collector/compiled CLI边界建立包含非空text/thinking/tool内容的RED，修复后GREEN，并保留typed捕获所需身份/stop/error/Usage/CallMetadata约束的反例。仅解析消费者实际使用的独立DTO，公共ai.Result不加反序列化。验证适用普通范围并独立两轴复核；来源变化重新冻结，旧付费原件不改pin/不覆盖/不自动复用grant。

## Answer

软件修复及12项公开CLI回归完成，v10独立两轴0/0；当前等待完整受影响普通实测后关闭软件票。当前来源正式live未完成，release.pass=false。

## Comments

- 2026-10-26：实际原生消费者拒绝触发；批次已停止、39调用/39HTTP原件全部保留，后15条未执行。修复为已授权implement-spec软件范围，付费重跑和后续人工阶段另取选择。

- 原缺陷公开RED/GREEN分别115842/115932；v9 Spec P2补充历史分支后121309实际12/12 PASS，生产DTO未再改变，新反例直接GREEN。见[v2证据](../evidence/generic-capture-decoder-v2.json)与[v10复核](../review/review-v10.md)。完整普通由集成分支统一执行，不启动付费或人工阶段。
