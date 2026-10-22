# 第 1 节七项复审修复

基线：3ddc774084fe502ca1a13ab8b41b208053e24378。用户仅授权修复及受影响普通验收。接缝沿用 plan.md §1.2：公共 CLI/子进程、临时文件系统和原生证据消费者；不运行后续计划。

先列失败方式，逐项先红后绿：

1. 普通进程继承持久 GOFLAGS=-race；子进程过滤环境后重新读取 HOME 中的配置；回执未记录隔离政策。只运行隔离配置下的 go env，不执行 race。
2. 有 PASS receipt 的目录同时含失败/未跑 manifest、未知文件或 grant/claim/账本，仍被隔离；清理分类须对整个候选范围保守判定。
3. 未被文档索引的失败原件引用成功目录；失败/未知保留运行的引用必须递归进入保护集合。
4. 状态写入遇到限额/磁盘失败或中断，截断旧状态并阻断恢复；旧有效状态必须留存，移动日志能恢复已移动内容。只用临时数据。
5. 后代持有输出管道并另建 session；外层必须有界结束并落失败回执，不因非 daemon 线程挂住。取消和输出限额仍保留诊断与资源释放。
6. 通用 live 的资格、预算、调用、身份或原件不完整；不能凭完整 PASS 标签和消费者自报 true 放行。复用各路线完整原生消费者，不执行厂商请求。
7. 原生 live package 含 build tag/timeout；按原生合同发现唯一 bundle，拒绝错包和歧义，不能由假生产者的简化格式掩盖。

原件使用 .evidence/barness-ai-test-process-review/e2e/ 下新的独立目录；失败原件不覆盖。清理仅在独立临时合成仓库验收。race、压力、付费 live、真实证据清理均不执行。

## 修复与失败原件

下列目录均位于 `.evidence/barness-ai-test-process-review/e2e/`；每个目录有实际命令/数字退出/时间及日志。RED 的非零退出保留原事实，不以成功重跑覆盖。

| 七项意见 | 先失败的进程回执 | 实现与验收边界 |
| --- | --- | --- |
| 持久 Go 配置可隐式开启 race | goenv-red-20261022T024513.135825Z；go-policy-record-red-20261022T025618.705155Z | GOENV=off、空 GOFLAGS、普通独立 HOME，子进程重建环境仍无 -race；只读判定拒绝不安全/缺失政策，测试仅读 go env |
| 通用 live 缺少完整资格 | live-qualification-red-20261022T025252.402101Z；live-consumer-red-20261022T025327.182800Z | 复用完整原生 Merge 和各路线构造器；只读消费者生成精确输入/结果/来源/命令/日志回执，删除自报 argv/exit 格式，拒绝零调用、预算、身份、请求和回执篡改 |
| 原生 live package 不匹配 | live-package-red-20261022T025136.594415Z | 保留 tag/timeout 原字符串，按原生格式发现唯一主包；合成 Python 生产者不执行 live |
| 混合候选掩盖失败/未知/授权 | cleanup-mixed-red-20261022T024639.561753Z；cleanup-audit-red-20261022T025743.906426Z | receipt 与全部 manifest/audit/声明库存同时成功，未知字节、失败、缺 audit、grant/claim/ledger 阻止整候选隔离 |
| 未索引失败的传递依赖未保留 | cleanup-dependency-red-20261022T024734.191263Z | 失败及未知候选也是递归引用根，不能仅依赖 Git 文档与三个固定索引 |
| 状态保存失败截断恢复依据 | atomic-state-red-20261022T024821.511506Z | 同目录临时写、fsync、replace、目录同步；临时假 payload 已移动后注入文件大小限制，保存失败保留原 MOVED 状态，公开 restore 成功 |
| 后代持管道导致超时挂住 | pipe-deadline-red-20261022T024936.311576Z | 无阻塞读取线程；同一 deadline 覆盖进程及管道，关闭读端、有界返回 CAPTURE_FAILED；真实小 Go 进程取消仍留数字退出 |

