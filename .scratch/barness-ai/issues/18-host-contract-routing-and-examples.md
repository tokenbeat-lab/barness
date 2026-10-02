# 18: 宿主契约、路由与最小示例（E10，批次 B 收口）

**What to build:** 上层模块开发者通过三个最小示例理解职责交接：本地显式装配租户与 key、宿主由认证结果建立 scope 并按租户读取历史与验证原生封套、工具往返；宿主汇合多个流时每个事件携带不可变调用归属；下游断开只取消本次生成（spec I2 末段、Testing Decisions §1、User Story 49）。批次 B（Responses + Anthropic）在本票通过后视为验收完成。

**Blocked by:** 07, 12, 13, 14, 17

**Status:** ready-for-agent

- [ ] 事件经信封或等价不可变引用携带 TenantID、RequestID、BindingID、实际 Provider/API/ModelID；多流汇合用例不依赖到达顺序推断归属
- [ ] 本地装配示例：宿主显式读取指定环境变量或秘密文件后注入，核心不自行读取
- [ ] 宿主接入示例：认证结果 → CallScope、按 `(tenant_id, session_id)` 读取历史、TrustNativeState、context 取消传播；恶意自报 TenantID / 历史引用由宿主边界拒绝，直接向库提交无效 scope/绑定由库拒绝——两类责任分别断言
- [ ] 工具往返示例：宿主执行工具并以新 RequestID 发起下一轮
- [ ] 下游断开或发送失败只取消本次生成，其他流不受影响
- [ ] Responses 与 Anthropic 分别通过 E01–E09、E11 适用场景（含 12/13 的资源与准入场景）
- [ ] 示例本身作为离线 E2E 运行并产出证据包
- [ ] 示例给出完整的有限资源策略（spec §9“正式示例”、ADR-0002），每个数值附适用负载、取值依据与调整说明；准入相关数值按下文“准入取值建议”推算，并由突发压力场景的数据确认，不写成已验证的通用默认值

## 准入取值建议

记录于 2026-10-02（工单 13 / ADR-0008 收尾时）。这些是推算方法和起点值，没有压测数据支撑。示例中的最终数值以本票压力场景的结果为准。

### 取值思路

1. **只让可能等到许可的尝试排队。** 等待者只有在 `AdmissionWait` 内有许可释放时才有用，否则必然超时，只是多等了一段。所以上限应接近等待期间预计释放的许可数：

   ```
   MaxAdmissionWaiters ≈ MaxConcurrentProcess × AdmissionWait ÷ 单次调用典型时长（p50）
   ```

   再留少量余量，最少取 1。

2. **LLM 调用时间长，所以这个值通常很小。** 流式生成常常要 10 到 30 秒。例如进程并发 32、等待 2 秒、典型时长 20 秒时，32 × 2 ÷ 20 ≈ 3，取 4 左右即可。设得更大，只会让更多请求等满 `AdmissionWait` 后照样被拒。

3. **按内存兜底。** 准入发生在请求体构造之后（ADR-0008 决策一），每个等待者都持有已编码的请求体，最大可达 `MaxRequestBytes`。排队时的最坏内存约为：

   ```
   MaxAdmissionWaiters × MaxRequestBytes
   ```

   这个值须在实例内存预算之内。

### 建议起点

| 场景 | AdmissionWait | MaxAdmissionWaiters | 说明 |
| --- | --- | --- | --- |
| 本地单用户 | 几秒 | 与 `MaxConcurrentProcess` 相同 | 请求少，排一下体验更顺 |
| 云端交互式 | 0 到 2 秒 | 按公式推算，通常是个位数 | 让客户端尽快收到拒绝 |
| 云端批处理/后台任务 | 0 | 不起作用，但仍须填正数 | 由任务队列负责排队与重试 |

云端宿主拿到 `admission_denied` 时，应向自己的客户端返回 429 与 Retry-After，让排队发生在 barness-ai 之外（客户端或任务队列）。内置等待只用来消化短暂的抖动。

### 验证方式

做一个突发场景：同时发起多于“并发上限 + 等待上限”的调用，记录：

- 立即被拒的比例；
- 等待后拿到许可的比例；
- 等待后仍超时的比例；
- 内存峰值。

“等待后仍超时”的比例高，说明 `MaxAdmissionWaiters` 或 `AdmissionWait` 设大了。示例配置的数值与说明以这组数据为依据，并随证据包一起保存。

## Comments

**2026-10-02 — from issue 17 (maintainer decision):** call identifiers are globally unique (ADR-0001): the host example must mint RequestIDs that never repeat across tenants or calls, and an injected admission/Observer may key on RequestID or AttemptID alone. BindingID and CredentialID stay unique per tenant only.
