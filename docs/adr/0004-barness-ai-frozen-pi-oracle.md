---
status: accepted
date: 2026-10-01
---

# barness-ai 以 npm 发布件重建冻结 pi，并以分类账判定差分

规范要求同一批 E2E 场景与冻结 pi-ai `1.0.0`（commit `a13d35a742c6`）比较请求和结果，且 CI 不依赖开发者本机研究目录。冻结 commit 缺少 `src/providers/data` 生成数据，`generate-models` 抓取线上目录，结果随时间漂移。barness 在 `ai/internal/testkit/pioracle/node` 中按 lockfile 固定 npm `@earendil-works/pi-ai@1.0.0`，并由 Go 测试以子进程驱动 Node.js 运行器；差异经版本化分类账逐项归类，待处理项使对应协议的差分门禁失败。

## Considered Options

- 引用研究目录：零准备，但违反“CI 不依赖本机研究仓库”，且该目录缺生成数据。
- 在 barness 内 vendor 冻结源码并自建：可完全控制，但要引入整个 monorepo 构建链，数据仍需外部来源。
- 重新生成模型数据：可从源码重建，但数据为生成时点快照，无法复现发布时的目录与价格。
- 固定 npm 发布件（采用）：1.0.0 发布件 ai / telemetry 的 `gitHead` 与 tag 都是 `a13d35a742c6`。冻结源码填入发布数据后 ai 811/811、telemetry 24/24 个 dist 文件逐字节一致。来源、逐文件哈希与两个干净环境的复现见工单 02 的 release-verification。旧 0.87.1 基线及其运行副本已由工单 03 替换。

## Consequences

- 重建只需 `npm ci --ignore-scripts`；运行时无网络。运行器空环境启动、拒绝凭据/endpoint 变量、套接字限于 loopback、key 仅来自用例；每次运行核对版本与模型数据哈希。来源与核验步骤见 `node/PROVENANCE.md`。
- 比较不删除字段：仅对象键语义排序，数字按精确值比较，只对显式的消息时间路径做一一映射。barness 结果投影到 pi 形状时保留全部序列化字段；Code/Phase 等 Go/租户扩展不进入 pi golden。
- 分类只有已修复/扩展/待处理三类；未登记差异视为待处理，“已修复”项再次出现视为回归。扩展须引用规范条款或明确批准，不能由实现者推断。
- 差分依赖 Node.js 与已安装副本，因此通过 `BARNESS_AI_PIDIFF=1` 显式开启；未开启时证据记为 NOT_RUN，普通 `go test ./...` 保持离线。维护者于 2026-10-01 决定首期保持显式开启：协议/公共语义变更由维护者显式运行受影响差分，待处理项按分类账跟踪到对应工单。这偏离规范 §6“常规门禁执行受影响的确定性差分”的字面要求；接入 CI 并转为常规门禁时另行决定，并更新本 ADR。
- 更换冻结版本即更换兼容基线：先改规范，再重做等价核验、更新 `provenance.json` 与分类账。


## 1.0.0 的 SDK 错误边界（2026-10-08，工单 03）

全量差分发现 OpenAI Node SDK 7.19.0 改变两类错误文本，维护者明确批准同步：HTTP 对象／数组缺少或含 null error 时，SDK 将整个值作为错误体；具名 SSE event:error 在 pi adapter 之前按 data.error ?? data 抛错。后者适用于所有 JSON 类型，不能只改 Responses 的 type:error 分支，也不能先按 Chat chunk 对象解码。

Responses 与 Chat 均使用 Go SDK 的原始 SSE decoder 保留事件名，在协议 DTO 解码前处理 SDK 错误规则；未具名 type:error 继续由 Responses adapter 处理，未具名 null/false error 不作为 SDK 错误。SDK typed stream 只处理初始请求与响应体关闭，避免另造 SSE 解析器或共享可变解析状态。错误文本仍经过已有脱敏、截断与分类边界。所有对齐项登记为 fixed，并以 HTTP 形状、具名／未具名帧及非对象错误帧的公共入口和冻结 pi 差分验证；未增加 extension。

## 1.1.0 实施基线（2026-10-11，ADR-0025）

唯一运行 oracle 替换为 pi-ai 1.1.0，固定 commit `abe508e1b89912adde45528136c3221eb69acdd7`。npm AI/telemetry 缺 gitHead，provenance 留空；官方 tag SHA、固定 codeload archive 哈希及两个干净重建的 AI 827/827、telemetry 24/24 dist 零差异证明来源，不推断缺失的 npm 字段。43 文件数据哈希 `8a06cf635be3392b977f7bb328e7ba5065cbe55825d997ca7d00309fedcdced8`，Anthropic Node SDK 为 0.129.0，另两 SDK 保持固定。版本、commit、数据错配均在启动前拒绝。

当前六个完整差分子集 814 条、pending 0，全部既有型号字段核对通过；来源与回放见[核验材料](../../.scratch/barness-ai-thinking-sampling/release-verification/README.md)。历史 1.0.0 记录保持原事实，运行时不保留第二套 oracle。整版正式发布必须通过 ADR-0017 的完整合同；当前 live/effect 未通过，一次压力失败另记未确认原因的缺陷，不因实现基线已核验而宣布整版交付。

## Decisions 的第七差分子集（2026-10-13，工单 07）

P10 使用同一冻结 1.1.0 的 `openai-decisions.classify` 与
`OPENAI_CLASSIFIER_MODELS`，按 classifier 类型选取 gpt-6-luna；同名聊天型号独立查找。
版本、commit、43 文件 model-data pins 与 lockfile 不变，不另设 oracle。
增加对原 827 个 dist 文件的排序清单哈希
`39a9951108d0e7faf60886302ce4bc953d8c51d5ebad22370301f23b29aa2504`，避免只校验目录却允许 API 源文件被替换。
进程保持空环境，socket 限定到本用例的 loopback HTTP host/port，拒绝转往其他本地端口。

完整 wire、命名答案、cost、score 分布/图例、refusal 与 usage presence 保留在比较中。
差异只按有限 case/path/kind 登记；每个扩展在自身公共 E2E 中用独立 literal 断言，
共同 model 字段与未登记 case 的负例继续失败。P07 原投影与断言保持；完整 Jev 目录序列化增加回归。
数值名称按名称匹配、响应数组乱序重放，不以 index 当身份；数字顺序与基本答案实际一致。

P10 在 P0、Required differential 与批准证据选择器中独立必需，属于冻结 Pi 原生路线。
记录包含 source/model-data/SDK lock/catalog、完整请求与结果哈希；导入门禁核对 observation 哈希。
本票没有正式 classifier 目录/live/effect 的发布结论；十组合完整门禁由工单 13 承接。
移除此路线时应撤销其发布要求与证据选择器的正式合同，不能以 NOT_RUN、Extension Route
或扩大账本代替成功差分。复现与完整捕获见[证据包](../../.scratch/barness-ai-openai-decisions/frozen-differential-evidence/README.md)。
