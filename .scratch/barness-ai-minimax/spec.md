# barness-ai 接入 MiniMax：Responses、Chat Completions 与 Messages

Status: claimed
Type: feature
Updated: 2026-10-18

本文件是交付契约，本身不表示 MiniMax 接入或真实冒烟已通过。2026-10-18 用户明确将接入目标改为国际站；当前依据为 [国际站研究](research-international.md)，[中国区研究](research.md)及原始失败捕获作为历史保留。官方声明、文档冲突与待实测项分别记录，不将其他地区或厂商的行为当作国际站事实。

当前交付为PARTIAL，见[交付状态](delivery-status.md)。国际站M3三候选与三份自身正式已PASS；M3.1无M Plan权益仍为线上三项NOT_RUN0/0，离线约束必交。用户新增区域路由验证后，本轮18个实际进程已全部关闭：其余15项collection及GLM Chat/Messages通过，GLM Responses为19PASS/3FAIL。新增245 calls/245 Attempts，连本次MiniMax续期历史300/300累计545/545；原失败与消费不改写，未用额度不转移。17身份默认矩阵原子导入16PASS/1FAIL。来源8fbd170的完整公共门禁已完成双vet/ordinary/race，数字退出0/0/1/1，两套各12439PASS/8FAIL/1NOT_RUN、审计0，报告pass=false。严格GLM三路raw/views续期仍拒绝；完整门禁暴露的固定映射、可选诊断注册和测试夹具inventory缺口由[票14](issues/14-release-preflight-gaps.md)离线补正，尚不授予修正后完整认证。10/09和spec保持开放；后续来源适用与最小重验范围独立复核，尚无新增付费批准。此前中国区401及全部历史仍保留；[工单11](issues/11-international-target.md)不重置预算或授予套餐资格。

领域与架构依据：[领域词汇](../../GLOSSARY.md)、[协议边界 ADR-0001](../../docs/adr/0001-barness-ai-protocol-boundary.md)、[DeepSeek Responses ADR-0014](../../docs/adr/0014-barness-ai-deepseek-responses.md)、[DeepSeek Chat ADR-0015](../../docs/adr/0015-barness-ai-deepseek-chat-completions.md)、[型号纳入 ADR-0018](../../docs/adr/0018-barness-ai-catalog-inclusion-criterion.md)、[Provider 自动字段 ADR-0032](../../docs/adr/0032-barness-ai-provider-automatic-fields.md)、[GLM 接入 spec](../barness-ai-glm/spec.md)、[GLM 正式支持 ADR-0036](../../docs/adr/0036-barness-ai-zhipu-formal-catalog-evidence.md)。

## Problem Statement

Host 希望在本地和云端通过 barness-ai 调用 MiniMax，自主选择 Responses、Chat Completions 或 Messages，保留已有 Client、消息、流事件、工具执行循环、用量与租户隔离契约。当前模块已有三套协议 adapter、DeepSeek 的独立 Provider 配置及 GLM 的兼容边界，但没有 MiniMax 的正式目录和自身支持证据。

仅替换 endpoint 或把 MiniMax 标成 OpenAI/Anthropic，会沿用错误的自动字段、思考默认、工具策略、缓存期限和计价。M3 在 Chat 默认思考，在 Responses/Messages 默认不思考；M3.1 强制思考且五档深度可调。M3 的 Chat 推理字段有官方资料冲突，Responses 的推理输入/输出 schema 不对称，Messages 的工具选择也小于原厂能力。有效文本响应不能证明工具历史能够续接。

Host 还需区分地区账户、按量与 M Plan 凭据用途、型号实际可用性和估算成本。M3.1 暂限 M Plan/MiniMax Code，旧 M2.x 正在迁移；自动换 Key、换协议或继续以旧型号价格估算会使授权、排障和费用归属失真。

## Solution

新增国际站 `minimax` Provider；沿用 `openai-responses`、`openai-completions`、`anthropic-messages` 三种 API Protocol，各有明确 Binding 和独立 Model 条目，共用现有协议 adapter、调用生命周期、SDK 传输、资源限制与可信续接边界。Host 负责认证授权、凭据、Transcript 和工具执行；模块一次处理一个 Generation Turn，不拥有会话或自动执行工具。

首期正式型号为 `MiniMax-M3`，三个协议身份均必交；同时为 `MiniMax-M3.1-Flash-Preview` 的三个候选身份实现并离线验证已知型号约束。首期交付文本输入/输出、流式消费、函数 Tool Round Trip、可观察推理及合法原生历史；多模态输入不在首期子集。M3 以国际站标准按量配置正式验收，M3.1 仅在有资格的 M Plan 配置中单独候选验收，二者不互相证明。当前研究尚无 M3.1 可直接用于美元 token 估算的公开费率；候选不能因此虚构价格或进入正式目录，也不阻塞 M3 的三协议交付。

保留公共调用形状，使用完整历史，不自动转换协议。首期只支持 standard 服务层级、自动缓存及 Responses 已公开的隔离缓存路由键；priority、Messages 主动缓存等尚未完整表达费用或期限的能力明确列为未接入。交付包含失败场景先行的公共 Client E2E、M3 三身份自己的正式真实证据、M3.1 候选约束和状态、目录/价格快照、差分或扩展登记、使用文档、可回滚变更和完整发布门禁。

## User Stories

以下编号对应稳定追溯标识 US01–US80；“每个型号”的能力约束同时覆盖正式 M3 与候选 M3.1，正式发布范围依 R06/R27。

