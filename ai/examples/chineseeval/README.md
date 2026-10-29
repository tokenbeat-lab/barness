# 独立中文分类评估

这是 Host 自己的四类、每类六条合成中文支持请求任务，共 24 条。数据集 bytes/hash 保持
`zh-support-v1`，标注规则、适用范围和每条依据随报告保存。它独立于协议真实冒烟和支持矩阵；
不制定业务阈值、不转人工、不从 fixture 推断真实质量。

当前默认读取 `jev-1.13.0-v2.json`、`jev-route-v2.json` 和 `luna-route-v2.json`，
它们固定 `2026-10-29.1` 目录与对应 Hash，并独立增加配置内容版本。固定任务、
Jev 1.13 / Luna 型号、价格、数据集和预算不变。`v1` 原字节保留历史目录固定值；
旧固定值不能搭配当前目录执行。历史包继续使用包内原始配置及原目录核验，
这些配置更新不把历史结果提升为当前目录下的真实 PASS。

## 公开命令

原命令和默认 Jev/schema 1 保持：

```sh
go run ./ai/examples/chineseeval/cmd/chineseeval
```

默认无网络，生成新的 NOT_RUN 包，全部样本保留，准确率为 null。
显式 `-route jev` / `-route luna` 选择各自配置，产生 schema 2 报告。
不支持的 Operation/Provider/API/Model 或不匹配的 route/config 在读取 key 前拒绝。

不依赖正式目录或 key 的 Host 自带目录流程：

```sh
go run ./ai/examples/chineseeval/cmd/chineseeval \
  -route jev -catalog ai/examples/chineseeval/testdata/host-catalog-v1.json \
  -config ai/examples/chineseeval/testdata/jev-host-v1.json -fixture normal

go run ./ai/examples/chineseeval/cmd/chineseeval \
  -route luna -catalog ai/examples/chineseeval/testdata/host-catalog-v1.json \
  -config ai/examples/chineseeval/testdata/luna-host-v1.json -fixture mixed

go run ./ai/examples/chineseeval/cmd/chineseeval \
  -route luna -catalog ai/examples/chineseeval/testdata/host-catalog-v1.json \
  -config ai/examples/chineseeval/testdata/luna-host-v1.json -fixture all-refusal

go test ./ai/e2e -run '^Test(ChineseEvaluation|ClassifierHostMigration)' -count=1
```

fixture 使用无法建立网络连接的合成 transport；普通答案固定 billing，故意不读取标签。
还有 `usage-absent`、`usage-partial`、`all-refusal-usage-absent`、`all-refusal-usage-partial`、
`cancel`、`pre-attempt-cancel`、`http-failure`、`dual-request-id`、`response-alias`；`budget` 配合显式缩小的
maxCalls/maxQuestions/maxAttempts 配置证明预算耗尽。每次都写新目录，保留失败、完整 raw
请求/响应脚本、Observer、报告、源代码文件 Hash、完整文件 inventory、manifest 和 audit。
fixture 与 `-live` 不可同时使用。

在独立进程中设置所选 route 的专用 key、账户/区域 alias 和开关，再执行真实评估。
Jev 原专用 key `BARNESS_AI_CHINESE_EVAL_TYPESAFE_KEY` 保持，Luna 使用
`BARNESS_AI_CHINESE_EVAL_OPENAI_KEY`。两者均需 `BARNESS_AI_CHINESE_EVAL=1` 和
`BARNESS_AI_CHINESE_EVAL_ACCOUNT_ALIAS`（仅字母、数字、`@._-`）；不得同时提供其他 key。

```sh
go run ./ai/examples/chineseeval/cmd/chineseeval -live
go run ./ai/examples/chineseeval/cmd/chineseeval -route luna -live
```