消费者新增边界也先验证失败：`consumer-boundary-red-20261022T025620.795957Z`（输出符号链接可写回原生包）、`consumer-invocation-red-20261022T025841.250906Z`（等价 Go flags 顺序产生不可消费回执）。对应修复禁止直接/间接写入 native bundle，保留实际 argv 并按真实 flag 语义核对。

中间 GREEN 尝试中的失败同样保留：`live-consumer-green-20261022T025432.329313Z` 为合成 observation 非法字段，`live-consumer-green2-20261022T025520.990702Z` 为临时目录 /var 与 /private/var 别名，`fixes-process-green-20261022T025643.860531Z` 为新输出目录检查误要求叶目录已存在。未放松原生消费者：修正假 observation、规范化临时仓库根、在创建新输出前检查已有父目录。

本轮所有清理子命令只在 TemporaryDirectory 假仓库执行。脱离 session 的测试后代由用例按已知 PID 回收；工具只终止自己建立的进程组，不发现或杀死任意外部 session。live 原件/授权/账本规则不变，未调用厂商。旧压力误跑事实继续保留在 status.md / execution-log.md，本轮没有新增压力执行。

中间 GREEN（fbbcab8 的原生包随后被 Host 审计拒绝）：`.evidence/barness-ai-test-process-review/e2e/fixes-process-source-bound-20261022T030106.680355Z`，39 组 exit0，源码前后相等且与冻结提交 `fbbcab8431e285ec4586b306c9b21ea728005685` 相等。完整普通 Go 原件 `.evidence/barness-ai-test-process-review/e2e/20261022T025935.242681Z-73oqveme/receipt.json` PASS / exit0，12890 通过节点；原生 {'PASS': 12086, 'NOT_RUN': 214}、audit0，原 E2E 通过集合保持相等。两种 vet 和 live-tag 仅编译全部通过。当前状态、只读判定与精确路径统一见 status.md / verification.json。

收尾自检追加失败方式：消费者回执中的 stdout/stderr 文件必须与公共 CLI 实际输出完全相同，不能只在磁盘生成名为 stdout 的结果快照。新的只读复核原件 `record-check-fixes-20261022T030635Z-a1pubf4m` 拒绝 E2E 原生 audit；当前状态已恢复为排查中，原状态快照及拒绝报告保留。

只读审计不一致已定位：普通 HOME 隔离阻止了持久 Go 配置，但旧进程报告只清除当前 HOME，原宿主 HOME 前缀不再被该清理覆盖。原生 producer 的 audit 使用隔离 HOME 因而漏报，Host 的只读复核仍拒绝该原件。新增公共 run.py → 现有单个受控 Go E2E → 原生 artifact 失败场景；保留两次审计的区别，不放松审计规则、不改旧 bundle。

最终 GREEN：`.evidence/barness-ai-test-process-review/e2e/final-process-source-bound-20261022T031329.728119Z`，41 组 exit0，源码前后相等且与冻结提交 `3aafd2d41436d222449f5b5f0e8c0596ed1154e9` 相等。完整普通 Go 原件 `.evidence/barness-ai-test-process-review/e2e/20261022T031330.272540Z-x_6lvwke/receipt.json` PASS / exit0，12890 通过节点；原生 {'PASS': 12086, 'NOT_RUN': 214}、audit0，原 E2E 通过集合保持相等。两种 vet 和 live-tag 仅编译全部通过。当前状态、只读判定与精确路径统一见 status.md / verification.json。

最终 Host 只读复核已通过 source / E2E / live 不适用三项；race 与正式压力仍 NOT_RUN，实际退出1，release.pass=false。命令回执：`.evidence/barness-ai-test-process-review/e2e/final-readonly-record-20261022T032023.972949Z/command-receipt.json`。宿主HOME脱敏与消费者真实 stdout/stderr 的追加失败场景均先 RED 后 GREEN；最终41组及完整普通验收覆盖它们。
