# barness-ai：多 LLM Provider 协议中间件

Status: ready-for-agent
Updated: 2026-10-02

本文是 barness 基础模块 barness-ai 的目标规范，综合既有需求、设计和 E2E 研究，并按本项目职责和工程原则重新定稿。状态表示可据此拆分实施任务；不表示代码、协议兼容或真实 API 验收已经完成。本次工作仅交付规范。配套的 [研究追溯表](research-traceability.md) 将原研究条目、当前契约、协议专项和验收场景逐项关联；实现以本规范为准，来源用于核查，不以“按基线处理”替代已经明确的行为。

## Problem Statement

barness 需要在本地程序和云端 Agent 环境中使用多种 LLM Provider。不同服务的消息、系统指令、工具调用、推理状态、图片、流事件、错误和用量格式不同；如果由各个上层模块分别处理，会产生重复适配、历史回放不一致、错误终态遗漏和依赖厂商 SDK 类型的问题。

云端场景同时存在多个租户和用户。同一 Provider、模型或 binding 名称可能对应不同租户的账户与 key。按 Provider 缓存凭据、从进程环境隐式兜底、复用可变鉴权状态或缺少原生状态归属检查，都可能串用账户和泄露数据。仅提供一个统一 HTTP 客户端不能解决这些问题。

既有研究按冻结 pi-ai 的行为给出了基础方案，但其定位仍是独立项目，部分范围和资源策略尚停留在提案阶段。barness 需要一份面向项目内基础模块、职责清晰、可直接实施并可重复验收的规范，供后续 Agent、Session、本地入口与云端宿主共同使用。

## Solution

交付可嵌入 Go 程序的 barness-ai，作为 barness 面向多种 LLM Provider 的协议中间件。宿主提供可信调用作用域、授权服务绑定和本次上下文；模块完成校验、凭据解析、确定性的协议转换、请求与流处理，返回统一事件、最终 assistant 消息、错误分类和用量。

一次调用处理一个生成轮次，可以接收多轮历史，但不拥有会话生命周期。工具声明、调用和结果属于协议数据；执行工具、组织下一轮上下文并再次调用模型属于宿主。依赖方向是 barness 上层模块依赖 barness-ai，基础模块不反向依赖 Agent、Session、TUI、数据库、任务队列或部署产品。

首期采用既有研究建议的范围：OpenAI Responses、Anthropic Messages、Gemini Developer API、OpenAI Chat Completions 四类协议，验收 OpenAI Responses/Chat、Anthropic Messages、Gemini Developer API、DeepSeek Responses/Chat 六个 Provider×协议组合。保留图片输入，使用直接厂商 API key；其他官方云 API、网关和生成模态按后续独立范围处理。

兼容基线冻结为 pi-ai `0.87.1`、commit `898ab804050730e9dcefb4443875d5a932aa6a32`。对齐已纳入范围的可观察行为；租户授权、可信配置、Go 并发访问和显式资源限制作为单独扩展验收。既有研究中的 SDK 探针仅作为选型证据，不能证明 barness-ai 已通过验收。

## User Stories

1. As a barness 模块开发者, I want 通过统一 Go 调用契约使用不同 LLM Provider, so that 上层业务不必理解每家厂商的请求和流格式。
2. As a 本地应用开发者, I want 使用与云端相同的协议模块并显式装配本地租户和 key, so that 本地行为可迁移到云端环境。
3. As a 云端宿主开发者, I want 传入经认证和授权的租户身份, so that 每次生成都归属正确的租户。
4. As a 租户管理员, I want 同一租户配置多个服务绑定, so that 生产、开发或不同厂商账户能够独立管理。
5. As a 租户用户, I want 同名 binding 在不同租户下独立解析, so that 不会使用其他租户的凭据和模型权限。
6. As a 宿主开发者, I want 缺失身份、越权模型和无效凭据在发推理请求前失败, so that 拒绝请求不会误用平台或环境 key。
7. As a 租户管理员, I want 新调用取得更新或撤销后的凭据状态, so that key 管理具有明确的生效语义。
8. As a 运维人员, I want 在途调用及其重试固定使用同一配置快照, so that 一次生成不会在中途切换账户身份。
9. As a 模型接入开发者, I want 分别表达 Provider、API 协议和模型, so that 不同厂商可以复用协议而不混淆实际服务身份。
10. As a 模型接入开发者, I want 同一 Provider 的多种协议由明确绑定选择, so that 请求不会因失败而隐式切换协议。
11. As a 应用开发者, I want 提供本次已选定的消息历史和系统指令, so that 模块只完成协议转换而不改变我的上下文策略。
12. As a 应用开发者, I want 保留 system 和工具声明变化的输入顺序语义, so that 回放后的有效指令和工具集符合既有 pi 行为。
13. As a 应用开发者, I want 发送模型支持的图片及工具结果图片, so that 图像输入可以参与同一个生成流程。
14. As a 应用开发者, I want 不支持图片的目标按基线生成占位文本, so that 历史转换具有可预测的降级行为。
15. As a Agent 开发者, I want 获得工具调用 ID、名称、参数和增量, so that 宿主能够关联工具执行及下一轮结果。
16. As a Agent 开发者, I want 部分工具 JSON 与完整可执行参数有明确区别, so that 不会执行截断或尚未校验的参数。
17. As a Agent 开发者, I want 显式使用工具 JSON 修复和 schema 校验助手, so that 参数处理可以复用而执行决策仍由宿主管理。
18. As a 应用开发者, I want 同模型回放完整保留适用的签名、redacted 块和原生续接状态, so that 推理连续性不会因序列化丢失。
19. As a 应用开发者, I want 同租户切换模型或 Provider 时自动转换可见推理与工具关联, so that 不必手工删除不适用状态。
20. As a 宿主开发者, I want 从持久化恢复原生状态前验证租户与账户归属, so that 客户端自报来源不能获得回放权限。
21. As a 应用开发者, I want 使用 simple 入口表达统一 reasoning 等级和预算, so that 常见推理控制不必按厂商分别编写。
22. As a 高级应用开发者, I want 使用按协议区分的完整选项, so that 已支持协议能力不会被最低公共参数集限制。
23. As a 应用开发者, I want 保留未设置、显式 null 和零值的区别, so that 参数编码、覆盖和默认值符合基线。
24. As a 应用开发者, I want 获得类型清晰且顺序稳定的文本、推理和工具流事件, so that 可以构建增量展示与消费流程。
25. As a 应用开发者, I want 不读取流事件也能等待最终结果, so that 仅需要最终消息时不会产生消费依赖或死锁。
26. As a 应用开发者, I want 同时安全读取累计响应视图和等待最终消息, so that 并发展示不会访问正在无同步修改的数据。
27. As a 应用开发者, I want 设置失败、鉴权失败和流中断仍返回最终 assistant 消息, so that 已收到的内容和停止原因不会丢失。
28. As a 应用开发者, I want 区分正常停止、长度截断、错误和取消, so that 不会把不完整生成误当成功或执行未完成工具。
29. As a 宿主开发者, I want context 取消和 Close 真正终止本次 I/O, so that 下游断开时可以释放连接与准入许可。
30. As a 租户用户, I want 我的取消、错误和超时不影响其他租户的调用, so that 共享 Client 仍保持隔离。
31. As a 运维人员, I want 明确配置重试次数、延迟与截止时间, so that 模块不会无边界重放请求或增加未知费用。
32. As a 应用开发者, I want 流开始后的失败保留部分结果并结束, so that 不会因自动重放收到重复内容或工具调用。
33. As a 宿主开发者, I want 请求、图片、流帧、工具参数、事件队列和输出均受显式限额约束, so that 慢消费者及异常服务不能使进程资源无界增长。
34. As a 宿主开发者, I want Complete 只等待最终结果而不积累无人读取的事件队列, so that 非流式使用具有清晰的资源成本。
35. As a 可信宿主开发者, I want 在指定时点使用请求体和响应元数据回调, so that 可以完成必要的协议扩展与集成。
36. As a 租户管理员, I want 非可信生成请求不能覆盖 key、endpoint、代理或认证头, so that 凭据只发往已授权目标。
37. As a 运维人员, I want 日志、错误和默认观测不包含 key、正文和原生密文, so that 诊断不会扩大敏感信息暴露范围。
38. As a 运维人员, I want 按租户、逻辑调用和请求尝试追踪状态, so that 可以定位重试、取消和用量归属。
39. As a 宿主开发者, I want 观测拥塞有界且不会改变模型结果, so that 观测系统故障不会破坏推理流程。
40. As a 成本管理开发者, I want 每次尝试明确记录用量完整性和价格版本, so that 未上报用量不会被解释为免费。
41. As a 成本管理开发者, I want 保留缓存读写、推理 token 和阶梯价格语义, so that 估算成本与兼容基线一致。
42. As a 云端宿主开发者, I want 使用内置的租户与进程并发限制，并通过注入的准入接口按厂商账户聚合准入, so that 单进程资源限制不会被误认为分布式配额或账户配额。
43. As a 应用开发者, I want 查询版本化模型能力并应用授权范围内的覆盖配置, so that 模型选择不依赖隐式在线发现。
44. As a 模型接入开发者, I want 新 Provider 按协议差异复用现有实现和测试场景, so that 不必复制整个调用流程。
45. As a barness 维护者, I want 在公开 Client 边界运行离线 E2E, so that 测试证明目标模块的行为而非 SDK 自身可用。
46. As a barness 维护者, I want 同一批场景与冻结 pi-ai 比较请求和结果, so that 兼容差异能够定位、解释并阻断发布。
47. As a barness 维护者, I want 真实 API 冒烟与离线测试分开运行, so that 日常验证无网络和真实 key 依赖，而发布仍有厂商接入证据。
48. As a barness 维护者, I want 每次验收保存可重放且脱敏的证据包, so that 其他维护者能够复核结论。
49. As a barness 上层模块开发者, I want 最小本地、宿主接入和工具往返示例, so that 能理解身份、历史、取消与执行职责的交接。
50. As a barness 维护者, I want 模块名称、文档和测试开关统一使用 barness-ai, so that 项目中不会出现旧工作名的并行入口与兼容层。

