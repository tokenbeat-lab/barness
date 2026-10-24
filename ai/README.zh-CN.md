# barness-ai

> 当前验证步骤见[测试流程 SOP](../docs/agents/verification.md)，决策见 [ADR-0051](../docs/adr/0051-scoped-verification-receipts.md)，实施事实见[唯一当前状态](../.scratch/barness-ai-test-process-review/status.md)；以下旧验收仅适用于其原受测源码。

> **目录与发布状态（2026-10-19）**：`2026-10-18.1-minimax` 包含96个聊天身份、两个图像型号和三个分类型号，发布合同要求十六协议组合、十七自身支持身份。中国区标准GLM-5.3/5.2各有Chat、Responses、Messages独立绑定；[Host示例](examples/glmhost/README.md)覆盖标准Key和full/simple/WithHooks。Jev1.14图像与Box继续要求实验图像开关和显式授权。当前支持矩阵17身份全部PASS；当前来源资格、完整发布门禁与最终审查均已完成，交付为 **COMPLETE**。见[当前交付状态](../.scratch/barness-ai-minimax/delivery-status.md)。

> **MiniMax 国际站**：`BuiltinCatalog()` 已包含 `minimax` × Chat、Responses、Messages 的三个 `MiniMax-M3` 身份，端点为 `api.minimax.io/v1` 或 `api.minimax.io/anthropic`。三份自身[当前正式报告](../.scratch/barness-ai-minimax/evidence/ticket-19/actual/remaining/README.md)均PASS，并已核对最终测量来源的适用性。M3.1仍为未推广候选，线上自身报告NOT_RUN，M Plan权益与正式价格未知。[Host示例](examples/minimaxhost/README.zh-CN.md)要求明确账户和凭据用途。[最终闭合](../.scratch/barness-ai-minimax/evidence/ticket-21/final-closure/README.md)记录11组来源资格、完整门禁和精确矩阵发布通过；原失败和旧来源map保留历史。

[English](README.md) | 简体中文

