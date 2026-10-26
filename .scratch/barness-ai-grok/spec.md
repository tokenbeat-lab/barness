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

维护者于2026-10-26明确选择“采用建议顺序，发布前保持未正式支持”：允许在未发布集成分支完成并审查最终目录接线、原生资格器和正式harness，再以同一冻结来源执行另行授权的正式live、race和压力；全部通过后才接受/发布正式晋升。目录源码接线本身不构成正式支持，候选PASS不自动触发该接线；不得在最终门槛完成后再修改运行源码或目录而继续复用旧pin。详见[冻结顺序方案](formal-qualification-plan.md)与[ADR-0054](../../docs/adr/0054-barness-ai-xai-responses.md)。

本次维护者另确认正式兼容范围为Grok自身与16条已发布旧目标，共17条正式路线。六条未晋升Kimi保持候选与受控兼容回归，未纳入本次正式兼容承诺，也不连带晋升；22条来源碰撞/复用受控路径全部保留。详见[已采用范围](formal-target-scope-plan.md)。

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
| 实现 | 01–04及07–16软件工单已解决；16局部DTO修复5f933af9、历史分支回归fd02b56a、v10两轴0/0及完整普通通过。05当前来源正式live、06人工阶段/发布未完成 |
| 冻结来源 | 软件pin fd02b56aa6177106ae1ef348fec0c3a4920b8a57；本轮完整普通实际受测81bdaeafbd2b79429daba6964d9c4d24fb37ae1d，2054默认来源文件及128选定输入固定；文档收尾沿用须全部文件/输入集合及内容相等，实际核验见final-source-equality-v8 |
| 审查 | v1–v8历史发现均已修复；16新增v9 Standards0/Spec1 P2（mandatory-history反例缺失），同一实现者补齐，独立[v10](review/review-v10.md)两轴均0项 |
| 当前完整普通（cases-v7） | 三包15,850 Go测试实际PASS、exit0/audit0、628.736秒；1469/1469必交case、121/121入口全部通过，主15020 PASS/214压力NOT_RUN，辅助GLM32 PASS/每case最多393文件。[receipt](../../.evidence/barness-ai-grok/e2e/20261026T122707.884222Z-3p40dgni/receipt.json)及[核验v7](../../.evidence/barness-ai-grok/final-ordinary-validation-v7.json) |
| 历史完整普通v3 | 采集器PASS/exit0；两个Go包15402测试PASS；主包14586个原生case PASS、214个opt-in压力/消费者case NOT_RUN；必需565/565 PASS、65入口齐全、audit0。v1/v2与手工校验失败原件均保留 |
| 原生消费 | 当前[索引v9](verification/record-v9.json)绑定新完整普通与本状态；source/E2E及其他五项事实由[readonly-v9原件](../../.evidence/barness-ai-grok/release/readonly-v9/native/release-report.json)逐项给出。原批次/readonly-v8仅为旧来源事实；当前正式live/race/压力未完成，release.pass=false |
| 自身live | 历史批次5a8正式xAI18/18 PASS、24调用/24HTTP、正式消费者exit0/pass=true、响应估算上界0.029358 USD；消费者修复进入其原生source后，当前来源正式资格NOT_RUN，不改旧原件pin |
| 共享历史live | 旧批次已按授权停止，2条执行/39调用39HTTP，后15条未跑；修复后的[冻结计划v3](formal-live-plan-v3.md)重新准备17路线，尚未获新付费授权；六条Kimi仍候选/受控 |
| race / 压力 | 均NOT_RUN；尚未人工选择，须先满足适用live后依序选择；压力适用，普通运行的214个opt-in场景未计为PASS |
| 目录/发布 | 未发布分支Builtin2026-10-26.1/xAI接线已完成；本次软件修复、独立复核与完整普通通过。当前新来源正式live及人工race/压力未完成，release.pass=false，仍未正式支持或发布 |

#### 最终完整普通与构建（当前来源）

实际受测 `97697bb2c2937974beb362aefc7a41bfaad3ee15`，源码与已复核f89e59a6的2052文件集合及内容相等。完整命令由[回执](../../.evidence/barness-ai-grok/e2e/20261026T093152.246258Z-ua2zgcqn/receipt.json)保存，显式128项`--input`、外层2400秒；Go命令为：

```sh
go test -json ./ai/e2e ./ai ./ai/internal/testkit/grokhistory \
  ./ai/internal/testkit/livegrant ./ai/internal/testkit/minimaxlive \
  ./ai/internal/testkit/release ./ai/internal/testkit/supportmatrix \
  ./ai/examples/chineseeval ./ai/live/cmd/... ./ai/release/cmd/... \
  -count=1 -timeout=30m
```

