# barness 地区 Provider 接入规格

Status: claimed
Type: feature
Updated: 2026-10-20

GLM 与 MiniMax 的中国站、国际站使用独立 Provider；每个服务分别配置型号、价格与授权，共享经过核验的协议实现。配套：[设计](design.md)、[研究](research.md)、[ADR-0049](../../docs/adr/0049-barness-ai-regional-provider-identities.md)。本文 R01–R20 是本功能的需求与验收合同。

## 当前状态

此节是本功能唯一的当前实施与验收状态入口。实施工单按依赖图推进；实现和普通通过不表示新增服务已取得自身正式支持。

| 项目 | 当前状态 |
| --- | --- |
| 架构选择 | 已确定方案 2，记录于 ADR-0049 |
| 本次交付 | 实现与普通验证已交付；工单 01–06 已解决，[07 最终验收](issues/07-final-acceptance.md) 进行中。分支 `codex/barness-provider-sites`，feature 保持 claimed |
| 最终续验准备 | 固定恢复3在实际6b87a584执行：三M3各13/13、ZAI Chat18/18 PASS；随后ZAI Responses完整18场景17 PASS/1 FAIL（glm-5.3/native-history的预固定tool-call源无推理），后19进程按约停止。[停止索引](../../.evidence/provider-sites/final-acceptance-20261020/recovery3-stopped/index.json)。14/15范围完成，[16来源选择修复](issues/16-zai-responses-native-source.md)已完成离线实现与六路线81受控场景；[17固定恢复4](issues/17-fixed-fourth-recovery.md)原生机制55公共进程场景通过，控制器[259工件](../../.evidence/provider-sites/final-acceptance-20261020/recovery4-controller-dev/index.json)与[138修复工件](../../.evidence/provider-sites/final-acceptance-20261020/recovery4-review-fix-dev/index.json)保留完整RED/GREEN，最终50/50通过（204.167s、exit0），三发现已由单一implementer修复且两轴均0；截至81e79b3离线复审结束，实际新付费授权4/grant4均无；后续授权/运行、实际冻结及根检查状态见[准备当前状态（CURRENT_STATE）](../../.evidence/provider-sites/final-acceptance-20261020/recovery4-preparation/index.json)与[具体计划](../../.evidence/provider-sites/final-acceptance-20261020/recovery4-preparation/plan.md) |
| 本轮 ZAI 原生推理研究 | 四次自身真实对照实际完成，4 Calls / 4 HTTP、零重试、合法终态；GLM-5.2 omitted 为 thinking0，enabled/budget1024/max4096 为 thinking97 / signature24，GLM-5.3 两形状均有原生块。工具 source 前后相同，Key 扫描零发现；只证实窄请求形状，尚未认证 native-history / 新候选。[固定研究事实](evidence/zai-thinking-research.md) |
| 文档核验 | 上轮收尾静态检查 exit 0，固定合同、格式 / 500 行及 334 处本地引用 PASS；该记录只证明其原来源；本次纯状态文档按引用、格式、状态与执行来源相等性核对；实际冻结及根检查状态见[准备当前状态（CURRENT_STATE）](../../.evidence/provider-sites/final-acceptance-20261020/recovery4-preparation/index.json) |
| 生产实现与目录变更 | 四服务各三协议、12路线/18身份；实现与Host隔离已完成。新增九身份已由六条真实candidate独立证明后原生晋升BuiltinCatalog，版本2026-10-20.1-regional，Hash sha256:9f00cfa579948ea506ce56eb1f901572b03ea64bfcd9ff4fcae3cb5347a674a1；晋升后formal支持仍待认证 |
| 上轮完整普通 E2E | 受测提交 `d4b63d9`；完整普通主包 13,244 原生场景全部 PASS，双 vet、差分、压力与原生审计通过；[该来源验证索引](evidence/verification.md) / [机器索引](evidence/index.json)。本轮新增源码尚未执行完整普通门禁 |
| 双轴审查 | 既有生产/Host/晋升及前轮修复审查保留。固定恢复3完整差异b22e013...6b87a584的Standards三项、Spec一项已由单一implementer修复，最终均0剩余；[最终双轴索引](../../.evidence/provider-sites/final-acceptance-20261020/recovery3-preparation/review/index.json)。16/17完整Go/Python差异经独立整合审查；Standards两项、Spec一项由同一implementer在81e79b3集中修复，最终Standards0/Spec0，无新增剩余。[本轮最终双轴索引](../../.evidence/provider-sites/final-acceptance-20261020/recovery4-preparation/review/index.json)SHA `8ca4aab7dcc940c481dc2fc17a27fb146029f6c8a6527ebad194ce794b9295dc` |
| 本轮定向普通 | 固定恢复3的[原生开发证据](../../.evidence/provider-sites/final-acceptance-20261020/recovery3-native-dev/index.json)与[控制器证据](../../.evidence/provider-sites/final-acceptance-20261020/recovery3-controller-dev/index.json)保留RED/GREEN、原件和219工件Hash。根6b87a584十新二进制、双vet与完整ai/live无Key检查exit0（包210.058s），[检查](../../.evidence/provider-sites/final-acceptance-20261020/recovery3/checks.json)。新增16的[永久索引](../../.evidence/provider-sites/final-acceptance-20261020/responses-native-source-fix/index.json)保留公共RED、六路线81受控场景及原生消费者；17的[原生索引](../../.evidence/provider-sites/final-acceptance-20261020/recovery4-native-dev/index.json)固定clean07807a1、55公共进程场景PASS（墙钟121.614878s）、目标vet及无Key事实。旧原生/controlled回归均通过；控制器259工件和集中修复138工件固定最终50/50、204.167s、exit0，覆盖旧R2/R3与新R4公开流程。三项review修复未改Go/guard/GLM，执行集合仍GLM245/MiniMax320/Regional414。这些离线结果不替代失败路线真实资格或完整正式普通门禁 |
| 本轮新身份自身真实验证 | 六candidate同ca03339全部81/81并原生晋升，六formal同600a4ec全部81/81历史保留。恢复3新409来源仅ZAI Chat18/18完整PASS；ZAI Responses17/18导致整轮停止，另外四formal未启动。16已同步harness与严格消费者；17原生机制加入后公开API实算Regional414、MiniMax320，六formal及三M3仍须按最终冻结源完整重采，不刷新旧证明 |
| 上轮新身份自身真实验证 | 六条各一次 candidate 已完成：81 Calls / 81 HTTP、零重试，80 PASS / 1 FAIL，六份 audit 零发现。ZAI Chat / Responses 与中国 MiniMax 三协议的 qualify 通过；ZAI Messages 因 [`glm-5.2/native-history`](issues/08-zai-messages-glm52-native-qualification.md) 失败未通过。未晋升 / 未运行新 formal。[固定汇总](../../.evidence/provider-sites/acceptance-20261020/candidate-summary.json) |
| 旧真实来源适用性 | 三中国GLM实际b384各完整22 PASS、raw/views公开消费且245项来源未变。恢复3三国际M3同6b87a584各完整13 PASS，三个原生正式消费者exit0；其余14旧路线与四效果没有本轮新资格。16的supportmatrix来源规则和17机制已改变MiniMax执行集合，当前M3成功仍保留原身份/来源历史，但不能冒充修改后的自身资格。07807a1公开API确认全部GLM245字节仍等b384，81e79b3控制器修复后GLM245/MiniMax320/Regional414字节集合不变；来源选择器、生产协议/SDK/目录及Go资金guard未改 |
| 正式前置 | 13与14已分别以获批的新完整真实资格解除；15固定恢复3机制及失败即停已实际验证。[16](issues/16-zai-responses-native-source.md)已离线修复ZAI Responses预先固定自身full-stream推理源，保留独立工具/Host链与严格原生断言；[17](issues/17-fixed-fourth-recovery.md)固定ID4/schema5及完整关闭历史的原生实现通过55场景。控制器最终50/50及独立两轴0已完成；后续授权/运行、实际源码/十binary/工具/计划冻结与根检查状态见[准备当前状态（CURRENT_STATE）](../../.evidence/provider-sites/final-acceptance-20261020/recovery4-preparation/index.json)及[具体计划](../../.evidence/provider-sites/final-acceptance-20261020/recovery4-preparation/plan.md)。新24范围506/518、累计1070/1130及US$30内17.30＋12.60＝29.90分配。截至81e79b3离线复审结束尚无实际新付费授权4/grant4，未读Key或调厂商；该阶段READY/race NOT_RUN、release.pass=false |
| race | 上轮 `470682d3…` 完整开发 race PASS：exit 0、墙钟 2589.635s、主包 13,244 原生 PASS、32 audit 零发现、无 race detector 警告；不是正式 check checkpoint。[固定报告](evidence/acceptance.md) / [机器索引](evidence/acceptance-index.json)。本轮新增源码 race NOT_RUN，取得新 READY 后一次完整正式 race 已明确获批 |
| release.pass | false；完整正式验收继续开放 |
| 实际厂商 / 付费调用 | 旧US$5轮次81/81已结束且不转用。US$30首轮/R2及已批Responses追加实际303/302、claimed456/492、整笔10.70关闭保留；恢复3实际新增75 Calls/72 HTTP（M339/36、ZAIChat18/18、ZAIResponses18/18含FAIL），claimed108/120、整笔1.90。跨轮实际378/374、claimed564/612、整笔US$12.60（非账单），父UNKNOWN3、退款0。后19进程未启动；无自动完整重跑、额外研究或最终race |
| 上轮完整受测来源 | 上轮完整普通/开发race集合470682d37e2d1c839bf5609f6898f9a1e4dafbd3dbfcd2cb77f2048a530f3e05仅证明原来源。候选实际ca03339、三GLM实际b384951保持；恢复3实际6b87a584，Regional409/MiniMax315及完整操作来源1233文件固定。[准备记录](../../.evidence/provider-sites/final-acceptance-20261020/recovery3-preparation/index.json)。16/17离线新源07807a1为Regional414/MiniMax320，GLM245相等；81e79b3控制器与最终两轴复审已完成，这些离线事实不构成最终发布证明；实际冻结及根检查状态见[恢复4准备当前状态（CURRENT_STATE）](../../.evidence/provider-sites/final-acceptance-20261020/recovery4-preparation/index.json)与[计划](../../.evidence/provider-sites/final-acceptance-20261020/recovery4-preparation/plan.md) |


