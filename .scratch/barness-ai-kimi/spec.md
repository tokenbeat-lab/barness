# barness-ai 接入 Kimi：Responses、Chat Completions 与 Messages

Status: claimed
Type: feature
Updated: 2026-10-23

本文件定义交付契约。中国站和国际站均在范围内；设计就绪不表示已经接入、型号已晋升或真实支持已通过。实施细节见[设计文档](design.md)，官方事实及冲突见[研究报告](research.md)。

## Problem Statement

Host 希望通过现有 barness-ai Client，在本地和云端调用 Kimi 的 Responses、Chat Completions 或 Messages，并沿用统一的消息、流事件、结果、工具往返、用量和隔离语义。目前模块已有三套协议 adapter 及 DeepSeek 的兼容厂商接入结构，但没有 Kimi 的独立 Provider 身份、协议能力配置、候选目录和自身支持证据。

仅替换 URL 会沿用不适用的 OpenAI 或 Anthropic 自动字段。Kimi 三协议的推理字段、工具选择、缓存位置和用量口径不同；Messages 的 Bearer 认证也不同于当前 adapter 的默认接线。中国站和国际站账户、Key、价格及资格独立，不能通过一个品牌身份或故障切换混用。

K3 自动缓存会产生写入费用。请求中的 TTL 不一定等于实际写入 TTL，Chat 又不返回实际 TTL；即使用量齐全，也不能把单一估值误称为费率已经确定。人民币价格直接进入美元成本字段同样会误导 Host。

## Solution

沿用 DeepSeek 的结构：Provider 与 API Protocol 分开，复用现有三套协议 adapter、SDK 和公共调用生命周期，通过协议内部不可变配置及必要的 Kimi 编码、校验、解析分支承载差异。新增国际站 `moonshot` 和中国站 `moonshot-cn`，不增加 Site 维度。

首批交付 `(chat, Provider, API, kimi-k3)` 六个独立 Model 身份。每条路线支持 Stream、Complete、StreamSimple、CompleteSimple 及对应 Hooks；范围为文本聊天、函数工具、推理输出和完整历史回放。输入能力暂声明为文本，支持的输出上限固定为 131072、上下文元数据为 1048576；这是接入范围的收窄，不是厂商能力上限声明。

Host 保持账户授权、上下文组织、工具执行和凭据存储职责。每个地区、协议有独立 Binding，使用标准按量 API。完成候选、公共 E2E、自身真实资格及正式目录阶段后，再更新支持矩阵；六条目标不能通过删去失败路线缩小为“已完成”。

## User Stories

