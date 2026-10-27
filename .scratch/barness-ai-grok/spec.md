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
| 实现 | 01–04及07–22软件已解决；21完成GLM预算/取消子树，22严格补齐工具合成证明与私有store边界。05真实资格、06人工阶段及发布未完成 |
| 冻结来源 | 软件审查pin a80bfe1b；完整实际受测dcf73abd6740d2107352909ba3cd3d07ad6684fb，2068默认来源/175输入。四native完整map与八镜像、原17 grant/config/auth/index均未改变；原v8权限未用，验证来源变化须明确补充同意 |
| 审查 | [v17](review/review-v17.md) Standards0/Spec0，原v16私有store符号链接P2已关闭；完整工具52/52 PASS。冻结前244项核验0未解决，操作脚本字段/哈希绑定P2在执行前修复；旧失败/审查原件均保留 |
| 当前完整普通（cases-v13） | 四包16,038个Go测试PASS、1572主必交/130入口、GLM32及MiniMax六辅助全部PASS，665.060秒/exit0/audit0；主15122 PASS、214压力NOT_RUN/max131原件。[receipt](../../.evidence/barness-ai-grok/e2e/20261027T030431.092824Z-681i69ck/receipt.json)/[有限核验](../../.evidence/barness-ai-grok/live-batches/gf8-20261027T012011Z-d0c21717/ordinary-validation-glm21-preconsent.json)；旧165输入FAIL不改 |
| 历史完整普通v3 | 采集器PASS/exit0；两个Go包15402测试PASS；主包14586个原生case PASS、214个opt-in压力/消费者case NOT_RUN；必需565/565 PASS、65入口齐全、audit0。v1/v2与手工校验失败原件均保留 |
| 原生消费 | [索引v17](verification/record-v17.json)绑定本次实际完整普通、175输入及唯一状态；[实际只读报告](../../.evidence/barness-ai-grok/release/readonly-v17/native/release-report.json)另存。live/race/压力缺少，release.pass=false |
| 自身live | 历史v6源507自身18/18真实PASS保留；当前v8全部17路线仍NOT_RUN，0厂商请求/0真实claim，原权限未用且不能自动重绑新验证来源 |
| 共享历史live | 原v8一次17授权仍未使用；本次仅[来源绑定补充提案](formal-live-source-amendment-v8.md)，不申请第二批付费额度。先175完整PASS、一次明确来源同意后新增两治理输入并取得177匹配普通，才可原17次。历史正式511逻辑/507HTTP保持，Kimi六条仍候选 |
| race / 压力 | 均NOT_RUN；尚未人工选择，须先满足适用live后依序选择；压力适用，普通运行的214个opt-in场景未计为PASS |
| 目录/发布 | 21/22软件、v17双轴与175完整普通已完成；尚待原未用v8权限的验证来源补充及177匹配普通、17原生正式资格、人工race/压力，仍未正式支持或发布 |

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

#### v3实际停止与资格合同修复

新批次已在第4条的真实正式消费者拒绝后停止，前三条正式资格通过，后13条NOT_RUN。OpenAI Chat的`reasoning-history`依既有`ai/live/combos_test.go`/`scenarios_test.go`合同为不支持、零请求；支持矩阵原生保留UNSUPPORTED语义，当前release资格层却要求所有场景正调用及PASS。已建立[17修复工单](issues/17-declared-unsupported-qualification.md)，只修复合同一致性并保留该能力不支持，不伪造成功、不扩大型号/重放能力。所有真实捕获、消费者拒绝、原始授权及65次调用由[停止索引](../../.evidence/barness-ai-grok/live-batches/gf3-20261026T131257Z-6a9ba85b/stopped-checkpoint-v1.json)固定；后续新source/付费需重新冻结与选择，当前授权不续跑。

#### 17资格合同修复与v4冻结前沿

原生已声明的不支持场景保持UNSUPPORTED/无新LastPassedAt，精确固定身份/原因、零活动及仅assertions原件均验证；伪造PASS、错误状态或任意其他能力借例外都会拒绝。Generic6当前目录分支审计证明只有OpenAI Chat固定noReplay可达。生产验证归属supportmatrix既有合同模块，删除release旧实现，不扩公开ai API。

