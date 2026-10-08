# barness

barness 面向本地与云端的 Agent 使用场景；barness-ai 是其中连接多种 LLM Provider 的基础协议模块。

## Language

**barness-ai（模型协议模块）**：
barness 中统一模型调用、消息、流事件与结果语义的基础模块，负责一次模型操作与目标模型服务之间的协议转换。
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

**Model Operation（模型操作）**：
宿主请求模型完成的一类工作：聊天、图像生成或分类。操作是型号身份与服务绑定授权的一部分，独立于 Provider 和 API 协议。
_Avoid_: 协议、工具调用、生成轮次

**Model（模型）**：
特定操作、Provider 和 API 协议下可被调用的型号，其身份由操作、Provider、协议和型号标识共同确定；同名型号可跨操作并存。
_Avoid_: 仅凭模型名称确定的全局模型

**Chat Model（聊天型号）**：
接收消息历史并生成 assistant 消息的型号，可接受图片作为输入。
_Avoid_: 图像型号、分类器

**Image Model（图像型号）**：
输出包含图片的型号，以输入输出模态、图片能力与分模态价格描述。
_Avoid_: 能看图片的聊天型号、图片文件

**Classifier Model（分类型号）**：
对给定状态作答单选、评分或是非问题的型号，以问题类型、容量范围与价格描述。
_Avoid_: 聊天模型提示词、业务规则

**Tenant（租户）**：
模型调用的授权与数据归属边界，拥有自己获准使用的服务绑定和历史；租户可以包含多个用户或服务主体。
_Avoid_: 用户、Provider 账户、API key

**Actor（调用主体）**：
经宿主确认、代表某个租户发起调用的用户或服务主体。
_Avoid_: 租户、worker 实例

**Binding（服务绑定）**：
某租户获准使用的一份模型服务配置，固定一种操作、Provider 与协议，以及服务目标、账户、可用型号和凭据引用。
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
宿主向 barness-ai 发起的一次模型操作调用，包含该调用内部获准进行的网络尝试。它由全局唯一的 RequestID 标识，其尝试由同样全局唯一的 AttemptID（`RequestID#序号`）标识；服务绑定、凭据等配置标识只在租户内唯一，不同租户可以同名。
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

**Managed Effort（托管推理强度）**：
厂商在对话内逐轮管理推理强度的模型能力（Go 中为 `ModelCompat.SupportsMidConvoEffort`）。每条 assistant 消息记录本轮强度（`ProviderThinkingLevel`），回放时向厂商说明每轮是以什么强度生成的，因此宿主可以在对话中途改变强度。本轮强度不是原生续接状态，回放时不需要封套（ADR-0019）。
_Avoid_: 请求级 effort、推理预算

**Call Attribution（调用归属）**：
一次逻辑调用不可变的归属：可信作用域中的租户、调用标识与可选的调用主体和作业，入口确定的操作、所请求的服务绑定，以及解析成功后实际服务调用的 Provider、协议、型号与厂商账户。解析前失败的调用保留入口操作并标为未解析，不回显请求中的型号。
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

**Provider Capabilities（Provider 协议能力）**：
某 Provider 在一种协议上实际提供的服务端能力与请求形状（如 Responses 的存储声明、加密推理、提示缓存、服务等级；Chat Completions 的 store、developer 角色、输出预算字段、thinking 开关与 assistant 消息上的 reasoning_content），按 Provider 而非模型确定；请求构建只从对方提供的能力派生字段，不因共享协议 adapter 而沿用另一 Provider 的字段。
_Avoid_: 模型 compat、"兼容 OpenAI"

**Extension Route（扩展接入路径）**：
冻结 pi 没有路由的 Provider × 协议组合（当前包含 DeepSeek × Responses、原生 OpenAI × Images 与 Google × Interactions 图像），以自身协议 fixture 和真实冒烟证明，登记在差分账本中，不计为 pi 差分通过。
_Avoid_: 差分通过、pi 已覆盖

**Live Smoke（真实冒烟）**：
同一 Client 以低权限测试账户连接厂商真实官方 API 的发布证据（Go 中为带 `live` build tag 的 `ai/live`，另需 `BARNESS_AI_LIVE=1`）；一个进程只跑一个操作 × Provider × 协议组合、只持有该组合的 key，结果只有 PASS、FAIL、NOT_RUN、UNSUPPORTED 四种，厂商故障在预算内重试后为 FAIL，不改成跳过。
_Avoid_: 集成测试、跳过、SDK 探针

**Support Matrix（支持矩阵）**：
每个首期操作 × Provider × 协议组合的真实冒烟状态（Go 中为 `ai/live/support-matrix.json`）：模型、SDK、测试账户/区域别名、各能力的结果与最后通过时间；只由该组合自己的冒烟报告合并更新，共享 adapter 的通过不连带标记其他组合。
_Avoid_: 兼容列表、"兼容 OpenAI"

**Design Load（设计负载）**：
资源策略示例在注释中声明所针对的负载（并发、请求与图片大小、单轮输出或工具参数大小）；压力场景（`E08-policy-pressure-*`）按它运行，记录每项限额的用量占比（headroom）与内存，作为该策略数值的依据。设计负载跑通不表示数值是通用默认值。
_Avoid_: 默认配置、基准测试

**Redaction Audit（脱敏审计）**：
对已写出的证据包（离线、live、race）逐文件检查：已注册秘密与审计进程环境中的凭据、主目录、主机名，厂商 key 形状，凭据请求头与 URL key 参数，Observer 记录中的非元数据字段，live 响应头白名单；发现即令运行失败，结论只给文件、规则与位置，不复述被发现的值。
_Avoid_: 日志过滤、事后清理

**Release Gate（发布门禁）**：
发布前对 barness-ai 的一次判定：vet、离线与 race、全部 P0、固定基线的差分、九组合自身完整真实证据、逐包脱敏审计、完整需求追溯、目录/价格快照、设计负载与独立中文效果均有效才可发布。必交能力不可声明不支持以放行；协议、资源容量与业务效果分别呈现。只有映射或历史通过而没有当前适用证据的条目不算通过。
_Avoid_: CI 绿灯、发布流程

**Chinese Classification Evaluation（中文分类评估）**：
宿主针对固定中文标注任务和固定分类型号形成的独立业务效果证据，包含准确率、误判、置信统计、完整样本状态与消耗。其适用范围由标注任务界定，独立于协议兼容与真实冒烟的支持结论。
_Avoid_: 协议准确率、支持矩阵通过、生产准确率保证
