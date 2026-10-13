# Pi 1.1.0 OpenAI Decisions 与 GPT-6 Luna：实现调研及 barness-ai 接入方案

研究日期：2026-10-12（Asia/Shanghai）。固定审查 Pi `v1.1.0` 源码与官方发布包，核对 OpenAI 第一方 API 文档；未修改生产代码、测试、冻结 oracle 或模型目录，未调用真实付费 API。

## 1. 结论

Pi 1.1.0 新增的是 **同一 GPT-6 Luna 的 classifier 模型类型和专用 API 路由**，不是一个带分类后缀的新模型 ID。分类目录项为 `type=classifier / provider=openai / api=openai-decisions / id=gpt-6-luna`；聊天目录仍是 `type=chat / api=openai-responses / id=gpt-6-luna`。分类请求的 `model` 原样使用 `model.id`，没有剥离后缀、另设 `modelId` 或选择 reasoning variant 的逻辑。[生成器][catalog-source]、[路由测试][route-tests]、[请求构造][decisions-call]

应该在 barness-ai **现有独立 `Classify` unary 路径上接入新的协议 adapter**，复用请求隔离、授权、资源限制、回调和显式重试运行时，并完整交付文本和图片两个切片。仅向聊天目录加入 GPT-6 Luna，或把 Decisions 包装成 Chat Completions 的 JSON 输出，均不能提供相同能力。此为设计建议；依据是 Pi 将 classifier 与 chat 分开建模，并为同 provider 注册独立 classifier API。[类型边界][classifier-types]、[Provider 注册][provider]、[Models 分派][dispatch]

Pi 的协议映射可直接作为差分基准，但不应无条件复制其结果处理：Pi 遇到单题 refusal 会清空全部答案，score 分布会丢弃，响应概率只检查有限数字，没有完整的名称、候选集合和取值范围校验。barness-ai 应保留既有严格边界；需要保留 refusal、score 分布等原生信息时，逐项登记差分扩展。[答案解析][answer-parser]、[终态处理][decisions-call]；本报告第 10 节已通过发布包离线复现。

## 2. 研究基线与版本差异

| 项目 | 已核对的第一方证据 |
| --- | --- |
| Pi 1.1.0 | tag `v1.1.0` 指向 `abe508e1b89912adde45528136c3221eb69acdd7`；[tag API][tag110] |
| 引入提交 | `ce8972a0e99cab79f2620a594a1c17baef436c67`，2026-10-07，新增 Decisions classifier 和 classifier images，同时将共有 HTTP helper 从 `system-one-shared.ts` 移至 `classifier-shared.ts`；[固定提交][feature-commit] |
| 官方发布说明 | 1.1.0 明列 GPT-6 Luna Decisions classifier、`OPENAI_API_KEY` 与 `ClassifierContext.images`；[AI changelog][changelog]、[GitHub release][release] |
| 最近 1.0.x | `v1.0.4` → `7c10bd4337495ee613f2224843ecdf349b80d1df`；[tag API][tag104] |
| 1.0.4 已有能力 | 官方 npm 发布包 OpenAI 数据仅有 `openai-responses` 分组，包含 `chat:gpt-6-luna`，没有 `openai-decisions` 或 `classifier:gpt-6-luna`；[官方 npm 元数据][npm104]、[官方 tarball][npm104-tar] |
| 1.1.0 发布包 | `@earendil-works/pi-ai@1.1.0`，OpenAI SDK 固定 `7.19.0`；npm 未提供 `gitHead`，不能把 tag commit 冒称 npm 自带的来源字段；[npm 元数据][npm110]、[package.json][package] |
| 发布包完整性 | npm `dist.integrity` = `sha512-1T7LAkc/5Bvc0v6w4vAGVdCrli0o/E0pEmYKTnixu95vSFArBjvbhS/G4ZwI0RUePgf0Imcu0VyqlM4EcXxqfw==`；[npm 元数据][npm110] |
| 本地基线 | barness 冻结 oracle 已是 Pi 1.1.0，记录了固定 commit、integrity、生成目录哈希及两次源码构建一致性；因此接入方案不需要再次升级冻结版本；[PROVENANCE](/Users/cyber/RestoX/barness/ai/internal/testkit/pioracle/node/PROVENANCE.md) |

源码以官方 tag archive 的 `/tmp/pi-1.1.0` 读取；生成模型数据同时读取仓库冻结的官方 npm `dist/providers/data/openai.json`，避免运行在线模型生成器后以最新目录替代 1.1.0 发布目录。本报告源码链接全部固定到上述 commit。研究笔记路径沿用仓库 `.scratch/<feature>/research.md` 约定。[仓库约定](/Users/cyber/RestoX/barness/docs/agents/issue-tracker.md)、[既有研究](/Users/cyber/RestoX/barness/.scratch/pi-1.1.0-thinking-sampling/research.md)

## 3. 准确的模型目录与能力

| 字段 | Pi 1.1.0 classifier 快照 |
| --- | --- |
| type / provider / api / id | `classifier` / `openai` / `openai-decisions` / `gpt-6-luna` |
| 生成 JSON 内部 key | `openai-decisions` 分组下的 `classifier:gpt-6-luna` |
| SDK 查找 | `models.getModelOfType('classifier', 'openai', 'gpt-6-luna')` |
| endpoint | base URL `https://api.openai.com/v1`，adapter 追加 `decisions` |
| 输入 | `['text', 'image']`；`contextWindow=922000` |
| 基础费率（USD / 1M tokens） | input `0.1`；output、cacheRead、cacheWrite 均为 `0` |
| 长上下文阶梯 | 总输入 **严格大于** `272000` 时，全请求 input 使用 `0.2`；其余费率仍为 `0` |
| classifier 特有字段 | 无 reasoning、thinking level、maxTokens、sampling defaults；这些属于 chat 模型 |

目录与内部 key 由生成器和 type-prefixed catalog 共同确认；公开查找使用无前缀 ID。不能将内部 `classifier:` key 发给 OpenAI，也不能将 chat 模型的缓存及输出费率沿用到 Decisions。[模型数据生成][catalog-source]、[catalog flatten][catalog-flatten]、[长上下文费率][pricing-helper]、[费率选择][cost-calculate]

**没有发现 Pi 对 OpenAI score 声明“2–10 档”的 metadata 或校验。** Pi 公共 score 是 `criteria: string[]`；codemode 只要求非空字符串数组。choice 也是非空的字符串说明 map，不在本地限制到 OpenAI 的候选数边界。不得把 llama.cpp adapter 的十档限制或 barness 当前 TypeSafe 的全局范围说成 Decisions API 的能力。[classifier 类型][classifier-types]、[codemode 校验][codemode-validation]、[llama.cpp 特有限制][classifier-readme]