`github.com/tokenbeat-lab/barness/ai` 是 barness 的模型协议模块：一个可信、并发安全的
`Client`，为**多个租户**向模型服务发起**一次模型操作**。它是
[pi-ai](https://github.com/earendil-works/pi)的 Go 复刻，以其为架构参考，并在其上增加了租户隔离、显式资源限额与可观测性。

本文面向两类读者：

- **人**：想了解模块做什么、怎么调用、各项保证记录在哪里。
- **Coding agent**：修改本模块前需要先掌握不变量、目录结构、约定与测试命令（见
  [写给 coding agent](#写给-coding-agent)）。

设计要求见[模型操作 spec](../.scratch/barness-ai-pi-1.0/spec.md)、[分档采样 spec](../.scratch/barness-ai-thinking-sampling/spec.md)、[Decisions spec](../.scratch/barness-ai-openai-decisions/spec.md) 与 [ADR](../docs/adr/)。本文 API 以当前实现（`go doc ./ai`）为准，支持声明依据正式目录和各路线自身的验收证据。术语定义见 [GLOSSARY.md](../GLOSSARY.md)。

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

| 操作 | Provider × 协议 | `ProviderID` / `API` | 自身真实报告历史 |
| --- | --- | --- | --- |
| chat | OpenAI × Responses | `openai` / `openai-responses` | PASS |
| chat | Anthropic × Messages | `anthropic` / `anthropic-messages` | PASS |
| chat | Google × Gemini Developer API | `google` / `google-generative-ai` | PASS |
| chat | OpenAI × Chat Completions | `openai` / `openai-completions` | PASS（协议本身不返回可回放的推理） |
| chat | DeepSeek × Responses（扩展路径，pi 无此路由） | `deepseek` / `openai-responses` | PASS |
| chat | DeepSeek × Chat Completions | `deepseek` / `openai-completions` | PASS |
| classifier | TypeSafe × System One，jev-1.13.0 | `typesafe` / `typesafe-system-one` | PASS（上下文探针返回 400；422 形状未确认） |
| classifier | OpenAI × Decisions，gpt-6-luna | `openai` / `openai-decisions` | PASS（当前来源资格已验证） |
| classifier | TypeSafe × System One, jev-1.14.0 | `typesafe` / `typesafe-system-one` | PASS（当前来源资格已验证） |
| chat | 中国区智谱 × Chat, glm-5.3 / glm-5.2 | `zhipu` / `openai-completions` | PASS（当前来源资格已验证） |
| chat | 中国区智谱 × Responses, glm-5.3 / glm-5.2 | `zhipu` / `openai-responses` | PASS（当前来源资格已验证） |
| chat | 中国区智谱 × Messages, glm-5.3 / glm-5.2 | `zhipu` / `anthropic-messages` | PASS（当前完整22场景；原取消失败保留） |
| chat | 国际站MiniMax-M3 × Chat | `minimax` / `openai-completions` | PASS（当前自身正式证据） |
| chat | 国际站MiniMax-M3 × Responses | `minimax` / `openai-responses` | PASS（当前自身正式证据） |
| chat | 国际站MiniMax-M3 × Messages | `minimax` / `anthropic-messages` | PASS（当前自身正式证据） |
| image | OpenAI × Images，gpt-image-2.5-sunburst-2026-09-08 | `openai` / `openai-images` | PASS（生成、JSON 编辑与 mask） |
| image | Google × Interactions v1beta，gemini-nano-banana-2.1 | `google` / `google-interactions` | PASS（生成与参考编辑，1K / 1:1） |

历史十组合曾在 **2026-10-13** 的正式目录 `2026-10-13.1` 与 schema 3 下各自重跑，路线状态见 [`live/support-matrix.json`](live/support-matrix.json)。只有矩阵中对应行完整通过，才可宣称该组合受支持；“兼容 OpenAI”不代表任何兼容服务已验收。可调用的型号是目录（`BuiltinCatalog()`，快照见 [`release/catalog-snapshot.json`](release/catalog-snapshot.json)）与 `Binding.AllowedModels` 在 `(Operation, Provider, API, ModelID)` 下的交集。

两条原生图像路线与 DeepSeek Responses 均登记为扩展路径，其自身证据不计为 pi 差分通过。发布结论与后续离线修复见[发布说明](../docs/barness-ai/README.md)。
内置目录包含 96 个聊天身份、两个图像型号与三个分类型号；同名型号可以分属不同操作或协议。

### 历史验证与发布基线

[当前MiniMax交付](../.scratch/barness-ai-minimax/delivery-status.md)已COMPLETE：测量来源`78843eb`两种vet、普通/race实际命令均exit0，两套主包各12495PASS、0FAIL、0NOT_RUN、audit0。17份自身支持与4项效果全部PASS，严格来源资格11/11；GLM Messages cancel在完整报告中1146ms PASS。后续文档/证据提交保持1078份选定源码字节，不冒充新Go执行。以下基线保持各自历史结果。

上一完整基线为[票14正式结论](../.scratch/jev-1.14-multimodal-box/14-remaining-real/verification.md)：11份自身真实支持、四份独立效果、normal/race各9098例、861份冻结pi案例和全部门禁通过。[原压力票07](../.scratch/barness-ai-thinking-sampling/issues/07-pressure-stream-integrity.md) 已由自身因果证据解决，历史失败保留。

[票13历史验收](../.scratch/barness-ai-glm/evidence/ticket-13/README.md) 在 `3d3cd0f` 为 PASS：normal 11,049、race 11,049 例，全部十门禁、两种vet及20份输入审计通过。当时目录 `2026-10-17.2` 保留十四份自身支持、四份独立效果；本票实际HTTP为282次：GLM66、已封闭原旧路线轮180、必要更正轮36。旧失败/捕获保留历史；[校验索引](../.scratch/barness-ai-glm/evidence/ticket-13/index.json)提供精确源码适用范围、有限需求/故事追溯和完整基线回滚。

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
			ProviderID: ai.ProviderOpenAI, API: ai.APIOpenAIDecisions,
			Endpoint: "https://api.openai.com/v1", AccountScopeID: "my-openai-account",
			CredentialRef: "openai-key", AllowedModels: []string{"gpt-6-luna"},
		},
	},
	Keys: map[string]localassembly.KeySource{
		"openai-key": {EnvVar: "BARNESS_OPENAI_KEY"},
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
	ai.Target{BindingID: "classifier", ModelID: "gpt-6-luna"},
	ai.ClassifierRequest{
		State: json.RawMessage(`{"message":"I cannot sign in."}`),
		Questions: map[string]ai.ClassifierQuestion{
			"needs_help": ai.BoolQuestion{Instructions: json.RawMessage(`"Does the user need support?"`)},
		},
	}, ai.OpenAIDecisionsOptions{}) // nil 同样选择协议默认
