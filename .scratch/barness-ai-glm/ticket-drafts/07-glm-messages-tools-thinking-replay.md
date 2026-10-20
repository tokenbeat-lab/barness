# 07: 完成 GLM Messages 工具、思考与原生续接

拆分草案，尚未获确认或发布；确认后才迁入正式工单并设置 ready-for-agent。

**What to build:** Host 接收 GLM Messages 的函数调用和可用思考块，执行工具后继续生成；签名、无签名 thinking 和工具结果都按 GLM 真实规则往返。

**Blocked by:** 06：建立 GLM Messages 实际契约并完成文本调用

依据：GLM 接入 spec R07、R13、R14、R15、R16；验收组 A02、A06、A07、A13；User Stories 26、27、28、30、31、32、33、40、41、56。

- [ ] 用 06 的真实契约确定探针及 fixture，先写 thinking/工具/签名失败场景，再实现具体适配；所有正式验收经公共 Client。
- [ ] 两型号完成 tool_use→Host 执行→tool_result→最终回答，工具 JSON 分片完整，ID/顺序与合法 SSE 终态正确。
- [ ] 可用 thinking、redacted_thinking、signature_delta、有/无 signature 历史按 GLM 规则保留；不伪造 Claude 签名，不将无签名 thinking 自动转 text 后称为原生续接。
- [ ] 思考/effort/budget 的映射只采用本型号/协议已证实的规则；不能把预算和强度猜换；非法显式选项及最终 hook 载荷在发送前拒绝。
- [ ] 默认和 auto 工具选择至少可用，其他形式只有自身证据、guard 和公共 E2E 同时支持才声明；不得把另一协议的选择限制直接迁入。
- [ ] 同模型、进程内可信结果与持久化经 Host 担保的状态可续接；不可信、跨账户/API/型号状态按现有规则降级而不泄漏账户引用。
- [ ] 若上游明确不返回可观察思考块，可按 spec 附来源及真实证据登记该可观察能力 UNSUPPORTED；工具往返和合法继续生成不可豁免。
- [ ] 产出两型号工具往返、思考/签名回放与非法字段拒绝证据；未解决必交能力保留 FAIL/未完成，不用纯文本成功替代整组合验收。
- [ ] 记录 GLM Messages 特有的原生回放和能力决定，同步 ADR；现有原生 Anthropic 工具、thinking 和缓存/beta 场景无回归。

**独立验收：** 两型号均完成 Messages 工具往返；每种实际返回的 thinking/签名有合法续接证据，不能回放的来源有明确降级结果。

**依赖理由：** 需要 06 的 GLM Messages 基本调用与经证实的 wire 契约。

