# 第 1 节维护工具

这些工具只负责一条已选择命令的回执、或一批已审阅文件的清理。不会发放授权、编排厂商轮次或自动进入下一阶段。当前实施范围没有真实 race、压力、付费 live 或证据清理许可；清理相关进程测试仅使用临时合成仓库。

## 普通回执

在仓库根运行 `python3 .scratch/barness-ai-test-process-review/tools/run.py --feature <slug> --stage e2e -- go test -json ./ai/e2e ./ai -count=1 -timeout=20m`。相关 CLI/消费者另做定向公共进程检查。`--input <仓库相对路径>` 可重复指定实际选定的目录快照、授权等非凭据文件；前后都检查内容哈希。argv 不得放密钥。

每次独占 attempt，原生包在 `bundles/<run-id>/`。`receipt.json` 保存实际 argv/退出/时间、源码文件哈希、选定输入、采集器/平台/可执行文件摘要、所有包 manifest/audit 指针、测试事件及日志哈希。失败、启动失败、超时与缺证据分别记录；成功重跑不覆盖失败。普通阶段清除重负载及凭据环境，固定 `GOENV=off` / 空 `GOFLAGS` 并在 toolchain 中登记；只读发布判定也强制检查这项政策。非 live 阶段使用独立 HOME，保留原 GOCACHE/GOPATH，防止重建环境的旧子进程重新读取用户持久 Go 配置。live 保留原 HOME 以维持一次性授权及账本边界，仍固定 GOENV/GOFLAGS。人工重负载阶段必须另有明确选择，不能从本次实施授权推断。

## 只读发布 CLI

`go run ./ai/release/cmd/releasegate -record <不可变回执索引.json> -out <新目录>` 只消费已有事实。schema 1 形状以 `ai/internal/testkit/release/record.go` 和 `tests/test_release.py` 的合成公共输入为准。索引包含 feature、currentState 文件指针/哈希、source、selectedInputs、impact，以及 e2e/live/race/pressure。

PASS 阶段必须有 receipt 路径/哈希和来自主包的 requiredCases；其余包的完整库存与审计同时检查；race/压力/live 另有 selection.by=human 与精确 argv。live 每条路线有独立记录，`consumers` 为恰好一份 `{path, sha256}` 原生消费者回执引用；旧自填 argv/exit/files 格式已删除。对精确原生 bundle 显式运行只读命令（不发厂商请求，不写支持矩阵）：

`releasegate -live-bundle <仓库相对 bundle 目录> -out <attempt 内的新消费者目录>`

它复用各路线原生报告构造器、manifest/audit 和完整 `supportmatrix.Merge` 资格检查，记录 `live-consumer.json`、stdout.jsonl、stderr.log、exit-code，绑定命令、时间、退出、可执行文件摘要、完整来源、全部原生文件哈希和结果摘要。消费者目录必须在仓库内且位于 native bundle 外，拒绝符号链接；输出不得写回原件。发布判定重新执行这些只读消费者规则并比对回执、原件及日志，不能用 `true`、过期来源、controlled 或零调用 PASS 标签代替。消费者不是另一套调度/授权系统，不启动任何验证阶段。

NA 必须有原因且与影响分析一致。Source.Files 是采集器选定的完整 Go/夹具/维护工具来源集合（含被 Git 忽略的源码，排除 Markdown 展示文档与生成缓存；参与实际消费的文档另用 --input 固定）；正式状态只能引用真实回执，不能拿自测的合成 race/live 字段冒充实际执行。

## 工具进程自测

`python3 -m unittest discover -s .scratch/barness-ai-test-process-review/tools/tests -v`。设置 `BARNESS_PROCESS_TEST_ARTIFACT_DIR=<新原件目录>` 可保留每次进程的参数、退出码和输出摘要。普通入口测试先编译再直接执行二进制，拦截所有隐藏负载调用；不能用会重写 PATH 的 `go test` 父进程代替这一安全接缝。

## 清理合同

真实使用前重新读取 plan.md 第 6 节。scan/plan/move/verify/restore/inspect/purge 的批次身份、清单哈希、锁、移动日志、基线和七天观察期都在工具内检查，不能绕过。工具不根据日志大小、名称或年龄直接删除；未知或失败状态保留。旧 grant/claim/账本不是清理对象。此文不构成真实清理授权。

## 原生 live 环境合同

live 只转发一个选定生产者的变量，不授予调用权限。通用路线沿用 `BARNESS_AI_LIVE_COMBO` 与该组合唯一 key；GLM 的 `chat/responses/messages` 映射为 `zhipu-*`，MiniMax 映射为 `minimax-*` 并保留原 ledger、账户/地区、entitlement 和续期字段；地区路线沿用完整 combo 与一次性 grant。Jev 1.14 正式路线使用 `BARNESS_MM_LIVE`、`BARNESS_MM_AUTHORIZATION` 及 `typesafe-classifier-jev-1.14.0`，版本中的点号允许。代理只使用相应原生显式变量。控制/回放开关不转发，多个路线选择拒绝。

本次该合同的验收只启动 Python 合成子进程和假凭据，未调用任何厂商。实施中误触发旧压力入口的偏差及纠正统一见 [当前状态](../status.md) 与 [执行日志](../execution-log.md)。

执行来源排除 `.DS_Store` / `Thumbs.db` 等非执行桌面缓存；命令回执固定登记 command.json、stdout.jsonl、stderr.log、exit-code。原生 bundle 库存与审计保持严格，真实命令文件或选定输入变化仍拒绝。

原生 live manifest 的 package 保留 `-tags live [-timeout <Go duration>] ./ai/live` 格式；采集器按该合同识别唯一主包，仍拒绝错包、附加包和歧义。外层超时同时覆盖进程与输出管道；脱离进程组的后代持有管道时，采集器关闭自身读端、记录 CAPTURE_FAILED 并有界退出，不把它写成 PASS。它不发现或终止任意外部 session；临时用例负责回收自己创建的已知后代。

清理判定必须同时核对所有 receipt、manifest、audit 与声明的文件库存；失败、缺 audit、未分类字节或 grant/claim/ledger 任一项都阻止整候选隔离。未被手工索引的失败和未知候选也作为递归依赖根。状态 JSON 通过同目录临时文件、flush/fsync、原子 replace 与目录同步保存，写入失败不会截断上一份状态；这不授予真实清理权限。

普通 E2E 的 TestMain 同时注册宿主账户 HOME 的脱敏，防止隔离 HOME 后进程报告泄露原宿主路径。消费者成功/拒绝的 stdout、stderr 文件与进程实际输出逐字相等，exit-code 与实际返回一致；不能以磁盘快照冒充标准输出。