1. As a Host developer, I want MiniMax to have its own Provider identity, so that routing identifies the actual service operator.
2. As a Host developer, I want Responses through the existing Client, so that I can keep my Responses integration.
3. As a Host developer, I want Chat Completions through the existing Client, so that I can keep my Chat integration.
4. As a Host developer, I want Messages through the existing Client, so that I can keep my Messages integration.
5. As a Host developer, I want Stream and Complete on every route, so that delivery style does not change supported behavior.
6. As a Host developer, I want StreamSimple and CompleteSimple on every route, so that protocol-neutral callers can use MiniMax.
7. As a Host developer, I want existing protocol options and hooks retained, so that I do not need a parallel MiniMax request API.
8. As a Host developer, I want incorrect protocol options rejected, so that a Binding change cannot reinterpret my request.
9. As a tenant administrator, I want separate protocol Bindings, so that authorization is explicit for every route.
10. As a tenant administrator, I want denied models rejected before credential access, so that unauthorized requests cannot read secrets.
11. As a tenant administrator, I want immutable Binding and Credential Snapshots, so that retries cannot change account or purpose.
12. As a tenant administrator, I want shared Clients to isolate tenant accounts, so that concurrent requests cannot mix credentials.
13. As a Host operator, I want China and international identities distinguished, so that regional service assumptions are not mixed.
14. As a Host operator, I want standard and M Plan credentials configured explicitly, so that quota purpose is known before a call.
15. As a Host operator, I want missing model entitlement reported honestly, so that another model's success cannot certify availability.
16. As a Host developer, I want independent model entries per protocol, so that one protocol's defaults cannot alter another's behavior.
17. As a Host developer, I want M3 and M3.1 configured independently, so that their reasoning constraints remain accurate.
18. As a Host developer, I want the first release's text-only subset stated clearly, so that upstream vision support is not mistaken for implemented input support.
19. As a Host developer, I want omitted reasoning to keep each protocol's documented default, so that absence is not reinterpreted as disabled thinking.
20. As a Host developer, I want M3 Chat thinking enabled by default, so that I retain the endpoint's intended behavior.
21. As a Host developer, I want M3 Responses and Messages thinking disabled by default, so that enabling reasoning remains an explicit choice.
22. As a Host developer, I want M3.1's mandatory thinking enforced, so that no entry point silently sends an invalid off request.
23. As a Host developer, I want M3.1's five supported depths, so that I can select the intended reasoning effort.
24. As a Host developer, I want M3 exposed as a thinking switch rather than artificial depth levels, so that controls do what they claim.
25. As a Host developer, I want full unsupported levels rejected and simple levels resolved predictably, so that reasoning selection is understandable.
26. As a Host developer, I want canonical levels separate from wire fields, so that sampling selection does not masquerade as measured model intensity.
27. As a Host developer, I want reasoning separated from visible answers where supported, so that my UI can present both correctly.
28. As a Host developer, I want M3's actual reasoning format preserved, so that conflicting documentation cannot corrupt history.
29. As a Host developer, I want tools on all three protocols, so that the Host can continue its existing tool loop.
30. As a Host developer, I want fragmented tool arguments assembled once, so that progress is observable without executing incomplete input.
31. As a Host developer, I want multiple tool calls kept distinct, so that arguments and results cannot be paired with the wrong call.
32. As a Host developer, I want reasoning and tool calls replayed together, so that subsequent generations retain required context.
33. As a Host developer, I want unsupported forced tool selection rejected, so that requested constraints are not weakened to auto.
34. As a Host developer, I want ignored parallel controls rejected when I depend on them, so that successful calls cannot imply an unenforced guarantee.
35. As a Host developer, I want complete Responses history replay, so that I do not depend on provider-side saved state.
36. As a Host developer, I want Responses input and output reasoning shapes checked independently, so that SDK decoding does not imply valid replay.
37. As a Host developer, I want authoritative final output reconciled with deltas, so that text and reasoning are neither lost nor duplicated.
38. As a Host developer, I want Messages thinking signatures retained, so that native continuation follows the provider contract.
39. As a tenant administrator, I want persisted Native State trusted only by the Host, so that client-supplied signatures do not grant replay authority.
40. As a tenant administrator, I want cross-account and cross-protocol history downgraded safely, so that account-specific state cannot cross a boundary.
41. As a Host developer, I want tool identifiers treated as opaque values, so that punctuation and long IDs cannot break pairing.
42. As a Host developer, I want protocol-specific output budgets, so that invalid fields or hidden budget inflation cannot change my request.
43. As a Host developer, I want reasoning-only truncation reported as truncation, so that an empty answer is not fabricated into success.
44. As a Host developer, I want final usage read after content stops, so that late usage chunks are not discarded.
45. As a Host developer, I want missing usage distinguished from reported zero, so that unknown consumption is not shown as free.
46. As a Host developer, I want cache and reasoning tokens normalized by protocol, so that estimates do not double count them.
47. As a Host developer, I want automatic cache usage retained, so that available cache savings remain observable.
48. As a tenant administrator, I want cache routing keys scoped to tenant and account, so that raw session identifiers are not shared.
49. As a Host developer, I want unsupported cache lifetimes described explicitly, so that a routing hint does not imply a retention guarantee.
50. As a Host operator, I want versioned international USD prices, so that another region's rates or exchange conversions cannot be applied to my account.
51. As a Host operator, I want long-context price tiers tested, so that estimates remain correct at threshold boundaries.
52. As a Host operator, I want unsupported priority requests rejected, so that an unpriced service tier cannot create misleading estimates.
53. As a Host operator, I want subscription usage distinguished from estimated standard API cost, so that estimates are not presented as M Plan deductions.
54. As a Host developer, I want sampling and callbacks applied in the existing order, so that request customization remains predictable.
55. As a tenant administrator, I want final requests revalidated after hooks, so that native payloads cannot bypass model or authorization constraints.
56. As a Host developer, I want immutable request snapshots, so that later map mutations cannot change an active call.
57. As a Host developer, I want unsupported stop and format controls rejected, so that ignored fields cannot imply enforced behavior.
58. As a Host operator, I want default zero retries and explicit bounded initial retries, so that duplicate generation is controlled.
59. As a Host operator, I want permanent quota failures distinguished from transient throttling, so that exhausted plans are not retried indefinitely.
60. As a Host operator, I want no replay once streaming starts, so that partial responses are not silently generated again.
61. As a Host developer, I want cancellation and deadlines to settle one Result, so that my application can release work reliably.
62. As a Host operator, I want existing byte, queue and concurrency limits enforced, so that MiniMax cannot bypass resource policy.
63. As a Host developer, I want malformed streams and premature EOF reported as failures, so that partial output cannot appear complete.
64. As a Host developer, I want provider stop reasons and available trace IDs retained, so that failures remain diagnosable.
65. As a tenant administrator, I want metadata-only observation and redacted artifacts, so that keys and conversation content remain private.
66. As a maintainer, I want one public Client E2E seam, so that acceptance exercises the complete caller-visible behavior.
67. As a maintainer, I want failure scenarios recorded before implementation, so that tests drive the supported behavior.
68. As a maintainer, I want replayable request, stream and result artifacts, so that acceptance can be checked without new paid calls.
69. As a maintainer, I want each model and protocol to produce its own live evidence, so that shared adapters cannot substitute for independent proof.
70. As a maintainer, I want bounded experiments for documentation conflicts, so that unknown behavior is resolved with attributable evidence.
71. As a maintainer, I want fixed pi differential or extension records, so that compatibility claims reflect the actual frozen route.
72. As a maintainer, I want M3.1 marked as outside the frozen model catalog, so that new behavior cannot be certified by a fabricated baseline.
73. As a maintainer, I want DeepSeek and GLM regressions covered, so that capability refactoring preserves their established contracts.
74. As a maintainer, I want controlled captures distinguished from live execution, so that replay cannot be reported as fresh support.
75. As a maintainer, I want source and catalog applicability checked, so that stale captures cannot certify changed execution code.
76. As a maintainer, I want additive release requirements, so that MiniMax cannot weaken existing providers' release gates.
77. As a Host developer, I want complete local and cloud configuration examples, so that both deployments use the same authorization contract.
78. As a Host operator, I want reversible configuration and software releases, so that MiniMax can be disabled without silent rerouting.
79. As a maintainer, I want architecture decisions and evidence limits recorded, so that future changes preserve their context.
80. As a maintainer, I want all six formal or candidate identities accounted for at delivery, so that missing access cannot be hidden by narrowing the report.

## Implementation Decisions

### R01 — Provider、操作与协议边界

新增 `ProviderMiniMax`，值 `minimax`，表示国际站服务；Operation 为 chat。API 保留 `openai-responses`、`openai-completions`、`anthropic-messages`；Completions 专指 Chat Completions，不包含旧文本 Completions。复用现有三套 adapter，不增加 MiniMax 专有协议、代理服务或第二套 Client/assembler/重试生命周期。此前尚未激活的中国区运行时路径在本次切换中删除，不新增别名或兼容shim。中国区历史仅用于来源与预算审计，不按 URL 猜测 Provider 或自动换区。