原报告合同RED与伪造PASS RED保留；[完整CLI证据](evidence/generic-unsupported-cli-v1.json)用11场景原生包和独立测试Git命名空间验证真正接线，三份实际RED分别证明原严格规则/报告接线/库存接线可被回归捕获，61项定向GREEN/audit0。合成全体release=false，未调用Claim或真实厂商/真实store。v11两轴P2经[v12](review/review-v12.md)独立复核均关闭。

[计划v4](formal-live-plan-v4.md)及[机器计划](verification/formal-plan-v4.json)固定新source/8个image/旧native预算及明确237 PASS+1固定UNSUPPORTED/零请求合同。129输入中119执行输入未变，额外绑定修改的原生合同文档。完整四包普通已实际PASS：15,982测试/1518必交/125入口与32辅助齐全、626.207秒、exit0/audit0；旧批次104次正式调用与13未跑项原样保留，不自动续用旧权限。

#### 原生限制修复后的完整验收与交付边界

完整四包命令`go test -json ./ai/e2e ./ai ./ai/internal/testkit/release ./ai/internal/testkit/supportmatrix -count=1 -timeout=20m`，外层1500秒/显式129输入，实际受测edd01fab。完整15,982测试PASS，主15283case（15069 PASS、214压力NOT_RUN）、辅助GLM32 PASS，所有原件audit0；[核验v8](../../.evidence/barness-ai-grok/final-ordinary-validation-v8.json)记录1518必交/125入口逐项通过及来源/输入比对。supportmatrix新增包79个Go测试实际PASS，不把历史M3来源不变材料冒作当前原生资格。

[17软件票](issues/17-declared-unsupported-qualification.md)已解决，完整公开CLI回归与三份实际RED由[证据索引](evidence/generic-unsupported-cli-v1.json)固定。[计划v4独立准备核验](../../.evidence/barness-ai-grok/preflight/formal-plan-v4-independent-check/receipt.json)348项PASS/0错误，只证明配置一致，不是调用授权或真实资格。当前新来源17正式live均NOT_RUN，[计划v4](formal-live-plan-v4.md)已获新一次明确授权，待新165输入完整普通通过后执行；原批次失败停止、不续用其13份剩余grant。

[record-v10](verification/record-v10.json)与[readonly-v10](../../.evidence/barness-ai-grok/release/readonly-v10/native/release-report.json)保存当前只读事实；正式live/race/压力尚未完成，release.pass=false。纯文档/索引收尾与实际受测edd01fab的2057来源及129输入集合/内容相等性另存[核验v9](../../.evidence/barness-ai-grok/final-source-equality-v9.json)。不改旧执行pin/时间/状态，不将UNSUPPORTED当作PASS或推理回放能力。

#### v4本次正式live授权

维护者明确回复“授权 v4 全部 17 条路线”；新批次`gf4-20261026T184225Z-e2beb101`的[原始授权](authorization/gf4-20261026T184225Z-e2beb101/authorization.json)与[17路线执行范围](authorization/gf4-20261026T184225Z-e2beb101/execution-scope.json)已保存。默认2057来源文件不变，新增36项授权/config/grant/index使选定输入为165；先取得同提交/输入的完整四包普通，再依固定顺序执行一次各路线与原生资格消费者。新批次当前NOT_RUN，未读取凭据、未取得claim或请求厂商；失败即停，不续用v3剩余权限。历史record-v10及readonly-v10只代表授权前检查点；race/压力及发布仍未授权，release.pass=false。

#### v4实际停止及无密钥启动修正

前5条新来源正式资格通过，共58 PASS及1项固定无请求UNSUPPORTED。智谱在仓库根初始化时读取相对testdata失败，后续入口检查覆盖更早的配置错误；未发HTTP、未领取grant，原生执行仍失败，按原授权停止后11条。完整[停止检查点](../../.evidence/barness-ai-grok/live-batches/gf4-20261026T184225Z-e2beb101/stopped-checkpoint-v1.json)保存所有原件，原进度及失败不修改。

现有env -C将native工作目录设为ai/live、binary/grant解析同文件绝对路径即可修正；[17条无Key/无grant初始化](../../.evidence/barness-ai-grok/preflight/gf4-20261026T184225Z-e2beb101-cwd-keyless-6em78m4s/receipt.json)全部PASS、0调用/0claim，不产生正式资格。源2057、执行输入165及HEAD均未改变，未修改源码或重新购买。

