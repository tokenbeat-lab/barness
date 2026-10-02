# barness

barness 面向本地与云端的 Agent 使用场景；barness-ai 是其中连接多种 LLM Provider 的基础协议模块。

## Language

**barness-ai（模型协议模块）**：
barness 中统一模型调用、消息、流事件与结果语义的基础模块，负责一次生成轮次与目标模型服务之间的协议转换。
_Avoid_: 独立 Agent 平台、模型托管服务、推理网关

**云端（场景）**：
多租户、多用户共享运行环境的使用场景，与开发者自行装配租户和凭据的本地场景相对。
_Avoid_: 远端、服务端

**Host（宿主）**：
使用 barness-ai 的本地程序、云端服务或 Agent 运行环境，是调用者身份、业务授权、上下文与工具执行的责任主体。
_Avoid_: Provider、租户

**Provider（模型服务提供方）**：
实际接收生成请求的模型服务提供方，其身份独立于采用的 API 协议；同一 Provider 可以提供多种协议。
_Avoid_: API、协议、凭据

**API Protocol（API 协议）**：
模型请求、响应与流事件所遵循的交互格式和语义，同一种协议可以由多个 Provider 提供。
_Avoid_: Provider、SDK

**Model（模型）**：
特定 Provider 和 API 协议下可被调用的模型，其身份由 Provider、协议和模型标识共同确定。
_Avoid_: 仅凭模型名称确定的全局模型

**Tenant（租户）**：
模型调用的授权与数据归属边界，拥有自己获准使用的服务绑定和历史；租户可以包含多个用户或服务主体。
_Avoid_: 用户、Provider 账户、API key

**Actor（调用主体）**：
经宿主确认、代表某个租户发起调用的用户或服务主体。
_Avoid_: 租户、worker 实例

**Binding（服务绑定）**：
某租户获准使用的一份模型服务配置，确定 Provider、协议、服务目标、账户、可用模型和凭据引用。
_Avoid_: Provider ID、API key、凭据版本

**Account Scope（厂商账户作用域）**：
Provider 侧账户或项目的归属范围，用于解释原生资源、厂商配额和缓存的适用性；它与租户是不同维度。
_Avoid_: 租户、API key

**Credential Snapshot（凭据快照）**：
某次调用获准使用的特定版本凭据及其租户、账户归属，贯穿该调用的请求尝试。
_Avoid_: 全局 key、当前 Provider key

**Call Scope（调用作用域）**：
宿主为一次逻辑调用建立的可信租户身份和关联信息，可包含调用主体与所属作业。
_Avoid_: 请求中自报的租户身份

**Logical Call（逻辑调用）**：
宿主向 barness-ai 发起的一次模型生成调用，包含该调用内部获准进行的网络尝试。它由全局唯一的 RequestID 标识，其尝试由同样全局唯一的 AttemptID（`RequestID#序号`）标识；服务绑定、凭据等配置标识只在租户内唯一，不同租户可以同名。
_Avoid_: Agent run、作业、HTTP 尝试

**Generation Turn（生成轮次）**：
模型根据本次给定上下文生成一个 assistant 消息的过程，消息可以包含工具调用；工具执行后的继续生成属于下一轮次。
_Avoid_: 整个会话、Agent loop

**Attempt（请求尝试）**：
一次逻辑调用中向 Provider 实际发起的一次请求；重试形成新的尝试，但不成为新的逻辑调用。
_Avoid_: 逻辑调用、生成轮次

**Retry Policy（重试策略）**：
运维人员为服务绑定显式配置的初始请求重试次数与延迟上限（Go 中为 `Binding.Retry`），默认不重试；同一逻辑调用的所有尝试固定使用解析时的策略，流开始后从不重放。
_Avoid_: 请求选项中的重试次数、SDK 默认重试

**Job（作业）**：
宿主管理的一项业务工作，可以触发多次逻辑调用；它的生命周期由宿主定义。
_Avoid_: RequestID、生成轮次