### R02 — Binding、端点和鉴权

每个 Binding 固定一种 Operation × Provider × API；同一型号的三种协议分别授权。正式国际站基址与最终资源路径如下；SDK 自动补路径、查询参数和版本头须由捕获验证，不通过重复拼接版本段适配。

| API | Binding.Endpoint | 生成请求地址 | 本期鉴权 |
| --- | --- | --- | --- |
| openai-responses | `https://api.minimax.io/v1` | `POST https://api.minimax.io/v1/responses` | Bearer API Key |
| openai-completions | `https://api.minimax.io/v1` | `POST https://api.minimax.io/v1/chat/completions` | Bearer API Key |
| anthropic-messages | `https://api.minimax.io/anthropic` | `POST https://api.minimax.io/anthropic/v1/messages` | x-api-key |

Messages 沿用当前 SDK 的 x-api-key 路径，这是官方明确接受的鉴权方式；首期不新增可选双鉴权或通用 AuthKind。最终请求不得同时带 Bearer 与 x-api-key，尤其不得依赖上游 Bearer 优先级选择另一把 Key。保留可信 header 回调与认证字段保护，不能利用回调追加第二套认证。正式请求使用本表地址；冻结 pi 的国际站原始配置须独立装配并核对，实际差异据实登记。

### R03 — 账户、凭据用途与型号资格

核心只使用 Host 解析出的 Binding 和 Credential Snapshot，不读取环境 Key、SDK profile 或进程默认凭据。按量与 M Plan 使用不同 Binding/凭据用途和适当 Account Scope；不从 Key 字符串猜测权益，不新增套餐识别、购买或余额查询。三协议使用相同普通兼容基址，不发明 Coding 专用地址。

同租户、同账户、同用途且有明确授权的协议 Binding 可引用同一 CredentialRef，但每次 Logical Call 仍独立解析并冻结版本。M3.1 仅在实际有资格的配置中验收，不能把本地交互套餐当成云端生产按量资格。错误不触发换 Key、套餐转余额、换型号或跨区。缺少合法测试配置时记录 NOT_RUN，不改变正式交付目标。

### R04 — 公共接口、选项与职责

保留 Stream、Complete、StreamSimple、CompleteSimple、WithHooks 和公共 Request/Transcript/消息/工具/事件/Result 契约。full 分别使用 ResponsesOptions、ChatOptions、AnthropicOptions；错误协议选项拒绝。nil/full/simple 的缺省语义依 R07，不被原厂默认覆盖。Complete 可继续消费相同流式实现，不新增必须支持非流式 HTTP reader 的要求。

Host 执行工具并自行发起下一轮；ValidateToolCall 和工具结果辅助能力继续使用。新增型号元数据仅表达已出现的真实语义，零值不改变旧 Provider；参与复制、序列化、目录版本和 Hash。不以型号 ID 分支隐藏硬约束，也不引入 Host 可写的 Provider 能力注册中心。

### R05 — Provider 协议能力

MiniMax 在三套内部能力配置中均须有显式分支，不能落入未知 Provider 的原厂全能力默认。按真实差异拆分 Chat thinking 的 adaptive/disabled 形状、Messages adaptive/effort 与原厂 display/beta、缓存路由与期限等被捆绑的能力。Provider × API 的固定 wire 事实归协议能力，型号只进一步收紧；Host Catalog 不能打开未实现的服务端能力。

不复制 DeepSeek 的 enabled 开关、assistant 补空 reasoning_content、GLM 的 tool_stream、单对象 reasoning 规则或特殊错误体。已有共享能力只有经 MiniMax 证据吻合才复用。内部重构直接删除过时字段和实现，不留 shim、双写或两个活跃兼容分支；旧 Provider 的可观察请求、事件和错误保持既有契约。

### R06 — 六个型号身份与目录

M3、M3.1 各有三个独立条目，不能从一个协议条目自动派生另两个。首期统一只声明 barness 已接入的文本/工具子集；这不表示厂商型号没有视觉能力。图片输入沿既有未声明图片能力时的占位/历史转换契约处理；回调不得绕过文本子集发送多模态容器。冻结 pi 的图片能力与 resize 策略差异登记为扩展，不默称完整模型数据等价。

候选上下文可固定为官方明确的 1000000；输出预算先采用不超过 131072 的保守目录上限，记录其是 barness 首期上限，未声称覆盖厂商最大 524288。Chat/Messages 的公开字段与 Responses 未给精确范围的事实分别记录，不能把一个协议的预算验收移作另一协议证明。元数据、型号价格和来源先在候选 Catalog 固定；M3 通过硬约束与自身候选证据后进入 BuiltinCatalog，正式报告必须再绑定正式目录 Hash。不得删除失败的 M3 协议或以另一型号代替以完成本 spec。

M3.1 的三候选条目仅用于明确标记的受控/候选 Host Catalog 和离线约束验收，不进入 BuiltinCatalog、正式 Support Matrix 或首期必需 live 清单。候选可用有标识的合成费率测试机制，但不得展示为真实 M Plan 估价；真实候选报告单列原始 token 用量及价格未获依据。取得合法账户和有依据价格后，仍须独立正式验收与显式扩展发布合同才能推广，不自动因 M3 通过而转正。缺资格为 NOT_RUN，不能改成 PASS/UNSUPPORTED；本期不扩展公共费用 DTO 来伪装未知费率。

### R07 — 思考默认、开关和真实深度

省略思考参数时，三种公共入口风格均保留该型号/协议的服务默认：M3 Chat 开启，M3 Responses/Messages 关闭，M3.1 三协议均开启且不可关闭。不得将 simple 未指定 Reasoning 一律翻译为 disabled/none，也不把省略冒充用户选择的某个 canonical 档位。

M3 只有开关语义；canonical high 作为“开启”的唯一声明档位，不承诺 low/high 不同计算深度，其他非 off 档位不在 full 的直接支持集合。Chat 显式开启用 adaptive、关闭用 disabled；Responses 开启用非 none 的已验证 effort high，关闭用 none；Messages 开启用 adaptive、关闭用 disabled。Chat/Responses 的显式关闭通过各自既有原生采样或可信 payload 入口表达，Messages 使用 ThinkingEnabled；不增加 simple 的 off/required 公共枚举。M3 Messages 的 Effort 显式值属于未接入深度控制，应拒绝，不能接受后忽略。

M3.1 canonical low/medium/high/xhigh/max 一一映射至该协议 effort，minimal/off 不支持；省略时保留服务默认 max。RequiresThinking 约束在全部入口和最终 wire 生效，不能仅把 Off 标 null 后静默吃掉显式关闭。Chat/Responses full 不支持档位在授权后、读凭据前拒绝；simple 沿用向上优先夹取，对 M3 夹到 high、M3.1 minimal 夹到 low，且结果可追溯。Messages 使用相同型号声明集合解析 simple，不套用 Claude budget 算法；full Effort 不支持时拒绝。M3.1 ThinkingEnabled=false、effort none 或最终回调关闭均 invalid_request，不隐式重开后假称按原请求执行。

