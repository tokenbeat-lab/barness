---
status: accepted
date: 2026-10-26
---

# xAI Responses：显式无状态、可信opaque历史与可核对的输出用量

Grok的Responses使用OpenAI兼容外形，但默认存储会话、默认加密输出与显式include的存储语义不同，官方usage示例又未清楚表达output是否包含reasoning。仅替换endpoint或复制DeepSeek的能力值，会让同一公共接口丢失原生历史或少计消耗。[研究](../../.scratch/barness-ai-grok/research.md)记录了官方事实及歧义。

本设计沿ADR-0014/0032新增独立xai Provider，只复用既有Responses adapter和调用生命周期；首期限定全球标准API、grok-4.7、文本/内嵌图片和函数工具。完整Transcript自包含，每次必须store:false并显式include加密推理，最终采样与回调后再次检查字段存在性；不以4.7默认返回密文推导同等留存行为，不用服务端response ID补历史。

原生reasoning按完整JSON语义保存并在可信tenant/account/provider/API/model下回放，不将密文投影为可见摘要。工具call_id与item ID采用本地版本化opaque二元组，避免继承OpenAI的分隔符、fc_前缀及长度假设；跨来源仍沿既有降级语义，保留可见内容和调用/结果关联。此表示可能进入Host持久化，因此格式与来源规则需要独立记录。

对本次无托管工具子集，规范全部输出B以同一报告的total−input确定；完整wire output O与reasoning R须满足B=O或B=O+R，且R不大于B。此为设计选择，须用本组合真实非零推理捕获确认，不是对厂商歧义的事实裁定。数字缺失保持partial；费用沿既有Attempt.CostEstimate独立表达，缓存分配未知但总量已知时算有限区间，不填造命中计数。长上下文按总输入含缓存≥200000整请求换价，不能借用OpenAI等级倍率或地区费率。

## Considered Options

- 复用OpenAI全部默认：实现最少，但存储、摘要、预算、ID及费用语义不成立。
- 完全复制DeepSeek profile：保留共享adapter却关闭xAI可用且本模块必需的无状态声明和密文请求，未采用。
- 独立Grok adapter或新公共协议：会复制租户/并发/取消生命周期并混淆Provider与API，未采用。
- 永久规定output必然包含或排除reasoning：单一官方示例不足以裁定，且可能双计或少计；选择可核对总量的局部转换及真实资格验证。
- 新建账单/货币/成本接口：已有CostEstimate可以表达xAI场景的不确定性，首期不扩公共形状。

## Consequences

Host必须完整保存消息与Native State Envelope，持久化读回后担保来源，并消费成本完整性。max_output_tokens不构成总费用上限；store:false不构成团队ZDR承诺。首期不接入其他协议、地区、型号、服务等级或托管工具；自带Catalog与Hook不能扩大该子集。

本决定已随2026-10-26的implement-spec请求采用：候选身份、无状态请求、opaque历史和用量/成本转换已进入集成分支。accepted记录的是架构取舍被采用，不表示真实资格、正式目录或发布已经通过，也不授权付费或人工验证阶段。交付契约见[spec](../../.scratch/barness-ai-grok/spec.md)，转换、验收和实施顺序见[design](../../.scratch/barness-ai-grok/design.md)；唯一当前状态只维护在[spec](../../.scratch/barness-ai-grok/spec.md#唯一当前状态)。

## 最终源码接线与正式发布的边界

维护者于2026-10-26明确确认“采用建议顺序，发布前保持未正式支持”。自身真实候选18/18通过后，先在未发布集成分支完成真实脱敏回放、xAI自身资格器/正式harness和最终Builtin目录接线，审查后冻结同一最终payload；之后才为该版本取得新的正式live授权及独立人工race/压力选择，全部适用门槛通过后接受/发布正式晋升。源码接线不等于已正式支持，不自动合并main、push或发布。

这样排序是因为现有正式消费者要求report绑定当前全Builtin版本/hash；若全部门槛完成后才改目录源码，会改变原生/外层来源并使旧门槛失去适用性。候选原件保留自己的stage/catalog/source/binary，不改写为正式。实现不新增控制器、构建标签或兼容层；发布保持false直到所有要求具备事实。详见[已接受的冻结顺序](../../.scratch/barness-ai-grok/formal-qualification-plan.md)。
