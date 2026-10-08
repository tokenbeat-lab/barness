# barness-ai 公共契约、错误分类与装配说明

本文是 barness-ai（Go 包 `github.com/tokenbeat-lab/barness/ai`）的发布说明之一，汇总宿主可依赖的公共契约。行为的权威来源依次是 [spec](../../.scratch/barness-ai/spec.md)、[ADR](../adr/) 与包文档（`go doc ./ai`）；本文不放宽其中任何一条。术语见 [GLOSSARY](../../GLOSSARY.md)。

## 1. 构造

`ai.NewClient(ai.Config)` 返回只读、并发安全的 `*ai.Client`；所有依赖凭据的 Provider 操作都经过它。

| `Config` 字段 | 必填 | 说明 |
| --- | --- | --- |
| `Policy *ResourcePolicy` | 是 | 显式有限资源策略（D1，ADR-0002）。缺失、任一容量/时限为零或负数、字段关系非法时构造失败（`*ConfigError`）。没有内置默认值，也没有“关闭限额”模式。 |
| `Bindings BindingResolver` | 是 | 宿主按 `(TenantID, BindingID)` 解析授权绑定。 |
| `Credentials CredentialResolver` | 是 | 宿主按绑定解析凭据快照；秘密值以 `ai.Secret` 传入，不进入任何输出。 |
| `Catalog *Catalog` | 否 | 默认 `BuiltinCatalog()`（快照见 [catalog-snapshot.json](../../ai/release/catalog-snapshot.json)）；三类型号分别存放，完整身份不得重复，能力与模态须自洽，价格有限且非负，图像声明的每个模态须有费率。 |
| `Transport http.RoundTripper` | 否 | 构造时固定，没有每调用覆盖。nil 选择不读代理环境变量的 transport；无 CookieJar，不跟随重定向。 |
| `Admission Admission` | 否 | 宿主注入的准入（按厂商账户聚合、分布式配额），在内置的单租户/全进程并发限制之后对每次尝试调用，参数含 TenantID、AccountScopeID、AttemptID。 |
| `Observer Observer` | 否 | 异步、有界（`MaxQueuedObservations`）的调用/尝试记录；需要正的 `MaxQueuedObservations`。 |
| `AllowLoopbackHTTP` | — | 仅测试装配使用，允许绑定指向回环 http。生产必须为 false，所有 endpoint 为 https。 |

