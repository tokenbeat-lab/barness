---
status: accepted
date: 2026-10-02
---

# barness-ai 真实 API 冒烟：每组合一进程、四值结果与只由本组合报告合并的支持矩阵

工单 23 交付发布前、SDK/模型升级后执行的真实官方 API 冒烟（spec Testing Decisions §2、§5）。spec 规定了双开关、六组合、场景、四值结果、记录项与支持矩阵字段，但没有规定入口形态、key 如何"仅注入对应 Provider×API 的测试进程"、重试预算、各组合用哪个模型证明哪些能力，以及矩阵放在哪里、如何更新。本 ADR 记录实现时的做法，维护者于 2026-10-02 确认（见文末）。

## 决策一：入口与 key 隔离

冒烟是 `ai/live` 包中带 `live` build tag 的测试，另需 `BARNESS_AI_LIVE=1`；无 tag 时 `go test ./...` 不编译它们，有 tag 无开关时全部 NOT_RUN。一个进程只跑一个组合（`BARNESS_AI_LIVE_COMBO`），key 只从该组合自己的变量 `BARNESS_AI_LIVE_KEY_<COMBO>` 读取，不读 `OPENAI_API_KEY` 等厂商变量或文件；进程中出现其他组合的 key 变量时整组 FAIL（类别 config），因为这说明 secret store 没有按组合分进程注入。有 key 时必须给出 `BARNESS_AI_LIVE_ACCOUNT_ALIAS`（测试账户/区域别名），矩阵靠它说明是哪个账户跑的；缺失同样 FAIL。各组合的模型在套件中固定，不提供覆盖：换成缺少能力的模型会把能力缺口变成 UNSUPPORTED 而使该组合"全部通过"。

Client 按宿主方式装配：一个租户、一个指向厂商真实 endpoint 的 binding、一个凭据、有限资源策略（测试值）；唯一的测试介入是 `Config.Transport` 上的记录器，它包住与 Client 默认相同的 transport（不读代理环境变量）。不提供 endpoint 覆盖：冒烟只证明真实服务。

### Considered Options

- 单进程持有六个 key、依次跑六组合：违反 spec "真实 key 仅向对应 Provider×API 的测试进程注入"。
- 显式命令入口（`go run`）：与 evidence 包的用例/重放机制脱节，需另写报告装置。
- build tag + 开关 + 每组合一进程（采用）。

## 决策二：场景与模型

每个组合跑六个共同场景：`text-stream`（Stream，完整选项）、`text-complete`（CompleteSimple）、`tool-round-trip`（Stream 调用工具 → 测试宿主经 `ValidateToolCall` 校验并回结果 → Complete 生成答复，答复轮关闭工具）、`cancel-after-first-frame`（start 之后第一个事件到达即取消）、`reasoning-history`（推理开启时调用工具，下一轮原生回放推理/签名与工具结果）、`image`。断言结构（start 在先、唯一终态在末且与 Result 一致、块按索引开闭）、关联（toolcall_end 与消息中的调用一致、答复请求带回调用 id、回放请求带回推理值）、终态、非空文本、归属与 usage，不断言生成文本。

默认模型选目录中成本低且具备首期承诺能力的模型：gpt-5-mini（OpenAI 两协议，非推理场景用 minimal）、claude-haiku-4-5、gemini-2.5-flash、deepseek-flash（DeepSeek 两协议）。P01/P04/P02/P03/P06 强制工具调用（P06 关闭 thinking，spec P06）；P05 的冒烟要求不含强制，用自动选择加指令性提示。OpenAI × Chat 的 `reasoning-history` 记为 UNSUPPORTED：该协议不返回可回放的推理内容或签名，推理只计入 usage。模型不收图片时 `image` 为 UNSUPPORTED（占位降级由离线用例覆盖）。

