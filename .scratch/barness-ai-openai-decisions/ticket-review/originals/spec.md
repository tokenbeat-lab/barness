# barness-ai 接入 OpenAI Decisions

Status: ready-for-agent
Type: feature
Updated: 2026-10-12

本文件是待实施的交付契约。当前生产代码仍仅支持 TypeSafe 的分类协议；研究中的离线 Pi probe 不构成 barness 已实现或真实服务通过的证据。

配套：[详细设计](design.md)、[验收矩阵](acceptance.md)、[实施计划](implementation-plan.md)、[研究和兼容审查](../pi-1.1.0-openai-decisions/research.md)、[拟议 ADR-0026](../../docs/adr/0026-barness-ai-openai-decisions-classification.md)。

## Problem Statement

Host 希望通过现有 Classify 操作使用 GPT-6 Luna 的低延迟分类、判断、评分和图片理解，同时继续使用已有 System One/Jev。现有模块已经具有隔离、准入、凭据固定、回调、用量和完整结果语义，但公共校验和执行分派还包含单协议假设。

两个协议的说明格式、候选数量、输入形状、答案分布、拒答和用量不同。仅增加型号或替换 endpoint，会造成合法 Jev 请求被收窄、图片丢失、回调绕过授权或拒答被误当作故障；同一型号也不能凭聊天 Binding 自动取得分类授权。

## Solution

在现有 Classify Module 的同一 Interface 上增加 OpenAI Decisions Adapter。Host 仍提交可信 Call Scope、Target、状态和命名问题；Protocol Model 在内部独立转换。新增内联图片与逐题拒答的公共表达，显式声明图片能力、Host 资源策略和 Binding 授权。

现有 Jev 调用、错误优先级、原生回调、评分舍入、用量和重试保持。公共 score 本期仍支持 2–10 个等级，该范围是 barness 的支持契约，不冒称 OpenAI 上限。每项与冻结 Pi 1.1.0 的差异独立验证和登记。

交付包含公共 E2E、图片隔离和资源证据、固定 Pi 差分、该组合自己的真实冒烟、正式目录/价格快照、独立中文效果报告和发布门禁；任何未执行项不能成为 PASS。

## User Stories

