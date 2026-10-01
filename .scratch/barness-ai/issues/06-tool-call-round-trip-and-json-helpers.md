# 06: 工具调用往返与工具 JSON 助手

**What to build:** Agent 开发者在 Responses 上声明工具、收到带 ID/名称/参数增量的工具调用事件，由宿主执行后把工具结果作为下一轮输入再次调用；部分 JSON 仅供展示，修复和 JSON Schema 校验由调用方显式使用（spec I6 工具部分、User Stories 15–17）。

**Blocked by:** 04

**Status:** ready-for-agent

- [x] Request 支持 Tools 声明；流产生 toolcall start/delta/end，保留原始 JSON 与增量，参数分片跨帧时正确拼接
- [x] 部分解析结果只用于展示，类型上与“完整且已校验参数”可区分；模块从不执行工具
- [x] 公开工具 JSON 修复与 JSON Schema 转换/校验助手：先枚举其全部失败方式并写行为测试，再实现
- [x] 本地 E2E 演示完整往返：模型工具调用 → 测试宿主执行并构造 ToolResult → 下一次逻辑调用（新 RequestID）→ 最终文本
- [x] 截断（length / error）中的工具调用不能通过助手得到“可执行”结论
- [x] Responses 工具参数分片与 toolChoice 请求字段接入 pi 差分，无待处理差异

## Comments

**2026-10-01 — implemented** (E2E `ai/e2e/tool_round_trip_test.go` with fixture `testdata/responses/tool-round-trip.json`; helper tests `ai/e2e/tool_json_helpers_test.go`; differential `PIDIFF-P01-E03-tool-*`, `PIDIFF-P01-E02-tool-*`).

- Protocol data: `Request.Tools` (`ai.Tool{Name, Description, Parameters}`), the `ToolCall` assistant block, `ToolResultMessage`, `AssistantMessage` as replayable history, `StopReasonToolUse`, and `toolcall_start/delta/end` events. The Responses parser ports pi's function_call path: deltas are joined per output index, `function_call_arguments.done` publishes the part beyond what was streamed as a final delta, `output_item.done` closes the call, and a completed turn with calls is `toolUse`. The tool id is pi's `call_id|item_id`. Requests send `tools` (`strict:false`, because pi's catalog marks these models `supportsStrictMode`; `Model.Compat` was added, catalog `2026-10-01.2`), `tool_choice` (full: `ResponsesToolChoice` mode or function; simple: `auto`/`none`), and replay `function_call`/`function_call_output` items as pi's `convertResponsesMessages` does. That includes text ids from the TextSignatureV1, pi's `msg_pi_N` fallback and `shortHash` for over-long ids.
- Display vs executable: `ToolCall.Arguments` is `UncheckedArguments`, pi's best-effort parse (`parseStreamingJson`, with a port of partial-json 0.1.7) stored in JSON.stringify form. Key order, duplicate keys and number spelling follow JavaScript, so a replay's `arguments` string is byte-identical to pi's. `ToolCall.RawArguments` keeps the provider's text. Only `ValidateToolCall(msg, i, tools)` yields an opaque `ValidToolCall`. It refuses unless the message ended `toolUse` (so length/error/aborted/pending are never executable), parses the *raw* text strictly with pi's repair, applies pi's `normalizeOptionalNulls` + `coerceWithJsonSchema`, then validates with santhosh-tekuri/jsonschema v6, with no loader: remote and file `$ref`s fail and are never fetched. `RepairToolJSON` is public. The module never executes a tool.
- Helper failure modes are listed in the test's header (G/A/S/J/V/OK/R) and were written before the implementation; pi's `validation.test.ts` conversion tables are included. Mutation-checked: dropping the stop-reason gate fails G1–G5; allowing the file loader fails S5; dropping the toolUse mapping or the done-delta fails the round trip.
- Differential: tools, `tool_choice` (`required`, `auto`, function object) and the round-2 history replay match pi exactly. No 06-owned pending items remain; the rest belong to 09/15. Ledger extensions (spec §6 "工具参数保留原始 JSON 与增量"): barness persists the raw argument text as `rawArguments`; pi only has a transient `partialJson` scratch buffer, which it leaves on an unclosed call of a length-truncated message. The approved disconnect-text decision now also lists `PIDIFF-P01-E02-tool-disconnect-mid-arguments-stream` (same undici cause).
- Deliberate differences from pi, not differential-covered because pi's helper is not driven by the oracle runner:
  - `ValidateToolCall` reads the raw text rather than the display parse, which is the completeness guarantee spec §6 asks for.
  - It does not run TypeBox's generic `Value.Convert` before pi's own JSON-schema coercion; that is the path pi takes for every non-TypeBox schema, which is all a Go host can pass.
  - Problem wording comes from the schema validator, laid out in pi's "Validation failed for tool …" format.
  - Where pi would return an unvalidated root after a failed root coercion, barness refuses.
  - partial-json can loop forever on inputs like `[1e5}`; the port stops the array instead.
- Scope notes:
  - Request structure checks reject unnamed or duplicate tools and non-object `Parameters` (phase `scope`, now documented as scope + request structure). pi would send them as is. This follows spec §6 "构造及解码检查非法变体组合"; no differential case reaches it.
  - `ResponsesToolChoice` covers `auto`/`none`/`required`/function. The remaining Responses `tool_choice` forms (allowed_tools, hosted tools) are for 09's "toolChoice 完整保留".
  - There is no message-free `validateToolArguments` counterpart. Executable status is only obtainable through the stop-reason gate.
- Left to later tickets: thinking blocks are not replayed until 07's trusted envelope. That is harmless while the catalog lists only non-reasoning models, but a same-model reasoning tool round trip needs 07. Cross-model tool id normalization, synthetic missing results, error-turn skipping and tool-result images belong to 08. Custom/grammar tool calls and tool namespaces are not parsed, because no tool barness declares produces them. Tool JSON byte limits belong to 12.

**2026-10-01 — review follow-up** (two-axis code review): `ValidateToolCall` detects a call by type (a blank-named call is `unknown_tool`). Clearer names (`schemaAccepts`, `primitiveReplaced`, `jsNumberLiteral`, `jsStringToNumber`). One `displayArguments` helper. The `fc_` replay condition is simplified. The adapter takes only value `ResponsesOptions`. Invalid host-built history arguments are documented as `invalid_request`. The function form of `tool_choice` is now covered offline and in the differential.