if err != nil { return err } // 答案校验失败时仍保留已上报用量
switch answer := classified.Answers["needs_help"].(type) {
case ai.BoolAnswer:
	fmt.Println(answer.Probability)
case ai.RefusalAnswer:
	fmt.Println("needs_help: refused") // 宿主决定重问或转人工
}
```

OpenAI 无 `ReferenceImages` 时生成，有内联 `ai.Image` 参考图时走 JSON 编辑。
`OpenAIImagesOptions` 提供数量、尺寸、质量、背景、格式、压缩、mask、输入保真度与审核级别，均受型号能力限制。
内置快照只开放一参考/一输出、1024×1024、low/medium 质量、mask 与透明背景；输入保真度须省略。

Google 使用 `OperationImage`、`ProviderGoogle`、`APIGoogleInteractions`，endpoint 为
`https://generativelanguage.googleapis.com/v1beta`，型号为 `gemini-nano-banana-2.1`。
`GoogleImagesOptions` 只提供 `AspectRatio` 与 `ImageSize`，内置型号只开放一参考/一输出、`1:1`、`1K`。
调用固定同步、`store=false`、无续接、输出内联；请求省略 `delivery`（ADR-0023），不保证精确输出数量。
两条路线均不读取文件、不下载 URL、不接受厂商 file ID。

### 分类输入与答案

| 内置分类器 | 状态 / 说明格式 | 问题与选项 |
| --- | --- | --- |
| TypeSafe `jev-1.13.0` | 仅文本；State 与说明支持 JSON 字符串/对象/数组 | 单选 1–255 个候选；评分 2–10 个有序等级；是非题；`TypeSafeOptions{}` |
| TypeSafe `jev-1.14.0` | 文本与明确选择的实验 PNG/JPEG；context/State/Instructions/Criteria 四位置 | 单选、评分、是非和 Box；图片/Box 需 `TypeSafeOptions{ExperimentalImages: true}` |
| OpenAI `gpt-6-luna` | 文本与获授权的内联图片；State 支持 JSON 字符串/对象/数组，说明为字符串 | 单选 2–255 个候选；评分 2–10 个有序等级；是非题；`OpenAIDecisionsOptions{}` |

两条路线都使用 `OperationClassifier`，`nil` 选择各自协议默认；typed nil 和错协议选项在凭据前拒绝。
Decisions 将 State 字符串转原文，对象/数组转保留数字精度的 compact JSON 文本。
其目录上下文窗口 922,000 来自冻结 pi 元数据；无日期服务 alias 的实际权重不由目录 Hash 固定。

`ChoiceAnswer` 保留选项、分布与置信度，`ScoreAnswer` 保留期望分、分布及验证后的图例，
`BoolAnswer` 保留“是”的概率。评分等级从零开始。Decisions 的合法拒答为 `RefusalAnswer{}`
（`{"type":"refusal"}`），终态 stop/nil error；混合或全部拒答仍保留每个问题名。
宿主显式处理拒答，决定业务阈值、以新逻辑调用重问或转人工。
答案按 Hook 后冻结的最终请求整体校验；坏答案清空集合，但保留独立有效的已上报用量。

#### 内联分类图片

`ClassifierRequest.Images` 须同时满足三个独立条件：

