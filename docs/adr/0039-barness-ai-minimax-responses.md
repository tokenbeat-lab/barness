---
status: accepted
date: 2026-10-18
---

# MiniMax Responses：独立能力、权威输出与无损候选历史投影

地区/Provider、来源和价格的当前适用性由[ADR-0042](0042-barness-ai-minimax-international-target.md)取代；下文保留初始中国区决策及历史证据上下文。既有协议约束、可信历史、正写入缺价保护和晋升边界继续有效，国际站须取得自身证据。

## Context

[MiniMax spec](../../.scratch/barness-ai-minimax/spec.md) R07–R08、R12–R14 与
R17–R21 要求复用公共 Client/Responses 生命周期，但不能沿用 OpenAI 的存储、加密推理、
16 token 最小预算、工具强制或 item ID 规则。依据为 2026-10-18 官方
[Responses schema](https://platform.minimax.cn/docs/api-reference/responses-create)；
具体快照由[研究](../../.scratch/barness-ai-minimax/research.md)记录。

输出 reasoning.content 是 reasoning_text 数组；输入只声明 summary_text 数组。
SDK 能读取输出并不证明该输出可以原样回放。此处只决定有明确来源的候选转换；
实际 M3 工具结果后续接仍须由工单 10 的独立真实证据确认。

## Decision

MiniMax 使用独立 Provider 能力，保留完整 input。默认不派生 store、reasoning、include、
parallel_tool_calls、truncation 或缓存期限/亲和头。显式 reasoning 按型号能力校验；
M3 high 仅为开启，M3.1 low/medium/high/xhigh/max 且不可关闭。标准服务层级可显式表达，
priority 拒绝，响应不会启用 OpenAI 的 tier 价格倍率。最终 sampling/可信回调之后再次验证请求子集、授权、精确预算及文本边界。

SSE 先验证原始数组形状与关联，再交 SDK 解码；不改写 transport 或把 reasoning 对象
修成数组。done/终态输出具有权威性，终态独自携带输出也能生成完整结果。重复、错误关联、
终态后内容与无终态 EOF 明确失败。只有 thinking 的 max_output_tokens 截断保留 length。
failed 的 ResponseID/usage 仍遵守既有 Responses 合同；实际 ResponseModel 独立记录。

完整原始 reasoning item 存入可信 Native State 的 Thinking.Signature。回放时将原始
content 的每个 reasoning_text 文本原封不动映为输入 summary_text；原始 summary 文本
也保留为独立片段，不进行语义摘要。原条目无文本而已合法收到完整 reasoning delta/done
时，保留其已收文本。没有原生 item 的 done-only 推理按可见 assistant 文本续接，不伪造
原生签名。未知原生字段或非文本形状明确拒绝回放，而不是静默剥除。

公开工具 ID 使用 `mmr1|` 加 base64url JSON 二元组，完整保存 call_id 与 item ID。
输入按真实 call_id 配对，条目 ID 不进入官方未声明的输入字段；不存在 fc_ 假设、净化或
截断。跨来源历史先包装整个旧 ID；跨账户/型号或不可信 Native State 按共享规则降级，
调用与结果同步保留关联。

## Considered Options

- 原样回送输出 reasoning：官方输入未声明 content，不能将 SDK 解码当作服务接受证据。
- 丢掉推理或生成摘要：丢失完整历史，违背 spec。
- 按官方两侧 schema 逐片段无损投影（采用作为候选），保留原始 Native State 并要求独立 live。
- 依赖 previous_response_id 或服务端保存：违反无状态完整历史与账户隔离合同。

## Consequences

离线公共 Client E2E 证明形状、事件、工具配对、信任降级及最终请求守卫；它们使用合成官方
schema，不能声明真实续接已通过。真实目标若拒绝该投影，应保留失败捕获，先更新来源与失败
E2E 再调整此边界；不得丢推理、自动跨协议或以 UNSUPPORTED 放行必交能力。

[工单 03](../../.scratch/barness-ai-minimax/issues/03-responses-protocol.md)保存 red/green
证据指针；[工单 10](../../.scratch/barness-ai-minimax/issues/10-formal-catalog-support.md)承担线上
收敛。MiniMax 首期仍仅候选装配；是否正式支持由全部发布门槛独立决定。