## Implementation Decisions

### 1. 定位、依赖与交付形态

- 本项目统一使用 **barness-ai** 表达该基础模块；Go 包标识采用合法 Go 标识符，模块名称不强迫标识符包含连字符。不以旧工作名建立独立产品、旧入口或兼容 shim。
- 交付统一调用 Client、公共领域类型、协议 adapters、历史转换、工具 JSON 助手、版本化模型目录、本地装配与宿主接入示例。Go module 的具体拆分不构成行为契约，不因原研究的独立仓库定位额外创建独立仓库或发布流程。
- 核心负责调用生命周期；协议 adapter 负责目标协议转换和响应归一化；宿主提供绑定、凭据、准入及观测能力。只为实际依赖边界定义精简接口，不预建通用插件系统、HookRegistry 或大量空接口。
- 协议模型与公共领域模型在 adapter 边界独立转换；SDK 类型、持久化记录、HTTP DTO 不泄漏到公共消息和事件。宿主自己的协议/持久化模型在宿主边界校验并转换。
- 复用不包含可变租户状态的实现与连接池；调用输入在接收时取得必要副本，返回集合及快照不能泄漏内部可变存储。调用方不得在交接期间并发修改输入；之后修改源数据不改变在途调用。
- 实现前核查已有依赖及候选 SDK 的文档、类型和公开扩展点。单文件不超过 500 行，以职责边界拆分；不因代码外形相似提前抽象。

### 2. 公开调用与标识契约

| 概念 / 入口 | 契约 |
| --- | --- |
| Client | 构造后配置只读、并发安全；所有 credential-dependent Provider 操作均经过该可信调用路径 |
| CallScope | TenantID、RequestID 必填；ActorID、JobID 可选，由可信宿主建立；身份显式传入，context 负责取消、截止时间与追踪 |
| Target | BindingID 与 ModelID；协议、endpoint、账户和凭据由授权绑定确定 |
| Request | 本次 Messages、可选 SystemPrompt 便捷输入与 Tools；SystemPrompt 按基线归一到初始 system 消息，不管理跨调用历史 |
| Stream / Complete | 使用与绑定 API 匹配的完整协议选项，不强制通过 simple 映射 |
| StreamSimple / CompleteSimple | 使用统一选项，按模型能力执行原有 reasoning 和预算映射 |
| Stream | 提供 Next、Event、Result、Err、Close；Next/Event 单消费者，Result 可独立或并行等待，Close 可并发且幂等。Err 采用 Scanner 风格：Next 返回 false 后返回与 Result 相同的错误，终态前为 nil |
| Result | 最终 AssistantMessage、不可变调用归属和附加错误/计量信息；普通请求或 Provider 失败仍有消息，不以 nil 结果替代 |
| Complete / Stream.Result 签名 | 返回 `(Result, error)`；StopReason 为 error 或 aborted 时 error 非 nil，且 Result 仍完整有效；成功终态时 error 为 nil |
| Error | 保留消息终态与 errorMessage，同时提供可用 errors.Is/As 访问的稳定 Code/Phase；不暴露秘密后端内部细节 |

RequestID 表示一次公开逻辑调用，内部重试沿用；Complete 内部复用流处理不生成第二个逻辑 ID。每次实际 HTTP 尝试具有独立 AttemptID；厂商 request ID 单独保存。宿主再次调用库应分配新 RequestID，不能将其等同于 Agent RunID、TurnID、JobID，也不据此承诺请求去重或厂商幂等。

RequestID 由宿主生成并保持足够唯一；ActorID 表示经授权的用户或服务主体，不是 worker 实例，JobID 只用于关联宿主作业。一个 Job 可以经过多次逻辑调用及工具往返；库只传递这些信息，不创建或调度 Job。完整入口不额外建立按厂商名索引的选项命名空间。

Result 与流提供不可变调用元数据，至少包含 TenantID、RequestID、BindingID 和解析成功后的实际 ProviderID、API、ModelID；用于保存或恢复原生状态的消息封套另含 AccountScopeID。ActorID/JobID 在提供时保留。预检失败前尚未解析出的真实模型或账户标记为未解析，不能把请求自报值冒充已授权身份。

宿主汇合多个流时，每个事件必须通过信封或等价的不可变引用携带上述调用归属，不依赖全局“当前租户”或事件到达顺序推断。该信封是 Go/租户扩展，事件内容与 pi 差分分开；原生状态封套不是仅凭字段即可信任的授权凭证。宿主控制对外响应暴露哪些内部元数据。

### 3. 租户授权、服务绑定与配置快照

