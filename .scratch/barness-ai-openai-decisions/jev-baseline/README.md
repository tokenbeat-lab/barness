# Jev 公共回归基线（工单 01）

2026-10-16 当前 HEAD 的独立重放、版本偏移、准确命令和产物 Hash 见[本次复核](recheck-20261016/README.md)。下文仍为 2026-10-12 的历史基线说明，原固定包及其状态保持不变。

受测 HEAD 与设计 Baseline 均为 `8ea61b8d70de6c84fe27be34dc7d21c5c40841d9`，没有偏移。开始时已有未提交的 GLOSSARY、Decisions 规划/研究和 ADR-0026；本票保留这些改动。生产源码、已有目录、fixture、oracle、账本及历史报告逐文件 Hash 不变。本票只新增公共 E2E 和证据，不实现 Decisions，也不运行真实 API。

| 本次执行 | 结果 | 证据位置（baseline.zip 内） |
| --- | --- | --- |
| 已有 Jev、目录、回调、unary、官方 wire replay | 542 PASS、2 NOT_RUN；命令退出 0 | `runs/existing/`、`existing.log` |
| 显式开启 Pi 的 Jev classify/catalog 差分 | 12 PASS，pending 0；退出 0 | `runs/jev-differential/`、`jev-differential.log` |
| 新增 plain/hooked 公共行为与拒绝顺序 | 98 PASS；退出 0 | `runs/supplement/`、`supplement.log` |
| Oracle 错 version/commit/model-data 启动拒绝 | 3 PASS；退出 0 | `oracle-startup.log` |
| 证据接缝先红 | 1 FAIL：验证器尚不存在；退出 1，审计零发现 | `runs/artifact-red-clean/`、`artifact-red-clean.log` |
| 初次进程诊断 | 验证器不存在；错误文本包含本机路径，未通过审计 | 只保留脱敏后的 `artifact-red.log`；不交付泄漏路径的原包 |

上述 2 个 NOT_RUN 是未启用差分的辅助目录检查；后续显式差分另记自己的当前结果，不回写旧状态。655 个固定 case 中，652 PASS、2 NOT_RUN、1 预期红灯 FAIL。新增基线验证器的绿灯和完整最终验证见 `validation.json`，不会重新生成固定包。

新对照测试固定单候选、空 string/object/array State、精确大整数、object/array 说明、null choice 说明、并列最高概率的两个合法答案、可选/结构化 bool criteria、score 2–10、有界百分位舍入及高精度反例、duplicate/case-alias 和未知 Jev refusal 的整体拒绝。每个响应对照同时检查原始 native 请求、答案、usage/价格、ResponseModel、归属/快照、Hook 顺序、Attempt/Observer、Host resolver/credential 次数、body 和 permit 释放。Hook 换题同时证明最终题答案成功、入口旧题答案失败。

| 验收 | 实际公共证据 |
| --- | --- |
| OD01 | 原 `TestClassifierBuiltinOfficial` / Catalog 测试；`P07-OD01-jev-*` 合法旧请求和范围对照 |
| OD03 | `P07-OD03-jev-*` 的 plain/hooked 归属和授权顺序 |
| OD09 | `P07-OD09-jev-*` 的 plain/hooked scope 拒绝和零 resolver |
| OD16 | 全部新 wire 对照及 `P07-OD16-jev-final-*`；原 callback/numeric-precision 对照 |
| OD34 | 原 `P07-E02-live-score-*` 官方舍入 replay；新 `P07-OD01-jev-rounded-*` |
| OD50 | 已有 Jev 和 12 条本次差分；`P07-OD50-jev-baseline-*` 固定包的完整性与反篡改 |

OD09 对坏 State、空题/空名/nil question、256 choice、1/11 score 同时提供不存在的 Binding，要求 `invalid_request/scope` 且 resolver、credential、Hook、Admission、Attempt、request 全部为零。宿主目录仍可声明 MaxChoices=300、MaxScoreLevels=20，实际公共范围不因此扩大。OD03 的 foreign/disabled/operation/provider/protocol/allowlist/catalog 拒绝固定 Code/Phase 和一个 resolver、零 credential/Attempt/request。

`manifest.json` 是独立 SHA-256 文件清单，包含压缩包哈希、每次执行的身份、版本和原始状态；`verify.py` 固定其哈希，拒绝将新 capture 与重新计算的 manifest 冒充原包。归档中保留每 case 的 fixture、Provider 脚本与捕获、结果、Attempt、Observer、assertions 和精确重放命令。整个归档在压缩前逐文件脱敏审计；ZIP 成员只有普通文件，解压字节和文件数量有界。

