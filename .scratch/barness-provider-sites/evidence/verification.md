# 地区 Provider 整合验证索引

本索引记录固定来源的验证事实；唯一当前实施与正式验收状态见 [spec](../spec.md#当前状态)。集成分支 `codex/barness-provider-sites`，实现工单 01–06。原生证据保留在仓库根 `.evidence/provider-sites/`，不复制重导出旧捕获。

## 来源与工具

- 最终受测提交：`d4b63d9b2ecec3dc7fcbdadf3b849e31c83cd92e`。
- 完整已跟踪 `ai/` 树及 `go.mod` / `go.sum` 的逐文件 SHA-256 清单与命令前后相等性在 [普通回执目录](../../../.evidence/provider-sites/final-v4/ordinary/)。聚合值：`470682d37e2d1c839bf5609f6898f9a1e4dafbd3dbfcd2cb77f2048a530f3e05`。
- 原生资格执行来源集合使用独立的现有算法，值为 `f38486193de4d19980123968616286259eb97a6cc500a29f8eee7a18ae1a28bc`；其完整文件清单、构建注入和 binary Hash 在受控原件内。它与完整普通测试树 Hash 的范围不同，不能互换。
- [工具身份与一次性审计脚本](../../../.evidence/provider-sites/final-tools/toolchain.json) 保存 Go、Node、Python、Git 的实际版本及脚本 Hash。冻结 pi、SDK、依赖、正式目录和旧 snapshot 字节未变。

## 最终普通结果

| 检查 | 实际结果 | 墙钟 | 原始回执 |
| --- | --- | --- | --- |
| `go vet ./...` | PASS，exit 0 | 0.729s | [receipt](../../../.evidence/provider-sites/final-v4/vet/receipt.json) |
| `go vet -tags live ./...`（无厂商调用） | PASS，exit 0 | 0.558s | [receipt](../../../.evidence/provider-sites/final-v4/vet-live/receipt.json) |
| 完整普通 `go test -json ./... -count=1 -timeout=20m` | PASS，exit 0 | 661.883s | [receipt](../../../.evidence/provider-sites/final-v4/ordinary/receipt.json) |
| 合同、格式、500 行与本地引用 | PASS，exit 0 | 0.301s | [receipt](../../../.evidence/provider-sites/final-v4/static/receipt.json) |
| releasegate CLI 构建 | PASS，exit 0 | 2.014s | [receipt](../../../.evidence/provider-sites/final-v4/build-cli/receipt.json) |
| 旧 HTTP 请求审计 | PASS，exit 0 | 0.887s | [receipt](../../../.evidence/provider-sites/final-v4/wire-audit/receipt.json) |
| 正式前置（缺少 live / 效果输入） | BLOCKED，exit 1（预期） | 2.095s | [receipt](../../../.evidence/provider-sites/final-v4/preflight/receipt.json) |

完整普通明确开启 `BARNESS_AI_PIDIFF=1` 与 `BARNESS_AI_PRESSURE=1`；`BARNESS_AI_EVIDENCE_DIR` 指向每次新输出目录。数字退出、UTC 起止、墙钟、原始 JSON stdout、来源前后 Hash、原生 manifest 与 audit 见各回执。父子测试节点不当成独立场景，包耗时不相加当墙钟。

主 E2E manifest 有 **13,244 个原生场景，全部 PASS**，另一个 `ai` 包原生 manifest 为 23 PASS，分别计数；主包无 NOT_RUN、重复 ID 或审计发现。32 份真正原生 Run 的 audit 均为零发现。六条正常 controlled 子进程合计 81 次受控 Attempt，通过；授权复用负例刻意运行的失败子进程不合并进主包计数。Go 节点为 14,306 PASS / 4 SKIP，包为 9 PASS / 27 无测试包 SKIP。四个节点 SKIP 是两个私有子进程入口、可选旧 bundle 诊断及可选 100 轮 loopback 诊断，不是必需差分/压力跳过。

主包实际产生 **889 份 pi 差分记录，全部 PASS、pending 0**，其中本功能为 18 次双侧调用与 3 次目录比较。名字含 pressure 的 204 个原生场景通过，包含实际压力及导入拒绝边界；完整记录见 [原生计数/审计汇总](../../../.evidence/provider-sites/final-v4/ordinary/native-audit-summary.json)。主包原件为 [manifest](../../../.evidence/provider-sites/final-v4/ordinary/bundles/20261020T030404.702080000Z/manifest.json)，Hash `9e5d1aa03afa2c9f96cf5d27adbf9b1284ffb62bd28852a72215b22f183cebbe`。

旧请求逐工件审计 **3,077 / 3,077 PASS**，覆盖 2,480 个实际 HTTP 请求及 body/零发送投影；108 份非 wire 的 Observer、错误、用量与 metadata 工件明确列为范围外。仅规范化 ADR-0049 指定的实际 `prompt_cache_key` 和既有 `Session_id` / `X-Client-Request-Id` 值，字段有无及其他 wire 值保持严格；七个明确的并发案例按 path/key_alias/model 分流后保留各流顺序，不断言跨流调度顺序。完整字段、Hash 与比较明细见 [comparison](../../../.evidence/provider-sites/final-v4/wire-audit/comparison.json)。

正式前置使用未提供 `-live` / 效果包的实际输入，退出 1；snapshot 与 difference-ledger 通过，新六路线从未合并自身正式 live，旧 live / 四项效果未选入。CLI 未启动后续验证命令。上述双 vet 与完整普通为开发整合诊断，不是可供 `-stage race -from` 续接的正式 check checkpoint；`READY_FOR_RACE` 不成立。

## 原件与审查

- [协议](../../../.evidence/provider-sites/protocols/index.json)、[缓存与隔离](../../../.evidence/provider-sites/runtime/index.json)、[Host/候选最终联调](../../../.evidence/provider-sites/host/final-runtime-integrated-host/receipt.json)。
- [独立冻结差分](../../../.evidence/provider-sites/differential/index.json)：ZAI Chat 两型号及 MiniMax CN Messages M3，18 次实际双侧调用比较与 3 次完整 model metadata 比较；四条缺少冻结路线的组合为 24 次 extension 执行，不计 pi parity。原八子集、锁文件与 pins 保留。
- [自身采集与消费流水线](../../../.evidence/provider-sites/release/index.json)：candidate/formal 分离、23 行支持目标、10 个差分子集、40 个独立有限需求项。
- [双轴原审查、修复与最终复审](review.md)。[第一轮修复](../../../.evidence/provider-sites/review-fixes/index.json) 处理授权身份记账、独立 endpoint 校对、原生终态/用量/内容/工具/native 重放及 callback policy；[边界修复](../../../.evidence/provider-sites/review-fixes/followup-1/index.json) 处理等价时间 UTC 与畸形取消前缀。均保留公开 red/green 和失败原件。
- [旧 OpenAI golden](../../../.evidence/provider-sites/review-fixes/chat-golden-v2/index.json) 与 [DeepSeek golden / GLM 当前拒绝阶段](../../../.evidence/provider-sites/review-fixes/legacy-golden-source/index.json)：三个固定 v2 期望由独立 Python `hashlib` / `struct` 算出；仅迁移活动期望，不改 47 份历史工件。GLM 派生视图精确断言 importer 漂移，原始执行/用量精确断言 execution 漂移，Audit 仍精确拒绝；旧/现执行与 importer 清单独立保留。

新增源码均遵守 500 行上限，格式与本地引用通过；原 17 条支持记录及 394 个追溯项逐对象不变。40 个地区项保留原全部 mandatory，仅按语义追加审查反例。固定合同拒绝通过删除映射缩减需求。

## 保留的失败与范围

- [实施前基线](../../../.evidence/provider-sites/baseline/receipt.json)：3376 个原生场景中 3367 PASS / 9 NOT_RUN，当时未打开差分/压力，不计为新功能证明。
- 临时 worktree 不包含旧 ignored archive 的消费失败原样留存于 protocols/07 与 integration-metadata/legacy-worktree-archives-missing；它们不表示厂商失败。根目录 [旧来源实际适用性失败](../../../.evidence/provider-sites/legacy-source-applicability/receipt.json) 明确为 execution source drift，未改原件。
- [第一次完整普通日志及收集错误](../../../.evidence/provider-sites/final-v3/ordinary/collector-failure.json)：原生主包 13238 PASS / 6 FAIL，失败是三个旧缓存期望与三个旧 GLM 拒绝阶段断言，原生审计零发现。外层 collector 退出 1；后处理把两份示例 manifest 误认为 native Run，未先落盘子进程数值退出码，子进程数值标记未捕获，原始 Go package FAIL 不改写。最终 collector 先保存 process-receipt，再按原生 Run schema 汇总。
- 初版 wire 审计错误纳入非 wire 工件并比较跨流调度顺序，得到 21 个审计差异；[原始比较](../../../.evidence/provider-sites/final-v3/wire-audit/comparison.json) 保留。修订审计范围后通过，未放宽其他 wire 字段。
- 其他 red、fixture/编译失误及目录名带 green/final 但实际退出非零的轮次，在各索引逐项保留；数字退出和原件优先于目录名。受控厂商替身的预期失败子进程不计为真实资格 FAIL，也不并入主包 PASS 数。


收尾文档提交 `6d90ec8` 后的 [最终静态回执](../../../.evidence/provider-sites/final-v4/closure-static/receipt.json) 再次通过全部合同、格式、500 行及 287 处本地引用检查，完整执行来源仍与普通受测集合逐文件相等。五个干净且全部合并的实现 worktree 已安全移除，分支和根目录原生证据保留，见 [清理回执](../../../.evidence/provider-sites/final-v4/worktree-cleanup/receipt.json)。后续仅补本段及索引链接，不改变执行来源。

## 人工 race 与自身 live 交接

本轮实际 race 为 `NOT_RUN`，实际付费/厂商调用为 0。预计完整开发 race 约 30–60 分钟（估算，不是测量或执行结果）。如人类明确选择该范围，可从本索引固定来源、相同输入与工具身份开始，使用新目录执行：

```sh
BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 \
BARNESS_AI_EVIDENCE_DIR=/absolute/path/to/new-regional-race \
go test -json -race ./... -count=1 -timeout=75m
```

覆盖范围是完整包集合，包含四服务并发快照、取消/队列/资源与旧 Provider 回归、差分和压力。执行前须按验证策略重核来源、已选输入和普通回执；本命令本次未执行。它不会补足自身 live 或正式外部效果证明。

将来正式发布须先取得新九身份自身完整 candidate 资格、独立晋升与新 formal 运行，并由当前原生消费者核对旧服务和四份效果包；再用 [发布指南](../../../docs/barness-ai/README.md) 的正式 check，获得真实 `READY_FOR_RACE` 后，人工选择 `-stage race -from <check-dir> -out <new-race-dir>`。本轮没有有效的该类 checkpoint。真实调用须另有身份、用途、范围和预算授权，入口与隔离环境见 [REGIONAL.md](../../../ai/live/REGIONAL.md)。

新9仍为候选，不在 BuiltinCatalog。一次性授权的持久化边界是可信 Git common directory；跨独立仓库或主机的统一审批由 Host 提供。迁移、缓存命中率影响、凭据用途和回滚步骤见 [地区服务指南](../../../docs/barness-ai/regional-providers.md)。完整正式验收继续开放，`release.pass: false`。