工单 21、22 留给冒烟的事项各为一个场景：Chat 两组合的 `usage-position`（usage 在 finish_reason chunk 内还是其后独立 chunk、是否早于 `[DONE]`）；DeepSeek Chat 的 `forced-tool-with-thinking`（厂商以 400 拒绝或照做都通过并记录，静默忽略则 FAIL，对应 ADR-0015 决策三的条件）、`prompt-cache-long-retention`（被拒则 FAIL，对应决策四）、`deepseek-v4-pro` 的 `reasoning-level-high/max`；DeepSeek Responses 的 deepseek-flash `reasoning-level-high/max`（被拒则 FAIL，对应 ADR-0014 决策三）；两个 DeepSeek 组合的 `auth-refused`（固定的伪造 key，记录 401 错误体形状与 id 类响应头）。推理项、函数调用项 id 等线上形状由记录器保存在每个场景的证据中，供修正 P05/P06 fixture；冒烟不断言这些形状。DeepSeek 的 402/429/503 错误体无法安全地主动触发（需要欠费、限流或厂商故障），只有自然出现时才被记录；`auth-refused` 是向真实厂商发送的唯一一次故障请求。

## 决策三：四值结果、重试与预算

结果只有 PASS/FAIL/NOT_RUN/UNSUPPORTED。缺 key 或未选中的组合为 NOT_RUN（测试 skip，证据记 NOT_RUN）；UNSUPPORTED 由 evidence 新增的 `Case.Unsupported` 记录，既非 skip 也非 PASS。

`rate_limited`、`upstream_error`、`transport`、`deadline_exceeded` 视为厂商或环境故障：在任何断言之前结束该次尝试，按 5 s 递增退避重试，每场景最多 1 次；仍失败则 FAIL 并标 `environment`，错误类别为该 Code。`upstream_auth` 同属环境故障但不重试（换不出另一个 key）。其他错误与断言失败为 FAIL，不重试。每进程最多 24 次逻辑调用（最大组合正常 14 次），每次输出上限 4096 token；预算耗尽的场景 FAIL（类别 budget）。报告记录预算与实际用量、每场景的模型、SDK、厂商 request id（barness 读取的头缺失时取厂商其他 id 类响应头）、耗时、调用与重试次数、错误类别和观察说明。

## 决策四：证据与脱敏

记录器在来源处脱敏：凭据请求头替换为 `[REDACTED]`，key 形状的文本（含 DeepSeek 401 中的掩码后缀）替换，响应头只保留 id、content-type、retry、rate-limit 类的值，其余只记名称（不保存组织、项目等账户信息）；请求与响应正文各有捕获上限。真实 key 注册给 evidence 的 `RedactSecret`，进程结束时再扫描整个证据包，发现 key 即令进程失败。证据包另含 `live-report.json`，即该组合的报告。

## 决策五：支持矩阵

矩阵是检入的 `ai/live/support-matrix.json`，每个首期组合一行：Provider、协议、模型、SDK、账户/区域别名、最后执行与最后全部通过时间，以及每个场景（能力）的模型、结果、错误类别、说明、最后执行与最后通过时间。它只经 `go run ./ai/live/cmd/supportmatrix <证据包>...` 合并报告而改变，合并规则（`ai/internal/testkit/supportmatrix`，先列失败方式再实现）：报告只能改自己组合的行，Provider 或协议与该行不符即拒绝，共享 adapter 的通过不会写到另一行；全为 NOT_RUN 的报告不改变矩阵，单个 NOT_RUN 场景保留该能力上次状态；FAIL 不清除上次通过时间；早于该行最后执行时间的报告被拒绝；含配置失败（类别 config，进程未发出请求）的报告被拒绝，不覆盖真实结果；任何一份被拒，矩阵文件不变。报告列出该组合应有的全部场景（`expected`），行的 `allPassedAt` 只在一次报告覆盖全部应有场景且均为 PASS 或 UNSUPPORTED 时更新（以 `-run` 缩小的运行不会置位），是工单 24 发布门禁读取的值。

## Consequences

