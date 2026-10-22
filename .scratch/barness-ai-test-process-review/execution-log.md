HEAD 118b82259d799b580df30d27d10ec43ab8ef9b3e 开始执行 plan.md

2026-10-22（Asia/Shanghai）：用户仅授权第 1 节入口迁移及普通验证。race、压力、付费 live 均不运行；不开始真实证据清理。公共测试接缝已由 plan.md §1.2 明确并获本次执行指令确认：Client、CLI/子进程、临时文件系统。当前无其他相关验证进程；原工作区仅有本功能三份未跟踪规划文档。

本文件记录实际执行，唯一当前状态见 status.md。实现与规范分别提交；历史失败原件不改写。

普通回执工具：RED `.evidence/barness-ai-test-process-review/e2e/bootstrap-20261022T013323Z` exit1；GREEN `.../run-green-20261022T013417Z` exit0，5组进程场景。清理工具：RED `.../cleanup-red-20261022T013521Z` exit1；GREEN `.../cleanup-green-20261022T013636Z` exit0，7组，仅临时合成目录。发布CLI新合同 RED `.../release-red-20261022T013712Z` exit1。

偏差：`ordinary-red-20261022T013759Z` 自测以 `go test` 为父入口，Go工具链重写PATH，拦截器未生效，旧 `TestMiniMaxReleaseLoadConsumer` 意外启动一次 `BARNESS_AI_PRESSURE=1` 的受控 MiniMax 六负载子进程并通过。此非用户授权，已向用户明确报告；不能将本次压力状态写为 NOT_RUN。没有 race 或厂商 live。后续改为仅编译测试二进制，再直接运行带硬编码拦截器的指定用例；拦截必须使旧入口先失败，才能修改入口。原回执保留，不改写成RED失败。

入口隔离接缝纠正后：`ordinary-guarded-red-20261022T013840Z` exit1，拦截器在旧入口先拒绝隐藏负载；`ordinary-green-20261022T014424Z` exit0，新普通入口读取留存夹具，不启动负载。夹具源为历史 `minimax-current-eleven-gate-2/gate/offline/20261019T064845.243220000Z` 六个 MM05 原生案例，逐文件哈希一致，压缩 JSONL 278680 字节；不产生当前容量或真实支持资格。

补充失败场景：`hardening-red-20261022T014510Z`（大小写 JSON 歧义、必需索引依赖缺失），`input-red-20261022T014711Z`（选定输入与工具身份），`multipackage-red-20261022T015031Z`（多包回执），`record-contract-red-20261022T015236Z`（缺工具链或影响标记），`source-inventory-red-20261022T015455Z`（忽略源码与纯展示文档边界），均 exit1，原件保留。对应修复后完整进程集在 `source-checks-20261022T015556Z` exit0，23组。

定向普通 Go 验证：`.evidence/barness-ai-test-process-review/e2e/20261022T014901.047673Z-00u3v4fm/receipt.json` PASS/exit0。维护工具全套、定向 vet、live-tag 只编译在 `implementation-checks-20261022T015346Z` 全部 exit0；未运行该 live 二进制。随后源码枚举补强仅排除展示 Markdown 与生成缓存，包含被 Git 忽略的源码；实际参与消费的 Markdown 必须通过 --input 指定。

删除原冻结差分、旧自动 check/race 调度及旧固定十门禁/全库选择器。历史原件未动；通用序列化投影与 SHA-256 改为中立测试语义，保留原生完整性、脱敏及资格消费者。旧固定发布断言由五项回执 CLI 进程验证替代，普通行为与消费者安全检查继续保留。
