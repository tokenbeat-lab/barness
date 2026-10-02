# 32: 流式工具参数逐 delta 全量重解析（CPU 随参数大小平方增长）

**What to build:** 工具调用参数流的展示解析不再随参数大小平方增长，使资源策略的 `MaxToolJSONBytes` 只受内存约束而非 CPU 约束；每个 toolcall delta 事件与 PartialView 中可观察的部分参数保持与冻结 pi 一致（spec I6 工具参数、I9 资源策略）。

**Blocked by:** —

**Status:** needs-triage

**Context:** 工单 24 的压力场景（`ai/e2e/policy_pressure_test.go`，`BARNESS_AI_PRESSURE=1`）发现：Responses 单次调用流式传入的工具参数，64 KiB 耗时 0.4 s，128 KiB 1.6 s，256 KiB 6 s，512 KiB 23 s（Apple M 系列，单调用，参数以 64 字符 delta 到达）。CPU 剖析显示热点为 `assembler.toolCallDelta` → `replaceToolCallArguments` → `displayArguments` → `parseStreamingJSON`（`ai/tool_json.go`）：每个 delta 都对累计的全部参数文本做一次 `parseJSONWithRepair`/`parsePartialJSON` 与 `stringifyJSON`，这是 pi `parseStreamingJson` 的逐行移植，pi 在每个 delta 上同样全量重解析。

**影响与风险：** 恶意或失控的模型输出可在 `MaxToolJSONBytes` 以内让单个调用占用一个 CPU 核数十秒；并发调用叠加。工单 24 据此把示例策略的 `MaxToolJSONBytes` 降为云端 128 KiB、本地 256 KiB（`ai/examples/*/policy.go` 注释记录了测得数值），这是临时方案。

**移除条件：** 解析改为增量或按需（例如只在消费者读取 PartialView/事件参数时解析，或复用上次解析状态），压力场景测得的单调用耗时与参数大小近似线性；之后可按内存预算重新放宽示例策略的 `MaxToolJSONBytes` 并更新注释与压力场景的设计负载。

- [ ] 先枚举可观察行为（每个 toolcall_delta 事件的 arguments、PartialView 快照、最终 arguments/rawArguments、修复与部分解析的边界），以现有 E03 工具 JSON 场景与 pi 差分为准，确认改动前后差分无新增差异
- [ ] 解析成本随参数大小近似线性；在压力场景加入大参数工具调用的耗时记录
- [ ] 放宽示例策略的 `MaxToolJSONBytes` 并更新数值依据
