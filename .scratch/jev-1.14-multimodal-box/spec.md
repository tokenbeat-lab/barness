# barness-ai 统一分类图片引用与 Jev 1.14 Box

Status: ready-for-agent
Type: feature
Updated: 2026-10-16

本规格按 to-spec 发布到项目本地 issue tracker。当前仅完成研究、设计和规格；本轮不实施代码，也不运行真实推理。实现状态与 triage 状态不同，ready-for-agent 不表示已实现或支持矩阵通过。

配套：[设计](design.md)、[研究](research.md)、[拟议 ADR](../../docs/adr/0030-barness-ai-unified-classifier-image-references.md)。

## Problem Statement

Host 已能通过 Classify 调用 Jev 文本分类及 Decisions 图片分类，现在需要使用 Jev 1.14 的可组合图片与 Box。若新增公共 Assets 并要求宿主掌握 Jev 的字面字符串引用，图片组织会随 Provider 分裂，协议中的 substring 匹配、multipart 和原生回显也会进入宿主代码。

Host 需要一种属于 barness-ai 的图片与引用定义，能够表达共享上下文、状态值、题目说明和判定标准中的图片；同一张图片可被多处使用，但可见范围、顺序、资源和权限必须明确。协议无法表达某种位置时，应在发送前得到可诊断的拒绝，避免误把局部标准提升为全局信息。

新型 Box 还要求 Host 区分位置、存在性和置信度，并在多图、Hook 换图和图例回显后仍能解释结果。协议类型校验、定位业务效果与正式支持不能混为一个结论。

## Solution

在现有 Classify Interface 上保留 Images 为唯一图片来源，增加显式 ImageReferences；每条引用选择一个来源及 context 或 value 位置。context 表示共享的有序图片，value 使用 JSON Pointer 指向明确的 JSON null 槽位。普通字符串和对象永不自动变成公共图片引用。

两个 Adapter 分别投影：Jev 生成私有素材 key 与 assets，并保持各位置；Decisions 将 context 引用映射为有序 image content，拒绝局部位置。复用既有 Call Scope、Binding、Credential Snapshot、Admission、Retry Policy、Usage、Observer 和原子结果。

新增 BoxQuestion / BoxAnswer，以及不含图片 bytes 的最终图片元数据和多模态 Score 图例引用。目录能力、Host 预算和 Binding 图片授权分别负责可表达性、资源与访问许可。先形成公开入口 E2E 的失败清单和证据，再实施；真实服务验证、目录发布及业务效果另有明确门槛。

## User Stories

