---
status: accepted
date: 2026-10-17
---

# 按 Provider 决定协议自动字段，型号只收紧自身约束

GLM spec R05 的第二个兼容厂商用例使 ADR-0014 中合并的 Responses 缓存能力不再足够：缓存路由键不代表保留期或 session 亲和头，加密推理也不代表 reasoning summary。保留按 Provider 选择的小型内部协议配置，将这五项分别表达；OpenAI 全部启用，DeepSeek 的 summary 继续接受并照发，其余四项不派生。它们决定请求字段 presence，Model 的 thinking level map、缓存保留期支持和输出预算仍只能描述该型号的约束。Provider 不提供的字段不会因 Host Catalog 的型号声明而出现。

Messages 在相同层级增加内部 Provider 配置，限定为已有自动字段的四组：缓存提示、原厂 thinking 扩展（display、adaptive/managed effort 及其 beta）、eager tool streaming、浏览器访问头。预算型 thinking、SDK 版本头、鉴权与流生命周期仍由协议 adapter 负责；模型 compat 决定原厂扩展中哪些规则适用。现有 Messages 路由保留完整 Anthropic 默认，回调的 beta 替换和字段授权流程不变。未来兼容厂商必须在此明确选择配置并取得自身 wire 证据；本工单不新增 Provider、GLM 型号或支持声明。

Chat 的 `thinkingToggle` 描述 enabled/disabled wire 机制，直接替换内部 DeepSeek 专名，并迁移构建与用量消费者；DeepSeek 原有空 `reasoning_content` 回放规则独立保留。没有旧字段 shim、双写、公共能力开关或插件注册中心。

此决策细化 ADR-0014 的缓存分组及 ADR-0015 的内部命名，继承两者的 Provider/Model 边界，并为 ADR-0011/0019 的原厂 Messages 默认建立明确归属。没有改动这些 ADR 已接受的厂商行为。采用独立 adapter 会复制生命周期；开放 Host 可写能力表会允许目录覆盖厂商事实，两者均未采用。GLM 的型号硬约束、兼容 Messages 回放和发送后验证由后续工单凭独立证据实现。

先执行现有公共 Client E2E 取得请求、事件、Result 和用量基线，再重构并比较同一案例；测试固定在已获确认的 Client 接缝，不增加内部 helper 单测。[工单 01 证据索引](../../.scratch/barness-ai-glm/evidence/ticket-01/README.md) 记录前后清单、归一化范围、重放命令及冻结差分结论。
