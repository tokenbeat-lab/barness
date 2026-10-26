# barness-ai 接入 Grok：OpenAI Responses

Status: ready-for-agent
Type: feature
Updated: 2026-10-26

本文是交付契约；[详细设计](design.md)说明实现边界，[研究报告](research.md)提供官方依据与文档疑点。用户已明确排除 Messages，并于本次设计中确认沿用现有公共 Client E2E 验收边界。ready-for-agent 表示文档可供实施，不表示代码、型号资格或发布已经完成。

## Problem Statement

Host 希望通过现有 barness-ai Client 调用 Grok，保留统一的消息、流事件、结果、推理输出、函数工具、用量和多租户隔离行为。模块已经具有 Responses adapter 和 DeepSeek 等兼容厂商的接入模式，但尚无 xAI 的独立身份、能力配置、型号约束和自身验证证据。

直接替换服务地址会继承 OpenAI 的字段和假设：Grok 的存储默认开启，摘要控制只为兼容，输出预算不限制全部消耗，工具可能整块返回，推理 token 的用量口径还存在官方示例歧义。缺少独立转换会造成历史丢失、工具配对错误、意外存储和费用少计。

## Solution

沿用 DeepSeek 的 Provider/API Protocol 分离与共享 adapter 方案，新增 Provider=xai，首期只交付 chat × xai × openai-responses × grok-4.7。使用全球标准 API、现有 OpenAI Go SDK、四种公共聊天入口及现有 Hooks；Host 继续负责认证、授权、凭据、Transcript 选择、持久化和工具执行。

请求显式保持无状态、完整历史和加密推理回放。厂商差异由 Responses 内部配置、DTO 转换及边界校验承担；用量与价格按 xAI 事实转换，不扩大公共调用接口。先交付候选与公共 E2E，再取得本组合自己的真实证据，满足适用门槛后才进入内置目录。

## User Stories