- [完整receipt](../../.evidence/barness-ai-grok/e2e/20261026T093152.246258Z-ua2zgcqn/receipt.json) SHA256 `05adb8a38450e38520db1b2dd18cb1f6ae016a494de6f9fe0f68988d4974b1c1`，exit0、1283.188秒；只记录真实墙钟，不累加各包并行时间。
- [独立核验](../../.evidence/barness-ai-grok/final-ordinary-validation-v5.json) SHA256 `d3fbbdb0d069a126d61f05dd7f07a78b43a50dd6d72c14526cbbd7d155578053`：1456必交/119入口全部PASS，Source/128输入逐项相等，所有辅助包完成且无失败。主原件15221case中15007PASS、214压力NOT_RUN；audit0。
- [静态清单](verification/cases-v4.json)、[独立枚举依据](verification/cases-v4-inventory.json)、[额外输入](verification/ordinary-inputs-v1.json)明确来源；11个旧草案helper遗漏已回源码补正，未重命名/删除原565基线case。
- [采集器进程补充](../../.evidence/barness-ai-grok/ordinary-aux/20261026T093339-forwarding-oidysu54/receipt.json)实际1项PASS，Source/128输入与完整普通相同；仅支持普通工具字段转发，不冒充主Go原生bundle或真实live。
- [冻结构建](../../.evidence/barness-ai-grok/build/20261026T091716-formal-freeze-99enr2ih/formal-build-receipt.json)保存所有实际argv、退出、source/binary与Go工具身份；零Key、零厂商调用、零claim。正式机器计划包含17条非激活模板（空id/零expiry），当前无执行权限。

#### 历史完整普通v3与候选构建

```sh
python3 tools/verification/run.py --feature barness-ai-grok --stage e2e \
  --input .scratch/barness-ai-grok/verification/cases-v3.json -- \
  go test -json ./ai/e2e ./ai -count=1 -timeout=20m
```

- 真实完整回执：[receipt](../../.evidence/barness-ai-grok/e2e/20261026T044535.989494Z-dh_qlt7t/receipt.json)，SHA256 `38d2f605589dfb1630bc03f4be713c6553ad6b7ae6780fc8ffd3d278a0c0e6f7`；实际受测 `36eaca012d45062242a67e5f48d01ef07bfe5c5e`，exit0、墙钟483.163秒。
- 逐项核对：[validation-v4](../../.evidence/barness-ai-grok/final-ordinary-validation-v4.json)，SHA256 `95d055750c21a19149aec0771ee39eefcc8996ae6334d92c3796caa8e3b7bd20`；565必需case/65入口齐全，15402 Go测试PASS，audit0。v3手工核对误拒绝了原生合同允许的单一末尾斜杠，失败记录保留；v4按`load.go`的physical目录规则核对，未改原件或导入器。
- [manifest](../../.evidence/barness-ai-grok/e2e/20261026T044535.989494Z-dh_qlt7t/bundles/20261026T044541.122196000Z/manifest.json)，SHA256 `dcfdf94e7a74b9e23da2758394e947b16a0374e28e87d4e9164134d5935eb2b7`；[audit](../../.evidence/barness-ai-grok/e2e/20261026T044535.989494Z-dh_qlt7t/bundles/20261026T044541.122196000Z/audit.json)，SHA256 `8aa973100a15bf817985395637131e838cf84e427be4792583def923f020f5fa`。
- 新[候选构建](../../.evidence/barness-ai-grok/build/20261026T044535-candidate-hvl6pw73/build-receipt.json)实际PASS，native source `05e89a133ceb71b6725ad7a6ee889170f6e49fa7a72845928b202416a7d2838b`，binary SHA256 `bf0f618074775a5e2f2679ffe9e50c19a2ce148dd1c819050c50670b03d416c1`；无Key自检NOT_RUN/Calls0。旧构建和旧授权仅为历史事实。

#### 自身候选真实验证 v3

原始授权响应“授权按计划 v3 运行一次”；grant `grok-20261026T050333Z-83abbc44`绑定source `05e89a133ceb71b6725ad7a6ee889170f6e49fa7a72845928b202416a7d2838b`、binary `bf0f618074775a5e2f2679ffe9e50c19a2ce148dd1c819050c50670b03d416c1`、账户GROK、地区global和显式代理。完整[索引](evidence/candidate-live-v3.json)，SHA256 `f1666c226de083abccff16f88fd37d0a470b023b0b914ad1ef2bc44d24aed773`。

