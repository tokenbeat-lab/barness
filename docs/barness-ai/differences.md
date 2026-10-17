# barness-ai 差异登记

本文登记 barness-ai 相对当前实现基线 pi-ai `1.1.0`（commit `abe508e1b89912adde45528136c3221eb69acdd7`）的全部可观察差异。spec Testing Decisions §5 只允许三种处置：**已修复**、**本规范定义或后续明确批准的扩展**、**待处理**；待处理阻断对应协议发布。逐条判定的权威记录是差分账本 [`ai/e2e/testdata/pidiff/ledger.json`](../../ai/e2e/testdata/pidiff/ledger.json)（每条带协议、路径、适用用例、处置与依据）；本文按主题归纳它，并补上不经差分比较的租户扩展。

发布门禁（`go run ./ai/release/cmd/releasegate`）读取账本：存在 `pending` 决定、任一差分用例有待处理发现、或扩展路径未登记，差分门禁即失败。

## 1. 账本概况

| 协议 | 扩展 | 已修复 | 待处理 |
| --- | --- | --- | --- |
| openai-responses（P01） | 61 | 25 | 0 |
| anthropic-messages（P02） | 48 | 4 | 0 |
| google-generative-ai（P03） | 32 | 0 | 0 |
| openai-completions（P04、P06） | 39 | 7 | 0 |
| typesafe-system-one（P07） | 25 | 0 | 0 |
| openai-decisions（P10） | 49 | 0 | 0 |

“已修复”条目只记录曾出现并已修复的差异；同一差异再次出现即为回归，按待处理计。扩展路径（`routes`）：DeepSeek × Responses、原生 OpenAI × Images。

## 2. 规范新增决策

| 决策 | 差异 | 依据 | 验收 |
| --- | --- | --- | --- |
| D1 显式有限资源策略 | pi 无字节、队列、并发与时限策略；barness 构造时必须提供有限策略，超限以 `resource_limit` 结束（限额内与 pi 一致；超限的帧/输出/工具 JSON/错误体差分登记为用例级扩展），超限的尝试不重试 | spec I9；ADR-0002、ADR-0007、ADR-0008 | E08、D1-construct-*、E08-policy-pressure-* |
| D2 不一致配置快照直接失败 | 原研究允许重新解析；barness 在两次解析冲突时发请求前失败（`credential_unavailable`/consistency），不在同一逻辑调用内重解析 | spec I3；ADR-0003 | E07 snapshot-*、*-between-reads |

## 3. 扩展接入路径

**DeepSeek × Responses（P05）**：冻结 pi 只经 openai-completions 路由 DeepSeek。barness 复用 Responses adapter，按 Provider 确定能力：不发送 `store`、`include`、提示缓存字段与亲和头，显式服务等级在发送前拒绝，历史以完整输入回放，推理流为 `response.reasoning_text.delta/done`。以官方协议 fixture（`ai/e2e/testdata/deepseek-responses`）离线证明，**不计为 pi 差分通过**，登记于账本 `routes`（维护者决定 2026-10-02，ADR-0014）。

DeepSeek × Chat Completions（P06）是 pi 自身的路由，全部场景进入差分（ADR-0015）。

## 4. 已批准的扩展（按主题）

| 主题 | 差异 | 依据 |
| --- | --- | --- |
| HTTP 栈默认头 | Accept-Encoding、Accept-Language、Connection 等 Node undici 与 Go net/http 的默认值不同；不属于 pi 或协议逻辑 | 维护者决定 2026-10-01（各协议同） |
| SDK 运行时指纹与宿主描述头 | barness 发送自己的 `User-Agent: barness-ai`；不发送 SDK 描述宿主 OS、架构与运行时版本的 `X-Stainless-*` 头（pi 的 SDK 发送） | 维护者决定 2026-10-01/02；工单 31；ADR-0016 决定 7 |
| live partial | pi 的 `partial` 是共享的持续更新视图；barness 以受同步保护的 `PartialView` 提供快照，事件中的 partial 比较按此投影。两边序列化 partial 的时机不同（pi 单线程处理完整个 chunk 后才序列化事件），所以 barness 的视图可能领先或落后于 pi 在同一事件上的视图，但不会落后于正在读取的事件 | spec I8；pi types.ts；工单 24（Gemini usage 方向） |
| 工具参数原始 JSON | barness 在每个工具调用上保留 Provider 原始参数文本（`RawArguments`） | spec I6 |
| 错误文本脱敏（安全差异） | pi 原样复制 Provider 错误体；barness 脱敏其中 key 形态的文本 | spec I9；ADR-0009 |
| 运行时错误文本 | pi 报告 JavaScript `JSON.parse` 与 undici `terminated`/AbortError 文本；barness 报告对应的 Go 文本 | 各协议账本条目 |
| 原生状态担保 | 无可信封套的原生状态按跨模型规则降级（pi 原样回放） | spec I6；ADR-0001 |
| 缓存/亲和标识派生 | 模块构造的 prompt cache key 与亲和头按租户与账户派生，不直接发送裸 session ID | spec I9 |
| 资源限额 | 超过 MaxFrameBytes、MaxOutputBytes、MaxToolJSONBytes、MaxErrorBodyBytes 的用例以 `resource_limit` 结束；超限错误体（429/5xx）不重试而 pi 会重试 | D1；ADR-0007 |

