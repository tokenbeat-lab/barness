# 七份补采账户地区与费率确认清单

七份范围按三个实际账户组确认即可：MiniMax国际一条、Kimi国内三条、Kimi国际三条。账户别名、地区、端点、型号、SDK及冻结哈希已有记录；别名本身不能证明仍是同一实际账户。当前全部未批准、未签发ID/有效期，下面只确认事实。

## 账户与地区

| 组 | 覆盖路线 | 已固定的账户别名和地区 | 已固定的端点及型号 | 仍需可信Host或你确认 |
| --- | --- | --- | --- | --- |
| MiniMax国际 | minimax-messages | MINIMAX / global | api.minimax.io/anthropic；MiniMax-M3 | 仍为旧13项对应的自有国际standard按量API账户，指定型号权限与可用额度正常；非M Plan/priority |
| Kimi国内 | kimi-cn-chat、kimi-cn-responses、kimi-cn-messages | KIMI / CN | api.moonshot.cn/v1，Messages为/anthropic；kimi-k3 | 三协议各自仍使用其旧九项对应的自有国内标准按量API账户；确认同站三协议共用同一实际账户，否则仅列例外 |
| Kimi国际 | kimi-intl-chat、kimi-intl-responses、kimi-intl-messages | KIMI / GLOBAL | api.moonshot.ai/v1，Messages为/anthropic；kimi-k3 | 三协议各自仍使用其旧九项对应的自有国际标准按量API账户；确认同站三协议共用同一实际账户，否则仅列例外 |

tenant固定live-smoke；MiniMax账户scope为live-minimax-messages，Kimi为regional:KIMI，并继续由各自Provider/站点隔离。Kimi网页订阅、Kimi Code套餐、第三方中转不替代这里的标准API身份。只需给出一致/例外/未知，或已有非敏感Host确认记录引用；不需要API key、token、密码、完整账号或具体余额。

## 待核对费率

下表全部为留存来源中的每百万token标准基准，不是本轮抓取的实时账户报价。低于基准的账户折扣无需提供明细；若标准费率、服务档位或另行收费项已改变，只列变化/未知的账户组，先离线复核估算。

| 组及原币 | 未缓存输入 | 输出（含按本协议归入的推理） | 缓存读 | 缓存写5min / 1h |
| --- | --- | --- | --- | --- |
| Kimi国内 CNY | 20 | 100 | 2 | 20 / 40 |
| Kimi国际 USD | 3 | 15 | 0.30 | 3 / 6 |
| MiniMax国际 USD低阶 | 0.30 | 1.20 | 0.06 | 未验证；单项出现正的独立写量即拒绝资格，不能当免费 |
| MiniMax国际 USD最高已知阶 | 0.60 | 2.40 | 0.12 | 同上 |

Kimi来源为2026-10-22快照；国内USD估算继续沿用ADR-0052已接受的2026-10-16历史汇率1 USD=6.7220 CNY，基础费率十位半偶舍入、1h按已舍入Input×2，不向你重复索要实时汇率或这项既定设计的批准。MiniMax来源为2026-10-18的国际standard快照；不以priority或套餐价替换。

## 七份已固定范围

| 路线 | 唯一场景 | 调用/HTTP | 输入预留/输出 | call/process | 拟议USD上限 |
| --- | --- | --- | --- | --- | --- |
| minimax-messages | thinking-on | 1/1 | 2048/2048 | 120s/180s | 0.01 |
| kimi-cn-chat | xai-origin-history | 1/1 | 4096/4096 | 120s/240s | 0.10 |
| kimi-intl-chat | xai-origin-history | 1/1 | 4096/4096 | 120s/240s | 0.10 |
| kimi-cn-responses | xai-origin-history | 1/1 | 4096/4096 | 120s/240s | 0.10 |
| kimi-intl-responses | xai-origin-history | 1/1 | 4096/4096 | 120s/240s | 0.10 |
| kimi-cn-messages | xai-origin-history | 1/1 | 4096/4096 | 120s/240s | 0.10 |
| kimi-intl-messages | xai-origin-history | 1/1 | 4096/4096 | 120s/240s | 0.10 |

全部零SDK/环境重试；Kimi真实首发等待60s已计入240s。按冻结最坏已知费率，MiniMax预留USD0.006144、Kimi国际每路USD0.086016、Kimi国内按既有舍入估算每路约USD0.085308，均低于各自拟议上限。合计拟议USD0.61、七个独立权限范围，不互借余额，不是厂商账单硬封顶。

失败、取消、超时、限流或原生消费拒绝时停止，不自动重试或提高预算；保留已用权限与全额预留。目录激活、race/压力、其他Provider与恢复暂停轮次都不在七份付费范围内。

## 你最少只需确认两项事实

1. **账户和权限**：三个账户组都仍对应各自旧证据的自有标准按量API账户；Kimi每站三协议共用对应账户；地区、指定型号权限及可用额度无变化。如有例外，仅列组名/路线及非敏感Host引用。
2. **费率口径**：这三组账户的标准费率/服务档位仍适用上述冻结基准，且没有基准未覆盖的额外收费项；不一致或未知时仅列该组。真实账户情况不能用离线快照推定。

可直接回复以下事实确认，或交由可信Host提供同等非敏感记录：

> 账户：MiniMax国际、Kimi国内、Kimi国际均与旧证据一致，均为自有标准API，型号权限/额度可用，Kimi各站三协议同账户。例外/未知：___。
>
> 费率：上述三组基准仍适用，无额外收费项。变化/未知：___。

不用再次抄写模型、端点、SDK、来源/二进制/manifest哈希，或先提供grant ID与到期时间。这两项回复只补事实，不是付费执行批准；将来执行仍须另外明确批准各路线及预算。

原七份提案[v3](../minimax-thinking-supplement/pending-scopes-v3/README.md)、[机器确认清单](../../.evidence/kimi-formal-offline-review/confirmation-checklist-v1.json)和[目录审阅](catalog-review.md)均保留各自边界。若实际账户已换，当前同账户组合前提不成立，应只冻结并重新评估该组最小证据范围；不直接复用旧九/十三项，也不默认重跑所有Provider。
