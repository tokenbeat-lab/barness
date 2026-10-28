---
status: accepted
date: 2026-10-20
---
> 当前规则：2026-10-27 的 [ADR-0055](0055-route-scoped-evidence-applicability.md) 已替代本文固定轮次、全局来源/目录失效和跨阶段输入恒等部分；原授权、预算、留存和暂停事实保留。以下原决定供历史追溯。


# 中国站与国际站使用独立 Provider 身份

GLM 与 MiniMax 的地区市场需要独立维护型号、能力、默认值、价格和账户授权。采用已有 Provider 身份承载独立服务契约，保留 `ProviderZhipu=zhipu` 表示智谱中国站、`ProviderMiniMax=minimax` 表示 MiniMax 国际站，新增 `ProviderZAI=zai` 与 `ProviderMiniMaxCN=minimax-cn`。型号身份继续为 `(Operation, Provider, API, ID)`，同名型号独立配置；不增加 Site 维度，也不将地区 ID 归一化为品牌 ID。

API adapter、SDK 与调用生命周期继续共享。每种协议以内部不可变配置明确选择已核验的服务能力和 codec；共享实现不得替代精确 Provider 的目录、授权、来源、价格或支持证明。Host 固定端点、账户、凭据来源与使用资格，不自动换站或换 key；标准 API 与套餐用途分别声明。

缓存与亲和标识的现有派生作用域加入 Provider，统一切换 `barness-ai/prompt-cache-key/v2`，输入为长度前缀编码的 Tenant、Provider、AccountScope 与 Session。这样即使账户别名相同也不会产生相同的跨服务路由键。代价是一次缓存命中率变化；不保留 v1 shim 或双发。NativeStateEnvelope 保持现有字段，继续按真实 Provider 精确比较，跨地区只走现有安全降级。

AdmissionRequest 同时增加实际 Provider，供 Host 在账户级准入中区分不同市场的同名账户；不改变库内租户 / 进程限额，也不新增配额系统。回滚先禁用新增服务并移除旧进程目录中的新身份，避免旧代码把新 Provider 作为通用兼容服务继续调用。

此决定沿用 ADR-0001、ADR-0020、ADR-0032，扩展 [ADR-0042](0042-barness-ai-minimax-international-target.md) 当前只接入 MiniMax 国际站的范围；不撤销其独立地区证明、历史消费、禁止回退和原件留存要求。新增中国站是新的正式接入目标，不能恢复旧失败捕获的支持资格。架构决定已选定，实施与发布资格按 [spec 当前状态](../../.scratch/barness-provider-sites/spec.md) 记录。

## Considered Options

仅在 Binding 增加站点能切换连接，但无法为同一完整型号身份配置不同元数据。让 Site 进入目录、授权、原生历史与观测也可满足需求，但增加贯穿全链路的身份维度；现有 Provider 已表达独立地区服务，因此选择独立 Provider。不同地区复制 adapter 会重复资源、重试与流生命周期，也不采用。

[研究](../../.scratch/barness-provider-sites/research.md)、[设计](../../.scratch/barness-provider-sites/design.md)、[规格](../../.scratch/barness-provider-sites/spec.md)。

## ZAI Messages 的独立窄思考形状

2026-10-20 的自身 standard API 四请求对照中，GLM-5.2 的省略形状没有原生 thinking/signature，精确 `enabled/budget1024/max4096` 工具回合返回完整原生块；GLM-5.3 两形状均返回原生块。这为 ZAI 自己的内部 Messages profile 建立 opt-in，并只为自身 GLM-5.2 候选设置 `MessagesThinkingBudgetTokens=1024`，不借用中国站证据。Provider 限定精确形状，Model 只能收紧；其他预算、输出配对及额外 thinking 字段仍被拒绝。

无选项仍省略，GLM-5.3 保留已证默认。单回合研究不证明预算被服务端强制执行，也不证明 canonical intensity 或实际 native-history 续接。Simple 不推导该形状；原始 SSE 在公共 Client 离线重放保持原字段，同源第二轮为明确标注的受控响应。候选资格和晋升后 formal 仍分别需要新来源、新授权的一次完整真实过程，原失败不改写。[研究与原件](../../.scratch/barness-provider-sites/evidence/zai-thinking-research.md)。

## ZAI Chat / Responses 的预定原生推理来源

ZAI 自身 Chat 与 Responses 的合法工具响应都可能只有工具调用、没有 reasoning；要求两种证据来自同一个工具回合会把原生续接与工具执行混为一项。两条路线在任何请求前按完整型号身份固定 `full-stream` 为 `native-history` 的来源，沿用该文本回合的真实用户／助手前缀，再保留独立的工具用户／助手、Host 工具结果与最终用户提示。其他四条地区路线保持各自已定义来源。

这项选择不根据返回内容变化。固定源缺少原生推理时仍须失败，即使另一工具回合含有推理也不能替代；工具响应中缺失的 Thinking 不由 Host 合成。报告声明每型号固定 sourceCase，独立消费者按同一合同核对来源、完整身份与租户／账户、确切前缀及真实 native wire，工具往返与零降级断言继续分别成立。代价是续接请求多保存一个自身文本回合，并按最终编码重新计量输入预留。实际失败原件与实施依据见[工单16](../../.scratch/barness-provider-sites/issues/16-zai-responses-native-source.md)，自身正式支持仍需当前来源的完整真实证据。

