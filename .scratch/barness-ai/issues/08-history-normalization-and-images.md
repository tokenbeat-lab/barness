# 08: 历史归一化、跨模型降级与图片输入（E03 其余部分）

**What to build:** 应用开发者传入任意已授权的多轮历史（含跨模型/跨 Provider 历史、工具调用与结果、system 和工具声明变化、图片），模块按冻结 pi 的规则确定性地转换为目标协议，不修改调用方原始历史（spec I6 表格、User Stories 11–14、19）。

**Blocked by:** 07

**Status:** resolved

- [x] content 缺失或 null 归一为空数组
- [x] 同模型 thinking：保留 redacted；带签名的空文本 thinking 保留；其他空白删除，非空保留
- [x] 跨模型 thinking：非空可见推理转 text，redacted 与空白丢弃；text 只保留文本；按原 truthy 条件删除非空 thoughtSignature，不合并缺失/null/空值（thinking/text 于本票实现；thoughtSignature 由 19 完成，见 Comments）
- [x] 按目标规范化工具 ID 并同步 toolResult 关联；缺少工具结果按基线顺序补 `No result provided` 的 isError 结果，调用与结果之间的 system 消息按原规则延后
- [x] error/aborted assistant 轮次按基线跳过；SystemPrompt 归一到初始 system 消息；system 与工具声明变化按输入顺序重放
- [x] 支持图片的模型发送 user 图片与工具结果图片；不支持时生成对应占位文本并按基线合并连续占位——两条路径分别有用例
- [x] 调用方原始历史在调用后逐字节不变
- [x] 以上全部场景在 Responses 上接入 pi 差分，无待处理差异

## Comments

**2026-10-01 — implemented** (E2E `ai/e2e/history_test.go` with fixture `testdata/responses/history.json`, 10 scenarios × 4 entry points plus 11 rejects × 4; differential `PIDIFF-P01-E03-history-*-{stream,streamSimple}`).

- New public model:
  - `SystemMessage{Content, Sections, ToolsAdded, ToolsRemoved}` joins the closed message set. It is pi's system message: `Request.SystemPrompt` and `Request.Tools` are the shorthand for a leading one. Sections are an ordered slice; a removed section is `Removed: true` (pi's `null`). Names that are array indices are ordered first, as JavaScript does in pi (documented on the field).
  - `Image{Data, MimeType}` is allowed in user and tool result content. `Data` is base64 sent as is, as in pi.
- Pipeline (`history.go`, `transcript.go`), run once in `call.go` before the adapter, ports pi in order:
  - `normalizeContext`, then `resolveTranscript`: later system messages collapse into one leading message unless the model takes them mid-conversation.
  - `transformMessages`: image placeholders, then the same-/cross-model rules (07's trust check decides "same"), then the adapter's tool call id normalization with the results remapped, then the second pass. That pass skips error/aborted turns, synthesizes `No result provided` results, and holds a system message until the open calls are answered.
  - Adapters get the prepared `transcript` and the current tool set. They supply only `historyRules` (mid-conversation flag, id normalizer).
- Downgrade counts now cover only replayed assistant messages. That fixes 07's known defect: an error/aborted turn is skipped and no longer counted.
- Responses (`responses_request.go`, `responses_history.go`):
  - Port of `normalizeToolCallId`. It works on UTF-16 code units, truncates to 64 and trims trailing `_`. A foreign item id becomes `fc_<shortHash>`. The target provider set is `{openai}`; pi's other two members are providers barness does not serve.
  - Port of `convertToolResultOutput`, and of the user `input_image` items.
  - Leading vs update rendering of system messages.
  - pi's `msg_pi_<n>` index rule: every non-leading system message counts.
- Catalog `2026-10-01.3`:
  - Adds `gpt-4`, the text-only model, so the placeholder path runs against pi's own catalog entry. It is in the test binding's allowed models.
  - Adds `ModelCompat.SupportsMidConvoSystemMessages`. No built-in model sets it: pi sets it only on reasoning models (09). The mid-conversation scenario sets it through a host catalog. The oracle runner's new `Case.ModelCompat` applies the same flags to pi's catalog model, as a pi custom model does. Without it, the "held system message" rule would be unobservable on Responses, because collapsing removes those messages first.
- Request validation (`invalid_request`, before any send), barness extensions where pi would crash or splice:
  - nil messages and blocks;
  - an image media type that is not `image/<subtype>`, since it is spliced into the data URL;
  - a section named twice, or removed while giving text;
  - a system message's tools checked like `Request.Tools`, and an empty removed tool name.
- Differential: all 20 history cases send pi's exact request body and produce pi's events and result. The only pending findings are the pre-existing categories owned elsewhere (`usage.cost`/`usage.reasoning` → 15, simple-entry `max_output_tokens` and its `Content-Length` → 09). Expected bodies in the fixture were captured from the frozen pi and checked by hand against the rules.
- Mutation-checked: dropping the collapse, the result remap, the error skip, the system hold, the image placeholders and their merge, the JS section order, the foreign hash or the system-message index each fail the E2E.
- **Deferred to 19:** pi deletes a truthy `thoughtSignature` from another model's tool calls without merging missing/null/empty. barness `ToolCall` has no such field: only Gemini produces and consumes it, and Responses never puts it on the wire. The Gemini adapter adds the field with a presence-aware type and this rule (`replay.go` `replayContent` notes it). Added as a criterion on 19; this ticket's criterion stays unchecked until then.
- Not covered: Responses `additional_tools`/tool search anchoring and grammar tools (pi `supportsAdditionalTools`/`supportsToolSearch`/`supportsOpenAIGrammarTools`). They only exist on reasoning models and arrive with those models (09).

**2026-10-01 — code review** (standards + spec, two parallel reviewers): no wrong implementations; the port matches pi line by line and the differential shows only the known categories. Applied: `Model.acceptsImages` replaces two copies of the image check; `answerToolCalls` → `settleTurns` and `replayNative` → `noDowngrade` (names now describe what they do); the test differential reuses the fixture scenario and one `withModelCompat` helper. Noted, not changed:
- The request rejections (nil message/block, non-`image/<subtype>` type, duplicate or contradictory sections, tool checks in system messages) are barness extensions pi lacks; documented above.
- `MaxImageBytes` is not enforced yet (ticket 12).
- An untrusted same-model message is normalized as cross-model, tool ids included. The spec requires this; 07's no-envelope E2E covers the body.

**2026-10-02 — closed from issue 24.** The thoughtSignature criterion handed to 19 is done there (19's checked criterion "ToolCall 增加区分缺失/null/空值的 thoughtSignature…", Gemini E03 cases); the release gate's traceability item C03 (all E03 cases, P01–P06) passes. Status set to resolved.
