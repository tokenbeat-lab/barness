# 11 — 国际站身份、美元价格与地区证据切换

Status: claimed
Type: task
Updated: 2026-10-18
Blocked by: 01, 02, 03, 04, 05, 06, 07, 08

## Context

用户明确要求“请使用国际站做接入”。[规格](../spec.md)、[国际站研究](../research-international.md)、[ADR-0042](../../../docs/adr/0042-barness-ai-minimax-international-target.md)、[任务图](../map.md)。只读源码映射在`/tmp/barness-minimax-international-seams.md`，原始国际站来源保存在ticket11证据目录；不得读取.env或调用厂商来完成本票。

## Scope and acceptance

删除尚未激活的中国区运行时Provider，使用ProviderMiniMax/minimax和国际站三精确端点；无别名、fallback或双Provider。三协议既有型号能力须由国际站来源复核。M3三身份采用国际站固定美元价，删除活动FX依赖；官方512k整数歧义明示为保守512000估算策略，正缓存写缺价继续保留已知用量并失败。M3.1缺价、资格和三协议离线必交边界保持。

所有历史中国区捕获、来源、intent、收据和hashchain原样保留。canonical candidate账本仍含已消费2 calls/2 attempts，阶段120/144、Chat/M3/standard预算槽20/24不刷新；精确目标身份与有限预算槽分开。历史收据解码仅供不可变预算数据校验，不能成为当前调用入口、跨区native predecessor或国际站actual/formal支持证明。新国际站launch名称、账户/地区别名明确，独占intent与未知消费fail-closed语义保持；总账汇总不得漏计历史或重复计数。

仅切换oracle包装器到冻结包已有MINIMAX_MODELS，保留唯一pi版本、完整数据和源码pins；M3 Messages实际差分、Chat/Responses路线扩展、M3.1型号扩展独立。保持旧317条决策、六扩展路线和前十四支持身份，替换三条未激活CN占位；总计仍十六路线、十七自身身份，所有原压力和四效果要求保持。

先列失败并写公共Client/必要辅助边界E2E red，再实现green。失败覆盖跨区可信native降级、错provider/endpoint/source/price拒绝、旧收据无损与共享预算、跨区研究不能证明当前、旧intent不覆盖、旧支持合同不缩减。更新当前Host/指南/来源/fixture归属及ADR适用性，历史证据不得重写。保留可重复red/green、普通/race、pi、预算/launcher无Key审计及来源清单。

本票完成后只解决国际站实现/离线范围；票10继续国际站自身真实研究、三候选与正式报告、数据激活、旧十四身份与四效果续期，票09汇合完整门禁与最终双轴审查。源敏感实现和审查全部冻结之前不得读取Key或运行真实调用。