### R08 — 采样、空值和输出预算

OpenAI 两兼容协议保留通用模型默认 → 有效 canonical 分档默认 → 调用级覆盖 → OnPayload 的既有顺序，回调后不重新选档；Messages 继续忽略 SamplingParams 和分档 map，合法原生字段经已有选项或安全 payload 表达。canonical 分档选择与 wire effort 分离，最终 wire 必须满足硬约束，不能借采样或回调绕过。省略 Reasoning 时，MiniMax 的分档默认选择使用其已声明的有效默认标签：M3 Chat 为 high（仅表示开关开启）、M3 Responses/Messages 为 off、M3.1 为 max；wire 仍省略未要求的思考字段，不把默认标签记录成用户显式档位。旧 Provider 的省略选档规则保持。默认不强加厂商推荐采样值；如候选目录设默认，须有来源并固定版本。

Chat/Messages temperature 的接入范围是 [0,2]、top_p 是 [0,1]；Responses 首期采用无冲突的 (0,1]，零值在文档冲突未收敛前明确拒绝，不能沿用 Chat 最大 2。null/省略遵循现有具名选项 presence，再按最终 wire 校验，不把未知 null 解释成有效数字。

Chat 只自动生成 max_completion_tokens，Responses 为 max_output_tokens，Messages 为 max_tokens。最终预算是精确整数 1..型号上限，不偷偷抬到 OpenAI 的 16 下限；小预算即便只产思考也保留原预算。沿用既有具名选项的 null/0 省略规则，但最终 JSON 中的零、null、分数、负值和越界必须拒绝，Messages 具名 MaxTokens=0 同样非法。Chat 不同时发送两个长度字段；旧 max_tokens 原生形状不在首期产品子集。默认/采样/回调后的校验保留精确数字，不能 float 舍入后放行分数。

### R09 — Chat 请求和工具策略

使用 system/user/assistant/tool 消息、嵌套 function 工具和 stream_options.include_usage=true。声明推理分离格式，正式默认 reasoning_split=true；不自动发送 store、developer、缓存路由/保留期、strict、tool_stream、旧 functions/function_call 或其他厂商字段。

Chat 首期只支持默认函数选择；公开 auto 作为相同默认语义，构建时省略未文档化 tool_choice 字段。安全回调若提供 tool_choice，只接受有自身证据的 auto；none、required、具名函数、allowed_tools、并行控制在首期明确拒绝，不静默变成默认。若候选证明连 auto 字段也不接受，继续省略并拒绝原生字段，不影响公共 auto 的默认语义。正文 JSON Schema、严格函数约束、response_format、seed 和 n≠1 不纳入子集。

### R10 — Chat 推理格式和工具历史

M3 的 reasoning_content 与 reasoning_details 文档冲突须在候选阶段按实际型号、端点和 Key 用途收敛。优先复用已有 reasoning_content reader 与 reasoning_details 的 typed entries/签名处理，不重新建立 reader；后者若形状不同，在协议边界最小转换并单独保留原生状态。不得压平原生数组后称为原样回放，不向所有 assistant 强行补 DeepSeek 的空字段。

2026-10-18国际站M3自身研究已观察到同一delta同时包含reasoning_content与reasoning_details，后者为reasoning.text/MiniMax-response-v1条目，拼接文本与前者完全一致。这种已证实的镜像只显示一次，保留原生明细及镜像归属以支持持久化后可信续接；只有严格相等的已集成镜像可接受，值矛盾或未集成跨帧格式切换仍失败。原始失败捕获按38b7098来源保留，不能在修复后改写成PASS。

正式默认采用已验证的分离格式；若 M3 只能以原始带 think 标签的 content 正确续接，必须记录这个格式限制、公开内容语义及差分决定后再验收，不能临时剥标签或伪造独立 Thinking。M3.1 不允许 reasoning_split=false。未验证的同协议格式变体明确拒绝，不能默许回调破坏 Native State 的可回放性。

工具往返保留该轮完整 assistant 正文、推理、调用和结果顺序。流中多个调用按实际 index/ID 分片独立组装，空 content/null assistant、并行调用、只在 done 有完整 arguments 均按实际 wire 验收。只有完整 ToolCall 交给 Host 验证执行；不在库内执行函数，不合并不同调用，不丢弃同轮思考。工具已返回但 finish_reason 语义冲突时不能当普通回答静默通过。

### R11 — Chat 流、终态和用量尾帧

支持正文、分离推理与工具参数的实际增量，权威完成值不能丢失或重复。finish_reason 的 stop/length/tool_calls/content_filter 依既有停止/错误合同映射，原始原因保留；未知原因不当作普通 stop。finish_reason 出现后继续读取合法 usage-only chunk，兼容 usage 与 finish 同帧的情形。

候选必须捕获完整 SSE 包括哨兵/EOF。是否出现 [DONE] 以自身证据确定；有有效 finish 且干净结束可按 Chat 既有生命周期完成，无有效 finish 的 EOF、无效 JSON 或 HTTP 200 流内 error 均失败。不能以人工 SDK 循环打印结束代替 raw wire 证明。

### R12 — Responses 请求子集

使用完整 input 历史、适用 instructions/messages、函数调用/结果、已验证 reasoning、max_output_tokens、metadata 与合法采样；函数工具为顶层 function 结构。支持默认/auto/none，required、具名强制、allowed_tools、strict、托管/namespace/custom 工具请求前拒绝，即使 Binding 允许某种 Hosted Tool 也不代表已接入该能力。

store、parallel_tool_calls、truncation 在官方端点被忽略，核心均不派生。首期原生回调同样不接受它们，不能用 store=false 自称请求保证了不存储，也不能用 parallel_tool_calls=false 宣称串行。默认使用厂商固定的无保存行为；explicit store=true 继续先受账户资源授权约束。previous_response_id、conversation、background、retrieve/delete/cancel、encrypted include、reasoning.summary 和非 text 输出格式不在子集，显式请求拒绝，不把忽略当满足。

### R13 — Responses SSE 与权威输出

复用现有 Responses 正文、reasoning_text、function arguments、output item 和终态处理；MiniMax 示例的 reasoning.content 是数组，不能照搬 GLM 的单对象修复。只有真实 schema 或捕获证明确有其他形状时才按边界结构转换，并保留原始数据，不通过 transport 改写或字符串替换掩盖 SDK 丢字段。

候选记录实际 created/item/content/delta/done/completed/incomplete/failed/error 事件。权威 done、只有 done、只有终态输出、混合多 item 的合法形状都须正确归一化，缺必要关联或畸形内容明确失败。completed 不能凭 HTTP 200/EOF 合成；incomplete.max_output_tokens 保留 length，包括只有 reasoning 没有 message；content_filter/failed/error 保留错误原因，不能生造正文。重复终态或非法顺序不得重复结算/重复块；缺有效终态的 EOF 必须失败，不能依赖 [DONE] 充当 Responses 终态。

### R14 — Responses 原生回放与工具 ID