## 4. 请求模型、图片与问题映射

调用链为 `Models.classify → 检查 classifier type / input modality → 解析 provider auth → provider.classify → classifiers[model.api] → lazy import → openai-decisions.classify → postClassifierRequest`。分类是单次完整 JSON 结果，不创建 stream event。裸 adapter 没有 OpenAI SDK import，使用共有 helper 的 `fetch`；因此 Pi 的 OpenAI SDK `7.19.0` 不阻碍这个新 endpoint。[Models][dispatch]、[类型检查][model-guards]、[lazy API][lazy]、[HTTP helper][classifier-http]、[依赖版本][package]

### 4.1 输入

Pi 将 JSON object 的 `state` 做 `JSON.stringify`。无图时 `input` 是该字符串；有图时转换为一个 `role=user` 消息，先放 `input_text`，再按原顺序放各个 `input_image`，`image_url` 为 `data:<mimeType>;base64,<data>`。没有 URL、file ID、image detail 或多个 user message 的公开 classifier 表达。[wireInput][input-builder]、[Context 类型][classifier-types]

模型必须声明 `input` 包含 image；`Models.classify` 在授权与发送前拒绝其他模型的图片输入。Decisions adapter 额外在发请求前拒绝 129 张及以上图片；上限固定为 128。codemode 检查 image block 的 type/data/mimeType 字段形状，但未校验 base64 是否可解码、MIME 白名单、单图体积，也未在此 adapter 自动 resize。[输入能力检查][model-guards]、[图片构造][input-builder]、[codemode 形状检查][codemode-validation]、[图片测试][image-tests]

Pi codemode 不直接信任脚本提交的 model URL/headers，而按 type、provider、id 重新查找目录项；最多允许四个嵌套 model call 同时运行。它把 read 工具返回的 image block 作为 `images` 接受，release 的“图片分类”是完整落到 classifier API 的能力。[宿主边界][codemode-run]、[并发上限][codemode-concurrency]、[官方示例][codemode-doc]

### 4.2 问题

| Pi 问题 | Decisions wire | 回到 Pi |
| --- | --- | --- |
| `choice` + criteria map | `{type:'choice', name:id, instructions, choices:[{value:<字符串key>,description:<说明>}]}`；空说明不发 description | `{type:'choice', choice:<字符串>, probabilities:<按value组成map>, confidence}` |
| `score` + 从低到高的 criteria | `{type:'score', name:id, instructions, levels:[{label:<说明>}]}` | `{type:'score', score, confidence}`；原生 probabilities/value/label 丢弃 |
| `bool` + true/false criteria | **始终** `{type:'predicate', name:id, instructions}`；非空 true/false 说明追加为两行 `True means: ...` / `False means: ...` | `{type:'bool', probability}`，表示 true 的概率 |

以上为 Pi 的具体映射，不是所有原生 API 类型的总集。Pi choice 只能表达字符串键，没有 typed boolean / number / null value；bool 也没有改用两个 boolean choice。[请求映射][question-builder]、[响应映射][answer-parser]、[映射测试][mapping-tests]

例如 state、choice、score、bool 混合请求的结构如下（说明文字为研究示例）：

```json
{
  "model": "gpt-6-luna",
  "input": "{\"message\":\"部署成功\"}",
  "questions": [
    {"type":"choice","name":"category","instructions":"分类","choices":[{"value":"success","description":"成功"},{"value":"failure","description":"失败"}]},
    {"type":"score","name":"urgency","instructions":"紧急程度","levels":[{"label":"低"},{"label":"中"},{"label":"高"}]},
    {"type":"predicate","name":"approved","instructions":"用户同意吗？\n\nTrue means: 同意\nFalse means: 不同意"}
  ]
}
```

`ClassifierOptions.temperature` 不进入这个请求，也不用于本地重算概率；`onPayload` 可以修改或替换整体 body。Pi 在发送 hook 后没有再次按原生 schema 校验，解析响应时仍参照原始 `context.questions`，因此 hook 改变问题集合可能导致请求与返回模型不一致。此为代码推论；barness 必须针对最终 wire 请求做校验和冻结，不能套用 TypeSafe 的 map payload guard。[共有 hook][classifier-http]、[解析调用][decisions-call]、[官方 Pi 选项说明][classifier-readme]

## 5. 结果边界与 token / 成本

Pi 先解析并保存 usage，后解析答案；只要 HTTP body 合法，缺题、答案类型错误、refusal 等错误仍保留已计费的 usage。每个请求 question 必须有同名 answer，返回数组顺序可以不同。一个 refusal 会抛解析错误，最终 `answers={}`、`stopReason=error`，先前正常答案也被丢弃；取消信号已 aborted 时才标 `aborted`。[答案完整性][answer-parser]、[usage 和终态][decisions-call]、[refusal 测试][refusal-tests]

Pi 的校验局限如下，均来自固定源码，并由第 10 节 probe 复核：[答案解析][answer-parser]、[数字检查][number-check]

- 所有概率、confidence、score 只要求 finite number；不保证概率/置信度在 `[0,1]`，不检查概率和，也不检查 score 落在 levels 的范围内。
- 不要求返回 choice 在请求候选集合内，probability value 集合也不要求与候选集合一致。
- 相同 `name` 的重复 answer 由 Map 最后一个覆盖；未知 name 忽略。choice probability 数组中重复 value 也最后一个覆盖。
- malformed/no-name 条目可以被跳过；只要所有请求题仍找到有效 answer，整体可成功。缺题或题目类型不符则整体失败。
- score 原生 distribution 与 labels 不保存；没有 refusal answer 领域类型，也没有 response ID 或 native envelope。

**Pi 没有从 `input_tokens` 中扣除缓存统计。** 共有 `parseClassifierUsage` 取正的 finite `input_tokens/output_tokens`，保留到 `Usage.input/output`；`cacheRead=cacheWrite=0`，`totalTokens=input+output`，忽略原生 `input_tokens_details`、`output_tokens_details.reasoning_tokens` 和 `total_tokens`。缺失或无法识别 usage 时不报解析错误。barness 如按 presence 保留 reasoning/cache 细节，应以明确扩展记录，不把省略字段伪造为厂商报告的零。[usage parser][usage-parser]

