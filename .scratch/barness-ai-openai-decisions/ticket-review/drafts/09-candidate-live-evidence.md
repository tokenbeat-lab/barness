# 09: 取得候选 Decisions 路线自己的有界真实证据

**Status:** needs-triage
**Publication:** draft，待用户批准后转 ready-for-agent；尚未发布到 issues。
**Blocked by:** 07: 纳入冻结 Pi 的 Decisions 差分与有限差异账本；08: 形成分类图片设计负载与可核验容量证据
**Source tickets:** 07
**Requirements:** R12, R16, R17, R19
**Acceptance cases:** OD51, OD54
**Updated:** 2026-10-12

**What to build:** 维护者以自带固定候选目录和公共 Client 验证真实 Decisions 文本及三格式图片，得到有来源、独立预算与审计的报告；凭据不足诚实 NOT_RUN，正式目录此时尚未纳入。

## Responsibility

负责第一方来源 fixture、候选目录、该路线专用 live harness/进程/完整场景与预算证据。离线完整实现、运行时、差分和容量先验收。沿用既有 live 基础设施；候选证明不写正式矩阵 PASS。协作时保留他人改动。

## Failure scenarios first

- harness 依赖内置 Luna，导致候选与正式纳入循环；候选直接改正式目录。
- 使用另一组合 key/报告、endpoint override 或未核对的账户权限。
- 失败/环境重试没预占预算，缺场景记 UNSUPPORTED，线上 refusal 被不断强求。
- 来源使用聊天 context/价格，922K 或 score 支持子集被误称已做服务上限测试。

## Acceptance criteria

- [ ] source/candidate/harness/report/budget 的失败方式先写公共 process E2E，再实现 live 接缝。
- [ ] 固定 classifier/openai/openai-decisions/gpt-6-luna、922000 Pi 分类快照及独立输入 .1/.2 阶梯来源，说明格式/模态/128/2–10 区分来源和本期子集。
- [ ] 官方 endpoint 固定；独立测试 key、账户/区域别名与低权限配置可核验。无凭据 NOT_RUN，错误配置拒绝；厂商错误 FAIL。
- [ ] 每进程最多 12 Logical Call/12 Attempt/24 问题/4 MiB 状态、累计 8 张及 4 MiB decoded 图片，180 秒总、30 秒单调用、Binding 零重试；失败和预留均入账。
- [ ] 自己的文本 choice/score/bool、混合、真 PNG/JPEG/WebP、usage/实际 Model/ID/归属/取消均有完整受控断言与真实捕获。
- [ ] 合法 refusal 用自己的类型化结果记录；不循环制造 refusal/429/504，2xx 从不重试。
- [ ] 用量 completeness 与协议报告完成区分，不为 complete 伪造缺省计数；缓存账单、区域 premium 等未观察项保持未知。
- [ ] 标记 candidate 版本/Hash，包含 source 日期/Hash、captures、结果/Attempt、Observer、预算、manifest/audit 与重放命令；不替代正式 Hash 报告。
- [ ] 有限真实响应精度/格式反例先保存、归因并限定 Decisions 自身规则；涉及 ADR 取舍时明确提出修订，不默改容差。

## Completion

candidate 实际完整通过且审计成功才可放行正式纳入。没有 key 时保留 NOT_RUN、本票未完成；实现和 fixture 通过不构成真实支持证据。