[具体同源续跑提案](formal-live-continuation-v4.md)只保留前5、对第6追加一次进程并继续后11：179必交PASS、最多292calls/320HTTP、6小时原生上界、原grant到2026-10-27 14:42（上海）过期。它明确要求人批准SOP4失败授权/原未claim grant及一次进程例外、解除STOP/精确cwd启动修正，以及仅补充计划与原始人类回复不加入165输入的SOP3治理回执例外。目前未授权、未续跑；若拒绝这些例外，不能混用不同输入回执。race/压力与发布仍未授权。

#### 限定续跑已授权及MiniMax实际停止

维护者明确回复“批准限定例外，续跑剩余 12 条”；[原始补充授权](../../.evidence/barness-ai-grok/live-batches/gf4-20261026T184225Z-e2beb101/continuation-authorization-v1.json)固定允许仅治理计划/原始回复不加入165输入、第6额外一次使用原未claim grant及解除STOP。没有改旧grant/config、source、pin或原失败记录，也未重复前5。新智谱24 PASS/24calls、ZAI20 PASS/20calls均取得正式资格。

第8国际MiniMax-M3 Chat实际14逻辑calls/13HTTP、21.045秒，14场景中13 PASS（含xai-origin-history），thinking-on FAIL；其真实请求thinking.type=adaptive/reasoning_split=true/2048输出，服务器HTTP200直接回答42且原始reasoning_tokens=0，无法满足自身“实际观察推理”正例。另一full-stream原始8 reasoning tokens及Thinking被正确保留，当前未见适配器丢失证据；推断探针任务可能过于简单，尚未修正或重新验证，保留FAIL不降级。完整[停止检查点](../../.evidence/barness-ai-grok/live-batches/gf4-20261026T184225Z-e2beb101/continuation-1/stopped-checkpoint-v1.json)保存ledger/claim/原件和9条NOT_RUN；没有自动重试或启动第9条。后续只做离线诊断，不新增付费权限。

#### 18/19离线修复与v13前沿

[18](issues/18-minimax-reasoning-probe.md)仅为M3 thinking-on追加固定有界CRT任务、继续要求自身非空推理，原模型/场景/预算/消费者不变；公开受控旧任务RED、新正例/零推理反例GREEN原件由[摘要v1](evidence/minimax-reasoning-probe-v1.json)固定。新增wrapper主1 case、native auxiliary6 case，不能将其计为真实资格。

18额外选择既有FreshLedger回归时漏查Git namespace隔离，实际追加3 synthetic ID/6 guards/3 ledgers；无真实Key或厂商请求，记录原样保留，根核对原件见[偏差核验](../../.evidence/barness-ai-grok/diagnosis/controlled-store-v1/receipt-v2.json)。[19](issues/19-controlled-grant-store-isolation.md)改为每用例私有Git common store，HEAD/object来源与真实checkout不伪造，先只读Git RED后私有GREEN；[原摘要v1](evidence/controlled-grant-isolation-v1.json)也保留。

[v13两轴](review/review-v13.md)指出旧/新增重启检查复用exec.Cmd，第二CombinedOutput在启动前报错，库存不变不能证明原生单次权限拒绝。原实现者正先补真实数字exit的RED，再用新命令对象及同私有namespace执行实际restart；不改旧v1和原restart原件，完成后新摘要和独立复核另行版本。当前尚未完成，不发放新付费或人工权限。

#### v14零发现与v5冻结准备

18/19以及v13真实restart缺口已完成：定向真实RED/GREEN均保留，新[私有授权v2](evidence/controlled-grant-isolation-v2.json)绑定实际第二native PID/runID、数字exit1、原生已消耗拒绝及零新增调用/四文件digest；[v14](review/review-v14.md)两轴均0。旧v1错误证明已撤回而不改原件。新[1522有限清单](verification/cases-v11.json)、[独立库存](verification/cases-v11-inventory.json)、[119实际外部输入](verification/ordinary-inputs-v8.json)由源码枚举，不以PASS列表反推。

[计划v6](formal-live-plan-v6.md)沿同17路线/237 PASS+1声明UNSUPPORTED合同，固定新8个image及正确package cwd/绝对grant路径。原始241/240正式调用及全部停止点保留；当前无新付费权限。先完成此129输入/2062来源的完整四包普通，再交付具体 paid选择。