由此，Pi 发布目录按完整 `input_tokens` 计费，并以它判断 272K 阶梯；300000 输入成本为 `$0.06`，即使 synthetic response 附带 cached/write token counts，结果仍同样为 `$0.06`。这是 **Pi 已证实的归一化行为**，不是此调研声称验证了 OpenAI 线上账单。[usage parser][usage-parser]、[费率计算][cost-calculate]、[上游长上下文测试][mapping-tests]

OpenAI 当前 guide 将 Decisions 定价描述为 input-only、无独立 cache-read/cache-write/output 费用，并提醒长上下文和区域价格调整仍适用。barness 至少应使用 Decisions 独立费率，保留原生 usage 供核账；在无真实非零缓存统计及账单证据时，不自行引入“扣缓存后按 input 计费”的修正，也不把总 input 与 cache 分桶重复相加跨过阶梯。[OpenAI Decisions guide][official-guide]。线上账单一致性列为发布前待验证项。

## 6. 认证、重试、超时与错误

`openai` provider 的 classifier 与 chat 共用 provider identity，但 Decisions 要 API key。Pi 的 availability filter 在 credential 为 OAuth 时隐藏全部 classifier；聊天模型仍可用。官方 Pi 文档特别提示，即使设置了 `OPENAI_API_KEY`，provider 当前持有 ChatGPT 登录凭据时仍隐藏 classifier；需要退出该登录再使用 key。这是可用目录过滤，不能理解为裸 adapter 自己做了凭据类型鉴别。[provider][provider]、[可用目录测试][route-tests]、[Pi 使用说明][models-doc]

共有 HTTP helper 要求非空 `options.apiKey`，以 bearer auth、JSON body 发 POST；头的覆盖順序为默认 → model → options。`onPayload` 在重试循环前执行一次，`onResponse` 只在成功 HTTP/JSON 解析后执行。每次尝试有新 timeout signal，并与调用取消信号合并；没有设置 timeoutMs 就没有该 helper 的额外定时器。[HTTP helper][classifier-http]

默认 `maxRetries=2`；共有 retry 尊重 `x-should-retry`，随后按 408/409/429/5xx 判断，读取 retry-after，并使用带抖动的指数 backoff。retry sleep 可被取消，默认最大服务器等待 60 秒。裸 fetch 网络错误通常不是此 helper 要求的 ProviderError 形状，不能笼统声称“所有网络错误均重试”；timeout 会转换成可重试 ProviderError。[HTTP helper][classifier-http]、[共有 retry][retry]

Decisions adapter 显式将 504 加入 no-retry 列表，覆盖默认 5xx 行为，并把 gateway HTML 错误替换成可读描述。约五秒、约 600K 输入就可能触发的 gateway 限制是 **Pi 源码注释和第一方使用文档记录的运行观察**；本次没有真实请求复现，不能把它作为永久 API 服务承诺或另设硬 600K 模型容量。[特殊 504 处理][gateway]、[官方 Pi README][classifier-readme]、[504 与 503 测试][error-tests]

922000 则是 Pi classifier 发布目录的输入窗口值；Pi adapter 本身不先统计 tokens 并硬拦截。barness 的字节预算、图片数量、超时和显式重试上限需要按本地运行时明确实施；额外保守限制须写清自己的依据，不能冒称 Pi 或 OpenAI 的协议限制。[目录快照][catalog-source]、[请求构造][decisions-call]

## 7. 应复用与应单独决策的实现

| 边界 | 接入建议 | 理由/证据 |
| --- | --- | --- |
| 模型身份 | 新增 classifier API，同 provider + 同 id 按 type 分离 | Pi 1.1.0 目录即如此；[路由测试][route-tests] |
| 传输 | 在现有 Classify unary 中新增 adapter，复用共享 runtime；不必为 endpoint 先升级全项目 SDK | Pi 用 raw fetch；[HTTP helper][classifier-http] |
| 公共问题 | 保持现有 choice/score/bool，bool 先按 Pi predicate 转换 | 可做确定性的固定版本差分；[问题构造][question-builder] |
| 任意 JSON 说明 | 原生 instructions/description/label 必须是 string；结构化内容应有明确转换/拒绝策略 | Pi 只暴露 string，不能自动丢掉 barness 更宽 JSON 的信息；[类型定义][classifier-types] |
| 图片 | inline image 字段、输入能力声明、最多 128、逐图与总字节校验、hook 后再校验；完整支持需要实际交付 | Pi 1.1.0 已含图；[图片测试][image-tests] |
| response | 检查最终问题名称集合、唯一性、type、候选集合、finite/range；协议错误保 usage | Pi 校验存在空缺；[答案解析][answer-parser] |
| score / refusal | 推荐保留 score 分布与逐题 refusal，以明确扩展账本记录相对 Pi 的差异 | Pi 将这些信息丢弃；[答案解析][answer-parser] |
| usage | 总输入计费与 cache 细节分离，优先固定 Pi 归一化基准，账单再核实 | Pi cache 桶为零，不扣 input；[usage parser][usage-parser] |
| 重试 | 保持 barness 显式重试和故障归类；默认重试、504 等差异逐项登记 | 避免复制一套协议专属 retry runtime；[Pi retry][retry] |
| 容量 | question shape、API 限制、宿主资源政策分开；score 2–10 如暂保留必须注明为 barness 保守政策 | Pi OpenAI 侧没有这一上限；[codemode 校验][codemode-validation] |

以上是源码事实推导的建议。以下补充 OpenAI 官方边界和本地工程方案，均为待实施设计。

## 8. OpenAI 官方边界与 SDK 选择

截至研究日，Decisions 是 public beta，只支持 `gpt-6-luna`。官方提供 predicate、choice、score；choice 的值可以是 string 或 boolean，类别是无序集合，score 则是有序等级的加权期望。官方称其约比 Responses 快 10 倍，此报告没有测得 barness 的实际延迟，不能直接写入产品 SLA。[官方 guide][official-guide]

原生 API 比 Pi 的公开分类接口更宽：允许纯文本或多个 user message；支持文本、图片，最多 128 张；图片可以使用 base64 data URL 或公开 HTTP(S) URL，`detail` 可选。它不接受其他角色、文件 ID、音频、工具调用或 item reference。choice 要求 2–255 个唯一候选，string `"true"` 和 boolean `true` 是不同值；score 的等级数量上限未在本次读取的 reference 中声明。[HTTP reference][official-reference]

