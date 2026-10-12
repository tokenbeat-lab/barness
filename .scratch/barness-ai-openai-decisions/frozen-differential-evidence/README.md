# 工单 07：冻结 Pi Decisions 差分证据

实现 `dbdd77c`，审查修复 `6acff6d`，起点 `724f09e`，当前分支 main。
本票交付 Classify/Catalog 的 P10 确定性差分与有限账本；使用合成 Host classifier
目录，不提前声明正式目录、live、压力或业务效果通过。其他用户的工作区改动保留。

唯一 oracle 仍为 Pi 1.1.0 / `abe508e1b89912adde45528136c3221eb69acdd7`。
43 文件 model-data hash `8a06cf635be3392b977f7bb328e7ba5065cbe55825d997ca7d00309fedcdced8`；
原 827 dist 文件的排序清单 hash `39a9951108d0e7faf60886302ce4bc953d8c51d5ebad22370301f23b29aa2504`；
未更换 SDK lock，checksum `e0b4084c372bc764b73faa25f2a311a3649e5b7ffb341352f25c13a891b7de65`。
源码和完整 npm lock 由仓库提交与 source-snapshot.json 固定，node 依赖可用原 lock 重建。

P10 使用冻结自身 openai-decisions.classify 与 OPENAI_CLASSIFIER_MODELS；同名 chat Luna
保持独立。空进程环境、显式 fixture key、loopback HTTP host/port 固定，转往其他端口被拒绝。
9 个 Node 反例覆盖 version/paired version/commit/model data/API source 篡改、分类/聊天目录、
旧聊天 Costs、真实受控 HTTP 分派、泄漏环境与 redirect。

| 验收 | 实际证据 |
| --- | --- |
| OD14 | numeric-looking 名称/候选按 ECMAScript 顺序冻结，string true/false 保持字符串；答案数组乱序，按名称匹配 |
| OD48 | 46 P10 classify/catalog 差分全 PASS，完整 wire/basic answers/source/hash/隔离；P10 为第七 Required 原生路线 |
| OD49 | 49 有限 case/path/kind 账本记录；独立字面值断言图片、精度、score、refusal、presence/total/reasoning、严格校验、headers/diagnostics/callback/retry；共同 model 变更及未登记 case 仍 pending |
| OD50 | P07 原 E2E、12 个 diff、官方 replay 和目录断言保持；新回归检查整个旧 Jev 序列化，没有整个 catalog/metadata 豁免 |

P10 投影保留 score 详情、refusal、cost 和 usage presence。wire/基本答案共同字段继续比较；
图片与 numeric 名称排序实际无差异。State string 的 quotes、原始 JSON 大整数、
逐题拒答、score 详情、严格校验、完整用量、最终 callback 题目、504 显式重试及
Go/fetch 头与诊断差异按有限 case 登记。每个 case 保留两侧完整 captures、原始脚本、
公开 Barness result、Observer 元数据、资源释放、assertions、实际 verdict、request/result
hashes、source/SDK/catalog hashes；没有协议级 extension。

首次公共红色 45 FAIL 保留原状。首次全套退出 1，旧 TestCostEstimate 发现按 ID 查价选中
同名 classifier；明确 chat 查询语义后修复。审查的 paired-version 与 SDK missing/wrong
反例先红后绿；原始收据与两轴审查见 validation.json、REVIEW.md。

最终全套退出 0：7268 个 E2E case，7259 PASS、9 个既有压力 opt-in NOT_RUN；
861 个差分全部 PASS，pending 0。七子集：P01 335、P02 70、P03 77、P04 203、
P06 118、P07 12、P10 46。专项 race 347 PASS；Node 9 PASS；两种 vet 和独立 cost/
release/oracle 检查通过。Standards 0、Spec 0。

实际门禁导入全套包，differential 与审计、目录/发布映射 snapshot 为 PASS；
仅提供 P10 时门禁继续要求原六子集。bundle-only 评估没有执行其 command/live/
pressure/effect 门禁，整体报告保持 FAIL，不把本票差分通过当完整发布。

`differential.zip` 保存红色/绿色/race 全部捕获；全套保存全部七个差分、全部 P07、
旧聊天目录核对的完整捕获，其他离线 case 保存整个 manifest/audit。大型全套原始
捕获在本地 `.evidence/decisions-07-owned/decisions-diff-full-reviewed/` 下保留，重放
命令会生成新的全套包。source-snapshot.json 固定 531 个代码/fixture 输入哈希；
installed-dist.sha256 固定 827 个源等价文件。所有 ZIP 成员均有 manifest SHA-256，
外层 SHA256SUMS 固定 archive/manifest/receipts/docs。Node 红色日志中的本机路径
替换为 workspace 标记，fixture 数据和实际运行结果保留。archive-audit.json 是对
最终 ZIP 解包后所有已交付文本/捕获的实际脱敏审计，findings=[]。

核验固定包（不访问厂商）：

```sh
(cd .scratch/barness-ai-openai-decisions/frozen-differential-evidence && shasum -a 256 -c SHA256SUMS)
unzip -t .scratch/barness-ai-openai-decisions/frozen-differential-evidence/differential.zip
```

仓库根目录重放当前源码；每例 assertions/manifest 也保存准确命令。以下使用根目录
绝对 evidence 路径，避免在 package cwd 生成未跟踪目录；不会覆盖已提交的固定包。

```sh
npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node
node --test ai/internal/testkit/pioracle/node/runner.test.mjs
BARNESS_AI_PIDIFF=1 BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/decisions-07-replay" go test ./ai/e2e -run '^Test(DecisionsPi|JevDifferential)' -count=1
BARNESS_AI_PIDIFF=1 BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/decisions-07-race" go test -race ./ai/e2e -run '^Test(DecisionsPi|JevDifferential|Classifier|JevRegression)' -count=1
go vet ./...
go vet -tags live ./...
BARNESS_AI_PIDIFF=1 BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/decisions-07-full" go test ./... -count=1 -json
```
