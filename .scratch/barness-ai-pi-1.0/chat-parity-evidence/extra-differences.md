# 工单 03：五项之外的 1.0.0 差异，已批准同步

2026-10-08 全量新基线差分发现两类差异，共四个用例、八个 pending 叶字段。
均已核查冻结发布件中的 OpenAI Node SDK **7.19.0** 源码，未扩大现有 extension 登记。
维护者于 2026-10-08 回复“按建议同步”，批准同步下列文本；原 extension 登记保持不变，新增处置为 fixed。

| 用例 | 旧行为 | 1.0.0 行为 | 来源 |
| --- | --- | --- | --- |
| PIDIFF-P04-http-500-no-error-stream | `500 status code (no body)` | `500 {"detail":"boom"}` | SDK client `makeStatusError` 将缺少/null `error` 的对象包装为 `{error: body}` |
| PIDIFF-P01-E02-http-500-no-error-object-stream | `OpenAI API error (500): 500 status code (no body)` | `OpenAI API error (500): 500 {"detail":"internal failure"}` | 同一 SDK 规则；pi `utils/error-body.ts` 保留 SDK 文本 |
| PIDIFF-P01-E02-stream-error-event-stream | `Error Code server_error: The server had an error while processing your request.` | `The server had an error while processing your request.` | SDK `Stream.fromSSEResponse` 在具名 `event:error` 上抛 `APIError(data.error ?? data)`，先于 pi adapter |
| PIDIFF-P01-E05-stream-error-not-replayed-stream | `Error Code server_error: The server had an error` | `The server had an error` | 同一 SSE 规则，不重放流的行为保持不变 |

每个用例的 `result.errorMessage` 和终态 `events[*].error.errorMessage` 各出现一项差异。
观察文本来自同一合成 fixture 的两侧公共入口输出；不是 SDK 推测或手工重建的 golden。

一手来源：[OpenAI SDK 7.19.0 client](https://github.com/openai/openai-node/blob/v7.19.0/src/client.ts#L863-L872)、
[7.19.0 streaming](https://github.com/openai/openai-node/blob/v7.19.0/src/core/streaming.ts#L149-L150)、
[pi v1.0.0 error body](https://github.com/earendil-works/pi/blob/v1.0.0/packages/ai/src/utils/error-body.ts)。
锁定依赖与核验来源见 oracle 的 PROVENANCE.md。

## 实施与回归控制

已先更新上述四个 fixture 的期望，再实现 HTTP 错误体包装及 SSE 事件名处理。
Responses 与 Chat 在 SDK 原始 SSE decoder 读取的帧上处理具名 error，再做协议 DTO 解码；
分类、超时、取消、限额、脱敏与响应体关闭仍经过原有边界。

额外 fixture 覆盖 missing/null error、带根 message 的对象、空对象、数组、嵌套对象与 false error，
以及具名／未具名、无 type、嵌套／null／false error 和字符串／数组／数字／false／null 数据帧。
未具名 type:error 仍进入 Responses adapter 并保留 Error Code 文本；未具名 null/false error 不误报。
thread.* 对照保留 SDK 原有包装语义：其内容不会成为协议事件，也不会误触发错误。

复核发现 Chat 提前将非对象具名错误当作普通 chunk 跳过，已先补失败 E2E，再前移错误判定；
两轴复核均确认修复。Responses 原始解码路径的 thread.* 回归对照也先失败再修复。
单用例 replay 的共享 fixture 在 key 注册前写入所导致的审计失败，已通过 TestMain 提前注册
已知合成 key 修复；不依赖其他测试先行构造世界。

最终全量差分、race、压力验证与脱敏审计结果见 [验证记录](README.md) 和 [机器报告](report.json)。