response 的 answers 数组按问题顺序返回；单题可为 refusal，其他题仍能正常回答。score 原生返回含整数 value、label、probability 的分布；usage 含 input/output/total、缓存统计和 reasoning 统计。这些信息需要在 barness 的协议边界独立解析，不能借用 TypeSafe wire decoder。[Go reference][official-go-reference]

当前 `go.mod` 固定 `openai-go/v3 v3.66.0`，本地该版本 Client 没有 Decisions service。官方示例最低 Go SDK 为 `3.73.0`、JS SDK 为 `7.30.0`；本次尝试读取 Go `3.73.0` module 时 `proxy.golang.org` 网络超时，未完成新 SDK 源码审查。[现有依赖](/Users/cyber/RestoX/barness/go.mod)、[官方 SDK 要求][official-guide]

**推荐直接使用已有 `sendUnary` 发送 native HTTP。** 它已经承接 HTTP client、绑定 endpoint、重试、响应限额与许可生命周期；新 adapter 只拥有请求/响应转换，因此无需新增依赖。使用 SDK 是可行备选，但应先升级到官方要求版本，核查类型、禁用 SDK 重试、保留原始 body 的 presence/重复字段检查，并对现有 OpenAI 聊天和图像路线做回归；不能只依赖 SDK 的零值结果判断协议是否完整。[现有 unary](/Users/cyber/RestoX/barness/ai/unary.go:13)、[SDK 请求配置](/Users/cyber/RestoX/barness/ai/openai_sdk.go:35)

Pi 的 `922000` classifier context 是本次对齐的发布目录快照；OpenAI 通用 Luna 模型页的 `1050000` context / `128000` output 属于通用型号描述，不能直接覆盖该 classifier 目录。通用页的输出、缓存费率也不用于 Decisions。当前官方只列出无日期 ID `gpt-6-luna`；不能编造一个日期快照。目录版本能冻结元数据和价格，不能保证此别名背后的权重永远不变。[Pi 目录][catalog-source]、[官方模型页][official-model]

## 9. barness-ai 推荐工程方案

本地审查基线为 `8ea61b8d70de6c84fe27be34dc7d21c5c40841d9`；内置目录版本 `2026-10-11.1`。已有模型身份 `(Operation, Provider, API, ID)` 可以直接容纳同名 Luna 的两种操作；当前 classifier 路由只允许 TypeSafe，执行也固定调用 TypeSafe，目录只有 `jev-1.13.0`。这三个接缝都必须修改，单纯加目录项不会使新模型可调用。[ADR-0020](/Users/cyber/RestoX/barness/docs/adr/0020-barness-ai-model-operations-and-binding-authorization.md)、[授权路由](/Users/cyber/RestoX/barness/ai/call_runtime.go:71)、[执行分派](/Users/cyber/RestoX/barness/ai/classifier.go:106)、[目录](/Users/cyber/RestoX/barness/ai/catalog.go:49)

### 9.1 职责与数据流

沿用 `Client.Classify` / `HookedClient.Classify`、`ClassifierRequest`、三种问题与 `ClassifierResult`。新增 `APIOpenAIDecisions = "openai-decisions"`，Provider 继续用 `ProviderOpenAI`；初期 `OpenAIDecisionsOptions` 可以是空结构，nil 仍选协议默认。它不需要 temperature、reasoning、stream 或 maxTokens。新绑定必须明确 `OperationClassifier`，白名单解释在该操作和协议之下；chat/image 绑定不能授权它。[现有入口](/Users/cyber/RestoX/barness/ai/classifier.go:11)、[Options 边界](/Users/cyber/RestoX/barness/ai/options.go:10)、[绑定原则](/Users/cyber/RestoX/barness/docs/adr/0020-barness-ai-model-operations-and-binding-authorization.md)

```mermaid
flowchart LR
  H[Host: 可信 CallScope + Target] --> P[Classify: 有界输入快照]
  P --> A[按操作 / Provider / API / ID 授权]
  A --> C[固定 Binding / Credential / Catalog 快照]
  C --> W[Decisions adapter: 构造 native payload]
  W --> G[Hooks 后重新校验并冻结最终请求]
  G --> U[sendUnary: 准入 / 超时 / 显式重试]
  U --> R[先记录 usage，再校验全部答案]
  R --> D[类型化结果 + 独立元数据]
```

建议保持两条 classifier 协议的固定 switch 分派，共享已有 runtime；请求/响应 wire 类型留在各自协议文件中。新增文件可按 `openai_decisions.go`、`openai_decisions_request.go`、`openai_decisions_guard.go`、`openai_decisions_response.go`、`openai_decisions_usage.go`、`catalog_openai_decisions.go` 分工。提取真正共用的校验函数即可，无需提前设计注册插件、第二个 HTTP client 或一套新的 retry runtime。[已有共享边界](/Users/cyber/RestoX/barness/ai/unary.go)、[ADR-0021](/Users/cyber/RestoX/barness/docs/adr/0021-barness-ai-typesafe-unary-classification.md)

### 9.2 输入与模型能力

| 决策 | 推荐的第一版契约 |
| --- | --- |
| State | JSON string 解码为原文；object/array 验证后压缩成 JSON 文本，保留数字精度，形成 input。这是 barness 已有宽输入的显式转换；Pi 共同子集保持 object state 差分 |
| 问题说明 | OpenAI 路线先接受 JSON string；instructions / criteria 的 object/array 在 capability 阶段明确拒绝。TypeSafe 保持现有结构化说明；若后续需要文本化，应另定格式和差分，避免隐式改变含义 |
| choice | 按确定顺序将字符串 key 转 value，null / 空说明按已确定的转换规则省略 description；新增 MinChoices 或等价协议检查，让 OpenAI 拒绝 1 个候选、TypeSafe 保留其能力 |
| score | ordered criteria 映射为 levels.label。将现有通用校验中的 TypeSafe `2–10` 限制移至该路线/模型；OpenAI 首批可声明经过自己验收的有限子集，但必须写清来源。官方上限未确认前不杜撰 server max |
| bool | 保持 Pi 的 predicate；可选 true/false 文本按 Pi 格式追加到 instructions。是否超过阈值由 Host 决定 |
| 图片 | 在 `ClassifierRequest` 增加有序 `Images []Image`，复用 inline `Image`；模型增加独立的输入模态与 MaxInputImages 能力，TypeSafe 只声明 text。不得把图片能力借给同 ID 的 chat 或另一 binding |

