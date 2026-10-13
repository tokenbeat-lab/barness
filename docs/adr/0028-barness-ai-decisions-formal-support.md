---
status: accepted
date: 2026-10-13
---

# Decisions 正式目录与独立 schema 3 支持记录

工单 09 的候选在自己的目录下 9/9 真实 PASS，83 项完整性 inventory 逐成员核验，
审计零发现。因此按工单 10 纳入 `classifier/openai/openai-decisions/gpt-6-luna`，
正式目录更新为 `2026-10-13.1`，内容 Hash 重新生成；候选版本、Hash、原报告不改写。
922000 是冻结 Pi classifier 元数据，score 2–10 和 PNG/JPEG/WebP 是本期子集。
Decisions guide 输入 $0.10/M 和 Luna >272000 的全请求 2x 阶梯产生 $0.20/M，
不导入同名聊天型号的输出或缓存价格；alias 权重、账单与区域 premium 不由 Hash 固定。

选择一次性迁移检入矩阵到 schema 3：九行身份、状态、能力、时间及 note 中的历史
Hash 精确保留，第十行初始化为九能力 NOT_RUN。运行时拒绝旧 schema merge；既有
九组合 producer 同步使用 schema 3，历史 schema 2 包继续作为原始证据保留。
不用运行时迁移、双写或兼容 shim，因为内部证据格式应有一个明确的当前合同。

Decisions 使用独立 validator、固定 route/model/endpoint、完整 expected 和自己正式
Catalog Hash。候选及 controlled 报告不能取得正式支持。所有必交能力禁止 UNSUPPORTED，
预算以累计预留和逐场景计数交叉校验，零重试；批次任一拒绝时不写文件。
NOT_RUN 保留历史 PASS，FAIL 更新当前能力但保留过去通过时间，因此当前仍会阻止发布。
正式报告只在审计、完整性读取及写入成功后 accepting；清单覆盖最终报告与捕获。

正式 live 沿用候选的 12 calls/12 attempt reservations/24 题/累计 4 MiB 状态，
8 图/累计 4 MiB decoded 图片，180 秒进程（留 10 秒收尾）/30 秒调用和零 Binding 重试。
用户在 09 已授权 `.env:OPENAI_KEY`、loopback HTTP 代理和无法核实时标记 unknown；
本票沿用授权且记录真实性质，不捏造账户、区域、权限或专用凭据验证。

新 Hash 与旧九组合历史 Hash 不同；它们不能自动成为当前发布证据。工单 13 收集
适用于新 Hash 的完整重跑，本票只形成新路线自己的支持记录；不宣称十组合可发布。
回滚撤销新增 classifier Binding 或图片授权，不改走 Jev/聊天；历史包与已报告用量保留。