1. As a Host developer, I want Kimi Responses through the existing Client, so that my Responses integration can use Kimi.
2. As a Host developer, I want Kimi Chat Completions through the existing Client, so that my Chat integration can use Kimi.
3. As a Host developer, I want Kimi Messages through the existing Client, so that my Messages integration can use Kimi.
4. As a Host developer, I want all four chat entry points on each route, so that delivery style does not change supported behavior.
5. As a Host developer, I want existing protocol options and Hooks retained, so that Kimi does not require another public calling interface.
6. As a Host developer, I want mismatched protocol options rejected, so that a Binding cannot silently reinterpret my request.
7. As a local Host operator, I want explicit standard API configuration, so that I can use my authorized Kimi account.
8. As a cloud Host operator, I want one Client to serve multiple tenants and both markets, so that shared runtime does not imply shared authority.
9. As a tenant administrator, I want independent China and international Provider identities, so that credentials and evidence cannot cross markets.
10. As a tenant administrator, I want a separate Binding per protocol, so that I can revoke one route independently.
11. As a tenant administrator, I want model and actor authorization before secret access, so that denied calls do not read credentials.
12. As a tenant administrator, I want credential ownership and configuration versions checked, so that one Logical Call uses a consistent snapshot.
13. As a Host operator, I want same-account credential sharing to be explicitly authorized, so that protocol reuse cannot grant new access.
14. As a Host operator, I want official endpoint presets checked against the selected market, so that a configuration mistake cannot send a Key to the other service.
15. As a Host operator, I want trusted gateway configuration to remain explicit, so that existing controlled deployments remain possible.
16. As a Host operator, I want no automatic market, Key, model or protocol fallback, so that failures do not change account or billing attribution.
17. As a Host developer, I want six independent model records, so that one market or protocol can evolve without changing another.
18. As a Host developer, I want candidate and formal models distinguished, so that offline success is not presented as production qualification.
19. As a Host developer, I want the initial text-only scope stated accurately, so that it does not imply K3 lacks multimodal capability.
20. As a Host operator, I want a supported output cap and predictable defaults, so that enabling Kimi does not select a million-token default.
21. As a Host developer, I want K3 low, high and max effort encoded per protocol, so that the chosen setting reaches the correct field.
22. As a Host developer, I want omitted effort to retain K3's max default, so that generic defaults do not disable thinking.
23. As a Host developer, I want Simple canonical levels clamped using the existing contract, so that protocol-neutral calls stay predictable.
24. As a Host developer, I want unsupported Full wire values rejected, so that requested constraints are not silently weakened.
25. As a Host developer, I want explicit thinking disablement rejected in this supported subset, so that undocumented behavior is not promised.
26. As a Host developer, I want fixed sampling constraints validated, so that generic options do not create invalid K3 requests.
27. As a Host developer, I want reasoning deltas retained separately from visible text, so that streaming and final results preserve their meanings.
28. As a Host developer, I want Chat preserved reasoning replayed with tool history, so that the next turn retains required context.
29. As a Host developer, I want Responses summary history replayed without encrypted-reasoning requests, so that stateless continuation uses Kimi's documented shape.
30. As a Host developer, I want Messages thinking and actual signatures preserved, so that replay does not fabricate provider state.
31. As a tenant administrator, I want unsigned native thinking to retain provenance checks, so that missing signatures do not bypass isolation.
32. As a tenant administrator, I want cross-account, cross-market and cross-protocol state downgraded by existing rules, so that visible history can continue without transferring native authority.
33. As a Host developer, I want function tools and matching results on each route, so that I can continue my Host-managed tool loop.
34. As a Host developer, I want tool-choice restrictions checked per protocol, so that Chat required and Messages any are not falsely promised on Responses.
35. As a Host developer, I want opaque tool IDs preserved and mapped without collisions, so that punctuation or long IDs cannot mispair results.
36. As a Host developer, I want malformed tool JSON and contradictory terminal data reported, so that incomplete calls cannot appear executable.
37. As a Host developer, I want complete history replay, so that I do not depend on stored responses or provider conversations.
38. As a Host developer, I want Messages sent through its stable endpoint with one Bearer credential, so that SDK defaults do not add incompatible authentication or beta requests.
39. As a Host operator, I want cache controls mapped to their actual protocol locations, so that a setting is not silently ignored.
40. As a Host developer, I want short and long cache preferences to mean 5m and 1h on Kimi, so that OpenAI's 24h or 30m values are not reused.
41. As a Host developer, I want an explicit cache-none request rejected when disabling is unavailable, so that automatic paid writes are not misrepresented as disabled.
42. As a tenant administrator, I want derived routing identifiers scoped to tenant, Provider and account, so that raw tenant/session identifiers are not sent as cache keys.
43. As a Host operator, I want the provider's org-level cache scope disclosed, so that routing keys are not mistaken for hard tenant isolation.
44. As a Host developer, I want cache read, write and uncached tokens normalized per protocol, so that input is neither omitted nor charged twice.
45. As a Host developer, I want required missing counters distinguished from zero and optional cached reads defaulted to zero, so that usage remains robust and meaningful.
46. As a Host developer, I want reasoning tokens counted inside output, so that cost does not double count thinking.
47. As a Host operator, I want cost certainty separate from usage completeness, so that complete token counts do not hide an unknown cache TTL.
48. As a Host operator, I want bounded cost estimates when known facts constrain uncertain quantities or rates, so that a valid call remains usable without an invented exact cost.
49. As a Host operator, I want China prices converted with an explicit historical FX snapshot, so that CNY values cannot be read as USD rates.
50. As a Host operator, I want catalog and source versions pinned, so that rates and estimates can be reproduced.
51. As a Host developer, I want callback payloads validated after customization, so that Hooks cannot expand supported capabilities or authority.
52. As a Host developer, I want mutable inputs copied at existing snapshot seams, so that later changes cannot affect active calls.
53. As a Host operator, I want quota exhaustion separated from transient throttling, so that HTTP 429 does not cause useless retries.
54. As a Host operator, I want bounded initial retries and no stream replay, so that recovery does not hide duplicate generation.
55. As a Host developer, I want cancellation, timeout and malformed-stream failures to settle once, so that my application can release resources reliably.
56. As a Host operator, I want existing admission, byte and queue limits retained, so that Kimi cannot bypass resource controls.
57. As a Host operator, I want safe attribution and request identifiers where verified, so that failures can be diagnosed without logging prompts or secrets.
58. As a maintainer, I want repeatable public Client E2E artifacts, so that protocol behavior can be reviewed without paid calls.
59. As a maintainer, I want each route's own finite live evidence, so that success on a shared adapter cannot qualify another market.
60. As a maintainer, I want existing Providers covered by affected regression tests, so that shared changes preserve their public contracts.
61. As a maintainer, I want one current feature status with immutable evidence pointers, so that planning, implementation and release cannot be confused.
62. As a Host operator, I want a documented disable-and-rollback path, so that Kimi can be withdrawn without rewriting history or other Providers.