1. As a Host developer, I want one Image definition for both classifiers, so that I can change Target without rebuilding image bytes.
2. As a Host developer, I want one image source collection, so that I do not maintain separate Images and Assets inputs.
3. As a Host developer, I want existing ordered image calls to retain their meaning, so that adopting references does not require a migration.
4. As a Host developer, I want explicit context ordering, so that I can compare images in the intended sequence.
5. As a Host developer, I want one image referenced several times, so that I can express repeated uses without duplicating source data.
6. As a Host developer, I want the same bytes in two source entries to remain distinct entries, so that input identity is not changed silently.
7. As a Host developer, I want images inside State values, so that image context can retain business structure.
8. As a Host developer, I want an image inside one question's Instructions, so that unrelated questions do not receive that example.
9. As a Host developer, I want image examples in Choice descriptions, so that category definitions can be visual.
10. As a Host developer, I want image examples in Score levels, so that a visual rubric preserves its level order.
11. As a Host developer, I want image examples in Bool criteria, so that both interpretations can be explicit.
12. As a Host developer, I want image references resolved only at declared positions, so that user text cannot accidentally trigger replacement.
13. As a Host developer, I want ordinary null values preserved, so that visual inputs do not change text-only business data.
14. As a Host developer, I want escaped JSON keys and array positions handled precisely, so that unusual question and criterion names remain usable.
15. As a Host developer, I want dangling and duplicate references rejected, so that a mistaken image placement never reaches a model.
16. As a Host developer, I want unused sources rejected in an explicit declaration, so that unnecessary image data is not sent.
17. As a Host developer, I want unsupported positions rejected before credentials, so that a Provider change cannot silently alter scope.
18. As a Host developer, I want Jev private asset keys hidden, so that protocol key generation remains an Adapter concern.
19. As a Host developer, I want context wrapping documented, so that I understand the native structure without guessing rewritten natural-language paths.
20. As a Host developer, I want Decisions images emitted in reference order, so that explicit reuse and ordering survive protocol conversion.
21. As a Host developer, I want image MIME support checked per model, so that WebP support on one route does not imply support on another.
22. As a Host developer, I want no automatic resizing or rotation, so that coordinates correspond to the pixels I submitted.
23. As a tenant administrator, I want picture permission on Binding, so that model catalog updates cannot grant image access.
24. As a tenant administrator, I want chat and image-generation permissions isolated from classification, so that an unrelated permission cannot authorize this operation.
25. As a Host operator, I want image experiments explicitly selected, so that a request cannot enable them through a hidden header change.
26. As a Host operator, I want finite source-image budgets, so that known image sizes can be refused before copying.
27. As a Host operator, I want finite reference-expansion budgets, so that reusing one image cannot evade resource limits.
28. As a Host operator, I want final encoded-body limits, so that Base64 and protocol framing cannot evade request limits.
29. As a Host operator, I want inherited finite limits for existing configurations, so that references do not create an unlimited default.
30. As a Host developer, I want legal Hook edits validated against final native content, so that answers correspond to what was actually submitted.
31. As a tenant administrator, I want Hook-added images checked against the same permission, so that callbacks cannot widen access.
32. As a Host operator, I want retries to reuse frozen request bytes and credentials, so that one Logical Call has one stable input snapshot.
33. As a Host operator, I want successful HTTP responses never replayed for decoding failures, so that a bad answer does not create another paid call.
34. As a Host developer, I want one Box mixed with other question types, so that location and existence can be evaluated together.
35. As a Host developer, I want normalized Box edges preserved, so that I can apply my own display and crop transformations.
36. As a Host developer, I want null Box confidence distinguished from zero, so that missing confidence is not interpreted as a reported value.
37. As a Host developer, I want a separate existence answer, so that an always-present Box is not treated as proof of an object.
38. As a Host developer, I want multi-image Box attribution described honestly, so that I do not rely on an invented image ID.
39. As a Host developer, I want visual Score legends returned with public references, so that I can interpret rubrics without vendor placeholders.
40. As a Host developer, I want final input MIME and content hashes, so that I can identify images after trusted Hook changes.
41. As a Host developer, I want malformed answers rejected atomically, so that I never consume a partially validated result.
42. As a Host operator, I want valid reported usage preserved on answer failure, so that consumption remains visible.
43. As a tenant administrator, I want concurrent calls isolated even with equal indexes and names, so that no image or result crosses tenants.
44. As a Host operator, I want Observer records limited to metadata, so that image contents and derived private details stay out of shared logs.
45. As a Host operator, I want cancellation and deadlines to release bodies and permits, so that failed image calls do not exhaust capacity.
46. As a maintainer, I want catalog capabilities deeply copied and hashed, so that mutable configuration cannot alter pinned calls.
47. As a maintainer, I want text and existing Decisions behavior covered by regression evidence, so that the extension does not narrow old workloads.
48. As a maintainer, I want repeatable E2E artifacts with synthetic images, so that behavior can be checked without production data.
49. As a maintainer, I want protocol support and localization quality evaluated separately, so that a typed answer is not presented as business accuracy.
50. As a Host operator, I want revocation to affect new calls while preserving pinned calls, so that rollback follows the existing snapshot contract.

## Implementation Decisions

### R01 — 单一公共契约

保留 ClassifierRequest 的 State、Questions、Images；新增 ImageReferences，不新增公共 Assets。复用 Image 的 Data/MimeType；分类入口不接受路径、URL、file ID 或 Provider 素材对象。ImageReference 包含必填 imageIndex、placement，以及仅 value 具有的 pointer。ImageIndex 是非负整数、必须落在 Images 范围；缺失不能默认成零。公共记录精确字段名、拒绝重复键、未知字段和跨 variant 字段。

### R02 — 规范化、顺序与使用完整性

省略 / null / 空 ImageReferences 将 Images 按原次序全部视为各一次 context；非空时是一份完整声明，每个来源至少被使用一次，不补隐式引用。重复同图的 context 使用合法且分别计数；value 目标唯一。相同 bytes 的不同来源不自动合并。两种书写只生成一份内部引用计划，两个 Adapter 共用这份领域语义，没有兼容 shim 或双写。

### R03 — 显式位置和有效内容校验

value 使用 RFC 6901 JSON 字符串形式，相对公共请求根；只允许 State、单题 Instructions、单个 Choice/Score/Bool criterion 的内容值及后代。目标存在且为 null；它与引用共同成为图片节点。允许被明确引用的 State / Instructions 根图片；未引用根仍按已有 JSON 格式和非空规则检查。其他 null 原义保留，字段容器、类型名、对象键和身份字段不可被引用。只读解析，不创建路径，不做字符串替换；句中图片用有序内容数组表达。