1. As a Host developer, I want Grok through the existing Client, so that I can reuse my application integration.
2. As a Host developer, I want Stream and Complete, so that I can choose streaming or final results.
3. As a Host developer, I want both Simple entry points, so that common calls use protocol-neutral options.
4. As a Host developer, I want existing ResponsesOptions and Hooks, so that Grok needs no separate calling interface.
5. As a Host developer, I want Provider and API Protocol kept separate, so that protocol compatibility does not change service identity.
6. As a Host operator, I want only the implemented Grok route selectable, so that an unsupported protocol cannot inherit another vendor's defaults.
7. As a local Host operator, I want explicit credential resolution, so that ambient environment keys do not change the target account.
8. As a cloud Host operator, I want tenant-owned Bindings, so that users cannot select another tenant's service.
9. As a cloud Host operator, I want model allowlists enforced, so that a valid key cannot widen model authorization.
10. As a Host operator, I want immutable Binding and Credential Snapshots, so that rotation cannot change an in-flight Attempt.
11. As a Host operator, I want the global endpoint fixed for this release, so that regional pricing cannot silently change.
12. As a Host developer, I want catalog discovery without reading secrets, so that configuration remains separate from calls.
13. As a Host developer, I want candidate models separated from the built-in catalog, so that unverified support is visible.
14. As a Host developer, I want model constraints enforced on custom catalogs, so that metadata cannot enable unsupported behavior.
15. As a Host developer, I want system instructions preserved, so that Grok receives the intended Transcript.
16. As a Host developer, I want complete historical input on every turn, so that continuation does not depend on a stored response ID.
17. As a Host operator, I want store:false enforced after callbacks, so that request customization cannot restore server-side conversation storage.
18. As a Host operator, I want the difference between stateless storage and ZDR documented, so that I do not infer an unsupported privacy promise.
19. As a Host developer, I want encrypted reasoning preserved even without visible text, so that later turns retain native context.
20. As a Host developer, I want reasoning summaries streamed, so that I can show available progress.
21. As a Host developer, I want default and explicit effort handled separately, so that omission preserves Grok's default.
22. As a Host developer, I want supported reasoning levels mapped precisely, so that unsupported wire values are never invented.
23. As a Host developer, I want invalid full options rejected before credentials, so that bad requests have no provider side effects.
24. As a Host developer, I want the output budget's limited meaning documented, so that I do not use it as a billing cap.
25. As a Host developer, I want small valid output budgets preserved, so that another provider's minimum does not change my request.
26. As a Host developer, I want explicit sampling rules, so that echoed values do not misrepresent actual sampling.
27. As a Host developer, I want text and inline PNG/JPEG inputs, so that Grok can inspect images without uploaded resources.
28. As a Host operator, I want image and request limits enforced before HTTP, so that large payloads remain bounded.
29. As a Host developer, I want ordinary function tools, so that my Host can execute model-selected work.
30. As a Host developer, I want auto, none and required tool choice, so that supported tool policies remain effective.
31. As a Host developer, I want parallel calls correlated independently, so that multiple results return to the correct calls.
32. As a Host developer, I want tools returned as complete items, so that missing argument deltas do not lose calls.
33. As a Host developer, I want complete final arguments validated, so that partial display JSON is never executable output.
34. As a Host developer, I want opaque call and item IDs preserved, so that punctuation, Unicode or length do not break correlation.
35. As a Host developer, I want incomplete and failed responses distinguished, so that I can handle truncation and failure correctly.
36. As a Host developer, I want final-only output supported, so that I can receive results even when intermediate events are absent.
37. As a Host developer, I want repeated snapshots reconciled, so that text and tool calls are not emitted twice.
38. As a Host operator, I want malformed streams to fail explicitly, so that EOF cannot masquerade as success.
39. As a Host developer, I want trusted native history to survive persistence, so that restored conversations can continue correctly.
40. As a Host operator, I want untrusted or cross-account state downgraded, so that ciphertext never crosses its source boundary.
41. As a Host developer, I want visible history retained after downgrade, so that losing native trust does not erase the conversation.
42. As a Host developer, I want missing usage distinct from zero usage, so that absent reports do not imply free calls.
43. As a Host developer, I want reasoning counted exactly once, so that normalized output agrees with total token use.
44. As a Host operator, I want cached input separated from uncached input, so that estimates use the correct rates.
45. As a Host operator, I want long-context prices applied at the exact boundary, so that a 200000-token prompt is priced correctly.
46. As a Host operator, I want cost uncertainty exposed, so that estimates do not invent missing cache or reasoning reports.
47. As a Host operator, I want catalog version and hash attached to calls, so that estimates remain reproducible.
48. As a Host developer, I want automatic caching treated as a hint-based service behavior, so that cache misses remain normal.
49. As a Host operator, I want session cache keys scoped to the tenant and account, so that raw session identifiers are not sent.
50. As a Host developer, I want callbacks validated after all mutations, so that sampling and Hooks cannot bypass the implemented contract.
51. As a Host operator, I want bounded retries and cancellation, so that failure does not cause duplicate generations or leaked permits.
52. As a Host operator, I want concurrent calls isolated, so that a shared Client never shares credentials or mutable stream state.
53. As a Host operator, I want diagnostic metadata without secrets or ciphertext, so that shared observability remains safe.
54. As a maintainer, I want public E2E failures before implementation, so that behavior changes have repeatable evidence.
55. As a maintainer, I want Grok's own live evidence, so that another Provider's adapter success does not grant support.
56. As a Host operator, I want a reversible rollout, so that disabling a Binding stops new calls without corrupting history.
57. As a maintainer, I want implementation, verification and release states separated, so that a finished document is not reported as a shipped feature.

## Implementation Decisions

### R01 — 身份、目标与路由

新增公开 Provider 常量 ProviderXAI，值为 xai；复用现有 openai-responses。首期唯一操作是 chat，唯一型号是 grok-4.7，唯一生产目标是全球标准 API。不新增 API 枚举、Grok 专用公共 Client 或品牌归一化层。已知 xAI 的其他操作、协议、型号或 US endpoint 在凭据读取前明确拒绝；不实现 Messages 请求、配置、目录或测试路线。

### R02 — 候选目录与型号约束

候选声明文本和图片输入、文本输出；根据官方500k上下文标签，本次保守声明上下文上限500000，首期输出能力上限128000。两者是本模块交付范围，不声称已实测厂商硬上限。Reasoning=true，RequiresThinking=true，默认档位 high；low/medium/high/xhigh 映射同名值，off/minimal/max 标为未直接支持。Simple 使用已有 clamp；Full 只接受型号实际支持档位。Host 可收窄能力、输出和上下文，并提供独立价格快照，不得改写必需推理、扩大模态/上限或创造 wire 档位。