1. `ResourcePolicy.Classifier.Images` 明确设置正数 `MaxImages` / `MaxTotalImageBytes`。
2. 型号声明 image 输入、`MaxInputImages` 与允许的 MIME。内置 Luna 最多 128 张 PNG/JPEG/WebP，
   宿主策略可更小；Jev 1.13 拒绝图片，固定 1.14 仍需实验图片选项。
3. 可信宿主配置授予 `Binding.AllowClassifierImages: true`。

上例绑定在开启第三项前只允许文本。传入有序的 `ai.Image{Data: base64Data, MimeType: "image/png"}`，
同时提供必需的 JSON `State`（可为 `{}`）。严格 base64 与图片格式头在不分配像素栅格的情况下校验。
全局 `MaxImageBytes` 限制每图 decoded 字节，子策略限制总 decoded 字节，
`MaxRequestBytes` 限制含 base64/data URL 的完整 native JSON。Hook 同样受固定权限与预算约束。
空 Images 保留纯文本 wire；不支持 URL、路径、file ID 或自动缩放。

最终 Hook、错误与撤销步骤见 [Decisions 公共契约](../docs/barness-ai/decisions.md)
及[可执行 Jev/Luna 迁移示例](examples/classifiermigration/README.md)。
[中文评估宿主](examples/chineseeval/README.md) 按路线独立报告业务效果；固定合成任务的 24/24 不保证生产准确率。

### 统一图片 Host 示例

运行 `go run ./ai/examples/classifierimages`，无需凭据即可经两个公开 Classify 入口重放受控 HTTP。
[完整示例](examples/classifierimages/README.md)用同一 PNG/JPEG context 请求只替换 Target/Options，
并演示显式复用、State/说明/标准图片、视觉图例、Box、拒答、撤销与可核验脱敏证据。
内置固定 `jev-1.14.0` 已具备实验图片位置和 Box，Host 仍需有限图片策略、型号白名单与
`TypeSafeOptions{ExperimentalImages: true}`。Decisions 在凭据前拒绝局部图和 Box。
受控 PASS 不代表真实协议或视觉质量通过；正式支持独立核验 1.14 自身证据，ADR-0030 保持 proposed。

`ImageReferences` 完整声明有序用途，value Pointer 指向现存 null 槽位；来源预算与
`ClassifierImagePolicy.References` 展开预算独立，Hook 后仍复核权限和限额。
结果 `ImageInputs` 是最终 MIME/decoded 文件 SHA-256，不能使用入口下标解释。
`ScoreAnswer.LegendImageReferences` 的 Pointer 相对 Legend 根，索引最终 ImageInputs。
`BoxAnswer` 保留必填 nullable Confidence 和归一化四边，与独立存在性 Bool 无条件关系。
Host 明确多图搜索关联，按实际方向、尺寸与裁剪转换像素，自行决定业务阈值。
失败不发布 Answers/ImageInputs，独立有效 usage 保留。Host 撤销 Binding/图片权限阻止新调用，
不自动切协议。真实模式必须显式 `-live -route -config -key-env` 注入 Host 配置，结果标为 OBSERVED。

## 核心概念