reasoning 输出 content 数组与输入 summary 声明不对称，是必收敛的协议问题。先保存原始输出和来源，再验证完整 reasoning 条目的真实可接受输入；若需要输入转换，必须有官方/真实证据、保留原始 Native State，并证明工具结果后继续生成。不能将完整推理自行摘要化、借服务端保存绕过完整历史，或以另一型号回放成功代替本型号。 2026-10-18自身Responses输出还包含文档未列的字符串encrypted_content；该原始输出须完整保留。首期不据此开放密文输入：仅把已公开文本无损投影到文档summary形状，保持输入hook拒绝密文；后续完整候选必须实际证明工具续接，不能把本地编码修复当作服务已接受。仍无法续接时该身份未完成，不标 UNSUPPORTED 放行必交历史。

function_call 与 function_call_output 用真实 call_id 配对，条目 ID 与调用 ID 分开处理。合法 ID 视为不透明值，覆盖 `%`、`|`、Unicode、空/超长约束及碰撞；只有服务契约限制时才净化。公开组合 ID 必须可无歧义关联且避免截断碰撞；不复制 OpenAI 的 fc_ 前缀假设或 GLM 专用编码。跨身份转换同步改名结果并保留工具配对，旧 Provider 历史无关场景不改变。

### R15 — Messages 请求与 adaptive 思考

复用已有显式 SDK service、SSE reader、headers 和关闭的 SDK 重试；先核验最终资源路径、版本头、SDK query 和认证，不将 SDK 附带 beta 查询误认为支持原厂 beta 功能。消息用 user/assistant content blocks 与顶层 system；函数定义使用 input_schema，结果经 user.tool_result 回传。

MiniMax adaptive 开关与 output_config.effort 独立于原厂 display、budget、interleaved beta、block_binding、managed per-turn effort、eager streaming 和 browser-access header。只生成已确认的 adaptive/disabled 与型号有效 effort；不得因打开 adaptive 能力顺带发送后者。M3.1 full 只有 Effort 且未指定 ThinkingEnabled 时仍应表达该 effort，不能因默认开启便漏掉它。M3 ThinkingEnabled=true 可开启但不发送无效深度；正数 ThinkingBudgetTokens、ThinkingDisplay、显式 InterleavedThinking 和未接入原厂扩展拒绝，不接受后忽略。

显式 Temperature 属于 MiniMax 已接入参数，不因为 adaptive 思考开启就沿 Claude 的规则静默丢弃；最终范围与存在性依 R08 校验。如果真实候选显示组合限制，先形成有证据的明确拒绝契约再修改 spec，不能吞掉参数。空 ThinkingDisplay、零 ThinkingBudgetTokens 代表未请求，不触发原厂默认 summarized/1024 的派生。

工具选择仅默认/auto/none；any、具名 tool、disable_parallel_tool_use 不接入。top_k、stop_sequences、mcp_servers、context_management、container 虽官方说忽略，本子集明确拒绝显式字段，避免把停止、安全或资源限制降格为提示。Metadata 继续按既有公共映射，不能自动把 TenantID 当作 user_id 上送。

### R16 — Messages 流、签名与续接

处理自身 message_start、ping、content_block_start/delta/stop、message_delta、message_stop 与 error。工具 input_json_delta/partial_json 和开始块 input 组合须经自身 capture 核验，不能因公开 enum 缺项断言不支持，也不能照抄 Claude fixture。校验块索引和 delta 类型的关联，独立组装多工具参数；非法顺序、重复/缺失终态或提前 EOF 必须诊断，完成只结算一次。

同模型工具往返原样保留完整 thinking/text/tool_use，包括签名增量和最终签名。默认采用有签名可信回放；无签名、空签名、redacted thinking 分别核验。若真实目标型号合法返回无签名块且续接需要原块，只为 MiniMax 已证实形状增加同模型可信规则，并纳入 Native State Envelope 保护；不得直接开启 GLM 的例外或放宽原生 Anthropic。未证实的 opaque/redacted 形状不伪造、不静默当普通正文，保持原始结果并明确不支持原生回放的错误。

### R17 — 缓存与期限

首期支持厂商自动重复前缀缓存及实际报告的缓存用量；不承诺命中、隔离缓存存储或保留时间。Responses 有 SessionID 且 CacheRetention 非 none 时可使用既有租户/账户/session 派生的 prompt_cache_key；无 SessionID 或 none 不发。short/long 在这一路都只是路由提示，不派生 TTL、retention 或 OpenAI session 亲和头，保留既有缓存提示合同并写清差异。回调不得替换成原始/他租户缓存键。

Chat 不派生未公开的 cache key/retention。Messages 首期不自动生成 cache_control，沿用自动缓存；short/long 均不代表已请求显式缓存，说明其为本期未实现的提示。显式 cache_control/ttl 经回调出现时拒绝，不能默发原厂 1h，也不能默认启用写价不齐的 M3/M3.1 主动缓存。主动缓存 5m/4 断点有文档但不等于本期已实现。

### R18 — 标准服务层级与价格快照

三协议只支持省略或 standard 服务层级；显式 priority 在具名选项和最终 payload 均拒绝。当前 Responses 可表达 tier，但其原厂 priority 倍率不适用；Chat/ Messages 的透传也缺完整 tier 计价，不能只修一条路径后放开另外两条。1.5 倍是官方事实，本期不以透传方式提供它。

CostRates/UsageCost 继续是美元/百万 tokens 与美元金额。M3 三身份按有来源、日期与Hash的官方国际站美元价建立 Price Snapshot，不使用中国区费率或人民币汇率换算。标准输入/输出/缓存读费率在输入侧≤512000时为0.30/1.20/0.06 USD/M，超过时为0.60/2.40/0.12 USD/M。定价总览是当前费率依据，其他指南的旧价示例单独记录冲突；不得用合成fixture正式发布或缺价填0。官方价格页仅写512k，未明示512000或524288；512000是本项目保守估算策略，不冒充官方精确整数。未来取得可靠整数定义时再修订此快照；当前保留边界±1验证，不执行高容量收费探测。阶梯包含输入/输出/缓存读，输入侧包含缓存读写，不以上下文窗口或未缓存 input 单独判阶梯。M3.1 的正式订阅标准价/估算基准须单独有出处；未取得可靠费率时维持 R06 的候选边界，不能捏造按量售价或自动复制 M3 价格。

订阅 Usage 保留 token 消耗；Estimated Cost 仅为明确基准下的估算，不是 M Plan 积分、额度扣减或账单。若平台返回意外的实际型号，记录 ResponseModel，不能自动换配置型号、归属或费率；未知实际型号造成估算不可靠须在报告披露并阻止正式支持验收，不能为迁移旧名默用原价。

### R19 — Usage、缺失与中断

Chat/Responses 的缓存 tokens 是输入总量的子集，归一化 Input 排除另列 CacheRead/CacheWrite；Messages 的 input_tokens 是未缓存部分，总输入按 input+read+creation。Reasoning 属于 Output，不再次计入总量或费用。数据可共享转换机制，字段 presence 和完整性仍按本协议证据判断；缺项保留 partial/unreported，合法显式零才是已上报零，不通过猜字段或补 0 宣称 complete。