- 首次检入的矩阵六行均未执行；本工单在开发环境中没有真实 key，未运行真实冒烟。冒烟装置的执行路径（全部场景通过、503 重试后通过、持续 503 的环境 FAIL、强制选择被静默忽略的断言 FAIL、401 记录、脱敏审计、报告合并）用指向本地 TLS 替身的临时文件验证过，替身未检入。
- 工单 21、22 的待确认事项在维护者以真实 key 运行后才有结论；按 ADR-0014、0015 的既定条件修改目录、adapter 与 fixture。
- 证据显示 openai-go 自行发送 `X-Stainless-Os`、`X-Stainless-Arch` 等头，与 `userAgent` 注释"不携带宿主 OS 与架构"的意图不一致；本工单不改，另行登记（工单 31）。
- 仓库尚无 CI；"每组合一进程、secret store 只向本组合进程注入 key"由进程侧检查与 `ai/live/doc.go` 的运行说明承担，CI 作业接入时照此配置。
- 冒烟默认模型或目录变化时，更新 `ai/live/combos_test.go` 并重跑受影响组合。

## 维护者决定（2026-10-02）

1. 决策一：采纳。一个进程只跑一个组合，key 只从本组合变量读取；进程中出现其他组合的 key、或有 key 而缺账户/区域别名，整组以 config 失败，报告不合并。本地手动运行同样只设一把 key。
2. 决策二（模型与场景）：采纳。默认模型 gpt-5-mini、claude-haiku-4-5、gemini-2.5-flash、deepseek-flash（DeepSeek Chat 的等级检查用 deepseek-v4-pro）固定在套件中，不提供运行时覆盖；首次真实运行前由维护者核对这些型号在测试账户中可用、花费可接受，需要换型号时改 `ai/live/combos_test.go`。各组合的强制/自动工具选择按决策二。
3. 决策二（OpenAI × Chat）：采纳。该组合的推理回放记为 UNSUPPORTED：协议不返回可回放的推理内容，OpenAI 的推理回放由 Responses 组合证明。
4. 决策二（`auth-refused`）：采纳，附条件。仅两个 DeepSeek 组合以固定伪造 key 各发一次 401 请求，其他组合不加。DeepSeek 的 401 错误体与请求 id 头经真实运行确认、P05/P06 fixture 据此修正后，删除该场景，此后不再向厂商发送故障请求。
5. 决策三：采纳。每进程 24 次逻辑调用、每场景最多重试 1 次（5 s 递增退避）、每次输出上限 4096 token。若真实运行中厂商偶发故障造成的误报 FAIL 过多，可把每场景重试提高到 2 次；不再放宽，spec 禁止无界重跑。
6. 决策五：采纳。矩阵位置 `ai/live/support-matrix.json` 与合并规则不变。
7. SDK 自带的宿主信息请求头（Consequences，工单 31）：去除。与 `userAgent` 不携带宿主 OS、版本与架构的决定（2026-10-01）一致，在 adapter 构建请求处统一删除 openai-go 与 anthropic-sdk-go 发送的 `X-Stainless-*` 宿主信息头，离线 E2E 断言服务端收不到；与 pi 的差异登记为差分账本扩展。实施归工单 31。

## 首次真实运行后的型号调整（2026-10-02）

首次运行时 Google 对 `gemini-2.5-flash` 返回 404：“no longer available to new users”，建议改用 `gemini-3.8-flash`。按决策二“需要换型号时改 `ai/live/combos_test.go`”，Gemini 组合的默认模型改为目录中的 `gemini-3.8-flash`（Gemini 3：level 推理、调用 id 与签名回放、嵌套的工具结果图片），价格为每百万 token 输入 0.75、输出 3.75 美元。其余默认模型首次运行可用。

`auth-refused` 场景按维护者决定于 2026-10-03 删除（401 错误体与请求 id 头已确认，工单 33）；支持矩阵中 DeepSeek 两行保留该能力的历史记录，合并规则不删除能力。

## 工单 08：分类组合与报告 schema 2（2026-10-08）

新增 classifier × typesafe × typesafe-system-one 独立进程；固定官方 endpoint 与
jev-1.13.0，双开关与本组合凭据变量不变。报告和矩阵的 schema 升为 2，增加
operation；六个历史行一次性注明 chat，保留其历史时间与能力，不伪造新执行。
旧 schema 报告拒绝，没有运行时兼容分支。固定组合身份独立校验，报告与矩阵同时错也不能通过。
分类完整集合是 mixed-questions / single-choice / context-422；只能使用固定型号，
必交能力不能降为 UNSUPPORTED。按调用、问题、提交状态字节预留预算：6 次、10 个、
每次 131072 字节；每场景最多 1 次环境重试。记录实际 HTTP 尝试、各场景问题和请求 ID。