资源策略示例（每个数值附适用负载、依据与调整说明，并由压力场景验证）：`localassembly.LocalPolicy`、`hostintegration.CloudInteractivePolicy`、`hostintegration.CloudBatchPolicy`。压力数据见 [README](README.md#资源策略示例与数值依据)。

`Model` 保持原字段与聊天含义；新增 `ImageModel`（输入/输出模态、图片能力、`ImagePricing`）和
`ClassifierModel`（上下文窗口、问题类型/容量、`ModelCost`），身份为 `(Operation, Provider, API, ID)`。
同一 ID 可跨操作、Provider 或协议并存；重复完整身份以 `ConfigError{Field: "Catalog"}` 拒绝。
图像输出须包含 image；参考图、mask 与 input fidelity 须有相应输入能力。分类支持 choice、score、bool；
choice 的选项上限至少 2，score 的级数最少 2 且上下界有序，不支持的问题类型不得声明非零容量。
`ClassifierCapabilities.Kinds` 使用 `[]ClassifierQuestionKind`，对应常量为
`ClassifierQuestionChoice`、`ClassifierQuestionScore`、`ClassifierQuestionBool`。
已有 Go 宿主的 `[]string` 构造须改为此类型；从动态字符串逐项转换后仍需目录构造校验。
JSON 继续编码为 `"choice"`、`"score"`、`"bool"`，未知或空类型仍在 NewClient 边界拒绝。
图像每个已声明输入/输出模态须有费率值（显式 0 合法，unset/null 缺失），缓存费率按协议声明；直接 OpenAI Images 禁止设置缓存输入费率。
Responses/Chat 型号默认 samplingParams 同样不能覆盖保留字段。

`Client.Catalog()` 返回构造时固定的完整目录副本。`Catalog.Lookup`/`ModelsOf` 只查聊天；
`LookupImage`/`ImageModelsOf`、`LookupClassifier`/`ClassifierModelsOf` 分别强类型查询另两类，
参数按 Provider、API（查找另含 ID）限定，列表保持目录顺序。所有返回值、切片、map、原始 JSON、
能力与阶梯价格均独立复制。配置在构造交接期间不得并发修改。发现目录无需读取绑定或凭据，
也不授予调用权限；内置目录列已有验收的聊天路线、jev-1.13.0 分类和通过自身 JSON 编辑真实门禁的
`gpt-image-2.5-sunburst-2026-09-08` 图像型号（目录 2026-10-08.4）。图像支持
一张参考图、一张输出、1024×1024、low/medium 质量、mask、透明背景；PNG/JPEG/WebP
均有本组合实测。宿主仍须启用 ImagePolicy 并逐个授权型号。
三类型号与全部价格都参加目录 Hash；内容改变应升 Version，宿主即使重用版本也会得到不同哈希（ADR-0010/0020）。

## 2. 调用

```go
s := client.Stream(ctx, scope, target, req, opts)          // 完整协议选项
s := client.StreamSimple(ctx, scope, target, req, simple)  // 统一选项
res, err := client.Complete(ctx, scope, target, req, opts)
res, err := client.CompleteSimple(ctx, scope, target, req, simple)
res, err := client.Classify(ctx, scope, target, classification, nil) // 同步分类
res, err := client.GenerateImages(ctx, scope, target, images, nil) // 同步图像
hc := client.WithHooks(hooks)                              // 可信回调，同样六个入口
```

| 输入 | 契约 |
| --- | --- |
| `CallScope` | `TenantID`、`RequestID` 必填，由可信宿主从认证结果建立，不得取自请求内容；`ActorID`、`JobID` 可选。RequestID 全局唯一（`examples/requestid`），一次逻辑调用一个，内部重试沿用，再次调用必须换新。 |
| `Target` | `BindingID` 与 `ModelID`；操作、Provider、协议、endpoint、账户与凭据由授权绑定确定。可调用的型号是该操作/Provider/协议下的目录与 `Binding.AllowedModels` 的交集。 |
| `Request` | 本次 `Messages`、可选 `SystemPrompt`（归一为初始 system 消息）与 `Tools`。调用在接收时复制输入，之后修改源数据不影响在途调用；交接期间不得并发修改。 |
| 完整选项 | 必须与绑定的协议匹配：`ResponsesOptions`、`AnthropicOptions`、`GeminiOptions`、`ChatOptions`；需要区分未设置/null/零值的字段使用 `Nullable[T]`。 |
| `SimpleOptions` | reasoning（minimal…max）、thinking budgets、toolChoice、maxTokens、cacheRetention 等公共参数，按模型能力映射。 |
| `Hooks` | `TransformHeaders`（认证头合并后、adapter 前）、`OnPayload`（adapter 构建请求体后、重试之外，一次）、`OnResponse`（初始响应成功后、start 前；Gemini 不调用）。回调失败为 `callback_failed`；回调不能放宽鉴权、目标、模型或托管工具（ADR-0005）。 |

宿主不得在普通生成请求中提交 key、凭据引用、endpoint、Authorization、代理或可执行函数；这些只来自可信装配。

`Operation` 只取 `chat`、`image`、`classifier`。一份 `Binding` 固定一种操作；`Binding.Operation`
零值规范化为 chat，是兼容旧宿主的公共契约例外：旧绑定仅有聊天权限，默认 chat 不扩大授权。
`Enabled` 和 `Credential.Active` 仍以零值拒绝。未知非空操作为 `invalid_request/binding`；
四个聊天入口（含 WithHooks）期望 chat，操作不匹配为 `tenant_denied/capability`。
核对操作后才查聊天目录与白名单，最后读取凭据并核对版本快照；操作拒绝没有凭据读取、Provider 请求或 Attempt。

### 同步分类（工单 05–07）

Client 与 HookedClient 的 Classify 接收 ClassifierRequest 和既有封闭 Options；nil 或空 TypeSafeOptions
表示协议默认，错协议选项在 capability 拒绝，SimpleOptions 不能传入。Binding 必须固定为
classifier × typesafe × typesafe-system-one；按独立分类目录与白名单授权后才读取凭据。

ResourcePolicy.Classifier 为 nil 时以 invalid_request/scope 返回完整失败结果。启用时 MaxQuestions、
MaxStateBytes、MaxQuestionBytes 均须为正且在构造时独立复制。状态为 JSON string/object/array；
Questions 至少一项、键非空，封闭为 ChoiceQuestion、ScoreQuestion、BoolQuestion。
Instructions 是非空 JSON string/object/array；单选 Criteria 为 1–255 个 JSON string/object/array/null 描述，
评分 Criteria 为 2–10 个有序 JSON string/object/array 描述，下标即分值。
是非可省略 Criteria，提供时 True/False 均须为 JSON string/object/array；null 不充当说明。
JSON 数字保留原始精度。接收时先检查已知长度再复制，交接期间不得并发修改源值。
公共请求 JSON 的 type 为 choice/score/bool；native wire 将 bool 转为 noul，反序列化拒绝未知类型和字段组合。
状态/说明的重复 JSON 键拒绝；对象的嵌套值允许任意合法 JSON。

一次调用提交完整状态和问题集合。Answers 是 map[string]ClassifierAnswer，值为 ChoiceAnswer、ScoreAnswer、BoolAnswer。
BoolAnswer.Probability 是“是”的概率，不设阈值。ScoreAnswer.Score 是期望值，Confidence 在 0–1；
可选 Probabilities 是 map[int]float64，Legend 是 map[int]json.RawMessage，保留字符串/对象/数组图例。
Score 必须在 0..级数−1，分布若提供必须覆盖每级、总和在 1±1e-6，期望与 Score 的绝对差不超过 1e-6。
图例若提供必须伴随分布、键集相同，且逐项与最终请求的等级描述按精确 JSON 数值相等；对象键序不影响相等，数组顺序有意义。
可选字段省略可接受，显式 null 或部分分布/图例拒绝。
每个最终请求问题恰有一个同类型答案，无额外键；分布选项集必须相同、概率和 confidence 有限且在 0–1，
总和在 1±1e-6，choice 属于最高概率项（允许并列）。任何答案失败时以 protocol/response 结束，
Answers 整体为空，已上报 Usage 和 Metadata 保留。重复响应字段、答案键或概率键拒绝，不归一化分布。
重复 usage 或 token 字段的用量存在歧义，以 unreported 拒绝；其他响应字段错误仍保留明确上报的用量。
成功 StopReason 为 stop，失败为 error 或 aborted，error 为对应 *ai.Error。

三种可信回调的 CallScope、Payload/ResponseInfo 都带入口确定的 Operation；入口覆盖宿主提交的
CallScope.Operation，该字段不授予权限。最终请求回调执行一次，回调后重新解码、校验并冻结问题集；
非法结果或分类容量超限为 callback_failed，改变授权模型、凭据或添加授权字段为 tenant_denied。
回调普通 JSON 值的可取得尺寸在序列化前检查；最终 MaxRequestBytes 在问题独立解码前检查。
自定义 JSON marshaler 的宿主代码自行负责执行和分配，其输出仍须先通过整份请求字节预算。
成功响应之后执行一次只看 HTTP 元数据的响应回调再读取。Gemini 聊天仍不调用响应回调。

unary 成功体只按 MaxOutputBytes 读取，不受 SSE MaxFrameBytes 限制；错误体按 MaxErrorBodyBytes。
读取、解析和答案校验完成前持有准入许可，任一失败关闭 body 并释放。默认不重试；
绑定显式允许的初始请求重试复用冻结快照，成功响应之后的失败不重放。
已结束的 context 在输入复制/配置读取前以 scope/error 拒绝；解析器返回时再次检查 context，
即使返回成功也在 binding/credential 阶段结束，保持未解析归属。准入等待、退避、读取成功体
期间的调用取消或总截止时间分别为 admission/request/response + aborted；尝试时限仍为 error。
连接失败按绑定预算重试；被拒绝的准入不生成 Attempt。响应回调失败仍为 request 阶段，
errors.Is/As 保留可信宿主原因；其他成功体读取/解码/校验失败为 PhaseResponse。

usage 的 input_tokens/output_tokens 均存在为 complete，缺项为 partial，无 usage 为 unreported。
TypeSafe 只按目录输入费率估价，输出为零；内置 jev-1.13.0 的官方输入费率为 $0.042/百万 token（2026-10-08）。
实际版本型号只放 ResponseModel Nullable[string]，授权 ModelID 不变，
x-typesafe-request-id 放 Attempt.ProviderRequestID。Observer 只记录操作、归属、尝试、分类及用量。
AttemptStarted 的 UsageReporting 为 unreported，AttemptFinished 与 CallFinished 的尝试记录
使用同一 unreported/partial/complete 轴。PhaseResponse 在 errors.Is 的 Code 匹配中与其他阶段相同：
Phase 为空匹配该 Code，Phase 非空同时匹配阶段。离线生命周期与资源证据见
[工单 07](../../.scratch/barness-ai-pi-1.0/unary-failures-evidence/README.md)。
内置目录 2026-10-08.3 仅列固定 jev-1.13.0，自己的真实冒烟已合并；禁用绑定或子策略即可停用新调用。
不在本地计算分类 token 容量，真实上下文 guard 为 400 max_tokens_exceeded，保持 invalid_request/request；422 线上形状未确认。
评分有界两位小数舍入与完整报告契约见 ADR-0021/0016；其他概率与图例规则保持。

## 3. 输出

| 输出 | 契约 |
| --- | --- |
| `*Stream` | 先返回、后台生产；`Next`/`Event`/`Envelope` 单消费者，`Result` 可独立或并行等待，`Close` 并发安全且幂等，`Err` 为 Scanner 风格。完整消费恰有一个终结事件（`done` 或 `error`）。只等 `Result` 不读事件的 Stream 仍受事件队列上限约束，输出超过 `MaxQueuedEvents`/`MaxQueuedEventBytes` 时以 `resource_limit`（Phase `event_queue`）结束——**仅需最终消息时用 `Complete`**。 |
| 事件 | `start`；`text_start/delta/end`、`thinking_start/delta/end`、`toolcall_start/delta/end` 可交错、索引稳定；`done`/`error`。`Stream.Envelope()` 附带事件发布时的 `CallAttribution`，多流汇合按信封路由，不按到达顺序推断。 |
| `Result` | `Message`（最终 `AssistantMessage`，失败时也完整有效）与 `Metadata`（`CallMetadata`：调用归属、绑定/凭据版本、目录版本与哈希、原生状态降级计数、每次 HTTP 尝试及其用量与用量完整性）。 |
| `(Result, error)` | StopReason 为 `error` 或 `aborted` 时 error 非 nil（`*ai.Error`），Result 仍完整；成功时 error 为 nil。 |
| 原生状态 | 本进程产出的消息自带可信来源，可直接回放；持久化读回的消息须由宿主鉴权后经 `TrustNativeState(scope, envelope)` 担保。JSON 反序列化得不到可信状态。无封套或账户不匹配时按跨模型规则降级（调用照常进行，计数进 `Metadata` 与 Observer）。 |
| 工具 | `ToolCall` 保留原始参数 JSON（`RawArguments`）与展示用部分解析（`Arguments`，不可执行）；`ValidateToolCall` 校验完整参数后才交宿主执行；`RepairToolJSON` 显式修复；`ToolResultText` 构造结果。工具执行与下一轮调用属于宿主。 |
| 用量与成本 | `Usage` 保持 pi 的数字（含按目录价格估算的 `Cost`，非账单）；每次尝试的 `UsageReporting`（未上报/部分/完整）另行记录，零值不表示免费。 |

`CallAttribution.Operation` 从入口开始即为 chat 或 classifier，包含 CallStarted、预检失败和事件信封；
Provider、API、ModelID、AccountScopeID 与绑定/凭据/目录版本仍只在一致快照解析成功后填写。
operation 是经审阅的 Observer 固定枚举元数据，已加入脱敏审计白名单。

## 4. 错误分类

`*ai.Error` 保留消息终态与 `ErrorMessage`（`Error.Message`），并提供稳定的 `Code`/`Phase`；`errors.Is(err, &ai.Error{Code: ai.CodeRateLimited})` 按分类匹配（Phase 为空时不比较 Phase），`errors.As` 取得 `HTTPStatus`、`ProviderRequestID`、`RetryAfter`。宿主回调或准入自身的错误只经 `Unwrap` 可见，不进入 Message。

| Code | 含义 | 常见 Phase | 是否到达 Provider |
| --- | --- | --- | --- |
| `invalid_request` | 请求结构、选项或历史非法；型号不在对应操作目录或不属于绑定的协议；不支持的 AuthKind 或未知 Operation | scope、binding、capability | 否 |
| `tenant_denied` | 绑定禁用、调用主体无权、入口操作与绑定不匹配、型号未获绑定授权、凭据访问被拒、回调试图放宽授权 | binding、capability、credential、request | 否 |
| `binding_not_found` | 绑定不存在，或属于其他租户 | binding | 否 |
| `credential_unavailable` | 缺 key 或空 key、凭据撤销、秘密后端故障；两次解析间配置快照不一致（D2，不内部重解析） | credential、consistency | 否 |
| `admission_denied` | 内置并发限制或宿主准入拒绝，或等待超时/超过等待者上限 | admission | 否 |
| `upstream_auth` | Provider 拒绝凭据（401/403）；不换 key、不换身份 | request | 是 |
| `rate_limited` | 429 | request | 是 |
| `upstream_error` | 5xx/408/409、流内错误、failed 响应、非长度原因的 incomplete | request、stream | 是 |
| `transport` | 建连失败、连接中断 | request、stream、response | 可能 |
| `protocol` | 缺协议终态的 EOF、无法解析的帧、非法事件序列或 unary JSON/答案校验失败 | stream、response | 是 |
| `canceled` | 宿主取消 context 或 Close | 任意 | 视阶段 |
| `deadline_exceeded` | 宿主 deadline、`CallTimeout`、建连/响应头/读空闲/协议 timeoutMs 中最早者 | 任意 | 视阶段 |
| `resource_limit` | 请求/图片字节（发送前）、帧/总输出/工具 JSON/错误体（读取中）、事件队列超限；超限的尝试不重试 | scope、request、stream、response、event_queue | 视限额 |
| `callback_failed` | 可信宿主回调出错或未产出可用请求体 | request | TransformHeaders/OnPayload 失败时否；OnResponse 在收到初始响应后运行，其失败时请求已到达 |

Phase 取值：`scope`、`binding`、`capability`、`credential`、`consistency`、`admission`、`request`、`stream`、`response`、`event_queue`。错误分类不替换 StopReason：设置阶段捕获的取消形成 `error`，进入 adapter 后的取消形成 `aborted`；单次尝试自身的时限到期为 `error`。

`Error.Message` 按 pi 原样引用 Provider 错误体，只脱敏 key 形态的文本，可能复述 Provider 回显的请求内容：它只返回给发起调用的租户，宿主不得写入共享日志，需要记录时使用 Observer 记录（批准的安全差异，ADR-0009）。barness-ai 自身不写日志。

## 5. 装配说明

### 本地（`ai/examples/localassembly`）

本地程序显式选择一个秘密来源（指定的环境变量或秘密文件，二者只能择一），读取后以单租户、单绑定装配；核心与 SDK 都不读 `OPENAI_API_KEY` 等环境变量或 endpoint。资源策略用 `LocalPolicy` 或按其注释调整。

### 云端宿主（`ai/examples/hostintegration`）

1. 认证结果 → `CallScope`（新 RequestID、ActorID）；拒绝自报 TenantID、他人会话与未认证调用者。
2. 按 `(tenant_id, session_id)` 读取历史；持久化的原生状态经 `TrustNativeState` 担保后回放。
3. 实现 `BindingResolver`/`CredentialResolver`：首期示例无缓存；自定义缓存须公布 TTL、失效与撤销传播上限，并以租户、凭据标识与版本区分。
4. 选择 `CloudInteractivePolicy` 或 `CloudBatchPolicy`；`admission_denied` 时向客户端返回 429 与 Retry-After，排队留在 barness-ai 之外。
5. 需要按厂商账户或跨实例配额时注入 `Admission`；需要记录时注入 `Observer` 并定期读 `ObserverStats` 的丢弃计数。
6. 重试在绑定上配置（`Binding.Retry`，默认不重试，ADR-0006）。
7. 下游断开或发送失败时取消该调用的 context 或 Close 其 Stream，只结束本次生成。

### 工具往返（`ai/examples/toolloop`）

宿主以 `ValidateToolCall` 校验完整参数后自行执行工具，截断（`length`）或中断的轮次不执行；每个下一轮都是新的逻辑调用与新的 RequestID。

## 6. 支持声明

协议组合只有在适用验收与真实冒烟通过后才可宣称支持；当前状态见 [README](README.md#支持矩阵) 与 [`ai/live/support-matrix.json`](../../ai/live/support-matrix.json)。“兼容 OpenAI”不等于任何兼容服务已验收；共享 adapter 的通过不连带标记其他组合。

## OpenAI 图像生成与 JSON 编辑（工单 09–10）

Client 与 HookedClient.GenerateImages 接收 ImagesRequest（Prompt、有序 ReferenceImages）和封闭的
ImageOptions，同步返回 ImagesResult，不产生 delta。OpenAI 无参考图走 generations，有参考图走 edits，
仅发送 JSON 内联图片。绑定必须为 image × openai × openai-images，并逐个授权目录型号；旧零值聊天
绑定不会得到图像权限。SimpleOptions 不能用于此入口。

ResourcePolicy.Image 为 nil 时以 invalid_request/scope 拒绝；ImagePolicy 的 MaxInputImages、
MaxOutputImages、MaxOutputImageBytes、MaxTotalOutputImageBytes 均须为正，单张 ≤ 总图片字节 ≤
MaxOutputBytes。策略深复制，旧聊天配置继续有效。请求、响应和 context 同受既有全局预算保护。

OpenAIImagesOptions 包含 N、Size、Quality、Background、OutputFormat、OutputCompression、Moderation、
InputFidelity（Nullable）与 Mask（*Image）。nil/省略取默认；显式 null 拒绝，压缩零有效。N 在 1–10 及有效输出上限内，
提示 1–32000 Unicode 字符，透明背景限 png/webp，压缩限 jpeg/webp 的 0–100。
参考图有序独立复制，MIME/文件头须为 png/jpeg/webp，base64 严格标准校验；不读取文件、
下载 URL 或接受 file_id/其他厂商引用。每个 data URL ≤20971520 字符，参考图不超过宿主、
协议 16 张与型号 MaxReferenceImages 的最小值；mask 占宿主输入张数和 MaxImageBytes，
要求参考图、Mask 能力且与第一张同尺寸。全部已知长度先检查，再复制/解码。
尺寸/质量/透明背景/保真度须由 ImageCapabilities 授权；Qualities、InputFidelity 空列表要求
省略相应选项，保真度仅编辑允许。gpt-image-2 省略；Sunburst 的允许值未被官方文档和本次实测确认，内置型号要求省略。
CustomSizes 是可选 ImageSizeConstraints（替换早期布尔开关），包含正值 EdgeMultiple、
MaxAspectRatio、MaxEdge、MinPixels、MaxPixels 与来源日期 Source，深复制；所有数字尺寸
须符合声明的硬约束，显式 Sizes 也不能绕过。nil 仅允许 Sizes 列出的值。
回调可调整允许的最终编辑字段；端点/操作先选定，不能增图把生成改成编辑、删除必需参考图
或改变型号。冻结后重做 mask、内联资源、预算与所有选项组合校验；容量失败为 resource_limit，
外部资源/操作/授权字段为 tenant_denied，未知字段/无效 schema 为 callback_failed。
拒绝 stream、partial_images、response_format、user 与续接状态；不回退 multipart。

Content 是独立封闭的 ImageOutputText/ImageOutputImage 有序块，图片为 base64 + MimeType。
响应格式优先，其次冻结请求格式，再取 png；严格验证 base64 与文件头，数量、单张与总量通过后
整体发布。失败 Content 为空，先登记的 Usage 与 Metadata 保留；2xx 后失败不重放。

Usage.Modalities 是可选 ModalityUsage，四项 Nullable token 计数区分缺失与零；没有第二个上报状态。
齐全为 complete，缺总量或任一输入/输出明细为 partial，usage 缺失/null 为 unreported；成本只算
已知模态，每个乘积单独舍入后相加。直接 Images 不设缓存费率或按张附加价格。模态用量在结果、
尝试与 Observer 间独立复制；聊天 JSON 不增加字段。x-request-id 为厂商请求 ID，ResponseID、
ResponseModel 不填，不保留 revised prompt。Observer 只含元数据。

完整行为、SDK 直连理由与后续 live 边界见 [ADR-0022](../adr/0022-barness-ai-openai-images-unary.md)。

### Google Interactions 图像生成与参考图编辑（工单 12–13）

绑定为 image × google × google-interactions，Endpoint 为带 `/v1beta` 的基址；
仅 POST `/interactions`，型号为裸 ID，认证仅用 X-Goog-Api-Key。GoogleImagesOptions
只有 AspectRatio 与 ImageSize（Nullable）；nil 用默认，null/其他协议选项拒绝，值须在
型号的 AspectRatios/ImageSizes 中。宿主自带目录须明确 OutputText 费率（显式 0 合法），
因为只请求图像也可能消耗文本思考。工单 14 已以本组合真实生成/参考编辑 PASS 纳入 gemini-nano-banana-2.1，
目录 2026-10-08.5 仅声明一参考/一输出、1K/1:1；文档化更大能力尚不声明。
ImagesRequest 为 prompt 文本在前、ReferenceImages 按原顺序在后，同一 `/interactions` 入口
提交内联 base64 与 MIME；入口 slice 在 resolver 前独立复制，字符串不可变。公共输入不提供
交错文本/图片与 mask，这一收敛登记为扩展。参考图数量 ≤min(MaxInputImages,
MaxReferenceImages, 14)，每张受 MaxImageBytes 和已有严格 png/jpeg/webp 文件头校验约束。
首批 Nano Banana 2.1 的官方能力为最多 10 个物体加 4 个角色；目录只记录总数，不推断图片
语义。型号允许 1K/2K/4K 时 512 拒绝；尺寸与宽高比始终按最终请求的授权型号验证。

请求固定 store=false、单个 image response_format，显式省略 delivery（真实服务拒绝该可选字段，
维护者批准修订，ADR-0023）；回调不能加入 delivery（inline/uri/null 也拒绝），以省略 background/stream
实现同步调用。回调后独立解码、重新校验固定项、允许字段、型号能力及最终大小；
续接、Agent、工具、环境、webhook、续读令牌、服务等级与网址资源为 tenant_denied，
无效/重复 schema 为 callback_failed；全局 MaxRequestBytes 与 Google 20 MB（20,000,000 字节） 取小值。
入口先以已校验 base64 长度和小型 JSON 信封估算准确编码字节，超限时不读凭据、不复制大图。
回调可在授权能力内添加/扩大参考图；普通容器、指针、带类型值与 raw JSON 在编码前重新检查
数量与单图大小，raw 转义用固定缓冲计长。自定义 marshaler 按既有资源契约交给编码器，
冻结后以实际 JSON 字节重验全部输入、单图与总量。提示沿用公共图像的 1–32000 字符边界。

仅 completed 发布内容，遍历全部 model_output 的 text/image 并保留出现顺序；不承诺精确张数。
failed/cancelled 为 upstream_error/response，其他状态或 continuation_token 为 protocol/response。
无图片为 protocol，有 errors 诊断的无图片为 upstream_error；错误文本固定，不暴露上游正文。
URI 交付不下载，混合内联与 URI 也拒绝。全部图片通过 MIME、严格 base64、文件头及
数量/单张/总字节限额后整体发布；失败清空 Content，保留已解析的用量、诊断 ID/型号和元数据。
ResponseID/ResponseModel 仅在响应报告时填写，授权 ModelID 不变，不保存原生续接状态。

Input 取 total_input_tokens，Output 取 total_output_tokens + total_thought_tokens，Reasoning
记已计入 Output 的思考量，TotalTokens 优先厂商 total_tokens，缺失才加已知 Input/Output。
CacheRead 记录厂商 total_cached_tokens，不重复加到总输入/总量，也不猜测缓存模态价格。
四个总量和两组完整的 text/image 明细为 complete，缺项 partial，缺 usage/null 为 unreported。
原始模态计数保持不变。输出与思考总量均报告时，完整输出明细和等于 total_output_tokens，计价视图把思考加到文本；
和等于 Output 时不再加。无文本明细或明确零文本时，只计已知思考；无法判断包含关系时
保持 partial，仅估已报告模态，不猜缺项。不一致/非法计数为 protocol，已知用量保留。
真实模态形状仍由工单 14 确认，两个包含规则变体是明确注明来源的合成 fixture。

成功响应先调用一次 OnResponse，再读取/解析/完整校验，期间持有准入许可；任何 2xx 后
失败都不重放。沿用 unary 的初始请求重试、时限、取消和元数据观测。Gemini generateContent
聊天继续不调用 OnResponse。P09 是原生扩展路线，fixture 不参加 pi 差分。


工单 14 的真实 Google 响应省略 id/请求 ID 头，报告 model/completed；ResponseID
保持 unset。output_tokens_by_modality 只有 image=1120，不含单独报告的文本思考；
Output 加 thought、Reasoning 为内含分项，TotalTokens 保留厂商值。未报输出文本
及未报输入图片明细保持 unset，usage 为 partial；成本仅估已知分项，thought 按
文本输出价计一次，不从内部 model invocation 明细补出未知成本。费率与能力来源
见 google-images/official.json 及 ADR-0023。
