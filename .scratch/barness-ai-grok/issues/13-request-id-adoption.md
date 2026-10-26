# xAI 已确认响应头的厂商请求 ID

Type: task
Status: claimed
Blocked by: 10

## Contract

[spec](../spec.md) R20；[SOP](../../../docs/agents/verification.md)。自身候选真实v3的24份响应均包含`x-request-id`，而当前公开Attempt因旧保守映射仍省略它。本次仅按现有Provider映射采纳已证实的头名，不改变错误正文脱敏或其他Provider选择。

## Ownership

ai/openai_sdk.go的request-ID Provider映射；新增公共E2E与被该明确行为变化直接取代的既有xAI请求ID断言。

## Failure modes before implementation

已约定的公共Client/受控HTTP接缝：

- Stream、StreamSimple、Complete、CompleteSimple成功时Result与终态Observer保留响应头ID。
- HTTP401、429或HTTP200后的协议失败，Attempt/Error/终态Observer保留同一已收到头ID。
- 该头缺失时，不从客户端RequestID、response body ID、Request-Id、x-ds-trace-id或traceparent捏造ID。
- OpenAI继续使用x-request-id，DeepSeek继续优先x-ds-trace-id。
- 既有HTTP/流失败分类、重试、用量及正文/密文排除断言保持。

## Verification

新公共E2E先RED，修改最小Provider映射后GREEN，并执行受影响xAI普通范围；不执行厂商请求、race、压力或最终完整冻结。原生来源会改变，历史v3仍只记录原始候选事实。

## Source evidence

`.evidence/barness-ai-grok/candidate-usage-observations-v3.json`在原仓库保留24个`x-request-id`头名的安全计数；原始bundle为`.evidence/barness-ai-grok/live/xai-responses/20261026T050450.562258Z-tlbttr1o/bundles/20261026T050451.200594000Z`。本工单不复制原始ID或密文到源码。

## Answer

待RED/GREEN及复核。

## Comments

- 2026-10-26：自身真实证据已经补足R20采纳响应头的前提；不以兼容API的其他Provider证据替代。
