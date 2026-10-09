# barness-ai

[English](README.md) | 简体中文

`github.com/tokenbeat-lab/barness/ai` 是 barness 的模型协议模块：一个可信、并发安全的
`Client`，为**多个租户**向模型服务发起**一次模型操作**。它是
[pi-ai](https://github.com/earendil-works/pi)（冻结基线 `1.0.0`）的 Go 复刻，并在其上增加了租户隔离、显式资源限额与可观测性。

本文面向两类读者：

- **人**：想了解模块做什么、怎么调用、各项保证记录在哪里。
- **Coding agent**：修改本模块前需要先掌握不变量、目录结构、约定与测试命令（见
  [写给 coding agent](#写给-coding-agent)）。

权威来源的优先级依次为：[1.0 升级 spec](../.scratch/barness-ai-pi-1.0/spec.md) 及其基础 spec、[ADR](../docs/adr/)、包文档（`go doc ./ai`）。spec 描述目标，本文的支持声明以实际目录和各路线自身的验收证据为准。术语定义见 [GLOSSARY.md](../GLOSSARY.md)。

## 模型操作

| 操作 | 入口 | 请求 → 结果 | 返回方式 |
| --- | --- | --- | --- |
| `chat` | `Complete`、`CompleteSimple`、`Stream`、`StreamSimple` | `Request` → `Result` / `Stream` | 最终 assistant 消息或 pi-ai 事件流 |
| `image` | `GenerateImages` | `ImagesRequest` → `ImagesResult` | 同步、有序的文本与图片块 |
| `classifier` | `Classify` | `ClassifierRequest` → `ClassifierResult` | 同步、按问题名返回类型化答案 |

`Client` 与 `HookedClient` 均提供这六个入口。图像与分类遵守 context 取消，不产生 delta 事件；接受对应协议选项，`nil` 表示协议默认。`SimpleOptions` 仅用于聊天。

## 职责边界

| barness-ai 负责 | 宿主（你的程序）负责 |
| --- | --- |
| 把宿主选定的消息历史转换为 Provider 请求，以 pi-ai 事件流式返回响应，并给出最终 `AssistantMessage` | 认证调用者，并据此构造每个 `CallScope`——绝不取自请求内容 |
| 通过宿主的 resolver 解析租户的**服务绑定**与**凭据**，并在整个调用中固定该快照 | 保存绑定、凭据与历史；决定每一轮放入哪些历史 |
| 执行有限的 `ResourcePolicy`（字节、队列、并发、时限） | 按自身负载选择策略数值 |
| 仅在来源匹配时回放 Provider 原生状态（签名、加密推理） | 用 `TrustNativeState` 为持久化的原生状态担保 |
| 报告工具调用，按需校验参数 | 执行工具，并把每个下一轮作为新的逻辑调用发起 |
| 以 unary 响应生成/编辑原生图像或返回完整校验的分类答案 | 显式启用操作绑定、目录型号与有限子策略；决定业务阈值与转人工规则 |
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
| classifier | TypeSafe × System One，jev-1.13.0 | `typesafe` / `typesafe-system-one` | PASS（上下文探针返回 400；422 形状未确认） |
| image | OpenAI × Images，gpt-image-2.5-sunburst-2026-09-08 | `openai` / `openai-images` | PASS（生成、JSON 编辑与 mask） |
| image | Google × Interactions v1beta，gemini-nano-banana-2.1 | `google` / `google-interactions` | PASS（生成与参考编辑，1K / 1:1） |

上表归纳 **2026-10-08** 九条路线各自的完整真实运行，以 [`live/support-matrix.json`](live/support-matrix.json) 为准。只有矩阵中对应行完整通过，才可宣称该组合受支持；“兼容 OpenAI”不代表任何兼容服务已验收。可调用的型号是目录（`BuiltinCatalog()`，快照见 [`release/catalog-snapshot.json`](release/catalog-snapshot.json)）与 `Binding.AllowedModels` 在 `(Operation, Provider, API, ModelID)` 下的交集。

两条原生图像路线与 DeepSeek Responses 均登记为扩展路径，其自身证据不计为 pi 差分通过。发布结论与后续离线修复见[发布说明](../docs/barness-ai/README.md)。

## 快速开始（本地）

需要 **Go 1.26.2 或更高版本**（见 [go.mod](../go.mod)）。下列示例为函数体片段，使用 `context`、`fmt`、`encoding/json`、`ai` 与 `ai/examples/localassembly`；调用者提供 `ctx` 并处理返回的错误。

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
	Policy: localassembly.LocalPolicy(), // 聊天示例策略；使用前核对设计负载
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

### 图像生成与分类

`OpenOperations` 在一个 Client 上装配多个绑定，要求显式策略，并保留各绑定的 `Enabled` 选择。程序只读取自己指定的 key 来源，按 `CredentialRef` 索引：

```go
asm, err := localassembly.OpenOperations(localassembly.OperationsConfig{
	TenantID: "local",
	Bindings: []ai.Binding{
		{
			BindingID: "images", Operation: ai.OperationImage, Enabled: true,
			ProviderID: ai.ProviderOpenAI, API: ai.APIOpenAIImages,
			Endpoint: "https://api.openai.com/v1", AccountScopeID: "my-openai-account",
			CredentialRef: "openai-key",
			AllowedModels: []string{"gpt-image-2.5-sunburst-2026-09-08"},
		},
		{
			BindingID: "classifier", Operation: ai.OperationClassifier, Enabled: true,
			ProviderID: ai.ProviderTypeSafe, API: ai.APITypeSafeSystemOne,
			Endpoint: "https://api.typesafe.ai", AccountScopeID: "my-typesafe-account",
			CredentialRef: "typesafe-key", AllowedModels: []string{"jev-1.13.0"},
		},
	},
	Keys: map[string]localassembly.KeySource{
		"openai-key":   {EnvVar: "BARNESS_OPENAI_KEY"},
		"typesafe-key": {EnvVar: "BARNESS_TYPESAFE_KEY"},
	},
	Policy: localassembly.MixedPolicy(),
})
if err != nil { return err }

images, err := asm.Client.GenerateImages(ctx, asm.Scope(),
	ai.Target{BindingID: "images", ModelID: "gpt-image-2.5-sunburst-2026-09-08"},
	ai.ImagesRequest{Prompt: "A small red sailboat on a calm lake."},
	ai.OpenAIImagesOptions{Size: ai.Value("1024x1024"), Quality: ai.Value("low")})
if err != nil { return err } // images 仍保留用量与调用元数据
for _, block := range images.Content {
	if img, ok := block.(ai.ImageOutputImage); ok {
		// img.Data 是严格 base64；宿主决定如何保存或展示。
		fmt.Println(img.MimeType)
	}
}

classified, err := asm.Client.Classify(ctx, asm.Scope(),
	ai.Target{BindingID: "classifier", ModelID: "jev-1.13.0"},
	ai.ClassifierRequest{
		State: json.RawMessage(`{"message":"I cannot sign in."}`),
		Questions: map[string]ai.ClassifierQuestion{
			"needs_help": ai.BoolQuestion{Instructions: json.RawMessage(`"Does the user need support?"`)},
		},
	}, nil)
if err != nil { return err } // 答案校验失败时仍保留已上报用量
fmt.Println(classified.Answers["needs_help"].(ai.BoolAnswer).Probability)
```

OpenAI 无 `ReferenceImages` 时生成，有内联 `ai.Image` 参考图时走 JSON 编辑。
`OpenAIImagesOptions` 提供数量、尺寸、质量、背景、格式、压缩、mask、输入保真度与审核级别，均受型号能力限制。
内置快照只开放一参考/一输出、1024×1024、low/medium 质量、mask 与透明背景；输入保真度须省略。

Google 使用 `OperationImage`、`ProviderGoogle`、`APIGoogleInteractions`，endpoint 为
`https://generativelanguage.googleapis.com/v1beta`，型号为 `gemini-nano-banana-2.1`。
`GoogleImagesOptions` 只提供 `AspectRatio` 与 `ImageSize`，内置型号只开放一参考/一输出、`1:1`、`1K`。
调用固定同步、`store=false`、无续接、输出内联；请求省略 `delivery`（ADR-0023），不保证精确输出数量。
两条路线均不读取文件、不下载 URL、不接受厂商 file ID。

分类状态与问题说明支持 JSON 字符串、对象或数组。`ChoiceQuestion` 返回 `ChoiceAnswer`（选项、分布、置信度），
`ScoreQuestion` 返回 `ScoreAnswer`（期望分值、置信度、可选分布与图例），`BoolQuestion` 返回“是”的概率。
评分标准按顺序排列，下标从零开始。答案按回调后的最终问题集合整体校验；不合格即拒绝，不重新归一化。
`TypeSafeOptions{}` 是空选项类型。[中文评估宿主](examples/chineseeval/README.md) 独立报告业务效果；[24/24 合成样本结果](../.scratch/barness-ai-pi-1.0/chinese-evaluation-evidence/README.md) 不代表生产准确率。

## 核心概念

```
宿主 ──CallScope + Target + 对应操作的请求/选项──▶ Client
  1. 入口固定 Operation；校验作用域与输入限额
  2. 解析 Binding；校验操作、对应类型目录与 AllowedModels
  3. 校验选项/输入；仅聊天：准备历史与原生状态
  4. 解析 Credential；固定一致的绑定 + 凭据 + 目录快照
  5. 回调后冻结请求；每次尝试：准入 → 协议实现 → Provider
  6. chat：合并事件/消息；unary：读取并校验完整响应
     → Result / Stream / ImagesResult / ClassifierResult，均附 CallMetadata
```

| 概念 | Go 类型 | 说明 |
| --- | --- | --- |
| 调用作用域 | `CallScope` | `TenantID`、`RequestID` 必填，入口设置 `Operation`。`RequestID` 每个逻辑调用全局唯一；内部重试沿用，新调用换新。它用于归属，不是去重键——见 [`examples/requestid`](examples/requestid)。 |
| 目标 | `Target` | `BindingID` + `ModelID`。协议、endpoint、账户与 key 来自绑定，从不来自请求。 |
| 服务绑定 | `Binding` | 固定一种操作、Provider、API、endpoint、账户，以及型号/工具白名单、`Retry`（零值不重试）。`Operation` 零值**仅表示 chat**，未知非空值拒绝；`Enabled` 与 `Version` 仍必填。 |
| 凭据 | `Credential` | 带版本的快照；`BindingVersion` 必须等于所解析绑定的版本，否则调用失败（`credential_unavailable`/`consistency`，ADR-0003）。key 以 `ai.Secret` 传入，不会出现在任何输出中。 |
| 请求 | `Request`、`ImagesRequest`、`ClassifierRequest` | 分别为聊天历史/工具、图像提示/参考图、分类状态/问题。可变输入独立复制，unary 先检查已知大小再复制。 |
| 选项 | `ResponsesOptions`、`AnthropicOptions`、`GeminiOptions`、`ChatOptions` / `SimpleOptions`；`OpenAIImagesOptions`、`GoogleImagesOptions`、`TypeSafeOptions` | 完整选项必须与绑定 API 匹配。`Nullable[T]` 区分未设置、`null` 与零值；图像选项拒绝显式 null。 |
| 结果 | `Result`、`ImagesResult`、`ClassifierResult` | 均含 `Metadata`：操作、调用归属、版本、目录哈希、尝试与用量完整性。聊天失败时仍有有效消息；unary 失败时内容/答案为空，已上报用量保留。 |
| 目录 | `Model`、`ImageModel`、`ClassifierModel` | `Catalog.Lookup` / `ModelsOf` 仅查聊天；图像用 `LookupImage` / `ImageModelsOf`，分类用 `LookupClassifier` / `ClassifierModelsOf`。`Client.Catalog()` 返回独立快照，发现能力不授予调用权限。 |

用不匹配的操作入口调用绑定，会在读取凭据和创建尝试前以 `tenant_denied/capability` 拒绝。同 Provider、同账户的多个操作绑定可以共用凭据引用，授权仍分别执行。

### 入口

```go
res, err := client.Complete(ctx, scope, target, req, opts)        // 完整协议选项
res, err := client.CompleteSimple(ctx, scope, target, req, simple) // 协议无关选项
s := client.Stream(ctx, scope, target, req, opts)
s := client.StreamSimple(ctx, scope, target, req, simple)
hc := client.WithHooks(hooks) // 可信请求回调；提供全部六个入口
```

请求头与 payload 回调每个逻辑调用只执行一次，重试复用冻结后的请求。
unary 响应回调在成功 HTTP 响应后、读取响应体前执行一次；Gemini 聊天继续不执行响应回调。

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

失败的调用返回对应操作的结果，以及带稳定 `Code` 与 `Phase` 的 `*ai.Error`：

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

unary 的响应读取、解码与输出校验失败属于 `PhaseResponse`。成功 HTTP 响应后的失败从不重放，已开始的聊天流同样不重放；只有 `Binding.Retry` 显式允许时才重试初始请求。

### 资源策略、准入与时限

每个 `Client` 都必须提供显式、有限的 `ResourcePolicy`——没有默认值，也没有“不限”模式（ADR-0002）。字节限额在编码与读取过程中执行。每次尝试（含重试）先取得单租户与全进程许可，再经过宿主的 `Config.Admission`（若设置，ADR-0008）。调用在 context deadline、`CallTimeout` 以及每次尝试的建连/响应头/读空闲/协议超时中最早者到达时结束。

经压力验证、带注释的起点：`localassembly.LocalPolicy`、`hostintegration.CloudInteractivePolicy`、`hostintegration.CloudBatchPolicy`。每个数值的依据写在源码注释中，实测数据见 [docs/barness-ai/README.md](../docs/barness-ai/README.md)。

`ResourcePolicy.Image == nil` 禁用图像，`Classifier == nil` 禁用分类。启用的子策略容量必须全部为正，构造时深复制：
图像限制输入/输出数量、单张与总输出字节；分类限制问题数、状态字节与单个问题字节。
全局请求/输出字节、并发与时限约束三种操作。unary JSON 按 `MaxOutputBytes` 读取，不受仅用于 SSE 的 `MaxFrameBytes` 限制；实际限额同时取协议/型号硬限制。

共用一个 Client 调用聊天、图像和分类时，使用独立的 `localassembly.MixedPolicy()` 或
`hostintegration.CloudMixedPolicy()`。它们显式启用有限子策略，按整体 JSON/base64 的
实测成本限制为 4/8 并发；原聊天容量不构成图像默认值。本地 `OpenOperations` 为每个
操作配置独立 Binding，Keys 按 CredentialRef 装配。
云端 `Host.GenerateImages` / `Host.Classify` 从 Principal 校验身份并传递取消 context。
宿主处理类型化结果与错误、业务置信阈值和转人工；记录使用 Observer。重放命令：

```sh
BARNESS_AI_PRESSURE=1 go test ./ai/e2e -run '^TestMixed' -count=1
go test -race ./ai/e2e -run '^TestMixed' -count=1
```

### 可观测性、用量与成本

`Config.Observer` 通过有界队列异步接收 `call_started`、`attempt_started`、`attempt_finished`、`call_finished`；慢的 Observer 不会阻塞调用，丢弃计入 `Client.ObserverStats()`。记录中不含 key、正文、工具参数或错误文本（ADR-0009）。`Usage` 保持 pi 的数字；`Usage.Cost` 是按目录价格的估算，不是账单。用量是否上报记录在 `Attempt.UsageReporting`——零值 `Usage` 从不表示免费（ADR-0010）。

图像调用增加可选 `Usage.Modalities` token 明细，按文本/图片费率估算；缺明细仍标为部分或未上报。
Google 思考 token 计入输出一次，并单独记录为推理量；TypeSafe 只对输入 token 计价。
观测包含 `Operation` 与用量，不含提示、图片、分类状态或答案。

### pi-ai 1.0.0 基线的聊天行为

- 无法解析或非有限的重试等待时间回退为有界、带抖动的指数退避。
- Responses 按响应报告的服务等级估价，缺失时取选项：flex 为 0.5 倍，priority/fast 为 2 倍（gpt-5.5 为 2.5 倍）。
- Anthropic 后续增量中的 1 小时缓存写入明细覆盖此前值。
- Responses/Chat 只合并一次型号默认 `SamplingParams`，调用级值优先；full/simple 入口一致，授权与状态等保留字段仍受保护。
- Responses 拒绝未完成的工具调用，包括被重复输出索引覆盖的旧块；宿主仍需在执行前校验工具参数。

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
4. **只重试初始请求**，由 `Binding.Retry` 决定。不重试资源限额失败、已开始的流或 unary 2xx 后的失败；不允许 SDK 自行重试。
5. **秘密不外泄。** `Secret` 不得进入消息、事件、错误、观测记录、证据包或日志。barness-ai 不写日志。
6. **没有环境隐式配置。** 不读取 Provider 环境变量、代理环境变量、凭据文件或默认 endpoint。endpoint 必须为 https，`AllowLoopbackHTTP`（仅测试）除外。
7. **默认与 pi 一致。** 可观察行为与冻结的 pi-ai `1.0.0` 一致。任何有意差异都必须登记到 [`e2e/testdata/pidiff/ledger.json`](e2e/testdata/pidiff/ledger.json) 与 [differences.md](../docs/barness-ai/differences.md)；`pending` 条目会阻断发布。
8. **能力按 Provider 确定，而非按协议。** 共享 adapter 不得发送其他 Provider 的字段（如 DeepSeek × Responses 不发送 `store`/`include`/缓存字段）。请求字段从该 Provider 的能力派生。
9. **Observer 从不影响调用；回调可以使调用失败，但不能放宽**操作、鉴权、目标、模型、原生引用或托管工具。unary 最终输入必须重新校验（ADR-0005）。
10. **工具参数不可执行**，直到对已结束的消息调用 `ValidateToolCall` 通过。
11. **操作独立授权。** 绑定操作零值只允许聊天，按操作查目录后才读凭据（ADR-0020）。
12. **unary 输出整体发布。** 所有图片/答案校验后一起交出；任何失败均保留已上报用量并归还许可。

### 包结构

所有生产代码都在单一的 `ai` 包中（扁平结构，按文件名前缀分组）。按约定，未导出类型不跨组泄漏。

| 领域 | 文件 |
| --- | --- |
| 共同调用运行时 | `client.go`、`call_runtime.go`、`unary*.go`（生命周期、固定快照、unary HTTP/JSON） |
| 聊天入口与流 | `entry.go`、`call.go`、`stream.go`、`event.go`、`assembler.go`、`partial*.go` |
| 领域类型 | `operation.go`、`message.go`、`request.go`、`result.go`、`binding.go`、`options.go`、`nullable.go`、`reasoning.go`、`errors.go` |
| 解析与安全 | `preflight.go`、`policy.go`、`admission.go`、`limits.go`、`timeouts.go`、`retry.go`、`transport.go` |
| 历史与原生状态 | `transcript.go`、`history.go`、`replay.go`、`native.go`、`cache_key.go` |
| 工具 | `tool.go`、`tool_validate.go`、`tool_coerce.go`、`tool_json.go`、`argument_text.go` |
| 目录 | `catalog*.go`（类型化型号、查询、校验、内置数据、快照与价格） |
| 用量与成本 | `usage.go`、`estimate.go`、`anthropic_effort.go` |
| 回调与可观测性 | `hooks.go`、`hooks_run.go`、`observer.go` |
| 聊天 adapter 契约 | `adapter.go`（`adapter` 接口与按 `API` 索引的固定 `registry()`）；unary 在 `images.go` / `classifier.go` 按操作与协议显式分派 |
| OpenAI Responses | `responses*.go`、`openai_sdk.go` |
| Anthropic Messages | `anthropic*.go` |
| Gemini Developer API | `gemini*.go`（REST，不用 SDK） |
| Chat Completions | `chat*.go`（`chat_compat.go` 存放各 Provider 的差异） |
| 原生图像 | `images.go`、`image*.go`、`openai_images*.go`、`google_images*.go` |
| TypeSafe 分类 | `classifier*.go`、`typesafe*.go` |
| 共享 HTTP/SDK 胶水 | `sdk_middleware.go`、`http_failures.go` |

其他目录：

| 目录 | 用途 |
| --- | --- |
| `examples/` | `localassembly`、`hostintegration`（聊天/混合操作）、`toolloop`、`requestid`、`chineseeval`（独立业务评估），均有离线 E2E |
| `e2e/` | 针对公共 API、本地受控 Provider 与宿主替身的离线验收测试；fixture 位于 `e2e/testdata/<protocol>/` |
| `live/` | 真实 API 冒烟（build tag `live`），更新 `support-matrix.json` |
| `release/` | 发布门禁命令、目录快照、追溯映射 |
| `internal/testkit/` | 仅测试使用：受控 Provider、宿主替身、pi oracle（Node）、证据、审计、发布门禁逻辑 |
| `internal/probe`、`internal/clock` | 经 `Config` 注入的测试钩子（宿主无法构造） |

### 常见改动

- **新增型号：** 按纳入标准（ADR-0018）更新 `catalog_builtin.go`、`catalog_images.go` 或 `catalog_typesafe.go`，再用 `go run ./ai/release/cmd/releasegate -write-snapshot` 重新生成目录快照并审阅差异。
- **在已有协议上新增 Provider：** 新增 `ProviderID` 及其能力集（如 `chat_compat.go` / `responses_capabilities.go`）、`e2e/testdata/` 下的 fixture、live 组合、ADR；pi 没有该路由时还需登记差分账本。
- **新增协议：** 明确操作、`API`、类型化选项与结果转换，添加 E2E 套件与 ADR。聊天实现 `adapter` / `registry()`；unary 复用共同运行时与显式分派。编码、读取和校验均执行限额，不提前添加注册抽象。
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
BARNESS_AI_LIVE_ACCOUNT_ALIAS='test-account' \
go test -tags live -count=1 ./ai/live
# 提前向该进程注入 BARNESS_AI_LIVE_KEY_DEEPSEEK_CHAT。
go run ./ai/live/cmd/supportmatrix <bundle-dir>...   # 把冒烟报告合并进支持矩阵

go run ./ai/examples/chineseeval/cmd/chineseeval       # 不联网，生成 NOT_RUN 包
go run ./ai/examples/chineseeval/cmd/chineseeval -verify <评估包>
go run ./ai/release/cmd/releasegate -evaluation <独立中文效果包> [-live <bundle-dir>]...
```

每次 E2E 运行都会在 `.evidence/` 下写出脱敏的证据包（可用 `BARNESS_AI_EVIDENCE_DIR` 覆盖）。新行为需要一个 E2E 用例，且其证据能被发布门禁追溯（`release/traceability.json`）。
完整发布门禁需要九条路线各自的真实证据与独立真实中文评估；离线 fixture 与 NOT_RUN 报告不能替代。

## 延伸阅读

| 文档 | 内容 |
| --- | --- |
| [docs/barness-ai/contract.md](../docs/barness-ai/contract.md) | 公共契约、完整错误表、装配说明 |
| [docs/barness-ai/differences.md](../docs/barness-ai/differences.md) | 与 pi-ai 的全部差异及处置 |
| [docs/barness-ai/README.md](../docs/barness-ai/README.md) | 发布交付物、门禁状态、策略实测 |
| [docs/adr/](../docs/adr/) | ADR-0001 … ADR-0024 |
| [GLOSSARY.md](../GLOSSARY.md) | 领域术语（租户、服务绑定、逻辑调用、原生续接状态……） |
| `go doc -all ./ai` | 包与类型文档 |
