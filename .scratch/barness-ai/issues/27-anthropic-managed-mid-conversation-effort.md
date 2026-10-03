# 27: Anthropic 托管推理强度（claude-fable-5-1、claude-opus-5、claude-opus-5-5）

**What to build:** 支持每轮推理强度可变的 Anthropic 模型（pi compat `supportsMidConvoEffort`）：记录每条 assistant 消息当时的强度，回放历史时按 pi 插入强度标记，使对话中途调强度不破坏 thinking 绑定；完成后把三款模型列入内置目录。按 ADR-0018 托管强度是硬约束（pi 恒用 adaptive thinking 加 `drop_block`，以免前缀不匹配时持续 400，并且从不发送 temperature），所以本工单仍是这三款模型的列入条件。

**Blocked by:** 16

**Status:** ready-for-human

**Context:** 基线为 pi-ai 0.87.1 `anthropic-messages.ts`：`stream` 中 `providerThinkingLevel`、`buildParams` 的 managed effort 分支、`insertThinkingLevelMessages`、`convertMessages` 的 `assistantLevels`、`getBetaFeatures`。三款模型同时开启了工单 26 的工具变更开关。按 ADR-0018 那是可选特性，不阻塞本工单：26 未完成时按 pi 的回退发送当前工具列表，差异按工单 34 的方式登记为扩展。

- [x] `ModelCompat.SupportsMidConvoEffort`；`AssistantMessage` 增加 pi 的 `providerThinkingLevel`（本轮强度：`effort` 选项，缺省 high），随消息序列化与回放
- [x] 请求：thinking 恒为 `{type: adaptive, display, block_binding: {prefix_mismatch_behavior: drop_block}}`，`output_config: {effort: "high"}`；每条同 Provider 的 Anthropic 历史 assistant 消息前插入 `{role: system, content: [], output_config: {effort: <其强度>}}`，末尾插入本轮强度；不发送 temperature
- [x] 请求头加 `mid-conversation-output-config-2026-07-01` 与 `thinking-binding-controls-2026-08-01` beta
- [x] 判定 `providerThinkingLevel` 是否属于原生状态（spec I6：需可信封套才回放，还是与 pi 一样按 Provider 回放），结论记入 ADR
- [x] simple 入口沿用现有 adaptive effort 映射（含 opus-5-5 的 minimal 为 null）
- [x] 三款模型列入目录（目录版本与 pin 同步升级）
- [x] 离线 E2E：首轮、带不同强度历史的多轮、跨模型历史、full/simple；pi 差分无待处理差异

## Comments

**2026-10-03 — implemented; live smoke not run.** Status is ready-for-human. Everything offline is done. This environment has no test keys, so the live scenario waits for the maintainer.

- **Model and message.** `ModelCompat.SupportsMidConvoEffort` is new. `AssistantMessage.ProviderThinkingLevel` (JSON `providerThinkingLevel`) holds the turn's effort: the `Effort` option, or high when it is unset. The adapter sets it at the start of `stream`, so a failed message carries it too, as in pi.
- **Request** (`ai/anthropic_request.go`).
  - Thinking is always adaptive with `block_binding: drop_block`, the request-level `output_config` is always high, and no temperature is sent.
  - `anthropicMessages` returns the effort of each replayed Anthropic Messages turn from the same provider. `withThinkingLevels` adds the empty system markers after the cache marker is placed, with the call's own effort trailing.
  - `anthropicBetas` adds the two betas.
  - A level that is not an Anthropic effort is ignored, as in pi.
- **I6 decision (ADR-0019).** `providerThinkingLevel` is not native state. It is replayed by provider, as pi does, without a trusted envelope. A downgraded or cross-model turn from the same provider keeps its marker. GLOSSARY gains "Managed Effort".
- **Simple entry.** The existing adaptive mapping is unchanged. opus-5-5's null minimal maps to low, which pi does too (the differential confirms it).
- **Catalog** `2026-10-03.2`: claude-fable-5-1, claude-opus-5 and claude-opus-5-5 are added, with fields checked against pi by `TestCatalogInclusion/builtin-models-are-pi's`. The snapshot and pin are regenerated. Issue 34's "still unlisted" check is removed.
- **Offline E2E** (`TestManagedEffort`, `testdata/anthropic/managed-effort.json`, 8 scenarios plus an in-process round trip). The scenarios cover:
  - a first turn;
  - history at max/medium, plus a legacy turn and an invalid level;
  - cross-model history (fable-5-1 marker kept, gpt-5.4 none, an untrusted opus-5-5 turn downgraded but still marked);
  - a tool loop;
  - simple minimal, unset and xhigh-with-history;
  - a 400 whose failed message still records the effort.
  - The round trip passes a Result straight into the next call at a new effort and checks the JSON field. A mutation that drops the markers fails 6 checks.
- **Differential.** Every scenario joins P02. Five `extension` entries are added:
  - `opus-5-5-cross-model-history`: the spec I6 downgrade (`messages[7].content` and `Content-Length`);
  - `opus-5-tool-loop`: issue 26's optional native tool changes (`tools[1]` placeholder, the tool-changes beta in `Anthropic-Beta`, `Content-Length`).
  - Issue 26 now says to delete those three entries too.
  - No pending differences.
- **Docs.** ADR-0019 is new. ADR-0018 gains an implementation note and ADR-0011's note is updated. `differences.md` §1 shows P02 at 48, and the §6 row about unlisted models is replaced by a managed-effort row.
- **Live smoke (to run).** `effort-changes-<model>` (anthropic-messages, the three models) runs a turn at high, then replays it at low. It checks the betas, the `high,low` markers and the recorded levels. Run `anthropic-messages` and merge the report. If the vendor refuses an effort change, take the model out under a new catalog version and record it in ADR-0019 and the matrix.
- **Verified.** `go vet ./...` passes with and without `-tags live`. `go test ./...` and `BARNESS_AI_PIDIFF=1 go test ./...` both pass. The release gate was not run, because it needs the new live bundle.
- **Review (2026-10-03).** Neither the standards review nor the spec review found a hard violation or a missing requirement.
  - Fixed after the review: the Claude 5 `offNull` level map is defined once.
  - Left as judgement calls, because the code mirrors pi's structure: the `SupportsMidConvoEffort` checks are spread over betas, the body and the level, and the levels are an index map between `anthropicMessages` and `withThinkingLevels`.
  - Known limit: as in issue 34, the catalog lists the models before the live smoke confirms them.
  - The `level.valid()` check is the only gate before a host-supplied level goes on the wire. Keep it.
