# 剩余 Provider 路线离线核验当前状态

2026-10-27：十二条路径的离线核验和最小缺口分析完成；因 TypeSafe 分两个型号版本，实际保存十三份记录。最终普通验收已通过，exit0；付费 live/race/压力均 NOT_RUN，暂停轮次 NOT_RESUMED，release.pass=false。

唯一逐路线结论：[decisions-v2.json](../../.evidence/provider-delivery-offline/decisions-v2.json)。授权范围见 [spec](spec.md)，具体边界及依据见 [影响审阅](impact-review.md)，后续执行顺序见 [最小清单](minimal-followup.md)。

## 新适用性与仍有阻塞的记录

| 原路径/型号 | 原场景 | 本次结果 |
| --- | --- | --- |
| OpenAI Images | 3 PASS | 新标准 schema2 消费 PASS/exit0，APPLICABLE |
| Google Interactions Image | 2 PASS | 新标准消费 PASS/exit0，APPLICABLE |
| OpenAI Decisions | 9 PASS | 新标准消费 PASS/exit0，APPLICABLE |
| TypeSafe Jev 1.13 | 3 PASS | 新标准消费 PASS/exit0，APPLICABLE |
| TypeSafe Jev 1.14 | 12 PASS | 原件/目录完整、当前公共 Classify 回放12/12；完整消费者仍因来源差异拒绝 |
| 智谱 Responses，GLM 5.3/5.2 | 22 PASS | 原件/目录完整、当前原生 wire/Result/Observer 检查22/22；完整消费者仍因来源差异拒绝 |
| Z.AI Responses，GLM 5.3/5.2 | 18 PASS | 原件/目录完整、当前原生场景检查及内存捕获回放18/18；完整消费者仍因来源差异拒绝 |

完整资格的实际命令、数字退出和哈希见 [最终消费索引](../../.evidence/provider-delivery-offline/current-consumption-final-v2/report.json)。初次目录缺失拒绝、恢复后的拒绝和所有诊断均保留。后面三项的“场景事实可复用”不能冒充“完整资格 PASS”；它们先做离线来源/消费者判断，当前没有依据把它们加入付费重测。

保守来源差异分别为 Jev 25、智谱 35、Z.AI 30 项，包含聊天专属代码、Provider 条件分支、注释和实际公共路径；不能统一称为 test-only。具体文件和型号投影见 [projection-replay-v1/report.json](../../.evidence/provider-delivery-offline/projection-replay-v1/report.json)。独立原生拒绝诊断见 [diagnostic-v2](../../.evidence/provider-delivery-offline/diagnostic-v2/run/receipt.json)。本轮未修改该保守检查来强行取得 PASS。

## Kimi 两站六条候选

| 路线 | 可复用的原场景 | 当前缺项 | 完整当前资格 |
| --- | --- | --- | --- |
| kimi-cn-chat | 9/9 | kimi-k3/xai-origin-history | 未通过 |
| kimi-cn-responses | 9/9 | 同上 | 未通过 |
| kimi-cn-messages | 9/9 | 同上 | 未通过 |
| kimi-intl-chat | 9/9 | 同上 | 未通过 |
| kimi-intl-responses | 9/9 | 同上 | 未通过 |
| kimi-intl-messages | 9/9 | 同上 | 未通过 |

结论为 **PARTIALLY_APPLICABLE：原54项事实可复用，新增6项真实场景缺失**。六条各自的 Provider/地区、型号/价格、SDK、原生 source/grant/build hash、producer/候选消费真实退出、grant 字节及两份可信账本均已核验，旧阶段仍为 candidate。当前现有原生逐场景检查全部54项无错误，见 [captured-scenes-v1/report.json](../../.evidence/provider-delivery-offline/captured-scenes-v1/report.json)。该分析探针只验证已记录场景，不产生完整资格或新增测量。

当前合同要求每路线10项，原 grant 固定9调用；缺口不能通过改写原预算、report/expected 或拼接未经验证的结果解决。正式目录仍 inactive，没有执行晋升。最少缺6场景不等于现有完整入口可只调用6次：现有入口六路线最多60调用/60HTTP。若只补6次，先评估并审阅单场景采集与组合消费的最小改动，本轮不新增这种框架。

