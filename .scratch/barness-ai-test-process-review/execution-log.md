HEAD 118b82259d799b580df30d27d10ec43ab8ef9b3e 开始执行 plan.md

2026-10-22（Asia/Shanghai）：用户仅授权第 1 节入口迁移及普通验证。race、压力、付费 live 均不运行；不开始真实证据清理。公共测试接缝已由 plan.md §1.2 明确并获本次执行指令确认：Client、CLI/子进程、临时文件系统。当前无其他相关验证进程；原工作区仅有本功能三份未跟踪规划文档。

本文件记录实际执行，唯一当前状态见 status.md。实现与规范分别提交；历史失败原件不改写。

普通回执工具：RED `.evidence/barness-ai-test-process-review/e2e/bootstrap-20261022T013323Z` exit1；GREEN `.../run-green-20261022T013417Z` exit0，5组进程场景。清理工具：RED `.../cleanup-red-20261022T013521Z` exit1；GREEN `.../cleanup-green-20261022T013636Z` exit0，7组，仅临时合成目录。发布CLI新合同 RED `.../release-red-20261022T013712Z` exit1。

偏差：`ordinary-red-20261022T013759Z` 自测以 `go test` 为父入口，Go工具链重写PATH，拦截器未生效，旧 `TestMiniMaxReleaseLoadConsumer` 意外启动一次 `BARNESS_AI_PRESSURE=1` 的受控 MiniMax 六负载子进程并通过。此非用户授权，已向用户明确报告；不能将本次压力状态写为 NOT_RUN。没有 race 或厂商 live。后续改为仅编译测试二进制，再直接运行带硬编码拦截器的指定用例；拦截必须使旧入口先失败，才能修改入口。原回执保留，不改写成RED失败。

入口隔离接缝纠正后：`ordinary-guarded-red-20261022T013840Z` exit1，拦截器在旧入口先拒绝隐藏负载；`ordinary-green-20261022T014424Z` exit0，新普通入口读取留存夹具，不启动负载。夹具源为历史 `minimax-current-eleven-gate-2/gate/offline/20261019T064845.243220000Z` 六个 MM05 原生案例，逐文件哈希一致，压缩 JSONL 278680 字节；不产生当前容量或真实支持资格。

补充失败场景：`hardening-red-20261022T014510Z`（大小写 JSON 歧义、必需索引依赖缺失），`input-red-20261022T014711Z`（选定输入与工具身份），`multipackage-red-20261022T015031Z`（多包回执），`record-contract-red-20261022T015236Z`（缺工具链或影响标记），`source-inventory-red-20261022T015455Z`（忽略源码与纯展示文档边界），均 exit1，原件保留。对应修复后完整进程集在 `source-checks-20261022T015556Z` exit0，23组。

定向普通 Go 验证：`.evidence/barness-ai-test-process-review/e2e/20261022T014901.047673Z-00u3v4fm/receipt.json` PASS/exit0。维护工具全套、定向 vet、live-tag 只编译在 `implementation-checks-20261022T015346Z` 全部 exit0；未运行该 live 二进制。随后源码枚举补强仅排除展示 Markdown 与生成缓存，包含被 Git 忽略的源码；实际参与消费的 Markdown 必须通过 --input 指定。

删除原冻结差分、旧自动 check/race 调度及旧固定十门禁/全库选择器。历史原件未动；通用序列化投影与 SHA-256 改为中立测试语义，保留原生完整性、脱敏及资格消费者。旧固定发布断言由五项回执 CLI 进程验证替代，普通行为与消费者安全检查继续保留。

首轮完整普通回归（785f123）原件：`.evidence/barness-ai-test-process-review/e2e/20261022T015737.105482Z-78u86_0y/`。实际发现 Decisions/Jev 历史目录比较未覆盖后续地区身份、旧 Host 回滚测试使用当前已晋升目录、promotion 边界测试把已激活文件当 inactive fixture。产品目录与这些旧测试的基线文件本次未改动，故按真实故障修复场景边界，不刷新历史 hash 或资格。

同轮暴露额外未 opt-in 的旧 AdmissionBurstPressure、PressureFixtureWriteBackpressure、PressureToolWriteIntegrity；它们已实际执行，属于本次禁止范围的偏差，已告知用户。后续将三个入口纳入显式压力开关；单次小响应的诊断场景以普通 loopback 故障语义保留。不得把这轮写为“零压力”或压力验收通过。

为避免在已知失败后继续运行，于本轮中断。进程组重复清理在 Go 进程退出后返回 EPERM，采集器退出1且未正常落最终 receipt；保留原始 RUNNING/局部文件，并另附 interruption.json（native exit未知，不伪造）。已核对相关自有 Go/采集进程均退出。该采集器故障也需修复并做公共取消回归。