## Implementation Decisions

### R01 — 身份、范围与目录

新增 `ProviderMoonshot=moonshot`（国际站）和 `ProviderMoonshotCN=moonshot-cn`（中国站）。首批 ModelID 均为 `kimi-k3`，Operation 为 chat，API 使用已有三个协议标识。每条记录独立拥有能力、默认值、约束和价格；协议配置可以共用实现，实际身份始终保留。

首批 Input 为 text、ContextWindow 为 1048576、MaxTokens 为 131072、Reasoning 为 true、默认规范档位为 max。Off、minimal、medium、xhigh 标为不直接支持，low/high/max 映射同名 wire 值；Simple 维持向上优先的已有 clamp。Host 目录可收窄，不能为 Kimi 打开尚未接入的型号、模态、输出上限或协议能力。

所有候选与正式条目使用各自的完整身份、版本和 Hash。正式纳入须有自身候选资格、目录审阅和晋升后自身验证；不能拿研究、受控回放或另一条路线的 PASS 代替。

### R02 — Host、Binding 与数据流

Host 为每站三协议配置独立 BindingID、固定官方目标、账户、凭据引用及标准 API 用途。凭据元数据同时记录实际 Provider、Tenant、Account Scope、用途、Allowed APIs 和版本；在读取秘密前校对。两站不得共用凭据记录，同站同账户且获准的协议可以共享凭据引用。

沿用 scope → Binding/Model 授权 → 选项与型号约束 → 凭据与一致性 → 请求构建/回调/最终校验 → Admission → HTTP 的顺序。命名选项拒绝不得读取凭据；回调后的拒绝发生在读取凭据之后、HTTP 之前。官方 Host 示例核对 URL；核心保留已有可信 endpoint/Transport 注入合同。

### R03 — 协议复用与最终校验

只扩展现有三个协议内部配置和必要 codec；不建立 Kimi 专属公共 Client、新 adapter 生命周期、厂商插件注册中心或 Host 可写能力表。既有 SDK、取消、限额、队列、Observer、初始重试和资源释放继续共享。

Kimi 的自动字段采用显式配置。型号、SamplingParams、Hooks 均不能扩大该协议已实现的能力；未知或未接入字段在最终请求校验时拒绝。先保留现有授权错误，再报告能力/格式错误；错误不回显不可信字段值。

### R04 — Chat Completions

