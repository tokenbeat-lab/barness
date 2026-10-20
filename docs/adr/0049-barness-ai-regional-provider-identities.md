---
status: accepted
date: 2026-10-20
---

# 中国站与国际站使用独立 Provider 身份

GLM 与 MiniMax 的地区市场需要独立维护型号、能力、默认值、价格和账户授权。采用已有 Provider 身份承载独立服务契约，保留 `ProviderZhipu=zhipu` 表示智谱中国站、`ProviderMiniMax=minimax` 表示 MiniMax 国际站，新增 `ProviderZAI=zai` 与 `ProviderMiniMaxCN=minimax-cn`。型号身份继续为 `(Operation, Provider, API, ID)`，同名型号独立配置；不增加 Site 维度，也不将地区 ID 归一化为品牌 ID。

API adapter、SDK 与调用生命周期继续共享。每种协议以内部不可变配置明确选择已核验的服务能力和 codec；共享实现不得替代精确 Provider 的目录、授权、来源、价格或支持证明。Host 固定端点、账户、凭据来源与使用资格，不自动换站或换 key；标准 API 与套餐用途分别声明。

缓存与亲和标识的现有派生作用域加入 Provider，统一切换 `barness-ai/prompt-cache-key/v2`，输入为长度前缀编码的 Tenant、Provider、AccountScope 与 Session。这样即使账户别名相同也不会产生相同的跨服务路由键。代价是一次缓存命中率变化；不保留 v1 shim 或双发。NativeStateEnvelope 保持现有字段，继续按真实 Provider 精确比较，跨地区只走现有安全降级。

AdmissionRequest 同时增加实际 Provider，供 Host 在账户级准入中区分不同市场的同名账户；不改变库内租户 / 进程限额，也不新增配额系统。回滚先禁用新增服务并移除旧进程目录中的新身份，避免旧代码把新 Provider 作为通用兼容服务继续调用。

此决定沿用 ADR-0001、ADR-0020、ADR-0032，扩展 [ADR-0042](0042-barness-ai-minimax-international-target.md) 当前只接入 MiniMax 国际站的范围；不撤销其独立地区证明、历史消费、禁止回退和原件留存要求。新增中国站是新的正式接入目标，不能恢复旧失败捕获的支持资格。架构决定已选定，实施与发布资格按 [spec 当前状态](../../.scratch/barness-provider-sites/spec.md) 记录。

## Considered Options

仅在 Binding 增加站点能切换连接，但无法为同一完整型号身份配置不同元数据。让 Site 进入目录、授权、原生历史与观测也可满足需求，但增加贯穿全链路的身份维度；现有 Provider 已表达独立地区服务，因此选择独立 Provider。不同地区复制 adapter 会重复资源、重试与流生命周期，也不采用。

[研究](../../.scratch/barness-provider-sites/research.md)、[设计](../../.scratch/barness-provider-sites/design.md)、[规格](../../.scratch/barness-provider-sites/spec.md)。

## ZAI Messages 的独立窄思考形状

2026-10-20 的自身 standard API 四请求对照中，GLM-5.2 的省略形状没有原生 thinking/signature，精确 `enabled/budget1024/max4096` 工具回合返回完整原生块；GLM-5.3 两形状均返回原生块。这为 ZAI 自己的内部 Messages profile 建立 opt-in，并只为自身 GLM-5.2 候选设置 `MessagesThinkingBudgetTokens=1024`，不借用中国站证据。Provider 限定精确形状，Model 只能收紧；其他预算、输出配对及额外 thinking 字段仍被拒绝。

无选项仍省略，GLM-5.3 保留已证默认。单回合研究不证明预算被服务端强制执行，也不证明 canonical intensity 或实际 native-history 续接。Simple 不推导该形状；原始 SSE 在公共 Client 离线重放保持原字段，同源第二轮为明确标注的受控响应。候选资格和晋升后 formal 仍分别需要新来源、新授权的一次完整真实过程，原失败不改写。[研究与原件](../../.scratch/barness-provider-sites/evidence/zai-thinking-research.md)。
