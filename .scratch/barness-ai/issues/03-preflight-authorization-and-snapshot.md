# 03: 授权预检、凭据解析与配置快照（E07 / D2）

**What to build:** 缺身份、越权、凭据不可用或配置不一致的调用在发出任何 Provider 推理请求前以错误终态结束；新调用取得更新或撤销后的凭据状态，在途调用固定使用原快照。覆盖 spec I3、T01–T05、T10–T11 与 ADR-0003（D2）。

**Blocked by:** 01

**Status:** ready-for-agent

- [x] 调用顺序按 spec I3：scope/结构/大小检查 → 绑定解析 → 能力/选项/历史来源检查 → 凭据解析 → 租户/账户/引用/版本一致性校验 → 准入 → 发送（结构/大小检查归 12、准入归 13，已在 `execute` 标出位置）
- [x] 缺 TenantID 或 RequestID、binding 不存在或不属于该租户、调用主体无权、模型不在 AllowedModels ∩ 目录、binding 禁用、缺 key、秘密后端故障：均以错误 assistant 消息结束，Code 分别可用 errors.Is/As 区分（invalid_request / tenant_denied / binding_not_found / credential_unavailable 等），本地 Provider 收到零推理请求
- [x] 以上拒绝在 full/simple × Stream/Complete 四个入口均验证，不存在绕过解析器的网络入口
- [x] 两次解析间发生更新或撤销且无法取得一致快照时直接失败，Phase 指明配置一致性阶段，无内部重解析，错误不泄漏秘密；宿主用新 RequestID 再调用可见新的一致快照
- [x] K1→K2 更新后新调用使用 K2；撤销阻止后续新调用；在途调用继续使用原快照（在途重试待 11 落地后验证）
- [x] 普通 Request 不能携带 key、凭据引用、endpoint、Authorization、代理或可执行函数（类型层面不提供该字段，或在边界拒绝）
- [x] 测试环境设置污染的 OPENAI_API_KEY / ANTHROPIC_API_KEY / GOOGLE_API_KEY 及 endpoint 变量，证明无隐式回退或误出网
- [x] 预检失败的 Result 中未解析出的 Provider/模型/账户标记为未解析，不使用请求自报值
- [x] 秘密值不出现在消息、事件、Result、错误文本中

## Comments

**2026-10-01 — implemented** (preflight in `ai/preflight.go`, order in `ai/call.go`; E2E in `ai/e2e/preflight_reject_test.go` and `ai/e2e/preflight_snapshot_test.go`).

- Resolver contract additions: `Binding.Enabled` and `Credential.Active` (zero value refuses, fail-closed), `Credential.AccountScopeID` (the credential snapshot's account ownership, per GLOSSARY), and sentinels `ai.ErrAccessDenied` (→ tenant_denied) and `ai.ErrSnapshotConflict` (→ credential_unavailable / Phase `consistency`). Any other resolver error is classified without echoing its text.
- D2 consistency check (`checkSnapshot`): credential owner = scope tenant (else tenant_denied), CredentialID = binding.CredentialRef, AccountScopeID equal and non-empty, both versions non-empty; otherwise credential_unavailable in Phase `consistency`. A consistent new credential version read between the two resolutions (K1→K2 rotation) is not a conflict. No re-resolution: E2E counts one read per resolver per RequestID.
- `CallMetadata.Resolved` (and Provider/API/Model/AccountScopeID, message identity) is now set only once the snapshot is consistent, so every preflight failure — credential and consistency included — reports unresolved identity.
- Binding/model rejections: binding of another tenant or a misfiled host record → binding_not_found (existence not revealed); disabled binding or actor refused by host → tenant_denied; unsupported AuthKind → invalid_request.
- Request surface is enforced at the type level (`TestRequestSurface` walks every per-call input type for authority-bearing names and func/chan fields).
- Binding-version consistency is a host duty: the library cannot see the backend's current binding version, so `CredentialResolver` must return `ErrSnapshotConflict` for a stale binding version (documented on the interface); the library itself only requires both snapshots to be versioned. `Credential.AccountScopeID` and `CallMetadata.AccountScopeID` are deliberate additions beyond the spec's field lists.
- Request structure/size checks remain ticket 12; admission slots in after the consistency check (ticket 13). In-flight retries using the original snapshot are verified once ticket 11 lands; this ticket verifies an in-flight call keeps its snapshot across a rotation and revocation.
- Test kit: the host double gains actor restrictions, misfiled records, credential backend faults, a hook between the two reads, per-RequestID read counts and a strict binding-version check; the controlled Provider gains `Reply.OnReceive`.