### R04 — Jev 映射

私有素材 key 由 Adapter 生成，经有界碰撞检查保证不误中原始文字；只写入目标槽位。一个来源对应一个 native asset，多次使用保留多个出现位置；native asset 采用 type=image、PNG/JPEG format 和 raw Base64。没有 context 时保持 State 内容位置；有 context 时将原 State 作为独立 state 子树，与 images 序列一起包裹，不向原业务对象注入保留字段。必须公开这种结构投影及其自然语言路径限制，不自动改写问题说明。

### R05 — Decisions 映射及能力拒绝

只接受 context；任一局部用途使整次调用在 capability 拒绝，不丢弃、不提升作用域。无图 input、字符串 / 精确 compact JSON 转换、唯一 user 消息、text-first / images-in-reference-order 保持。每引用输出一个 input_image data URL，重复引用重复 block。问题类型转换、字符串说明、两候选下限和公共 2–10 评分等级保持；Box 不转为聊天或 Choice。

### R06 — 模型能力与协议边界

ClassifierCapabilities 增加 ImageLocations，明确 context/state/instructions/criteria；增加 box kind。图片位置、MIME、数量受协议支持与目录声明交集限制。省略新位置列表时，旧 Decisions 图片型号解释为 context，text-only 解释为空；默认不写回目录、不改旧 Hash。未知、重复、矛盾能力在目录边界拒绝，查询 / 构造 / 快照深复制。新增 Jev 图片与 Box 型号须用明确版本，不能从型号名称猜能力或借 alias 扩大授权。

### R07 — 图片授权与实验选择

所有图片用途同时受 Classifier 图片策略、型号能力和 Binding.AllowClassifierImages 控制；图像生成策略或聊天 Binding 不授权本操作。TypeSafeOptions 增加默认 false 的 ExperimentalImages，图片与 Box 还需显式选择才发 Images 实验 header；该选择不能授予权限。Long Context 首期不开放。最终 header 大小写归一，重复 / 矛盾值及 Hook 越权拒绝。

### R08 — 来源与展开预算

保留单图、所有来源 decoded bytes、来源数量、State / Question bytes、完整请求 / 响应 bytes 限制。新增可选 References 子策略，显式时 MaxReferences / MaxExpandedImageBytes 均为正；省略时分别继承 MaxImages / MaxTotalImageBytes 的有限上限。引用预算按每次出现累加；共享 State 只算一次。Decisions native image block 上限按展开数取 128 与型号限额的更小值。禁止在复制 / 序列化后才检查已知超限，累加不能整数溢出。

### R09 — 格式与厂商资源限制

Jev 首期静态 PNG/JPEG，拒绝 APNG；decoded 单文件、边长、面积、最终 body 和 key 字节须遵守官方已知边界及更小 Host 限制。Decisions 保留已验证 PNG/JPEG/WebP 子集。流式验证 Base64、签名 / 格式和尺寸，不自动缩放、旋转、转码或解压完整栅格。Host 对预处理后的图像及坐标负责。本地 decoded 预算不冒充厂商重新编码后容量或精确图像 token 预测。

### R10 — 独立 native freeze 与 Hook

每协议独立拥有 native 请求和最终验证记录，不把 TypeSafe 原生模型回解成公共请求后跨层复用。Hook 后重建题目、图片出现、MIME、预算和答案依据；支持合法增删题 / 换图 / 重排。Jev 最终接受子集仅 JSON Base64、全值素材引用；substring / part / 外部资源偷渡拒绝。Decisions 最终沿用闭合 content 树。认证、模型、目标、权限与实验选择始终固定，合法 Hook 不能扩大它们。

### R11 — 结果来源和图例

ClassifierResult 新增 ImageInputs，包含最终图片 MIME 和 decoded 文件 SHA-256 元数据，不含 bytes 或 vendor key；只有全部答案合法后才发布。最终索引按设计中的确定遍历次序建立，不保证等于入口索引。ScoreAnswer 新增 LegendImageReferences；先校验 native 图例精确符合最终 Criteria，再还原图片为 null 槽位并附引用。引用索引指向最终 ImageInputs，pointer 相对该答案 Legend 根；仅允许 value。无图片的图例、分布与有界舍入保持。

### R12 — Box 与业务语义

