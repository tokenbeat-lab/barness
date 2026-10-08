# 工单 15：混合操作隔离与设计负载证据

本地 OpenOperations 和云端 Host 在同一个只读 Client 中并发调用聊天、OpenAI 图像、
Google 图像与 TypeSafe 分类。显式有限策略、独立 Binding、合法共享账户凭据引用、
宿主身份和类型化结果均经公共 E2E 验证。业务阈值、转人工与路由由宿主决定。

验证代码提交 `817595c`：全量 4,411 个场景 PASS，包含冻结 pi 差分 623 个（pending 0）、
压力 13 个；最终混合操作 103 个全部 PASS。go build、两种 go vet、完整 race 回归通过。
双轴审查报告 3/5 项（含交付脚本及默认开关关联残留），均修复并独立复核，无遗留项。
详细计数和版本见 verification.json。

无需厂商凭据即可复跑：

```sh
go build ./...
go vet ./...
go vet -tags live ./...
BARNESS_AI_PRESSURE=1 go test -race ./ai/e2e -run '^TestMixed' -count=1
BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test -race ./... -count=1
```

压力默认 NOT_RUN，显式开关启用；not-run.json 保存未启用时的实际记录。每次 E2E
写 `.evidence/barness-ai/<UTC run-id>/` 的版本、断言、逐场景回放命令和自动脱敏审计。

本地设计为一租户 4 并发，云端两租户 8 并发（各 4）：128 KiB 聊天输入、16 KiB 输出；
每次图像编辑两张 256 KiB 参考图，OpenAI 另有 256 KiB mask，输出两张 512 KiB 图；
分类状态 128 KiB、8 个约 4 KiB 问题。全部输入/输出逐项限额占比 ≤75%，并发按设计占满，
混合缓冲增长预算分别 256/512 MiB。PNG 为已知签名加 padding，验证协议字节与编码负载。

客户端 Body.Read 已交付全部 unary 响应字节后暂停 EOF，确认 4/8 个许可仍持有；
释放屏障后记录完整解析和原子输出。强制 GC 分别保留完整 body 与已发布结果，记录可达堆
增长（采样捕获下界）；调用期 TotalAlloc 是新增可达内存的保守上界，包含 GC 采样之间
的暂存副本及 Provider/Observer 开销，两者均须在预算内。Provider 每个调用仅捕获一次
有界请求（约 1.95/3.90 MiB），响应脚本按路线共享且在基线前生成（约 3.54 MiB）。
报告保留 base64 膨胀、实际限额用量/余量、吞吐和生成参数，压力正文不复制进证据。
精确耗时不作机器无关的通过条件。

本次完整 race 运行的测量值（darwin/arm64，Go 1.26.2）：

| 策略 | 并发 | GC 捕获的堆增长 | 调用期总分配上界 | 增长预算 | 吞吐 |
| --- | ---: | ---: | ---: | ---: | ---: |
| 本地 | 4 | 8.42 MiB | 144.72 MiB | 256 MiB | 6.51 次/秒 |
| 云端 | 8 | 16.42 MiB | 289.55 MiB | 512 MiB | 12.42 次/秒 |

mixed-cases.json 保存 103 个场景的实际请求、结果、断言、观察记录及原始 artifact 哈希，
覆盖四维授权、零值聊天 Binding、前置拒绝、快照冲突、轮换/撤销与固定重试、输入独立
冻结、取消/失败与准入等待、完整响应持有许可、原子超限输出、Observer 失败隔离及
三条新路线独立停用。resources.json 的全部资源读数归零；pressure.json 是实际负载报告。
red.json 保存实现前和审查修复前的真实失败；review.md 分列两轴结论。

manifest.json、run-audit.json 和完整日志来自修复后全量 race/差分/压力运行。
trace.json 是既有 releasegate 对五个新增 T/C/H 项的 PASS 结果，并记录离线、差分、
脱敏与目录快照门禁。`-bundle` 评估不会运行命令，其命令门禁显示 NOT_RUN；本次实际
命令退出码另存 command-results.json 与 verification.json，每条记录关联完整日志 SHA-256、
运行前后提交；源码哈希逐项绑定该测试提交。缺失日志、失败命令或版本不一致均拒绝采集。
九路线真实 API 发布验收由工单 17 交付。

核验源文件、交付哈希、断言、默认开关、资源和预算：

```sh
python3 .scratch/barness-ai-pi-1.0/mixed-operations-evidence/verify.py
go run ./ai/internal/testkit/audit/cmd/auditbundle .scratch/barness-ai-pi-1.0/mixed-operations-evidence
```

重新采集时先审计完整 run，运行 `releasegate -bundle <run> -out <ignored-evaluation>`，
可先执行 `run_commands.py <ignored-output>` 留存真实命令结果和日志，
再执行 `capture.py <仓库相对run路径> <仓库相对evaluation路径> <command-results.json路径>`；审计交付后重跑 capture
以纳入 audit.json 的哈希，最后执行 verify。每个工单独立保存 capture/verify 的版本语义。
