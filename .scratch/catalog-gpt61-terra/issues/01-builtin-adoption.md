# GPT-6.1 Terra 内置目录采用

Type: task
Status: resolved

按 [spec](../spec.md) 的内置采用范围，将本地已批准 revision 通过 CLI 导出的
整组 Provider、原运行快照和完整基线一起入库。保留原 118 条定义及相对顺序。
GPT-6.1 Terra 子 agent 编排并执行受影响验证，先保留公共 RED，再验证采用后的
公共 Builtin/Client、推理、图片输入、Usage、价格阶梯、来源完整性与新种子导出。

本次状态与证据统一记录在 [status.md](../status.md)。不扩大发布或付费调用范围。

## Answer

已采用批准 revision 的整组维护源、内置原快照与完整 20 文件基线。
GPT-6.1 Terra 完成 28 项必需原生场景和完整受影响普通回归，PASS/exit0；
生成、构建、vet 与两个独立审查均通过。必要的 ChineseEvaluation v4 目录 pin
同步已验证，旧配置/快照/测量保持。全部失败和取消原件保留。
实际受测提交、原件摘要和有限结论见 [唯一当前状态](../status.md) 的内置采用最终结果。
