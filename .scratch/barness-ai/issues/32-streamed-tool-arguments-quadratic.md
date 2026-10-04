# 32: 流式工具参数逐 delta 全量重解析（CPU 随参数大小平方增长）

**What to build:** 工具调用参数流的展示解析不再随参数大小平方增长，使资源策略的 `MaxToolJSONBytes` 只受内存约束而非 CPU 约束；每个 toolcall delta 事件与 PartialView 中可观察的部分参数保持与冻结 pi 一致（spec I6 工具参数、I9 资源策略）。

**Blocked by:** —

**Status:** resolved

**Context:** 工单 24 的压力场景（`ai/e2e/policy_pressure_test.go`，`BARNESS_AI_PRESSURE=1`）发现：Responses 单次调用流式传入的工具参数，64 KiB 耗时 0.4 s，128 KiB 1.6 s，256 KiB 6 s，512 KiB 23 s（Apple M 系列，单调用，参数以 64 字符 delta 到达）。CPU 剖析显示热点为 `assembler.toolCallDelta` → `replaceToolCallArguments` → `displayArguments` → `parseStreamingJSON`（`ai/tool_json.go`）：每个 delta 都对累计的全部参数文本做一次 `parseJSONWithRepair`/`parsePartialJSON` 与 `stringifyJSON`，这是 pi `parseStreamingJson` 的逐行移植，pi 在每个 delta 上同样全量重解析。

**影响与风险：** 恶意或失控的模型输出可在 `MaxToolJSONBytes` 以内让单个调用占用一个 CPU 核数十秒；并发调用叠加。工单 24 据此把示例策略的 `MaxToolJSONBytes` 降为云端 128 KiB、本地 256 KiB（`ai/examples/*/policy.go` 注释记录了测得数值），这是临时方案。

**移除条件：** 解析改为增量或按需（例如只在消费者读取 PartialView/事件参数时解析，或复用上次解析状态），压力场景测得的单调用耗时与参数大小近似线性；之后可按内存预算重新放宽示例策略的 `MaxToolJSONBytes` 并更新注释与压力场景的设计负载。

- [x] 先枚举可观察行为（每个 toolcall_delta 事件的 arguments、PartialView 快照、最终 arguments/rawArguments、修复与部分解析的边界），以现有 E03 工具 JSON 场景与 pi 差分为准，确认改动前后差分无新增差异
- [x] 解析成本随参数大小近似线性；在压力场景加入大参数工具调用的耗时记录
- [x] 放宽示例策略的 `MaxToolJSONBytes` 并更新数值依据

## Comments

**2026-10-04 — agent implementation:**

可观察行为枚举（以 pi 为准，改动前后一致）：toolcall_start 时 arguments 为 `{}`（不论条目宣告的原始文本）；每个 toolcall_delta 时读到的 PartialView 参数 = 当时原始文本的展示解析（完整/可修复→精确解析，否则部分解析，再否则 `{}`）；toolcall_end 事件与最终消息的 arguments = 最终原始文本（空为 `{}`）的展示解析，rawArguments 为提供方原文；超 `MaxToolJSONBytes` 或中途截断时保留最后一个限额内文本及其解析；Responses `arguments.done` 不延续已流文本时静默替换；并发读取视图无竞态。冻结 pi 差分（逐事件快照 PartialView）改动前后均为 470 个子用例全部通过、无新增差异。

两处平方成本：

1. `assembler.toolCallDelta` 每个 delta 调用 `displayArguments` 全量解析。改为按需：`PartialView` 只记录原始文本并标记该块待解析，Snapshot/MarshalJSON/终态读取时解析并缓存到下一个 delta（`ai/partial.go`）。读到的值与 pi 在该时刻持有的相同；不读取视图的消费者（只转发 delta 的宿主、Complete）不付解析成本。
2. 去掉 1 之后剖析显示适配器的 `raw += delta`（每个 delta 复制整段文本、并造成大块分配与 GC）仍呈平方增长。Responses、Anthropic、Chat Completions 改用只追加的 `argumentText`（`ai/argument_text.go`），已返回的文本不再变化；调用结束（toolcall_end，或被截断的调用在消息终态时）复制一次以释放缓冲区的多余容量。

新增 `E08-policy-pressure-tool-arguments-scaling`（`ai/e2e/tool_args_scaling_test.go`，`BARNESS_AI_PRESSURE=1`）：改动前 64 KiB 0.45 s、256 KiB 6.2 s；改动后 128 KiB 12 ms、512 KiB 48 ms、2 MiB 192 ms，每字节耗时之比约 1.0（判据 ≤ 3）。示例策略恢复为云端 512 KiB、本地 1 MiB，注释按内存预算说明；设计负载改为本地 512 KiB、云端 256 KiB 工具调用，压力场景通过（工具 JSON 用量 50%，堆增长在预算内）。ADR-0017 追加“后续”一节，README 与差异登记同步。

遗留（不另开工单）：消费者在每个 delta 后都读取视图时，每次读取仍完整解析一次参数，成本与 pi 相同；仓库内暂无这样的消费者。若出现（例如逐 delta 渲染部分参数的 TUI），再考虑增量解析。