新增封闭 BoxQuestion / BoxAnswer variant，保留非空说明。每个 Box 至少能看到 context、State 或自己 Instructions 中的一张图，不能借其他题的局部图满足前提；无可见图为 invalid_request/scope，Hook 后丢失可见图为 callback_failed/request。答案四边显式、有限、归一化且有正面积；confidence 显式存在，允许 null，与数字零区分。Box 不带 criteria、not-found、虚构图像 ID 或存在概率；不按置信度推断目标存在。Host 用独立 BoolQuestion 与自己的业务阈值决定是否消费 Box，同批问题没有条件执行。像素换算、搜索图归属与动作安全由 Host 处理。

### R13 — 调用一致性、并发和重试

保持租户 / Actor 可信作用域、Binding / Credential / Catalog 一致快照；来源、引用及 Options 深复制，无全局素材缓存。回调每 Logical Call 一次，初始请求重试复用最终 bytes / headers / 素材 key / 凭据；默认不重试，2xx 后解码失败不重放。取消、超时、body 关闭和 Admission 释放沿用现有运行时，不分拆问题或跨协议 fallback。

### R14 — 错误与原子结果

坏公共引用 / Base64 / 文件签名为 invalid_request/scope；Host 已知预算超限为 resource_limit/scope。不支持位置、格式、Box 或缺实验选择为 invalid_request/capability；Binding 图片拒绝为 tenant_denied/capability。Jev 专属 APNG / 最小尺寸等检查在 resolve 后、pin 前完成，非法为 invalid_request/capability，超过文件 / 尺寸 / 面积最大值为 resource_limit/capability，不施加到 Decisions。以上均不读凭据、不建 Attempt。Hook 损坏为 callback_failed/request，越权为 tenant_denied/request，最终容量超限为 resource_limit/request。坏答案 / 图例为 protocol/response，Answers 和 ImageInputs 整体不发布，唯一合法 reported usage 与 ResponseModel 按既有规则保留。无新引用旧请求的错误优先级保持。

### R15 — 用量与可观测性

使用服务端 reported input_tokens（含图片）按对应目录快照估价一次；输出 / 缓存 / 阶梯延续各协议规则，不套其他 Provider 的图像计价公式。缺失 / null usage 不伪造成已报告零。Observer 仅保留既有归属、状态与用量；图片、引用路径、私有 key、框、内容 hash 和错误正文不得进入共享观测。最终图片元数据属于返回当前租户的结果内容，不是观测标签或授权证明。

### R16 — 兼容、发布与撤销

保留现有公共字段与默认行为，新增字段为 additive；同步 Host 结果消费者。未实施的 Assets 草案直接移除，内部重构删除旧路径。新目录内容换版本和 Hash；历史 Jev / Decisions 证据不自动成为新增能力或新 Hash 的 PASS。冻结 Pi 不升级，新增语义逐项登记扩展。撤销图片授权、移除型号白名单或禁用 Binding 影响新调用，已 pin 调用照原快照结束；不撤回 usage，不转文本或其他 Provider。

### R17 — 文档边界与实施状态

设计、规格、领域术语和拟议 ADR 同步表达本方案。相关既有 ADR 的差异须明示，已有 proposed 状态不能因设计完成而改成 accepted。本轮只交付这些文档；代码、测试实现、目录、支持矩阵及收费 API 调用不属于本轮执行范围。

## Testing Decisions

主要 Seam 为既有 Client.Classify / HookedClient.Classify：同一个公开 Interface 的两个入口，接真实库运行时和受控 HTTP Provider。检查返回值、最终 wire、Attempt、凭据读取、Observer 和资源释放；不为私有 helper 开新业务 Seam，不测试函数拆分或 map 内部结构。用户已在 2026-10-16 的 to-spec 接缝核对中明确确认采用这个既有公共接缝。

测试先于实现：先列所有失败方式，再写公共 E2E 令其失败，之后实施对应切片；不在实现后补单元测试。先例为既有 Decisions 图片 / Hook / 字节预算 / 撤销 E2E、TypeSafe 混合答案 / 图例 / usage / 真实 wire replay 与两租户并发隔离。沿用其证据记录器、可控 Provider 和脱敏审计。