Chat 读取 finish 同帧或后续 usage-only chunk；Messages 合并 start 输入明细与 delta 的权威输出，用量更新不是盲目累加；Responses 按实际终态提供的字段。保留已合法获得的 Attempt usage；failed-response 是否接纳用量遵守现有 Responses 公开合同，不借本功能全局改规则。CacheWrite1h 不按原厂默认推导；缺失新型号写缓存费率不能借“未返回”猜免费。

### R20 — 最终请求校验、回调和不变性

保持 header → payload → response 顺序与 WithHooks，OnPayload 接收该协议原生请求。所有默认、模型/调用采样与回调之后统一校验最终 stream、模型、消息、函数工具、原生引用、认证、缓存键、型号约束、已接入能力与资源限制；无 hook 也走同样边界。只检查协议容器，函数 JSON Schema 属性名、arguments 或正文中的同名字段仍是业务数据，不能递归误拒。

绑定/操作/型号授权优先；具名非法能力在身份已授权后尽早拒绝，最终 payload 先做既有授权校验再做能力校验，保留 tenant_denied 等优先级。采样的中间非法值可由合法 OnPayload 修复，合法 full 选项不能靠后续覆盖掩饰未声明档位；最终协议预算精确检查在回调后执行。发送前失败零 Provider 请求、零 Attempt；发送后 hook 失败保留实际 Attempt 与已知用量。调用快照独立复制，最终合法字节冻结一次，所有初始重试复用同一份，不重新执行 hook 或重选凭据。

### R21 — Native State 与租户隔离

完整模型身份始终是 Operation × Provider × API × ModelID；共享名字不授予跨协议或账户的续接权。进程内可信 Result 可同模型续接，持久化读回需 Host 校验完整性并 TrustNativeState；客户端 JSON 不能携带信任。跨租户、账户、型号、API/Provider 或不可信状态依既有规则降级，保留可见文本与工具配对，剥离账户原生引用，记录降级元数据。

同租户已授权的跨模型转换不因旧消息来自其他账户整体拒绝；只检查最终仍保留的专有引用。两租户同名 BindingID/ModelID/session、不同凭据用途及并发重试不得串用缓存、原生状态、选项或 HTTP header。RequestID/AttemptID 的全局唯一 Host 义务保持；测试不靠重复 RequestID 制造伪隔离结论。

### R22 — 错误、请求 ID 与重试

沿用稳定 Error Code/Phase、HTTPStatus、RawStopReason 和可验证的 ProviderRequestID。分别解释各兼容协议的 HTTP/body/SSE error，不按 MiniMax 自有 base_resp 反推三种错误格式。已知参数错误、鉴权/权限、余额/套餐耗尽/权益不足不可恢复；可重试限流/过载仅按 Binding 显式有界策略在流开始前重试。未知错误保留已有通用分类，不仅凭消息含词猜业务码。

SDK 重试关闭、Binding 默认零重试；已知永久 quota 错误即使 HTTP 429 也不自动重试，未知 429 维持既有策略。任何已开始的流不重放，不换协议/账户/型号。HTTP 200 error、过滤、长度截断、网络失败分别保留原因和已知用量。只有实际返回的 trace_id/header 或协议 request_id 才作为厂商 ID；实际优先级和安全格式经捕获固定，不把 Host RequestID 或 DeepSeek/GLM 的头照抄成 MiniMax 证据。

### R23 — 资源、取消、观测与隐私

全部路径经过现有 Admission、并发/等待、请求/响应头/调用时限、输入/输出/帧/工具 JSON/事件队列限制；不以厂商最大模型容量覆盖 Host ResourcePolicy。取消、超时、断流、hook 异常和限额失败必须仅有一个终态，Result 可独立等待，body、许可、goroutine 与队列占用释放。取消只保证本地停止，不承诺远端停止生成或免计费。

复用既有协议资源套件覆盖 MiniMax 三路；在本地/云端既有 Design Load 下加入与旧 Provider 并发的 MiniMax 调用，保持原策略、原负载和每项限额占用不超过 75% 的验收标准，记录峰值占用、内存依据与资源归零。不能靠减少旧负载、新增无依据的宽松策略或提升限额解决回归。

Observer 继续只含既有授权元数据和调用归属，不加入 prompt、推理、工具参数、Native State、原始错误正文、Key 或敏感账户信息。部署地域/凭据用途由 Host 使用非秘密别名管理，不能为观测扩展不必要的业务数据。证据仅保存合成/获准测试输入与脱敏捕获，逐包 Redaction Audit 失败即验收失败，不在错误中回显敏感正文。

### R24 — 冻结 pi、差分与扩展

保持唯一 pi-ai 1.1.0 版本、commit、完整模型数据与源码 pins，不升级或修改冻结包。该基线完整目录包含国际站 minimax × Messages 的M3/M2.7数据；保持包与完整数据pins，仅将oracle装配改为已有MINIMAX_MODELS；M3 在本期文本/工具行为适用的场景进入真实差分，当前端点、图像子集、默认字段或回放差异逐 case/path 登记。不能因上游有 Provider 就把尚未运行的差分标 PASS。

MiniMax Chat/Responses 是冻结 Provider 数据没有的路线，登记 Extension Route、自身 fixture 与明确 pi skip；M3.1 不在冻结型号目录，单独登记型号扩展，不套 M3 型号冒充直接差分。差异只限实际受影响字段和场景，不豁免整个协议/结果/请求体，未解释项保持 pending。OpenAI/Anthropic/DeepSeek/GLM 共享 adapter 的通过不证明 MiniMax，既有差分与扩展覆盖不得减少。

### R25 — 文档冲突的候选收敛

官方文档未列字段表示未获契约，不等于已证实拒绝；本 spec 的拒绝集合是明确的首期产品边界。候选探索围绕预先登记的问题，以官方快照与型号/端点/凭据用途自己的原始请求响应记录选择实际支持形状，再先写对应失败 E2E，最后实现。不能把未知散落成无上下文 TODO，不能在失败后重写 fixture 或删验收项让它通过。

M3 正式三路径必须收敛：Chat 分离推理字段与完整历史；工具增量/调用配对；Responses SSE 终态与原生 reasoning 输入；Messages thinking 签名与实际块序；晚到 usage、缓存口径、实际型号和诊断 ID。M3.1 相应问题独立记录候选证据，未实测项不得继承 M3 的线上结论；其离线约束仍按官方声明覆盖。Responses 采样零值、强制工具、priority/主动缓存等明确排除项无需为“全部参数兼容”扩大探索。若正式必交行为无法成立，保留可复现失败、对应身份未完成；不得用跨协议转换、删掉失败路径或 UNSUPPORTED 豁免。

### R26 — 自身 Live Smoke、预算与证据适用性

每进程仅一个 Operation × Provider × API × 型号 × 凭据用途，且只持有本配置 Key；M3 三身份分别生成自己的候选与正式证据，M3.1 三身份仅生成独立候选证据，二者不在同进程读取两种 Key。沿用现有 live tag 和运行开关。正式配置关闭 SDK/Binding 重试，显式 Retry Policy 的正确性由离线 E2E 验证。