- Binding 由 `(TenantID, BindingID)` 定位，包含版本、ProviderID、API、Endpoint、AuthKind、AccountScopeID、CredentialRef、AllowedModels 和策略。首期 AuthKind 只支持 API key；一个 binding 固定一种 API。
- BindingResolver 验证归属、调用主体权限及模型许可；CredentialResolver 根据该绑定获取含 OwnerTenantID、CredentialID、Version、有效状态和内部秘密值的凭据快照。秘密值不进入公共消息、事件、Result、日志或错误。
- 调用顺序是：检查 scope/请求结构与大小 → 解析授权绑定 → 检查能力、选项和历史来源 → 解析凭据 → 校验租户、账户、引用及版本一致性 → 获取本次尝试的准入许可 → 构造请求与发送 → 收束结果并释放资源。
- 两次解析之间发生撤销或更新且无法取得一致快照时，首期在发请求前失败，不拼接不同版本配置，也不在同一逻辑调用内自动重新解析。这是决策 D2 对原研究“重新解析或失败”的明确收窄。缺绑定、越权、禁用、缺 key、秘密后端故障均不得发出 Provider 推理请求或回退到其他身份。
- 普通生成请求不可包含 key、凭据引用、endpoint、Authorization、代理、云项目身份或可执行函数。云端宿主对可提交的协议选项设允许范围；完整选项和回调产生的最终模型/原生引用还需授权复核。
- 首期 endpoint 来自已验收 Provider 配置。管理操作扩展目标时必须验证允许的 HTTPS 目标与凭据绑定关系；本地测试专用 loopback HTTP 由测试装配提供，不作为生产请求覆盖能力。
- 核心不缓存明文 key 或“已认证的当前 Provider client”。本地宿主可以显式读取指定环境变量或秘密文件再注入；核心及 SDK 均不得自行使用环境 key、endpoint 或 ambient identity 兜底。
- 每次新逻辑调用重新解析；K1→K2 更新后新调用用 K2，在途调用及内部重试继续使用原快照。禁用阻止后续新调用；立即终止全部在途请求由宿主索引并取消 context，不承诺撤回厂商已接受的计算。
- 首期示例解析器无缓存；宿主自定义缓存必须公布 TTL、失效和撤销传播上限，并以租户、凭据标识和版本区分，不能省略授权。Go 运行时内存中出现秘密不可避免，不承诺彻底内存擦除。
- 宿主按 `(tenant_id, session_id)` 读取历史并鉴权；库无法判断任意文本是否来自其他租户。TenantID 自报、同进程不可信插件和厂商自身的账户隔离不属于库能单独保证的安全边界。
- 云端任务 payload 只携带经授权的引用及业务输入，不写入凭据秘密。未来平台账户供多租户共享时，也必须逐租户建立显式授权的账户绑定，不能恢复全局 key 兜底；后续短期服务身份更新仍固定租户和账户，不能使用无归属的凭据链。

### 4. Provider、协议与模型目录

| 首期 Provider × API | 凭据与 adapter | 需要单独验收的能力 |
| --- | --- | --- |
| OpenAI × Responses | 租户 API key；Responses adapter | 文本、工具、reasoning、受支持模型图片、完整/不完整/失败终态 |
| Anthropic × Messages | 租户 API key；Messages adapter | thinking 签名、redacted、工具、图片、缓存用量 |
| Google × Gemini Developer API | 租户 API key；Gemini adapter | thought 与 thoughtSignature、function call/result、图片和结束原因 |
| OpenAI × Chat Completions | 租户 API key；Chat adapter | 工具分片、流末 usage、finish_reason、模型支持的推理/图片 |
| DeepSeek × Responses | 租户 API key；复用 Responses adapter | 独立能力配置、reasoning_text、工具与历史回放；属于冻结 pi 路由之外的接入扩展 |
| DeepSeek × Chat Completions | 租户 API key；复用 Chat adapter | reasoning_content、工具结果回放、终态前 usage |

ProviderID 表示真实服务方，API 表示协议，模型按 `(ProviderID, API, ModelID)` 定位。按 Binding.API 从显式注册、构造后只读的注册表选择 adapter；共享协议不改变结果中的 Provider 身份，不共享账户配置，也不在失败后自动换协议、模型或 Provider。

Responses、Chat Completions 是协议简称，对照冻结 pi 的 API 标识分别为 `openai-responses`、`openai-completions`；协议名中的 OpenAI 不限定实际 Provider。DeepSeek 两个协议使用不同 BindingID，但可经授权引用同租户、同账户的同一凭据；模型配置必须保持协议维度独立。

模型目录只包含无秘密的版本化元数据：能力、模态、上下文/输出上限、兼容选项、reasoning 映射与价格。实际可调用范围是目录与 binding 授权的交集。受授权的配置可以显式补充自定义模型；请求不能临时绕过模型许可。首期不联网自动发现或更新目录。

新 Provider 完全符合已有协议时增加配置和验收场景；存在实际差异时在对应 adapter 内集中处理，不复制整套生命周期。支持声明按 Provider×API×模型能力发布；“兼容 OpenAI”不等于所有兼容服务已验收。DeepSeek 的专有字段和不支持能力根据冻结研究对应的官方 fixture 单独验证，不能仅替换地址就声称具备全部 Responses 能力。

冻结研究记录的 DeepSeek Responses 配置使用 `https://api.deepseek.com`，不支持 `previous_response_id`、`conversation`、`store`，推理流使用 `response.reasoning_text.delta/done`。请求构建不得自动沿用 OpenAI 的服务端续接或存储字段；历史通过完整输入回放，Provider 能力配置不将这些字段列为支持项。厂商能力是研究时点快照，升级时通过 P05 更新配置和差异记录。能力按 Provider 而非模型确定：DeepSeek 不发送 `store`、`include`、提示缓存字段与亲和头，显式请求服务等级在发送前拒绝；该组合在差分账本的 `routes` 中登记为扩展接入路径（维护者决定 2026-10-02，见 ADR-0014）。DeepSeek Chat 复用 Chat adapter，请求形状同样按 Provider 确定，照 pi 的 DeepSeek compat：不发送 `store`、不用 `developer` 角色、输出预算为 `max_tokens`、以 `thinking` 开关推理并只在开启时发送 `reasoning_effort`、推理模型的每条回放 assistant 消息带 `reasoning_content`；thinking 与强制工具选择的组合照发、由厂商拒绝，long 保留期照 pi 发送派生缓存键，二者以真实冒烟为准修订；该组合是冻结 pi 路由，全部场景进入差分（维护者决定 2026-10-02，见 ADR-0015）。

### 5. SDK 与传输边界

- 按既有选型研究，首期默认使用 OpenAI Go SDK 承担 Responses/Chat、Anthropic Go SDK 承担 Messages、Google GenAI Go SDK 承担 Gemini 的 HTTP 与流解码；统一语义仍由 barness-ai 负责例外：Anthropic Messages 的 SSE 由 adapter 按冻结 pi 自己的解码器解码，SDK 只承担 HTTP 与错误体读取，因 SDK 的两种解码方式均与 pi 的可观察行为不同（维护者决定 2026-10-02，见 ADR-0011）；Google × Gemini Developer API 整体直接 HTTP，流按冻结 pi 所用的 @google/genai 解码方式解码，因 Google GenAI Go SDK 无法无损保留 thoughtSignature 与 functionCall.args、不提供原始响应且会读取环境 key 与 endpoint（维护者决定 2026-10-02，见 ADR-0012）；OpenAI × Chat Completions 由 SDK 发送请求并切分 SSE 事件，事件内容由 adapter 按冻结 pi 所用的 openai-node 读取方式解读，因 SDK 的类型化流遇到 `[DONE]` 即停止读取、把 null 的 error 字段与非对象 chunk 当作失败，且不承载 pi 读取的推理字段与工具调用索引的缺失（维护者决定 2026-10-02，见 ADR-0013）。
- 研究记录的候选锁定版本是 OpenAI `v3.66.0`、Anthropic `v1.75.0`、Google GenAI `v1.71.0`，作为实施起点而非当前最新版声明。纳入依赖或升级时重新验证目标行为，不直接继承旧探针的通过状态。
- 保留原研究 H1–H5 的含义：H1 为调用级凭据/endpoint/header 隔离；H2 为字段 presence、签名、工具和 pi 行为差分；H3 为协议终态与部分消息；H4 为 pi 等价重试；H5 为读取/分配前的资源限制。请求回调时序、取消和资源释放亦为强制契约，不由 SDK 默认行为代替。
- 先核查 SDK 的公开扩展、原始 JSON 与受控 transport 能否满足缺口；确实无法满足时，仅为受影响 Provider×API 采用直接 HTTP。直接实现也通过同一组场景；多个协议出现已证明的系统性阻碍后才重议统一直连。
- 关闭 SDK 默认重试；不修改共享 SDK client 的 key。连接池不持有租户认证、cookie 或请求正文，禁用 CookieJar，默认不跟随携带凭据的重定向。协议终态由模块检查，不能把 SDK 无错误的 EOF 视为成功。
- 后续增加 mTLS 或带身份的代理时，按不同身份隔离 transport profile，不能复用携带另一账户传输身份的连接池。

锁定版本研究已暴露以下约束，必须转成目标模块 fixture，不得以升级 SDK 的推测代替验证：

