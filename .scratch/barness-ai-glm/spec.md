# barness-ai 接入 GLM：Responses、Chat Completions 与 Messages

Status: ready-for-agent
Type: feature
Updated: 2026-10-17

本文件是待实施的交付契约，不表示 GLM 已接入或真实冒烟已通过。依据为[GLM 三种兼容协议研究](research.md)，研究包含官方来源、文档冲突与已复现的 Go SDK 解析风险；本次制定 spec 不使用真实凭据或执行收费请求。

领域与架构依据：[领域词汇](../../GLOSSARY.md)、[协议边界 ADR-0001](../../docs/adr/0001-barness-ai-protocol-boundary.md)、[DeepSeek Responses ADR-0014](../../docs/adr/0014-barness-ai-deepseek-responses.md)、[DeepSeek Chat ADR-0015](../../docs/adr/0015-barness-ai-deepseek-chat-completions.md)、[型号纳入 ADR-0018](../../docs/adr/0018-barness-ai-catalog-inclusion-criterion.md)。

## Problem Statement

Host 希望在本地和云端通过现有 barness-ai 调用 GLM，自主选择 Responses、Chat Completions 或 Messages，同时保留统一的消息、工具、流事件、结果、用量与租户隔离语义。当前模块已具备三套协议 adapter 和 DeepSeek 的兼容厂商接入方式，但没有智谱独立的 Provider、协议能力配置与正式模型目录。

仅更换 endpoint 或把 GLM 标为 OpenAI/Anthropic，会混用服务端能力、鉴权和模型配置。现有 Responses Go SDK 对 GLM 文档允许的单对象 content 不报错却丢失内容；GLM-5.3 Chat 不允许关闭思考；Messages 的高级字段缺少完整公开兼容表。直接沿用现有默认值可能导致有效调用失败、推理回放改变语义、工具强制选择失效或产生错误的支持声明。

2026-10-16 的 Key 分类调整又使凭据用途与 endpoint 不再一一对应。Host 还需要知道实际获准使用的账户、型号和计费路径，不能依赖全局 Key、自动切换协议或历史经验推定套餐额度。

## Solution

新增中国区智谱 Provider，沿用 DeepSeek 的接入方式：Provider 身份独立，三个 API Protocol 各自使用明确的 Binding 与模型条目，共享现有协议 adapter、调用生命周期和资源控制，按 Provider 协议能力与型号硬约束生成请求。首期交付三个协议上的文本聊天、流式结果、函数工具往返、思考输出及合法历史回放。

首批目标型号为 `glm-5.3` 与 `glm-5.2`，各自覆盖三种协议，共六个独立 Model 身份。两者均按文本输入/输出纳入；所有承诺的型号与协议组合都须有自身证据，不能从另一协议复制支持结论。以标准 API Key 和标准/通用官方端点作为正式验收配置。

保持公共 Client 调用形状；Host 负责认证授权、凭据装配、上下文组织和工具执行。Responses 使用无状态完整历史；三协议不自动互转或故障切换。交付包含可重放公共 E2E 证据、三个组合各自的真实冒烟、目录与价格快照、差分登记、文档和完整发布门禁。

## User Stories