- 新实际输入cases-v3/plan-v3/授权/grant/config的匹配定向普通回执：[receipt.json](../../.evidence/barness-ai-grok/e2e/20261026T050340.625526Z-vqbdjmhv/receipt.json)，565必需case实际PASS、exit0；执行来源与完整普通v3集合及内容相等。
- [真实采集器回执](../../.evidence/barness-ai-grok/live/xai-responses/20261026T050450.562258Z-tlbttr1o/receipt.json)，SHA256 `c61e9100958fc9af2c63a23aa05c86201f04c19ee8c5a623777ab9c96da318b5`；exit0、墙钟37.136秒，18场景完整、24调用/24HTTP、零自动重试、完整捕获24次。
- [原生report](../../.evidence/barness-ai-grok/live/xai-responses/20261026T050450.562258Z-tlbttr1o/bundles/20261026T050451.200594000Z/xai-candidate-report.json)、[manifest](../../.evidence/barness-ai-grok/live/xai-responses/20261026T050450.562258Z-tlbttr1o/bundles/20261026T050451.200594000Z/manifest.json)、[audit](../../.evidence/barness-ai-grok/live/xai-responses/20261026T050450.562258Z-tlbttr1o/bundles/20261026T050451.200594000Z/audit.json)保留；audit0，complete=true、formalQualified=false。估算上界0.031036 USD是响应费用估算累计，不是厂商账单。
- [正式消费者外层schema2](../../.evidence/barness-ai-grok/live-consumers/attempt3-wrapper-v1/consumer/live-consumer.json)实际exit1、pass=false；其报告与全部原件保留，候选不能自动授予BuiltinCatalog支持资格。

#### 历史普通v2回执与可重复入口

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

所有实施工作树原件已复制至原仓库 `.evidence/barness-ai-grok/` 并逐文件SHA256核对；11回放原件另经preservation/20261026T053443Z核对。v4索引只代表候选完成检查点；之后新增回放及资格源码须重新冻结，原件不改pin。用户未提交AGENTS.md修改和两个scratch目录保留。

#### 编译产物与尚未执行项

[候选构建](../../.evidence/barness-ai-grok/build/20261026T040254-candidate-6uadcnzf/build-receipt.json)与[只读门禁构建](../../.evidence/barness-ai-grok/build/20261026T040254-releasegate-el73wpu7/build-receipt.json)实际exit0，GOENV=off、GOFLAGS为空、隔离HOME/凭据；source/binary/catalog/SDK机器事实已固定。候选无开关/无Key精确入口实际为NOT_RUN、Calls0，不产生真实资格。

自身候选已在第3次运行完整真实PASS；前两次失败、授权/claim及原生拒绝结果保留。11回放与13 request-ID已完成；12正式软件已合入，15审查修复与14软件回归已完成；单case库存布局已修复，完整适用复验与实际原生source/E2E消费均通过。新的[17路线具体计划v2](formal-live-plan-v2.md)已准备，待完整普通通过后另行请求付费执行授权；六条Kimi仍为候选，race/压力未选择，release.pass=false。

回滚先禁用并更新xAI Binding版本，固定调用按原快照结束或由Host取消。保留Transcript、Native State Envelope、旧价格快照及验证原件，不自动切Provider/API；无数据库迁移。


#### 库存布局修复后的适用完整普通

仅两个E2E测试文件调整，v8独立两轴0/0；[本次原件](../../.evidence/barness-ai-grok/e2e/20261026T103441.672150Z-zab9zi6a/receipt.json)实际完整执行 `go test -json ./ai/e2e ./ai ./ai/internal/testkit/release -count=1 -timeout=20m`，128项冻结输入、1457必交/120入口全部PASS、辅助GLM32case全部PASS且每case最多393文件、audit0/exit0。独立[核验v6](../../.evidence/barness-ai-grok/final-ordinary-validation-v6.json)保留实际数字与墙钟；此前M3完整269测试PASS仅按其源码未变范围沿用，不伪造本次Go包范围。

原v5普通执行PASS与原生拒绝均保留；不放宽512文件上限、不删除原始4350文件。当前正式执行计划为[v2](formal-live-plan-v2.md)，四类native/binary来源未变，仍无新付费权限。


#### 本次正式17路线已授权并按失败规则停止

维护者明确回复“授权全部17条路线”，对应冻结计划v2。批次 `gf-20261026T112558Z-14ec1226` 的[原始授权与范围](authorization/gf-20261026T112558Z-14ec1226/execution-scope.json)保持不可变；本批原生入口、预算、代理和单路线凭据隔离均按已授权范围执行。新增164项输入的完整普通预检已实际通过，受测与付费HEAD均为 `5a8b7a9b3a4bb1f680a05ed13788dc035b51f426`。