| 研究发现 | barness-ai 的实现与验收要求 |
| --- | --- |
| OpenAI Responses、Gemini 缺协议终态时，SDK 可能以无错误 EOF 结束 | 自行识别对应终态；中断保留 partial/usage，以 E02、P01/P03 验证 |
| Google typed 流结果可保留已知 thoughtSignature/function call/usage，但无法重新序列化保留测试注入的未知字段 | 必需字段若未被 SDK 类型承载，检查原始响应或公开扩展能否无损保留；失败则按组合转直接 HTTP，不能先丢字段再声称差分一致 |
| OpenAI 的公开扩展曾用于 DeepSeek reasoning_content 等字段，SDK 不会自动选择全部 pi 默认值 | 核查 WithJSONSet、ExtraFields、ExtraBody、受控 RoundTripper 的实际可用性；显式配置默认字段并验证实际请求，不能仅比较 typed 对象 |

这些结论只描述冻结研究的 SDK 能力与缺口；历史探针数量、通过范围和尚未通过的项目门槛记录在研究追溯表。

### 6. 消息、历史与工具语义

公共模型包含 Message、ContentBlock、Tool、ToolCall、ToolResult、Model、Event、Usage、Result 和 Error。块与事件具有明确类型和索引；构造及解码检查非法变体组合，保持内容顺序。工具参数保留原始 JSON 与增量；部分解析仅供展示，修复及 JSON Schema 转换/校验由调用方显式使用，不触发工具执行。Go 静态类型不能替代运行时参数校验；参数完整且校验成功后才可交由宿主决定执行。

历史转换遵循冻结 pi 的上下文归一化、system/tools 重放、通用历史变换及 adapter 后续转换；调用方原始历史不被修改：

| 输入条件 | 必须保留的行为 |
| --- | --- |
| content 缺失或 null | 归一为空内容数组 |
| 目标不支持图片 | user/toolResult 使用对应占位文本并按基线合并连续占位，不改成统一拒绝 |
| 同 Provider/API/model 的 thinking | 保留 redacted；带 thinkingSignature 的块即使可见文本为空也保留；其他空白 thinking 删除，非空 thinking 保留 |
| 跨模型 thinking | 非空可见推理转普通 text；redacted 与空白推理丢弃 |
| 跨模型 text/toolCall | text 只保留文本，去除源模型附加状态；按原 truthy 条件删除 toolCall 的非空 thoughtSignature，不擅自合并缺失/null/空值；按目标规范化工具 ID，并同步 toolResult 关联 |
| 缺少工具结果 | 按基线顺序补 `No result provided` 的 isError 结果；调用与结果间的 system 消息按原规则延后 |
| error/aborted assistant 历史 | 按基线跳过该轮次 |

普通历史的租户归属由宿主在读取时鉴权，库无法也不尝试判断任意文本的来源；库只校验原生状态封套。已授权同租户跨模型/跨 Provider 的历史自动降级，不要求调用方先删状态，也不因源 binding/账户不同而整体拒绝；仅对转换后仍发送的账户专有引用检查适用性。

原生状态连同来源归属（TenantID、AccountScopeID、Provider/API/Model）保留在原生状态封套中。可信性由宿主担保，库不负责密封：宿主按自己的存储完成鉴权和完整性校验后，调用一个不能从 JSON 反序列化得到的 Go 构造入口（如 `TrustNativeState(scope, envelope)`）取得可信值；库只核对来源字段与当前 scope 和 binding 是否匹配。云端请求中的签名、密文、TenantID 或 `trusted=true` 不能自行获得这种权限。库在本进程内产出的消息自带可信来源（已解析的租户、账户与模型），宿主直接传入下一次调用时按同模型回放；经持久化读回的消息必须重新经该构造入口担保（维护者决定 2026-10-01，见 ADR-0001）。原生状态包括推理签名/密文、redacted 块、文本消息条目 ID 和工具调用的厂商条目 ID（如 Responses 的 `fc_` ID）；未经担保时工具调用去掉厂商条目 ID 但保留调用关联（维护者决定 2026-10-01）。同账户换 key 不应仅因凭据版本不同就丢弃有效状态。

没有可信封套、或同 Provider/API/model 但 AccountScopeID 不一致的原生状态，按上表跨模型规则降级，调用照常进行，不返回 invalid_request。降级计数及原因（无封套、账户不匹配、跨模型）写入 Result 调用元数据和 Observer，不进入 pi 兼容消息，也不参与 pi 差分。

### 7. 完整选项、统一选项与可信回调

- full 入口接收按 API 区分的完整选项并校验绑定协议；simple 入口保留 reasoning 的 minimal/low/medium/high/xhigh/max、thinkingBudgets、toolChoice 和公共参数。`off` 属于能力/映射值，不擅自增加到 simple 输入枚举。
- 移植支持等级、clamp、thinkingLevelMap 的禁用/重映射与未设置分支；按模型分别生成 OpenAI effort、Anthropic adaptive effort/token budget、Google level/budget，不以统一新公式覆盖协议差异。
- 默认 thinking budget 保留 minimal=1024、low=2048、medium=8192、high=16384 及自定义规则；保留 maxTokens、回答空间、输入 token 估算、4096 token 安全余量与无有效 contextWindow 的例外。这里的估算不授权模块裁剪历史。
- 需要区分未设置、null、零值的字段采用 presence-aware 表达，单独使用 omitempty 不足以表达三者。模型 samplingParams 与调用参数逐键合并、调用值优先；OpenAI-compatible 路径保留最后覆盖已命名请求字段的原语义，其他首期 API 按冻结实现忽略 samplingParams。发送前复核模型及资源授权，不通过采样参数越过认证/网络目标边界。
- 完整保留已纳入协议的 cacheRetention、metadata、toolChoice 及其他完整选项，不能只因 simple 路径没有同名字段而删掉。cacheRetention 的协议映射、metadata 的发送内容/字段 presence 和不适用协议行为分别按 fixture 验证；不把 TenantID 自动填进厂商 metadata。
- 可信宿主可提供每调用的 onPayload、onResponse 和 header 变换，与可反序列化的 Request 隔离。transport 只在 Client 构造时装配，不提供每调用覆盖；测试用 loopback 由测试装配注入。固定本次函数和配置引用，显式传入 scope；不依赖全局当前租户，不向云端 JSON 开放执行能力。
- onPayload 支持观察、原位修改或返回替换对象；“不替换”不撤销原位修改。回调等待完成并遵守 context，失败进入对应阶段错误终态。header 变换及 payload 修改不能绕过绑定的鉴权、目标和最终资源授权。payload 修改默认只能声明函数工具；托管工具（如 web_search）须由绑定的 AllowedHostedTools 按类型显式放行，放行不放宽模型、缓存键、存储和厂商侧状态引用的拒绝（维护者决定 2026-10-02，见 ADR-0005）。
- transformHeaders 在认证头与请求头合并后、分派给协议 adapter 前执行；onPayload 在 adapter 构建原生请求体后执行，二者不是同一个时点。可信变换后的最终认证和目标仍须满足绑定授权。每次调用的 onPayload/onResponse 各为一个可选函数，不引入多处理器注册；Go 必须区分“不替换”和“返回替换对象”，不能用含义不清的 nil 混合两者。

| Adapter | onPayload / onResponse 契约 |
| --- | --- |
| Responses / Chat | onPayload 在初始请求重试包装之外；onResponse 只在成功取得初始响应后、start 前执行，不按 token 或失败尝试触发 |
| Anthropic | 同上；替换 payload 后仍按基线强制 stream=true |
| Gemini | 保留 onPayload，回调看到 REST 请求体而非 pi 的 SDK 参数；冻结实现不调用 onResponse，且拒绝非默认 fetch 的行为不被悄悄改写；新增支持必须登记扩展（维护者决定 2026-10-02，见 ADR-0012） |

onResponse 读取 HTTP status/headers 和模型信息，不接收或替换最终 assistant 消息。Agent 的 before_run、transform_context、before_tool、after_tool、after_response、compaction 等生命周期机制不进入模块。Observer 与请求回调分开：前者失败不改变生成，后者位于执行路径且可以使调用失败。

