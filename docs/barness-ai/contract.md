# barness-ai 公共契约、错误分类与装配说明

本文是 barness-ai（Go 包 `github.com/tokenbeat-lab/barness/ai`）的发布说明之一，汇总宿主可依赖的公共契约。行为的权威来源依次是 [spec](../../.scratch/barness-ai/spec.md)、[ADR](../adr/) 与包文档（`go doc ./ai`）；本文不放宽其中任何一条。术语见 [GLOSSARY](../../GLOSSARY.md)。

## 1. 构造

`ai.NewClient(ai.Config)` 返回只读、并发安全的 `*ai.Client`；所有依赖凭据的 Provider 操作都经过它。

| `Config` 字段 | 必填 | 说明 |
| --- | --- | --- |
| `Policy *ResourcePolicy` | 是 | 显式有限资源策略（D1，ADR-0002）。缺失、任一容量/时限为零或负数、字段关系非法时构造失败（`*ConfigError`）。没有内置默认值，也没有“关闭限额”模式。 |
| `Bindings BindingResolver` | 是 | 宿主按 `(TenantID, BindingID)` 解析授权绑定。 |
| `Credentials CredentialResolver` | 是 | 宿主按绑定解析凭据快照；秘密值以 `ai.Secret` 传入，不进入任何输出。 |
| `Catalog *Catalog` | 否 | 默认 `BuiltinCatalog()`（快照见 [catalog-snapshot.json](../../ai/release/catalog-snapshot.json)）；价格必须有限且非负。 |
| `Transport http.RoundTripper` | 否 | 构造时固定，没有每调用覆盖。nil 选择不读代理环境变量的 transport；无 CookieJar，不跟随重定向。 |
| `Admission Admission` | 否 | 宿主注入的准入（按厂商账户聚合、分布式配额），在内置的单租户/全进程并发限制之后对每次尝试调用，参数含 TenantID、AccountScopeID、AttemptID。 |
| `Observer Observer` | 否 | 异步、有界（`MaxQueuedObservations`）的调用/尝试记录；需要正的 `MaxQueuedObservations`。 |
| `AllowLoopbackHTTP` | — | 仅测试装配使用，允许绑定指向回环 http。生产必须为 false，所有 endpoint 为 https。 |

资源策略示例（每个数值附适用负载、依据与调整说明，并由压力场景验证）：`localassembly.LocalPolicy`、`hostintegration.CloudInteractivePolicy`、`hostintegration.CloudBatchPolicy`。压力数据见 [README](README.md#资源策略示例与数值依据)。

## 2. 调用

```go
s := client.Stream(ctx, scope, target, req, opts)          // 完整协议选项
s := client.StreamSimple(ctx, scope, target, req, simple)  // 统一选项
res, err := client.Complete(ctx, scope, target, req, opts)
res, err := client.CompleteSimple(ctx, scope, target, req, simple)
hc := client.WithHooks(hooks)                              // 可信回调，同样四个入口
```

| 输入 | 契约 |
| --- | --- |
| `CallScope` | `TenantID`、`RequestID` 必填，由可信宿主从认证结果建立，不得取自请求内容；`ActorID`、`JobID` 可选。RequestID 全局唯一（`examples/requestid`），一次逻辑调用一个，内部重试沿用，再次调用必须换新。 |
| `Target` | `BindingID` 与 `ModelID`；协议、endpoint、账户与凭据由授权绑定确定。可调用的模型是目录与 `Binding.AllowedModels` 的交集。 |
| `Request` | 本次 `Messages`、可选 `SystemPrompt`（归一为初始 system 消息）与 `Tools`。调用在接收时复制输入，之后修改源数据不影响在途调用；交接期间不得并发修改。 |
| 完整选项 | 必须与绑定的协议匹配：`ResponsesOptions`、`AnthropicOptions`、`GeminiOptions`、`ChatOptions`；需要区分未设置/null/零值的字段使用 `Nullable[T]`。 |
| `SimpleOptions` | reasoning（minimal…max）、thinking budgets、toolChoice、maxTokens、cacheRetention 等公共参数，按模型能力映射。 |
| `Hooks` | `TransformHeaders`（认证头合并后、adapter 前）、`OnPayload`（adapter 构建请求体后、重试之外，一次）、`OnResponse`（初始响应成功后、start 前；Gemini 不调用）。回调失败为 `callback_failed`；回调不能放宽鉴权、目标、模型或托管工具（ADR-0005）。 |

宿主不得在普通生成请求中提交 key、凭据引用、endpoint、Authorization、代理或可执行函数；这些只来自可信装配。

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

## 4. 错误分类

`*ai.Error` 保留消息终态与 `ErrorMessage`（`Error.Message`），并提供稳定的 `Code`/`Phase`；`errors.Is(err, &ai.Error{Code: ai.CodeRateLimited})` 按分类匹配（Phase 为空时不比较 Phase），`errors.As` 取得 `HTTPStatus`、`ProviderRequestID`、`RetryAfter`。宿主回调或准入自身的错误只经 `Unwrap` 可见，不进入 Message。

| Code | 含义 | 常见 Phase | 是否到达 Provider |
| --- | --- | --- | --- |
| `invalid_request` | 请求结构、选项或历史非法；模型不在目录或不属于绑定的协议；不支持的 AuthKind | scope、binding、capability | 否 |
| `tenant_denied` | 绑定禁用、调用主体无权、模型未获绑定授权、凭据访问被拒、回调试图放宽授权 | binding、capability、credential、request | 否 |
| `binding_not_found` | 绑定不存在，或属于其他租户 | binding | 否 |
| `credential_unavailable` | 缺 key 或空 key、凭据撤销、秘密后端故障；两次解析间配置快照不一致（D2，不内部重解析） | credential、consistency | 否 |
| `admission_denied` | 内置并发限制或宿主准入拒绝，或等待超时/超过等待者上限 | admission | 否 |
| `upstream_auth` | Provider 拒绝凭据（401/403）；不换 key、不换身份 | request | 是 |
| `rate_limited` | 429 | request | 是 |
| `upstream_error` | 5xx/408/409、流内错误、failed 响应、非长度原因的 incomplete | request、stream | 是 |
| `transport` | 建连失败、连接中断 | request、stream | 可能 |
| `protocol` | 缺协议终态的 EOF、无法解析的帧、非法事件序列 | stream | 是 |
| `canceled` | 宿主取消 context 或 Close | 任意 | 视阶段 |
| `deadline_exceeded` | 宿主 deadline、`CallTimeout`、建连/响应头/读空闲/协议 timeoutMs 中最早者 | 任意 | 视阶段 |
| `resource_limit` | 请求/图片字节（发送前）、帧/总输出/工具 JSON/错误体（读取中）、事件队列超限；超限的尝试不重试 | scope、request、stream、event_queue | 视限额 |
| `callback_failed` | 可信宿主回调出错或未产出可用请求体 | request | TransformHeaders/OnPayload 失败时否；OnResponse 在收到初始响应后运行，其失败时请求已到达 |

Phase 取值：`scope`、`binding`、`capability`、`credential`、`consistency`、`admission`、`request`、`stream`、`event_queue`。错误分类不替换 StopReason：设置阶段捕获的取消形成 `error`，进入 adapter 后的取消形成 `aborted`；单次尝试自身的时限到期为 `error`。

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