这些是对现有 TypeSafe 专属公共假设的调整；同步修订公共说明、JSON decoder、克隆、目录校验与哈希。现有 `ClassifierCapabilities` 只有 MaxChoices，没有 MinChoices 或图片输入能力；`ClassifierPolicy` 的通用 question validator 硬编码 choice 1–255、score 2–10，不能直接解释为 OpenAI 能力。[现有类型](/Users/cyber/RestoX/barness/ai/catalog_operations.go:59)、[现有问题校验](/Users/cyber/RestoX/barness/ai/classifier_policy.go:155)、[JSON 边界](/Users/cyber/RestoX/barness/ai/classifier_json.go:65)

图片只接受内联 base64；逐图校验 MIME、base64、`MaxImageBytes`，约束总图片数和完整 JSON 的 `MaxRequestBytes`，在复制/编码前先检查已知大小。宿主已有图像解码能力可以复用，但 OpenAI Images 编辑的 16 张 / 20 MiB 常量不属于 Decisions，不得照搬。首次声明的 PNG/JPEG/WebP 等格式应以此路线自己的证据确认。[Image 领域类型](/Users/cyber/RestoX/barness/ai/message.go:50)、[现有图片校验](/Users/cyber/RestoX/barness/ai/image_policy.go:41)、[编辑专属限制](/Users/cyber/RestoX/barness/ai/openai_images_input.go:12)

### 9.3 最终 payload、答案与消费量

`OnPayload` 必须保持 native `input / questions[]` 形状。现有 `ClassifierPolicy.checkPayload` 读取 `state / questions map`，不能用于 Decisions。新 guard 应校验 model 固定、禁止认证/目标/操作/工具字段，限制数组和图片大小，重新解析最终 native 问题，冻结名称、kind、候选值与 levels；响应只依据这份快照验证，Host 的入口 map 和回调持有的可变数据都不能继续影响请求。[现有 guard](/Users/cyber/RestoX/barness/ai/classifier_payload.go:12)、[现有冻结模式](/Users/cyber/RestoX/barness/ai/typesafe.go:36)、[可信回调](/Users/cyber/RestoX/barness/ai/hooks.go:8)

继续整体发布答案：重复 JSON 字段、重复 answer name / probability value、缺题/多题、错误 kind、未提供的 choice、非有限数或越界概率、非法分布均应失败，Answers 清空但保留已明确上报的 usage。score 校验其分布与有序等级的一致性、分值范围和加权期望；TypeSafe 的两位小数特例只留在 TypeSafe，OpenAI 的精度容差依据自己的真实响应确定。[现有响应顺序](/Users/cyber/RestoX/barness/ai/typesafe_response.go:10)、[答案校验](/Users/cyber/RestoX/barness/ai/typesafe_answers.go:25)、[TypeSafe 特例](/Users/cyber/RestoX/barness/ai/typesafe_score_rounding.go)

推荐新增封闭 `RefusalAnswer`，用 `type=refusal` 表示正常的单题拒答。合法 response 含拒答时返回 nil error，其他有效答案仍可用；即使全部拒答，也是一份已完成的分类结果，交由 Host 选择后续流程。缺题/损坏响应仍整体失败。保留 score 的原生 distribution，并在验证 label 与最终 rubric 一致后映射到已有 `ScoreAnswer.Probabilities/Legend`；不要编造拒答原因。以上与 Pi 的两项差异需要独立 ledger 和 E2E，而不能豁免整个结果。[现有封闭答案](/Users/cyber/RestoX/barness/ai/classifier_answers.go)、[原生 refusal / score][official-reference]

消费量建议独立解析：canonical Input 保留 Decisions 上报 input，cache 桶为零，先与 Pi 基准一致；保留有效原生 TotalTokens 和 Reasoning 的 presence，分别作为已有 barness 语义的扩展。若要保存 cache 明细，应另设明确的诊断记录，避免改变费用及阶梯分桶，也不能把原始 response 或题目写入 Observer。缺失、null、显式零要区分，token 计数必须为非负整数并检查溢出。不要直接复用 `responsesUsage.of` 的“input 减缓存”算法，也不要依赖 `typeSafeUsage` 的字段集。[现有 Usage](/Users/cyber/RestoX/barness/ai/usage.go:5)、[Responses 专属拆桶](/Users/cyber/RestoX/barness/ai/responses_usage.go:25)、[TypeSafe 用量](/Users/cyber/RestoX/barness/ai/typesafe_response.go:62)

基础价格为 input `.1`，严格 `>272000` 时 `.2`，其他 canonical 桶价格零；价格和能力进入独立版本的目录哈希。这里是全球标准处理的估算快照，区域额外费用与缓存账单解释须在发布证据中核对。ResponseModel 独立保存，不替换已授权 ModelID；厂商 `x-request-id` 进入 Attempt。默认不重试，允许重试时复用最终 body 和凭据，收到 2xx 后不因拒答、解析失败或读取失败再重放。[费率快照](/Users/cyber/RestoX/barness/ai/catalog_snapshot.go:61)、[现有请求生命周期](/Users/cyber/RestoX/barness/ai/unary.go:13)、[ADR-0006](/Users/cyber/RestoX/barness/docs/adr/0006-barness-ai-retry-policy-on-binding.md)

原生 `safety_identifier` 如确有业务需求，由可信 Host 派生不含个人信息的标识，按租户和账户隔离；它不是认证主体，不要用原始 ActorID 自报权限。首版空 Options 可以省略该字段，后续添加时核对官方长度与 Hook 后 guard。[官方字段语义][official-reference]、[可信 CallScope](/Users/cyber/RestoX/barness/ai/binding.go:13)

### 9.4 分期交付及验收

| 切片 | 交付物 | 完成判定 |
| --- | --- | --- |
| A：完整文本协议 | API/Options、授权与固定分派、请求映射、严格答案/拒答/用量、自己路线的 fixture | 先列失败目录并写公共 Classify E2E，再实现；Pi 1.1.0 共同子集差分零 pending，显式扩展各有证据 |
| B：图片分类 | 公共 Images、模型模态/数量、policy 与 clone、native input、Hook 后图片 guard | 本路线 E2E 验证顺序、有效图片、129 张、坏 base64/MIME、字节限额和并发；A 只完成时不得声明完整图片支持 |
| C：正式目录和发布 | 官方来源 fixture、自己的有预算 live、目录/价格新版本、支持矩阵与 release 合同、独立中文效果报告 | 对此组合取得完整证据后才纳入内置目录；无法执行 live 则保持 NOT_RUN，不沿用 chat 或 TypeSafe 的 PASS |