每个候选或正式身份进程上限 20 个 Logical Call、24 个实际 HTTP Attempt、40000 个请求输出预算 tokens，单次输出上限 4096、单调用 120 秒、进程 1800 秒且服从 Host 更短时限。输入按固定合成文本/历史的保守估算累计预留最多 120000 tokens，实际已知用量超出估算时停止后续请求，不把取消后的未知量当零。场景运行前预留最坏输出/Attempt，合法历史须完成至少一次 Host 工具往返；text/thinking、开关/档位、截断、缓存/usage、error/cancel 场景固定 inventory。预算为低成本协议取样，长上下文阶梯用离线合成用量验证，不执行高容量收费压测；输入估算边界与误差须记录，不能把这些本地预算宣称厂商计费硬上限。

候选阶段含所有 SDK/HTTP 探索，当前六个国际站身份与此前中国区历史消费共同受最多 120 个 Logical Call/144 个 Attempt 限制；中国区已消费的2次调用/2次Attempt保留，地区切换不刷新阶段余额或route/model/purpose预算槽；该Chat/M3/standard槽仍受20/24限制，历史计数与当前精确Provider/endpoint证明身份分别校验；正式阶段首期三个 M3 身份合计最多 60 个 Logical Call/72 个 Attempt。M3.1 线上候选依合法配置执行，缺资格可 NOT_RUN，但其三协议离线约束仍为必交，不影响 M3 正式门槛。不得循环启动新进程绕开同阶段上限。场景重跑至多一次且占同一预算，预算不足为 FAIL。固定 inventory 或预算需变更时先更新规范和证据计划，不运行中放宽。candidate/controlled/capture replay 不发布正式 PASS；正式报告绑定实际 SDK、执行源码完整集合/依赖、目录/价格 Hash、型号、endpoint、区域与账户用途别名及捕获来源。执行源码或目录改变令不适用证据过期，允许的文档变更不得篡改已测身份。

### R27 — 支持矩阵、门禁与交付判定

新增 minimax-chat、minimax-responses、minimax-messages 三个正式协议组合；每条必须有 M3 自身完整正式报告。M3.1 的候选记录与矩阵隔离，不用候选 Hash、合成价格或 M3 报告给它授予正式支持。后续多型号正式推广若需要汇合报告，必须在已有证据边界保留每型号独立来源、当前失败与完整覆盖，不让门禁/聚合进程持有厂商 Key；本期不为尚未转正的 M3.1 新造多型号合并机制。

PASS/FAIL/NOT_RUN/UNSUPPORTED 含义保持；缺 Key/权益不是平台不支持，真实故障不是跳过。文本、工具往返、思考开启与合法续接、隔离、终态和资源约束均为必交，不能 UNSUPPORTED 放行。新增三组合与六型号追溯是增量合同；全部既有十三组合、十四自身支持身份、分类/图像/业务效果、差分、压力、普通/race、两种 vet、脱敏与回滚要求继续有效。不得删除旧映射、降低预算负载或减少能力选择器以通过。

候选通过不等于正式目录支持；正式纳入后的完整当前报告、目录/价格快照与门禁通过均需可核验。研究、代码编译、合成 fixture 或历史 PASS 单独都不是本 spec 完成证据。三个 M3 身份任一必交缺失，或 M3.1 三身份的离线型号约束缺失，则本 spec 交付不完整。最终声明须明确“MiniMax 国际站 M3 三协议文本/工具正式支持，M3.1 为候选”，不称所有型号、所有参数或多模态全面支持。

### R28 — 文档、ADR 与回滚

交付完整 Host 配置示例，说明两种凭据用途、三绑定、文本子集、六型号默认/档位、full/simple、hook、缓存提示、标准服务层级、USD 估算、诊断、原生状态信任及真实支持状态。关键能力拆分、M3 二元思考、M3.1 必选深度、Responses 原生输入和与冻结 pi 的差异同步 ADR；继承 ADR-0032/0033–0036 的边界，但不把 GLM 特例泛化成所有兼容 Provider。

无数据库迁移、无新增对外 HTTP 接口、无新包级环境凭据。配置回滚可禁用 MiniMax Binding，调用明确拒绝；软件回滚恢复本变更前完整实现、目录、价格、矩阵、差分与发布合同，保持用户其他工作。回滚需验证旧路线仍工作、MiniMax 被移除后无凭据读取或 Provider 请求；不新增永久兼容 shim、自动转投或双写。

## Testing Decisions

### T01 — 唯一业务验收接缝

用户已于 2026-10-18 确认采用现有公共 Client E2E。经 Stream、Complete、StreamSimple、CompleteSimple 和 WithHooks 观察最终请求、流事件、Result、Error、Attempt/Usage、Call Attribution 与释放结果。目录/配置校验沿现有公共入口；离线受控 TLS Provider 与真实服务均从同一个 Client 发起，不为 MiniMax 再建低层业务接口。

沿用 DeepSeek 两协议的路由/工具回放、GLM 三协议的边界/真实捕获、Anthropic 内容块与签名、可信回调、多租户和资源套件先例。Host 的 Resolver、Catalog、ResourcePolicy、HTTP transport 为现有装配接缝；测试可捕获最终 wire，但不以私有 helper、能力结构字段或 SDK 类型布局为断言。来源检查、证据 CLI/合并器和门禁是既有辅助边界，不能代替业务 E2E；必要隔离测试仍须先列失败方式再实现。

### T02 — 先写失败方式，再写实现

每项变更先写可失败的外部场景并留 red 证据，再实现至 green。未知 wire 先有受限候选记录，再将已确认形状写成失败 E2E，最后适配；不得实现完补镜像单测。每场景标明合成官方 schema、真实捕获回放或实际 live；不同证据类型不能互冒。下表全部为必验组，其中排除能力验证的是明确拒绝行为。

