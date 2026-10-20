# ZAI Messages 显式思考形状的自身研究

2026-10-20 用户明确批准本轮完整 726 Calls / 762 HTTP、US$30、九服务 standard API / 测试账户资格，以及最终 READY 后一次完整正式 race。授权原文及范围：[authorization.json](../../../.evidence/provider-sites/final-acceptance-20261020/authorization.json)。本研究只使用其中四个已固定请求，不代表 candidate / formal 资格通过。

研究工具先完成独立复审，修复自动 redirect、响应头值秘密留存和首次 claim 目录持久化；最终 helper SHA `962cedadc13563a40b5f9a01c1b63f789dc0d1484e7e10875ebfae2d96231c18`、精确 plan SHA `ca00881ec3b628b88c5429ff1d4a03dd96c4be6933969d11604267c272ac2706`。[独立审查](../../../.evidence/provider-sites/final-acceptance-20261020/plan/revision-3/research-review.md)。

同一官方 Messages URL、同一既有 Host 工具 prompt/schema、max_tokens4096，两个自身型号分别对照省略 thinking 与 `thinking={type:enabled,budget_tokens:1024}`。实际 **4 Calls / 4 HTTP、零重试、20 秒内完成，进程 exit 0**，四次均 HTTP200、一个 message_stop、一个 tool_use，原始请求/响应及每次回执完整保留。

| 自身型号 / 请求 | thinking 字符 | signature 字符 | 原生终态 |
| --- | ---: | ---: | --- |
| GLM-5.2 / omitted | 0 | 0 | message_stop |
| GLM-5.2 / enabled-budget1024 | 97 | 24 | message_stop |
| GLM-5.3 / omitted | 102 | 24 | message_stop |
| GLM-5.3 / enabled-budget1024 | 99 | 24 | message_stop |

本轮证实 ZAI 自身精确 enabled/budget1024/max4096 请求被接受，并使 GLM-5.2 这个工具回合返回原生 thinking/signature；没有证明任意预算、预算被服务端严格执行、canonical effort 映射或默认必然行为。GLM-5.3 对照同样保留原生块。历史九次 GLM-5.2 无 thinking 的真实捕获及原资格 FAIL 保持不变，不能通过修改解析器凭空补字段。

下一步先在公共 Client 建立 RED，开放服务内部经过本次自身证实的窄形状，再由 GLM-5.2 独立 ModelCompat 约束 opt-in。无选项仍省略，GLM-5.3 保留已验证默认，不推导 canonical 档位。候选数据、快照/pin、真实采集的 full 请求及普通/负向边界在本轮首次 candidate 前一并完成、复审、冻结。完整 native-history 续接仍须新 candidate 与晋升后 formal 实际证明，研究单回合不能替代。

[原始命令回执](../../../.evidence/provider-sites/final-acceptance-20261020/research-command.receipt.json) / [逐次观测](../../../.evidence/provider-sites/final-acceptance-20261020/research/summary.json) / [原件Hash与Key审计](../../../.evidence/provider-sites/final-acceptance-20261020/research-index.json)。源 helper 前后相同，14 个研究文件的选定 Key 字节扫描零发现，不保存秘密或其 Hash。研究在本轮美元账务中整笔占用 US$0.20 保守预留，未将返回用量当作最终账单。
