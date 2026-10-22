# 第 1 节维护工具

这些工具只负责一条已选择命令的回执、或一批已审阅文件的清理。不会发放授权、编排厂商轮次或自动进入下一阶段。当前实施范围没有真实 race、压力、付费 live 或证据清理许可；清理相关进程测试仅使用临时合成仓库。

## 普通回执

在仓库根运行 `python3 .scratch/barness-ai-test-process-review/tools/run.py --feature <slug> --stage e2e -- go test -json ./ai/e2e ./ai -count=1 -timeout=20m`。相关 CLI/消费者另做定向公共进程检查。`--input <仓库相对路径>` 可重复指定实际选定的目录快照、授权等非凭据文件；前后都检查内容哈希。argv 不得放密钥。

每次独占 attempt，原生包在 `bundles/<run-id>/`。`receipt.json` 保存实际 argv/退出/时间、源码文件哈希、选定输入、采集器/平台/可执行文件摘要、所有包 manifest/audit 指针、测试事件及日志哈希。失败、启动失败、超时与缺证据分别记录；成功重跑不覆盖失败。普通阶段清除重负载及凭据环境。人工重负载阶段必须另有明确选择，不能从本次实施授权推断。

## 只读发布 CLI

`go run ./ai/release/cmd/releasegate -record <不可变回执索引.json> -out <新目录>` 只消费已有事实。schema 1 形状以 `ai/internal/testkit/release/record.go` 和 `tests/test_release.py` 的合成公共输入为准。索引包含 feature、currentState 文件指针/哈希、source、selectedInputs、impact，以及 e2e/live/race/pressure。

PASS 阶段必须有 receipt 路径/哈希和来自主包的 requiredCases；其余包的完整库存与审计同时检查；race/压力/live 另有 selection.by=human 与精确 argv。live 每条路线有独立记录及消费者回执，消费者日志放在 attempt 内，不能塞进封存 native bundle。NA 必须有原因且与影响分析一致。Source.Files 是采集器选定的完整 Go/夹具/维护工具来源集合（含被 Git 忽略的源码，排除 Markdown 展示文档与生成缓存；参与实际消费的文档另用 --input 固定）；正式状态只能引用真实回执，不能拿自测的合成 race/live 字段冒充实际执行。

## 工具进程自测

`python3 -m unittest discover -s .scratch/barness-ai-test-process-review/tools/tests -v`。设置 `BARNESS_PROCESS_TEST_ARTIFACT_DIR=<新原件目录>` 可保留每次进程的参数、退出码和输出摘要。普通入口测试先编译再直接执行二进制，拦截所有隐藏负载调用；不能用会重写 PATH 的 `go test` 父进程代替这一安全接缝。

## 清理合同

真实使用前重新读取 plan.md 第 6 节。scan/plan/move/verify/restore/inspect/purge 的批次身份、清单哈希、锁、移动日志、基线和七天观察期都在工具内检查，不能绕过。工具不根据日志大小、名称或年龄直接删除；未知或失败状态保留。旧 grant/claim/账本不是清理对象。此文不构成真实清理授权。