现有 Provider 的原始支持与消费记录保持各自适用范围；本表不撤销历史事实，也不把历史通过计为本功能通过。实施后在此维护结果及原件指针，其他文档只链接此状态。

## Problem Statement

两个市场可能使用相同型号名，却有不同的型号集合、能力、默认参数、价格、账户或凭据资格。当前 Model 元数据按 Provider、协议和型号确定；如果只在 Binding 增加站点或换 Endpoint，同一个目录身份仍共用一份元数据，不能完整表达差异。

barness 同时面向本地与云端共享 Client。Host 需要显式选择服务，按租户授权，并保证原生历史、缓存路由、凭据快照和失败处理不会因同品牌而串用。新市场的支持证明也必须有自己的实际服务来源。

## Solution

保留 `zhipu` 为智谱中国站、`minimax` 为 MiniMax 国际站，增加 `zai` 和 `minimax-cn`。使用既有 `(Operation, Provider, API, ModelID)`、Catalog、Binding 和 Credential Resolver，不增加 Site 或品牌维度。

共享三套协议 adapter 的生命周期，通过协议内部不可变配置选择已核验的请求与解析行为。Model 数据独立维护，Provider 保留到最终请求、消息来源、事件、用量和支持证明。缓存派生作用域增加 Provider 并统一升级 v2。

