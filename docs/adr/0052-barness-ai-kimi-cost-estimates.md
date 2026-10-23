---
status: accepted
date: 2026-10-22
updated: 2026-10-23
---

# Kimi 在共享协议中接入，可证明的计费不确定性以成本范围表达

Kimi 两个市场的三协议共用 barness-ai 现有 adapter，身份沿 ADR-0049 分为 `moonshot` 与 `moonshot-cn`；首批限定 K3 文本和 131072 输出，逐协议独立目录与自身证明。沿用 ADR-0014/0015/0032 的内部 Provider 能力配置，不复制生命周期或让 Host 通过兼容开关扩大能力。完整设计与实施就绪范围见 [spec](../../.scratch/barness-ai-kimi/spec.md) 和[设计](../../.scratch/barness-ai-kimi/design.md)；本决定已随 2026-10-22 的 implement-spec 实施请求采用，身份与成本接口已进入集成分支；不表示六路线实现、真实资格或发布已完成。

K3 默认有收费缓存写入，实际 TTL 可能沿用既有前缀；Chat 响应没有证明其写量属于 5m 还是 1h 的字段。请求 TTL 只能表达偏好。拒绝所有正写量会使 Chat 难以使用，猜 5m 则会把未知费率伪装为确定值。决定在 Attempt 增加可选的值类型 CostEstimate，区分 resolved、bounded、partial、unreported，保留可证明的 USD 成本范围；token 的 UsageReporting 继续独立表达上报程度。

完整收费计数、价格和实际档位已知时范围收敛为一点；仅实际 TTL 未知时，用本站 5m/1h 写价形成上下界，调用可以成功，Usage.Cost 为明确标注的区间下界。计数缺项若有明确总量约束和已知费率，可计算完整有限区间；无法证明上下界时保持 partial。协议定义的可选默认值与必要报告缺失分开处理，无报告的必要计数零值不表示免费。Host 展示须消费成本状态，预算预留使用保守上界。新字段对旧 Provider 省略，原消息形状、其他 Provider 数字、失败 usage 和不跨 Attempt 求和的规则保持。

这是对 ADR-0010“Attempt 承载报告、Catalog 固定价格”的补充：它没有成本可靠性轴，也没有 Kimi 的收费 TTL 缺项。Kimi 的收费写量仍是用量完整性必要报告项，但缺项可在明确总输入分区内形成有界费用；不改变既有 Provider 的可选写量判定。成本区间由协议所保留的实际证据计算，不把请求值、合成签名或另一市场的事实当成观测。

美元成本合同继续沿用：中国初批取仓库已留存的 2026-10-16 CFETS 历史汇率，原币/换算/来源进入独立价格记录，基础费率十位小数半偶舍入；1h 按现有 Input×2 运算次序估值，记录末位舍入约定。不引入在线换汇或新的货币模型。更新快照必须更换版本并取得适用验证。

## 2026-10-23 修订：可选读取计数与有界写量

用户明确选择保守费用区间，并将 Kimi Chat/Responses 的 cached_tokens 定义为可选、缺失/null 默认0。该默认参与用量完整性与成本计算，不仅是兼容数字字段的占位；Chat 先使用有效的 nested 值，再使用原有 legacy 顶层回退，最后使用默认。显式非法值继续拒绝。Messages 的 cache_read_input_tokens 与其他 Provider 维持自身合同，原始捕获不改写。

对 Chat/Responses，若总输入P、输出O及读取R已知（R可来自上述默认），写量W缺失，则0≤W≤P−R、未缓存输入=P−R−W。固定价格及实际响应的TTL证据确定所有允许分配的最小/最大成本；请求TTL不作为实际证据。两个端点相等为 resolved，否则 bounded；UsageReporting 仍为 partial，不能把未知写量或1h分量改成观测。必要总输入/输出不足以证明完整有限费用时继续 partial/unreported。

这扩展了初版仅允许未知TTL产生 bounded 的限制。初次真实中国站 Chat 的缺 cached_tokens 响应及失败授权永久保留，见 [执行记录](../../.scratch/barness-ai-kimi/evidence/candidate-live-result-01.md)。新规则下，该原始响应在公共 Client 离线回放得到 CacheRead0、UsageComplete、CostResolved；这不是对旧来源真实资格的追认。实现与验收见 [08](../../.scratch/barness-ai-kimi/issues/08-live-usage-reporting.md)。

## Considered Options

- 只发 5m、按请求 5m 估价：不能证明同 org 中先前创建的前缀不是 1h。
- 对所有未知 TTL 正写量返回 protocol 错误：把合法服务响应变成不可用的 Chat 路线。
- 把 UsageReporting 改成 partial：混淆完整 token 报告与未知计费档位，无法给预算消费者可靠上界。
- 对所有 Provider 重构货币和账单模型：超出本次真实差异，增加无关迁移成本。
- 在 Attempt 加有界、可省略的成本范围（采用）：新增一个必要的公共值，保留调用可用性和价格不确定性。

## Consequences

Kimi 的 Host 必须升级成本展示/预算消费；只看原点值会低估未知 1h 写价。旧 Provider 不受该新增轴影响；公共结构的位置式 literal 需迁移。完整成本依据与原用量须在 Observer 入队前同步固定，避免最终 Result 与观测不一致。

首批不增加 DefaultMaxTokens 公共字段，而将支持上限收窄至厂商默认 131072；将来扩大输出时须重新设计默认/能力上限关系和资格范围。其他模态、K2.x、严格采样或托管工具不由本次协议接入自动授权。

验证依 ADR-0051 分阶段，以用户确认的公共 Client E2E 为主；本 ADR 不授权真实调用或昂贵阶段。唯一当前状态、完成与未完成项统一见 [spec](../../.scratch/barness-ai-kimi/spec.md#唯一当前状态)。