本地已有冻结 Pi 1.1.0，可在 runner 的 APIS 表增加 `openai-decisions` 与 `OPENAI_CLASSIFIER_MODELS` 目录，保持现有版本/commit/model-data hash 校验和 loopback 隔离。新增独立 `openai-decisions-classifier` live combo；现有矩阵 identity、TypeSafe report validator 和九组合发布合同需要增加这条路线。目录测试目前硬断言 classifier 数量为 1，必须改为按完整身份和来源验证，并重建版本/哈希快照。[oracle 路由](/Users/cyber/RestoX/barness/ai/internal/testkit/pioracle/node/runner.mjs:25)、[目录断言](/Users/cyber/RestoX/barness/ai/e2e/classifier_builtin_test.go:32)、[矩阵身份](/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/identity.go)、[发布追溯](/Users/cyber/RestoX/barness/ai/release/traceability.json:58)

当前中文效果示例从装配、plan 校验到 report 校验均固定 TypeSafe / Jev。为 Luna 使用相同数据集但独立 route/config/report，比较准确率、混淆、拒答率、置信校准、延迟和估算消耗；协议支持与任务质量分别呈现。样本不够时不将小样本结果推广为生产质量保证，阈值继续由 Host 依据误判代价决定。[效果计划](/Users/cyber/RestoX/barness/ai/examples/chineseeval/plan.go:97)、[效果报告](/Users/cyber/RestoX/barness/ai/examples/chineseeval/report.go:160)、[ADR-0024](/Users/cyber/RestoX/barness/docs/adr/0024-barness-ai-chinese-classification-evaluation.md)

**先写的失败目录**：错操作/协议/型号、跨租户、错 Options、不可用 key、单候选、非法说明与图片、回调改 model/认证/目标、回调换题后仍按旧题解析、重复/遗漏/多余答案、概率越界/不归一/score 不一致、混合 refusal、usage 缺失/零/负数/溢出、429/503/504、读 body 超时/取消/字节超限、准入释放、并发冻结和 Observer 脱敏。生产实现之前先在 E2E 表达这些场景；不在实现后补镜像式 unit test。[项目测试原则](/Users/cyber/RestoX/barness/AGENTS.md)、[已有分类 E2E](/Users/cyber/RestoX/barness/ai/e2e/classifier_validation_test.go)

每次 E2E 写出 fixture、捕获请求/响应、消费结果、错误分类、Observer 元数据、断言、manifest 哈希和重放命令。真实图片证据保留独立测试资产；504/大输入仅做有界预算探针，不能无界扩张。架构决策在实施时修订 ADR-0020/0021/0018 并增加 Decisions ADR，特别说明拒答、分布、容量来源、价格和回滚。回滚通过禁用此 classifier binding、移除此目录授权或关闭分类策略停止新调用；已发请求的费用和历史证据仍独立核对，不自动改走聊天或 TypeSafe。[发布约定](/Users/cyber/RestoX/barness/docs/barness-ai/README.md)、[隔离边界](/Users/cyber/RestoX/barness/docs/adr/0021-barness-ai-typesafe-unary-classification.md)

### 9.5 尚未确认的事项

- API 实际 score 等级范围、最大问题数、图片格式及总请求容量；Pi 代码和官方 reference 没有给出全部边界。
- 真实缓存非零统计与账单之间的关系，区域 premium，以及 reply 的 total / reasoning presence。
- Pi 所述约 5 秒 / 大输入 504 的当前区域、账户适用性；922K 目录声明不等于该输入量已有可靠服务证据。
- Luna 在目标中文任务上的质量、置信校准与真实 p50/p95 延迟。以上须由该路线自己的验收回答，不构成此次文档调研已通过的结论。

### 9.6 与现有 Classify / System One / Jev 的兼容审查

2026-10-12 补充：System One 是 TypeSafe 的协议，Jev 是走该协议的模型。接入目标是两种 native 协议都转换为 Classify 的领域契约；不能要求其 HTTP payload 或全部合法输入相同。下表区分现状、协议事实与建议，不表示生产实现已修改。

| 兼容面 | 当前 System One / Jev 路径 | Decisions 的差异 | 推荐处理 |
| --- | --- | --- | --- |
| 图片 | 公共请求只有 State / Questions，目录无分类输入模态 | 有图时 input 是 user message / parts | 扩公共图片输入及能力；没有图片的 Jev 请求保持现有 wire；Jev 收到图片明确拒绝，不静默丢弃 |
| State | JSON string/object/array 原生发送到 state | input 是 text 或 message parts | 明确 string 解码及 object/array 文本化规则；区别 JSON 原值与被引号包裹的文本，保留数值精度 |
| 说明格式 | instructions 和 criteria 允许 JSON object/array，部分允许 null | instructions / description / label 是 string | 原来的合法 Jev 请求不一定能直接改绑 Luna；先按协议验证。若采用文本化，需要公开约定与测试，不全局收窄 Jev 的格式 |
| choice | 当前允许 1–255 个，候选是 string key | 至少 2 个，原生 value 支持 string/bool | 在 Decisions 的协议检查中限制最小值，公共字符串候选先保持；两个原生值 true / "true" 不能混作一个键 |
| score | 通用 validator 当前硬编码 2–10；criteria 可为原生 JSON，legend 回显它 | 有序 levels.label/description，分布是数组；本次未确认其等级数量上限 | 将 TypeSafe 容量从通用 shape 校验移出；转换原生分布及 label。已有 TypeSafe 舍入兼容规则仅适用其自己的响应 |
| bool | native type / 字段为 noul；两侧 criteria 是原生字段 | predicate / probability；Pi 将两侧说明拼入 instructions | 公共 BoolQuestion / BoolAnswer 可保持；adapter 自己转换，结构化两侧说明也受上述格式约束 |
| 拒答与成功语义 | 封闭答案只有 choice/score/bool；任何协议错误整体清空 Answers | 单题 refusal 是合法响应，其他题可有正常答案 | 如采纳 RefusalAnswer，宿主必须处理新 variant；nil error 表示协议调用完成，不再意味着每题均有普通答案。Jev 的旧结果规则保持 |
| Native Hooks | state + questions map；回调后按最终 TypeSafe 问题校验 | input + questions array / names | 原有 payload callback 不能直接跨协议复用；按 Payload.API 分支。各协议独立 guard / freeze，共享回调调用顺序与授权保护 |
| Usage / 价格 | input/output，total 求和；input-only 自己的费率；complete 判断依据这条 wire | 另有 total、cache/reasoning details，Luna 独立长上下文阶梯 | 每协议独立解析 presence 和计费，复用公共 Usage；缺失明细不伪造零，不套用 chat 的缓存拆桶或费率 |
| Options / Binding | TypeSafeOptions 属于 typesafe-system-one，路由仅允许 TypeSafe | 新 API 与独立 classifier binding | 仅改 Target 而保留 TypeSafeOptions 会被正确拒绝；nil 可选目标协议默认。OpenAI chat 的绑定和白名单不会自动获得分类权限 |
| 目录及默认值 | Jev 类型默认隐含 text-only；旧宿主可能自己构造 ClassifierModel | 新模态、限额字段以及同名 chat/classifier 条目 | 新字段的零值/省略语义必须明确；避免新必填字段让原有 Jev 自带目录突然无效。图片能力只由显式声明授予；整个目录仍需换版本/哈希 |
| 模型效果 | 宿主可能已有 Jev 的置信阈值、分值解释和任务校准 | 数值形状相同不保证统计校准相同 | 旧阈值不直接搬给 Luna，按相同任务另做混淆/拒答/校准/延迟和消耗评估 |