使用 system/user/assistant/tool 消息、`max_completion_tokens` 和顶层 `reasoning_effort`；省略 effort 保持 max。请求 stream usage，读取正文、reasoning_content、工具参数和末尾用量；finish_reason 后仍接收合法 usage。

支持 tool_choice 的省略、auto、none、required；具名强制和 allowed-tools 形式不在首批合同内，发送前拒绝。回放原 assistant 推理、工具声明与匹配结果，禁止合成缺失推理。固定采样值仅按设计列出的公开范围接纳；不发送 store、developer 角色、thinking toggle 或 DeepSeek 专属字段。

### R05 — Responses

使用真实 Responses endpoint，推理写入 `reasoning.effort`，工具结果以 call_id 配对；只支持省略或 auto 工具选择。完整历史回放，不发送服务端续接、存储、后台、加密推理、summary 请求配置或 service tier。

读取 reasoning summary 事件和终态 output，保留原始原生条目，按独立输入 schema 投影回放 summary。不能把输出条目的所有字段、OpenAI 合成 ID 或 encrypted_content=null 原样当成必要输入；投影不得改写原始状态。completed、incomplete、failed、error 分开处理，不把 HTTP 200 当成功证明。

### R06 — Messages

使用稳定 Messages 资源、单一 Bearer 头和已核对的版本头；不发送 beta query、原厂浏览器头或 beta。effort 独立写入 `output_config.effort`，不借 adaptive、预算或托管逐轮 effort 功能生成。

支持省略和 auto/any/none 工具选择，拒绝具名工具、未接入的 thinking 选项和结构化输出扩展。保存实际 thinking/signature 并按原生来源回放；不存在的签名不补造。Messages 的“关闭 thinking”与 K2.x 文档冲突由限定 K3、只承诺 effort 的首批合同解决，不尝试未声明的关闭形状。

### R07 — 工具历史、原生状态与隔离

Host 负责工具执行。三协议保留完整工具往返；不透明 call/item ID 需要可逆且无分隔歧义的内部编码，并在转换后维持调用与结果的一一配对。跨模型归一化不得让两个调用碰撞。

Native State Envelope 继续精确比较 Tenant、Account Scope、Provider、API 和 Model。Messages 的 unsigned thinking 作为原生块回放时同样需要来源担保；跨域仅沿现有安全降级保留允许的可见内容，不能附带签名、密文或原生引用。持久化恢复仍由 Host 信任入口负责。

### R08 — 缓存偏好与租户信息

默认/short 明确编码 5m，long 编码 1h；Chat/Responses 使用 implicit 缓存选项，Messages 使用顶层 ephemeral 控制。明确的 none 拒绝为不支持，不静默映射为自动付费缓存。long 被 Host 型号约束关闭时也明确拒绝。

有 SessionID 时 Chat/Responses 使用现有 v2 派生键，作用域含 Tenant、实际 Provider、Account Scope；无 SessionID 不合成全局键。Messages 不添加未声明的会话头。厂商 org 缓存不是本库可保证的租户隔离；需要厂商侧硬隔离时由 Host 使用独立 org/account。

### R09 — 用量与成本范围

Chat/Responses 的 Input 由总输入扣除缓存读写；Messages 的 input_tokens 已是不含读写的输入。推理是 Output 的子集。所有报告计数须是非负整数且满足包含关系，溢出或矛盾为 protocol 错误；必要报告缺失与零分开保留。Chat/Responses 的 cached_tokens 是可选字段，缺失/null 默认0，默认本身不降低报告完整度；Chat 先保留 nested 与 legacy 顶层回退。显式非法读取量仍拒绝，Messages 的独立字段维持原合同。

Kimi 的收费写量属于必要报告项，缺失不得宣称完整；其他 Provider 继续按原合同判定。Messages 的 thinking token 明细仍是可选项。保持失败响应、重试间不求和及已知用量留存的现有规则。

