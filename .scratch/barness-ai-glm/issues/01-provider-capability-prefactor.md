# 01: 拆分厂商协议能力并保持既有调用行为

Status: resolved
Type: task

由 2026-10-17 implement-spec 实施指令发布。

**What to build:** 在准备接入第二个兼容厂商时，Host 对现有 OpenAI、DeepSeek 和 Anthropic 的请求、流、结果与费用行为保持一致；后续 GLM 可以独立选择协议能力，不会被迫继承原厂默认。

Blocked by: None

依据：GLM 接入 spec R05、R22；验收组 A14；User Stories 58、62。

- [x] 先固定现有公共 Client 的请求、终态和结果证据，再完成有界前置重构；使用既有 E2E 接缝，不以新增私有 helper 单测代替。
- [x] 将 Responses 缓存路由键、保留期、亲和头及 reasoning summary 分别表达，使原有 OpenAI/DeepSeek 的默认值和字段 presence 保持不变。
- [x] 为 Messages 建立按 Provider 确定自动生成字段的边界；现有 Anthropic cache/beta/thinking/工具字段和 header 行为保持不变。
- [x] 移除需要泛化的内部 DeepSeek 专名时，同票迁移消费者并删除旧字段/实现；不引入并存 shim、双写或通用能力插件注册中心。
- [x] Provider 协议能力和 Model 硬约束的职责有明确记录；本票不引入 GLM 型号或提前发布支持声明。
- [x] 受影响的 OpenAI、两条 DeepSeek 和原生 Anthropic 公共请求/事件回归通过，适用冻结差分无新增未解释差异。
- [x] 产出可重复的前后请求/事件/Result 对比和运行索引，同步解释能力边界的 ADR；本票可以独立合并且既有检查保持通过。

**独立验收：** 现有 Host 调用的 E2E 对照证明请求和结果保持一致；GLM 后续可分别控制缓存、summary 与 Messages 默认字段。

## Answer

2026-10-17：完成有界重构。Responses 拆开 cache key/retention/affinity/summary，保留 OpenAI 和 DeepSeek 原有字段 presence；Chat 的 `thinkingToggle` 直接替代内部 DeepSeek 专名并迁移构建与 usage 消费者；Messages 仅增加四组已有自动字段的 Provider 配置，保持原厂语义。[ADR-0032](../../../docs/adr/0032-barness-ai-provider-automatic-fields.md) 记录 Provider/Model 边界。

公共 Client 前后各 2,340 案例，10,598 个 artifact 对比零未解释差异；适用冻结差分 726 PASS/零 pending，最终 Messages 配置另有 6 PASS。vet 通过；完整 suite 的 46 个 worktree 缺失既有 ZIP/新文件未纳入 source inventory 的失败已通过定点恢复验证，初次失败原样保留。未读取真实 Key、未跑收费请求、未新增 GLM 路由或发布支持声明。[证据、已解释归一化与重放](../evidence/ticket-01/README.md)。

