# Responses 基础公共 Client 契约探索（2026-10-17）

两个候选型号各一次，2 Logical Call / 2 Attempt / 2 外部请求，零重试；每次 high effort、store=false、最多 4096 输出 token、120 秒。使用既有公共 Client、显式中国区标准用途凭据及可信 payload 回调去掉未接入的 OpenAI 字段，不修改响应或 transport 上的 JSON。

两次均 HTTP 200/completed，实际 reasoning 与 message 的 content 都为数组，公共 Result 的正文、思考与 usage 保持完整；本次没有复现官方 schema 允许的单对象 content 丢失。该失败方式仍须用官方 schema 导出的离线公共 E2E 驱动修复，不能把数组样本称为对象兼容已通过。实际流含 reasoning_text/output_text 的 delta/done 和合法 completed；X-Request-Id 可用。

[计划](plan.md)、[归档索引](index.json) 与原始包位于根 `.evidence/barness-ai-glm/protocol-exploration/responses/01-basic/`。本目录 ZIP 按项目约定忽略，索引保留完整归档与包索引 SHA256。包内含源码、原始请求/SSE、公开事件/结果/Attempt/usage、SDK/目录 Hash、断言和脱敏审计；没有 Key 正文，全部审计通过。

这是先行 wire 取证，不是默认生产 Provider 配置、对象解析修复、工具回放、六条正式目录或完整 live/发布通过证明。

## 工具与 reasoning 请求转换的独立取证

新增 [固定四调用计划](tools-plan.md)，每型号自动函数调用及 Host 工具结果续接各一次，全部 HTTP 200；公共 ValidateToolCall 和 ToolResultText 参与，call_id 与 fc_ item ID 保持独立，参数 delta/done 均已捕获。5.2 首轮返回 reasoning 数组；可信 payload hook 验证原始原生条目与可见推理，再仅转换为官方 InputReasoning 的单对象 content 并移除不支持的原生 id/status/summary/encrypted 字段，续接成功且未发生 Native State 降级。5.3 工具首轮未返回可观察 reasoning，所以本组不证明其对象型推理回放；该项目仍须在工单 11 的预登记 native-history 场景取得自身证据。

合计 6 Logical Call / 6 Attempt / 6 外部请求，无重试和替代试参。两次基础与四次工具探索各自有锁定源码、预算、实际公开结果和脱敏审计，未改生产实现。新增原始包位于 `.evidence/barness-ai-glm/protocol-exploration/responses/02-tools/`，归档和索引 Hash 见 index.json。