在 Attempt 上增加可选、只含数值和枚举的成本估算元数据，区分 resolved、bounded、partial、unreported。旧 Provider 默认省略，旧消息 Usage 数字形状不变。Kimi 收到完整收费计数且实际 TTL 可证时上下界相等；未知TTL或写量可由已知总量与价格约束时，计算所有允许计费分配的有限上下界，保留成功并明确 bounded；上下界相等则 resolved。Chat/Responses 已知总输入P、输出O和读取R时，缺失写量W满足0≤W≤P−R；用量仍partial，公开未知写量保持兼容零值，不伪造观察。必要总量/费率缺失、无法证明完整有限区间时继续partial/unreported。

Kimi 的 Usage.Cost 使用同一元数据定义的点估值：resolved 为该估值，bounded 为区间下界，partial 为现有缺项规则下的非完整估值；Host 展示和预算不能把 bounded/partial 的点值当成完整费用。UsageReporting 只表达 token 报告程度，与成本状态独立；重试许可的预算预留使用保守上界而不是点估值。

### R10 — 两站价格快照

两站分别保留原币定价、来源日期与 Hash。国际 K3 的每百万费率为输入/5m 写入 3 USD、1h 写入 6 USD、读取 0.30 USD、输出 15 USD；中国分别为 20/20/40/2/100 CNY。运行时仍用既有美元成本合同，不引入在线换汇。

中国首批明确复用仓库已留存的 2026-10-16 CFETS 历史快照：1 USD=6.7220 CNY；这是固定估算基准，不声称为当日汇率。基础 USD 费率按十位小数半偶舍入，1h 写入使用既有输入两倍算式；原币事实、换算规则及来源作为价格快照的留存输入。更新来源或费率须换版本，不能借国际价格填充中国记录。

### R11 — 错误、重试与可观测性

按 HTTP 状态和已识别的 Kimi error.type 分类。明确额度耗尽不可按瞬时 429 重试；限流/过载只能按 Binding 的有界策略重试，流开始后不重放。取消、超时、提前 EOF、无终态和非法工具参数必须产生单一可诊断失败，并释放许可、正文与队列资源。

诊断只保留固定错误类别、HTTP 状态、已核对的 request ID、Retry-After 和既有归属，不回显厂商自由文本、prompt、工具参数或凭据。实际 request ID 头只根据自身官方资料/捕获登记，不猜测签名头用途。

### R12 — 交付、回滚与兼容

交付包括六个候选身份、Host 本地/云端装配示例、公共 E2E 夹具和重放产物、自身真实资格入口与只读证据消费、正式目录及文档。复用现有证据结构；不建立固定全库调度器、不恢复 pi oracle/差分门禁或历史恢复控制器。

仅新增 Provider 常量和可选 Attempt 成本字段等必要公共扩展，现有 Target、Binding、凭据、消息、Native State Envelope 和调用方法保持。涉及公共结构位置式 literal 的调用者须迁移；其他 Provider 的字段、默认值、估价和历史行为由回归锁定。

回滚先禁用 Kimi Binding，再从旧进程将加载的 Catalog 移除 Kimi 身份，随后回滚代码；持久化历史保留原身份，不映射成 OpenAI、Anthropic 或另一地区。旧证据和已消费授权保持原件。

## Testing Decisions

用户于 2026-10-22 明确选择：**以现有公共 Client E2E 为主要测试 Seam**。Host Resolver、Catalog、SDK、Transport、assembler、Result、Observer 与资源 Probe 均通过同一公共调用验证，不增加私有 helper 单测或专用测试 Client。

好测试验证最终请求、可见事件/结果、工具配对、隔离、计数与资源释放；不绑定内部函数或分支数量。先留存失败场景再实现。可控服务与 channel/hook 决定取消、并发和故障时点，正常验收不顺带跑设计负载或厂商请求。

