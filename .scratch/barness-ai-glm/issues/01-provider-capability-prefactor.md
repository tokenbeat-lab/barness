# 01: 拆分厂商协议能力并保持既有调用行为

Status: ready-for-agent
Type: task

由 2026-10-17 implement-spec 实施指令发布。

**What to build:** 在准备接入第二个兼容厂商时，Host 对现有 OpenAI、DeepSeek 和 Anthropic 的请求、流、结果与费用行为保持一致；后续 GLM 可以独立选择协议能力，不会被迫继承原厂默认。

Blocked by: None

依据：GLM 接入 spec R05、R22；验收组 A14；User Stories 58、62。

- [ ] 先固定现有公共 Client 的请求、终态和结果证据，再完成有界前置重构；使用既有 E2E 接缝，不以新增私有 helper 单测代替。
- [ ] 将 Responses 缓存路由键、保留期、亲和头及 reasoning summary 分别表达，使原有 OpenAI/DeepSeek 的默认值和字段 presence 保持不变。
- [ ] 为 Messages 建立按 Provider 确定自动生成字段的边界；现有 Anthropic cache/beta/thinking/工具字段和 header 行为保持不变。
- [ ] 移除需要泛化的内部 DeepSeek 专名时，同票迁移消费者并删除旧字段/实现；不引入并存 shim、双写或通用能力插件注册中心。
- [ ] Provider 协议能力和 Model 硬约束的职责有明确记录；本票不引入 GLM 型号或提前发布支持声明。
- [ ] 受影响的 OpenAI、两条 DeepSeek 和原生 Anthropic 公共请求/事件回归通过，适用冻结差分无新增未解释差异。
- [ ] 产出可重复的前后请求/事件/Result 对比和运行索引，同步解释能力边界的 ADR；本票可以独立合并且既有检查保持通过。

**独立验收：** 现有 Host 调用的 E2E 对照证明请求和结果保持一致；GLM 后续可分别控制缓存、summary 与 Messages 默认字段。