需要本地出站代理时显式加 `-proxy http://127.0.0.1:7890`。只允许 loopback IP、
有效端口和无凭据的 HTTP proxy，拒绝 query、fragment 与非根路径；仅供 `-live` 使用。
默认不读取环境代理。目标仍是所选路线的官方 HTTPS endpoint，保留正常 TLS 验证。
该参数只改变 Host transport，不改变目录、评估预算或质量指标；离线重放不使用代理。

真实运行属于独立执行票 12。无 key 或未启用为 NOT_RUN，厂商/协议/预算故障为 FAIL，
已发送 Attempt 和已知消耗仍入账；没有伪造 PASS 或制造拒答的重复请求。

## 报告与完整性

schema 1 保持原 Jev 配置与指标含义。schema 2 独立记录 Operation/Provider/API/Model、route
Hash、配置和数据集 bytes Hash、目录/价格内容 Hash、实际代码内容 Hash、Git commit、日期及
每条实际 ResponseModel。Luna 无日期 alias 的权重不称为固定，实际响应型号不改写授权 ModelID。

每条状态 `answered`、`refused`、失败或未执行独立。refusal 进入全部 24 条的准确率分母，不计
正确，不是调用失败。普通答案的条件准确率、混淆、confidence/chosen probability 各自十等宽 bin
和 ECE、多类 Brier 独立计算，不复用 Jev threshold，空 bin 的均值/准确率为 null、不插值。
延迟（全部执行调用的 mean/p50/p95）和费用独立列出。

schema 2 的 protocolStatus 描述已调用的协议完成情况；status 描述任务完整性；metrics 描述
质量。所有样本都得到普通答案或拒答、每次正好一个 Attempt、usage complete 且整包可核验时
status 才是 COMPLETE。全拒答可 COMPLETE：总准确率 0、拒答率 1，条件准确率/ECE/Brier
为 null，费用照实保留。partial/unreported 用量保留，协议成功也不能补造成任务 COMPLETE；
它们不会抹掉合法普通答案的质量统计。COMPLETE 不设置准确率门槛。

最大 24 Call、24 Question、24 Attempt、零重试、串行、总 600 秒、每调用 15 秒；原 schema 1
的 180 秒配置保持。取消不发后续请求。执行的每条开始/结束时间及 duration 随报告保存，
未执行项没有延迟。固定小样本不能证明生产分布或未覆盖置信区间的校准。

## 核验与历史

```sh
go run ./ai/examples/chineseeval/cmd/chineseeval -verify <bundle-directory>
go run ./ai/examples/chineseeval/cmd/chineseeval -route luna -verify <luna-bundle-directory>
go run ./ai/examples/chineseeval/cmd/chineseeval \
  -verify .scratch/barness-ai-pi-1.0/chinese-evaluation-evidence/live
```

核验从包内 dataset/config/catalog 装配，检查价格、route/代码 Hash、完整样本、指标、预算、
Observer 和文件 inventory，并重新审计。schema 2 的 `captures.json` 由同一有界 collector
记录实际传输：完整固定任务请求、响应文本（成功最多 16 KiB + 1、错误最多 4 KiB + 1）、
HTTP status 和路线对应的 request-id header。它不保存认证 header、URL query 或错误字符串。
完整字节通过公共 Client 在无网络 transport 重放；取消/超时的缺失或部分字节按
Attempt/Observer 对账，不为重放而伪造完整响应或重新生成质量指标。原 Jev 历史包没有 catalog 文件，按它自身的
`2026-10-08.5` Hash 解析从原 commit 导出的不可变 recorded snapshot；不重写历史报告、
时间、Hash 或结论。历史核验不赋予当前发布适用性，fixture/NOT_RUN/live 身份不可互换。
当前发布消费者仍须把报告与当前预期 route/config/catalog/code/日期相比较。

Host 输入能力发现、nil/协议选项、结构化说明和封闭 Answer switch 的迁移示例见
[迁移文档](../classifiermigration/README.md)。离线交付与 verifier 反例见
[工单 11 证据](../../../.scratch/barness-ai-openai-decisions/host-tooling-evidence/README.md)。