| 验收组 | 必须证明的外部行为 | 对应需求 |
| --- | --- | --- |
| K01 六路线与四入口 | 每个完整身份实际被选择；Stream/Complete 与 Simple 均有完整事件/Result | R01–R03 |
| K02 授权与快照 | 越权先拒绝、拒绝不读秘密、不发 HTTP；版本冲突；并发账户不串线 | R02、R07 |
| K03 最终请求 | 精确 URL、稳定 Messages、单一 Bearer、允许字段及默认 presence | R03–R06 |
| K04 档位与预算 | low/high/max、默认 max、Simple clamp、Full 拒绝、131072 上限和固定采样 | R01、R04–R06 |
| K05 工具选择 | 三协议允许值与拒绝值；强制模式不被静默改为 auto | R04–R06 |
| K06 工具往返 | 工具 JSON 分片、空/多工具、opaque ID、结果配对和跨模型碰撞处理 | R07 |
| K07 推理与历史 | 实际 reasoning/summary/thinking/signature 回放；不补造状态、不混输入输出 schema | R04–R07 |
| K08 原生来源 | 持久化信任、缺封套/错账户/跨站/跨协议降级；unsigned thinking 同样受限 | R07 |
| K09 缓存请求 | short/long/none、顶层 Messages、v2 派生、最终 hook 不绕过 | R08 |
| K10 用量 | 可选读取默认0、必要写量缺失、读写互斥、包含关系、显式零、溢出与负数；末帧 usage 不丢 | R09 |
| K11 成本 | 已知/未知TTL、缺写量有限分区、相等/反转费率、分档配置仍拒绝、零写量、partial/unreported、价格/FX 来源 | R09–R10 |
| K12 故障与取消 | 错误终态、无终态、非 JSON HTTP 错误、额度/限流、取消/超时和有界重试 | R11 |
| K13 共享运行时 | 请求/帧/输出/工具/队列限额、背压和许可归零；双站并发归属 | R03、R11 |
| K14 Hooks 与快照 | 认证与授权字段受保护、最终校验、输入修改不影响活跃调用、成本元数据独立 | R02–R03、R09 |
| K15 回归与回滚 | 受影响既有 Provider 请求、历史、成本、Observer 和目录保持；禁用/目录移除有效 | R12 |
| K16 自身真实资格 | 六路线各自完整有限场景、当前来源、真实历史续接及脱敏可重放捕获 | R01、R12 |

已有先例是 DeepSeek 两条路线的公共协议夹具、三协议共享资源套件，以及地区 Provider 的身份、Host 装配、目录、原生历史和证据回放场景；复用其测试 Seam 和产物，按 Kimi 合同新建自身案例。

每个普通 E2E bundle 留存实际请求、服务脚本/响应、事件、Result、Attempt、Observer、资源 Probe、断言、manifest/audit、来源 Hash 和精确重放入口。合成与真实捕获明确区分；回放不访问厂商，也不产生真实支持资格。

按测试 SOP 分阶段：先普通 E2E → 冻结来源 → 获准的自身 live → 人工选择 race → 人工压力或有依据的不适用 → 只读发布判定。普通范围包含完整受影响公共 E2E 包及核心包；新增共享成本元数据要求覆盖既有 Provider 的序列化、快照和 Observer。race/压力的候选路径及有限 live 场景见设计文档，本文不授予付费、race 或压力执行许可。

## Out of Scope

- Kimi 自有协议、Kimi Code/会员套餐、GLM 或其他 Provider 新功能。
- legacy `/v1/completions`、K2.x、工具专用型号别名和厂商 moving aliases。
- 图像/视频/音频接入、图像生成或分类操作；首批文本 Input 不声称厂商不支持这些能力。
- 超过 131072 的输出支持、百万 token 容量实测或任意参数组合保证。
- Structured Output/严格工具采样、namespace/custom/apply_patch、托管搜索/MCP、远端文件和公网图片引用。
- 保存/检索 Response、previous_response_id/conversation、后台执行、自动记忆、自动工具执行和自动断流续写。
- 新增通用网关、运行时能力发现、实时价格/汇率、跨地区/Key/协议 fallback、兼容 shim 或双写。
- 改变其他 Provider 已有缺项、失败 usage、估价或正式资格规则。

## Further Notes

