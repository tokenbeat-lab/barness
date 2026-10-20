# 09 最终验收的 MiniMax 国际站独立续验授权

Status: claimed
Type: task

## Context

[规格 R18 / R20](../spec.md)、[07 最终验收](07-final-acceptance.md)、[验证策略](../../../docs/agents/verification.md)。用户要求修复 ZAI Messages 后固定最终源码，重新取得七条严格旧路线及四份效果的当前证据。最初原生授权只有已关闭的 `current-source-eleven-renewal-1`；不得修改旧 grant、账本、路由 claim 或借其剩余额度启动新进程。

## Required behavior

当前实现范围为固定续验 `provider-sites-final-acceptance-2`。首个最终候选批次1已关闭，实际40 Calls/40 HTTP、已认领52/52、USD1.5整笔预留均不得回退；该批次原生 M3 grant/ledger/nonce 从未发放。schema3须在独立15份原生历史之外验证 canonical `closed-cohort-1/seal.json`、完整原件/源码归档与31未启动claim，并由 grant/human/review 三重绑定seal摘要。旧1显式选择、无selector的旧1账本路径、旧1新增grant/nonce/intent均在Key/HTTP前拒绝。新scope为33进程722/758、USD28.4，跨轮774/810、总预留29.9不增加原USD30。没有新增研究；九服务资格与唯一最终race沿用原审批SHA，新付费scope仍待实际明确授权。仅[规格当前状态](../spec.md#当前状态)记录执行结果。

以已封存首批次后的具体续验2演进原生授权边界。先给出架构与失败方式清单，再在公共 CLI / 独立进程边界建立 RED，最后实现。新轮次只允许本次明确批准的 `standard` 国际 `MiniMax-M3` 三协议；保留正式阶段、注册场景、身份隔离、可信 canonical 存储、执行来源、review / human / plan / binary pins、一次路由 claim、持久账本及有限调用 / Attempt / 输入 / 输出预算。

新轮次必须有独立的新 canonical ID / grant / 账本，并绑定所有已关闭的原始 formal 与前次续验历史，不能覆盖旧 authority 或让改 ID、路径、复制 grant、删除输出、并发进程、部分写入或换 worktree 重置授权。旧轮次原件与失败原样保留；不得仅替换常量并把旧 grant 改成新来源，不增加通用无限轮次入口或旧 ID 兼容 shim。无完整批准、来源有变、历史被改动或已有同路线 claim 时，读取秘密与 HTTP 之前拒绝。

本工单只准备并实现有限授权机制及离线验证，不授予真实请求。实际本轮额度仍须人类批准。首选保持既有每身份 20 Calls / 24 Attempts、总计 60 / 72、单次输出最多 4096、累计请求输出 40000、输入预留 120000 的窄合同；任何扩张需另行说明。

## Ownership / validation

实现者负责 `ai/internal/testkit/minimaxlive/` 续验授权相关文件、直接使用它们的 live / 公共 CLI 测试与必要文档；不修改生产协议、区域候选目录或旧原件。先列失败方式，再按 `implement-spec` / `tdd` 在已由规格约定的公共 Client / CLI 接缝逐片 RED→GREEN。仅运行受影响的普通进程 E2E；记录建议 race 范围但不执行。保留真实命令、数字退出、时间、源码与原生工件。所有代码文件不超过 500 行。

## Acceptance

独立审查确认新轮次真正隔离且旧历史不变；普通公共进程 E2E 与旧拒绝合同通过；不弱化证据消费者，不伪造当前自身支持。该机制进入最终执行来源冻结，实际执行及最终来源审查在 07 汇总。

## Implementation（关闭批次1的历史）

已在独立分支实现 schema 2 固定续期、完整 15 份 canonical 前序绑定、只读 launcher preflight、pre-secret `BeginRenewal`、持久单进程 nonce claim 与闭合的 pending 首笔预留。原正式和首轮续期各自 39/36 未变，旧 ID 活动入口永久拒绝，新上限沿用 60/72、20/24、4096、40000/120000。普通公共进程 RED→GREEN 与受影响消费者通过；未读取真实 Key、未请求厂商、未运行 race。[实现与原件记录](../evidence/ticket09/README.md)、[ADR-0050](../../../docs/adr/0050-final-acceptance-renewal-authority.md)。独立最终审查和冻结仍由 07 汇总，工单在该确认前保持 claimed。

## 固定续验 2 实现

活动原生ID为2，schema3与上述固定封存绑定，原15历史39/36×2和本身份60/72、20/24、4096、40000/120000不变。封存散列采用4096文件/64MiB单文件/512MiB总量的稳定文件流，归档遍历限8192条目/深度32。普通公共进程覆盖旧1显式/隐式入口、grant/nonce/intent残留、三重pin、完整历史、3MiB二进制正向、超限拒绝及并发/linked-worktree。永久回执索引由[本轮开发记录](../../../.evidence/provider-sites/final-acceptance-20261020/round2-dev/index.json)关联；这些离线事实不授予付费权限。实际执行及最终审查仍由07汇总。