| 概念 | Go 类型 | 说明 |
| --- | --- | --- |
| 调用作用域 | `CallScope` | `TenantID`、`RequestID` 必填，入口设置 `Operation`。`RequestID` 每个逻辑调用全局唯一；内部重试沿用，新调用换新。它用于归属，不是去重键——见 [`examples/requestid`](examples/requestid)。 |
| 目标 | `Target` | `BindingID` + `ModelID`。协议、endpoint、账户与 key 来自绑定，从不来自请求。 |
| 服务绑定 | `Binding` | 固定一种操作、Provider、API、endpoint、账户，以及型号/工具白名单、`Retry`（零值不重试）。`Operation` 零值**仅表示 chat**，未知非空值拒绝；`Enabled` 与 `Version` 仍必填。 |
| 凭据 | `Credential` | 带版本的快照；`BindingVersion` 必须等于所解析绑定的版本，否则调用失败（`credential_unavailable`/`consistency`，ADR-0003）。key 以 `ai.Secret` 传入，不会出现在任何输出中。 |
| 请求 | `Request`、`ImagesRequest`、`ClassifierRequest` | 分别为聊天历史/工具、图像提示/参考图、分类状态/问题/可选图片。可变输入独立复制，unary 先检查已知大小再复制。 |
| 选项 | `ResponsesOptions`、`AnthropicOptions`、`GeminiOptions`、`ChatOptions` / `SimpleOptions`；`OpenAIImagesOptions`、`GoogleImagesOptions`；`TypeSafeOptions`、`OpenAIDecisionsOptions` | 完整选项必须与绑定 API 匹配。`Nullable[T]` 区分未设置、`null` 与零值；图像选项拒绝显式 null。 |
| 结果 | `Result`、`ImagesResult`、`ClassifierResult` | 均含 `Metadata`：操作、调用归属、版本、目录哈希、尝试与用量完整性。聊天失败时仍有有效消息；unary 失败时内容/答案为空，已上报用量保留。 |
| 目录 | `Model`、`ImageModel`、`ClassifierModel` | `Catalog.Lookup` / `ModelsOf` 仅查聊天；图像用 `LookupImage` / `ImageModelsOf`，分类用 `LookupClassifier` / `ClassifierModelsOf`。`Client.Catalog()` 返回独立快照，发现能力不授予调用权限。 |

用不匹配的操作入口调用绑定，会在读取凭据和创建尝试前以 `tenant_denied/capability` 拒绝。同 Provider、同账户的多个操作绑定可以共用凭据引用，授权仍分别执行。

### 回调与流式调用

`client.WithHooks(hooks)` 返回提供全部六个入口的 `HookedClient`。
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

`ResourcePolicy.Image == nil` 禁用图像生成，`Classifier == nil` 禁用分类；
`Classifier.Images == nil` 独立禁用分类图片输入。启用的子策略容量必须全部为正，构造时深复制：
图像限制输入/输出数量、单张与总输出字节；分类限制问题数、状态字节与单个问题字节。
全局请求/输出字节、并发与时限约束三种操作。unary JSON 按 `MaxOutputBytes` 读取，不受仅用于 SSE 的 `MaxFrameBytes` 限制；实际限额同时取协议/型号硬限制。

共用一个 Client 调用聊天、图像和分类时，使用独立的 `localassembly.MixedPolicy()` 或
`hostintegration.CloudMixedPolicy()`。它们显式启用有限子策略，按整体 JSON/base64 的
实测成本限制为 4/8 并发；原聊天容量不构成图像默认值。本地 `OpenOperations` 为每个
操作配置独立 Binding，Keys 按 CredentialRef 装配。
云端 `Host.GenerateImages` / `Host.Classify` 从 Principal 校验身份并传递取消 context。
宿主处理类型化结果与错误、业务置信阈值和转人工；记录使用 Observer。混合与分类图片设计负载只在人工选择的独立压力阶段运行，race 单独选择；未开压力时用例为 NOT_RUN。

### 可观测性、用量与成本

`Config.Observer` 通过有界队列异步接收 `call_started`、`attempt_started`、`attempt_finished`、`call_finished`；慢的 Observer 不会阻塞调用，丢弃计入 `Client.ObserverStats()`。记录中不含 key、正文、工具参数或错误文本（ADR-0009）。`Usage` 保持 pi 的数字；`Usage.Cost` 是按目录价格的估算，不是账单。用量是否上报记录在 `Attempt.UsageReporting`——零值 `Usage` 从不表示免费（ADR-0010）。

图像调用增加可选 `Usage.Modalities` token 明细，按文本/图片费率估算；缺明细仍标为部分或未上报。
Google 思考 token 计入输出一次，并单独记录为推理量；TypeSafe 只对输入 token 计价。
观测包含 `Operation` 与用量，不含提示、图片、分类状态或答案。