设计依据为领域词汇及 ADR-0001、0010、0014、0015、0018、0020、0032、0049、0051。沿用 DeepSeek 的协议复用与 Provider 能力边界；历史差分门禁已被 ADR-0051 取代，不能照搬。新增成本范围的公共取舍记录于 [ADR-0052](../../docs/adr/0052-barness-ai-kimi-cost-estimates.md)。

官方文档冲突采用明确首批子集处理：仅 K3；Messages 只承诺 effort；认证选择参考页明确的 Bearer；输出与模态上限收窄；Chat 缓存 TTL 不可证时返回成本区间。新增官方事实或真实捕获与本合同不符时，保留失败，更新研究/设计及受影响场景，不偷偷切换型号或删除目标。

### 唯一当前状态

本节是唯一可变状态入口。2026-10-23用户明确选择费用保守区间，并指定 cached_tokens 可选、未返回补0；[08](issues/08-live-usage-reporting.md)已完成实现与普通验收。修订前状态字节保留于 [快照](history/spec-before-cache-default-bounds-20261023.md)，既有真实失败及授权不改写。

| 项目 | 当前状态 |
| --- | --- |
| 集成分支/修订基线 | codex/barness-ai-kimi-integration；本轮固定比较点2d55fcfa8957f430cb2b9950942429a58fc14a19，原大功能基线68b7da0a3400f909dc535d7f585c93771df1569c |
| 当前实现范围 | Kimi Chat/Responses 两站可选 cached_tokens 默认0；独立判断 token 完整性与可证明费用界；缺写量有界分配；真实SSE回放；Result/Observer一致；资格断言与消费者采用同一成本合同。Messages及其他Provider维持原合同 |
| 公共验证入口 | 继续采用用户确认的公共Client/CLI E2E。先留存真实捕获/缺项/自定义费率等失败，再实现并转绿；不增加私有helper单测 |
| 普通验收范围 | 完整受影响包：go test -json ./ai/e2e ./ai ./ai/internal/testkit/supportmatrix ./ai/internal/testkit/release ./ai/live/cmd/regionalverify ./ai/live/cmd/regionalpromote -count=1 -timeout=20m；通过SOP run.py固定全部实际输入。[最终有限清单](evidence/ordinary-selection-05.json)在运行前冻结：14045个必需ID，214个重负载NOT_RUN |
| 旧普通事实 | 实际受测0c2a03dfc177699cf0163fa4b7b76e61e17fea71，13956原生PASS/214重负载NOT_RUN；[原核对](../../.evidence/barness-ai-kimi/acceptance/04/verification.json)。此为旧来源历史，未用来替代本轮新来源的完整普通验收 |
| 本轮普通/审查 | 冻结实现3a722ba，交付02aea23；398个定向公共用例PASS，[实施证据](evidence/cache-bounds-implementation-01.md)。[双轴审查](evidence/cache-bounds-review-01.md)均PASS/0发现；最终完整普通实测2ffa761c437cdb6403cbaf05c7f1846e11965f49、exit0/PASS，14045必需用例全PASS、214重负载NOT_RUN；14913个实际Go PASS节点 |
| 实际输入 | 共135项：122份实际历史/价格/脚本输入、新native二进制和6份新grant草案、5份实际编排/审计工具及本次有限清单。85/6/5整目录集合另作前后审计；真实脱敏夹具在ai/来源内。新native来源22b841baaf4b6a61f982a29cef8ed4199bdb05748a413705944955911f406eea；新构建全exit0，凭据不进入输入/argv |
| 最终普通原件 | [调用/退出/墙钟](../../.evidence/barness-ai-kimi/freeze/05/invocation.json)、[回执](../../.evidence/barness-ai-kimi/e2e/20261023T011642.944924Z-ylkxot0p/receipt.json)、[逐项断言与五辅助审计](../../.evidence/barness-ai-kimi/acceptance/05/verification.json)。主审计零发现；辅助ai26、supportmatrix79、release15节点PASS，两CLI仅编译，实际行为由公共进程E2E覆盖。1351份来源、135项输入及85/6/5目录消费集合前后相等 |
| 最终普通Hash | receipt ab3bb1a3d00f225f48a798801f11f6cc20c66a405a4151c707880a7f54993210；manifest 99bb7f3b8a41a517752790181409b5637dd4b58ef6edb51ecc60a4827e314771；audit 8aa973100a15bf817985395637131e838cf84e427be4792583def923f020f5fa |
| 真实历史 | 旧来源7b41a21c96ee33238100ce8bea5218da480a6b6a的中国站Chat候选1调用/1Attempt、零重试，因cached_tokens缺失的旧成本规则FAIL；余8场景及其余五路线NOT_RUN，原生与外层消费者exit1。见[执行结果](evidence/candidate-live-result-01.md)及[原件](../../.evidence/barness-ai-kimi/paid-execution/01/summary.json) |
| 新live适用范围 | 六条Kimi成本/报告与资格来源受影响，须自身新候选与正式资格；保留共享历史涉及旧18条chat路线的既有影响判断。旧grant/来源不复用，不以离线回放补新真实资格 |
| 凭据/授权 | 中国站KIMI/CN使用KIMI_CN_KEY，国际站KIMI/GLOBAL使用KIMI_KEY；旧候选范围首失败停止。本次[具体提案03](evidence/candidate-live-review-03.md)六路线已获明确批准；[原始批准02](../../.evidence/barness-ai-kimi/authorizations/02/approval.json)绑定新来源/二进制/六份新grant及≤US$9。范围首失败即停，执行状态与原件在本次结果留存。正式/旧路线/race/压力未授权 |
| race/压力候选 | 均NOT_RUN；共享Usage快照、Observer与取消/Admission并发和受影响旧Provider在live后由人选择；工具/队列/背压等压力按SOP后续独立选择 |
| 工单/目录 | 01–06及08 resolved；07普通实现/审查已完成，仍等待自身真实资格、晋升和人工阶段。六路线正式目录仍inactive，不减少原目标 |
| 原件/回滚 | 所有旧失败/中断/批准/账本与捕获保留。旧八个实施worktree已清理；本轮6250份原件已逐项核对并保留；app两次归档请求仅queued，随后受控CLI清理exit0、checkout已移除，分支02aea23保留，见[清理回执](evidence/worktree-cleanup-06.json)。回滚先禁用Binding/移除加载目录，再回退本轮代码/合同，保留原生身份与账本 |
| 当前release.pass | false；[不可变索引04]历史绑定[执行前状态快照](history/spec-before-candidate-retry-20261023.md)，本次执行后新建索引。[索引04](../../.evidence/barness-ai-kimi/records/04/record.json)及其真实只读命令/报告保留为修订后普通验收的执行前事实。当前文档提交与受测2ffa761的完整来源/输入集合相等性另存[核对](../../.evidence/barness-ai-kimi/closure/03/source-equality.json)。索引03为旧来源历史，不刷新旧pin/状态 |

依据[测试SOP](../../docs/agents/verification.md)执行；当前修订决策优先于研究保留的旧schema必填标记。旧官方原文、旧真实失败及来源Hash不修改，当前软件归一化行为由修订合同与自身公共证据定义。

### 交付检查记录

2026-10-22 完成只读文档检查：Python 检查进程退出 0，核对 46 个本地引用及锚点、七个模板章节、连续 62 条用户故事、12 项需求、16 组验收、六条路线和每路线九个 live 计划场景；全部文件低于 500 行，围栏与行尾格式有效。FX 半偶换算及成本区间示例算术一致，既有 FX 原始 Hash 已核对；Git diff 格式检查退出 0。另人工对照研究、代码及 ADR 检查协议字段、Simple/Full 语义、单一当前状态、成本与用量两轴及候选/正式资格。

上述是规划文档质量检查，没有程序受测提交或新的运行资格；不引用既有 Provider 的 PASS 作为 Kimi 结果。未来运行阶段的真实命令与不可变回执另按 SOP 留存。