现状依据：[公共请求与结果](/Users/cyber/RestoX/barness/ai/classifier.go)、[宽 JSON 问题](/Users/cyber/RestoX/barness/ai/classifier_questions.go)、[容量检查](/Users/cyber/RestoX/barness/ai/classifier_policy.go:155)、[native 编解码](/Users/cyber/RestoX/barness/ai/typesafe.go:78)、[严格答案/用量](/Users/cyber/RestoX/barness/ai/typesafe_response.go)、[score 与 noul](/Users/cyber/RestoX/barness/ai/typesafe_answers.go)、[封闭答案](/Users/cyber/RestoX/barness/ai/classifier_answers.go)、[Options/路由](/Users/cyber/RestoX/barness/ai/call_runtime.go:71)。Decisions wire 依据 [HTTP reference][official-reference]，问题类型及统计阈值指导依据 [guide][official-guide]。

其中最容易造成实际宿主不兼容的三种情况是：将带 JSON object 的 Jev instructions 原样切给 Luna；仅切 Target 却仍传 TypeSafeOptions；将新 route 的 nil error 当作普通答案保证，继续无分支地断言 ChoiceAnswer。这些都需要在迁移示例与公共 E2E 中明确覆盖。

授权、凭据固定、准入、取消、网络尝试、Observer 和 unary 资源释放可以继续共用。新路线的错误 body / request ID / 协议校验分别处理，默认不重试，2xx 后不重放的契约保持。不要因为 Pi 对 504 的观测更改 Jev 的共享重试行为；如为 Decisions增加例外，必须局限于这条路线且有故障证据。[共享 runtime](/Users/cyber/RestoX/barness/ai/call_runtime.go)、[unary](/Users/cyber/RestoX/barness/ai/unary.go)、[重试契约](/Users/cyber/RestoX/barness/ai/retry.go:153)

验收需同时证明：原有 Jev 请求/响应、错误分类、回调、用量、重试与资源释放不回归；新路线只在自己声明的能力下接受输入；同一公共请求在两路线上的合法性差异能在发网前正确解释。现有 Jev 差分、官方 wire replay 和 live 状态须保留，新路线增加自己的支持矩阵、目录证据与效果报告。[Jev 回放与目录](/Users/cyber/RestoX/barness/ai/e2e/classifier_builtin_test.go)、[Jev 差分](/Users/cyber/RestoX/barness/ai/e2e/classifier_pidiff_test.go)、[分类 live 合同](/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/classifier.go)

## 10. 离线验证与复现

使用仓库冻结的 **官方 `pi-ai@1.1.0` 发布包**直接调用 `dist/api/openai-decisions.js`，注入 synthetic fetch，不访问 OpenAI，不依赖未发布 main。Node `v24.15.0` 上 9 项 probe 通过，源输入与断言如下。数值成本以 `1e-12` 容差校验，因为 JS 浮点结果为 `0.060000000000000005`。

| Probe | 观察到的 Pi 行为 |
| --- | --- |
| 正常乱序答案 + 非零 cache details | 成功；按 name 回配；300000 input、cache 桶均零、成本 `$0.06`；score 分布丢弃 |
| 非法候选及数值 | 未请求 choice、重复概率 value、概率 `1.4/-0.2`、confidence `1.7/2`、score `99` 都被接受；未知 answer 忽略 |
| 重复答案 name | 最后一条覆盖前值，bool 从 `.8` 变 `.3` |
| 缺少一题 | 整体 error、空答案，保留 300000 input usage |
| 一题 refusal | 整体 error、空答案，保留 300000 input usage |
| 声明 PNG MIME 的 synthetic image | 成功；捕获单 user message，先 state text 再 data URL；此用例不验证真实 PNG 可解码 |
| 129 张图 | error，fetch 未调用 |
| 504 | 只发送一次，error，错误描述无 HTML |
| 503 后成功 | 发送两次，成功 |

可从 repo root 重放以下代码；最后输出的 JSON 是可保存、比对的验证 artifact。通过的预期来自上表和固定源码，不代表线上质量或费用已经验收。