源码身份在 `initial-inventory.json`、`inputs.json`；生产代码、oracle pins、Jev fixture 及来源证明按原字节保存。完整测试源码与故意包含 key 形状的攻击 fixture 以 Hash 标识，其执行输入/捕获保存在各自已经脱敏的 case 目录，未放宽审计规则。归档文件的 SHA-256 固定实际交付字节，case 的原 fixture Hash 与脱敏后 artifact Hash 分开保留。

唯一运行 oracle 仍为 Pi 1.1.0，commit `abe508e1b89912adde45528136c3221eb69acdd7`，43 文件模型数据 Hash 为 `8a06cf635be3392b977f7bb328e7ba5065cbe55825d997ca7d00309fedcdced8`。原 lock/integrity、安装副本的版本与 43 文件实测字节均在包内。2026-10-11 两份源码等价证明的原 SHA256SUMS 本次重新核验；它们是历史重建，不宣称本次重新下载或重建。

包内 `history/` 保存 2026-10-11 原始发布 FAIL 报告及压力缺陷描述；该归档不是当前 live/effect 状态。2026-10-12 的本地代理重测已在本票起点 HEAD 之前提交：[Anthropic](../../barness-ai-thinking-sampling/anthropic-rerun-20261012/summary.json) 11 PASS；[TypeSafe](../../barness-ai-thinking-sampling/typesafe-rerun-20261012/summary.json) 3 PASS，独立中文评估 COMPLETE、24/24 正确；[其余五条路线](../../barness-ai-thinking-sampling/remaining-proxy-rerun-20261012/summary.json) 28 PASS、1 个既有协议 UNSUPPORTED。当前九路线矩阵没有 FAIL/NOT_RUN，不能再将旧 live/effect 失败描述为当前阻断。

代理重测各报告明确记录完整 releasegate 尚未重新执行；原发布 FAIL 是历史判定，不据专项通过推断整版 gate 已 PASS。压力流完整性异常的根因仍由[原缺陷工单](../../barness-ai-thinking-sampling/issues/07-pressure-stream-integrity.md)追踪。本票自身执行的是 Jev 离线基线，没有额外真实调用；Decisions 差分由工单 07 负责，第十组合发布 gate 由工单 13 负责。

从仓库根目录验证固定包（不产生真实调用）：

```sh
python3 .scratch/barness-ai-openai-decisions/jev-baseline/verify.py --audit
go test ./ai/e2e -run '^TestJevBaselineEvidence$' -count=1
```

重新执行当前工作树的 Jev 对照，输出目录必须全新，固定包不被覆盖：

```sh
npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node
python3 .scratch/barness-ai-openai-decisions/jev-baseline/replay.py \
  --out .evidence/jev-baseline-fresh
```

runner 顺序执行已有测试、Jev 差分、98 个补充对照、基线反篡改、oracle pin 拒绝、相关 race 和两种 vet。首次固定 supplemental 的 evidence 环境值相对 Go package 目录，原包收集到根输出目录但字节与原 manifest 不变；后续 runner 使用绝对输出目录。

实现 commit `7939308` 已按 code-review 做两项独立审查，Standards 0、Spec 0；审查者另行核验固定包与审计。完整报告见 [REVIEW.md](REVIEW.md)。

最终验证命令如下；日志和命令退出码独立记录在 `validation.json` 与 `validation.zip`。完整套件未开启设计压力/真实调用；相关 race 不宣称全仓库 race 或完整发布 gate。

完整套件退出 0，用时 239.134 秒，6,227 case（6,218 PASS、9 NOT_RUN），其中 815 差分全部 PASS；相关 race 20.320 秒、394 PASS。`replay.py` 的整条新目录流水线也实际执行并退出 0。`validation.zip` 保存最终日志、原 run manifest/audit、4 个 artifact 绿灯 case 和重放收据；完整最终普通/race case 捕获仍在收据所指的本地 `.evidence` 目录，固定基线中的原 Jev 捕获独立完整。validation 的 SHA-256 清单保护其实际交付字节；它不替换基线固定 manifest。

```sh
go vet ./...
go vet -tags live ./...
BARNESS_AI_PIDIFF=1 go test -race ./ai/e2e \
  -run '^Test(JevRegression|Classifier|Unary|JevBaselineEvidence)' -count=1 -json
BARNESS_AI_PIDIFF=1 go test ./... -count=1 -json
```