## 5. 不进入差分、单独断言的租户扩展

以下行为不塞入 pi golden，由离线 E2E 单独断言（spec Testing Decisions §5）：租户/调用/尝试归属与事件信封（E10）、Code/Phase 错误分类（E02、E07）、用量完整性与价格快照标记（E11，ADR-0010）、资源限制与准入（E08，ADR-0007/0008）、并发与同步访问（E01、E06）、Observer 记录与脱敏（E09，ADR-0009）、绑定上的重试策略（E05，ADR-0006）、托管工具放行（E04，ADR-0005）。

## 6. 记录在案的有意差异与已知限制

| 项目 | 内容 | 依据 / 跟踪 |
| --- | --- | --- |
| `timeoutMs: 0` | barness 视为未设置；pi 会立即超时 | ADR-0008 维护者决定 |
| 失败响应携带的用量 | 与 pi 一样丢弃，尝试标为未上报 | ADR-0010 决策三 |
| Gemini 回调 | 回调看到 REST 请求体而非 pi 的 SDK 参数；不调用 onResponse（与 pi 同） | ADR-0012 |
| 流式工具参数解析时机 | pi 在每个 delta 上全量重解析；barness 在读取 PartialView 时才按当时的原始文本解析，读到的参数与 pi 相同，不读取的消费者不付解析成本（每个 delta 都读取则与 pi 成本相同） | 工单 32；ADR-0017 后续 |
| 未实现的可选特性 | 中途工具变更（Anthropic 原生 tool changes）、服务端备用模型（`fallbacks`）、OpenAI 的 additional_tools/tool search 未实现；开启它们的模型（claude-opus-4-8、claude-fable-5、托管强度的 claude-fable-5-1/opus-5/opus-5-5 与 gpt-5.4 起的十个 OpenAI 模型）照常列入目录，按 pi 中特性关闭时的行为发送当前工具列表、不发 `fallbacks`，pi 多出的占位工具、beta、`fallbacks` 与 `additional_tools` 条目登记为限于工单 34、27 用例的扩展 | 工单 26、28；ADR-0018 |
| 托管推理强度 | 与 pi 一致：每条消息记录本轮强度（`providerThinkingLevel`），按 Provider 回放强度标记；它不是原生状态，不需要可信封套，降级的同 Provider 消息仍保留标记 | 工单 27；ADR-0019 |
| SDK 响应体兜底 | SDK 在调用方 context 结束时不关闭恰好到达的响应体，barness 以中间件兜底关闭 | 工单 29 |
| 真实冒烟确认的 DeepSeek 行为（2026-10-02） | thinking 加强制工具选择被 400 拒绝；提示缓存字段被接受；Chat 的 usage 在 finish_reason chunk 内；high/max 推理等级被接受；错误体为 `error{code,message,param,type}`，key 只显示后四位；厂商请求 id 取 `x-ds-trace-id`（DeepSeek 不发送 `x-request-id`）；关闭思考时缺推理计数视为完整上报。P05/P06 fixture 已按实测形状修正 | 工单 23、33；ADR-0014/0015/0016 |

## pi-ai 1.0.0 迁移（工单 03）