1. As a Host developer, I want GLM to have its own Provider identity, so that routing and attribution identify the actual service operator.
2. As a Host developer, I want GLM Responses through the existing Client, so that I can retain my current Responses integration.
3. As a Host developer, I want GLM Chat Completions through the existing Client, so that I can reuse my Chat integration.
4. As a Host developer, I want GLM Messages through the existing Client, so that I can reuse my Messages integration.
5. As a Host developer, I want Stream and Complete on every GLM route, so that delivery style does not change the supported operation.
6. As a Host developer, I want the simple entries on every GLM route, so that protocol-neutral calls remain available.
7. As a Host developer, I want existing protocol-specific options retained, so that adopting GLM does not require another request API.
8. As a Host developer, I want incorrect protocol options rejected, so that switching a Binding cannot reinterpret options silently.
9. As a tenant administrator, I want separate Bindings for each protocol, so that each route has explicit authorization.
10. As a tenant administrator, I want model allowlists checked before credentials are read, so that denied calls do not access secrets.
11. As a tenant administrator, I want shared Clients to isolate tenant accounts, so that concurrent requests cannot mix credentials or state.
12. As a Host operator, I want a fixed credential and configuration snapshot per Logical Call, so that retries retain one authorized identity.
13. As a Host operator, I want standard Key examples, so that initial deployment has an explicit billing purpose.
14. As a Host operator, I want Key categories distinguished from endpoints, so that URL changes do not imply a different quota source.
15. As a Host developer, I want independent model entries per protocol, so that one protocol's settings cannot alter another's behavior.
16. As a Host developer, I want text-only model capabilities described accurately, so that an image container in the protocol is not mistaken for model vision support.
17. As a Host developer, I want GLM-5.3's mandatory thinking respected, so that omitted options cannot produce an invalid disabled request.
18. As a Host developer, I want GLM-5.2's thinking behavior configured independently, so that older model mappings are not applied to GLM-5.3.
19. As a Host developer, I want canonical levels separated from wire effort, so that sampling and model capability remain understandable.
20. As a Host developer, I want full and simple reasoning behavior documented, so that unsupported levels have predictable outcomes.
21. As a Host developer, I want reasoning deltas and final thinking retained, so that available model output is not silently lost.
22. As a Host developer, I want Responses object and array content decoded correctly, so that valid final text and reasoning remain complete.
23. As a Host developer, I want authoritative final content to reconcile streamed deltas, so that the Result reflects the completed response.
24. As a Host developer, I want the original native reasoning preserved separately, so that parsing normalization cannot corrupt replay.
25. As a Host developer, I want complete Responses history replay, so that I do not depend on provider-side saved sessions.
26. As a Host developer, I want function tools on all three protocols, so that the Host can continue its existing tool loop.
27. As a Host developer, I want tool argument deltas where supported, so that streaming tools expose progress before completion.
28. As a Host developer, I want tool calls and results paired across turns, so that tool round trips retain their meaning.
29. As a Host developer, I want interleaved Chat reasoning returned with tool results, so that the next generation has its required context.
30. As a Host developer, I want unsupported tool choices rejected before sending, so that a requested constraint cannot be silently weakened.
31. As a Host developer, I want Messages thinking and signatures handled using GLM evidence, so that Claude-specific assumptions do not corrupt my history.
32. As a tenant administrator, I want persisted native state restored only through Host trust, so that client-supplied signatures do not grant replay authority.
33. As a tenant administrator, I want cross-account native state downgraded using existing rules, so that state cannot cross an unauthorized boundary.
34. As a Host developer, I want supported prompt cache hints isolated by tenant and account, so that raw session identifiers are not shared with the provider.
35. As a Host developer, I want cache retention promises kept explicit, so that a routing key does not imply a guaranteed lifetime or cache hit.
36. As a Host developer, I want usage omissions distinguished from reported zero, so that missing metadata is not represented as free usage.
37. As a Host developer, I want reasoning tokens counted within output, so that estimates do not double count reasoning.
38. As a Host operator, I want prices expressed in the existing dollar contract, so that RMB prices cannot be misinterpreted as dollar rates.
39. As a Host operator, I want price snapshots attributable to sources and dates, so that estimates can be reproduced.
40. As a Host developer, I want trusted callbacks on GLM routes, so that my existing request customization still works.
41. As a tenant administrator, I want the final callback payload revalidated, so that callbacks cannot bypass routing, state or model constraints.
42. As a Host developer, I want immutable request snapshots, so that later changes to my maps do not affect an active call.
43. As a Host operator, I want explicit bounded initial retries, so that transient failures can be handled without hidden duplicate requests.
44. As a Host operator, I want quota exhaustion distinguished from transient throttling, so that HTTP 429 does not cause pointless retries.
45. As a Host operator, I want no replay after streaming starts, so that a partial response is not regenerated or charged twice by the library.
46. As a Host developer, I want cancellation and deadlines to produce one terminal result, so that my application can finish calls reliably.
47. As a Host operator, I want byte, queue and concurrency limits retained, so that GLM cannot bypass resource controls.
48. As a Host developer, I want premature EOF and malformed streams reported as errors, so that incomplete output cannot appear successful.
49. As a Host developer, I want provider stop reasons and request IDs preserved when available, so that failures remain diagnosable.
50. As a tenant administrator, I want metadata-only observation and redacted artifacts, so that operations do not leak keys, prompts or reasoning.
51. As a maintainer, I want one public Client E2E seam, so that tests verify caller-visible behavior across all three adapters.
52. As a maintainer, I want failure scenarios written before implementation, so that the implementation is driven by observable contracts.
53. As a maintainer, I want replayable wire and result artifacts, so that acceptance can be verified repeatedly without paid calls.
54. As a maintainer, I want each GLM protocol tested against its own official API, so that shared adapter success cannot substitute for GLM evidence.
55. As a maintainer, I want model-specific live results, so that one model's success does not certify the other model.
56. As a maintainer, I want documented unknowns resolved by bounded probes, so that absent documentation is not mistaken for unsupported functionality.
57. As a maintainer, I want scoped pi differential or extension records, so that the compatibility claim reflects the actual frozen route.
58. As a maintainer, I want existing providers covered by regression checks, so that capability refactoring cannot break DeepSeek or native providers.
59. As a maintainer, I want current support-matrix and release-gate evidence, so that incomplete routes cannot be published as supported.
60. As a Host operator, I want a reversible release, so that I can disable GLM without changing accounts or rerouting calls automatically.
61. As a Host developer, I want complete Host configuration examples, so that local and cloud deployments use the same module contract.
62. As a maintainer, I want architecture decisions and evidence limits recorded, so that future changes preserve the reasons for the design.

