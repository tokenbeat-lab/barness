# 工单 08：分类图片容量证据

公共设计负载保留原本地/云端四路线混合合同，另增加每租户四个 Luna 图片分类调用的
成功、完整响应读取处取消、同负载恢复三批。各批同时占满 4/8 并发，并沿用原
256/512 MiB 预算。每次 128 KiB state、八个 4000 字节问题、两张 256 KiB PNG；
MaxImages=4、MaxTotalImageBytes=1 MiB、MaxImageBytes=1 MiB、MaxRequestBytes=4 MiB。
图片 count 占用50%，单图25%，总图50%；每图 base64 349528 字节、data URL 349550
字节，实际完整请求 JSON 862760 字节（20.57%）。全部占用率 <=75%，剩余余量 >=25%。

失败方式先记录于 [classifier-load-failures](../classifier-load-failures.md)。先红后绿验证
政策缺图片和旧 Jev 报告错误通过；实际 public pipeline 写出全部三批 artifacts，导入
删字段/错值/旧报告、错误身份/预算/并发/资源的反例，并重哈希每份变异 manifest，
使反例验证语义而不是仅依赖 hash 变化。导入核对真实捕获的 count/bytes/whole JSON、
Host 政策/目录/hash/Binding、全部结果与单 Attempt、取消和恢复，以及每批资源。
GC reachable heap 是采样下界，TotalAlloc 是涵盖采样间暂存的保守上界；两次强制 GC
覆盖完整 response bytes 未交 EOF 和完整结果保留，另有持续采样、Provider 复制量和吞吐。
输入和响应脚本在基线前生成；各批记录自己的分配增长，跨批捕获保留在堆基线中。

准确命令、退出状态见 validation.json，最终实测表见 [capacity](capacity.md)；load.zip 记录开发失败历史、审查前全仓失败、最终 full-current 旧聊天压力失败及其完整反例、
最终适用压力/导入完整 artifacts、源码、命令日志、普通/full race manifest与审计。
其他全仓 PASS case 原始 artifacts 留在本地 .evidence，仅在包中保存全仓 manifest/audit。
整包 SHA-256 清单和审计供离线重复核验。起点 c4cc495，开发运行 git_commit 为当时 HEAD，
不能解读成未修改的 HEAD 通过；源码随包固定。

这只是本票独立容量及导入交付，不宣称正式 Luna/live 或十组合 Release Gate PASS。
完整发布普通/race消费者接线与真实支持/效果证据由后续票完成。

重放（每次产生新证据目录）：

```sh
BARNESS_AI_PRESSURE=1 go test ./ai/e2e -run '^Test(MixedPolicies|MixedPressure|ClassifierImagesPressure|ClassifierImagesPressureImport)$' -count=1
BARNESS_AI_PRESSURE=1 go test -race ./ai/e2e -run '^Test(MixedPolicies|MixedPressure|ClassifierImagesPressure|ClassifierImagesPressureImport)$' -count=1
go test ./ai/internal/testkit/release -count=1
BARNESS_AI_PRESSURE=1 BARNESS_AI_PIDIFF=1 go test ./... -timeout 30m -count=1
BARNESS_AI_PRESSURE=1 BARNESS_AI_PIDIFF=1 go test -race ./... -timeout 30m -count=1
go vet ./...
go vet -tags live ./...
```

固定包核验：

```sh
(cd .scratch/barness-ai-openai-decisions/classifier-load-evidence && shasum -a 256 -c SHA256SUMS)
unzip -t .scratch/barness-ai-openai-decisions/classifier-load-evidence/load.zip
classifier_load_dir="$(mktemp -d)"
unzip -q .scratch/barness-ai-openai-decisions/classifier-load-evidence/load.zip -d "$classifier_load_dir"
go run ./ai/internal/testkit/audit/cmd/auditbundle "$classifier_load_dir"
```

审查前全仓 fd1b05a 有一个旧 TypeSafe 100ms CallTimeout/ReadIdleTimeout 分类竞争失败，
原 FAIL 与完整 case artifacts 保留，未调整原时限或压力合同。专项重放和 0856b2b 的
最终完整普通/race各自记录，不将后续通过冒充初次通过。

最后按 actual native JSON 补正问题量测：public BoolQuestion 4033 字节、native predicate
带 name/type 为 4050 字节。报告使用所有捕获的较大值，限额仍 8192，比例 49.438% 。
该差异先红后绿，0856b2b 固定最终源码，早期 4033 报告不被改写为当前适用证据。

旧 TypeSafe 超时的原样独立进程重放五次为四 PASS、一 FAIL；同一 100ms CallTimeout/
ReadIdleTimeout 竞争被再次复现。每次独立包均保留；更早 count=5 诊断为全部 PASS，
该进程复用同名 artifact 目录，仅保留末次文件及完整执行日志，manifest 如实记录五个状态。
不能将它称作五套独立捕获，更不能据后续通过清除原始失败。

最终源码 0856b2b 的普通完整运行：7445 PASS、1 FAIL；本票适用的 182 个容量/政策/
导入用例全部 PASS，861 个 PIDiff 用例全部 PASS，审计无发现。唯一 FAIL 为旧 local
chat 压力中的两次 loopback TCP ENOBUFS；准确错误、Provider 结果与原判定完整保留。
该旧负载独立原样重放 PASS，不改变失败历史。最终 race 完整运行 7446 PASS、0 FAIL，182 个本票适用用例和 861 个 PIDiff
用例全部 PASS，审计无发现；完整收据见 validation.json。live 与 Release Gate 均为 NOT_RUN。