| 验收组 | 必须可观察的行为 | 覆盖要求 |
| --- | --- | --- |
| T01 默认/显式共享图片 | 默认与相同显式 context 得到相同语义；旧 Decisions wire 保持；无图回归 | R01–R02、R05、R16 |
| T02 顺序与复用 | [1,0,1] 保留；一个来源多出现；重复 bytes 的来源不静默合并 | R02、R04–R05、R08 |
| T03 公共位置 | State、题内说明、Choice/Score/Bool 标准和根图片正确投影；其他题不获题内图片 | R03–R04 |
| T04 Pointer 边界 | 转义 / 数字对象键正确；坏 escape、缺字段、负 / 越界 / 缺失索引、重复目标、禁区拒绝 | R01–R03、R14 |
| T05 文本保持 | 普通 null、类似 key 的文字、普通对象及精确数字不变；不自动识别文本引用 | R03–R04 |
| T06 协议能力 | Decisions 含任何 value 则整次零凭据零发送拒绝；不支持 Box / MIME 同理 | R05–R07、R14 |
| T07 Jev context 结构 | 包裹不覆盖用户 state/images 字段，复用不因题数重复；无 context 保持 State 位置 | R04 |
| T08 格式与容量 | 合法至少 280×280 PNG/JPEG；坏 Base64/MIME、APNG、256×256、单图/引用/展开/完整 body 超限拒绝 | R08–R09 |
| T09 安全默认 | 缺策略、图片权限、能力或实验选择均不能借另一条件放行；Long Context 与 header 重复拒绝 | R06–R07、R14 |
| T10 最终 Hook | 合法换图换题重排按最终输入验；加 URL/part/substring 或改权限/目标/认证拒绝 | R10、R14 |
| T11 图例来源 | private key 不泄漏；null 槽位与 LegendImageReferences 可解释，最终 hash 匹配 Hook 后图像 | R11、R15 |
| T12 Box | 无可见图及仅其他题有图拒绝；null 与 0 confidence 区分；缺失/越界/退化/倒置/非有限坐标拒绝；合法不存在目标仍需 Host existence 判定 | R12、R14 |
| T13 原子和用量 | 坏一题使全部 Answers / ImageInputs 清空；合法 usage 保留；refusal 与已有舍入不变 | R11–R15 |
| T14 生命周期 | 取消/时限释放资源；默认零重试；允许重试字节一致；2xx 后不重放 | R13 |
| T15 隔离与深复制 | 同名 / 同索引两租户并发不串图；构造后变更配置和借用返回值不改已 pin 结果 | R06、R13 |
| T16 撤销与回归 | 新调用受撤销限制；已 pin 快照完成；Jev 1.13 文本 / Decisions 图片 / Pi 既有投影保持 | R16 |

所有适用组覆盖两个公开入口，容量 / 隔离 / 回调组纳入 normal 和 race 证据。仅断言契约，不把随机 key 的固定值、内部调用顺序或偶然耗时当作外部保证；同一 Logical Call 重试 key/bytes 固定是必须断言的外部保证。

E2E 结束必须产出可复核并可重放的合成图片和 hash、公共输入、最终请求/响应、结果/错误、Attempt/Observer、资源及释放断言、运行配置、完整性清单和脱敏审计结论。不得只给 go test 通过行。记录真实服务测试为独立来源；未运行为 NOT_RUN，不能用 mock 取代 PASS。

Jev 固定 1.14 候选目录的真实验证需要有限请求/题目/图片/字节/时长/重试预算；Jev 与 Decisions 各自校验。协议类型 / 图片容量、人工标注定位 IoU、目标缺失误报与中文任务效果分别报告，不借旧版本效果或他协议 PASS。首期保持 JSON 上传；multipart 不成为离线闭环的前置条件。

## Out of Scope

- 本轮代码、测试实现、真实 API 调用及部署；本规格供后续实施使用。
- 远程图片抓取、文件路径 / file ID、上传服务、持久资产与跨租户图片缓存。
- 聊天 / 图像生成接口重塑，通用文档 AST 或任意模态注册系统。
- 音频、视频、动画、自动 resize / EXIF 旋转 / 转码 / 裁剪。
- 公共字符串内偏移引用、Jev substring 模板透传、任意未知 native schema 放行。
- Decisions 局部图片、Box 仿真、逐题拆请求、自动 Provider fallback。
- 自动模型别名升级、Long Context 实验、multipart 实现和精确图像 token 估计。
- Host 的业务存在阈值、交互页面、工具动作或安全决策；不内置通用生产准确率门槛。

## Further Notes

本规格是原研究初设的收敛版本：公共 Assets 不再保留；唯一来源及显式位置属于 barness-ai，原生上传和引用属于 Adapter。ADR 记录为 proposed，术语定义不构成运行时支持声明。

公开协议仍有文档 / SDK 漂移和未公开语义；研究记录了 confidence nullable、instructions optional、处理后容量与模型别名等证据。只支持明确子集并如实保留未知项。Jev context 包裹、多图 Box 语义和定位质量需要独立效果证据；如果投影需要改变，先更新本规格和自己的验证，不能悄悄改变输入作用域。

本次保留已存在的无关工作区改动。文档检查仅验证规格、示例和交叉引用自洽；任何运行时能力仍待实现与验证。