修复验收：`boundary-fixes-20261022T020649Z` exit0，25组维护工具公共进程场景；取消真实小型 Go 子进程得到 CANCELLED 数字退出回执，手动压力入口默认全部 NOT_RUN。定向历史目录/回滚/晋升/自身快照回归 `.evidence/barness-ai-test-process-review/e2e/20261022T020732.756216Z-_jpbzjmu/receipt.json` PASS/exit0。inactive fixture取自 Git 2f19d45，SHA256 7923e919d03e274d940a164da65d747ef06247a74fbc99d442f93996c004787a，与原晋升边界常量一致；未改生产启用文件或放松 guard。

整包普通回归（冻结来源 e1a336de0e7d9249c3bde9ed44551f502555fa90）：`20261022T020749.741275Z-ns_hsib6/receipt.json` PASS/exit0，12811个实际通过测试节点，墙钟375.59秒；原生库存与 audit 有效。`final-static-20261022T020758Z` 普通/live-tag两种全 ai vet与live-tag只编译均exit0。

复审发现live回执环境合同缺少通用路线及MiniMax授权字段、GLM/MiniMax短路由映射和Jev版本点号。先在`live-contract-red-20261022T021445Z`留下Python合成子进程失败，修复后`live-contract-green-20261022T021505Z`全套26组进程验证exit0。此测试只有临时合成原件和假凭据，不执行Go live生产者或厂商请求。随后再次冻结实现，最终普通回执显式固定仓库内外层历史输入；真实证据清理仍未开始。

最终冻结普通回执：`.evidence/barness-ai-test-process-review/e2e/20261022T021552.728508Z-37lnpja_/receipt.json` PASS/exit0，受测提交 `0bbb33896c2b11d13c870aa9bbf9738e9bcaec83`，实际通过12811测试节点，原生12086 PASS / 214 NOT_RUN，audit零发现，墙钟372.9秒。完整来源1240文件及8636份外层历史输入前后哈希相等。普通通过的有限case集合与上一轮整包通过集合完全相同；没有以删除/跳过原普通用例缩小范围。

外层输入枚举方式：`git ls-files .scratch`，排除本功能 `.scratch/barness-ai-test-process-review/`；为每个文件构造独立 `--input <path>` 参数，再通过同一 run.py main（Python runpy避免操作系统argv长度限制）执行回执所记原生Go argv。原件中的 selectedInputs 保存全部精确路径及哈希；不扫描或变更真实 `.evidence`。最终静态检查 `frozen-static-20261022T021650Z` 两种vet和live-tag只编译均exit0。

收尾只读判定 `record-check-20261022T022247Z-w_a2gs3l` 实际退出1；source/live适用性通过，E2E因attempt根目录的`.DS_Store`在回执封存后被macOS更新而拒绝，race/压力未执行。原receipt和拒绝报告均保留，不改旧哈希。`desktop-metadata-red-20261022T022346Z`先证明该问题；随后明确区分执行来源/命令文件与桌面缓存，Go/Python来源枚举一致排除桌面缓存，回执只登记自身生成的四份命令/日志文件。再次冻结与普通验收后才提交规范。

最终冻结普通回执：`.evidence/barness-ai-test-process-review/e2e/20261022T022434.220792Z-qopwtw85/receipt.json` PASS/exit0，受测提交 `ad2ca22aad39cfc041319204687b9bd5ba6acf0b`，实际通过12811测试节点，原生12086 PASS / 214 NOT_RUN，audit零发现，墙钟370.57秒。完整来源1240文件及8636份外层历史输入前后哈希相等。普通通过的有限case集合与上一轮整包通过集合完全相同；没有以删除/跳过原普通用例缩小范围。

外层输入枚举方式：`git ls-files .scratch`，排除本功能 `.scratch/barness-ai-test-process-review/`；为每个文件构造独立 `--input <path>` 参数，再通过同一 run.py main（Python runpy避免操作系统argv长度限制）执行回执所记原生Go argv。原件中的 selectedInputs 保存全部精确路径及哈希；不扫描或变更真实 `.evidence`。最终静态检查 `metadata-static-20261022T022458Z` 两种vet和live-tag只编译均exit0。

桌面元数据修复 `ad2ca22aad39cfc041319204687b9bd5ba6acf0b` 冻结后，`desktop-metadata-green-20261022T022359Z` 27组进程场景通过；最终普通原件 `20261022T022434.220792Z-qopwtw85` PASS/exit0，370.57秒，12811个通过节点，原生12086 PASS/214 NOT_RUN、audit0；四份受控命令文件不登记桌面缓存。`metadata-static-20261022T022458Z` 双vet和live-tag仅编译均exit0。旧桌面缓存哈希失效的回执和拒绝报告原样保留。