## 原件追溯与最小实现

四份早期捕获没有包内完整目录，但原固定提交的 ai/release/catalog-snapshot.json 仍在，version/完整 Catalog.Hash 均与原报告一致。当前消费者新增有界、只读的 Git 原件读取，校验固定40位提交、原 version、原 hash 与实际 catalog 内容；并继续核对当前所用型号，不修改原 bundle 或目录时间。九种失败方式先 RED 后 GREEN，包括错误提交、缺原件、损坏 JSON、错 hash/version、工作区变化及当前型号漂移。

全部选中原包、封存库存、逐 case artifacts、断言及 audit 已核验；1162项已枚举来源（path+SHA）均能从当前相同字节或固定 Git blob 定位，缺原字节0。Kimi实际输入与原 producer 回执、六份旧冻结二进制哈希、原候选退出0与12份可信账本均核验，未执行旧二进制。见 [originals-audit-v1.json](../../.evidence/provider-delivery-offline/originals-audit-v1.json)及[授权链](../../.evidence/provider-delivery-offline/authorization-links-v1.json)。

与前次16条独立结论合并后，现有20份路线/型号正式证据条目；这不是20/29条路径完成交付，TypeSafe两个型号版本共享一条路径，Kimi仍候选。旧16条没有新增 live，也不因本轮消费者/测试修改而全部失效。见 [合并索引](../../.evidence/provider-delivery-offline/current-qualification-index-v2.json)。MiniMax国际Messages仍保留原 incomplete/FAIL，本轮不诊断或恢复它。

## 验证与隔离异常

基线/受测 HEAD仍为69ff4af880e68ef3ca245e5128bdae5e59a2ef7b；实现尚未提交。实际受测身份按2001项Source.Files与179项实际输入保存，不以HEAD冒充已提交实现。活动变更只涉及testkit目录消费者及普通测试，产品SDK/adapter/catalog不变。

目录恢复九场景定向PASS；隔离守卫先RED、修复后定向PASS。最终56项工具进程测试与go vet均PASS；最终完整普通PASS/exit0，墙钟591.353秒，主包15173 PASS/214原有opt-in NOT_RUN，辅助GLM32、MiniMax77+6、visualeval25均PASS。来源集合SHA256：`a3595792fd3d41ba23d1d4e740868f234de5b1eb6a8a2f0eff4cb970daf64f91`。见[最终回执](../../.evidence/provider-delivery-offline/e2e/20261027T102623.879318Z-fdn3_tgs/receipt.json)及[验证/保留性核对](../../.evidence/provider-delivery-offline/verification-and-retention-v1.json)。原编译失败、RED、早期普通PASS及隔离问题全部独立保留，见[attempt索引](../../.evidence/provider-delivery-offline/attempts-index-v1.json)。

扩大保留性核对发现第一次普通/工具回归向公共Kimi/地区/xAI存储新增24份合成记录，包括故意不完整的写入；第二轮又发现未初始化私有命名空间的xAI共享夹具新增3份，合计27份。9936项原受保护文件未改，无真实授权消费或厂商请求。新增27项全部保留并索引，不删除、不迁移；见 [异常原件索引](../../.evidence/provider-delivery-offline/synthetic-store-exposure-v2.json)。

已在既有Kimi节奏/中断、地区别名测试及xAI共享原生夹具中复用privateGrantStore，并在领取或写入之前检查实际common Git属于声明的私有命名空间。先保存各入口失败守卫，再修正；最终定向、完整普通及工具套件前后公共存储均保持8156项，集合与每项字节完全相同，没有后续追加。旧25份过程测试隔离异常属于前次历史，仍保留，本轮未冲销。

所有来源差异、捕获解释与资格拒绝分别记录；纯诊断没有绕过正式门禁。没有新付费、race、压力、暂停恢复、权限复活、清理、提交、推送或正式发布。