该次迁移的唯一 oracle 为 1.0.0，commit `a13d35a742c6ef8462812a28fbe1d8c8b7431c32`；
发布件等价证明、逐文件数据哈希和安装锁见
[PROVENANCE](../../ai/internal/testkit/pioracle/node/PROVENANCE.md)。旧安装与运行路径已替换。

原基线的五项行为已改为 1.0.0：无效/非有限 Retry-After 退避；fast 服务等级计价；
Anthropic 增量覆盖 1h 明细；full/simple 在 builder 合并型号默认采样参数；
Responses 在 EOF 与终态检查之后核对所有工具块的 done（包括被覆盖的块）。
每组 fixture 写明原行为、新行为和冻结来源；差分账本中的新增项都是 `fixed`，
没有扩大现有 `extension` 的适用范围。重试等待由假时钟验证，pi 侧真实等待仍有明确时限。

内建目录升为 `2026-10-08.1`，原有字段及价格逐字段一致，新型号的纳入依据见
[ADR-0018](../adr/0018-barness-ai-catalog-inclusion-criterion.md#100-基线迁移工单-032026-10-08)。


维护者于 2026-10-08 批准同步差分发现的两类 OpenAI Node SDK 7.19.0 错误归一化：
HTTP 对象／数组缺少或含 null error 时保留完整错误体；具名 SSE event:error 按 data.error ?? data
在协议解码前抛错，覆盖对象和非对象 JSON。未具名 type:error 保留 Responses 的 Error Code 文本；
未具名 null/false error 不误报。来源、前后文本与最终验证见工单 03 的 chat-parity-evidence，
处理边界见 ADR-0004，新增登记均为 fixed。

## TypeSafe 混合问题与 1.0.0 差分（工单 06）

P07 是 pi 自身的 classifier 路由，进入必需差分。对象 state、字符串 instructions/choice criteria
以及基本 choice/score/bool 答案同场景比较；每个请求尝试保持固定 jev-1.13.0，pi 的 maxRetries 与 Binding 一致。
以下扩展按具体字段与用例登记，依据 [ADR-0021](../adr/0021-barness-ai-typesafe-unary-classification.md)：

| 扩展 | 边界与证据 |
| --- | --- |
| typesafe-native-json-input | string/array state、object/array instructions/criteria、choice null、JSON 原始数值精度；mixed native fixture 单独断言，不能算 pi 差分通过 |
| typesafe-score-details | score 完整分布及原生 JSON 图例保留、按最终等级校验；pi 仅给 score/confidence，差分投影明确移除这两个字段 |
| typesafe-strict-answers | 完整键集、概率范围/和、期望与图例一致性，任一失败整体清空答案并保留用量；三项真实差分用例证明 pi 接受而 barness 拒绝 |
| typesafe-versioned-model | pi jev-latest 通过 modelPatch 固定到本次 jev-1.13.0；目录别名单独登记，不把 ResponseModel 当授权 ID |
| typesafe-input-pricing | pi 直连 Jev 费率为零；barness 按宿主快照只计输入，差分双方显式移除 usage.cost。工单 08 已以官方事实和本路线真实证据纳入固定版本，输入每百万 token $0.042、输出免费 |
| transport metadata | Node fetch 与 Go net/http 的 Accept、Accept-Encoding、Accept-Language、Connection、Sec-Fetch-Mode、User-Agent 六个字段逐项登记，限定 P07 差分用例 |

classifier 目录逐字段比较 pi 拥有的共同字段，固定 ID 与价格的扩展明确登记；
baseUrl 归属 Binding，pi 的 input=[text] 对应分类隐含模态，两者显式断言并按字段登记，目录比较也产出 pidiff.json。
未知 pi 目录字段会失败，没有整个目录、请求或结果的豁免。证据及复现见
[工单 06 验证记录](../../.scratch/barness-ai-pi-1.0/typesafe-mixed-evidence/README.md)。


工单 08 的真实 score 证据说明期望需要有界两位小数舍入（ADR-0021），
概率和仍是 1±1e-6，保留原值；未放宽差分的删除字段或判定账本。
真实 400 max_tokens_exceeded 已固定为自身协议回放 fixture；422 不由该 400 冒充确认。

## 原生 OpenAI Images（工单 09–10）

image × openai × openai-images 是独立扩展路线 P08。冻结 pi 1.0.0 路由 OpenRouter 图像，
没有此原生路线；fixture 每项明确 pidiffSkip，routes 账本与离线 E2E 断言登记，不计为 pi
差分通过。宿主请求收敛为提示加有序参考图，输出为独立封闭的有序文本/图片块；整组校验失败
清空输出而保留用量。原生图像目录明确豁免聊天型号字段比对，按自己的证据纳入。

生成与内联 JSON 编辑共用 GenerateImages，参考图按顺序映射，mask 与输入保真度按型号能力
校验。回调不能改变已选端点/操作；编辑 fixture 和容量/越权失败均为 P08 扩展离线证据。
实际型号/费率、JSON 编辑接受性与真实支持声明由工单 11 交付；运行时不回退 multipart。
直连的 SDK 例外、有限资源与模态计价见 [ADR-0022](../adr/0022-barness-ai-openai-images-unary.md)。

工单 11 的固定 Sunburst 快照已通过本路线 JSON 编辑、生成和 mask 真实冒烟；
目录 2026-10-08.4 与带来源的 real-response replay 固定其费率和已验证能力。
该真实接受性仍为扩展路线证据，不计为 pi 差分通过。原始 PNG 响应超过 512 KiB
捕获上限，live-contract fixture 明确记录以公共结果恢复图片和 usage 的归一化；
原始截断标志、实际结果、图片校验与审计均保留在本组合证据包。

## 原生 Google Interactions 图像（工单 12–13）

image × google × google-interactions 是独立扩展 P09，固定 v1beta、store=false、同步与
内联交付。冻结 pi 1.0.0 没有这条原生路由；所有 fixture 显式 pidiffSkip，routes 与原生
目录豁免有公共离线断言，不计为 pi 差分通过。提示加参考图的收敛接口未扩成交错输入，
工单 13 接入有序参考图编辑，公共输入不支持 mask；这项输入收敛与协议可表达的交错输入
差异作为本扩展路径的一部分登记。内联格式沿用公共图像严格 png/jpeg/webp 校验，协议列出的
其他格式尚未开放；将来增加格式需独立文件头验证和 fixture。参考图数量受宿主、型号与 14 张
协议上限交集约束；总请求包含编码开销，≤min(MaxRequestBytes, 20,000,000)。全部 model_output 的文本/图片保序且原子
发布；真实终态、响应诊断 ID/型号和原生续接分开。

Google 的思考加到 Output，但是否已经包含在模态输出明细中用完整总量关系判断，
避免重复计价；两个变体明确为带来源的合成 fixture。工单 14 确认真实形状、费率与能力
后才能纳入 Google 内置图像型号或宣称 live 支持。既有 generateContent 聊天行为保持原契约。
详见 [ADR-0022](../adr/0022-barness-ai-openai-images-unary.md)。

## pi-ai 1.1.0 迁移（2026-10-11）

两个干净环境完整 AI 827 + telemetry 24 dist 文件等价。唯一运行 oracle 已替换为 1.1.0；官方 npm 缺 gitHead，未伪造，固定 tag/archive 身份与字节重建证明来源。当前全部六子集 814 条差分、目录逐字段核对通过，pending 0；完整发布仍受九组合及业务效果门禁约束。旧 1.0.0 记录是历史事实。

已修复：上下文估算改为 ceil(chars/3.5)，三个输出预算差异逐 case 登记 fixed；五处真实价格分层同步；模型分档采样按 R01–R09 实施，公开请求不删字段、不改差分投影。已支持型号没有采样默认，不合成填充。

用户于 2026-10-11 明确批准补齐 durationMs 并允许独立执行耗时的数值差异。账本限定四种聊天 API、有限 case ID，以及 result/terminal/partial 的单一 durationMs 字段。每次差分先断言结果为非负整数、唯一终态值与结果一致；缺失/类型错误/不一致无法由账本放行。非终态 PartialView 字段出现时机沿用既有共享视图契约，仅登记该字段的 presence 差异；最终字段必须存在。没有宽泛 result/event 豁免。

仍成立的既有扩展（SDK 错误文本、共享 PartialView、目录可选能力、隔离/资源限额与扩展路线）完整保留并重新执行。DeepSeek Responses 仍独立断言和显式 skip，不计为 pi 原生差分。1.1.0 上层整轮 retryAssistantCall 的 server_busy 文本等不在本库初始请求重试承诺内，未扩展产品范围。迁移明细及执行证据见[专项索引](../../.scratch/barness-ai-thinking-sampling/evidence.md)。

## Decisions 原生差分（工单 07，2026-10-13）

冻结 Pi 1.1.0 自身提供 OpenAI × Decisions × classifier，P10 是第七个必需差分子集。
它没有 `routes` 豁免。共同对象状态、PNG 图片输入、numeric-looking 名称/候选、string
"true"/"false" 候选、命名 choice/score/bool 基本答案、输入阶梯计费及显式 429 重试完整比较。
响应数组刻意乱序，按名称核对；公共字节保留大整数而 Pi 使用 Number 的差异有独立对照。

| 有限差异 | 自身证据与边界 |
| --- | --- |
| State string / JSON 精度 | OD49-string/precision；只限 input，字符串长度变化另核对 Content-Length |
| score 详情 / 逐题拒答 | OD49-score/refusal；score/confidence 共同字段保留，拒答的兄弟题与 billed usage 自身断言 |
| presence / total / reasoning | OD49-usage-absent/empty/full、total-only、tier；完整 known counts、Reasoning unset/value 与 input-only 估价逐例断言 |
| 严格校验 | OD49-extra-answer/duplicate-answer/confidence-range/distribution-sum/missing-answer/usage-invalid；整体清空答案，唯一合法用量保留 |
| callback / HTTP 错误 / retry | OD49-malformed-json/error-400/retry-429/retry-504/hook-final；明确回调数组、终态、每个尝试的完整 wire 与请求次数 |
| headers / catalog metadata | 六个 fetch/Go 默认头有固定值；OD48-catalog 比较整个条目，Binding endpoint 和 capability metadata 各有来源；P07-OD50 独立保留完整旧 Jev 内容 |

49 条账本记录均有有限 case 列表、path/kind、理由和需求；移到未登记 case 一律 pending，
改变共同 wire model 也 pending。P10 投影保留 cost、score 详情、refusal 和 usage presence，
P07 投影未改。来源、逐例完整 captures、实际 verdict、manifest/audit 与命令见
[证据包](../../.scratch/barness-ai-openai-decisions/frozen-differential-evidence/README.md)。

## MiniMax 中国区（工单 06，2026-10-18）

固定 pi 1.1.0 的既有 minimax-cn Messages M3 export 现已装配进 wrapper；版本、commit、
完整 model-data、827 dist 源码和 SDK lock pins 保持不变。M3 Messages 实际执行六条
公共 Client / frozen pi 成对差分：full/simple 文本、带签名思考的两个工具调用及 Host
工具结果后的完整续接。没有剪裁 request、events 或 Result 以掩盖差异。

新增账本仅限这六个 case 与实际字段：默认预算512000/131072、adaptive 与冻结 Claude
budget/display、simple 显式 disabled 与当前省略默认、原厂 beta/browser/eager/ephemeral
自动字段、无一小时缓存写入契约的可选零值 presence、合成候选价格及独立耗时。
每条 wire/token/cost 差异在分类前有独立固定值/算例断言；未知 case、字段、协议仍 pending。
完整目录比较另保留当前 `api.minimax.cn` 与冻结历史 `api.minimaxi.com`、文本子集与冻结
image/resize、保守输出上限、binary/default 元数据和候选合成费率的精确差异。

Chat/Responses 在冻结 provider 数据中没有路由，登记两条 Extension Route；M3.1 在三个
协议中均为型号扩展，不套用冻结 M3 的身份。五个独立 fixture 的四入口验收明确保存
pidiffSkip 与 `piParity=NOT_RUN`，不计原生差分通过。Messages 整条路线没有豁免。
源、有限 case/path 边界与重放见 [MiniMax 差分说明](../../ai/e2e/testdata/minimax-pidiff/README.md)
和 [工单证据](../../.scratch/barness-ai-minimax/evidence/ticket-06/README.md)。

这些是合成受控协议比较，不是当前厂商收费调用或正式支持声明。候选价格明确为机制测试值，
不将冻结历史美元报价当当前人民币换汇价格，也不代表 M Plan 扣费；正式目录/费率和各协议
自己的真实证据仍由后续发布工单验收。既有原生差分、分类账与扩展覆盖没有减少。