1. As a Host developer, I want to call Luna through Classify, so that I can reuse my model operation interface.
2. As a Host developer, I want Jev calls to retain their behavior, so that adopting another protocol does not break existing workloads.
3. As a Host developer, I want the same model ID separated by operation and protocol, so that chat and classification remain distinct.
4. As a tenant administrator, I want a dedicated classifier Binding, so that chat authorization cannot grant classification access.
5. As a tenant administrator, I want model allowlists enforced before credentials are read, so that denied calls do not access secrets.
6. As a Host developer, I want nil options to select protocol defaults, so that simple classification calls remain concise.
7. As a Host developer, I want wrong protocol options rejected, so that switching a target cannot silently reinterpret options.
8. As a Host developer, I want old Classifier Model declarations to remain usable, so that additive metadata does not invalidate my Jev catalog.
9. As a Host developer, I want input formats discoverable, so that I can determine which requests a model accepts.
10. As a Host developer, I want JSON state converted predictably, so that text evidence retains its meaning and numeric precision.
11. As a Host developer, I want Jev's structured instructions preserved, so that its existing native JSON features remain available.
12. As a Host developer, I want incompatible Luna instructions rejected before sending, so that unsupported formats are explicit.
13. As a Host developer, I want choice candidates represented by stable strings, so that categories remain consistent in my application.
14. As a Host developer, I want a single Jev candidate to remain legal, so that current requests continue working.
15. As a Host developer, I want Luna's two-candidate minimum enforced, so that invalid requests are caught locally.
16. As a Host developer, I want ordered score criteria preserved, so that score indices keep their established interpretation.
17. As a Host developer, I want the current public score range preserved, so that validation order and failures remain compatible.
18. As a Host developer, I want bool questions to return probabilities, so that business thresholds remain my responsibility.
19. As a Host developer, I want bool explanations translated without loss, so that both alternatives remain understandable.
20. As a Host developer, I want ordered inline images, so that I can classify text and visual evidence together.
21. As a tenant administrator, I want image classification disabled by default, so that a new capability does not expand existing permissions.
22. As a tenant administrator, I want explicit image authorization per Binding, so that tenants can have different image permissions.
23. As a Host operator, I want classification image budgets independent of image generation, so that one operation cannot enable another.
24. As a Host operator, I want image count and byte limits enforced before copying, so that large input cannot exhaust memory.
25. As a Host developer, I want MIME and base64 checked, so that malformed image evidence is rejected predictably.
26. As a Host developer, I want TypeSafe image requests rejected, so that images are never silently discarded.
27. As a Host developer, I want a typed refusal answer, so that a normal model refusal is distinguishable from a broken call.
28. As a Host developer, I want other answers retained alongside refusals, so that one refused question does not erase useful results.
29. As a Host developer, I want an all-refusal response represented as a completed result, so that I can choose the next business action.
30. As a Host developer, I want malformed results to fail atomically, so that unvalidated answers cannot reach my application.
31. As a Host developer, I want score distributions retained, so that I can inspect uncertainty rather than only an average.
32. As a Host developer, I want response names and candidate sets checked, so that duplicate or unrelated answers cannot overwrite results.
33. As a Host developer, I want probability ranges and score consistency validated, so that impossible values are caught.
34. As a Host developer, I want Jev's evidenced rounding behavior retained, so that valid existing responses remain accepted.
35. As a Host developer, I want native callbacks to identify their protocol, so that I can safely handle different payload shapes.
36. As a tenant administrator, I want final callback payloads reauthorized, so that hooks cannot widen models, credentials or image access.
37. As a Host developer, I want answers checked against final frozen questions, so that legitimate callback changes work correctly.
38. As a Host developer, I want my request and callback objects copied, so that later mutation cannot affect an ongoing call.
39. As a Host operator, I want retries to reuse one payload and credential snapshot, so that calls remain attributable.
40. As a Host operator, I want no retry after a successful HTTP response, so that refusals and parse errors do not duplicate charges.
41. As a Host operator, I want cancellation and limits to release resources, so that failed calls do not retain permits or bodies.
42. As a Host developer, I want known usage retained when answers fail, so that consumed resources remain visible.
43. As a Host developer, I want missing and zero usage distinguished, so that an unreported call is not described as free.
44. As a Host operator, I want Decisions-specific pricing, so that chat cache and output rates cannot distort estimates.
45. As a Host operator, I want the long-context price boundary tested, so that the full-request tier is applied correctly.
46. As a tenant administrator, I want concurrent Jev and Luna calls isolated, so that shared identifiers and accounts do not mix results.
47. As a Host operator, I want metadata-only observation, so that images, state and answers do not leak into operational records.
48. As a maintainer, I want public Classify E2E tests, so that caller-visible behavior is the verification surface.
49. As a maintainer, I want fixed Pi 1.1.0 differential evidence, so that parity has a reproducible source.
50. As a maintainer, I want scoped extension records, so that deliberate differences cannot hide unrelated failures.
51. As a maintainer, I want candidate live evidence before catalog inclusion, so that metadata does not imply an unimplemented route.
52. As a maintainer, I want independent support matrix results, so that chat or Jev passes cannot stand in for Decisions.
53. As a Host developer, I want route-specific Chinese evaluation, so that Jev confidence thresholds are not assumed valid for Luna.
54. As a maintainer, I want repeatable audited artifacts, so that failures and release conclusions can be independently verified.
55. As a Host operator, I want explicit rollback, so that the new route can be disabled without rerouting calls to another provider.
56. As a maintainer, I want architecture and migration documentation, so that future work understands the contract and its limits.

## Implementation Decisions

