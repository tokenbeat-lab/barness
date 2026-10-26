# xAI 已确认响应头的厂商请求 ID

Type: task
Status: resolved
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

`511fa00c`移除xAI对已确认响应头的保守抑制，沿现有映射采用`x-request-id`。新增两顶层公共测试/34个case；16个有头场景先RED，缺头及两个旧Provider场景保持通过。修复后34/34通过，合入最新回放的全部xAI普通范围568 case/609 Go测试/66入口PASS，audit0、exit0。[独立证据指针](../evidence/request-id-adoption-v1.json)。既有14个HTTP重试、26个流故障、7个HTTP用量失败case ID均未变，错误分类、重试及正文排除保持。

本工单实现验证完成；最终两轴复核、完整包冻结及正式新来源资格由集成阶段继续，不据此宣称发布通过。

## Comments

- 2026-10-26：自身真实证据已经补足R20采纳响应头的前提；不以兼容API的其他Provider证据替代。

- 2026-10-26：完整RED/GREEN原件、source patch及精确新case清单保留在 `.evidence/barness-ai-grok/request-id-adoption-v1.json` 指向的独立attempt；没有厂商、race或压力调用。