## Implementation Decisions

### R01 — 操作、Provider 和协议身份

新增 `ProviderZhipu`，值为 `zhipu`，表示本次研究的中国区智谱服务。Operation 为 chat；API 继续使用 `openai-responses`、`openai-completions`、`anthropic-messages`。这里的 Completions 指 Chat Completions，不包括旧文本 Completions。复用现有三套 adapter；不新增 GLM 专有 adapter、协议名、代理服务或 Agent 运行时。

### R02 — Binding、端点与鉴权

每个 Binding 固定一种 Operation × Provider × API，三份配置互不替代。正式接入使用下列 SDK 基址，SDK 附加资源路径后必须精确得到对应请求地址；配置示例不得重复附加版本或资源段。

| API | Endpoint | 生成请求地址 | 鉴权 |
| --- | --- | --- | --- |
| openai-responses | `https://open.bigmodel.cn/api/v1` | `POST https://open.bigmodel.cn/api/v1/responses` | Bearer API Key |
| openai-completions | `https://open.bigmodel.cn/api/paas/v4` | `POST https://open.bigmodel.cn/api/paas/v4/chat/completions` | Bearer API Key |
| anthropic-messages | `https://open.bigmodel.cn/api/anthropic` | `POST https://open.bigmodel.cn/api/anthropic/v1/messages` | x-api-key |

核心继续仅接收 Host 解析出的 Binding 和 Credential Snapshot，不读环境 Key、不使用 SDK 的环境/profile 自动发现。同租户、同账户且授权一致时可以显式共享 CredentialRef；协议、租户或账户不同不能据同一模型名共享调用快照。取消、错误或限流都不自动切换 endpoint、Provider、API、Key 或型号。

### R03 — 凭据用途和默认部署

正式验收仅使用中国区标准 API Key 和上述端点。标准、个人套餐、团队套餐、历史兼容 Key 的用途与来源由 Host 管理，不新增凭据字符串识别、账户权益查询或库内账务系统。示例说明 2026-10-16 分类后的 Key 规则，区分新 Key 的用途与历史兼容 Key 的 endpoint/账户历史行为。Coding Plan 不属于首期正式支持变体；不得把标准 Key 的通过解释为套餐资格或扣费行为已验证。

### R04 — 公共接口与模块责任

保留 Client 的 Stream、Complete、StreamSimple、CompleteSimple 和 WithHooks，以及现有 Request、Transcript、工具、消息、事件、结果、Usage 与 Call Attribution。full 分别使用 ResponsesOptions、ChatOptions、AnthropicOptions；nil 选择该 Provider × API 的有效默认，错误协议选项仍拒绝。Host 执行工具并发起下一轮，模块一次只完成一个 Generation Turn。

新增内部能力或型号元数据时只表达已出现的 GLM 差异；不增加调用级 credentials/endpoint、通用策略插件注册或另一套重试/组装运行时。若既有 ModelCompat 不能表达新的型号硬约束，可增加语义清晰、默认不改变旧模型行为的字段；不得以硬编码模型 ID 判断代替目录数据。公共 DTO 的新增字段应可省略、可复制、参与目录版本与 Hash。

### R05 — Provider 协议能力与型号硬约束

按 Provider 区分 Chat 请求形状和 Responses 服务端能力；为 Messages 建立同层级的小型 Provider 能力配置。拆开当前绑定在一起的缓存路由键、保留期和亲和头，单独控制 reasoning summary、加密推理与可选厂商字段。重构内部 DeepSeek 专名时直接迁移并删除旧实现，不保留兼容 shim；OpenAI、DeepSeek、Anthropic 原有请求与语义保持。

Provider 能力不能被 Host Catalog 打开到超出该端点的已实现范围；模型能力只能进一步收窄。型号的不可关闭思考、支持档位、输入模态和输出上限归 Model；同一 Provider 的所有型号不必拥有相同思考行为。未在文档出现的字段按“未接入”处理，不称为官方已拒绝。

### R06 — 目标型号、目录和来源

首期六个目标条目是 `glm-5.3`、`glm-5.2` 各自的三个 API 身份，独立保存，不能从 Chat 条目自动派生 Responses 或 Messages 数据。均为 text-only，不因 Responses 图片示例增加视觉能力；图片输入沿用既有文本型号历史降级/占位契约，不扩大图片或其他 Operation 授权。

型号上下文、输出上限、档位、价格和来源在候选阶段固定为有日期的元数据；研究中 1M/128K 的单位不得直接冒充已核验的精确整数。实现先从官方 schema/模型资料取得明确界限，无法取得时使用明确标为 barness 保守上限且不超过已证实容量的候选配置，禁止凭空填充“厂商精确值”。按 ADR-0018 遵守全部硬约束后，候选通过才纳入内置目录；六个目标均为本 spec 的交付范围，不能默默删去失败型号或协议以放行。