#### 计划scope哈希元数据修正与v6

v5独立准备核验765项发现11条重复route.scope.binarySha256仍旧，实际producer/inactive模板/8个新image始终正确。旧129输入完整普通在274.559秒由root主动SIGTERM取消，原生exit-9/CANCELLED及原件保留，没有厂商调用；相关Go进程组已清空。机器/人类v5和核验不覆盖。

新[计划v6](formal-live-plan-v6.md)直接从实际模板生成11 scope，与v5仅这些binary字段值改变；代码2062、119执行输入与8个image不变。独立[867项准备核验](../../.evidence/barness-ai-grok/preflight/formal-plan-v6-independent-check/receipt.json)全部PASS/0发现，只证明准备一致，不产生普通/真实资格或付费权限。新的129集合仅替换两计划路径，先重新运行匹配完整普通。

#### 修订版完整普通实际完成

实际受测dcfa262278499aaf04f5153312f8672980880f51/2062来源/129输入；完整四包15,987 Go测试PASS，1522主必交/127入口与GLM32、MM18六辅助全部PASS，633.015752秒/exit0/audit0。主15287 case中15073 PASS/214压力NOT_RUN；辅助32+6全部PASS。主/GLM/MM18每case最大131/393/3原件；没有扩大512上限。完整[实际核验](../../.evidence/barness-ai-grok/final-ordinary-validation-v9.json)固定所有bundle/manifest/audit/source/input哈希。v5元数据取消原件和所有历史失败继续保留。

18/19软件票在上述完整普通与v14双轴0/0后关闭；真实MiniMax零推理FAIL和7条旧资格仍是历史，当前新来源没有正式live权限或资格。当前[计划v6](formal-live-plan-v6.md)准备核验867 PASS，模板ID空/expiry零；待维护者明确选择一次固定17的范围及失败处理策略，先生成新grant/config/授权并取得匹配普通后才可付费。race/压力后续分别人工选择，未正式支持或发布。

#### v6 一次17路线全部执行与已定位余项

维护者明确选择“授权17条，收集全部独立结果”。原始回复与新36授权/config/grant/index保存在[本批授权](authorization/gf6-20261026T233415Z-ceb8d8c1/authorization.json)，有效期间17条均按顺序只运行一次；匹配普通、实际消费者、调用与原生预算原件完整保留。整批14正式资格通过、2资格器拒绝、1原生失败，未重发权限或重跑。独立总账实际exit1表示完整已知失败，integrityPass=true/0errors；前版审计脚本误把Markdown当JSON的失败原件同样保留。

- 国际MiniMax Chat与Responses各14原生场景全PASS，thinking-on自身非空推理通过；正式资格器在验证fresh ProcessGrant之前无条件读取legacy one-shot origin，旧sourceCommit与本批507不符，真实exit1。[20工单](issues/20-minimax-fresh-authority-dispatch.md)仅离线修复权威分派，不能改旧origin或据原生PASS冒称正式资格。
- 国际MiniMax Messages thinking-on的HTTP200流已有并保留非空推理，随后读取在可见正文/成功终态前中断，保持FAIL；底层vendor/代理/网络原因无法由原件确定。thinking-control初始unexpectedEOF已按原合同rerun1恢复，不能把它的错误原文套给thinking-on。精确诊断见[独立原件](../../.evidence/barness-ai-grok/live-batches/gf6-20261026T233415Z-ceb8d8c1/read-only-diagnoses/receipt.json)。不降低推理或成功终态门槛，不擅自增加重试。
- 238必交实际236 PASS/1固定声明UNSUPPORTED/1 FAIL；逻辑270/HTTP267。进度消息曾口算HTTP多1，检查点与原始预算始终是267，未改旧原件。与此前正式241/240合计511/507，候选费用另计。

所有live原件仍绑定507/2062/165；20改变来源后必须独立评估适用性，不能预报新源合格。本次授权已全部消耗，没有自动付费重跑、race、压力或发布权限。

#### MM20 完整普通闭环与 v8 待授权

20只修复MiniMax正式消费者的权威分派：fresh ProcessGrant材料始终经过完整原生校验，不依赖旧legacy origin；没有fresh材料的历史路径仍保持原origin校验。任何fresh marker都受其源、binary、scope、claim/namespace、账本及完整场景约束，foreignmodel或混合renewal不能绕过。公开CLI先保留三个协议实际RED，再得45新场景GREEN；既有受影响回归53主/6辅助通过，v15两轴各0。新源完整普通又实际通过16,033 Go测试、1568主/129入口/38辅助；20软件据此关闭，不产生真实支持资格。