宿主可将 before_payload 桥接到 onPayload；onResponse 先收集 HTTP 元数据，宿主在最终消息产生后再执行自己的 after_response。后者可能替换宿主层的消息，不改变 barness-ai 的 onResponse 时点及权限。

### 8. 流、终态、取消与重试

- Stream 系列先返回流对象，随后由后台生产过程执行可等待的绑定/凭据解析、建连、流读取和归并；慢解析不能让调用者无法取得流并 Close。预检失败也经流终态交付，不以同步 nil/error 替代。事件生产与最终结果通知独立。保留 start、text/thinking/toolcall 的 start/delta/end、done/error；块可交错且索引稳定。完整消费的流恰有一个终结事件，预检失败可在 start 前直接 error，成功增量不得先于 start。
- 累计部分响应保持 live partial 语义，以受同步保护的 PartialView 提供只读 Snapshot；多个事件可引用同一持续更新视图。delta、块索引和 end 数据保持各事件自身值，最终消息在终态后稳定。
- Result 不消费事件、不要求先 Next、不返回“流仍活跃所以不能取结果”的错误；可与事件消费并行。Complete 系列复用相同生产/归并过程，只等待最终结果，不为无人读取的事件建积压队列。
- 设置/租户/凭据失败与建立 HTTP 失败均产生错误 assistant 消息；流中断或协议错误保留部分内容、usage 和 errorMessage。正常 length 仍是明确截断终态，不等于完整工具参数；缺协议结束标志的 EOF 是错误。
- 成功终结的 done 携完整消息，Result 返回相同最终消息且 Err 为 nil；流中协议/Provider 失败产生 error 与 StopReason=error，Result/Complete 同时返回非 nil 的分类 error 且 Result 仍完整。调用方既不读事件也不等结果时必须 Close 或取消 context；宿主下游断开及发送失败均取消本次生成。
- Close 幂等；未完成时取消本次设置或 I/O，已完成时不改写结果。阶段行为保留基线：lazy setup 捕获的取消形成 error，进入 adapter 后按对应路径形成 aborted；附加错误分类描述取消/截止时间。所有路径都释放响应体、准入许可和等待者，不依赖再次 Next。
- 重试由运维人员在服务绑定上配置（`Binding.Retry`），不是请求选项（维护者决定 2026-10-02，见 ADR-0006）。默认 maxRetries=0。显式重试优先按 x-should-retry 判断，再按基线网络错误形状及 408/409/429/5xx 判定；不能把所有 Go error 视为可重试。401/403 不触发换 key 或其他身份兜底。
- 延迟依次采用 retry-after-ms、retry-after 秒数/日期、指数退避；基础从 500ms 起、封顶 8s、最多下浮 25%。maxRetryDelayMs 默认 60 秒，设 0 只关闭该项等待上限，仍受整个调用截止时间约束；退避可取消。
- 仅在各 adapter 的初始请求位置重试，流开始后不自动重放。所有尝试固定租户、绑定版本、凭据和账户；每次分别观测。请求到达厂商但响应丢失时，重试可能重复产生费用。

### 9. 资源、准入、错误与观测

本规范将原研究尚未决定的资源默认值收敛为**显式、有界的宿主策略**：Client 构造必须提供有效资源策略；未提供、把必需容量/时限设为零或负数以表示无限制、或字段关系非法时构造失败。首期不提供“关闭全部限额”模式，也不在协议 adapter 内散落隐式默认数值。这是本规范新增的决策 D1，并非冻结 pi 的默认行为。正式示例随实现提供已跑通压力场景的有限配置及数值依据；数值是部署策略，不构成 pi 兼容承诺。

策略至少覆盖请求与图片字节、单 SSE frame、工具 JSON、错误体、总输出、事件队列条数及字节、单租户及全进程并发、有限准入等待、调用总时限、建连/响应头与 Provider 读空闲时限。宿主 deadline、策略与协议 timeoutMs 按更早者结束；协议自身默认值仍按基线映射，宿主上限作为扩展单独验证。限额要在相应读取/解码/分配阶段生效，而非无限读入后才检查。总输出按一次调用读取的流式响应体字节计量；达到限额的尝试不重试（维护者决定 2026-10-02，见 ADR-0007）。

事件队列超限时结束本次请求并发布 resource_limit 错误，其 Phase 指明事件队列超限；保留已排队事件，为终态预留位置，不静默丢 delta，不阻塞 Result。不采用阻塞生产者的背压。Stream 后仅等待 Result 是有效用法，但仍受该流的队列上限约束，输出较长时会以 resource_limit 结束；仅需最终消息的调用方应使用 Complete，公开文档须写明这一点。达到并发上限时采用有界拒绝或等待，不建立无限准入队列。Provider 读空闲只计算上游数据等待；下游发送截止时间由宿主管理。

准入在每次尝试前获取许可并在结束/取消时释放；内置策略只覆盖单租户与全进程并发，只保证本进程边界。准入接口参数包含 TenantID 与 AccountScopeID；按厂商账户聚合的准入不内置，由宿主注入实现。宿主可注入分布式准入，但跨实例配额、硬费用预算预约与最终对账由宿主实现。租户与 AccountScopeID 分别用于资源归属和厂商账户聚合，不把 API key 当作账户身份。

错误至少区分 invalid_request、tenant_denied、binding_not_found、credential_unavailable、admission_denied、upstream_auth、rate_limited、upstream_error、transport、protocol、canceled、deadline_exceeded、resource_limit、callback_failed（可信宿主的请求回调出错或未产出可用请求体，与 invalid_request 区分，维护者决定 2026-10-02），并保留发生阶段、可安全公开的 HTTP status、厂商 request ID 和 Retry-After。错误分类不替换消息停止原因；基线 errorMessage 涉及秘密时必须脱敏并登记安全差异。

Observer 记录 CallStarted、AttemptStarted、AttemptFinished、CallFinished；调用级记录关联 TenantID/RequestID，实际尝试额外关联 AttemptID，预检失败不伪造 HTTP 尝试。记录已解析的 BindingID、AccountScopeID、配置/凭据版本、真实 Provider/API/model、时间、厂商 request ID、错误类别及已知 usage；ActorID/JobID 在提供时保留，无效身份的拒绝事件不能假装属于已授权租户。预检被拒的调用仍记录可信 scope 中的 TenantID，但标为未解析（`Resolved=false`），不出现账户、模型与版本；用量与成功率统计只取已解析记录（维护者决定 2026-10-02，见 ADR-0009）。

观测异步、容量有界，拥塞和失败不改变模型结果，丢失计数可查询；它不保证持久账本或 exactly-once。key、Authorization、正文、工具内容、原生密文不进入默认日志/观测/错误，授权结果仍完整保留原生状态。例外：Provider 错误体按 pi 原样进入 ErrorMessage 与 `Error.Message`（只脱敏 key 形态的文本），其中 Provider 回显的正文随之保留；该错误只返回给发起调用的租户，属于租户内容，宿主不得写入共享日志，需记录时使用 Observer。此为批准的安全差异（维护者决定 2026-10-02，见 ADR-0009）。受控调试采集由宿主显式启用并限制租户访问；原始 TenantID 不自动进入厂商 payload，也不默认成为无限增长的全局指标标签。

首期无应用结果缓存和厂商服务端会话依赖。保留协议 prompt cache 配置；模块构造的缓存/亲和标识按租户和账户作用域派生，不直接传裸 session ID 或 TenantID。厂商侧缓存隔离仍取决于账户/项目和厂商规则。

后续若增加 Provider response ID、文件 ID、原生缓存句柄等引用，必须验证租户与账户归属；应用缓存键至少包含租户、绑定/账户、模型、配置版本及请求摘要。不同 key 可能属于同一厂商账户，不能以 key 不同推断厂商缓存隔离。这些约束登记为后续能力的前置条件，不把缓存实现加入首期。

### 10. 用量、成本与交付完成条件