```bash
PI_AI_DIST="$PWD/ai/internal/testkit/pioracle/node/node_modules/@earendil-works/pi-ai/dist" node --input-type=module <<'JS'
import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import {pathToFileURL} from 'node:url';
const d=process.env.PI_AI_DIST;
const {classify}=await import(pathToFileURL(d+'/api/openai-decisions.js'));
const cat=JSON.parse(await fs.readFile(d+'/providers/data/openai.json','utf8'));
const m=cat['openai-decisions']['classifier:gpt-6-luna'];
assert.equal(m.type,'classifier');assert.equal(m.id,'gpt-6-luna');assert.equal(m.api,'openai-decisions');
const c={state:{text:'approved'},questions:{
 q:{type:'choice',instructions:'pick',criteria:{a:'A',b:'B'}},
 p:{type:'bool',instructions:'approve?',criteria:{true:'Approve',false:'Decline'}},
 s:{type:'score',instructions:'score',criteria:['low','high']}}};
const a=[
 {type:'choice',name:'q',choice:'a',probabilities:[{value:'a',probability:.8},{value:'b',probability:.2}],confidence:.6},
 {type:'predicate',name:'p',probability:.8},
 {type:'score',name:'s',score:.8,probabilities:[{value:0,label:'low',probability:.2},{value:1,label:'high',probability:.8}],confidence:.6}];
const u={input_tokens:300000,input_tokens_details:{cached_tokens:100000,cache_write_tokens:10000},output_tokens:0,total_tokens:300000};
const rows=[];
async function run(name,context,answers,stopReason){
 let attempts=0,payload;
 const result=await classify(m,context,{apiKey:'synthetic-key',maxRetries:2,fetch:async(_,init)=>{
  attempts++;payload=JSON.parse(init.body);
  if(name==='504')return new Response('<html>timeout</html>',{status:504});
  if(name==='503'&&attempts===1)return new Response('busy',{status:503,headers:{'retry-after-ms':'0'}});
  return Response.json({answers,usage:u});}});
 assert.equal(result.stopReason,stopReason);
 rows.push({name,attempts,result,...(name==='image'?{payload}:{})});return result;
}
const good=await run('out-of-order',c,[...a].reverse(),'stop');
assert.deepEqual(good.answers.s,{type:'score',score:.8,confidence:.6});
assert.equal(good.usage.input,300000);assert.equal(good.usage.cacheRead,0);
assert.equal(good.usage.cacheWrite,0);assert.ok(Math.abs(good.usage.cost.total-.06)<1e-12);
const bad=await run('invalid-values',c,[
 {...a[0],choice:'unrequested',confidence:1.7,probabilities:[{value:'a',probability:-.2},{value:'a',probability:1.4}]},
 {...a[1],probability:-.2},{...a[2],score:99,confidence:2},
 {type:'predicate',name:'unrequested',probability:1}],'stop');
assert.equal(bad.answers.q.probabilities.a,1.4);assert.equal(bad.answers.q.choice,'unrequested');
assert.equal(bad.answers.q.confidence,1.7);assert.equal(bad.answers.p.probability,-.2);
assert.equal(bad.answers.s.score,99);assert.equal(bad.answers.s.confidence,2);
assert.deepEqual(Object.keys(bad.answers).sort(),['p','q','s']);
const dup=await run('duplicate',c,[...a,{...a[1],probability:.3}],'stop');
assert.equal(dup.answers.p.probability,.3);
for(const [name,answers] of [['missing',a.slice(0,2)],['refusal',[a[0],{type:'refusal',name:'p'},a[2]]]]){
 const r=await run(name,c,answers,'error');assert.deepEqual(r.answers,{});assert.equal(r.usage.input,300000);}
const img={type:'image',data:'aW1hZ2U=',mimeType:'image/png'};
await run('image',{...c,images:[img]},a,'stop');
await run('129-images',{...c,images:Array(129).fill(img)},a,'error');
await run('504',c,[],'error');await run('503',c,a,'stop');
assert.deepEqual(rows.find(r=>r.name==='image').payload.input,[{role:'user',content:[
 {type:'input_text',text:JSON.stringify(c.state)},
 {type:'input_image',image_url:'data:image/png;base64,aW1hZ2U='}]}]);
assert.equal(rows.find(r=>r.name==='129-images').attempts,0);
assert.equal(rows.find(r=>r.name==='504').attempts,1);assert.equal(rows.find(r=>r.name==='503').attempts,2);
assert.ok(!rows.find(r=>r.name==='504').result.errorMessage.includes('<html>'));
process.stdout.write(JSON.stringify({model:m,rows},null,2)+'\n');
JS
```

## 第一方来源索引

[tag110]: https://api.github.com/repos/earendil-works/pi/git/ref/tags/v1.1.0
[tag104]: https://api.github.com/repos/earendil-works/pi/git/ref/tags/v1.0.4
[feature-commit]: https://github.com/earendil-works/pi/commit/ce8972a0e99cab79f2620a594a1c17baef436c67
[release]: https://github.com/earendil-works/pi/releases/tag/v1.1.0
[changelog]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/CHANGELOG.md#L3-L16
[npm110]: https://registry.npmjs.org/@earendil-works%2fpi-ai/1.1.0
[npm104]: https://registry.npmjs.org/@earendil-works%2fpi-ai/1.0.4
[npm104-tar]: https://registry.npmjs.org/@earendil-works/pi-ai/-/pi-ai-1.0.4.tgz
[package]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/package.json#L69-L79
[catalog-source]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/scripts/generate-models.ts#L2689-L2708
[catalog-flatten]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/model-catalog.ts#L19-L80
[pricing-helper]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/scripts/generate-models.ts#L391-L439
[cost-calculate]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/models.ts#L1200-L1219
[classifier-types]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/types.ts#L649-L710
[provider]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/providers/openai.ts#L9-L34
[dispatch]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/models.ts#L972-L988
[model-guards]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/utils/model-operations.ts#L39-L53
[lazy]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/api/openai-decisions.lazy.ts
[classifier-http]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/api/classifier-shared.ts#L41-L104
[number-check]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/api/classifier-shared.ts#L30-L39
[usage-parser]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/api/classifier-shared.ts#L106-L128
[question-builder]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/api/openai-decisions.ts#L43-L71
[input-builder]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/api/openai-decisions.ts#L73-L92
[answer-parser]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/api/openai-decisions.ts#L94-L144
[gateway]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/api/openai-decisions.ts#L146-L158
[decisions-call]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/api/openai-decisions.ts#L160-L197
[retry]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/utils/provider-retry.ts#L24-L127
[classifier-readme]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/README.md#L891-L954
[models-doc]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/coding-agent/docs/models.md#L131-L144
[codemode-doc]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/coding-agent/docs/codemode.md#L168-L184
[codemode-validation]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/coding-agent/src/extensions/codemode/execute.ts#L118-L170
[codemode-run]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/coding-agent/src/extensions/codemode/execute.ts#L638-L684
[codemode-concurrency]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/coding-agent/src/extensions/codemode/execute.ts#L49-L50
[route-tests]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/test/classifier-models.test.ts#L200-L246
[mapping-tests]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/test/openai-decisions.test.ts#L80-L139
[image-tests]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/test/openai-decisions.test.ts#L141-L183
[refusal-tests]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/test/openai-decisions.test.ts#L185-L216
[error-tests]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/test/openai-decisions.test.ts#L235-L290
[official-guide]: https://developers.openai.com/api/docs/guides/decisions
[official-reference]: https://developers.openai.com/api/reference/resources/decisions/methods/create
[official-go-reference]: https://developers.openai.com/api/reference/go/resources/decisions/methods/create
[official-model]: https://developers.openai.com/api/docs/models/gpt-6-luna
