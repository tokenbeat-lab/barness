# 03: 完成 GLM Chat 工具流与交错思考往返

Status: resolved
Type: task

由 2026-10-17 implement-spec 实施指令发布。

**What to build:** Host 收到完整且可逐步观察的 GLM 函数调用，执行工具后将同 turn 推理和工具结果回传，模型继续回答；未接入的工具选择和保留式思考请求得到明确错误。

Blocked by: 02

依据：GLM 接入 spec R07、R11、R12、R15、R16；验收组 A02、A05、A07、A08、A13；User Stories 26、27、28、29、30、40、41、42。

- [x] 先写两型号工具参数分片、多工具 index/ID、工具结果回传和最终回答的公共 E2E，再实现相应行为；Host 执行工具，库每次只生成一个轮次。
- [x] 带函数工具的流式请求发送 tool_stream=true，依正确 index/ID 组合参数；完整参数、ToolCall 和结束事件保持一致，不提前执行未完整工具。
- [x] 合法工具往返完整保留必要的 reasoning_content、调用 ID 和工具结果顺序；不继承 DeepSeek 给每条 assistant 补空推理字段的规则。
- [x] 仅默认/auto 工具选择可用；full 的 required/具名强制和 simple 的 none 在发送前拒绝，SimpleOptions 不增加 required。
- [x] 标准端点保持 clear_thinking=true 的跨 turn 行为；显式未接入的 clear_thinking=false 拒绝，不以裁剪或重排后的历史假称 Preserved Thinking。
- [x] 对具名选项、SamplingParams、OnPayload 的最终 wire 统一检查工具选择与型号思考硬约束；保持合法 hook 顺序和授权错误优先级。
- [x] 跨型号/不可信历史沿既有降级规则转换，保留可见内容与有效工具配对；持久化原生状态仍须 Host 重新担保。
- [x] 产出两型号至少一个完整工具往返及分片/错误用例的可重放请求、事件、Result、Attempt 与断言；受影响 DeepSeek Chat 回归通过。

**独立验收：** 两型号各完成一次思考→函数调用→Host 工具结果→继续回答，同时演示非法强制工具和未接入保留式思考均无网络请求。

**依赖理由：** 依赖 02 的 zhipu Chat 请求、思考编码与公共候选调用入口。


## Answer

2026-10-17：两型号的公共 Client 工具往返、参数分片与 Host 执行、同 turn reasoning_content 回放、默认/auto 工具选择边界、clear_thinking:false 拒绝和最终请求检查完成。具名校验在凭据读取前，采样/回调之后按实际 wire 重新校验，授权错误优先，非法请求零网络与零 Attempt。持久化信任、跨身份降级和不可变快照沿既有公共契约验证。

实现提交 `2769e46`；合并最新集成 `9ea22d0` 后的已提交重放源 `4975214`：完整 GLM Chat 358 PASS，本票新增 192 例 race PASS；受影响回归 1999 PASS 加单独 oracle guard 1 PASS；冻结 OpenAI Chat/DeepSeek Chat 321 PASS，零 pending；vet/diff checks 通过。附带修复 GLM 最终预算微小小数被 float64 舍入的问题，精度 red/green 覆盖 model/profile/call/hook，旧 Provider 语义不变。

上下文：[ADR-0033](../../../docs/adr/0033-barness-ai-zhipu-chat-candidate.md)、[失败方式与逐包证据索引](../evidence/ticket-03/README.md)。全部为合成离线公共接缝证据，未读取 `.env` 或运行收费调用；真实标准 API 验收、正式目录与支持矩阵仍由后续票负责。