- xAI Responses：[执行原件](../../.evidence/barness-ai-grok/live/xai-responses/20261026T113909.701356Z-hi0i7x_h/receipt.json) exit0，18/18场景、24调用/24HTTP；[原生正式消费](../../.evidence/barness-ai-grok/live-consumers/gf-20261026T112558Z-14ec1226/01-xai-responses/live-consumer.json) exit0/pass=true。
- OpenAI Responses：[执行原件](../../.evidence/barness-ai-grok/live/openai-responses/20261026T114333.867655Z-4st8edcu/receipt.json) exit0，9/9场景、15调用/15HTTP；[实际正式消费者](../../.evidence/barness-ai-grok/live-consumers/gf-20261026T112558Z-14ec1226/02-openai-responses/live-consumer.json) exit1/pass=false，原因已定位为非空AssistantContent捕获JSON解码缺陷，见[16修复工单](issues/16-generic-capture-decoding.md)。
- 按原始授权“任一路线失败或原生资格拒绝即停止后续路线”，批次状态STOPPED。其余15条NOT_RUN，未使用剩余grant、未重跑前两条；本批累计39调用/39HTTP。[不可变停止检查点](../../.evidence/barness-ai-grok/live-batches/gf-20261026T112558Z-14ec1226/stopped-checkpoint-v1.json)保存来源、逐条回执及其哈希。
- race、压力及发布仍NOT_RUN；live未完整有效，不能启动后续人工阶段，release.pass=false。先定位并准备可审查的软件修复，再按SOP重新冻结与取得所需的新执行选择；不以本批旧授权继续失败后的调用。

#### 正式消费者修复与新冻结计划

[16](issues/16-generic-capture-decoding.md)最小生产修改为`chat_capture.go`局部DTO，不改变公共领域JSON合同；保留归属、Observer、usage、预算及库存检查。原缺陷先在公开Audit/compiled CLI中RED，修复GREEN；v9指出的历史身份/终态/错误反例已补齐，使用公共Client实际loopback编码与Host/provenance/wire/CrossModel验证，共12/12定向case PASS、audit0，[v2原件索引](evidence/generic-capture-decoder-v2.json)。新反例在正确实现上直接GREEN，未虚构RED。

[独立v10两轴复核](review/review-v10.md)均0项。[静态清单](verification/cases-v7.json)独立确认1469主必交/121入口，119额外读取未变。[新计划v3](formal-live-plan-v3.md)及[机器计划](verification/formal-plan-v3.json)固定同17路线预算/顺序/代理与新原生source/binary；所有模板id空、expiry零，尚未发放本次权限。完整普通已实际通过：三包15,850 Go测试、1469主必交/121入口与32辅助case齐全，628.736秒/exit0/audit0；新付费执行不在旧停止批次的剩余授权内，不自动续用grant。race/压力及发布继续NOT_RUN。

#### 消费者修复后的完整验证与交付边界

实际受测`81bdaeafbd2b79429daba6964d9c4d24fb37ae1d`，命令为`go test -json ./ai/e2e ./ai ./ai/internal/testkit/release -count=1 -timeout=20m`，外层1500秒且显式128项输入；[实际receipt](../../.evidence/barness-ai-grok/e2e/20261026T122707.884222Z-3p40dgni/receipt.json) exit0，628.736秒。主manifest含15234 case（15020 PASS、214压力NOT_RUN），辅助GLM32 PASS/audit0且最大393文件。独立有限清单1469/1469、121入口全部完成；[核验v7](../../.evidence/barness-ai-grok/final-ordinary-validation-v7.json)保存实际数字及来源/输入比对。

[16软件票](issues/16-generic-capture-decoding.md)已解决，旧真实失败和合成RED/GREEN均保留。[v3计划独立准备核验](../../.evidence/barness-ai-grok/preflight/formal-plan-v3-independent-check/receipt.json)301项PASS/0错误，只证明配置和来源准备一致，不代表真实资格或新的授权。源码接线、软件修复与普通通过不构成正式支持；当前来源17条正式live均NOT_RUN，需[计划v3](formal-live-plan-v3.md)的新一次明确付费选择，之后才可分别选择race/压力。

当前只读判定使用[record-v9](verification/record-v9.json)与[readonly-v9](../../.evidence/barness-ai-grok/release/readonly-v9/native/release-report.json)保存实际source/E2E及未完阶段检查；最终release.pass=false。纯文档/索引收尾与实际受测提交的2054来源及128输入集合/内容相等性另存[核验v8](../../.evidence/barness-ai-grok/final-source-equality-v8.json)，不改旧receipt的pin、时间或状态。

#### 正式计划v3已获本次执行授权

维护者明确回复“授权 v3 全部17条”；新批次`gf3-20261026T131257Z-6a9ba85b`的原始回复、17份新grant/config与[执行范围](authorization/gf3-20261026T131257Z-6a9ba85b/execution-scope.json)已留存。新164项输入先取得同来源完整普通预检，再按冻结17路线串行、失败/取消/超时/原生资格拒绝即停。当前本批尚无厂商请求或claim；旧批次及候选grant不复用。race/压力及发布仍未授权。