### R01 — 独立型号身份和协议

新增 APIOpenAIDecisions，值为 openai-decisions。首期路线为 OpenAI × Decisions × classifier，内置/真实验收型号为 gpt-6-luna；wire model 从已授权目录 ID 取得，允许离线宿主合成目录而不硬编码型号。Binding 的 Operation、Provider、API 和允许型号共同授权；同名聊天型号不参与该 lookup。网络使用现有 unary HTTP Module，不引入新 SDK、注册插件或另一套调用运行时。

### R02 — System One/Jev 不回归

旧方法签名、State 的 JSON 范围、说明结构、单候选、公共 score 2–10、bool 概率、严格原生字段、回调顺序、错误 Code/Phase、用量和评分舍入保持。没有图片的新请求与旧请求在 Jev wire 上相同。保持现有通用形状校验先于 Binding 解析；新增协议规则在 capability 阶段检查。score 超过 10 的跨协议扩展不属于本期。

### R03 — 公共请求和选项

ClassifierRequest 新增可省略的有序 Images；State 仍必填且为 JSON string/object/array，图片调用也提交 State。一份状态/图片/全部问题是一个 Logical Call，不分拆或内部并发凑答案。问题名非空，choice/score/bool 的公共形状不变。OpenAIDecisionsOptions 首期为空，nil 表示目标协议默认，错误协议选项拒绝；没有 temperature、reasoning、stream、timeout 或 authority 选项扩张。

### R04 — 格式与容量

公共 choice 仍为 1–255，score 为 2–10；模型能力可进一步收窄。Decisions choice 额外要求至少 2 个。Jev 允许现有字符串、对象、数组说明；Decisions 首期仅接受 JSON string，choice 的 null 说明表示省略。结构化说明不隐式 stringify，bool 的两侧说明也遵守该规则。

### R05 — 模型能力与安全默认

ClassifierModel 增加输入模态；ClassifierCapabilities 增加最少候选、说明格式、图片数量和图片 MIME 声明。新字段省略时是旧的 text-only、最低 1 个候选、现有 JSON 说明格式、不支持图片；既有 Jev Host 目录仍合法。新增协议能力声明不得超出其支持范围，但不收紧旧 MaxChoices/score 字段已有的构造接受范围；实际调用仍受公共范围与有效能力交集约束。目录和返回值深复制，变更参与版本和 Hash。

### R06 — 图片资源策略

ClassifierPolicy 增加独立 Images 子策略；nil 关闭图片分类，非 nil 必须给正的 MaxImages 和 MaxTotalImageBytes。ResourcePolicy.Image 仍只控制图像生成。按解码后的单图/总图字节、完整 JSON 字节和有效图片数量分别限制；首期仅 inline PNG/JPEG/WebP，拒绝 URL、路径、file ID、detail、mask 和自动 resize。

### R07 — Binding 图片授权

Binding 新增 AllowClassifierImages，默认 false。图片调用须同时满足 Binding 明确授权、模型输入能力及 Host 图片策略。回调添加图片同样检查；绑定快照中的权限贯穿全部 Attempt。没有图片的旧 Jev 调用不受新开关影响；对其他 Operation 开启此字段是无效配置。

### R08 — 请求转换

Decisions 的 State string 解码成原文，object/array 压缩为 JSON 文本而不经 float64。无图用 input string，有图用一个 user 消息，先状态文本再顺序图片。choice 映射字符串 value，score 保持有序 label，bool 映射 predicate 并采用固定 true/false 说明拼接格式。字符串 "true" 不改为 boolean。原生候选 boolean 等宽 API 表达不扩入公共 choice。

### R09 — native 回调和最终请求

复用 header → payload → response 顺序；payload hook 仍接收 native 形状。每协议拥有 guard 和冻结记录，Decisions 使用 input/questions array，TypeSafe 继续使用 state/questions map。回调后重新检查授权、形状、图片、大小及题目集合；响应依据最终快照校验。禁止改 model 或添加认证、目标、工具、外部资源与图片授权字段。

