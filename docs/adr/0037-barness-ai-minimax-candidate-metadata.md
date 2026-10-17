---
status: accepted
date: 2026-10-18
---

# MiniMax 候选身份与默认采样标签

Provider 新增 `minimax-cn`，API 仍为既有三协议。六个候选条目在测试装配中各自声明，
不运行时从 Chat 条目派生另外两条；文本子集、上下文和输出保守上限独立于厂商更大能力。
在各自真实证据和可核验价格齐备前，不进入 BuiltinCatalog。

M3 只有 off/high 两个 canonical 能力，high 表示开关开启而非深度。M3.1 复用已存在的
`RequiresThinking` 表达必选思考，并声明 low/medium/high/xhigh/max；协议实现负责在
具名选项、采样和最终 payload 边界拒绝非法关闭，不能仅凭 map 夹取消费掉显式关闭。
Provider 的实际 wire 形状仍归协议能力，型号元数据不能开启未经实现的服务功能。

新增 `Model.DefaultThinkingLevel` 只为省略 reasoning 选择采样标签：M3 Chat high，
M3 Responses/Messages off，M3.1 全为 max。它不物化为 wire effort 或用户显式选择。
零值保留已有 off→clamp 行为；非零必须为 canonical 且型号支持，mandatory 不得声明 off。
标签作为值字段参与 JSON、Hash 和 Catalog/Client 快照，无新的可变共享状态。
共享采样只在输入为空时采用标签，回调之后不重选；Messages 继续忽略采样 map。

没有选择通过协议默认强行写入用户 Reasoning，因为它混淆省略与显式设置，也会污染续接
和采样的归属。没有添加二元思考兼容开关，因为已有支持档位 map 已足以表达真实集合。

候选价格全部是明确标记的合成测试费率，版本与显示名不允许它冒充 M Plan/标准报价。
真实 M3 价格及其 RMB→USD 来源在后续独立候选/正式快照取得，不用零值假装免费。
共享元数据验收通过已有公共 Catalog/NewClient 与 Client E2E；这不证明 MiniMax 协议
或线上可用性。[工单 01](../../.scratch/barness-ai-minimax/issues/01-provider-model-contract.md)
保留失败清单和可重放证据；协议与 live 证据由各自后续工单承担。