### R03 — 授权与一致快照

沿用 Tenant、Actor、Binding、Account Scope、Credential Snapshot 和 Catalog 交集授权。Provider/API/endpoint/型号约束与命名选项先校验，再读取凭据；账户及配置版本不一致直接失败。认证仅使用绑定快照中的 Bearer key；不读环境、不切 key、不跟随重定向。所有可变输入在既有边界复制。

### R04 — 公共调用与模块边界

保留 Stream、Complete、StreamSimple、CompleteSimple、Target、Request、Result 和 Hooks 的现有签名。复用现有 Responses adapter、SDK、assembler、Admission、限额和重试；xAI 内部拥有请求、流、历史及用量转换，DTO 不泄漏到领域模型。不复制生命周期，不建立可由 Host 扩大厂商事实的能力注册表。

### R05 — 请求与无状态保证

每次发送完整 input 数组、stream:true、store:false，默认 effort 省略时也显式请求 reasoning.encrypted_content。系统消息使用 system。不使用 previous_response_id、conversation、prompt、后台任务、服务端压缩、item_reference、file_id 或任何上传文件；不自动生成 developer、OpenAI 会话亲和头、缓存保留期或摘要控制字段。最终请求缺少 store:false 或必需 include 也必须失败。

### R06 — 选项与预算

Full 预算省略/null 时使用 Host 型号的已收窄上限；显式预算为1至该上限的精确整数，0和越界拒绝，不抬高到16。Simple 先按现有上下文估算收窄，再校验型号上限。max_output_tokens 仅是可见输出预算，不能承诺限制推理、工具参数或总费用。首期 temperature 只接受0..1，top_p只接受(0,0.95]；这是为避免服务端静默 cap 而收窄的接入范围。显式摘要风格、服务等级、惩罚、停止词及未集成采样参数拒绝。档位省略与显式 high 分开保留。

### R07 — 最终字段与回调校验

最终 guard 在无 Hook 时也执行，位于全部采样合并与 OnPayload 之后、HTTP 之前。按授权、协议子集、预算和资源限制的顺序检查。型号、固定无状态语义、派生缓存 key、原生资源授权不可扩大；合法业务 JSON 中恰好名为 store、model、file_id 的属性不能被当作协议控制字段扫描。最终 wire 子集由设计中的字段表固定。

### R08 — Transcript 与原生回放

保留消息与工具结果顺序，不裁剪历史、不生成摘要、不借服务端 ID 补全历史。reasoning 原始条目保存在既有 Thinking.Signature，密文及不引入新资源语义的附加字段按合法 JSON 原样保留；有可信来源时回送完整 reasoning 条目。保留未知字段不授予文件引用、其他native种类或新工具能力。可见推理与原生条目分开。原生条目畸形时明确失败，不能静默丢弃或伪造签名。assistant 文本和函数调用按公开输入 schema 独立投影，不生成 OpenAI 专属合成 item ID。

### R09 — 图片输入

首期只接纳严格内嵌 base64 PNG/JPEG，单图不超过20 MiB且仍受更小的 Host 资源策略约束。普通 user 与工具结果图片分别按已声明输入形状编码；工具结果图片如需拆成后续 user 图片条目，必须保留结果配对和顺序并取得本路线证据。不抓取公网 URL、不上传、不接受文件 ID、不把能看图片当成图像生成授权。

### R10 — 函数工具子集

首期至多350个函数工具，使用平铺 name/description/parameters。参数 schema 根限定 object；更宽的组合根、具名工具强制、allowed-tools、custom/grammar、结构化回答及全部托管工具不在首期范围。tool_choice 支持省略/auto/none/required；required必须有工具。parallel_tool_calls 可省略或为布尔。strict省略，不能声称可关闭厂商隐式约束；不改写 schema 的 required 或 additionalProperties。

### R11 — 流合并与终态

支持正文、两种 reasoning 文本事件、函数参数、完整 output item 以及只有终态 output 的结果。以 output_index、内容/摘要索引和 item/call ID 关联，不以工具名关联。done/终态快照具有权威性，必须与已完成内容一致，重复快照不重复发出领域内容。无真实终态的 EOF/[DONE]失败；completed、incomplete、failed、error按设计逐项处理，HTTP200不构成成功。只增加必要的 xAI 协议状态，不更改旧路线宽松度。