| 验收组 | 可观察行为与主要失败方式 | 对应决策 |
| --- | --- | --- |
| A01 身份/认证 | 三种精确资源路径、单一认证与SDK版本头；环境 Key 不生效、双认证拒绝；禁用/越权/错 API 在读 Key/请求前拒绝 | R01–R04 |
| A02 六身份 | M3正式/M3.1候选、协议/用途配置独立；候选与正式 Hash/价格不混；未授权套餐不自动转按量；文本子集占位与回调多模态拒绝 | R03、R06 |
| A03 思考默认 | nil/full/simple 下 M3 Chat on、另两路 off、M3.1 全 on；不派生无效关闭；M3 high只表示开关 | R07 |
| A04 档位与强制 | M3.1 五档、minimal/none、full 拒绝/simple夹取；M3不伪装深度；effort-only Messages 不丢选项；采样/hook关闭必选思考被拒 | R07、R15、R20 |
| A05 采样/预算 | 各层次覆盖顺序、省略的默认采样标签与wire省略分离、不随回调重选；预算1/上限/越界/精确分数；Responses零采样冲突；Messages忽略采样map | R08 |
| A06 Chat请求 | system/max_completion_tokens/include_usage；禁止store/developer/tool_stream/strict/双长度；公共auto与原生auto的固定行为 | R09 |
| A07 Chat推理历史 | M3字段冲突按自身wire定；reasoning_details原生结构、reasoning_content与think原始格式不混；完整assistant工具回传 | R10 |
| A08 Chat流末 | 正文/思考/多工具分片；finish同帧与晚到usage；有无DONE的合法EOF；缺finish、过滤、未知原因、流内error不假成功 | R10–R11 |
| A09 Responses形状 | 数组正文/推理、只有done或终态、done校正、重复/乱序/缺终态、reasoning-only截断；无静默SDK丢字段 | R12–R13 |
| A10 Responses续接 | 原始reasoning→输入转换→工具结果后继续；call_id/item_id不混，标点/Unicode/长值不碰撞；不用存储引用 | R14、R21 |
| A11 Messages请求 | adaptive/disabled与effort，显式temperature不因思考被吞；不带budget/display/beta/eager/browser；auto/none可用、any/具名/stop被拒；metadata不泄露租户 | R15 |
| A12 Messages流历史 | 块索引与类型、签名delta、多工具JSON；原样thinking工具续接；已确认有/无签名/空签名规则；非法顺序/缺stop/error失败 | R16、R21 |
| A13 能力守卫 | 强制/禁并行、store/truncation/background/include/summary/格式/忽略停止字段、priority/主动缓存从具名/采样/hook均无法绕过；schema同名业务属性合法 | R05、R09、R12、R15、R17–R20 |
| A14 缓存 | 自动用量保留；Responses隔离路由键、none不发、short/long不承诺期限；Chat无未公开key、Messages无默认主动写入或1h | R17 |
| A15 用量与费用 | 缺/零/部分/中断usage；Messages输入加缓存而Chat/Responses拆缓存；reasoning不双计；国际站USD来源、512000边界±1、priority拒绝、未知实际型号阻止正式通过 | R18–R19 |
| A16 错误与重试 | 401/403/400/429/5xx、余额与套餐永久错误、流内error；实际trace与原始停止原因；默认零/显式有限重试、流开始后零重放 | R22 |
| A17 回调/不变性 | header→payload→response顺序；授权先于能力、发送前零Attempt；回调修复合法、越权不合法；重试不重跑hook；输入/结果map修改不串扰 | R20 |
| A18 隔离与释放 | 两租户同名binding/model/session、不同用途并发；持久化信任/跨账户降级；Admission、取消、deadline、帧/队列/工具限额、唯一Result及资源归零 | R21–R23 |
| A19 回归与差分 | 原厂/DeepSeek/GLM旧wire与错误保持；冻结Messages适用场景实际差分；Chat/Responses及M3.1扩展、无无关skip/未解释pending | R05、R24 |
| A20 正式交付 | M3三份自身正式live、M3.1离线必交与线上候选独立状态；预算/inventory/source/catalog适用性、矩阵原子更新保留失败、全部既有门禁/脱敏/完整回滚 | R25–R28 |

### T03 — 真实缺口收敛与固定预算

执行前形成按六身份区分的有限场景表、研究问题、预算预留与通过断言。M3 Chat 同一轮完整工具历史、Responses 原生 reasoning 续接、Messages thinking签名与工具续接必须各自产生真实证据；M3.1 若执行候选，也按其自身相同能力核验，不继承 M3 的结果。某次模型没有返回 thinking 或工具不能直接算通过该能力；可按预定重跑一次，仍无证据保持该能力失败/未完成，不能用另一个型号补齐。不得用强制工具参数绕开本期未支持策略。

自动缓存可用重复合成前缀进行有界对照；未命中不伪造命中，报告原始count/presence。厂商不保证缓存命中，额外“命中能力”仅凭自身证据声明；必交的是正确读已返回用量和合法继续调用。服务真实永久拒绝保留 error 捕获，临时网络故障保留 FAIL；回放这些捕获只能证明解析，不产生新 live。

### T04 — 可重放、可审计交付物

每次公共 E2E/live 产出有限 case inventory、原始合成/脱敏请求与响应SSE、事件与最终Result、Attempt/已知usage、Observer、断言、资源读数和运行模式；大负载按既有方式保存生成参数/长度/Hash，不复制无界正文。文件清单、SHA-256、来源/SDK/执行源码/目录价格/计划身份完整，失败历史不覆盖。离线重放命令不读取真实 Key、不调用厂商，报告清晰区分记录与重放。

完整验收需普通和race公共E2E、受影响冻结差分零未解释pending、两种vet、M3三身份正式自身报告、M3.1三身份离线约束和候选状态、全部现有资源压力与业务效果、逐包脱敏、目录价格快照和追溯。门禁使用可信实际case与不可变来源，不能只信可编辑manifest声明；缺文件/篡改/多报case/漏Attempt/错误预算/旧目录或执行来源都要失败。源码改变须按既有适用性规则更新证据，不为省调用改写历史捕获。

## Out of Scope

- MiniMax 自有生成协议、旧式文本 Completions、中国区运行时接入、第三方网关及自动跨协议转换。
- 图片/视频/音频输入、图像生成、Files/mm_file资源、resize与新媒体DTO；厂商多模态能力不等于本期已接入。
- M2.x、高速别名与迁移旧名的正式目录；不自动把旧请求改成M3.1。
- M3.1 在缺少可验证价格基准/权益时的正式目录与正式支持声明；其已知三协议型号约束的候选实现和离线验证仍在范围内。
- priority及其他服务层级、Messages主动缓存/TTL保证、缓存命中保证、套餐扣费/余额/权益管理。
- 强制/具名工具选择、禁并行控制、严格JSON Schema、原厂托管工具、MCP、服务器保存/续接/后台任务及远端取消承诺。
- Claude预算型/managed thinking、原厂display/beta/块绑定、fallback或中途工具变更；不增加Agent/Session/TUI职责。
- 新数据库、对外HTTP接口、通用Provider插件中心、运行时汇率服务或隐藏的token估算远程请求。
- 真实厂商百万上下文容量压测、生产效果保证以及用合成/小样本通过推定生产质量。

## Further Notes

1. 本 spec 将研究建议收敛为固定交付边界：M3三身份正式交付，M3.1三身份实现已知约束并保持候选；首期文本/工具、standard和自动缓存，Responses路由键独立支持。M3.1缺价不成为虚构免费或阻断已有正式型号的理由。后续推广M3.1、增加priority、主动缓存或视觉，均须另有明确契约、价格和自身验收，不靠通用payload透传变相开放。
2. 官方事实的检索日为2026-10-18；国际站来源、逐项文档Hash、冲突与URL集中在[研究报告](research-international.md)。编码前若官方契约改变，保留旧来源并记录差异，不能把新文档静默替换成当时已知事实。
3. 三协议主来源：[Responses](https://platform.minimax.io/docs/api-reference/responses-create)、[Chat Completions](https://platform.minimax.io/docs/api-reference/text-openai-api)、[Messages](https://platform.minimax.io/docs/api-reference/text-anthropic-api)。模型资格与迁移依据：[型号概览](https://platform.minimax.io/docs/guides/models-intro)、[M Plan FAQ](https://platform.minimax.io/docs/m-plan/faq)。未知不能升级为已支持，缺凭据不能记为厂商不支持。
4. 计划/实施必须追溯US01–US80、R01–R28与A01–A20至实际有限case；实施任务若另行拆分，沿本地工单约定逐票建档，不把本spec当作已完成实现或真实支持声明。
