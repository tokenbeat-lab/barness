# 07: 原生状态封套与 reasoning 同模型回放（E03 原生状态部分）

**What to build:** 应用开发者把带签名/密文的推理状态连同来源封套保存，宿主鉴权后通过 Go 构造入口（如 `TrustNativeState(scope, envelope)`）取得可信值并回放给同 Provider/API/model；无可信封套或账户不匹配时自动降级、调用照常成功（spec I6 后三段、User Stories 18–20）。

**Blocked by:** 06

**Status:** ready-for-agent

- [x] Responses reasoning 在同模型回放中完整保留适用的加密内容/签名与 redacted 状态
- [x] 原生状态封套记录 TenantID、AccountScopeID、Provider/API/Model；可信值只能经 Go 构造入口获得，JSON 反序列化（含 `trusted=true`、自报 TenantID、签名字段）无法得到
- [x] 库只核对封套来源与当前 scope、binding 匹配；同账户换 key 不因凭据版本不同而丢弃状态
- [x] 无可信封套、或同 Provider/API/model 但 AccountScopeID 不一致时，按跨模型规则降级且不返回 invalid_request
- [x] 降级计数与原因（无封套 / 账户不匹配 / 跨模型）写入 Result 调用元数据（Observer 部分见 14），不进入 pi 兼容消息、不参与 pi 差分
- [x] 回放后的实际 wire 字段与 pi 差分一致

## Comments

**2026-10-01 — implemented** (E2E `ai/e2e/native_state_test.go` with fixture `testdata/responses/native-state.json`; differential `PIDIFF-P01-E01-reasoning-{call,replay,cross-model}-*`).

- Public contract: `NativeStateEnvelope{TenantID, AccountScopeID, ProviderID, API, ModelID}` is plain, storable provenance. `TrustedNativeState` has only unexported fields, and `AssistantMessage.NativeState` is `json:"-"`, so decoding JSON never yields trust. That covers a forged record with `trusted:true`, a self-reported tenant or the original signatures (E2E `forged-record`, `TestTrustNativeState` T1/T5). Trust has two sources:
  - `TrustNativeState(scope, env)`, called by the host after its own storage checks. It refuses an incomplete envelope or one naming another tenant than the scope, with `ErrInvalidNativeStateEnvelope`. That is a host fault, never a call error.
  - The library itself, which stamps every message of a resolved call. A Result passed straight into the next call is therefore replayed natively, which is why the existing tool round trip is unchanged. Recorded in ADR-0001 Consequences.
  - `Envelope()` returns the provenance for the host to store.
- Call-time check (`native.go` `nativeReplay`, run after the model is authorized and before the credential read, per spec I3 "检查能力、选项和历史来源"):
  - The message's own provider/API/model must equal the target's (pi's rule).
  - The trusted envelope must be for the calling tenant, the same provider/API/model, and the binding's `AccountScopeID`.
  - Credential and binding versions are not compared, so K1→K2 on the same account keeps replaying (`key-rotated`).
  - Otherwise the message is converted by pi's cross-model rules and counted in `CallMetadata.NativeStateDowngrades{NoEnvelope, AccountMismatch, CrossModel}`. A state trusted for another tenant counts as `NoEnvelope`, and an envelope naming another model than the message counts as `CrossModel` (`relabeled-model`). The call succeeds.
  - The counts are call metadata only (`omitzero`), never in the message JSON or pi golden.
- What counts as native state: a reasoning signature or redacted thinking, a text signature (Responses message item id/phase), and a tool call's provider item id (`call|fc_…`). A message carrying none has nothing to vouch for: it follows pi's field rule alone and is not counted.
- Same-model replay (`replay.go`): ports the thinking/text half of pi's `transformMessages`.
  - Same-model keeps redacted and signed thinking (also empty) and drops other blank thinking.
  - Cross-model turns visible thinking into text, drops redacted and blank thinking, and strips text signatures.
  - Responses then sends each remaining signed thinking block as `JSON.parse(signature)`, and omits the `fc_` id for a downgraded or different-model call, as pi does for another model. `Thinking.Redacted` was added. A trusted signature that is not JSON is `invalid_request`; untrusted ones are never parsed.
- Stream side (`responses.go`): `thinkingSignature` is now pi's `JSON.stringify(item)`, using the JS-compatible JSON helpers, so escapes and numbers normalize as in pi. pi's `backfillReasoningSignatures` is ported: a truthy `encrypted_content` from the terminal response's output fills a finished block whose stored item lacks one. A key that is present is replaced in place; otherwise it is appended.
- Differential: round 1 (signatures and backfill), the trusted same-model replay (reasoning items, redacted, message phase/id, `fc_` id) and the cross-model replay match pi exactly on the request body, events and result content. The only pending items in these cases are the pre-existing categories owned elsewhere, which also fail the baseline: `usage.cost`/`usage.reasoning` (15) and simple-entry `max_output_tokens` with its derived `Content-Length` (09). Downgrading same-model state has no pi counterpart. Offline, the downgraded body is asserted equal to the body pi sends for the cross-model replay (same fixture `downgradedBody`).
- Mutation-checked: ignoring the tenant, account or envelope-model check, skipping the trust check, dropping the backfill, and keeping raw signatures each fail the E2E.
- Left to later tickets:
  - Observer emission of the downgrade counts (14).
  - Cross-model tool ID normalization, synthetic missing results and skipping error/aborted turns (08). Until 08 lands, a downgraded error turn is still replayed and counted. The fixture's ids are unchanged by pi's normalization.