### R12 — 工具 ID 与约束兑现

使用版本化、不歧义的本地 opaque ID 保存 call_id/item ID 二元组，不假定分隔符、fc_前缀或64字符限制。同源可信回放保留两者；降级仅保留调用关联，去除账户专有 item 状态。工具调用与结果同步重命名，重复或冲突关联拒绝。正常完整终态若 required没有函数调用、none却出现函数调用，或最终参数不是完整 JSON object，则 protocol失败；不自动执行、修补为{}或改成普通成功。

### R13 — 来源隔离

原生回放必须满足可信 tenant/account/provider/API/model 封套；持久化反序列化不能恢复信任。同租户的无担保、跨账户或跨型号历史沿用既有降级，保留可见文本与工具配对并记录原因。不同租户历史授权由 Host 保证；库不能从普通文本识别越权。凭据轮换不改变同账户来源归属。

### R14 — xAI 用量转换

严格读取原始数字与字段 presence，不依赖 SDK 的零值。首期只含模型 token 与 Host 函数，不含托管工具。总输入P、总量T完整时，规范输出B=T−P；完整原始output O和reasoning R只接受B=O或B=O+R两种与事实一致的表示，R不得大于B；该算法是本模块针对研究歧义的设计决策，须由真实捕获确认适用。缓存C从P中分离，Reasoning只作B的子集，绝不再次加进TotalTokens。缺失、null、零、非法和矛盾各自处理，详见设计用量表。

### R15 — 估算成本与价格

复用现有 Attempt.CostEstimate 表达 resolved/bounded/partial/unreported，不新增账单或货币接口。全球标准价按每百万token输入2、缓存0.5、输出6美元；prompt总输入达到200000时整次请求分别改为4、1、12美元，包含缓存输入，边界为大于等于。缺缓存计数但总输入/总输出确定时，以允许的缓存分配算上下界，不填造缓存观测。费用使用固定Catalog版本及哈希，不在线更新，不复用OpenAI等级倍率。厂商reported cost字段不在首期公共交付合同，不用其替代Estimated Cost。

### R16 — 缓存提示

自动缓存不构成隔离保证。SessionID仅通过已有tenant/provider/account/session派生键发送；不发原始SessionID或另一套会话头。CacheRetention none只省略派生键，不保证厂商禁缓存；short/long均不产生TTL字段，long只是无法兑现时长的提示。缺缓存或被驱逐仍可正常调用，不把cache hint当资源授权或必然折扣。

### R17 — 失败与重试

沿用稳定Code/Phase：401/403为upstream_auth，429为rate_limited，5xx/408/409为upstream_error，其余4xx为invalid_request；畸形协议为protocol。默认不重试，显式策略仅覆盖获准的初始请求，流开始后不重放。上下文取消、超时、字节/队列限制不通过重试绕过。不自动跨Provider、账户、endpoint、协议或型号回退。

### R18 — 并发、限额与释放

每次Logical Call独立持有编码、解析、原生状态及用量快照；共享Client不存当前账户或key。帧、输出、工具JSON和事件队列均有界；加密内容、重复终态快照也计入既有字节限额。任何失败、取消、断流与回调错误都关闭正文、释放Admission并产生唯一终态。Complete不创建流事件队列。

### R19 — Hooks 合同

保持headers→payload→response顺序，payload回调在初始重试之外只执行一次，response回调在领域start之前。认证、目标、无状态与原生引用受保护。普通选项不支持在前置校验失败；仅Hook产生的非法子集为callback_failed，扩大授权为tenant_denied。Hook panic维持既有公共语义，不借xAI接入更改其他Provider或新增通用recovery。

### R20 — 归属与观测

沿用Event Envelope、Call Attribution、Attempt与Observer，记录实际操作、xAI身份、账户别名、型号、版本、错误类别与用量完整性。实际返回型号可通过既有ResponseModel保留；不自动扩大可调用型号。request ID只采纳有来源且经真实捕获确认的响应头，缺失不伪造。默认日志/Observer不包含正文、密文、工具参数、Key或未经清洗的上游错误文本。

### R21 — 候选、真实资格与正式目录

