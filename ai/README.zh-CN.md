# barness-ai

[English](README.md) | 简体中文

`github.com/tokenbeat-lab/barness/ai` 是 barness 的模型协议模块：一个可信、并发安全的
`Client`，为**多个租户**向模型服务发起**一次模型操作**。它是
[pi-ai](https://github.com/earendil-works/pi)（冻结基线 `1.0.0`）的 Go 复刻，并在其上增加了租户隔离、显式资源限额与可观测性。

本文面向两类读者：

- **人**：想了解模块做什么、怎么调用、各项保证记录在哪里。
- **Coding agent**：修改本模块前需要先掌握不变量、目录结构、约定与测试命令（见
  [写给 coding agent](#写给-coding-agent)）。

权威来源的优先级依次为：[1.0 升级 spec](../.scratch/barness-ai-pi-1.0/spec.md) 及其基础 spec、[ADR](../docs/adr/)、包文档（`go doc ./ai`）。本文只做归纳，不放宽其中任何一条。术语定义见 [GLOSSARY.md](../GLOSSARY.md)。

## 职责边界

| barness-ai 负责 | 宿主（你的程序）负责 |
| --- | --- |
| 把宿主选定的消息历史转换为 Provider 请求，以 pi-ai 事件流式返回响应，并给出最终 `AssistantMessage` | 认证调用者，并据此构造每个 `CallScope`——绝不取自请求内容 |
| 通过宿主的 resolver 解析租户的**服务绑定**与**凭据**，并在整个调用中固定该快照 | 保存绑定、凭据与历史；决定每一轮放入哪些历史 |
| 执行有限的 `ResourcePolicy`（字节、队列、并发、时限） | 按自身负载选择策略数值 |
| 仅在来源匹配时回放 Provider 原生状态（签名、加密推理） | 用 `TrustNativeState` 为持久化的原生状态担保 |
| 报告工具调用，按需校验参数 | 执行工具，并把每个下一轮作为新的逻辑调用发起 |
| 以 unary 响应生成原生图像或返回完整校验的分类答案 | 显式启用操作绑定、目录型号与有限子策略 |
| 错误分类、成本估算、发出观测记录 | 把错误映射到自己的 API（如 `admission_denied` → HTTP 429） |

不做的事：没有 Agent loop，没有内置记忆或历史裁剪，不执行工具，不写日志，不读取 `OPENAI_API_KEY` 之类的环境变量，没有默认资源限额。

## 支持的组合

| 操作 | Provider × 协议 | `ProviderID` / `API` | 真实冒烟 |
| --- | --- | --- | --- |
| chat | OpenAI × Responses | `openai` / `openai-responses` | PASS |
| chat | Anthropic × Messages | `anthropic` / `anthropic-messages` | PASS |
| chat | Google × Gemini Developer API | `google` / `google-generative-ai` | PASS |
| chat | OpenAI × Chat Completions | `openai` / `openai-completions` | PASS（协议本身不返回可回放的推理） |
| chat | DeepSeek × Responses（扩展路径，pi 无此路由） | `deepseek` / `openai-responses` | PASS |
| chat | DeepSeek × Chat Completions | `deepseek` / `openai-completions` | PASS |
| classifier | TypeSafe × System One，jev-1.13.0 | `typesafe` / `typesafe-system-one` | PASS |
| image | OpenAI × Images，gpt-image-2.5-sunburst-2026-09-08 | `openai` / `openai-images` | PASS（生成、JSON 编辑与 mask） |
| image | Google × Interactions，gemini-nano-banana-2.1 | `google` / `google-interactions` | PASS（生成与参考编辑） |

以 [`live/support-matrix.json`](live/support-matrix.json) 为准。只有矩阵中对应行完整通过，才可宣称该组合受支持；“兼容 OpenAI”不代表任何兼容服务已验收。可调用的模型是目录（`BuiltinCatalog()`，快照见 [`release/catalog-snapshot.json`](release/catalog-snapshot.json)）与 `Binding.AllowedModels` 的交集。

## 快速开始（本地）

本地程序与云端宿主使用同一个 `Client`，只是显式写出云端宿主会按租户解析的那一个租户、绑定和 key。
[`examples/localassembly`](examples/localassembly) 完成这一装配：

```go
asm, err := localassembly.Open(localassembly.Config{
	TenantID: "local",
	Binding: ai.Binding{
		BindingID:      "openai",
		ProviderID:     ai.ProviderOpenAI,
		API:            ai.APIOpenAIResponses,
		Endpoint:       "https://api.openai.com/v1",
		AccountScopeID: "my-openai-account",
		AllowedModels:  []string{"gpt-5-mini"},
	},
	// 由程序指定唯一来源，不会回退到 OPENAI_API_KEY。
	Key: localassembly.KeySource{EnvVar: "BARNESS_OPENAI_KEY"},
	// Policy 为 nil 时使用 localassembly.LocalPolicy()。
})
if err != nil { return err }

req := ai.Request{
	SystemPrompt: "You are concise.",
	Messages:     []ai.Message{ai.UserText("Say hello in French.")},
}
res, err := asm.Client.CompleteSimple(ctx, asm.Scope(), asm.Target("gpt-5-mini"), req,
	ai.SimpleOptions{Reasoning: ai.ThinkingLow})
if err != nil { return err } // *ai.Error；res 仍是完整的 Result

for _, c := range res.Message.Content {
	if t, ok := c.(ai.Text); ok {
		fmt.Println(t.Text)
	}
}
```

## 核心概念

```
宿主 ──CallScope + Target + Request + Options──▶ Client
                                                  │ 1. 校验作用域、请求、图片字节
                                                  │ 2. BindingResolver(tenant, bindingID)
                                                  │ 3. 授权模型（目录 ∩ AllowedModels），映射选项
                                                  │ 4. 准备历史；降级不可信的原生状态
                                                  │ 5. CredentialResolver → 校验快照一致性
                                                  │ 6. 每次尝试：准入 → adapter → Provider
                                                  ▼
                        Stream（事件 + EventEnvelope） /  Result{Message, Metadata}
```

| 概念 | Go 类型 | 说明 |
| --- | --- | --- |
| 调用作用域 | `CallScope` | `TenantID`、`RequestID` 必填。`RequestID` 每个逻辑调用全局唯一；内部重试沿用，每个新调用（包括工具往返的下一轮）必须换新——见 [`examples/requestid`](examples/requestid)。 |
| 目标 | `Target` | `BindingID` + `ModelID`。协议、endpoint、账户与 key 来自绑定，从不来自请求。 |
| 服务绑定 | `Binding` | 租户的一份服务配置：Provider、API、endpoint、账户、可用模型、可用托管工具、`Retry`（零值即不重试）。`Enabled` 与 `Version` 必填；零值一律拒绝。 |
| 凭据 | `Credential` | 带版本的快照；`BindingVersion` 必须等于所解析绑定的版本，否则调用失败（`credential_unavailable`/`consistency`，ADR-0003）。key 以 `ai.Secret` 传入，不会出现在任何输出中。 |
| 请求 | `Request` | `SystemPrompt`、`Messages`、`Tools`，接收时复制。模块从不选择、裁剪或保存历史。 |
| 选项 | `ResponsesOptions`、`AnthropicOptions`、`GeminiOptions`、`ChatOptions` / `SimpleOptions` | 完整选项必须与绑定的 API 匹配。`Nullable[T]` 区分未设置、`null` 与零值。 |
| 结果 | `Result` | `Message`（始终有效，失败时也是）与 `Metadata`（`CallMetadata`：调用归属、各版本、目录哈希、尝试记录、用量完整性、原生状态降级计数）。 |

### 入口

```go
res, err := client.Complete(ctx, scope, target, req, opts)        // 完整协议选项
res, err := client.CompleteSimple(ctx, scope, target, req, simple) // 协议无关选项
s := client.Stream(ctx, scope, target, req, opts)
s := client.StreamSimple(ctx, scope, target, req, simple)
hc := client.WithHooks(hooks) // 可信请求回调；同样四个方法
```

只需要最终消息时用 `Complete*`。`Stream` 会排队每一个事件；从不读取事件的 Stream 在输出超过 `MaxQueuedEvents` / `MaxQueuedEventBytes` 后以 `resource_limit`（`PhaseEventQueue`）结束。

```go
s := client.StreamSimple(ctx, scope, target, req, ai.SimpleOptions{})
defer s.Close() // 幂等；结束 I/O 并归还许可
for s.Next() {   // 单消费者
	switch e := s.Event().(type) {
	case ai.TextDeltaEvent:
		fmt.Print(e.Delta)
	}
	_ = s.Envelope().Call // 本事件的调用归属，供多流汇合时路由
}
res, err := s.Result() // 也可与 Next 并行等待
```

事件与 pi 一致：`start`；可交错、索引稳定的 `text_*`、`thinking_*`、`toolcall_*`（`start`/`delta`/`end`）；恰好一个终结事件 `done` 或 `error`。

### 工具往返

barness-ai 从不执行工具。以 `StopReasonToolUse` 结束的轮次把调用交给宿主：宿主用 `ValidateToolCall` 逐个校验、执行，用 `ToolResultText` 构造结果，再以新的逻辑调用发起下一轮。`ToolCall.Arguments` 只是尽力解析、仅供展示，绝不可执行。以 `length`、`error` 或 `aborted` 结束的轮次一律不执行。见 [`examples/toolloop`](examples/toolloop)。

### 原生状态

本进程产出的消息自带可信来源，可直接回放。从存储（JSON）读回的消息在宿主自行检查后调用 `TrustNativeState(scope, envelope)` 之前都不可信。不可信或（租户、账户、模型）不匹配的原生状态按 pi 的跨模型规则降级，而不是拒绝；降级次数计入 `Metadata`。

### 错误

失败的调用返回完整的 `Result` 以及带稳定 `Code` 与 `Phase` 的 `*ai.Error`：

```go
var e *ai.Error
switch {
case errors.Is(err, &ai.Error{Code: ai.CodeRateLimited}): // Phase 为空时匹配任意阶段
case errors.As(err, &e):
	log.Println(e.Code, e.Phase, e.HTTPStatus, e.ProviderRequestID, e.RetryAfter)
}
```

| 到达 Provider 之前 | 到达 Provider 时或之后 |
| --- | --- |
| `invalid_request`、`tenant_denied`、`binding_not_found`、`credential_unavailable`、`admission_denied` | `upstream_auth`、`rate_limited`、`upstream_error`、`transport`、`protocol` |
| 任意阶段：`canceled`、`deadline_exceeded`、`resource_limit`、`callback_failed` | |

`Error.Message` 按 pi 原样引用 Provider 错误体（仅脱敏 key 形态的文本），可能复述请求内容：只返回给发起调用的租户，不得写入共享日志；需要记录时使用 `Observer`。完整错误表见 [contract.md §4](../docs/barness-ai/contract.md)。

### 资源策略、准入与时限

每个 `Client` 都必须提供显式、有限的 `ResourcePolicy`——没有默认值，也没有“不限”模式（ADR-0002）。字节限额在编码与读取过程中执行。每次尝试（含重试）先取得单租户与全进程许可，再经过宿主的 `Config.Admission`（若设置，ADR-0008）。调用在 context deadline、`CallTimeout` 以及每次尝试的建连/响应头/读空闲/协议超时中最早者到达时结束。

经压力验证、带注释的起点：`localassembly.LocalPolicy`、`hostintegration.CloudInteractivePolicy`、`hostintegration.CloudBatchPolicy`。每个数值的依据写在源码注释中，实测数据见 [docs/barness-ai/README.md](../docs/barness-ai/README.md)。

共用一个 Client 调用聊天、图像和分类时，使用独立的 `localassembly.MixedPolicy()` 或
`hostintegration.CloudMixedPolicy()`。它们显式启用有限子策略，按整体 JSON/base64 的
实测成本限制为 4/8 并发；原聊天容量不构成图像默认值。本地 `OpenOperations` 为每个
操作配置独立 Binding，Keys 按 CredentialRef 装配，同 Provider/账户允许共享引用。
云端 `Host.GenerateImages` / `Host.Classify` 从 Principal 校验身份并传递取消 context。
宿主处理类型化结果与错误、业务置信阈值和转人工；记录使用 Observer。重放命令：

```sh
BARNESS_AI_PRESSURE=1 go test ./ai/e2e -run '^TestMixed' -count=1
go test -race ./ai/e2e -run '^TestMixed' -count=1
```

### 可观测性、用量与成本

`Config.Observer` 通过有界队列异步接收 `call_started`、`attempt_started`、`attempt_finished`、`call_finished`；慢的 Observer 不会阻塞调用，丢弃计入 `Client.ObserverStats()`。记录中不含 key、正文、工具参数或错误文本（ADR-0009）。`Usage` 保持 pi 的数字；`Usage.Cost` 是按目录价格的估算，不是账单。用量是否上报记录在 `Attempt.UsageReporting`——零值 `Usage` 从不表示免费（ADR-0010）。

## 云端宿主接入

参考实现为 [`examples/hostintegration`](examples/hostintegration)。检查清单：

1. 认证结果 → `CallScope`（新的 `RequestID`、`ActorID`）。拒绝自报的 `TenantID` 与他人会话。
2. 按 `(tenant, session)` 读取历史；用 `TrustNativeState` 担保持久化的原生状态。
3. 实现 `BindingResolver` / `CredentialResolver`。若加缓存，键必须区分租户、凭据标识与版本，并公布 TTL 与撤销传播上限。
4. 选择云端策略；`admission_denied` 时返回 429 与 `Retry-After`，排队留在 barness-ai 之外。
5. 需要按厂商账户或跨实例配额时注入 `Admission`；注入 `Observer` 并关注 `ObserverStats`。
6. 在绑定上配置重试（`Binding.Retry`，ADR-0006）。只重试初始请求；已开始的流从不重放。
7. 下游断开时取消该调用的 context 或 `Close` 其 Stream。汇合多个流时按 `EventEnvelope.Call` 路由，而非到达顺序（`hostintegration.Merge`）。

不得从请求 JSON 接受 key、凭据引用、endpoint、`Authorization` 头、代理或回调——它们只来自可信装配。

---

## 写给 coding agent

修改前请阅读本节、[GLOSSARY.md](../GLOSSARY.md) 以及与改动相关的 ADR。仓库级规则见 [AGENTS.md](../AGENTS.md)（单文件不超过 500 行、删除优于兼容层、以 E2E 为主的测试）。

### 不变量——不得破坏

1. **身份只来自可信输入。** `CallScope` 绝不取自请求内容、环境变量或 `context.Context`。所有租户范围的查找都携带 scope 经过 resolver。
2. **限额没有默认值。** 不得新增默认 `ResourcePolicy`、“关闭限额”开关，或无界的缓冲、队列、重试、等待。每个新缓冲都必须受策略限额约束，并以 `CodeResourceLimit` 失败。
3. **快照固定。** 解析完成后，一个调用的所有尝试使用同一份绑定 + 凭据快照；调用内不得重新解析（D2，ADR-0003）。
4. **只重试初始请求**，由 `Binding.Retry` 决定。不重试资源限额失败或已开始的流；不允许 SDK 自行重试。
5. **秘密不外泄。** `Secret` 不得进入消息、事件、错误、观测记录、证据包或日志。barness-ai 不写日志。
6. **没有环境隐式配置。** 不读取 Provider 环境变量、代理环境变量、凭据文件或默认 endpoint。endpoint 必须为 https，`AllowLoopbackHTTP`（仅测试）除外。
7. **默认与 pi 一致。** 可观察行为与冻结的 pi-ai `1.0.0` 一致。任何有意差异都必须登记到 [`e2e/testdata/pidiff/ledger.json`](e2e/testdata/pidiff/ledger.json) 与 [differences.md](../docs/barness-ai/differences.md)；`pending` 条目会阻断发布。
8. **能力按 Provider 确定，而非按协议。** 共享 adapter 不得发送其他 Provider 的字段（如 DeepSeek × Responses 不发送 `store`/`include`/缓存字段）。请求字段从该 Provider 的能力派生。
9. **Observer 从不影响调用；回调可以使调用失败，但不能放宽**鉴权、目标、模型、原生引用或托管工具（ADR-0005）。
10. **工具参数不可执行**，直到对已结束的消息调用 `ValidateToolCall` 通过。

### 包结构

所有生产代码都在单一的 `ai` 包中（扁平结构，按文件名前缀分组）。按约定，未导出类型不跨组泄漏。

| 领域 | 文件 |
| --- | --- |
| 公共入口与流 | `client.go`、`entry.go`、`call.go`、`stream.go`、`event.go`、`assembler.go`（所有入口共用的生产与合并过程）、`partial.go`、`partial_json.go` |
| 领域类型 | `message.go`、`request.go`、`result.go`、`binding.go`、`options.go`、`nullable.go`、`reasoning.go`、`errors.go` |
| 解析与安全 | `preflight.go`、`policy.go`、`admission.go`、`limits.go`、`timeouts.go`、`retry.go`、`transport.go` |
| 历史与原生状态 | `transcript.go`、`history.go`、`replay.go`、`native.go`、`cache_key.go` |
| 工具 | `tool.go`、`tool_validate.go`、`tool_coerce.go`、`tool_json.go`、`argument_text.go` |
| 目录、用量与成本 | `catalog.go`、`usage.go`、`estimate.go`、`anthropic_effort.go` |
| 回调与可观测性 | `hooks.go`、`hooks_run.go`、`observer.go` |
| Adapter 契约 | `adapter.go`（`adapter` 接口与按 `API` 索引的 `registry()`） |
| OpenAI Responses | `responses*.go`、`openai_sdk.go` |
| Anthropic Messages | `anthropic*.go` |
| Gemini Developer API | `gemini*.go`（REST，不用 SDK） |
| Chat Completions | `chat*.go`（`chat_compat.go` 存放各 Provider 的差异） |
| 共享 HTTP/SDK 胶水 | `sdk_middleware.go`、`http_failures.go` |

其他目录：

| 目录 | 用途 |
| --- | --- |
| `examples/` | `localassembly`、`hostintegration`、`toolloop`、`requestid`——最小示例，均由离线 E2E 运行 |
| `e2e/` | 针对公共 API、本地受控 Provider 与宿主替身的离线验收测试；fixture 位于 `e2e/testdata/<protocol>/` |
| `live/` | 真实 API 冒烟（build tag `live`），更新 `support-matrix.json` |
| `release/` | 发布门禁命令、目录快照、追溯映射 |
| `internal/testkit/` | 仅测试使用：受控 Provider、宿主替身、pi oracle（Node）、证据、审计、发布门禁逻辑 |
| `internal/probe`、`internal/clock` | 经 `Config` 注入的测试钩子（宿主无法构造） |

### 常见改动

- **新增模型：** 按纳入标准（ADR-0018）加到 `catalog.go` 中对应的 `builtin*Models()`，再用 `go run ./ai/release/cmd/releasegate -write-snapshot` 重新生成目录快照并审阅差异。
- **在已有协议上新增 Provider：** 新增 `ProviderID` 及其能力集（如 `chat_compat.go` / `responses_capabilities.go`）、`e2e/testdata/` 下的 fixture、live 组合、ADR；pi 没有该路由时还需登记差分账本。
- **新增协议：** 实现 `adapter`（`adapter.go`），在 `registry()` 中注册，新增 `API` 常量、对应 `Options` 类型、E2E 套件（`e2e/<protocol>_*_test.go`）与 ADR。编码与读取时必须执行 `call.limits`。
- **改变可观察行为：** 运行 pi 差分，然后更新账本与 `differences.md`。

### 测试

按 AGENTS.md，E2E 是主要测试手段。测试只驱动公共 `Client`；`e2e/` 中的代码不触及包内部。

```sh
go test ./ai/...                               # 离线 E2E 与包内测试，不访问网络
go test -race ./ai/...
BARNESS_AI_PRESSURE=1 go test ./ai/e2e         # 策略压力场景
BARNESS_AI_PIDIFF=1  go test ./ai/e2e          # 与冻结 pi 的差分
#   需要 Node.js，并先执行：npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node

# 真实冒烟：一个进程只跑一个组合，只持有该组合的 key
BARNESS_AI_LIVE=1 BARNESS_AI_LIVE_COMBO=deepseek-chat \
BARNESS_AI_LIVE_ACCOUNT_ALIAS=<alias> BARNESS_AI_LIVE_KEY_DEEPSEEK_CHAT=<key> \
go test -tags live -count=1 ./ai/live
go run ./ai/live/cmd/supportmatrix <bundle-dir>...   # 把冒烟报告合并进支持矩阵

go run ./ai/release/cmd/releasegate -evaluation <独立中文效果包> [-live <bundle-dir>]...
```

每次运行都会在 `.evidence/` 下写出脱敏的证据包（可用 `BARNESS_AI_EVIDENCE_DIR` 覆盖）。新行为需要一个 E2E 用例，且其证据能被发布门禁追溯（`release/traceability.json`）。

## 延伸阅读

| 文档 | 内容 |
| --- | --- |
| [docs/barness-ai/contract.md](../docs/barness-ai/contract.md) | 公共契约、完整错误表、装配说明 |
| [docs/barness-ai/differences.md](../docs/barness-ai/differences.md) | 与 pi-ai 的全部差异及处置 |
| [docs/barness-ai/README.md](../docs/barness-ai/README.md) | 发布交付物、门禁状态、策略实测 |
| [docs/adr/](../docs/adr/) | ADR-0001 … ADR-0019 |
| [GLOSSARY.md](../GLOSSARY.md) | 领域术语（租户、服务绑定、逻辑调用、原生续接状态……） |
| `go doc -all ./ai` | 包与类型文档 |