### R10 — 拒答是类型化结果

新增封闭 RefusalAnswer，JSON discriminator 为 refusal，不编造原因。合法单题/混合/全拒答均为 StopReasonStop、nil error；Answers 的完整名称集合包含拒答。宿主决定阈值、路由和人工处理。Jev 仍只接受其已定义的三种 native 答案，不将未知 Jev type reinterpret 为 refusal。

### R11 — 严格结果发布

任何缺题、多题、重复字段/名称/候选、类型错误、非法数值或分布导致整次失败，Answers 清空，已唯一上报的用量保留。Decisions score 保存概率及经验证的图例，等级从 0 开始；有限范围、概率和与期望检查不归一化、不修补。Jev 的有界舍入例外不共享给 Decisions。

### R12 — 用量和价格

各协议独立解析用量与 presence。Decisions canonical Input 保留完整 input_tokens，cache 桶为零，Reasoning 只在明确上报时设置，TotalTokens 保留合法上报或由已知项求和并标 partial。估价只用独立 Decisions 输入费率；基础 .1 USD/百万，严格超过 272000 输入为 .2 的全请求阶梯。不扣缓存、不重复加缓存、不套聊天费率；区域 premium 不宣称已包含。

### R13 — 调用生命周期

共享固定凭据/目录快照、默认零重试、Binding 的显式重试、准入、截止时间、取消、bounded body 与资源释放。任何 2xx 之后不重放，包括拒答、回调、解码或读取失败。非 2xx 错误保留有界脱敏诊断和 x-request-id。Pi 的默认重试及 504 不重试与当前运行时的差异逐项登记。

### R14 — 隔离和不可变状态

以可信 Call Scope 定租户/主体/作业；请求自报 Operation 不授予权限。请求、问题、图片、选项、目录和 final payload 各边界独立复制。共享账户不改变租户隔离；相同 BindingID/ModelID 的 Jev、Luna、聊天或图像操作不得混用授权、快照、答案或 Observer 归属。

### R15 — 固定 Pi 差分

保持当前唯一 Pi 1.1.0 oracle 和来源 pins，增加 Decisions classify/catalog 分派。比较完整共同请求和基本答案；图片、JSON 精度、拒答、score 详情、usage presence、严格校验、回调时机、header/故障和重试差异各有 case/path 有界账本，不豁免整个协议、body 或结果，不把扩展记为未经解释的 parity。

### R16 — 目录纳入

先用宿主候选目录验证真实路线，后纳入内置 gpt-6-luna 分类型号。contextWindow 使用固定 Pi classifier 快照 922000；初期 score 能力 2–10 是本期支持子集。正式候选 metadata、格式、价格和自己 live 证据齐备后换目录版本和 Hash；无日期 alias 的服务权重不能由目录 Hash 保证固定。

### R17 — 独立支持与发布

增加 Decisions classifier 的第十组合，矩阵只接受该路线自己完整、同版本的报告。保留九条现有路线的固定合同。候选通过后仍需正式目录 hash 下的报告；缺 key 为 NOT_RUN，真实错误为 FAIL，不伪造通过。最终完整门禁要求当前适用的全部离线、race、差分、压力、真实报告和脱敏审计。

### R18 — 宿主示例和效果

旧 Jev 示例默认和历史效果报告保留。新 Luna 配置具有独立 route、key、任务/目录哈希和报告；共享固定标注数据但不共享阈值或质量结论。正确率、混淆、拒答率、置信统计、延迟和消耗各自呈现；合法拒答计入完整性和任务分母，但不计为正确类别。

### R19 — 证据和观测

Observer 继续 metadata-only、有界、不会改变调用结果；不增加图片/答案/原始缓存明细/错误正文。E2E 与 live 输出版本、身份、fixture、捕获、消费结果、Attempt、Observer、assertions、审计、SHA-256 manifest 和重放命令。artifact 属于已声明的测试数据范围，真实用户数据不进入测试包。