## User Stories

1. 作为 Host 开发者，我可以在一个 Client 中同时调用同品牌的两个市场，不必创建可变的全局站点选择。
2. 作为型号维护者，我可以为不同 Provider 的同名型号分别声明能力、思考档位、采样默认、限额和价格。
3. 作为租户管理员，我可以单独授予或撤销某个市场、某种协议和某个型号，不影响另一站。
4. 作为 Host 开发者，我通过同一套聊天入口和工具循环使用四个 Provider，收到准确的事件与结果归属。
5. 作为账户管理员，我能明确选择账户、凭据来源与标准 API 或套餐用途，配置错误不会触发隐式换站。
6. 作为会话使用者，我可以显式切换市场继续可见对话，原站签名和账户专有状态不会被当成新站原生状态。
7. 作为云端运维人员，我可以并发使用不同市场，即使账户别名、SessionID 或其他租户内标识相同，也能保持正确隔离。
8. 作为维护者，我能复用已实现的协议能力，同时用新服务自己的证据判断支持范围。
9. 作为费用使用者，我得到与实际 Provider、型号和目录版本一致的 USD 估算，不把缺价或缺用量当免费。
10. 作为维护者，我能通过公共 Client E2E 与保留工件复现行为，并能禁用或回滚新服务而不自动重路由。

