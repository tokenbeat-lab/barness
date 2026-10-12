# 工单 06：运行时隔离与回滚证据

新增 63 个公共 E2E 验证 Decisions 的默认零重试、有限 Binding retry/header/backoff、
取消/时限、bounded unary 响应、共享 Client 的 Jev/Luna/chat/image 隔离、
快照 mutation、Observer 拥塞/error/panic 和三种授权撤销。生产改动只补充
非 2xx JSON error.message 的有界脱敏诊断；没有另建 runtime 或自动 fallback。

撤销对照分别关闭 Enabled、移除 AllowedModels 和关闭 AllowClassifierImages，
覆盖入口带图与 Hook 加图。新调用拒绝且不发送；共享账户/key 的其他 Tenant 和
其他路线继续成功，旧图片调用及重试仍使用 b1，已报告费用保留；恢复授权由新调用读取 b3。
图片撤销后的纯文本调用仍可用。重试 bytes/headers/key、目录/hash 与图片权限固定，
凭据后端修改 Binding 入参不影响副本；mutation 在明确复制/实际发送 barrier 后执行。

八路同时在途保留每租户捕获、答案、图片、usage、每操作独立的 USD 费用和完整归属。
Unary 输出/error exact/over 最多读 limit+1，JSON 深度 10000 外拒绝；Frame/event allowance
只有 16 bytes 的对照仍可读取 1024-byte unary 成功结果。每次结束的 body/call/permit/waiter/event
和 Host permit 都为零，取消等待者不影响持有者。Observer 成功/失败不受 slow/error/panic
影响；队列 1 的固定 barrier 下 16 条中 2 delivered / 14 dropped，无正文、图片或秘密。

起点 `26d955b`，实现 `c9124ef`，补强断言 `fb30d25`。两轴 code-review 最终均 0，
见 [REVIEW](REVIEW.md)。准确完整的执行命令、版本、退出状态、case 计数与审计结果见
[validation](validation.json)。新用例全部 PASS；既有 2xx usage/refusal/read/hook 反例和
图片验证取消一并重放。完整普通与 race 均开启原有冻结 Pi 差分，vet（普通和 live tag）执行。
没有将既有 opt-in 压力 NOT_RUN 改为 PASS。

runtime.zip 保留全部开发执行历史和专项 race 的完整捕获，最终全仓执行保留本票
63 新例、9 个 DecisionsUsageAfter2xx 及图片取消的完整 artifacts；其他全仓 case
只保留 manifest/audit 汇总，其原始文件在本地 .evidence。最终适用 case 保存实际产生的
响应脚本、request capture、result/error/usage/Attempt、Observer、assertions、资源、fixture/hash
和精确 replay；前置取消无响应脚本，保留零请求/零 Attempt 与资源读数。历史两项候选装配
失败只产生 assertions.json，按实际保留。同时保存源码/spec/版本与执行日志。
聚合 observer-records.json 保留原名，增加字节相同的 observations-observer-records.json
审计视图，触发现有 Observer 字段白名单；两份的相同 SHA 及映射在内部 manifest 登记。
manifest.json 列出所有 ZIP 成员 SHA-256，
SHA256SUMS 固定压缩包与说明/收据。整包审计 audit.json 无发现。

初次 14 case 的 6 FAIL 是诊断缺口的红→绿；之后生命周期 4 FAIL 是请求阶段 aborted
预期修正，混合场景 2 FAIL 是候选 MinChoices 装配修正。一次 CostRates 嵌入字段编译错误
在执行前修正，未伪造运行 artifact。失败包保持原 verdict，不改写历史。
首次全仓 race 的所有 case 通过，但收尾整包审计超过 go test 默认进程时限，
在 11 分钟被终止，命令 FAIL 保留。该包另用相同普通 Go auditor 完成审计，
不据此把原命令改为 PASS；显式 -timeout 30m 重跑完整 race，所有调用时限和断言保持原值。
历史 manifest 的 git_commit 是当时 HEAD，开发运行含当时未提交的测试/实现，
不是该 HEAD 纯源码的通过结论；最终全仓执行固定实现/测试源码为 fb30d25。

自己的 [retry 披露](retry-disclosures.json) 记录冻结 Pi classifier 默认 2 重试、Decisions
504 不重试与 Binding 策略的差异，来源是固定代码 SHA 与 barness 捕获。
新 Decisions 自身 Pi differential、设计负载、live 和发布结论均为 NOT_RUN，归后续工单。
本票不修改正式内置型号或支持矩阵。

仓库根目录核验固定包：

```sh
(cd .scratch/barness-ai-openai-decisions/runtime-evidence && shasum -a 256 -c SHA256SUMS)
unzip -t .scratch/barness-ai-openai-decisions/runtime-evidence/runtime.zip
runtime_bundle_dir="$(mktemp -d)"
unzip -q .scratch/barness-ai-openai-decisions/runtime-evidence/runtime.zip -d "$runtime_bundle_dir"
go run ./ai/internal/testkit/audit/cmd/auditbundle "$runtime_bundle_dir"
```

重放当前源码（新运行新目录，不覆盖本包）：

```sh
go test ./ai/e2e -run '^TestDecisionsRuntime' -count=1
BARNESS_AI_PIDIFF=1 go test -race ./ai/e2e -run '^Test(DecisionsRuntime|DecisionsUsageAfter2xx|DecisionsImagesCancellation|Unary|Mixed|JevRegression|Classifier)' -count=1 -json
BARNESS_AI_PIDIFF=1 go test ./... -count=1 -json
BARNESS_AI_PIDIFF=1 go test -race ./... -timeout 30m -count=1 -json
go vet ./...
go vet -tags live ./...
```

pack.py 只归档已有结果，参数指向实际 .evidence 的 history/race/full 根目录、validation
收据及日志；默认不会重跑测试。来源参数记录于 validation 的 packagingCommand。

最终完整普通执行与显式 30 分钟时限的完整 race 均退出 0：各 7,221 case，
7,212 PASS / 9 既有压力 NOT_RUN；各 815 个既有冻结差分全部 PASS，pending=0。
两种 vet、编译检查及所有已完成的整包审计无发现。最终完整执行本票适用的
63 新例 + 9 个 2xx 反例 + 图片取消共 73 个 case 均 PASS，源码为 fb30d25。
