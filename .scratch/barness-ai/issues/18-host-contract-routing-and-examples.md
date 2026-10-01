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