## Implementation Decisions

### R01 服务身份

增加 `ProviderZAI ProviderID = "zai"`、`ProviderMiniMaxCN ProviderID = "minimax-cn"`。现有 `ProviderZhipu = "zhipu"`、`ProviderMiniMax = "minimax"` 的含义与序列化值保持不变。Provider 区分服务契约，API 区分协议；同品牌、共享 codec 或相同 wire ModelID 均不构成身份相等。

### R02 首批路线与型号

Operation 固定为 chat；每个服务分别接入 `openai-completions`、`openai-responses`、`anthropic-messages`。首批验收集合如下，三种协议都必须包含表中每个型号：

| Provider | 目标 ModelID | 路线数 | 型号身份数 | 本功能性质 |
| --- | --- | --- | --- | --- |
| `zhipu` | `glm-5.3`、`glm-5.2` | 3 | 6 | 原有服务回归 |
| `zai` | `glm-5.3`、`glm-5.2` | 3 | 6 | 新服务独立接入 |
| `minimax` | `MiniMax-M3` | 3 | 3 | 原有服务回归 |
| `minimax-cn` | `MiniMax-M3` | 3 | 3 | 新服务独立接入 |

新增 6 条路线、9 个型号身份；本功能合计 12 条路线、18 个型号身份。原有服务按证据影响确定重验范围，新服务每个完整身份必须有自身完整证明。缺资格或失败记录为 NOT_RUN / FAIL，保留目标与未完成验收项，不改为 UNSUPPORTED 来缩减合同。

### R03 独立 Model 元数据

Model 身份维持四维。允许两个 Provider 使用相同 ModelID 并在同一 Client 的 Catalog 中并存；每个完整身份的能力、思考映射、默认值、采样、Compat、容量与 Cost 均独立。Lookup / ModelsOf 不混入另一 Provider 或协议，返回独立副本。Host 自定义目录支持各市场分别配置，但不能通过 Model 扩大 Provider 协议能力。

ModelID 保留实际 wire 名称，不追加地区后缀或翻译别名。AllowedModels 只授权本 Binding 的 Operation × Provider × API 下的型号，不复制元数据，不让另一市场的同名型号补足目录缺项。

### R04 目录校验与快照

构造和 Hash 继续拒绝重复完整身份、非法模态 / 型号约束 / 采样以及负数或非有限价格。沿用现有深复制与快照，不共享可变嵌套数据。目录或价格内容变化更换 Version 和 Hash；同一 Logical Call 固定其所依据的快照。各市场可独立维护数据，Client 仍持有整个 Catalog 的一个固定版本。

### R05 协议复用与能力隔离

四个服务在三套内部能力配置中都必须有明确选择。现有配置增加协议私有的 codec / 实现选择，复用经过核验的共同形状；差异按协议明确表达。配置不可由 Host Catalog 或请求改写，不保存租户、账户、凭据或调用级可变状态。

逐项分类并迁移现有 Provider 判断：协议形状分支使用相应配置；授权、目录、原生来源、价格和归属仍比较真实 ProviderID。不得在输入、消息、封套或结果中把 `zai` 改写成 `zhipu`，或把 `minimax-cn` 改写成 `minimax`。

### R06 请求与流协议合同

各新身份按自身资料与实际捕获确定 thinking 默认与可用档位、预算字段、工具选择、reasoning 形状、流终态、用量与错误解释。不能直接继承原厂 OpenAI / Anthropic 默认字段，也不能用一个协议的成功推导另两个协议。

必须覆盖文本、工具往返、分片工具参数、原生 reasoning 续接、usage、失败终态和取消。full 不支持的具名选项明确拒绝，simple 保留现有按模型能力解析规则；无选项时保留该条目已声明和验证的默认。工具由 Host 执行；Responses 使用完整历史，不新增厂商保存会话或资源引用能力。现有 Provider 行为除 R11 派生键和 R14 新增准入元数据外保持原合同。

### R07 Binding 与端点

