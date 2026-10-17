---
status: accepted
date: 2026-10-17
---

# 中国区标准 GLM Chat 复用生命周期，型号约束守住最终请求

按 ADR-0032 的 Provider/Model 边界，中国区标准智谱以 `ProviderZhipu=zhipu` 和既有 `openai-completions` 路由进入同一 Client/adapter/SDK 生命周期。Provider 配置选用 system、max_tokens、thinking 开关，省略 store、OpenAI 缓存与未核验的 stream_options 默认；省略 effort 保留服务默认思考。DeepSeek 给 assistant 补空 reasoning_content 的规则独立保留，不传给智谱。明确提供 stream_options 的有界候选对照仍可经安全回调发送，字段未出现在 schema 不被解释为官方拒绝。

型号以独立候选 Catalog 声明能力；GLM-5.3 的 off/minimal/medium/xhigh 为 null，low/high/max 有映射，GLM-5.2 则独立采用官方标准 Chat 的开关和别名映射。新增可省略的 `ModelCompat.RequiresThinking` 表达不可关闭思考，零值不改变既有模型，值随目录复制和 Hash 固定。使用型号 ID 分支会把型号事实藏入 adapter，使用 Provider 全局强制思考则会错误限制 5.2，因此均未采用。

full 显式 canonical 档位在绑定/型号授权后、凭据读取前校验；simple 仍按既有向上优先规则夹取。模型采样、canonical 分档、调用采样与 OnPayload 后只校验实际最终 wire，不重新选档；回调可以修复中间采样值，但不能使非法 full 具名档位悄悄变成另一个档位。最终授权检查先于型号能力检查；无 hook 的路径同样校验。最终 max_tokens 必须为 1 到 min(型号上限, Chat schema 131072) 的整数，max_completion_tokens 形状拒绝；具名 null/0 仍省略，实际 wire 上的 null/0 则拒绝。Host 目录可以收窄输出预算，不能扩大服务 schema 上限。发送前失败为零 Provider 请求与零 Attempt。原 OpenAI/DeepSeek full wire 与 canonical 采样分离的契约不变。

候选数据位于 `ai/internal/testkit/glm`，Chat 两条身份各自固定 text-only、官方 schema 的 131072 输出上限与明确标为 barness 保守值的 1000000 上下文，价格仅为合成 fixture 价格。它们未进入内置目录。Responses/Messages 必须有自己的数据和证据，不从 Chat 自动派生；工具、完整 live、价格与正式支持仍由后续票验收。候选 Host/controlled live 只共享已有 Client 装配，单调用策略固定在 120 秒，后续隔离入口另行执行 spec 的双开关、单组合凭据和 24/32 预算。

[工单 02 的证据索引](../../.scratch/barness-ai-glm/evidence/ticket-02/README.md) 给出逐片 red/green、回归、race、冻结差分与重放命令。删除候选装配和 Provider 配置即可撤回尚未发布的路线；内置目录和既有路由不变。