候选可用于受控E2E和明确授权的本组合live，不能自动激活BuiltinCatalog。正式晋升必须具有明确来源的型号/价格约束、自身文本/推理/加密续接/工具/图片/用量证据、脱敏捕获回放、适用回归及当前SOP要求。不得以移除必交能力、改成UNSUPPORTED或借用DeepSeek证据完成此功能。

### R22 — Host 装配与回滚

提供现有接口的完整装配说明：全局endpoint、绑定白名单、凭据快照、资源策略、持久化信任恢复、工具往返、成本状态和足够但有限的超时。优先使用现有Host resolver，不为单一路线新建认证框架。禁用并更新Binding版本阻止新调用；已固定调用按原快照结束或由Host取消。回滚不删除历史原件、不修改旧价格快照、不自动切换协议。

### R23 — 实施与验证顺序

先在公共Client边界编写失败场景，再实现并取得完整受影响普通回执；冻结相关源码和输入后，按SOP依次推进受影响live、人工选择的race及压力。现有依赖足以完成方案；SDK升级、通用协议重写或全Provider成本重构不是本功能的默认工作。新文件保持职责单一并低于500行。

### R24 — 研究遗留项的关闭

用量包含关系使用R14算术规则并以真实非零reasoning捕获确认；工具数使用350的明确接入上限；具名强制及结构化输出直接排除；工具整块/终态输出和双摘要事件在E2E覆盖，再以自身捕获修正。若实测与设计矛盾，保留失败原件，先更新设计及失败E2E再改实现；不得吞错、改写原始证据或未经来源自行开放字段。

## Testing Decisions

验收接缝已由用户确认：只通过现有公共Client及Host resolver/Transport/Hooks注入受控HTTP/SSE行为，观察实际请求、事件、Result、Attempt、Observer和资源释放。四种聊天入口都属于同一个接缝；Catalog公开构造/查询是必要的配置边界，不增加内部helper单测。真实API也走相同Client，SDK探针不能代替正式资格。

测试仅断言对Host可见的行为、错误、无HTTP副作用、数据关联和限额；不绑定私有函数名、switch分支或内部对象布局。先列失败方式并保留red证据，再编码。参考已有DeepSeek Responses公共E2E、Kimi/MiniMax的终态输出及opaque ID场景、混合租户隔离和共享资源释放场景；不重建冻结pi差分体系。

| 验收组 | 必须覆盖的外部行为 | 需求 |
| --- | --- | --- |
| E01 身份与目录 | 候选与内置分离、元数据快照、非法自带目录、其他操作/协议/型号拒绝且未读凭据 | R01–R04、R21 |
| E02 请求 | 精确目标、单一Bearer、system、完整input、默认/显式effort都有store:false与include | R03–R06 |
| E03 最终guard | Hook删除store/include、改型号/缓存键、加入资源引用、业务schema同名字段、零HTTP | R05、R07、R19 |
| E04 选项 | full/simple档位、unset/null/0、1/15/16/128000及越界、采样边界、禁止字段 | R02、R06、R24 |
| E05 文本与推理 | 两类reasoning事件、无可见推理的密文、done-only/final-only、终态回填 | R08、R11 |
| E06 工具往返 | auto/none/required、并行、完整item无delta、参数不完整、350/351边界 | R10–R12 |
| E07 ID关联 | 分隔符、Unicode、长ID、成对碰撞、重复call_id、跨来源配对、不伪造item ID | R12、R13 |
| E08 历史信任 | 同源可信、JSON读回未担保、账户/型号/协议变化、凭据轮换、畸形原生条目 | R08、R13 |
| E09 图片 | PNG/JPEG、user与工具结果顺序、畸形base64/MIME、20MiB及Host更小限额、外部资源拒绝 | R09 |
| E10 流故障 | 重复终态、内容矛盾、错误index/ID、事件错序、EOF/[DONE]无终态、上游error | R11、R17 |
| E11 用量 | inclusive/exclusive两种有效算术、全零、字段缺失/null、缓存、负数/分数/溢出/矛盾 | R14 |
| E12 成本 | 输入199999/200000/200001、缓存计入门槛、两档全量费率、未知缓存区间、无报告不免费 | R15 |
| E13 缓存 | 同域稳定派生、跨租户/账户/Provider不同、none/short/long无TTL、无二次会话头 | R16 |
| E14 HTTP与回调 | 401/403/429/5xx、有限重试、回调次数与顺序、key轮换、目标不变、脱敏 | R17、R19、R20 |
| E15 并发资源 | 两租户同名Binding、混合旧Provider、受控取消/超时、队列溢出、释放许可与正文 | R03、R18 |
| E16 Host回滚 | 禁用后零新HTTP、已固定调用一致、历史可恢复、旧快照不改、无自动回退 | R22 |
| E17 自身live | 四入口文本、非零推理、密文续接、工具往返/required/none、图片、usage/成本与模型归属 | R21、R24 |
| E18 回归与发布 | 受影响旧Responses、共享native/ID/usage回归、当前来源与原件绑定、五项事实只读判定 | R04、R23 |

