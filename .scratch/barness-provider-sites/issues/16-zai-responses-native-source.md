# 16 ZAI Responses 预先固定自身文本推理来源

Status: claimed
Type: task

## Context

[07最终验收](07-final-acceptance.md)、[12同类Chat来源修复](12-zai-chat-native-source.md)、[规格R06/R17/R18/R20](../spec.md)。本票修正验收harness及其原生来源合同，不把原真实失败改写为成功。

## Actual failure

固定恢复3在实际6b87a584来源完成M3三路39/39与ZAI Chat18/18；ZAI Responses完整18场景17 PASS/1 FAIL，`glm-5.3/native-history`断言`native=false`。全部走显式代理7890、零重试，失败进程实际18 Calls/18 HTTP；父UNKNOWN和US$0.65全额占用不改。[停止索引](../../../.evidence/provider-sites/final-acceptance-20261020/recovery3-stopped/index.json)固定1012原件，跨轮实际378/374、claimed564/612、整笔12.60USD；后19进程未启动。

失败断言检查预先固定的tool-call源消息。该实际源SSE没有reasoning item/事件，保存Thinking0且reasoning_tokens=0；同进程full-stream自身保存完整推理/原生状态。旧成功tool-call原件由当前公共Client正确保留；新旧最终tool-call请求逐字相等。没有parser回归或代理丢帧的证据。

## Required fix

沿用12已验证的设计，将ZAI Responses的原生推理源也预先固定为同轮同型号full-stream，保持独立tool-call/tool-result链与Host标记；不得事后挑成功消息、从兄弟协议借来源、伪造/拼接原生状态、降低推理/零downgrade断言或将真实FAIL推广。harness记录与`supportmatrix.RegionalNativeSourceCase`必须同步，严格消费者仍核对原件和精确续接前缀。

先公共CLI/Client受控失败覆盖“工具响应合法但无推理，先前文本流有自身完整推理”，再实现并跑受影响完整六路线、原生消费者及拒绝场景。生产协议/请求/SDK/模型目录不改。supportmatrix消费者进入MiniMax执行来源，因此三M3与六formal需新的自身资格；GLM245逐项相等才可保留。不能缩小任何来源选择器以绕过重验。

## Paid boundary

固定恢复3已按实际批准的失败即停条件停止。本票的离线实现、普通测试和审查不授予新付费调用。新的完整有限范围、全部历史/UNKNOWN/预留、预算分配、源码/二进制/工具及双轴审查必须成为具体计划，再取得明确批准；预算上限仍US$30。无新READY，race仍NOT_RUN，完整发布继续开放。