- Usage 保留 input/output/cacheRead/cacheWrite/totalTokens/cost、可选 cacheWrite1h/reasoning、初始化零值与 adapter 更新规则；reasoning 是 output 子集，不能重复加入总量。
- 成本移植冻结 pi 的阶梯价格、缓存读写、1h 写入及 adapter 专用调整，记录价格版本；金额为估算，不是账单承诺。
- Result/观测元数据另行区分未上报、部分上报、完整上报，不修改兼容消息的原数字字段。每次尝试分别记录；不能将全部尝试合计反写为 pi 最终消息 Usage，也不能从零值推断失败请求免费。缺失的计数按 0 计，Responses 缺 `cache_write_tokens` 不影响完整上报；`response.failed` 携带的用量与 pi 一样丢弃、尝试标为未上报（工单 25 已作废，见 ADR-0010 决策三）；价格版本以目录版本与哈希标在调用元数据上（维护者决定 2026-10-02，见 ADR-0010）。
- 交付完整公共契约、错误和装配说明、支持矩阵、模型/价格快照、fixture、差异登记及最小示例。协议组合只有在适用验收通过后才可宣称支持；尚未完成的其他组合不能因共享 adapter 被标记通过。
- 实施顺序为：冻结基线与失败场景 → 先建验收装置 → 调用内核及 Responses/Anthropic → Gemini/Chat/DeepSeek → 宿主接入示例、资源/隔离验收与发布冒烟。每批新增代码前先定义对应行为和故障用例。

## Testing Decisions

### 1. 按 barness 当前状态调整测试边界

当前仓库只有工程约定，尚无 Go 实现、既有测试或可复用的 Agent/云端宿主入口。采用**一个主验收边界：库外 Go 调用程序 → 公开 Client → 完整 barness-ai → Provider HTTP/SSE 边界**。用例以公共事件、结果、错误、观测及服务端实际收到的请求判定行为，不断言内部函数、包结构、锁或 SDK 累计器。

本地受控 Provider 是外部服务替身，不替换模块内部协议流程；使用真实 SDK/HTTP 栈。可信宿主替身只提供绑定、凭据、准入、观测及已认证身份，覆盖这些集成依赖的可见契约。库负责的预检必须通过服务器收到零推理请求证明，不能仅检查错误文本。

最小宿主接入示例通过一个调用程序验证“认证结果建立 scope、按租户读取历史、验证原生封套、取消传播”，无需另建 HTTP 云端服务、JWT 登录或数据库。恶意自报 TenantID/历史引用由该宿主边界拒绝；对库直接提交无效 scope/绑定则由库拒绝。这两类责任分别断言。未来完整登录→Agent loop→持久化会话 E2E 归对应 barness 宿主模块，不阻塞本模块交付，也不宣称本次已覆盖。

### 2. 三类证据复用同一主入口

| 证据类型 | 作用与运行时机 | 门禁 |
| --- | --- | --- |
| 确定性离线 E2E | 每次相关变更、合并、发布；Client 连接本地受控 Provider，无外网、无真实 key | 对应范围全部通过，含资源与隔离场景 |
| 冻结 pi 差分 | 复用适用 E2E 的逻辑输入和响应脚本，分别运行 pi 和 barness-ai，作为兼容判定器；协议/公共语义变更必跑 | 未解释差异阻断受影响协议；不是第二套独立业务测试体系 |
| 真实官方 API 冒烟 | 同一 Client 连接选定真实账户/模型；发布前、SDK/模型升级后执行 | 声明支持的组合必须通过；开发环境未配置报告 NOT_RUN，不冒充 PASS |

不执行会根据环境 key 自动访问线上服务的整套 pi 测试。普通 `go test ./...` 无外网；live 测试采用独立 build tag 或显式命令入口，加 `BARNESS_AI_LIVE=1` 双开关。真实测试只使用独立低权限测试账户，并限制 token 和调用次数；真实 key 仅由 CI secret store 向对应 Provider×API 的测试进程注入，不复用开发者默认账号或进程环境回退。

### 3. 测试装置与可重复性

- 用服务端脚本控制分帧、CRLF、多行 SSE、心跳、未知事件、响应头延迟、HTTP 错误、协议错误及中途断开，捕获 method/path/query、请求体、允许的认证信息和连接关闭顺序。逐字节拆帧和交错块必须能重放。
- 假 key 使用测试别名；服务端断言真实收到的测试 key，报告只显示别名。设置污染的 OPENAI_API_KEY、ANTHROPIC_API_KEY、GOOGLE_API_KEY 与 endpoint 环境变量，测试 transport 仅允许本地目标，证明无隐式回退或误出网。
- 带版本的绑定/秘密替身可控制更新、撤销、故障及读取次数；A/B 使用同名 binding/model/session。服务端屏障强制交错，不依赖概率碰撞或真实 sleep。
- 时钟、退避随机源和 ID 在必要的最小控制点可替换，不为测试引入通用框架。每个等待有截止时间；资源探针确认响应体关闭、许可释放、活动请求归零、goroutine 有界收敛，不断言脆弱的精确 goroutine 数。
- fixtures 保存输入、能力配置、响应脚本、预期请求/事件/结果、来源与哈希；固定 pi commit、运行时/SDK 版本、目录和价格快照。golden 更新附行为依据，不能用批量更新掩盖回归。

### 4. P0 验收场景

| ID | 必须覆盖的外部行为 | 主要验收证据 |
| --- | --- | --- |
| E01 调用生命周期 | full/simple、Stream/Complete、阻塞解析器、块交错、Result 不调用 Next、Result 与消费并行、live partial、安全快照和唯一终态 | 解析阻塞时已获得 Stream 且可 Close；有序事件、稳定最终消息、成功时 Result/Complete 的 error 与 Err() 均为 nil、Complete 无事件积压、输入不改写 |
| E02 失败与截断 | start 前设置失败、401/429/5xx、流内错误、半个 JSON、异常 EOF、正常 length、取消及 deadline | StopReason/Code/Phase 匹配，保留部分消息/usage；error/aborted 时 Complete/Result 返回非 nil error 且 Result 完整，Err() 与之相同；无终态 EOF 不成功，截断工具不被执行 |
| E03 历史与工具 | 同模型空文本签名、redacted、跨模型/Provider 降级、图片占位、工具 ID/缺结果、system/tools 重放、错误历史过滤、部分 JSON/修复/schema | 实际 wire 字段、内容顺序、调用结果关联与 pi 一致；仅经 TrustNativeState（或本进程库产出）且来源匹配的原生状态按同模型回放，无封套/账户不匹配时降级且调用成功，降级计数与原因出现在 Result 元数据与 Observer |
| E04 选项与回调 | 未设置/null/零值、全部 reasoning 等级、映射与预算边界、maxTokens、samplingParams、cacheRetention、metadata、工具选择、回调修改/替换/失败 | 非 OpenAI-compatible 路径忽略 samplingParams；header 合并后且 adapter 前变换；重试外 payload 次数；start 前 response 时点；Anthropic stream/Gemini 差异；回调跨租户不串用 |
| E05 重试 | 默认 0、显式次数、x-should-retry、可重试错误、Retry-After 秒/日期/ms、过大延迟、退避取消、开始流后断开 | 次数/延迟可重复，SDK 不额外重试，快照固定，不在流中重放，每次尝试独立记录 |
| E06 租户并发 | 同 Client/transport 的 A/B 同名配置交错，A 鉴权失败/取消/超时，B 正常 | 捕获 key/endpoint 和结果归属，无串流/串配置/串费用，不从其他租户或环境兜底 |
| E07 授权与更新 | 缺 TenantID/RequestID、越权 binding/model、缺/禁用 key、秘密后端故障、解析间更新、K1→K2、撤销、在途重试、注入认证/目标 | 拒绝时服务器零推理请求；D2 冲突失败无内部重新解析；宿主另起 RequestID 后见新一致快照；新调用与在途快照语义一致；无授权旁路 |
| E08 资源与释放 | D1 策略缺失/非法、慢消费者、只取 Result 的流队列满、各字节/并发限额、准入等待、建连/头/读空闲/总超时、Close 并发/重入 | 构造拒绝非法策略；边界值及 limit+1；只取 Result 的流超限以 resource_limit 结束且 Phase 指明事件队列；不静默丢已排队 delta；终态预留；Result 不死锁；I/O、许可与等待者释放；准入接口收到 TenantID 与 AccountScopeID |
| E09 观测与秘密 | 成功/预检拒绝/失败/重试/中断/Close、慢或失败 Observer、敏感错误体 | Call/Attempt 关系准确、无伪造尝试、拥塞有丢失计数；秘密/正文不泄漏，授权原生结果仍完整 |
| E10 路由与宿主契约 | OpenAI/DeepSeek 共享 Responses、DeepSeek 双协议、本地装配、可信身份/历史封套与伪造请求、下游取消、多流汇合 | 实际 Provider/API/model/endpoint/key 正确；逐事件信封及 Result 归属一致；未解析身份不伪造；宿主拒绝自报权限；JSON 反序列化无法得到可信原生状态；不存在每调用 transport 覆盖入口；只取消本次生成 |
| E11 用量与成本 | 成功/失败/重试 usage 的缺失/null/零值、缓存读写与 1h、reasoning、阶梯计价 | 公共 Usage/presence 对齐 pi；未知与部分计量独立表达；reasoning 不重复加总、价格版本固定 |