每份 Binding 固定一个租户内的 Operation、Provider、API、endpoint、账户和凭据引用。标准 API base 与 Host 校验规则以[设计第 6 节](design.md#6-端点与可信-host-配置)为准。官方装配入口拒绝另一市场主机、错误协议路径或带凭据的 URL；实际生成地址、版本头、SDK 查询参数和认证须由捕获确认。

可信 Host 的网关、Transport 和既有 loopback 测试能力保留，非官方目标必须显式配置；核心不自动按域名推导 Provider。端点选择与网络代理配置分离，不按地理标签决定传输或数据驻留。新增服务不扩大其他 Model Operation 的授权。

### R08 账户与凭据用途

Host 凭据记录保存 Provider、租户、账户和用途，Resolver 校对后才读取秘密。不同服务不能引用同一租户内的同一凭据记录；同服务、同账户、同用途且获准的协议可共享 CredentialRef。公共 Credential 字段形状不变，核心继续核对引用、账户、所有者、版本和 BindingVersion。

标准 API 与 Coding Plan / M Plan 资格分别声明，不能以 endpoint、品牌、key 文本或已有另一市场资格推断。认证失败、余额不足、限流和网络故障不得触发跨站、换 key、套餐转余额或换型号。

### R09 授权和失败阶段

保持 scope → binding → operation / model / options → credential → snapshot → admission → send 的既有边界。先完成型号目录与白名单交集；不匹配时零凭据读取、零 HTTP 请求、零 Attempt。凭据归属与快照不匹配不得发送请求，不能重新解析另一份配置救场。

| 失败场景 | 预期处理 |
| --- | --- |
| 缺少或串租户 Binding | 延用 `binding_not_found / binding`，不泄露另一租户记录 |
| Binding 不允许型号或操作 | `tenant_denied / capability`，不读凭据 |
| 当前 Provider × API 目录缺型号 | `invalid_request / capability`，不从兄弟目录查找 |
| Host 凭据记录 Provider / 用途不匹配 | Resolver 返回 ErrAccessDenied，对应 `tenant_denied / credential`，零发送 |
| Credential 账户或配置版本冲突 | `credential_unavailable / consistency`，不重新取 key |
| 官方预设的端点与服务不匹配 | Host 装配返回配置错误，尚未发起 Logical Call |
| 跨 Provider 的原生历史 | 依 R10 降级，合法调用继续，不把它变成配置错误 |

### R10 原生历史

Native State 的可信回放继续精确核对 Tenant、AccountScope、Provider、API、Model。相同账户别名与 ModelID、相同实现选择仍不能跨服务原生续接。可见正文、推理文本和工具配对按现有跨模型政策保留；不透明状态按政策去除 / 转换，并记录降级。

持久化数据须 Host 校验后 TrustNativeState；签名缺省但 native kind 需授权的内容同样检查。已有封套不重写、不增加 Site 字段。同服务同账户换凭据版本继续有效，不把 key 轮换误认为换市场。

### R11 缓存与亲和作用域

统一采用设计第 8 节的精确 v2 算法：SHA-256 输入为版本域及长度前缀编码的 TenantID、ProviderID、AccountScopeID、SessionID，输出 64 位小写十六进制。相同租户 / 账户别名 / SessionID、不同 Provider 必须得到不同派生键；分隔符和 Unicode 不造成元组编码碰撞。

同 Provider、租户、账户与 session 的结果稳定，换 CredentialRef / Version 不改变键。原有“不发送缓存字段”的条件继续生效；只在原协议能力允许时使用键或亲和头。切换覆盖所有现有 cacheScope 消费者，不保留 v1 回退 / 双发；明确记录既有 wire 中键值变化及一次命中率影响，不扩大为其他行为豁免。

### R12 Hooks 与最终请求

保留现有 Hooks 顺序与授权。模型身份、认证、账户专有引用、托管工具授权与派生缓存键在最终请求重新校验；合法回调仍可修复或设置已接入的参数，不能开启本服务未接入的能力。最终认证不得同时出现另一套 Bearer / x-api-key。

回调、采样与协议转换不改变真实 Provider、ModelID 或调用归属。非法请求发送前失败，不创建 Attempt；错误阶段和诊断保留现有语义，不把可信 Host hook 的失败伪装为厂商返回。

### R13 并发与重试快照

同一 Client 并发调用四服务、两个租户时，必须分别固定 Binding、Credential、Catalog、请求体、协议配置与归属。只按现有显式有界 Retry Policy 重试初始请求，重试复用原 endpoint、key 版本和冻结 bytes；默认零重试，流开始后不重放。

修改 Host 后续配置、禁用 Binding 或轮换 key 不改变已固定调用；新调用重新解析并接受撤销。不使用全局当前 Provider / 当前账户，不因同名 BindingID 或 CredentialID 混淆租户。

### R14 生命周期与资源

沿用现有调用 / 尝试生命周期、准入、请求与事件预算、超时、取消、背压和资源释放。四服务均覆盖队列受限、畸形或截断流、取消与重试终止；Result 只结算一次，关闭响应体并归还许可。共享 adapter 不能成为绕过策略的入口。

`AdmissionRequest` 增加 `ProviderID ProviderID`，每次 Attempt 使用已固定 Binding 的真实 Provider。Host 账户级准入以服务与账户区分不同市场；同一真实服务账户是否跨租户共享限额仍由 Host 明确决定，库内租户 / 进程限额不变。新字段不接受请求或 Observer 事件反向填充。

Admit 方法签名不变；Host 的位置式 struct literal、固定字段集合断言和账户聚合键须一并迁移，不宣称新增字段对所有源码形式均兼容。

### R15 用量与价格

按实际 Provider × API × Model 的固定目录价格估算。人民币费率经固定来源与日期的汇率归一到 USD；价格与档位不跨市场复制，不在调用中联网换汇。保存官方原件、Hash、币种、单位、档位条件和换算依据。

保持缺少 usage、部分 usage、明确零值、缺价和实际扣费的区别；已有缓存写入缺价保护不移除。合成价格仅用于离线 E2E，不表示真实费率；标准 API 估算不能展示成套餐余额或最终扣费。

### R16 候选与正式目录

新增 9 个型号身份先进入各自候选目录，约束和价格取自该服务自身来源。每条身份的候选验收与正式目录晋升单独核对，未通过不能写入 BuiltinCatalog；正式支持报告绑定晋升后的目录 Hash。ModelID 的新旧别名不自动互换，也不将另一型号成功用于补位。

### R17 可观测性与工件

CallAttribution、EventEnvelope、Result、Hooks 和 Observer 继续保存实际 Provider / API / Model / AccountScope。解析前失败保持 Resolved=false；不能回显未经解析的型号身份。默认日志与 Observer 只有获准元数据，不泄露 key、正文、签名、秘密引用内容或错误原文。

普通 E2E 使用合成数据和脱敏凭据别名，保留最终请求、脚本 / 原始流、事件、Result、Attempt、usage、降级统计、断言、资源归零、来源与文件 Hash。实际捕获按现有规则脱敏、审计并保留原件，不用重导出文件刷新采集时间。

### R18 独立支持与发布证明

支持矩阵新增 R02 的 6 条路线，逐型号核对自身完整能力，不合并、重命名或覆盖原服务条目。报告的实际 Provider、endpoint、账户别名、协议、型号、SDK、执行来源和目录必须相符；另一市场、另一型号、离线 replay 或过期来源不能认证当前支持。

冻结 pi 基线不更新。按现有基线实际内容分别登记可适用差分或扩展，不能从最新上游或共享 adapter 推导差分 PASS。追加新身份要求，保留所有旧门禁、效果合同及失败原件。共享 codec、缓存算法和目录 Hash 的影响均参与旧证明适用性审查，不能仅改 pins、时间戳或 Hash 复用旧证据。

### R19 迁移与回滚

保留现有 Provider 值、Binding 与历史含义；新市场使用新 Binding、独立凭据记录及明确账户授权。禁用一站影响该站新调用，其他站正常；已固定调用按原快照结束。

软件回滚先禁用新增服务的 Binding，并从旧进程将加载的 Catalog 移除新身份，再回滚代码与目录版本；旧 Host 不加载新身份的自定义目录。保留新增服务的历史及消费账本，不能依赖旧版本的未知 Provider 默认分支实现拒绝，也不自动映射到兄弟 Provider。内部迁移删除过时实现，不添加兼容 shim、别名、双写或隐式 fallback；缓存 v2 升降级的命中率变化按设计说明。

### R20 文档与验证流程

实现前先在公共 Client / Host 配置边界建立下表失败场景，再写生产代码；不在实现后补 helper 单元测试。必须隔离验证时先写失败方式，遵循 [verification.md](../../docs/agents/verification.md)。每项需求有实际场景与保留工件，不用需求映射替代执行结果。

普通阶段按变更影响选择 E2E 与旧服务回归。发布按现有阶段执行前置、双 vet、完整普通集合、差分、压力和原生证据消费；race 只有人类明确选择该范围时运行。新增付费验证需另有身份、资格、范围与预算授权，设计 / 实施批准和旧剩余额度不构成新授权。当前状态、实际 argv、退出码、时间、源码、输入与原件统一记录在本节指向的状态入口。

## Testing Decisions

### 公共边界验收矩阵

下列 RP01–RP20 为待实现场景编号，普通功能场景通过公共 Client 或 Host 装配入口执行，不直接测试私有 profile helper。涉及全部目标身份时使用 R02 的明确集合；无法在受控服务中证明的真实资格单独验收。

| 场景 | 前置或失败方式 | 必须验证的可观察结果 | 需求 |
| --- | --- | --- | --- |
| RP01 | 四个 Provider，同品牌同名 wire 型号 | 同一 Client 按真实 Provider / API 找到独立条目；wire 型号无地区后缀 | R01 R02 R03 |
| RP02 | 两站对同名合成型号设置不同预算、档位、采样和价格 | 请求默认与估算各自正确；只改一站不改变另一站；错误 full 档位不发送 | R02 R03 |
| RP03 | 重复完整身份、跨站目录缺项、非法元数据 | 重复构造失败；不从另一站补目录；失败在正确阶段 | R03 R04 R09 |
| RP04 | 修改 Catalog 及查询返回的嵌套 map / slice，跨市场价格档位与固定汇率边界 | 活跃 Client / 调用不变；新内容有新版本与 Hash；币种换算不混用 | R03 R04 R15 |
| RP05 | 共用 adapter 的已知服务具有不同能力，Host Compat 或 Hook 试图扩大能力 | 各自只使用明确 profile；新地区不落原厂默认；真实 Provider 从头到尾不变 | R05 R06 |
| RP06 | 迁移前后旧服务同一脚本，核对三套 codec 全路径与冻结差分 | 除精确声明的 v2 键差异外旧行为一致；差分 / 扩展如实登记 | R05 R06 R18 |
| RP07 | 新服务各协议文本、工具分片、native history、迟到 usage、非法终态与 EOF | 内容不丢失 / 重复；工具往返完整；失败和 usage 保留上下文 | R02 R06 R14 |
| RP08 | 官方 Host 预设给错市场或协议端点，另有显式网关 / loopback 配置 | 官方配置失败且零请求；显式可信目标仍能按原服务身份工作 | R07 R09 |
| RP09 | 账户名相同但凭据来源 Provider / 用途错误，或配置版本冲突 | Host 和核心分别按 R09 拒绝；不换凭据、不发送 | R08 R09 |
| RP10 | 同租户另一站的 AllowedModels、另一操作 Binding、另一租户同名 Binding | 白名单不能互授；操作 / 租户隔离；不应读凭据时读取次数为零 | R03 R07 R09 |
| RP11 | 同名型号、同账户别名的跨 Provider 受信历史；未信任持久化历史；同账户轮换 key | 跨站签名等状态降级，正文与工具配对完整；同源合法回放保留 | R10 |
| RP12 | Tenant / Provider / Account / Session 元组边界、分隔符、Unicode、换 key、retention=none | v2 字节算法与稳定性正确；不同 Provider 键不同；省略条件不变；无 v1 双发 | R11 |
| RP13 | Hooks 改 model、认证、缓存键或账户引用，及合法的参数修复 | 非法最终请求零 Attempt；合法回调顺序不变且只执行一次 | R05 R12 |
| RP14 | 初始可重试错误期间修改配置，随后发生已开始流的错误 | 有界初始重试固定原目标、凭据、请求 bytes；流错误不重放、不换站 | R08 R13 |
| RP15 | 两租户复用配置名，四服务账户别名 / SessionID 同名；受控账户准入、撤销与轮换 | 同租户各站 Binding / CredentialRef 独立；Admission / 事件 / Attempt / Result 均带真实服务；一站受限不串到另一站 | R08 R13 R14 R17 |
| RP16 | 取消、deadline、满队列、超预算、观察器失败、截断流 | 一次结算，body / permit / waiter 等归零，脱敏记录可核验 | R14 R17 |
| RP17 | 未报 usage、部分 usage、明确零值、缓存写缺价、协议错误但有已知用量 | 按既有合同报告与估算；不伪装免费或最终套餐扣费 | R06 R15 R17 |
| RP18 | 另一站 / 型号报告、错 endpoint / Hash / source、缺少新身份或只有候选证明 | 证据消费者拒绝；不可提前晋升，旧矩阵记录和真实失败原件保留 | R02 R16 R18 |
| RP19 | 禁用一站后调用另一站，回滚后读取新身份历史 | 禁用与新调用规则正确；历史不改写；未知目标明确失败而非重路由 | R01 R13 R19 |
| RP20 | 新路线完整自身 live、旧来源适用性审查、普通门禁与文档工件完整性 | 新身份逐项认证；缺任一必交证据保持未完成；旧要求不弱化 | R02 R16 R18 R20 |

RP02 / RP04 的合成限额与费率验证配置机制，不宣称厂商实际支持这些合成值。RP05 选择已有实际能力差异的服务（如 OpenAI / DeepSeek Responses）并覆盖新地区的明确配置，不注入可写能力表，也不为测试虚构两地区的能力差异。厂商能力仍需自身来源。

### 工件与重放

复用 [evidence testkit](../../ai/internal/testkit/evidence/evidence.go)、既有受控 Provider、审计和来源工具。每次执行使用新的输出目录，保留 manifest、场景断言、请求 / 流 / 结果及各文件 Hash；失败轮次同样保留。成功的充分条件包含实际执行的场景清单、数字退出码和工件完整性，不能只接受“没有匹配测试”的退出码 0。

实施时新增公共测试统一以前缀 `TestRegionalProvider` 登记。以下为完成测试实现后的定向普通命令，不是本次已执行结果：

```sh
BARNESS_AI_EVIDENCE_DIR=/absolute/path/to/new-regional-evidence \
go test ./ai/e2e -run '^TestRegionalProvider' -count=1
```

回放读取保留的合成脚本或本身份自己的脱敏捕获，通过同一 Client 比较 wire、事件、归属、usage、降级与资源工件，不发真实网络请求。RP20 的新 live 要求由独立获准合同执行，不能混入上述普通命令或以回放通过代替。

### 完成条件

实现交付要求 R01–R20 的生产与 Host 装配变更、普通功能场景、旧服务回归、文档与可重放原件齐全。正式功能验收还要求新增 9 个身份的自身完整支持、晋升后的目录与来源证明、所有追加发布检查以及原有发布合同通过。

缺 live 授权、race 选择或其他必需证明时，可以交付实现和已有普通结果，但本功能完整验收继续开放，`race: NOT_RUN`、`release.pass: false` 按实际状态记录。不能把未执行、真实失败或来源过期合并成同一状态。

## Out of Scope

- 新增 Site / Region 字段、Provider 品牌层、通用插件注册中心或新的 API 协议。
- 自动跨站 fallback、自动挑 key / 型号、语言 / IP 地区推断、账户合并、额度迁移或余额查询。
- 将所有套餐拆为 Provider，或自动把新服务接入资格推广为 Coding Plan / M Plan 资格。
- `MiniMax-M3.1-Flash` 的新增与 Preview 迁移、视觉 / 视频接入，以及依据最新文档自动扩张现有型号能力。
- 新建前端产品、改变 Actor / Tenant 鉴权体系、替换秘密管理产品或升级冻结 pi / SDK / 依赖版本。
- 重写历史捕获、原生封套、旧失败、预算账本和已经关闭的付费授权。

## Further Notes

本规格扩展 [ADR-0042](../../docs/adr/0042-barness-ai-minimax-international-target.md) 的目标服务范围，保留其账户、证据与消费隔离要求。型号身份遵循 [ADR-0020](../../docs/adr/0020-barness-ai-model-operations-and-binding-authorization.md)，协议能力边界遵循 [ADR-0032](../../docs/adr/0032-barness-ai-provider-automatic-fields.md)，执行授权与发布阶段遵循 [verification.md](../../docs/agents/verification.md)。

2026-10-20 官方来源与上游参考保留在 [research](research.md)。其动态价格只用于说明独立数据的必要性，实施须冻结实际使用的各站来源；不把另一个市场或另一个日期的资料当作本身份支持证明。
