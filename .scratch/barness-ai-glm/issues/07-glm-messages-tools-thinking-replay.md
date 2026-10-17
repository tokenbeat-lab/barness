# 07: 完成 GLM Messages 工具、思考与原生续接

Status: resolved
Type: task

由 2026-10-17 implement-spec 实施指令发布。

**What to build:** Host 接收 GLM Messages 的函数调用和可用思考块，执行工具后继续生成；签名、无签名 thinking 和工具结果都按 GLM 真实规则往返。

Blocked by: 06

依据：GLM 接入 spec R07、R13、R14、R15、R16；验收组 A02、A06、A07、A13；User Stories 26、27、28、30、31、32、33、40、41、56。

- [x] 用 06 的真实契约确定探针及 fixture，先写 thinking/工具/签名失败场景，再实现具体适配；所有正式验收经公共 Client。
- [x] 两型号完成 tool_use→Host 执行→tool_result→最终回答，工具 JSON 分片完整，ID/顺序与合法 SSE 终态正确。
- [x] 可用 thinking、redacted_thinking、signature_delta、有/无 signature 历史按 GLM 规则保留；不伪造 Claude 签名，不将无签名 thinking 自动转 text 后称为原生续接。
- [x] 思考/effort/budget 的映射只采用本型号/协议已证实的规则；不能把预算和强度猜换；非法显式选项及最终 hook 载荷在发送前拒绝。
- [x] 默认和 auto 工具选择至少可用，其他形式只有自身证据、guard 和公共 E2E 同时支持才声明；不得把另一协议的选择限制直接迁入。
- [x] 同模型、进程内可信结果与持久化经 Host 担保的状态可续接；不可信、跨账户/API/型号状态按现有规则降级而不泄漏账户引用。
- [x] 若上游明确不返回可观察思考块，可按 spec 附来源及真实证据登记该可观察能力 UNSUPPORTED；工具往返和合法继续生成不可豁免。
- [x] 产出两型号工具往返、思考/签名回放与非法字段拒绝证据；未解决必交能力保留 FAIL/未完成，不用纯文本成功替代整组合验收。
- [x] 记录 GLM Messages 特有的原生回放和能力决定，同步 ADR；现有原生 Anthropic 工具、thinking 和缓存/beta 场景无回归。

**独立验收：** 两型号均完成 Messages 工具往返；每种实际返回的 thinking/签名有合法续接证据，不能回放的来源有明确降级结果。

**依赖理由：** 需要 06 的 GLM Messages 基本调用与经证实的 wire 契约。


## Answer

在 `codex/glm-07` 实现（`c00012c`、`5606fda`），已于 `85b6fa9` 合并集成分支
`f0fb0de`。两型号的独立 Messages 原始工具/签名 SSE 与请求从已审计的 03
探索产物原样固化，04 提供仅省略 signature 的真实接受证据；公共四入口覆盖 default/auto → ValidateToolCall → Host 固定工具执行
→ ToolResultText → 最终回答，JSON 分片、ID、内容顺序与合法终态保持。

GLM-5.3 省略思考仍保留 Provider 默认；5.2 仅支持已证实的 full enabled、budget1024、
max4096 精确形状。新增可省略的语义 ModelCompat.MessagesThinkingBudgetTokens（0 不启用），
Provider 只实现该固定子集，型号元数据不能扩充预算；canonical、effort、adaptive 和其他
预算/输出组合仍拒绝，simple 不再通过 Claude 分档预算算术猜换到这个 full 形状。

可信同来源 signed/unsigned thinking 可直接或持久化经 Host TrustNativeState 续接；unsigned
保留 thinking 类型和正文，仅省略 signature 字段。未担保、跨租户/账户/API/型号均按已有规则
降级并记录原因。没有伪造签名，也没有将 unsigned 改 text 后宣称 native。未观察到真实
redacted thinking 或 unsigned Provider 输出：redacted 原生回放、空签名和未接入工具扩展
在发送前明确拒绝；不可信 redacted 状态降级时丢弃。存储完整性继续由 Host 担保。

GLM 独立 SSE 顺序检查拒绝 18 种两型号故障脚本；多索引交错参数独立组装。最终
callback 的模型/原生引用授权先于能力校验，未接入选项和文本型号图片注入均零请求/零
Attempt，schema/input 用户数据不误判为协议字段。原生 Anthropic 既有语义保留。

最终提交源码公共 Client **193 PASS**，race **190 + 3 PASS**，原生 Anthropic 回归
**104 PASS**、运行时 **168 PASS**、锁定 pi Messages 差分 **70 PASS / 0 pending**；
vet、whitespace 与全部产物清单/脱敏审计通过。红测、harness 迭代、ZIP 与哈希见
[证据索引及复现命令](../evidence/ticket-07/README.md)，决策见
[ADR-0035](../../../docs/adr/0035-barness-ai-zhipu-messages-basic.md)。

本票零付费调用、未读取 `.env`。既有 wire 探索及离线生产实现回放不能替代 ticket 11
正式 production Client live，ticket 12 内置目录与发布状态仍独立验收。