每个协议至少覆盖一次 full/simple 文本与最终结果、一条完整 Stream 消费、一条完全不调用 Next 的 Result（输出在事件队列限额内，超限路径由 E08 覆盖）、一条 Complete，以及成功/失败终态。能力组合按规则分配，不要求无意义的全笛卡尔积。请求层扩展和无权限行为也必须覆盖 full/simple 各条入口，避免出现低层旁路。

下列 P01–P06 是首期必过的协议专项，与 E01–E11 组合形成用例标识（例如 P03/E02/无结束原因），仍使用同一个 Client 主入口。每行都要本地证明工具往返、原生状态回放和适用图片行为，不能用其他协议的横向用例代替。

| ID / Provider × API | 本地 P0 正常路径 | 必须有的协议专项与断言 | 真实 API 最小冒烟 |
| --- | --- | --- | --- |
| P01 OpenAI × Responses | 文本、工具调用→宿主结果→下一轮、reasoning、支持模型的图片 | 分别处理 response.completed/incomplete/failed；无终态 EOF 为 error；工具参数分片、reasoning replay、usage presence；必需原始字段不丢失 | 文本 Stream/Complete、强制工具往返、首帧后取消；能力支持的推理回放/图片 |
| P02 Anthropic × Messages | 文本、工具往返、thinking 签名回放、图片 | message_stop；signed empty thinking、redacted、交错块、input_json_delta/signature_delta、SSE error；缓存读/写与 1h 计价；替换 payload 后 stream=true | 文本、工具往返、取消；能力支持的 thinking/图片 |
| P03 Google × Gemini Developer API | 文本、function call/result、thought signature 回放、图片 | 区分 thought 与 thoughtSignature；无 finishReason EOF 为 error；结束原因、usage、工具结果图片路由；未知必需字段原始通道；不以 Vertex 用例代替 Developer API | 文本、工具往返、取消；能力支持的 thinking/图片 |
| P04 OpenAI × Chat Completions | 文本、工具往返、支持模型的推理/图片 | 流末 usage chunk、finish_reason、工具 JSON 分片及 reasoning 扩展字段；结束信号前不得遗漏 usage；以协议终态判定成功 | 文本、强制工具往返、取消；能力支持的推理/图片 |
| P05 DeepSeek × Responses | 独立 binding 的文本、工具往返、推理及完整历史回放 | response.reasoning_text.delta/done、incomplete/failed；不自动注入 previous_response_id/conversation/store；Provider/模型/endpoint/key 与 OpenAI 独立；不支持字段行为单列 fixture | 文本、工具往返、取消、推理和完整历史；按实际能力增加图片 |
| P06 DeepSeek × Chat Completions | 独立 binding 的文本、工具往返、reasoning_content 回放 | reasoning_content 增量及带工具调用/结果的 assistant 回放、[DONE] 前 usage；不混用 Responses 配置；thinking 与工具选择的限制单独验收 | 文本、工具往返、取消、推理及工具历史；按实际能力增加图片 |

各 P 场景同时检查 full/simple 的适用映射。支持图片的模型跑实际图片请求；不支持的模型跑占位降级，两者不能互相替代。依据补充选型记录，冻结研究的 DeepSeek Chat thinking 模式不支持强制 tool_choice=required：强制工具冒烟在关闭 thinking 时运行，thinking+tools 的自动选择与推理历史回放另跑；不能用关闭 thinking 的通过结果证明组合能力。

### 5. 差分与真实组合的判定

差分比较实际请求的路径、headers、认证来源、JSON 字段 presence，以及 Provider/API/model、消息块/工具关联/原生状态、事件类型/块索引/delta/顺序、StopReason、errorMessage、Usage 与成本。仅语义排序 JSON 对象键；数组与事件不能排序。时间与生成 ID 采用明确字段的一一映射并保持后续引用，不能统一删除 null、零值、空数组、未知字段或所有 ID。

租户元数据、Code/Phase、计量完整性、资源限制、同步访问与秘密脱敏等扩展单独断言，不塞入 pi golden。DeepSeek Responses 按研究记录属于额外接入路径，使用官方协议 fixture 和真实冒烟证明，不伪装成冻结 pi 差分通过。差异只能归类为“已修复”“本规范定义或后续明确批准的扩展”“待处理”；待处理项阻断对应协议发布。

真实冒烟覆盖六个首期组合，各经目标库执行短文本 Stream/Complete、模型工具调用→测试宿主返回结果→下一次生成、首帧后取消，以及受支持模型的推理/签名历史和图片。断言结构、关联、终态、非空有效内容和 usage，不断言随机文本全文。能力不支持时记录 UNSUPPORTED，并为首期承诺的能力选择支持模型，不能用 skip 代替覆盖。

重试、异常 EOF、资源上限和恶意跨租户输入只在本地受控服务确定性注入，不向真实厂商发送大量故障请求。真实测试每次记录模型、SDK 版本、厂商 request ID、耗时和错误类别；厂商临时故障标记 FAIL/环境故障并在既定预算内重试确认，不把失败改成 skip，也不无界重跑。

结果统一为 PASS/FAIL/NOT_RUN/UNSUPPORTED；缺 key 为 NOT_RUN，厂商故障为 FAIL/环境故障，不能自动改成跳过。支持矩阵保存 Provider、协议、模型、SDK、测试账户/区域别名、能力与最后通过时间。其他协议或旧 SDK 探针通过不能代替目标组合通过。冒烟以 `live` build tag 加 `BARNESS_AI_LIVE=1` 运行，一个进程只跑一个组合、只持有该组合的 key；支持矩阵为 `ai/live/support-matrix.json`，只经合并本组合报告更新，报告覆盖全部应有场景且均为 PASS 或 UNSUPPORTED 时才记为完整通过（维护者决定 2026-10-02，见 ADR-0016）。

### 6. 证据产物与门禁

每次 E2E 必须生成可核验、可重放的产物：运行清单、case ID、版本与模型/价格/fixture 哈希、脱敏请求与响应脚本、事件和最终结果、观测/资源释放摘要、断言报告、差分首个不同位置，以及单场景重放命令。敏感原始正文仅限合成离线 fixture；真实请求输出先脱敏，不保存 key 或不必要的用户数据。

差分记录至少有 case_id、pi_commit、sdk_versions、model_catalog_hash、原始请求与帧哈希、首个不同 JSON path/事件位置、分类、处理决定。用例报告关联研究 T/C 条目、当前 E/P 场景及 D1/D2 扩展，做到“需求 → 断言 → 实际证据”可追溯。追溯表中的“已映射”只表示文档覆盖，不能填为测试 PASS。

Go 模块建立后，常规门禁执行 `go test ./...`、`go test -race ./...`、`go vet ./...` 和受影响的确定性差分。发布要求相应 P0、race、无待处理差异及该支持组合真实冒烟全部通过。SDK/模型/adapter 改动重跑受影响组合及横向隔离场景，报告分别列出离线、差分和 live 结果。