### R07 — 思考档位、默认和采样

保留 canonical level 与 wire effort 的独立含义。GLM-5.3 Chat 只声明 low/high/max；off/minimal/medium/xhigh 明确标为不支持，simple 按既有向上优先的规则夹取，full 对不支持的显式档位请求前返回 invalid_request。省略档位不发送 disabled，保留服务端默认思考；不声称该默认等于客户端采样选择的 canonical level。GLM-5.2 按其独立目录和官方已确认的开关、effort 映射处理。

Responses 的 GLM-5.3 effort 文档与型号页有差异；首期只承诺 low/high/max 的输入子集，仍记录 Responses 文档对 low 的 wire/服务端映射，不能承诺与 Chat 相同强度。none/minimal 的关闭行为、medium/xhigh 的别名必须核验后才能加入公开能力；未核验时 full 拒绝、simple 按已声明子集夹取。省略值使用该模型/协议已证实的默认形状，不自动套入 OpenAI 的 none。

OpenAI 两种兼容协议继续按既有顺序应用通用模型采样、有效 canonical 档默认、调用覆盖及 OnPayload，回调后不重新选档；GLM Messages 继续忽略 SamplingParams 与分档 map。最终 wire 不得违反型号硬约束，即使采样或回调覆盖了字段也须重新检查。GLM 专属校验不改变现有 OpenAI/DeepSeek full wire 与采样选档分离的契约。

### R08 — Responses 请求能力

显式发送 store=false，使用完整 input 历史，不发送 previous_response_id、conversation 或服务端保存引用。核心请求包含已支持的 messages、函数调用/结果、reasoning、输出预算与适用采样字段；reasoning effort 不自动带 summary=auto 或 encrypted include。默认不生成 service tier、缓存保留期、OpenAI 亲和头、background 或未接入字段。

支持 tool_choice 的 none/auto；required、具名强制、allowed_tools 等本期未接入形式在请求前拒绝，不降级为 auto。函数 strict 不自动生成。GLM max_output_tokens 按其 schema 的 1 起下限和型号上限处理，不继承 OpenAI 的 16 起隐藏抬升；显式预算超出型号有效范围时明确拒绝。格式化输出只在当前公共选项/安全回调已能表达且经证实的 text/json_object 子集内使用，不新增严格 JSON Schema 能力承诺。

### R09 — Responses wire 解析与原生回放

在原始 JSON 边界处理 GLM 文档允许的 reasoning.content 单对象、message.content 单对象或数组；内部统一为解析器可处理的形状，再进入既有 assembler。对象必须按结构解码，不做字符串替换，不依赖 SDK 不报错来认定内容有效。正文、推理与工具参数的 delta 和 authoritative done 都要处理；只有 done 内容的流也须完整。

保留原始 GLM reasoning 条目与解析用规范形状的区别。Native State 回放遵循 GLM 的 InputReasoning 形状；内部规范化产生的数组、OpenAI synthetic ID、summary/encrypted 字段不得污染原生回放。工具 call_id、工具条目 ID 与函数结果配对按本协议证据转换，保留现有可信来源检查。

### R10 — Responses 生命周期与缓存

处理 response.reasoning_text 及现有适用的正文、item、工具参数事件，以 completed/incomplete/failed/error 等既有协议终态规则完成；GLM 不发 [DONE] 时应正常收束，缺有效终态的 EOF 必须失败。结果与终态事件一致，不允许单对象 done 把已流出的正文错误清空。

有 SessionID 且 CacheRetention 不为 none 时，可发送现有租户、账户、session 派生的 prompt_cache_key；无 SessionID 或 retention=none 不发送。short/long 仅沿用路由提示，不派生 GLM 未接入的 TTL/retention/亲和头，不承诺缓存命中或期限。显式请求 reasoning summary、ServiceTier、encrypted include 等未接入能力返回 invalid_request；缓存提示不派生的既有约定保持并写明。

### R11 — Chat 请求形状与工具流

使用 system 指令、max_tokens、嵌套 function 工具和合法 thinking/reasoning_effort；不生成 store、developer、max_completion_tokens 或 OpenAI 缓存字段。首期仅接入 tool_choice=auto；其他明确指定形式请求前拒绝，包括 full 的 required/具名强制和 simple 的 none，不能静默改写或靠上游忽略满足请求。SimpleOptions 的公共枚举继续只有 auto/none，不为 GLM 增加 required。

当声明函数工具并流式调用时发送 tool_stream=true，按实际工具 index/ID 组合参数，参数完整后再交给 Host 执行。stream_options.include_usage 在未核验前不发送；通过有/无字段的候选对照记录兼容事实，但正式配置不得因为缺少 usage 强制套用 OpenAI 字段。Chat 的 [DONE]、finish_reason、流末 usage 和同帧内容分别读取，不能在 finish_reason 出现时漏掉后续用量。

