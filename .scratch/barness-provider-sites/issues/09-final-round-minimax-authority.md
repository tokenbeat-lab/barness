# 09 最终验收的 MiniMax 国际站独立续验授权

Status: claimed
Type: task

## Context

[规格 R18 / R20](../spec.md)、[07 最终验收](07-final-acceptance.md)、[验证策略](../../../docs/agents/verification.md)。用户要求修复 ZAI Messages 后固定最终源码，重新取得七条严格旧路线及四份效果的当前证据。当前国际 MiniMax 原生授权仅接受已经关闭的 `current-source-eleven-renewal-1`；不得修改旧 grant、账本、路由 claim 或借其剩余额度启动新进程。

## Required behavior

以第二个实际续验用例明确演进原生授权边界。先给出架构与失败方式清单，再在公共 CLI / 独立进程边界建立 RED，最后实现。新轮次只允许本次明确批准的 `standard` 国际 `MiniMax-M3` 三协议；保留正式阶段、注册场景、身份隔离、可信 canonical 存储、执行来源、review / human / plan / binary pins、一次路由 claim、持久账本及有限调用 / Attempt / 输入 / 输出预算。

新轮次必须有独立的新 canonical ID / grant / 账本，并绑定所有已关闭的原始 formal 与前次续验历史，不能覆盖旧 authority 或让改 ID、路径、复制 grant、删除输出、并发进程、部分写入或换 worktree 重置授权。旧轮次原件与失败原样保留；不得仅替换常量并把旧 grant 改成新来源，不增加通用无限轮次入口或旧 ID 兼容 shim。无完整批准、来源有变、历史被改动或已有同路线 claim 时，读取秘密与 HTTP 之前拒绝。

本工单只准备并实现有限授权机制及离线验证，不授予真实请求。实际本轮额度仍须人类批准。首选保持既有每身份 20 Calls / 24 Attempts、总计 60 / 72、单次输出最多 4096、累计请求输出 40000、输入预留 120000 的窄合同；任何扩张需另行说明。

## Ownership / validation

实现者负责 `ai/internal/testkit/minimaxlive/` 续验授权相关文件、直接使用它们的 live / 公共 CLI 测试与必要文档；不修改生产协议、区域候选目录或旧原件。先列失败方式，再按 `implement-spec` / `tdd` 在已由规格约定的公共 Client / CLI 接缝逐片 RED→GREEN。仅运行受影响的普通进程 E2E；记录建议 race 范围但不执行。保留真实命令、数字退出、时间、源码与原生工件。所有代码文件不超过 500 行。

## Acceptance

独立审查确认新轮次真正隔离且旧历史不变；普通公共进程 E2E 与旧拒绝合同通过；不弱化证据消费者，不伪造当前自身支持。该机制进入最终执行来源冻结，实际执行及最终来源审查在 07 汇总。