批次 B 的 Responses/Anthropic 先通过适用 E01–E09、E11 与 P01/P02，再接 E10 宿主契约；批次 C 新增 P03–P06 时复用所有适用横向用例；发布批次补齐六组合 live、压力和脱敏审计。不能以某批开始实施代替上一批验收通过。

优先使用上述 E2E 作为功能验收，不在实现后补单元测试。必须隔离验证公开工具 JSON 助手等独立系统时，先枚举所有失败方式并写行为测试，再编写实现；不为内部实现细节扩大测试入口。参考既有研究归纳的 pi 流、终态、历史转换、工具 JSON、重试、system/tools 重放测试及三家 SDK 的 HTTP/SSE 故障样本；迁移的是行为场景，不能直接继承其通过结论。

## Out of Scope

- Agent loop、工具执行、系统指令编写、上下文选择/裁剪/摘要/记忆检索、Session 存储、任务队列、沙箱、TUI 与前端体验；它们由 barness 的相应宿主模块承担。
- 独立部署的推理网关、稳定云端 HTTP 服务协议、用户登录/JWT、租户或 key 管理后台、特定数据库/KMS/Vault 产品、部署编排。
- 订阅账号登录、浏览器回调、device code 和订阅 token 刷新，包括 ChatGPT/Codex、Claude Pro/Max、Copilot；不提供相关后端或旧 npm/JS 宿主兼容入口。
- pi-messages 私有协议、Cloudflare AI JS binding 和 Node/Bun/Workers 宿主绑定不交付；这不影响已列出的官方 API 协议范围。
- 首期暂不包含 Azure OpenAI、Google Vertex、AWS Bedrock、Mistral、OpenRouter 等网关，以及服务身份 IAM/service account/workload identity。官方服务身份 OAuth 与订阅账号登录不同，后续可按独立规范纳入，不视为永久排除官方云 API。
- 首期暂不包含图片生成、音视频生成、批处理/后台生成、官方 WebSocket、动态模型发现/自动目录更新、紧凑流帧持久化、应用结果缓存和服务端会话续接依赖；图片输入在首期范围内。
- 分布式配额、持久审计、账单/outbox、财务 exactly-once、失败免费或厂商幂等保证、立即撤回已接受计算、同进程不可信代码隔离和厂商自身缓存隔离承诺。
- 通用插件加载器、多处理器优先级 HookRegistry、脚本执行引擎、跨 Provider 自动故障转移、旧内部路径兼容层与双写。

## Further Notes

### 本次综合采用的范围决策

| 原研究待定项 | 本规范的决定 |
| --- | --- |
| R1 首期 Provider | 四类协议、六个接入组合；先 Responses/Anthropic，再 Gemini/Chat/DeepSeek |
| R2 官方 API | 首期直接厂商 API key；官方云服务身份和商业网关后续独立验收 |
| R3 多模态 | 首期保留图片输入；生成及音视频能力后续 |
| R4 资源策略 | 构造要求显式有限策略；缺失/非法即失败，首期无无限制模式，示例数值以压力场景验证 |
| R5 交付 | barness 内可嵌入基础模块与最小调用示例，不为本期额外构建云端服务 |
| 测试边界调整 | 用户要求结合项目情况调整后，采用单一 Client 主边界；pi 差分复用场景、live 作为发布证据、只验宿主接入契约 |

以上是本次 spec 的综合决策，不将原研究中的“建议”伪写成先前已获用户批准的事实。协议能力和 SDK 结论采用研究冻结时点；本次没有重新访问厂商、安装依赖或执行模型请求。实现中若发现新证据与规范冲突，应先登记具体差异并更新本规范及相关 ADR，不能静默删减兼容行为。

### 新增决策及其代价

| 决策 | 相对原研究的变化与理由 | 影响、风险和重新评估条件 | 验收 |
| --- | --- | --- | --- |
| D1 显式有限资源策略 | 原研究将默认值和可关闭策略留待评审；本规范要求构造时提供策略，避免本地与云端使用不同的隐式无界路径，并让每项容量和时限可验证 | 本地装配也增加配置负担；过小值会拒绝合法请求，过大值削弱保护。示例必须说明数值依据；未来有可重复压力数据支持时，可另立决策提供版本化有限默认配置，仍保留显式覆盖与构造校验 | E08；限额内协议保持兼容，超限单独登记扩展 |
| D2 不一致配置快照直接失败 | 原研究允许发请求前重新解析或失败；首期选择失败，避免隐式配置重试与 Provider 重试混杂、版本更新时形成重试风暴，并简化一次调用的身份快照 | 凭据轮换/更新竞争可能造成短暂前置失败；只针对无法取得一致快照的情况，不把所有更新都拒绝。错误应指出配置一致性阶段且不泄漏秘密，宿主可用新 RequestID 在预算内再调用；若后续出现可量化的可用性问题，再评估有界重解析及其截止时间/观测契约 | E07；零 Provider 请求、无内部重解析、后续新调用可恢复 |

D1/D2 是 barness-ai 的新增取舍，不是 pi 兼容行为，也不是原研究已经证明的最优解。两者分别记录在 [ADR-0002](../../docs/adr/0002-barness-ai-resource-policy.md) 和 [ADR-0003](../../docs/adr/0003-barness-ai-snapshot-consistency.md)，更改时同步更新契约、扩展登记和 E07/E08。

术语以 [GLOSSARY](../../GLOSSARY.md) 为准，核心边界及取舍记录在 [ADR-0001](../../docs/adr/0001-barness-ai-protocol-boundary.md)。待实现的验收事项是交付条件，不是已知实现缺陷；本次不创建推测性的技术债务或实施票据。

### 研究来源与追溯

下列原文只读，文中的旧工作名仅属于历史来源；barness 规范正文、后续实现和测试开关使用 barness-ai。绝对链接用于保留本次已有研究的准确位置，移植 fixture 时应将必要基线快照及出处纳入 barness 自身可重放的验收产物，不能让 CI 依赖开发者机器上的另一个仓库。

| 来源 | SHA-256 |
| --- | --- |
| [既有需求研究](/Users/cyber/RestoX/harness/.scratch/pi-ai-go/spec.md) | `ed9b352f1ed115f69afeeb8efad96c73bea7421f0acf0109ef9ccefa2c5a3c67` |
| [既有技术设计](/Users/cyber/RestoX/harness/.scratch/pi-ai-go/design.md) | `e72cf866231e98c40c5e0b27ab382bf5defb7e2974b82dc57e95728c70c66aad` |
| [既有 E2E 计划](/Users/cyber/RestoX/harness/.scratch/pi-ai-go/e2e-test-plan.md) | `bd0fa17b595f127b3a86cd956001f05ea8b029f61c22d639fb862b0f41ff8489` |

原设计记录当时收集的 pi checkout 缺少 `src/providers/data` 生成 JSON，且尚未完成完整参考构建；后续选型探针记录的四协议简单请求差分通过，不能推导完整兼容基线已建成。实施时在固定提交的独立副本补齐生成模型数据与依赖，记录数据来源/哈希和可重建步骤，不修改原收集源码，也不让 CI 依赖本机研究目录。

历史选型记录报告三 SDK 锁定版本下 28 个离线顶层测试及 race 通过、四协议简单文本请求 JSON 差分通过，以及六个组合的真实文本、强制工具调用和首帧取消通过；另有 DeepSeek 双协议推理与历史回放。它们只证明探针覆盖的 SDK 能力，不证明 barness-ai 的解析器、公共事件/结果、重试、异常 EOF 或资源策略通过。[选型验证记录](/Users/cyber/RestoX/harness/docs/research/pi-ai-go-sdk-validation.md)

原实现、具体用例和 SDK 测试来源在 [研究追溯表](research-traceability.md) 提供直接链接与迁移边界。Go 接口草图、示例目录树和 TS 语法对照中的实现形式可按本项目调整；已提取的外部行为和 P0 断言不能因形式不同而删除。以后更新基线需明确版本、行为变化、差异处置与回退版本。