### R12 — Chat 推理历史

读取 reasoning_content 增量并按原来源字段回放；GLM 不继承 DeepSeek“每条 assistant 必须补空 reasoning_content”的特殊规则。工具往返须完整保留同一 turn 中必要的 reasoning_content、工具调用和结果顺序。

标准端点采用 clear_thinking=true 的跨 turn 行为；不新增 Preserved Thinking 的公共开关或自动保存/裁剪策略。可信回调若要求本期未接入的 clear_thinking=false，应明确拒绝而不假称完整保留。正常同 turn 回放不等于开启跨 turn 保留；持久化恢复仍须 Host 担保来源。

### R13 — Messages 的证据建立与能力配置

复用现有 Anthropic SDK 请求服务和 SSE 解码路径，使用显式配置、x-api-key 和关闭的 SDK 重试。先用基本 messages 请求核对路径、SDK 默认 anthropic-version 与最终 headers，再建立 GLM 自身的文本、thinking、工具和 usage fixture。未取得契约证据前不自动生成 cache_control/TTL、eager_input_streaming、thinking display、interleaved beta、adaptive/block_binding、逐轮 managed effort 或服务端 fallback。

Messages 思考参数、effort 映射、签名、工具选择和用量位置必须由官方逐字段资料或有界真实探针确认，不能从 Chat/Responses 外推。探针只用于确定 wire 事实；正式支持必须由公共 Client 再验证。不能确定基础思考和工具往返时，本组合保持未完成，不得凭基本文本成功发布为完整支持。

### R14 — Messages 的公共行为与回放

必须交付文本、可用思考输出、函数 tool_use/tool_result 往返、同模型续接和合法 SSE 终态。nil/simple 省略思考时使用 GLM 已证实的默认形状；始终思考型号不得由现有 simple 默认产生 ThinkingEnabled=false。显式支持档位仅通过经证实的 wire 映射，未接入的 budget/effort/adaptive 选项请求前拒绝，不能把预算和强度互相猜换。

首期工具选择至少支持默认和 auto；其他模式只有在本组合证据确认、公共 guard 与 E2E 同时覆盖后才声明为支持，否则明确拒绝。signature、redacted_thinking、无签名 thinking 的回放使用 GLM 的实际规则；不伪造 Claude 签名，不把无签名 thinking 自动改成 text 后称作原生回放。若上游明确不返回可观察的思考块，可以有来源与真实证据地登记该可观察能力 UNSUPPORTED；工具往返与合法继续生成仍为必交，不可豁免。

### R15 — 最终请求校验与 callback 顺序

保持 header → payload → response 的既有顺序，OnPayload 仍接收目标协议的原生请求。模型默认、采样覆盖、回调完成后，校验实际最终请求的模型、目标、工具/原生引用授权、协议能力、型号硬约束和资源限制；检查具名选项、SamplingParams、OnPayload 各入口，不能只检查原始参数。

拒绝未接入字段时保留可诊断原因；模型、凭据、目标和账户引用越权继续按现有 tenant_denied 等授权错误处理，不为新能力改变既有优先级。这些发送前校验失败均为零 Provider 请求、零 Attempt；OnResponse 等发送后失败保留实际发生的 Attempt 与已知用量。已允许的合法回调修改冻结一次，重试复用最终字节，不共享可变 map 或原生 JSON。

### R16 — Native State 与租户隔离

所有消息、流事件、结果和 Attempt 带实际 zhipu/目标 API/型号/账户归属；完整模型身份包含 Operation。库内可信结果可同模型续接，持久化读回必须由 Host 重新 TrustNativeState。跨模型、Provider、API、账户或不可信来源按既有规则降级；降级保留可见内容与有效工具配对，不回放账户专有状态，并记录降级元数据。

同租户授权的跨模型转换不因历史来自另一账户整体拒绝；只检查最终仍保留的账户引用。相同 BindingID、ModelID 或 session 名不得造成跨租户 credential/cache/native-state 混用。全局唯一 RequestID 的 Host 义务保持，隔离测试使用不同 RequestID。

### R17 — 用量和价格

三个协议独立解析 usage 字段与 presence，沿用 Input 排除已单计缓存、Reasoning 属于 Output 的聊天约定；不重复叠加 reasoning token，不把缓存写入或缺失明细猜成已上报零。缺失是 partial/unreported，完整合法零才是零用量证据；Responses 缺 total_tokens 时按已知项计算，完整性按该协议现有规则判断。流中断保留已合法上报的用量，failed-response 用量沿用既有 Responses 契约，不借本特性扩大已接受的取值行为。

CostRates 与 UsageCost 继续采用美元/百万 token 和美元金额。人民币官方价格必须以有来源、日期和固定换算率的版本化 Price Snapshot 转为美元；不增加运行时汇率请求、不将人民币数直接填入美元字段。换算依据、缓存费率和上下文阶梯在正式纳入前齐备；无依据的零价格不得作为免费发布。估算费用不表示套餐最终扣款或厂商账单。