**Transcript（消息历史）**：
宿主为本次调用选定的消息、系统指令与工具变化序列，是协议转换的输入，不代表模块拥有整个会话。
_Avoid_: 模块内置记忆、自动上下文管理

**Native State（原生续接状态）**：
Provider 返回、用于后续生成的签名、密文或其他协议状态（含文本与工具调用的厂商条目 ID），具有来源和适用范围，不能仅凭客户端声明获得可信归属。
_Avoid_: 普通推理文本、跨账户通用状态

**Native State Envelope（原生状态封套）**：
随原生续接状态一起保存的来源信息，记录其租户、厂商账户作用域和模型；只有经宿主担保后才构成可信回放依据。
_Avoid_: 签名、凭证、可信标记

**Call Attribution（调用归属）**：
一次逻辑调用不可变的归属：可信作用域中的租户、RequestID 与可选的调用主体和作业，所请求的服务绑定，以及解析成功后实际服务调用的 Provider、协议、模型与厂商账户（Go 中为 `CallAttribution`，内嵌于 `CallMetadata`）。解析前失败的调用标为未解析，不回显请求中的模型。
_Avoid_: 请求自报身份、当前租户

**Event Envelope（事件信封）**：
流事件连同产生它的调用归属（Go 中为 `EventEnvelope`，经 `Stream.Envelope` 取得）；宿主汇合多个流时按信封归属分派事件，不依赖到达顺序或全局“当前租户”。它是 Go/租户扩展，事件本身仍是 pi 的事件。
_Avoid_: 原生状态封套、事件序号

**Tool Round Trip（工具往返）**：
模型返回工具调用，宿主执行工具并提供结果，再由宿主发起下一次模型调用的完整交互。
_Avoid_: 模块自动执行工具

**Usage（用量）**：
模型服务报告并经协议转换的 token 使用信息；缺少报告不代表没有发生消耗。
_Avoid_: 厂商账单、应付金额

**Estimated Cost（估算成本）**：
依据模型价格版本和已知用量计算的估算金额，不等同于厂商最终账单。
_Avoid_: 实际扣费、确定费用

**Usage Reporting（用量完整性）**：
一次请求尝试的用量由 Provider 报告到什么程度：未上报、部分上报或完整上报（Go 中为 `Attempt.UsageReporting`）。它与兼容消息的 Usage 分开记录，使零值用量与未上报可以区分。
_Avoid_: 用量为零、免费

**Price Snapshot（价格快照）**：
估算成本所依据的模型目录及其价格，以目录版本和内容哈希标识（Go 中为 `CallMetadata.CatalogVersion`/`CatalogHash`）；内容变化必须换新版本。
_Avoid_: 实时价格、账单价格

**Admission（准入）**：
一次请求尝试发往 Provider 前取得的有限并发许可，作用于本进程或由宿主注入的范围。
_Avoid_: 配额、费用预算

**Request Callback（请求回调）**：
可信宿主为单次逻辑调用提供的 header 变换、请求体回调与响应元数据回调（Go 中为 `Hooks`），位于执行路径上、可使调用失败，但不能改变绑定授权的认证、目标、模型与原生引用。
_Avoid_: Observer、生命周期钩子、每调用 transport

**Observer（观测器）**：
宿主注入的调用与尝试记录接收方（Go 中为 `Config.Observer`），接收 CallStarted、AttemptStarted、AttemptFinished、CallFinished；异步、有界投递，不在执行路径上，失败或拥塞只计数、不改变调用结果，记录中不含秘密、正文或错误文本。
_Avoid_: 请求回调、日志、指标

**Hosted Tool Allowance（托管工具放行）**：
绑定上列出的托管工具类型（Go 中为 `Binding.AllowedHostedTools`），表示可信宿主担保该账户的托管资源及工具访问的网络目标归该租户使用；只有列出的类型可以由请求回调声明。
_Avoid_: 工具白名单、全局工具开关
