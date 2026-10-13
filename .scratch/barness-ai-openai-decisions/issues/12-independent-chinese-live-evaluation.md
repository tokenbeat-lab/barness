# 12: 取得 Luna 当前正式目录下的独立中文效果报告

**What to build:** 维护者对同一固定中文任务实际运行 Luna，得到与 Jev 完全独立、基于正式目录的准确率、拒答、延迟和消耗报告；未执行或失败不能变成质量结论。

**Blocked by:** 10: 纳入正式目录并形成第十组合独立支持记录；11: 交付 Host 迁移示例与独立拒答评估流程

**Status:** resolved
**Type:** task
**Source tickets:** 08
**Requirements:** R18, R19
**Acceptance cases:** OD55
**Updated:** 2026-10-13

## Responsibility

负责独立 live 评估执行、当前适用性和真实 evidence bundle；沿用已验证工具，不在此票重写评估引擎。Jev 旧报告按原事实保留，当前发布需其自身适用报告。协作时保留他人改动。

## Failure scenarios first

- fixture、candidate Hash 或 Jev 报告充当正式 Luna 效果。
- 两协议同一进程/错误 key、沿用 Jev threshold 或全拒答被称为高准确率。
- 调用/时间预算超额仍 COMPLETE，漏失败样本，取消继续发送。
- 新目录后修改旧报告日期/Hash，或把协议冒烟当效果评估。

## Acceptance criteria

- [x] 执行前核验工具票已有的拒答、错 route、预算/取消反例和第一方 source；新增失败方式先由 public process 表达。
- [x] Luna 正式 route/model/version/Hash、固定 dataset/config/code Hash 与独立 key/账户别名真实一致；记录实际 ResponseModel 和日期。
- [x] 一进程仅一协议，最多 24 Call/Attempt、零重试、串行、总 600 秒；预算与准确命令可核验。
- [x] 24 样本均有状态；合法 refusal 计入任务分母、可核验完整性，不计正确，不伪造概率或置信。
- [x] 两种准确率、混淆、拒答率、confidence/chosen probability、ECE/Brier、延迟与费用独立呈现，null 指标和未上报用量不被零填。
- [x] 缺 key 形成 NOT_RUN 且不完成；实际失败 FAIL。COMPLETE 仅任务完整性，不设置或降低质量阈值。
- [x] 新包和旧 Jev bundle 核验分别保留，Jev 当前报告适用性明确；过期则由最终发布票收集自己新证据，不改旧包。
- [x] actual captures、消费结果、Attempt、Observer、完整样本、manifest/audit 与重放命令齐备，报告不写入支持矩阵。

## Completion

仅该 route 自己的正式实际完整报告与审计成立才完成本票。效果与协议支持保持独立，不能据小样本保证生产质量。

## Comments

2026-10-12：用户批准 13 张草案并替换旧票。此票按新编号发布；下列原票评论按来源保留，ready-for-agent 不表示已实施或依赖已完成。

### 原票 08

2026-10-12：协议支持与标注任务质量是独立结论，拒答不是协议故障。

### 2026-10-13 — 实际独立执行与交付

正式 route `classifier/openai/openai-decisions/gpt-6-luna`，目录版本 `2026-10-13.1`、
Hash `sha256:faae843ac61fea6e3ae59bb75165dc97fcf6aa409dabfe0a2f7c9dccbba12029`。
源码 `d8442da`、固定 `zh-support-v1`/Luna route config；2026-10-13 10:49:27–10:49:35
（Asia/Shanghai）仅一次真实评估，24 Call / 24 Attempt、串行、零重试、8.157236 秒。
全部实际 ResponseModel 为 gpt-6-luna，24 条 answered，24 正确、0 拒答/失败/未执行。
两种准确率均 1，refusal rate 0，confidence ECE 0.0145833333、chosen probability ECE
0.0116666667、Brier 0.0006916667；延迟 mean/p50/p95 339.870/282.498/376.854 ms。
已报告 input 6238/output 0、usage 全完整，目录估价 $0.0006238；不宣称厂商账单或生产质量。

专用子进程只持有 `.env:OPENAI_KEY` 映射的 Luna key，账户 alias
`local-openai@region-unreported`；真实账户/区域、key 专用性、权限和厂商额度未核实，
明确 unknown。直连超时后按公共 CLI/process 接缝先 RED 后补显式 loopback proxy，保留
官方 HTTPS endpoint/TLS 验证、固定评估引擎和预算。代理失败仍 FAIL，缺 key 实际 NOT_RUN，
不制造拒答或用额外请求调参。执行前原工具 105 PASS，新增代理 13 PASS，专项 race 118 PASS。

Luna actual captures、消费结果、24 Attempt、96 Observer、完整样本、manifest/files/audit
已通过现有公共 verifier 无网络重放；460 个源文件与执行快照一致，原始源码归档保留。
Jev 原包按自己的 recorded snapshot 单独核验、全部字节保持，旧 Hash 当前不适用，
新 Jev 证据仍由工单 13 收集。支持矩阵 before/after SHA 相同，没有合并质量报告。

[完整结果、Hash、重放命令与证据](../chinese-live-evidence/README.md)，
最终检查及两轴审查见同目录 validation.json / REVIEW.md。

最终全量开启冻结 Pi：7,991 PASS / 188 SKIP / 0 FAIL，普通/live-tag vet 通过。
Standards 的一项 launcher 非法代理收据脱敏发现先经公开进程三例 RED，再修复并 GREEN；
复审 Standards 0 / Spec 0。外层交付审计无发现、全部文件 SHA 可重复校验。本票完成自身
正式实际完整报告，不代表十组合发布门禁已通过。
