# 第 1 节三项复审修复后的普通验收

唯一当前状态见 [status.md](status.md)，实际冻结提交 `3bce7a61cb85662e25faf13137feee723c226503`；精确回执、命令、数字退出与时间见 [verification.json](verification.json)。[三项修复及 RED/GREEN 原件](rereview-fixes.md)与[前轮七项修复](review-fixes.md)分别保留，不改写历史结论。

- 47 组工具进程场景 exit0，完整普通 Go 回执 PASS / exit0、375.15 秒、12,890 通过节点；原生 12,086 PASS / 214 NOT_RUN、audit0，与上一轮完整普通 E2E 的 PASS case 集合相同。
- 1,247 份执行文件与 8,636 份选定历史输入在执行前后相同。外层历史输入由 git ls-files .scratch 枚举并排除本功能目录，通过 runpy 将逐文件 --input 交给原回执工具，避免操作系统 argv 长度限制。并非只凭 HEAD 判断。
- 普通 Go 范围沿用完整 ai/e2e、ai、release、releasegate 和 supportmatrix；共享原生证据读取边界改变，重做完整普通回归，保留原有公共能力和历史资格消费者验证。没有启动压力入口或负载子进程。
- 普通及 live-tag 全 ai vet、live-tag 仅编译均 exit0，没有运行编译出的 live 二进制。工具自测和 Go 父子进程保持 GOENV=off、空 GOFLAGS，并隔离凭据、重负载配置和普通 HOME；宿主 HOME 脱敏继续有效。
- 新增失败回归覆盖相对目录/二进制/无扩展名依赖的传递保留，以及普通 chat 的零调用、预算、身份、捕获、结果、Observer、重复调用与计数篡改。完整合成正例仍只读，不能用作真实支持资格。
- 原生 live CLI 只生成资格结果；公开外层 consume_live.py 实际采集唯一原生子进程的输出与退出，原子落 schema 2 回执，发布判定重验两层证据。原生最终写失败实际退出1；外层终结写失败即使子进程退出0也保留 RUNNING 非通过记录。输出直入原件及符号链接均拒绝。

普通、工具和静态验证分别保留真实命令、数字退出、起止/墙钟、日志、工具身份和来源。只读发布判定随后消费不可变索引；race / 正式压力 NOT_RUN，release.pass=false。付费 live、真实清理和后续计划未执行；此前压力误跑事实继续由执行日志保留。