context-422 是有界形状核验：40,000 个合成词经公共 Client，本地结构和字节校验通过，
厂商实际返回 HTTP 400，detail.error_type=max_tokens_exceeded，分类 invalid_request/request。
仅此明确上下文形状可说明探针完成；备注必须写 422 UNCONFIRMED，不声称观察到 422。
任意其他 400 失败；不扩大输入、绕过入口或改发空问题。此项修正实施设计的 422 假设，
按工单“无法预算内确认时明确未确认”处理。空问题由离线公共入口拒绝、零网络请求证明。

本次账户别名 local-typesafe@region-unreported 来自用户提供的 .env Key；
账户区域与权限级别未核实，不据此宣称该账户符合低权限要求。runner 只提取 TYPESAFE_KEY，
不 source 整份 .env；子进程只注入本组合 Key 和运行所需环境。Key 不写报告或命令行。
最终状态只能由自己的脱敏报告合并；NOT_RUN 保留有效历史，拒绝批次不部分写入。

## 工单 11：P08 原生 OpenAI Images（2026-10-08）

沿用每组合一进程、双开关、唯一 Client、固定官方 endpoint、组合专用凭据与现有
报告/脱敏审计/原子矩阵合并，增加 image × openai × openai-images。固定日期快照为
`gpt-image-2.5-sunburst-2026-09-08`。第一场景必须是内联 JSON 编辑；仅该进程的
编辑 PASS 才执行生成和 mask，窄化执行也不能绕过。JSON 编辑和生成不得报告
UNSUPPORTED；mask 只有 HTTP 400 明确表示 mask 不支持时可记 UNSUPPORTED。

独立预算为 4 次公共逻辑调用、4 张请求图片、每次一张、1024×1024；失败请求与
环境重试先预占调用/图片预算，实际 HTTP 尝试与报告用量逐项累计，耗尽 FAIL。
Binding.Retry 保持关闭，避免内部尝试绕过预算。mask 也占请求预算。
报告/矩阵沿用 schema 2，预算增加可选图片计数与分辨率，场景增加可选 ImageCalls
（路径、状态、请求 ID、格式、读图校验、用量完整性）；旧分类/聊天语义不变。
合并校验以固定预期能力验证，不能自行缩小/重复预期列表；固定操作、Provider、
协议、目标和型号都须相符，计数、顺序与实际请求证据完整才写完整通过时间。
共享 adapter 和其他操作的通过不能给 P08 背书。

2026-10-08 的自身实测三个场景 PASS：PNG JSON 编辑、JPEG 生成、WebP mask 编辑
（透明背景），均完整解码为 1024×1024。一共 3 次请求、3 张图、0 次重试，usage
均 complete。账户别名 `dotenv-openai@region-unconfirmed` 记录凭据来自宿主 `.env`；
权限级别和区域未由 API 响应证明，别名不等于低权限认证。正式 CI 仍须配置专用
低权限凭据及真实区域别名。子进程只收到 OPENAI Images 的 key，不读取 `.env`。

[真实报告与校验](../../.scratch/barness-ai-pi-1.0/openai-images-live-evidence/README.md)
包含原始有界脱敏捕获、公共结果、Observer、图片文件/SHA-256、零发现审计及离线
反例进程。PNG 响应超 512 KiB 捕获上限，截断明确标记；图片与用量以 Client 完整
结果保留。其 fixture 归一化明确记录此来源，不能作为另一条 live 接受性证据。

## 工单 14：Google 图像真实证据（2026-10-08）

加入第九组合 image/Google/google-interactions（google-interactions-image），固定 v1beta 与 Nano Banana 2.1。两项必交能力不能 UNSUPPORTED；每进程至多四调用/四图片、1K/1:1，失败与重试计预留。响应可能无 ID/请求 ID 头，只报告实际存在的型号；不能凭共享 Google 聊天通过声明支持。Google 报告必须完整覆盖两项，错身份、旧报告、配置失败、缺形状或预算证据原子拒绝。详见 ADR-0023 与工单 14 交付证据。
