# 07: 原生状态封套与 reasoning 同模型回放（E03 原生状态部分）

**What to build:** 应用开发者把带签名/密文的推理状态连同来源封套保存，宿主鉴权后通过 Go 构造入口（如 `TrustNativeState(scope, envelope)`）取得可信值并回放给同 Provider/API/model；无可信封套或账户不匹配时自动降级、调用照常成功（spec I6 后三段、User Stories 18–20）。

**Blocked by:** 06

**Status:** ready-for-agent

- [ ] Responses reasoning 在同模型回放中完整保留适用的加密内容/签名与 redacted 状态
- [ ] 原生状态封套记录 TenantID、AccountScopeID、Provider/API/Model；可信值只能经 Go 构造入口获得，JSON 反序列化（含 `trusted=true`、自报 TenantID、签名字段）无法得到
- [ ] 库只核对封套来源与当前 scope、binding 匹配；同账户换 key 不因凭据版本不同而丢弃状态
- [ ] 无可信封套、或同 Provider/API/model 但 AccountScopeID 不一致时，按跨模型规则降级且不返回 invalid_request
- [ ] 降级计数与原因（无封套 / 账户不匹配 / 跨模型）写入 Result 调用元数据（Observer 部分见 14），不进入 pi 兼容消息、不参与 pi 差分
- [ ] 回放后的实际 wire 字段与 pi 差分一致