## 原生来源合同变化后的资格边界

上述 ZAI Responses 固定来源合同将原生推理来源扩展到自身 `full-stream`；harness 与 `RegionalNativeSourceCase` 消费者同步。消费者属于国际 MiniMax 执行来源，因此这项验收设计变化同时使三条 M3 与六条地区 formal 的旧来源资格过期；不能通过刷新旧报告或缩小选择器继续使用。GLM 的保留取决于其完整245项执行来源逐项相等。此前真实 PASS／FAIL、父 UNKNOWN、grant／nonce／账本与美元占用均保留各自事实。

离线受控流程只验证新来源合同、独立工具链及拒绝边界；新的真实执行必须另行形成完整有限计划、冻结来源与二进制、保留此前消费并经过明确批准。[工单16](../../.scratch/barness-provider-sites/issues/16-zai-responses-native-source.md)不新增实际轮次、执行授权或付费预算，实际状态继续只由[规格](../../.scratch/barness-provider-sites/spec.md)记录。


## 中国 MiniMax Responses 的独立固定文本源

恢复5自身MiniMax-M3工具响应在相同high请求下没有reasoning事件或原生item；terminal usage省略reasoning_tokens，公共归一值为0。旧自身相同请求曾返回推理，而本轮full-stream原件保存150字Thinking。公共Client原SSE逐字回放复现差异，没有parser丢弃的证据；“请求了推理”不能替代“本回合实际提供原生推理”。

仅`minimax-cn-responses/MiniMax-M3`也在请求前固定自身full-stream为native-history来源。harness按同一固定合同保存原user/assistant前缀，随后保留独立工具用户/助手、Host结果及最终提示；严格消费者核对来源、作用域、精确前缀和真实summary/native wire。工具回合无Thinking仍是合法工具响应，但不能提供原生推理证明；指定文本源若空仍失败，即使另一工具回合有native也不得替代。生产协议、SDK、模型、guard和价格不改。

代价是多保存一个自身文本回合、按实际最终wire计量输入，以及消费者合同进入M3执行来源而再次使三M3/六地区formal旧资格过期。GLM245字节不受此改动；八generic保留另须明确已审来源差异与公共wire/账本不变证明，不能自动扩张旧白名单。原FAIL/UNKNOWN/消费与美元预留不改，新的真实有限范围与授权另行准备。[21实施和原件](../../.scratch/barness-provider-sites/issues/21-minimax-cn-responses-native-source.md)，唯一实际验收状态仍见spec。

## MiniMax 验收的本地授权与连接时限

恢复6的国际M3 Chat两次REQUEST失败原件明确记录`ConnectTimeout (15s)`，各次exchange约25.5秒后返回；原生授权门禁位于Client已开始计时的Host transport内，需要在下游建连前逐次重新验证完整历史。原件的recorder／账本记录不能证明厂商已接收请求。公共Client受控门禁以有界本地阶段复现相同错误且vendor到达数为0；保持授权阶段不变、仅调整真实harness policy可恢复成功。

仅MiniMax验收harness将`ConnectTimeout`设为其既有`CallTimeout=120s`；`ResponseHeaderTimeout=90s`、`ReadIdleTimeout=90s`、逻辑调用120秒、重试和次数上限不变，生产SDK及其他Provider默认值不变。连接与本地校验合计仍受90秒响应头／120秒调用时限约束，可能继续按真实失败停止；此调整不保证任何厂商延迟。每次Ledger.Attempt继续在最终请求抵达下游transport前完整校验，不迁出门禁、不缓存、不预领凭据或跳过原件。代价是MiniMax验收的连接阶段不再单独受15秒限额约束；公共回归同比例缩放同一policy工厂，核对本地授权、一次真实账本记录、vendor到达和资源释放。问题与资格影响见[工单24](../../.scratch/barness-provider-sites/issues/24-minimax-native-authorization-timeout.md)。

固定恢复7的历史继续增加，外层进程与消费者虽已采用3600秒，真实report初始化仍保留1800秒、harness尾部仅留20秒，会先于外层截止。仅在`live/formal/M3`且既有`RenewalAuthority(固定7,true)`成功、此前配置无错时，报告截止改为started+3600秒；研究、未批准ID、旧轮和非live模式保持1800秒。真实harness在同一已验证报告上取报告与Go测试截止的较早值，再固定预留120秒给未缓存的完整尾部`VerifyCapture`。旧轮保持报告截止减20秒、测试截止更短时减1秒的既有规则。证据清单使用实际`test.timeout`值，不从轮次环境变量推导授权。所有调用、HTTP、token和美元上限保持；代价是第7轮总进程墙钟边界扩大，不能被解释为付费许可或资格通过。[工单26](../../.scratch/barness-provider-sites/issues/26-recovery7-fixed-continuation.md)的最终冻结与直接人工批准仍是独立前置。