### R20 — 文档和回滚

更新公共契约、包说明、差异账本、架构决策和发布追溯。保留 System One 历史 ADR，使用 Decisions ADR 明确新增语义。回滚禁用新 Binding、撤销 classifier 型号或图片授权；不自动改走 Jev/聊天，已发请求用量仍保留。内部过时的单协议分派直接替换，不新增 shim。

## Testing Decisions

- 用户于 2026-10-12 确认：Client.Classify / HookedClient.Classify 为主要执行 Seam；目录查询与发布 artifact 验证是辅助 Interface 检查。
- 好的测试观察调用者可见的请求、结果、Code/Phase、快照、调用次数、许可和 body 释放。禁止以私有 parser/helper 测试或镜像实现的 unit test 替代公共 E2E。
- 每个实施切片先列全失败场景，先写红色 E2E，再生产实现；基线已通过的 Jev case 继续运行且断言不收窄。
- Host 和 Provider 替身沿用现有装配与受控 HTTP Seam；网络错误、读取取消、回调修改、重复 JSON 和图片都从公共入口注入。
- 先例包含现有 Jev 严格 envelope/混合问题/JSON 精度/舍入、unary 生命周期与回调、mixed-operation 隔离、图像预算、固定 Pi 差分、真实矩阵和独立中文评估。
- 新路线单独证明，Jev regression、catalog clone/default、拒答/score、图片边界、回调授权、用量、并发/race 和 artifact 审计均为必交；详见验收矩阵。
- 性能与质量分开报告；不把公共接口通过、Pi 差分或小样本正确率等同于生产 SLA。

## Out of Scope

- 实施生产代码、运行付费 API 或宣称真实支持：本次请求只交付设计与实施计划，实际执行属于工单。
- 任意新 Provider、动态 Adapter registry、聊天/SSE 伪装、Agent loop、工具执行、结果自动 fallback。
- 原生 boolean choice、多 user message、URL/file ID 图片、图片 detail/resize、图像生成、音频、结构化输出抽取。
- 公共 score 超过 10 档、将 Jev 宽 JSON 说明自动文本化、增加业务阈值或自动判断 yes/no。
- 升级整个 OpenAI SDK、模型权重固定承诺、tokenizer、本地 context token 精确预测、重复记缓存账单或实现区域计费系统。
- 吞错、无界重试、共享另一组合 PASS、事后 unit test、未经执行的发布 PASS。

## Further Notes

事实依据为[固定研究](../pi-1.1.0-openai-decisions/research.md)及其中第一方源码/官方文档。有关[模型操作与授权](../../docs/adr/0020-barness-ai-model-operations-and-binding-authorization.md)、[Jev unary](../../docs/adr/0021-barness-ai-typesafe-unary-classification.md)、[目录标准](../../docs/adr/0018-barness-ai-catalog-inclusion-criterion.md)、[用量](../../docs/adr/0010-barness-ai-usage-reporting-and-price-snapshot.md)和[中文效果](../../docs/adr/0024-barness-ai-chinese-classification-evaluation.md)的既有决定继续适用。

本设计对“移出公共 score 上限”的研究建议作出取舍：本期维持公共 2–10 及原有 scope 拒绝顺序，不为没有真实需求的宽范围改动旧契约；模型/协议检查仍独立。不支持 11 档的原因是 barness 本期契约，而不是 OpenAI 服务被证实只支持 10 档。

Decisions 仍是 beta；score 服务上限、缓存账单、区域 premium 和大输入 504 适用性没有本路线真实证据。计划用有限且可审计的自己的探针核对，未知项不变成无限能力，也不阻止先交付明确的支持子集。

所有工单使用 ready-for-agent；Dependencies 描述实施顺序，外部凭据缺失按实际 NOT_RUN 记录。spec 的就绪标签不表示工单已完成或本次已经部署。