### R18 — 错误分类与重试

沿用稳定 Error Code/Phase、HTTPStatus、原始 stop reason 和可取得的 ProviderRequestID，不泄露 Key 或上游回显的敏感正文。Responses 字符串业务码、Chat 数字业务码、Messages 错误对象分别解释；HTTP 200 的流内错误或缺终态不能作为成功。

Chat 的 sensitive、network_error、model_context_window_exceeded 不映射为普通 stop。对已知欠费、额度耗尽、套餐到期、型号权益不足等不可恢复业务码，即使 HTTP 429 也禁止自动重试；暂时速率/过载只按 Binding 的显式有界策略在流开始前重试。未知 429 保留既有 HTTP 分类与策略，不猜测为某种账务错误。SDK 默认重试关闭，流开始后不重放；身份和 payload 在全部 Attempt 中固定。

### R19 — 资源、取消与观测

继续经过 Admission、调用/请求时限、请求/响应/帧/输出字节和事件队列限制；所有取消、错误、回调失败和限额终态释放 body、goroutine 与许可，最终结果仅结算一次。厂商 7200 秒建议不覆盖 Host ResourcePolicy；本地取消只保证本地终止，不承诺远端取消或免计费。

Observer 沿用只含元数据的合同，不加入提示、思考、工具参数、原生状态、原始错误正文或 Key 类别中的秘密。request id 的 header/错误体来源按实际证据实现，不照抄 DeepSeek 的 trace header，也不把本地 RequestID 当作厂商已返回的 ID。

### R20 — pi 基线与扩展登记

保持当前唯一冻结 pi-ai 1.1.0 oracle 及其来源 pins。逐一核查该基线的 Provider 身份与 API 路由；不能仅用模型名相同、国际站名相似或基址相似判定同一官方服务。存在实际适用路由时进入该组合差分；无路由时按 ADR-0014 登记 Extension Route，以官方协议 fixture 与自身 Live Smoke 证明。

差异按 case/path 和来源登记 fixed/extension/pending；未解释差异保持 pending，不能豁免整个协议、请求体、结果或仅因为新 Provider 就宣称差分通过。Shared adapter 的 OpenAI/Anthropic/DeepSeek 差分通过不计为 zhipu 通过。

### R21 — 候选核验、正式支持与发布

新增 zhipu-chat、zhipu-responses、zhipu-messages 三个 chat 组合，每个组合的必交场景覆盖 glm-5.3 和 glm-5.2，并单独记录型号、SDK、目录 Hash、endpoint、标准 Key 类别、账户/区域别名和能力结果。候选 Host Catalog 先取得真实协议证据，再生成正式目录；候选通过不代替正式内置目录 Hash 下的完整报告。

沿用 live 双开关和每进程仅本组合 Key 的装配；正式 live 的 Binding 重试关闭，显式重试策略由离线 E2E 验证。每次组合运行上限为 24 个 Logical Call、32 个 HTTP Attempt，每次生成最多 4096 输出 token，进程时限 1800 秒、单调用不超过 120 秒且受 Host 更短时限约束。两个型号的预定场景在执行前登记和预留预算；预算耗尽则失败，不隐式增加额度。需修改固定预算时先改规范和证据计划，不能运行中放宽。厂商故障的场景级重试最多一次并计入同一预算，SDK/Binding 重试与场景重跑不能重复绕过预算。

Support Matrix 仅由本组合自己的完整报告更新，保留 PASS/FAIL/NOT_RUN/UNSUPPORTED；缺 Key 是 NOT_RUN、真实故障是 FAIL。两型号中的必交失败均阻止该组合完整通过。UNSUPPORTED 必须附确切来源/真实证据，只用于未承诺的可选能力，不能豁免文本、工具往返、合法历史、隔离、终态和资源约束。既有十个正式组合继续保留，新三组合纳入同一发布合同，不删除旧 P0、差分、压力、race、业务效果或脱敏要求以放行。

### R22 — 文档、ADR 与回滚

交付文档说明三份 Host Binding、标准 Key 装配、型号差异、full/simple 选项、已接入/未接入能力、Callbacks、用量与币种、无状态历史、诊断和真实支持状态。关键 Provider 能力拆分、Messages 非原厂行为与 Native State wire 决策在实施时同步 ADR，并说明与既有 ADR 的继承/修改关系；不把候选推断写成 accepted 厂商事实。

无数据库迁移，无新增对外 HTTP 接口。Host 可禁用 GLM Binding；软件回滚恢复完整先前实现、目录、能力配置、矩阵与发布证据版本，不改投另一个 Provider，不增加旧路径 shim、双写或自动兼容分支。

## Testing Decisions

### T01 — 唯一业务验收接缝

