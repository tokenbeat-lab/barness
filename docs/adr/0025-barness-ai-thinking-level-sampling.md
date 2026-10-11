---
status: proposed
date: 2026-10-11
---

# 模型分档采样按 canonical level 解析，沿用 pi 的入口差异

pi 1.1.0 已提供模型级分档采样默认，而 barness-ai 当前的 1.0.0 基线仅有通用默认；本方案将配置归属 Chat Model，在已有采样请求边界按模型支持的 canonical level 选择，并以单一、经发布件等价核验的 1.1.0 oracle 验收。采样夹取与 full 的 wire reasoning 分开保留，避免以“实际强度”之名静默改变 pi 的 full 契约；simple 的既有夹取行为继续有效。

`proposed` 保留为整版尚未生效的状态。公共采样功能、单一 1.1.0 oracle 的来源与差分已实施核验；完整当前发布受外部 live/effect 失败阻断。B01–B04 全部通过后再记录正式生效，保留历史证据。

## Considered Options

- **按请求原始档或厂商 wire 值选参**：无法表达 unsupported 档的 clamp，也丢失多个 canonical 档映射同一 wire 值时的不同默认。
- **把 full/simple 都改成最终 wire 强度驱动采样**：直观，但 full low 的采样可夹取为 high、summary-only wire medium 的采样也可夹取为其他档；统一会改写上游契约，回调覆写还会引入反馈循环。
- **保留 1.0.0 并长期把新采样记为扩展，或增加第二套 oracle**：减少即时迁移成本，但同一公开语义有两个兼容基准，难以解释目录、差分和发布合同。
- **采用 canonical 选档并迁移唯一 1.1.0 基线**：沿用 pi 的有效档位定义和既有单基线原则，代价是发布件来源证明、全量差分、目录与正式发布证据都要同步。

## Consequences

- 新 `SamplingParamsByThinkingLevel` 只有模型级字段；canonical keys 为 off/minimal/low/medium/high/xhigh/max，调用级仍用 SamplingParams，SimpleOptions.Reasoning 不新增 off 输入。
- 具名字段 → 通用默认 → 夹取档默认 → 调用覆盖 → OnPayload，顶层浅合并保留 null/0/false。Go 在请求边界应用一次，避免把模型默认物化到 simple 选项；用公共 E2E 验证与 pi 两次内部合并的外部等价。
- Full wire 与 sampler 的分离、off=null 的省略 effort、summary-only 默认 medium、回调之后不 reselection 都是需要记录的兼容边界，不能以后当作无上下文 bug 自动“修正”。不声称已获知厂商真实推理强度。
- 只有已实现的 OpenAI/DeepSeek × Chat/Responses 消费新 map；原生 Anthropic/Gemini 忽略。DeepSeek Responses 仍按 ADR-0014 是扩展路线，不获得原生 pi 差分声明，不新增 Azure 或 Provider 注册。
- 新 map 全档结构校验；消费协议沿用保留键。Catalog 各边界深拷贝 map 与 JSON bytes；Hash/Version、重试冻结和授权沿用 ADR-0003/0005/0006，不能复制 JS 的嵌套引用共享到 Go。
- 迁移必须遵守 ADR-0004 的发布件等价证明与分类账、ADR-0018 的型号硬约束、ADR-0017 的完整发布证据。核验前不假定发布模型默认的内容，不把切换 oracle 版本当作完成兼容迁移。
- 回滚恢复完整先前版本和配套配置/证据合同，不增内部 shim 或双写。本次实施已同步主规范、生产代码、目录和 frozen oracle；完整发布结果独立记录。

[目标规范](../../.scratch/barness-ai-thinking-sampling/spec.md)、[详细设计与实施顺序](../../.scratch/barness-ai-thinking-sampling/design.md)、[验收合同](../../.scratch/barness-ai-thinking-sampling/acceptance.md)、[研究第 9 节](../../.scratch/pi-1.1.0-thinking-sampling/research.md#9-barness-ai-接入映射与建议尚未实现)。用户已确认公共 Client E2E 边界；公共请求矩阵、隔离与差分已验证，实际状态见专项证据索引。

## durationMs

用户于 2026-10-11 明确选择“补齐 durationMs，批准耗时数值差异”。为迁移 1.1.0 聊天终态的公开字段，使用 Nullable[int64] 区分尚未结算与已完成的 0ms；在唯一 assembler 终态边界记录单调 elapsed，包含 Host 准备、回调及初始请求重试，四舍五入到毫秒。结果与 done/error 事件使用同一值。

两次独立运行不可能要求 elapsed 数值相等，因此仅有限 case 的 result.durationMs、terminal durationMs、partial durationMs 数值登记 extension，完整比较仍保留这些字段。每次差分先断言最终值是非负整数、唯一终态与结果一致；终态缺失不获豁免。非终态 partial 的字段出现时机属于既有可前进共享视图契约，有限 case 仅豁免该字段的 presence，避免对 PartialView 作不真实的不可变快照承诺。未增加 Observer 正文或采样日志。

[专项实际证据](../../.scratch/barness-ai-thinking-sampling/evidence.md)分别记录离线、差分、race、live、独立效果与完整发布。整版失败仍保持本 ADR proposed，不新增用户审批要求。
