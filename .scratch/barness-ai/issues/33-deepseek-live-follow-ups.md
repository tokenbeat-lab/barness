# 33: DeepSeek 真实冒烟后的跟进：请求 id、用量完整性与 fixture 修正

**What to build:** 按 2026-10-02 首次真实冒烟观察到的 DeepSeek 行为（工单 23）落实 ADR-0014/0015 的条件决定与维护者 2026-10-03 的决定：读取 DeepSeek 的厂商请求 id，关闭思考时缺推理计数不算部分上报，P05/P06 fixture 改为实测形状，删除 `auth-refused` 冒烟场景。

**Blocked by:** 23

**Status:** resolved

- [x] DeepSeek 两协议以 `x-ds-trace-id` 作为 `ProviderRequestID`（DeepSeek 不发送 `x-request-id`）；OpenAI 仍为 `x-request-id`
- [x] Chat 用量只在本次调用请求了推理时要求推理计数：DeepSeek 关闭思考时缺 `completion_tokens_details` 为 complete，开启思考时缺失仍为 partial
- [x] P05 fixture：UUID 形式的响应/条目 id、`call_00_…` 调用 id、推理条目的 `encrypted_content`、消息条目的 `phase`；401 错误体为实测形状（key 只显示后四位，消息末尾带 request_id）并带 `x-ds-trace-id` 响应头；P06 的 401 同样修正
- [x] 删除两个 DeepSeek 组合的 `auth-refused` 冒烟场景

## Comments

**2026-10-03 — implemented.**

- **Request id** (`requestIDHeaderOf` in `ai/openai_sdk.go`). It is keyed by provider, as `responsesCapabilitiesOf` and `chatCompatOf` are, and used by both OpenAI-protocol adapters for errors and attempts. Body vs header: DeepSeek's 401 body also carries a `request_id` UUID, which differs from `x-ds-trace-id`. The header was chosen because it comes on every response, success included, so `Attempt.ProviderRequestID` gets it too. `ProviderRequestID` is a barness extension, so the differential is unaffected.
- **Usage reporting** (`chatUsage.reasoningExpected`, `ChatOptions.reasoningExpected`). A reasoning count is expected only from a reasoning model and, under DeepSeek's thinking switch, only when thinking is on. OpenAI always sends the count, so nothing changes there. A payload callback that switches thinking on itself is not seen; such a call's missing count reads as complete (stated on the method).
  - Fixtures: `usage-without-reasoning-details` (thinking off) now expects complete; the new `usage-thinking-without-reasoning-details` (thinking on) expects partial.
  - A mutation check (the rule forced to "never expected") fails the thinking-on case.
- **P05 fixtures.** `text.json` (`reasoning`, `reasoning-simple`, `reasoning-tool-call`) and `history.json` (`tool-round-trip`, which replays round one) now use DeepSeek's live shapes. Expected values follow from pi's rules, not from the output:
  - the thinking signature is the done reasoning item's JSON, now with `encrypted_content`;
  - the text signature (TextSignatureV1) carries the message's `phase`;
  - a replayed `function_call` keeps only an `fc_` item id, so DeepSeek's UUID item id is dropped on replay. DeepSeek accepted that live.

  The other P05 scenarios keep illustrative `msg_…` ids: they exercise no shape-dependent rule.
- **401 fixtures (P05, P06).** These now use the live body. They expect `providerRequestId` from `x-ds-trace-id`. P05's 401 no longer exercises key-echo redaction, because DeepSeek masks the key itself; P01's key-echo cases still cover the redaction.
- **Live.** `auth-refused` and its helper are removed. The largest combination now makes 13 calls. The support matrix's DeepSeek rows keep their past `auth-refused` capability until the next merge, which never removes capabilities; it is a PASS and does not affect the gate.
- **Verified.** `go test ./...`, the pi differential and the DeepSeek E2E all pass. This changes DeepSeek adapter code (request id and usage reporting), so spec §6 asks for the two DeepSeek combinations to be rerun live.