采用现有公共 Client：Stream、Complete、StreamSimple、CompleteSimple 和 WithHooks。Host 提供受控 Binding/Credential Resolver、Catalog、ResourcePolicy 和 HTTP transport；离线 TLS Provider 替身捕获最终请求并发送脚本化响应，真实 API 验收仍通过同一个 Client。覆盖目录查询与配置隔离的公共入口，不以私有 helper 单测替代业务验收，不要求上层 Agent/TUI。

用户已于 2026-10-17 确认沿用公共 Client E2E 边界，离线替身与真实 API 共用该入口。这个接缝沿用 DeepSeek 两协议、原生 Anthropic 与可信回调/资源隔离测试的先例；测试关注可见请求、事件、Result、错误、用量、归属和释放结果，不断言内部函数名、能力结构布局或 SDK 类型。

### T02 — 先失败场景，再对应实现

先列出下表所有失败方式并写公共 E2E，然后实现相应行为；不能在实现后补镜像单测。对文档缺失的 Messages wire，先用已有公共入口及有界候选探针获取真实形状，再把该形状转成离线失败场景，最后补具体适配。协议探索脚本、来源哈希检查和目录核验不构成第二套业务验收接口。

| 验收组 | 必验外部行为与失败场景 | 对应决策 |
| --- | --- | --- |
| A01 路由与认证 | 三种基址得到正确资源路径和鉴权；错误 API 选项、禁用 Binding、Operation/型号越权在读取 Key/发请求前拒绝；不读取环境凭据 | R01–R04 |
| A02 型号与思考 | 两型号、三协议独立；5.3 不产生 disabled；simple 夹取与 full 拒绝可预测；5.2 映射不泄漏；省略与显式选项、callback 覆写均遵守硬约束 | R05–R07 |
| A03 Responses shape | reasoning/message 单对象与数组，只有 done、delta 后 done 校正、不同 item index；权威终帧无丢失/重复；畸形结构明确失败 | R09–R10 |
| A04 Responses 全历史 | 无保存的文本/思考/函数调用→结果→继续生成；原生 reasoning 保持 GLM 形状；call_id 与条目 ID 不混；无 [DONE] 正常、无终态 EOF 失败 | R08–R10、R16 |
| A05 Chat 流与历史 | reasoning_content、正文和多工具参数分片；tool_stream=true；finish_reason 后 usage；同 turn 推理回放；跨 turn clear_thinking 默认不变 | R11–R12 |
| A06 Messages | GLM 自身 message_start/content blocks/message_delta/message_stop；工具 JSON 分片；有/无 signature 的合法回放；流内 error、非法顺序和缺终态；不伪造签名或发送未接入 beta/cache | R13–R14 |
| A07 能力边界 | Responses none/auto 与 Chat auto；未接入强制工具、summary/tier/TTL/strict/budget 被明确拒绝；具名选项、sampling、OnPayload 同样校验 | R05、R08、R10–R15 |
| A08 采样与预算 | 通用/分档/调用/回调顺序，canonical 与 wire 分离，Messages 忽略采样 map；GLM 输出预算 1、有效上限及越界；不隐藏抬到 OpenAI 下限 | R07–R08 |
| A09 用量与价格 | 正常/零/缺项/中断/缓存/推理明细；不双计；reported 与 zero 区别；两型号费率、固定 RMB→USD 快照与价格阶梯；估算非账单 | R17 |
| A10 错误与重试 | 401/403、参数错误、可重试与不可恢复 429、5xx、流内错误；正确 Code/Phase/stop/id；默认零重试、显式有界策略、流开始后零重放 | R18 |
| A11 限额与释放 | 首帧前/后取消，截止时间、拒绝 Admission、超大 body/帧/输出、队列耗尽、回调异常；唯一终态和可独立等待 Result，许可/body/goroutine 释放 | R19 |
| A12 隔离与不可变 | 两租户同名 Bindings/ModelID/session、不同账户/协议并发；可信/不可信/跨账户原生状态；目录/选项/请求/结果 map 修改不串扰；Observer 不泄露内容 | R02、R15–R17、R19 |
| A13 回调顺序 | header→payload→response 顺序；最终模型/Key/目标/原生资源授权不变；合法自定义可用；回调失败零无效请求，重试复用冻结 payload | R15 |
| A14 回归与差分 | 两种 DeepSeek、OpenAI 两协议、原生 Anthropic 的受影响请求和事件保持；冻结基线不变；GLM 逐路由差分/extension 正确且零未解释 pending | R05、R20 |
| A15 正式发布 | 两型号的三个组合自身 live，正式目录 Hash、矩阵独立合并、全部既有门禁、脱敏审计；窄化报告、缺 Key、环境失败不能冒充完整通过 | R06、R21–R22 |

### T03 — 真实协议缺口的收敛规则