每次E2E保留可重复的请求/脚本、流事件、Result、断言及manifest/audit，引用实际命令、退出码、来源和输入；失败attempt不覆盖。live只跑被明确授权的有限真实场景，错误风暴与限额故障只在受控服务注入。race和压力由后续人工选择；本次文档请求不授权这些执行。

## Out of Scope

Messages协议整体、Chat Completions、xAI自有协议；其他型号、US/其他地区endpoint、priority/fast/flex；服务端会话、资源读取/删除、压缩、文件上传、URL抓取；托管工具、MCP、图像生成、音视频、分类；具名强制、复杂schema根、JSON Schema回答、可配置strict；自动执行工具、Agent loop、自动上下文裁剪；团队ZDR管理、厂商账单公共接口、在线价格更新、硬费用预算或分布式限流实现。

对xAI未实现组合的前置拒绝只是路线授权收敛，不构成Messages协议的设计或实现任务。已存在的其他Provider协议继续服务，不因本功能移除。

## Further Notes

- 依据：[研究](research.md)、[详细设计](design.md)、[ADR](../../docs/adr/0054-barness-ai-xai-responses.md)、[术语](../../GLOSSARY.md)、[测试SOP](../../docs/agents/verification.md)。研究仍保留此前两协议的历史范围，本spec以用户最新的Responses-only要求为准。
- 本spec沿用已有Provider能力、回调授权、原生状态与目录原则；xAI用量转换和来源保留的设计取舍由ADR记录。官方示例存在歧义，不将本地算法冒称为xAI已证实的全量协议事实。

### 唯一当前状态

| 项目 | 当前事实 |
| --- | --- |
| 实现 | 候选实现与证据修复完成：01–04、07–09已解决，集成分支 `codex/barness-ai-grok`；05真实资格、06正式交付仍未关闭 |
| 冻结来源 | 实际受测 `02b422bfffddefcfb0d95ec018867ed4dff816dc`，1962个完整执行文件；选定输入 [cases-v2.json](verification/cases-v2.json)，SHA256 `c14e8116158c5213e6560fa87fb745221cae11c3ffc016d7c1905556acb31a1f`；收尾提交与受测来源的集合相等性另存 |
| 审查 | 原Standards2项、Spec3项全部修复并独立复核0/0；证据标签修正的增量复核亦0/0；[v1](review/review-v1.md)、[v2](review/review-v2.md)、[v3](review/review-v3.md)分别保留 |
| 完整普通v2 | 采集器实际PASS/exit0；两个Go包15402个测试PASS；主包14586个原生case PASS、214个opt-in压力/压力消费者case NOT_RUN；必需565/565 PASS、65预期入口齐全，audit0、非规范目录0 |
| 原生消费 | 真实565-case定向包的公开原生消费者已验证source/e2e=true、live/race/压力=false、CLI exit1；完整包只读结果见[readonly-v2](../../.evidence/barness-ai-grok/release/readonly-v2/)及[不可变索引](verification/record-v2.json) |
| 自身live | NOT_RUN；18场景/24调用只有controlled证明，冻结二进制无Key自检Calls0；[有限待授权计划v2](next-stage-plan.md) |
| 共享历史live | xAI外另22条旧聊天目的路线存在xAI来源ID分配变化，分别待计划/授权；完整23路线见[影响表](live-impact.md)。旧来源报告不视为当前通过 |
| race / 压力 | 均NOT_RUN；尚未人工选择，须先满足适用live后依序选择；压力适用，普通运行的214个opt-in场景未计为PASS |
| 目录/发布 | BuiltinCatalog未激活xAI，无正式支持声明；release.pass=false，保留付费与人工阶段所缺事实 |

