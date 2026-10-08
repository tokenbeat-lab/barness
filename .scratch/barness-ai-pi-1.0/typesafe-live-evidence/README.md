# 工单 08：TypeSafe 真实冒烟与目录

2026-10-08；独立组合 classifier × typesafe × typesafe-system-one，固定 jev-1.13.0。
先记录 [失败方式](FAILURES.md)，沿公共 Client、现有报告合并/CLI 与 evidence 审计接缝验收。
[官方能力 fixture](../../../ai/e2e/testdata/typesafe/official-capabilities.json) 保存官方来源、抓取日期和内容哈希；
目录 2026-10-08.3 输入每百万 token $0.042、输出免费，别名不纳入。

[最终 live 报告](live/live-report.json) 与脱敏 wire 完整保存在 live/，三项均 PASS：
混合 choice/score/bool、单选、有界上下文错误探针。每进程 6 次调用、10 个问题，
每次状态 131072 字节，最多 1 次环境重试；最终实际 3 次调用/HTTP 尝试、5 个问题、
0 重试，明确上报输入 714、输出 92 token。错误请求未报用量，消耗未知。
Observer 独立记录只含元数据，各场景 result、exchanges 与 observations 可核验。

**422 形状未确认**：公共 Client 提交 80002 字节（40,000 个合成词）的有效状态、一个有效问题；
没有触发宿主容量限制，官方实际返回 HTTP 400，detail.error_type=max_tokens_exceeded，
被分类为 invalid_request/request。探针完成只证明这个具体上下文 guard；不声明观察到 422。
任意其他 400 仍失败；不扩大探针、改送空问题或绕过入口。
缺 usage 为 unreported，结果零值不代表未消耗。空问题由离线 Client 本地拒绝、零请求证明。

另一次 200 的 score=1.32、分布 .01/.67/.32（期望 1.31）先被现有严格校验拒绝。
真实回放先红后绿后，仅对 .01 网格启用有界舍入可行性：每项 ±.005，必须有一个
总和为 1 的潜在分布，其期望区间与 score±.005 相交。高精度仍用 1e-6，
概率和/范围/键集/图例仍严格检查，原值保持；1.34 与高精度 1.315 不一致会整体拒绝。
这是根据自己的 wire 作出的兼容推断，厂商未承诺其舍入算法，详见 ADR-0021。

[运行历史](historical-runs.json) 保留所有五次真实进程的调用/问题/用量和失败结果，
失败进程的完整证据留在 failed-*/；FAIL 未降为 NOT_RUN/UNSUPPORTED。
首次失败暴露 400/422 假设，另一次失败暴露 score 舍入，期间发生一次有界连接重试。
首次通过使用宿主候选目录，之后才纳入生产目录并复跑；最后 live 的实现提交为 00038a7，
报告边界审查修正为 fd5495f（不改变分类协议）。
账户别名 local-typesafe@region-unreported：用户提供 Key，账户区域与权限级别未核实。
[run-live.py](run-live.py) 只读取根 .env 的 TYPESAFE_KEY，子进程只持有本组合凭据；不 source .env。

矩阵 schema 2 增加 operation，六个旧组合明确 chat 并保留历史。
[matrix-before.json](matrix-before.json)、[matrix-after.json](matrix-after.json) 与
[matrix-verification.json](matrix-verification.json) 证明只更新 TypeSafe。
[not-run 报告](not-run/live-report.json) 为三项 NOT_RUN，合并后矩阵字节不变。
离线 CLI 验证一批报告后项拒绝不会部分写入；旧 schema、错身份、错型号、缺能力、
超预算或实际计数不一致的报告拒绝。成功报告还必须有提交字节和非空厂商请求 ID。
审计白名单复核没有开放 Observer 正文；x-typesafe-request-id 与 operation 都是元数据。

红/绿日志按边界保存。编译红记录代表公共报告字段尚未实现；身份、目录、真实 score 与
复审边界另有实际行为失败。测试装置曾修正浮点精确相等和切片长度的错误预期，未据此改生产行为。
两轴 [code-review](review.md) 初审发现均修复，最终两轴零遗留。

复现（仓库根目录，不联网的命令在前）：

    go test ./ai/internal/testkit/supportmatrix ./ai/internal/testkit/release -count=1
    BARNESS_AI_PIDIFF=1 go test ./ai/e2e -run '^Test(Classifier|SupportMatrixCLI)' -count=1
    go test -tags live -race ./ai/live -run '^TestClassifierHarness' -count=1
    go vet ./...
    go vet -tags live ./ai/...
    BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test -race ./... -count=1
    python3 .scratch/barness-ai-pi-1.0/typesafe-live-evidence/run-live.py

真实请求、响应和官方事实的公共离线回放固定在 testdata/typesafe/live-contract.json；
全部能力结论来自本组合，未借用聊天 Provider 的结果。整个版本发布仍需旧六条路线在升级后复跑，
以及后续两条图像路线自己的证据；本工单不宣布整版可发布。

追溯的 V6-typesafe-live 条目按自己的 P07 官方 wire 回放统计离线证据；
真实通过单独由 liveCombos 中的 typesafe-classifier 行与该行自己的已审计 live bundle 判定。
未把通用 LIVE 聚合塞进此条目，因为既有 LIVE 语义聚合所有组合，会让别的 Provider 的状态
混入 TypeSafe 专项结论。离线回放与真实调用证据分开报告。

最终验证提交 c1209d0：`BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test -race ./... -count=1`
退出码 0、耗时 498.794 秒；3740 用例全部 PASS，含 623 差分（0 待处理差异）、11 压力场景。
[完整 manifest](full-bundle-manifest.json) 与 [完整包审计](full-bundle-audit.json) 保留判定，
[304 个 P07/差分用例](offline-cases.json) 保存重放命令与各产物哈希；新增 10 个验收用例的
实际请求、响应和断言也可在 offline/ 独立查看。原始 179 MB 全量包保留在本地 .evidence/，
不提交重复的全量捕获。完整执行日志见 [full-race.log](full-race.log)，最终装置 race 和两种 vet 也通过。
首轮仅旧目录 pin 失败，记录保留在 [full-pin-failure.json](full-pin-failure.json)，并有红绿断言。

[发布数据核对报告](release-evaluation/release-report.md) 的 TypeSafe live 行、V6/H2 专项追溯、
P0、差分、审计与快照均通过。`-bundle` 模式没有执行 command gates，且未提供旧六条路线的
新审计包，因此该报告整体 FAIL，命令退出 1；不将它冒充完整版本发布。独立执行的 vet/race
通过证据由 [report.json](report.json) 与对应日志报告。最终交付目录审计零发现。

无需 Key 或联网即可校验交付文件和源码哈希、矩阵隔离、NOT_RUN 保留以及 own live/离线结论：

    python3 .scratch/barness-ai-pi-1.0/typesafe-live-evidence/verify.py

[manifest.json](manifest.json) 固定所有交付产物哈希。若本地原始全量包仍在，验证器也逐项核对
其 manifest、审计与 304 个所选用例的原始文件；包不存在时，仍可核对可移植的新增用例，并按
offline-cases.json 的命令重放。涉及生产分类/目录的修改均记录在 ADR-0021 和对应契约中。