[计划v8](formal-live-plan-v8.md)保持既定17路线、模型、SDK、完整238场景与预算，收集已确定独立结果、每路线一次，不新增进程重跑或备用权限。v7尚未执行的提案将付费授权放在SOP3完整普通前，独立核查P2后保留原稿及FAIL；v8严格先取得本次129完整普通，再提新付费授权；若批准，再生成新36授权/config/grant/index，冻结实际新提交/165输入并完成同范围完整匹配普通，之后才可发任何厂商请求。v8有限核验830项0发现，源/镜像/范围未因文案排序修正改变。

本次普通实际受测10326/2064/129；构建E1的全部相关源码/实际119读取字节与受测来源匹配，单列原生合同文档新hash固定。三个原生来源map改变、GLM342 map未变；全局Stage来源/选定输入已经变化，旧v6的14资格仅保留历史。当前新源17正式live均NOT_RUN，race/压力尚未人工选择，发布保持false。原真实失败、transport中断及所有已消耗grant/ledger不修改，不借本软件闭环自动重买路线。

#### v8 授权后普通失败，厂商执行未启动

原始v8回复、36新授权/config/grant/index与11份keyless scope检查全部留存；实际654750/2064/165完整普通以exit1结束，两个公共GLM包装在10.01秒失败。直接native32个断言PASS但未完成manifest/audit，有写入发生在父包装失败之后；内嵌包装完成9个formal case后进入验证。共享ctxFor只有10秒，不能据此宣称全耗于Go编译，也不能把未定稿辅助包当通过。此前同源码授权前129完整PASS/包装8.69秒仅说明余量很窄。

[21](issues/21-glm-controlled-process-budget.md)先做有界受控RED，修正完整受控子进程预算与真正子树终止边界，保持GLM32辅助归属/512上限和严格普通门槛；不预热缓存掩盖失败，不全局放宽公共调用context。真实普通与两级不完整native原件都保留。四原生map与8镜像仍逐字节匹配；测试-only改动是否继续保持需以实际diff独立核验。

本批全部17厂商路线NOT_RUN，真实vendor调用/真实claim均0，原grant和expiry不改；不能把普通失败推断成付费额度已花，也不据原批准自动重绑变化的全局来源。修复后的完整普通与授权适用性另按SOP明确判断后，才可进入厂商阶段。历史正式511/507不变，race/压力及发布未授权。

#### 21/22 完整普通闭环与 v8 未用来源绑定

21通过真实公共过程RED修复短调用context截断完整native工作及收尾，并在外层取消时给予内层owner有界清理窗口，非合作子树强制终止；新增4项主case，原32项GLM辅助及每case512上限保持。22只更新Google七场景合成证明和测试专用类型桥；缺grant、标签占位、provenance/downgrade篡改及目录逃逸均严格拒绝。v16发现store根符号链接P2后，先在完全测试拥有目录复现外写，再于输出/Claim前拒绝；v17独立两轴均0。完整52项Python过程实际通过，作为独立工具范围旁证，与主Go bundle分开。

本次实际受测dcf73abd6740d2107352909ba3cd3d07ad6684fb/2068来源/175输入，16,038个Go测试、1572主必交/130入口、32+6辅助全部PASS，665.060秒/exit0/audit0。原始654750/165普通FAIL、所有RED/INVALID_EVIDENCE/CANCELLED及真实已停止批次均保留。当前全局验证来源变化；四native map完整SET/内容与八个实际镜像仍相同，不等于自动取得来源重绑或本源真实资格。

[来源补充提案](formal-live-source-amendment-v8.md)不新发或延长权限：保留原v8尚未使用17个grant及原21:20到期、每路线一次/同一444逻辑调用472HTTP预算。先完成本次175普通，再取得一次明确来源同意，新增原始回复与最终选择两项后冻结177并完成匹配普通；结果作输出不再改冻结输入，不再索要确认PASS。尚无此补充同意，厂商请求/真实claim均0。原授权和17配置/ID/expiry未改，历史正式511/507不变；race、压力和发布仍未授权。