#### 最终普通回执与可重复入口

```sh
python3 tools/verification/run.py --feature barness-ai-grok --stage e2e \
  --input .scratch/barness-ai-grok/verification/cases-v2.json -- \
  go test -json ./ai/e2e ./ai -count=1 -timeout=20m
```

- 实际受测SHA：`02b422bfffddefcfb0d95ec018867ed4dff816dc`；exit0，墙钟476.188秒；[receipt.json](../../.evidence/barness-ai-grok/e2e/20261026T040239.130313Z-uw3gpgli/receipt.json)，SHA256 `aac6ba38aa6c33e1a4ec61d5e91c6fa5ce765e8ea040bdda7f080d5b40e36cba`。
- 独立逐项核对：[validation-v2](../../.evidence/barness-ai-grok/final-ordinary-validation-v2.json)，SHA256 `1a106f750315c50c74f876ef54b3ca3384e8c9f59ec7de04b4d362b6d3951d63`；565必交case、65入口齐全，两个Go包实际PASS，源码/选定输入一致。
- [原生manifest](../../.evidence/barness-ai-grok/e2e/20261026T040239.130313Z-uw3gpgli/bundles/20261026T040243.945177000Z/manifest.json)，SHA256 `e9d046fab847af0f90ee32a79888824009bb8758ff07809df78b60c086cd3177`；[audit](../../.evidence/barness-ai-grok/e2e/20261026T040239.130313Z-uw3gpgli/bundles/20261026T040243.945177000Z/audit.json)，SHA256 `8aa973100a15bf817985395637131e838cf84e427be4792583def923f020f5fa`。
- 精确清单为515个xAI与50个旧ID回归；包含真实producer→collector→原生CLI消费接缝。[独立枚举依据](verification/cases-v2.md)。未编写内部helper单测。

#### 原失败及来源收尾

原v1普通采集器PASS，但其E04省略effort的case目录含中间`//`，原生消费者实际拒绝。该问题通过明确`default`证据标签修正，实际请求仍省略effort，并补真实公开CLI失败/成功回归；导入器合同保持。原[readonly-v1报告](../../.evidence/barness-ai-grok/release/readonly-v1/native/release-report.json)、[v1索引](verification/record-v1.json)、全部失败/无效/成功attempt及[修正前状态原字节](history/spec-before-canonical-case-fix-v2.json)独立保留，未改写旧原件。前一可变状态另存[收尾快照](history/spec-before-final-v2.json)。

所有实施工作树原件已复制至原仓库 `.evidence/barness-ai-grok/` 并逐文件SHA256核对。最终来源后只做文档/索引收尾，完整1962文件集合与内容、case-v2选定输入集合及内容须与受测来源相等；相等性回执另记录实际受测提交和收尾当前提交。用户另有未提交AGENTS.md修改及两个scratch目录，均保留为原有工作区内容。

#### 编译产物与尚未执行项

[候选构建](../../.evidence/barness-ai-grok/build/20261026T040254-candidate-6uadcnzf/build-receipt.json)与[只读门禁构建](../../.evidence/barness-ai-grok/build/20261026T040254-releasegate-el73wpu7/build-receipt.json)实际exit0，GOENV=off、GOFLAGS为空、隔离HOME/凭据；source/binary/catalog/SDK机器事实已固定。候选无开关/无Key精确入口实际为NOT_RUN、Calls0，不产生真实资格。

本次仍缺明确付费授权、账户/地区别名及USD停止阈值。05须取得自身捕获、脱敏回放、各受影响路线与原生正式资格；06另需正式晋升、人工race与适用压力选择后再只读判定。race候选是当前Client/Responses和共享opaque来源路径；压力候选含大工具参数、密文/重复终态、队列/准入/取消及消费者，精确argv和负载待人类选择。NOT_RUN不等于不适用；真实证据维护/隔离/purge未执行。

回滚先禁用并更新xAI Binding版本，固定调用按原快照结束或由Host取消。保留Transcript、Native State Envelope、旧价格快照及验证原件，不自动切Provider/API；无数据库迁移。