Decisions 保留包含缓存的完整 `input_tokens` 与已上报输出，规范缓存桶为零，不重复扣费。
input/output/total/cached/cache-write/reasoning 六项均明确报告（包括零）才为 complete；
缺失或 null 字段为 partial，整个 usage 缺失/null 为 unreported。无显式 reasoning 时保持未设置。
损坏或有歧义的 usage 使响应失败并标 unreported；坏答案仍保留独立合法的用量与输入估价。
内置 Luna 按输入计价：不超过 272000 token 为 $0.10/百万，超过后全部输入为 $0.20/百万；
输出/缓存不另收费。区域 premium 与账单不由估价核对，详见[用量契约](../docs/barness-ai/contract.md#decisions-用量与估价工单-04)。

### pi-ai 1.1.0 基线的聊天行为

- 无法解析或非有限的重试等待时间回退为有界、带抖动的指数退避。
- Responses 按响应报告的服务等级估价，缺失时取选项：flex 为 0.5 倍，priority/fast 为 2 倍（gpt-5.5 为 2.5 倍）。
- Anthropic 后续增量中的 1 小时缓存写入明细覆盖此前值。
- Responses/Chat 按具名选项 → 型号通用默认 → canonical 分档默认 → 调用覆盖 → `OnPayload` 一次浅合并。
  分档使用 `off/minimal/low/medium/high/xhigh/max`；原生 Anthropic/Gemini 忽略两种采样 map。
  Full 保留原 wire effort，Simple 独立 clamp；采样/Hook 覆盖后不重新选档。内置型号未声明采样默认。
  见[完整 Host 示例](../docs/barness-ai/thinking-sampling.md)。
- 最终聊天消息与终结事件带非负整数 `DurationMs`（JSON `durationMs`），从调用建立计至结束，包含 Hook 与重试。
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
7. **自身契约与证据。** pi 只作架构和设计参考；公共行为由 barness 契约、公共 E2E 和自身真实证据验证，流程见 [ADR-0051](../docs/adr/0051-scoped-verification-receipts.md)。
8. **能力按 Provider 确定，而非按协议。** 共享 adapter 不得发送其他 Provider 的字段（如 DeepSeek × Responses 不发送 `store`/`include`/缓存字段）。请求字段从该 Provider 的能力派生。
9. **Observer 从不影响调用；回调可以使调用失败，但不能放宽**操作、鉴权、目标、模型、原生引用、托管工具或分类图片权限。unary 最终输入必须重新校验（ADR-0005）。
10. **工具参数不可执行**，直到对已结束的消息调用 `ValidateToolCall` 通过。
11. **操作独立授权。** 绑定操作零值只允许聊天，按操作查目录后才读凭据（ADR-0020）。
12. **unary 输出整体发布。** 所有图片/答案校验后一起交出；任何失败均保留已上报用量并归还许可。

### 包结构

生产代码位于单一的 `ai` 包，按职责与协议前缀组织文件。

| 领域 | 文件 |
| --- | --- |
| 共同调用运行时 | `client.go`、`call_runtime.go`、`unary*.go`（生命周期、固定快照、unary HTTP/JSON） |
| 聊天入口与流 | `entry.go`、`call.go`、`stream.go`、`event.go`、`assembler.go`、`partial*.go` |
| 领域类型 | `operation.go`、`message.go`、`request.go`、`result.go`、`binding.go`、`options.go`、`nullable.go`、`reasoning.go`、`errors.go` |
| 解析与安全 | `preflight.go`、`policy.go`、`admission.go`、`limits.go`、`timeouts.go`、`retry.go`、`transport.go` |
| 历史与原生状态 | `transcript.go`、`history.go`、`replay.go`、`native.go`、`cache_key.go` |
| 工具 | `tool.go`、`tool_validate.go`、`tool_coerce.go`、`tool_json.go`、`argument_text.go` |
| 目录 | `catalog*.go`（类型化型号、查询、校验、内置数据、快照与价格） |
| 用量与成本 | `usage.go`、`estimate.go`、各协议 `*_usage.go`；`anthropic_effort.go` 处理托管推理强度 |
| 回调与可观测性 | `hooks.go`、`hooks_run.go`、`observer.go` |
| 聊天 adapter 契约 | `adapter.go`（`adapter` 接口与按 `API` 索引的固定 `registry()`）；unary 在 `images.go` / `classifier.go` 按操作与协议显式分派 |
| OpenAI Responses | `responses*.go`、`openai_sdk.go` |
| Anthropic Messages | `anthropic*.go` |
| Gemini Developer API | `gemini*.go`（REST，不用 SDK） |
| Chat Completions | `chat*.go`（`chat_compat.go` 存放各 Provider 的差异） |
| 原生图像 | `images.go`、`image*.go`、`openai_images*.go`、`google_images*.go` |
| 分类协议 | `classifier*.go`，独立的 `typesafe*.go` 与 `openai_decisions*.go` |
| 共享 HTTP/SDK 胶水 | `sdk_middleware.go`、`http_failures.go` |

其他目录：

| 目录 | 用途 |
| --- | --- |
| `examples/` | `localassembly`、`hostintegration`（聊天/混合操作）、`toolloop`、`requestid`、`classifiermigration`、`chineseeval`（独立业务评估），均有离线 E2E |
| `e2e/` | 针对公共 API、本地受控 Provider 与宿主替身的离线验收测试；fixture 位于 `e2e/testdata/<protocol>/` |
| `live/` | 真实 API 冒烟（build tag `live`），更新 `support-matrix.json` |
| `release/` | 只读发布回执判定命令与目录快照 |
| `internal/testkit/` | 仅测试使用：受控 Provider、宿主替身、证据、审计和只读发布判定 |
| `internal/probe`、`internal/clock` | 经 `Config` 注入的测试钩子（宿主无法构造） |

### 常见改动

- **新增型号：** 按纳入标准（ADR-0018）更新对应 `catalog_*.go` 数据（分类还包括 `catalog_decisions.go`），同步审阅 `release/catalog-snapshot.json` 的目录内容与公共 `BuiltinCatalog().Hash()`；`TestBuiltinCatalogReleaseSnapshot` 核对二者。
- **在已有协议上新增 Provider：** 新增 `ProviderID` 及其能力集（如 `chat_compat.go` / `responses_capabilities.go`）、`e2e/testdata/` 下的 fixture、live 组合、ADR；真实支持由自身路线证据判定。
- **新增协议：** 明确操作、`API`、类型化选项与结果转换，添加 E2E 套件与 ADR。聊天实现 `adapter` / `registry()`；unary 复用共同运行时与显式分派。编码、读取和校验均执行限额，不提前添加注册抽象。
- **改变可观察行为：** 先写公共失败场景，再验证受影响 E2E；涉及厂商行为另列所需 live 路线与授权。

### 测试

修改测试、选择验证、判定发布或维护证据前，遵循[测试流程 SOP](../docs/agents/verification.md)。步骤、命令模板、完成条件与失败处理以 SOP 为准；先建立公共 Client/CLI 失败场景，再实现，每次独立保存可验证的原生证据。

[工具合同](../tools/verification/README.md)定义回执形状与只读发布 CLI。付费 live、race 和压力仍分别明确选择，离线/受控回放不产生真实支持资格。本次流程启用事实及未执行项见[当前状态](../.scratch/barness-ai-test-process-review/status.md)。

## 延伸阅读

| 文档 | 内容 |
| --- | --- |
| [docs/barness-ai/contract.md](../docs/barness-ai/contract.md) | 公共契约、完整错误表、装配说明 |
| [docs/barness-ai/differences.md](../docs/barness-ai/differences.md) | 历史 pi-ai 差异研究（非当前门禁） |
| [docs/barness-ai/decisions.md](../docs/barness-ai/decisions.md) | 分类输入、拒答、图片授权、用量与回滚 |
| [docs/barness-ai/thinking-sampling.md](../docs/barness-ai/thinking-sampling.md) | 思考分档、合并顺序与 full/simple 边界 |
| [docs/barness-ai/README.md](../docs/barness-ai/README.md) | 发布交付物、门禁状态、策略实测 |
| [docs/adr/](../docs/adr/) | 当前与历史 ADR |
| [GLOSSARY.md](../GLOSSARY.md) | 领域术语（租户、服务绑定、逻辑调用、原生续接状态……） |
| `go doc -all ./ai` | 包与类型文档 |
