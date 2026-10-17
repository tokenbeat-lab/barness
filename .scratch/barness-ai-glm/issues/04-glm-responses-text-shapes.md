# 04: 接通 GLM Responses 并修复对象型内容丢失

Status: resolved
Type: task

由 2026-10-17 implement-spec 实施指令发布。

**What to build:** Host 使用独立的 GLM Responses Binding 获得完整文本和思考；GLM 的单对象或数组 content、只有 done 的结果与无 [DONE] 流都被正确呈现。

Blocked by: 02

依据：GLM 接入 spec R02、R04、R05、R06、R07、R08、R09、R10；验收组 A01、A02、A03、A08；User Stories 2、5、6、7、8、20、21、22、23、24、25、48、51、52、53。

- [x] 先通过公共 Client 重现单对象 reasoning/message content 被现有 SDK 静默丢失的失败，再实现 raw JSON 结构转换；不做字符串替换，不依赖 Unmarshal 无错误认定内容完整。
- [x] 两型号通过独立 Responses 候选条目和 Binding 使用正确官方基址、资源路径与 Bearer；复用现有 Responses adapter 和公共四种调用入口。
- [x] reasoning.content 单对象、message.content 单对象/数组、只有 done、delta 后 done 校正及交错 item index 都不丢失、重复或错误清空内容；畸形结构明确失败。
- [x] GLM 原始 reasoning 条目与解析用规范形状分开保留；为下一票历史回放提供未污染的原生状态，不存入伪造 summary/encrypted 或 OpenAI synthetic ID。
- [x] 默认 store=false，既不依赖服务端存储也不发送 previous_response_id/conversation；effort 不自动带 summary=auto、encrypted include、tier、缓存保留期或亲和头。
- [x] 5.3 的 Responses 输入先限定 low/high/max；不从 Chat 假定服务端强度，记录文档映射和需探针确认的事实；simple/full 与采样遵守本协议目录。
- [x] max_output_tokens 接受 GLM 合法的 1 起范围，不隐藏抬到 OpenAI 的 16；越界明确拒绝，最终 sampling/hook 不可绕过。
- [x] 无需 [DONE] 也按合法协议终态完成；缺终态 EOF、failed/error/incomplete 按既有协议规则表现，事件与 Result 一致。
- [x] 留下两型号、所有关键 content 形状与终态的可重放产物；已有 OpenAI/DeepSeek Responses 的受影响请求和结果回归通过。

**独立验收：** 分别演示单对象、数组、仅 done 和 delta 后 done 的完整文本/思考结果，以及不含 [DONE] 的合法 GLM Responses 流。

**依赖理由：** 仅依赖 02 提供的共享 Provider/候选 Host 接入；不等待 03 的 Chat 工具功能。


## Answer

已在 `codex/glm-04` 实现独立 Responses 候选目录/Binding、官方路径和 Bearer、全部公共入口以及 nil/full/WithHooks。公共 Client 首次红灯明确复现官方 schema 对象 content 成功解码却文本/思考为空；在 raw JSON 边界按结构校验、为 SDK 转换解析副本，并另存未转换的原生 reasoning。真实基础样本均为数组，已按各自型号离线重放；没有把对象合成 fixture 称为 live 复现。

GLM 处理 authoritative done、仅 done/仅终态、交错 output_index、delta 校正、合法无 [DONE] 终态和畸形/缺终态错误；不向原生状态注入 summary/encrypted/synthetic ID。默认 store=false、effort 省略保留服务默认；5.3 full 只接受 low/high/max，simple 夹取；共享精确整数校验与本协议原始预算保留确保 1..min(型号上限,131072) 通过各采样层/最终回调检查，不受 OpenAI16下限影响。缓存路由能力与保留期/亲和头独立。

[ADR-0034](../../../docs/adr/0034-barness-ai-zhipu-responses-boundary.md)、[完整重放索引](../evidence/ticket-04/README.md) 记录来源冲突、保守上下文/精确输出及合成价格边界。受影响公共回归1978 PASS，冻结 Responses 差分335 PASS/零 pending；单独启用 oracle 后的合并回归无 NOT_RUN；go vet 通过。所有已完成证据包脱敏审计通过；本票未读取 GLM 凭据、未发收费请求，正式支持矩阵与内置目录未改。工具/历史请求转换和 enhanced guard 由05消费本票保留的原生签名继续实现。