候选计划按型号与协议分别核验：Responses effort 冲突、对象/数组和无状态工具回放；Chat temperature=0、include_usage 对照、工具参数流和 reasoning_content；Messages 的 thinking/budget/effort、工具块、签名、usage、version/default headers。先登记假设、预期、最大调用次数及输出预算；一项只改变一个待确认变量，保留失败，不以无限试参、换账号或换协议“直到成功”。

默认不主动触发欠费、平台限流、套餐到期、超长上下文或不必要的认证失败；这些使用来源明确的离线 fixture，真实运行自然出现则记录。若必交能力经预算内核验仍无法实现，保留 FAIL/未完成与具体证据，不静默缩小本 spec；可选未接入能力无需逐一收费探测。

### T04 — 可验证、可重复的证据产物

每个 E2E 场景产出已脱敏的 fixture 与来源、最终请求、脚本化响应/SSE、事件序列、Result、Attempt/Usage、错误和断言；资源场景还记录释放与限额结果。证据有固定场景 ID、运行命令、SDK/目录/提交版本和校验索引；离线重放能复现相同语义，不比较不可复现的墙钟数值。

真实冒烟额外保存实际 wire 捕获、账户/区域与 Key 类别别名、型号、耗时、预算/实际尝试、费用估算、request id 来源、完整报告与逐包脱敏审计。捕获截断必须标明，不能用不完整 capture 声称逐字节验证。公开案例只用合成任务和低权限账户，不含真实业务资料；一协议或一型号的产物不能替代其他条目。

## Out of Scope

- GLM 自有协议、旧式文本 Completions、国际站或第三方代理服务。
- Coding Plan 的正式接入认证、套餐资格/额度探测、账号登录、自动续费、账单与计费后台；套餐端点不计为标准端点验收通过。
- 图像/音视频/文件输入的新支持、图像生成、分类、托管 Agent、Agent SDK 或上层 Agent/TUI 流程。
- Responses 服务端会话存储、previous_response_id 续接、查询、删除、cancel 和 background 作业；本地 context 取消仍在范围。
- namespace/custom/web_search、托管工具、容器、MCP、服务端模型 fallback、工具搜索和严格 JSON Schema 约束采样的新接入。
- Chat 跨 turn Preserved Thinking 的新公共配置、自动上下文裁剪、思考摘要或重排；必要的同 turn 工具推理回放在范围。
- Anthropic 专有缓存 TTL、managed effort、原生中途工具变更和 speculative beta 的兼容承诺；已确认对基础 GLM 运行必需的 wire 不以此排除。
- 全局能力插件注册中心、新模型协议框架、第二套 SDK/调用生命周期、自动跨协议或跨账号 fallback、实时汇率抓取。
- 升级冻结 pi 基线、修改已有 Provider 的公开语义或取消既有发布门禁。

## Further Notes

本规范采用研究日期为 2026-10-17 的官方资料快照。后续官方变更先记录来源、日期和影响范围，再更新对应 fixture、目录/能力与 ADR；不能用更新后的网页假称历史探针已经验证。

| 尚需实施证据确认的事实 | 本 spec 已确定的处理方式 |
| --- | --- |
| Responses GLM-5.3 effort 与模型页不同 | R07 限定首期子集，记录真实映射；关闭思考/别名未证明前不开放，不把 Chat 限制当作 Responses 事实。 |
| Messages 缺字段级兼容表 | R13–R14/T03 在同一预算合同下建立 GLM 自身 wire 证据；缺文档不代表拒绝，文本成功不代表整协议交付。 |
| Responses 图片示例与 text-only 型号冲突 | R06 采用 text-only 型号能力，不增加视觉承诺。 |
| temperature=0、流末 usage 和 SDK 默认头 | 按 T03 做有界对照，确定适配事实并验证最终 Client 请求；不能只依据单页文档或 SDK 示例。 |
| 历史套餐账号限制与新 Key 公告有时差 | R03 以标准 Key 正式验收；历史账号和套餐扣费不构成本次通过声明。 |
| 1M/128K、人民币报价与精确目录值 | R06/R17 在候选纳入前固定可追溯的容量与美元换算快照；元数据缺证据不以零或臆测替代。 |

上述是实施中的证据任务，不是要求用户补充业务需求。ready-for-agent 表示可据此开始实施和核验，不表示未知事实已成立。公共 Client E2E 测试接缝已获用户确认；该确认不改变表中尚需取得的厂商协议证据。

主要证据入口：[研究及本地解析探针](research.md)、[官方 Responses 指南](https://docs.bigmodel.cn/cn/guide/develop/responses/introduction)、[OpenAI 兼容指南](https://docs.bigmodel.cn/cn/guide/develop/openai/introduction)、[Claude 兼容指南](https://docs.bigmodel.cn/cn/guide/develop/claude/introduction)、[2026-10-16 Key 公告](https://docs.bigmodel.cn/cn/coding-plan/notice/api-key-management)。实施工单后续按本地 issue tracker 规则单独拆分；本次不创建实施工单或修改运行时代码。
