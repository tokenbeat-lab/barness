# 06: 工具调用往返与工具 JSON 助手

**What to build:** Agent 开发者在 Responses 上声明工具、收到带 ID/名称/参数增量的工具调用事件，由宿主执行后把工具结果作为下一轮输入再次调用；部分 JSON 仅供展示，修复和 JSON Schema 校验由调用方显式使用（spec I6 工具部分、User Stories 15–17）。

**Blocked by:** 04

**Status:** ready-for-agent

- [ ] Request 支持 Tools 声明；流产生 toolcall start/delta/end，保留原始 JSON 与增量，参数分片跨帧时正确拼接
- [ ] 部分解析结果只用于展示，类型上与“完整且已校验参数”可区分；模块从不执行工具
- [ ] 公开工具 JSON 修复与 JSON Schema 转换/校验助手：先枚举其全部失败方式并写行为测试，再实现
- [ ] 本地 E2E 演示完整往返：模型工具调用 → 测试宿主执行并构造 ToolResult → 下一次逻辑调用（新 RequestID）→ 最终文本
- [ ] 截断（length / error）中的工具调用不能通过助手得到“可执行”结论
- [ ] Responses 工具参数分片与 toolChoice 请求字段接入 pi 差分，无待处理差异
