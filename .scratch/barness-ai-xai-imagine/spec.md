# barness-ai xAI Imagine 图像生成与编辑

Status: ready-for-human
Type: feature
Updated: 2026-10-29

本文是本功能的交付契约，综合用户已确认的 Q1–Q7。领域用语遵循根词汇表；架构决定遵循 ADR-0020、ADR-0022、ADR-0054 和已接受的 ADR-0056。详细依据见 [设计方案](design.md) 与 [官方研究](research.md)。候选代码、真实候选验证和受限内置目录激活完成；正式原生资格和后续人工阶段仍未完成。

## Problem Statement

Host 希望通过 barness-ai 现有 image 路径调用 xAI Imagine，完成文生图、单图编辑和有序多图编辑，并继续使用已有的租户授权、凭据解析、资源策略、取消、回调、观测和结果处理方式。

当前图片操作只接入 OpenAI 和 Google，xAI 的 Binding 与图像目录校验明确拒绝图片操作。仅替换 endpoint 无法完成接入：xAI 编辑要求不同的 JSON 资源对象，质量和分辨率共同决定按张价格，现有 ImagePricing 仅表达 token 费率。xAI 还可能只报告收费而不报告 token，若沿用现有假设，会发生请求不兼容、缺失费用被当作零费用，或厂商报告费用与目录估算混淆。

Host 需要一个边界明确、可验证的原生图像操作，而不是新的客户端、聊天工具包装或异步作业系统。

## Solution

在既有 Client.GenerateImages 和 HookedClient.GenerateImages 中增加 image × xai × xai-images 路线，首期只接全球标准 API 的 grok-imagine-image-2.0。沿用 ImagesRequest 的提示词和有序内联参考图，同步返回 ImagesResult；复用既有 unary 调用生命周期。

文生图最多输出 10 张；编辑最多接收 5 张参考图且输出 1 张。候选分辨率为 1k、1.5k、2k；质量为 low、medium、auto。所有上限受型号目录和 Host 资源策略进一步收紧。省略选项沿厂商默认，固定示例和验收显式指定质量、分辨率。单图编辑暂拒绝显式比例，多图编辑可选择已授权比例。

2026-10-29 用户另行要求激活内置目录。内置范围收紧为生成最多 2 张、编辑输出 1 张、最多 5 张参考图，显式比例仅 1:1/16:9；分辨率与质量保持上述三档。该范围以自身六份真实捕获为依据，候选目录继续独立保留。源码激活不改原候选阶段或正式资格，详见[激活范围](catalog-activation.md)。

新增封闭的 XAIImagesOptions 与独立协议转换，将请求固定为同步 JSON、响应固定为 b64_json。整组图片验证通过后才返回内容；失败保留已经取得的合法用量、收费与诊断事实。

图像价格增加明确的按张分支。目录 Estimated Cost 继续通过 Attempt.CostEstimate 表达；Reported Cost 独立通过 Attempt.ReportedCostUSD 表达。token、图片数量、目录估算和厂商收费各自保留含义，未知信息明确标识。

## User Stories

1. As a Host developer, I want to generate Imagine images through GenerateImages, so that I can use the existing image integration.
2. As a Host developer, I want to edit one reference image through the same entry, so that editing needs no separate client.
3. As a Host developer, I want to edit multiple ordered reference images, so that the model can distinguish the intended subjects and their order.
4. As a Host developer, I want to resubmit a returned image for another edit, so that my Host can manage iterative refinement.
5. As a Host developer, I want HookedClient.GenerateImages support, so that existing trusted request customization remains usable.
6. As a Host developer, I want the route to use Image Model identity, so that image generation remains distinct from chat and classification.
7. As a Host operator, I want the first release limited to global grok-imagine-image-2.0, so that aliases and regional prices cannot silently change the service.
8. As a Host developer, I want model discovery without credentials, so that configuration can be inspected independently of execution.
9. As a Host developer, I want typed XAIImagesOptions, so that incompatible OpenAI and Google options fail clearly.
10. As a Host developer, I want one to ten generated images in one request, so that variations do not require a client-side request loop.
11. As a Host developer, I want one to five reference images and one edited output, so that the first release has a defined editing contract.
12. As a Host developer, I want 1k, 1.5k and 2k resolution choices, so that I can select the intended output tier.
13. As a Host developer, I want low, medium and auto quality choices, so that I can balance output quality and cost.
14. As a Host developer, I want omitted quality and resolution to retain vendor defaults, so that omission does not secretly select a library-specific tier.
15. As a Host developer, I want supported aspect ratios for generation and multi-image editing, so that I can request an appropriate composition.
16. As a Host developer, I want explicit aspect ratios rejected for single-image editing, so that unsupported behavior is not silently ignored.
17. As a Host operator, I want catalog and policy limits to narrow protocol capabilities, so that application budgets remain authoritative.
18. As a Host developer, I want explicit null options and typed nil pointers rejected, so that invalid requests cannot masquerade as defaults.
19. As a Host developer, I want prompt validation consistent with the existing image entry, so that malformed text fails before a vendor request.
20. As a Host developer, I want PNG, JPEG and WebP reference bytes accepted with matching MIME, so that common input formats remain usable.
21. As a Host operator, I want strict inline base64 validation, so that paths and external resource references cannot enter the image operation.
22. As a cloud Host operator, I want an explicit image Binding, so that a chat Binding does not authorize generation charges.
23. As a cloud Host operator, I want tenant-owned Binding resolution, so that one tenant cannot use another tenant's configuration.
24. As a cloud Host operator, I want the model allowlist enforced before credential access, so that a valid account key does not widen model authorization.
25. As a Host operator, I want a pinned Binding and Credential Snapshot, so that rotation cannot change an in-flight logical call.
26. As a local Host operator, I want credentials supplied through existing resolvers, so that ambient environment keys cannot change the selected account.
27. As a Host operator, I want caller-owned input and options snapshotted before resolvers run, so that later mutation cannot change the request.
28. As a Host developer, I want catalog and result copies to be independent, so that modifying one consumer's data cannot affect another call.
29. As a Host developer, I want text generation and image editing to select their native endpoints automatically, so that the request shape matches the operation.
30. As a Host developer, I want trusted Hooks to change authorized content and options, so that application-specific adjustments remain possible.
31. As a Host operator, I want Hook output revalidated after encoding, so that callbacks cannot bypass authorization or resource limits.
32. As a Host operator, I want Hooks prevented from switching generation and editing, so that endpoint selection remains stable within a logical call.
33. As a Host operator, I want URL, file and deferred-output injection rejected, so that a callback cannot introduce another resource lifecycle.
34. As a Host developer, I want a callback executed once before retries, so that each attempt uses the same frozen request.
35. As a Host developer, I want standard ImageOutputImage results, so that existing image consumers need no vendor-specific output handling.
36. As a Host developer, I want all returned images validated before publication, so that a partially invalid collection is never presented as success.
37. As a Host developer, I want MIME derived from image bytes when the vendor omits it, so that valid inline results remain usable.
38. As a Host developer, I want conflicting MIME and file headers rejected, so that malformed results do not reach consumers.
39. As a Host developer, I want empty results and malformed JSON distinguished from success, so that a successful HTTP status cannot conceal a failed operation.
40. As a Host developer, I want explicit moderation refusals classified correctly, so that the Host can explain why no image was produced.
41. As a Host developer, I want missing REST moderation flags handled honestly, so that undocumented fields are not mandatory success conditions.
42. As a Host developer, I want reported request IDs and serving models preserved when available, so that vendor incidents can be investigated.
43. As a Host developer, I want absent response metadata kept absent, so that requested values are not misrepresented as vendor reports.
44. As a Host operator, I want URL-only and deferred responses rejected without additional requests, so that one synchronous call cannot spawn downloads or polling.
45. As a Host operator, I want request and image limits checked before expensive copying and encoding, so that oversized inputs remain bounded.
46. As a Host operator, I want output image counts and bytes bounded, so that generated media cannot exceed the configured memory budget.
47. As a Host operator, I want shared admission and finite waiters, so that concurrent image requests respect process and tenant capacities.
48. As a Host developer, I want cancellation and existing timeout behavior, so that abandoned image operations release their resources.
49. As a Host operator, I want every terminal path to release response bodies and permits, so that failures do not exhaust the client.
50. As a Host operator, I want default retries disabled and explicit retries bounded, so that repeated requests do not create uncontrolled charges.
51. As a Host operator, I want no replay after a successful HTTP response, so that decoding and moderation failures do not regenerate paid images.
52. As a Host developer, I want prices expressed per input and output image, so that estimates use the vendor's actual pricing unit.
53. As a Host developer, I want output prices selected by quality and resolution, so that different image tiers receive different estimates.
54. As a Host operator, I want token and per-image prices mutually exclusive, so that the same consumption cannot be priced twice.
55. As a Host operator, I want complete, unique and finite price tables, so that missing rates cannot become free generation.
56. As a Host operator, I want catalog versions and hashes to include the price table, so that estimates remain attributable to a fixed snapshot.
57. As a Host developer, I want resolved estimates only when their inputs are known, so that a point amount has a verifiable basis.
58. As a Host developer, I want bounded estimates for auto quality when its possible charges are finite, so that I can reserve the upper amount.
59. As a Host developer, I want incomplete estimates to remain partial or unreported, so that uncertainty is not hidden as zero cost.
60. As a Host developer, I want provider-reported charges separate from estimated cost, so that I can distinguish actual reports from catalog calculations.
61. As a Host developer, I want ReportedCostUSD to distinguish missing amounts from explicit zero, so that free and unreported requests remain different.
62. As a Host operator, I want raw integer cost ticks retained in real captures, so that the reported USD amount can be independently checked.
63. As a Host developer, I want missing token usage left unreported even when a charge is known, so that money does not fabricate token consumption.
64. As a Host developer, I want reported reasoning and cached tokens counted once, so that diagnostic token totals remain consistent.
65. As a Host developer, I want valid usage and charges retained after image validation fails, so that a failed result does not erase known consumption.
66. As a Host operator, I want Observer records limited to approved metadata, so that diagnostics do not expose prompts, image bytes or credentials.
67. As a Host operator, I want simultaneous tenants and credential rotation covered through the public client, so that isolation and snapshot behavior are verifiable.
68. As a Host developer, I want existing OpenAI and Google image behavior preserved, so that the new route does not break current consumers.
69. As a Host developer, I want existing xAI chat behavior preserved, so that extending Provider identity does not change Responses semantics.
70. As a maintainer, I want public E2E failure scenarios written before implementation, so that the feature is driven by observable requirements.
71. As a maintainer, I want repeatable request, response, result and image artifacts, so that failures can be reproduced and reviewed.
72. As a Host operator, I want this image route to obtain its own live evidence, so that another route's success cannot become an unsupported claim.
73. As a Host operator, I want candidate capabilities separated from formal support, so that documented limits are not presented as tested capabilities.
74. As a Host operator, I want paid validation tied to an explicit finite scope and budget, so that research and implementation cannot silently authorize spending.
75. As a maintainer, I want release status to distinguish incomplete, failed and not-run stages, so that a partial result cannot become a release PASS.
76. As a Host operator, I want rollback by disabling the image Binding, so that new image calls can be stopped without changing chat or migrating data.

## Implementation Decisions

### 1. 操作身份与职责

- 唯一服务身份为 image × xai × xai-images × grok-imagine-image-2.0；公开常量采用 OperationImage、ProviderXAI、APIXAIImages。
- 全球基址固定为 https://api.x.ai/v1。沿用现有 xAI endpoint 规范化与校验，拒绝其他区域、查询参数、片段和用户信息。不自动解析旧型号别名或更换型号。
- 扩展图片入口分派、Binding 操作校验和 ImageModel 校验，精确允许该路线；不能只删除现有拒绝条件而放行所有 xAI 操作或型号。
- Host 继续拥有认证、授权、Call Scope、账户/凭据、输入图片访问权、持久化与多轮工作流。图片 Binding 显式授权该操作和型号；共享同一真实账户的凭据引用不等于共享操作权限。
- 依赖方向为公共图片入口→授权与快照→xAI 协议转换→共享 unary 运行时→规范结果。协议 DTO 不进入领域请求、结果或持久化模型。
- 复用既有 Client、HTTP 客户端、Hooks、Binding.Retry、Admission、取消、超时和 Observer；使用有界 JSON 直连，不新增 SDK、下载器、通用适配注册框架或作业系统。

### 2. 公共请求与选项

- 沿用 ImagesRequest 的 Prompt 和 ReferenceImages；输入为有序内联 Image，输出沿用 ImagesResult。参考图 slice 和可变选项在任何 Host resolver 执行前独立快照。
- XAIImagesOptions 实现现有封闭 ImageOptions，字段为 N、AspectRatio、Resolution、Quality，均保留 Nullable 的省略/null/值语义。nil 接口采用默认；typed nil、显式 null、其他协议选项和非法值拒绝。
- 生成 N 为 1–10，省略为厂商默认的一张。编辑参考图为 1–5 张，N 仅允许省略或 1。编辑多输出不在本期范围。
- Resolution 值为 1k、1.5k、2k；Quality 值为 low、medium、auto。省略时不注入库自己的分辨率或质量默认值；不将当前 auto 的选择永久写死。
- 生成和多图编辑的 AspectRatio 候选值为 1:1、3:4、4:3、9:16、16:9、2:3、3:2、9:19.5、19.5:9、9:20、20:9、1:2、2:1、21:9、5:2、auto。值须同时得到型号目录授权。
- 单图编辑拒绝任何显式 AspectRatio，包括 auto。此为已确认的保守边界，不能静默忽略参数或当作厂商永久限制。
- 每项数量、能力和字节限制取协议、目录与 Host 策略的交集。沿用 ImageCapabilities 的 MaxReferenceImages、MaxOutputImages、ImageSizes、AspectRatios、Qualities，不新增推测性能力体系。
- 提示词沿用公共图片入口的 1–32000 个有效 Unicode 字符约束。输入图片只接受 MIME 与文件头一致的 PNG、JPEG、WebP 和严格标准 base64；限制图片字节、数量和完整请求字节。

### 3. 原生 JSON 与最终请求守卫

- 无参考图时 POST /images/generations；有参考图时 POST /images/edits。端点在回调前确定，同次调用不得被 Hook 切换为另一种操作。
- 生成请求包含授权型号、提示和已提供的数量/比例/分辨率/质量字段；响应交付固定为 response_format=b64_json。
- 一张参考图使用 image 资源对象，多张使用按原顺序排列的 images 对象数组；每个资源包含 type=image_url 和由本次内联字节构造的 data URI，image 与 images 互斥。
- 不发送 multipart；不从聊天协议复制 store、include、会话或原生续接字段。不允许 deferred、stream、外部 URL、file_id、托管输出、上传目标或其他资源生命周期。
- 回调前先检查已知长度与输入限额，再编码。Hook 后对最终独立字节严格解析，重做 schema、重复字段、null、型号、选项、参考图、数量、MIME 和字节验证。
- Hook 可修改已授权内容和选项，但不能改变授权型号、目标、响应交付方式或端点类别；编辑必须保留合法且非空的参考图。费用和响应校验均以最终冻结请求为依据。
- 保留既有错误分类：无效入口为 invalid_request；回调生成非法 schema 为 callback_failed；越过授权/资源边界为 tenant_denied；容量超限为 resource_limit。拒绝尽可能发生在凭据读取或 HTTP 发送前，不增加厂商尝试。
- Hook 只执行一次；显式获准的初始请求重试复用冻结后的相同请求体，不重复读取配置或更换凭据快照。

### 4. 同步响应与原子结果

- 仅接受一次同步完整结果，按返回顺序转为 ImageOutputImage；不产生 delta、聊天 assistant 消息或 Native State。
- 必须取得合法且符合最终请求数量的完整图片集合；数量不符、空图或任一图片失败均不得发布部分 Content，也不得补发生成请求。整组失败保留已经独立解析成功的用量、收费和诊断元数据。
- b64_json 必须是裸的严格标准 base64，遵守单图、总图像字节和总响应字节限额。拒绝 URL 交付或混入外部资源，不发起下载。
- MIME 上报存在时必须与文件头一致；缺失/null 时只从支持的 PNG/JPEG/WebP 文件头识别。未知或冲突格式失败，不固定为 OpenAI 的默认 PNG，也不转换编码。
- 仅根据明确审核信息判断审核拒绝并映射 upstream_error。REST 未承诺 respect_moderation 必返，因此字段缺失本身不构成失败；空结果缺乏明确审核依据时按损坏/不完整协议处理，不伪造审核原因。
- 损坏 JSON、重复响应字段、不完整响应、非法计数和异常协议结构映射 protocol；取消沿现有 aborted 语义；传输、超时、容量和非成功 HTTP 沿已有分类。HTTP 202 或异步 request_id 不能成为同步成功。
- ResponseID、ResponseModel、ProviderRequestID 仅保存真正上报且合法的事实。缺失保持未报告；请求型号不能伪装成实际服务型号，也不能用响应型号覆盖授权目标。
- 保持先登记合法消费事实、再校验图片的顺序。不能因为后续解码或图片错误丢失已报告收费；也不能因收费已知而将图片失败改成成功。

### 5. 按张价格、用量与报告收费

- ImagePricing 增加可选 PerImage 分支，类型为 ImagePerImagePricing。输入图片单价使用 InputImage；输出单价表使用 Output，每行 ImageOutputPrice 包含 Quality、ImageSize、Price。单价使用 Nullable 浮点值，单位为 USD/张，来源和日期由 Source 记录。
- PerImage 与现有 token/cache 费率互斥；按张分支不要求虚构文本或图片 token 费率。现有 token 字段保持每百万 token 单位和旧目录含义。
- 对所有开放的 low/medium × 分辨率组合，费率必须唯一、完整、非负且有限；显式零费率合法，缺失/null 不等于免费。auto 是选择策略，不是固定价格行。
- 按张价格指针与列表参与目录深复制、版本和内容哈希。没有新分支时保持既有 token 目录的编码含义；价格变化必须换目录版本，历史快照不覆盖。
- 2026-10-29 的初始公开价目快照为输入图片每张 USD 0.01；输出 low 的 1k/1.5k/2k 分别 USD 0.04/0.05/0.06，medium 分别 USD 0.06/0.07/0.08。该数据是带日期的目录依据，不是永久费率或厂商最终账单承诺。
- 目录估算复用 Attempt.CostEstimate，不新增计费引擎。只有图片数量、适用费率和档位依据齐全时为 resolved；数量已知但 auto 的实际质量未确定、且可限定在已知费率之间时为 bounded；事实不足时保持 partial/unreported。
- 有界费用提供 Minimum 和 Maximum，Host 预算使用 Maximum。若分辨率或其他计费事实无法限定，不虚构有限上界。不能从厂商报告金额反推图片计数、折扣或目录单价。
- 生成估算依据最终请求及完整合法返回集合；单输出编辑按已知输入图数量与输出档位估算。请求 n 不是失败时的实际计费张数；编辑 n>1 的输入费重复计量公式不作本期承诺。
- Attempt 新增独立 ReportedCostUSD，类型为 Nullable[float64]，JSON 名为 reportedCostUsd。仅表达厂商报告本次已收费金额，不写入 CostEstimate 或 Usage.Cost。
- 原生 usage.cost_in_usd_ticks 先严格解析非负 int64，拒绝小数、溢出和重复字段，再除以 10^10 换算有限 USD。缺失/null 为未报告，显式零为已报告零。真实捕获保留原整数，公共浮点金额不承诺逐 tick 无损累计，不新增第二份整数金额双写。
- token 用量仅转换实际上报计数。input_tokens 已含文本和图片；cached_tokens 为文本子集；output_tokens 已含重写文本、reasoning 和图片，Reasoning 是输出内分项，均不重复累加。
- 完整可核对计数关系须检查非负整数、溢出和总量一致性。缺项保持 partial/unreported；图片张数不能装进 token 或 Modalities 字段。
- token 全未报告时，Usage 保持既有零值合同；按张估算仍可由 Attempt.CostEstimate 独立提供。明确补充既有成本说明的适用范围：本图片路线不依赖 Usage.Cost 承载费用下界；不得据零 Usage 或未报告 token 推导免费。
- 失败只保留有依据的事实；完整图片验证未通过时不补造输出张数来获得 resolved 估算。合法 Reported Cost 与部分 token 报告不因 Content 清空而消失。

### 6. 生命周期、观测与回滚

- ImagePolicy 必须显式启用，继续使用有限的输入/输出数量、单图和总字节容量，以及既有请求、响应、准入、排队和超时策略；不引入无限容量或新全局可变状态。
- 取消、连接/响应头/读取空闲超时、成功、拒绝、解析失败及 Hook 失败均关闭正文并释放准入与调用资源。准入持有到响应关闭和验证结束。
- 默认不自动重试；Binding.Retry 仅控制现有有限初始请求策略。2xx 后的读取、解码、审核、图片验证或回调失败均不重放；不宣称网络重试具有厂商侧幂等性，保留可能重复收费的含义。
- Observer 使用既有异步有界投递与终态归属。独立收费字段随 Attempt 值复制传播，并加入允许的标量元数据审计；不记录提示、图片、revised prompt、原始响应、密钥或错误正文。
- 正常调用不新增自动正文捕获。真实验证使用现有有界捕获与脱敏审计保留请求/响应和精确 ticks，敏感凭据不进入产物。
- 回滚通过 Host 禁用该 image Binding 并更新版本；已固定的调用按原快照结束或由 Host 取消。保留历史目录、结果与证据，不自动切换聊天、旧型号别名或其他 Provider。本功能无数据库迁移和内部兼容层。

### 7. 候选、正式支持和实施顺序

- 先在公共边界建立失败场景，再实现最小完整路线和独立候选目录。候选设计支持的上限不等于正式目录已验证能力。
- 对实现、目录、夹具、工具和实际输入审查并冻结，完成受影响普通验证后，再按现有验证流程获得本图像组合自身真实证据及适用人工阶段事实。
- 正式支持限定为自身证据实际覆盖的型号和能力；未验证的更宽能力继续留在候选，不以官方文档、共享代码、xAI 聊天或其他图片 Provider 的通过替代。
- 正式目录与资格报告须绑定最终适用来源和价格快照，避免验收后修改目录使来源失配。源码接线、真实资格、正式支持和发布分别记录，不提前宣布完成。
- 使用现有验证工具和本路线的捕获/资格扩展，不复制授权、预算、审计或发布控制器。需要新增厂商事实时，先形成精确场景、有限调用/尝试/图片数、质量/分辨率、USD 停止阈值和来源，再取得该次明确付费授权。

## Testing Decisions

- **主要测试边界已经确认**：用户在 Q7 接受沿用公共 Client E2E。通过公开 Client/HookedClient 的 GenerateImages、NewClient 和 Catalog 查询，配合受控 HTTP 上游、Host resolver 与 Hook，验证完整可见行为。不新增可供业务绕过公共边界的测试入口。
- **好测试的标准**：断言输入接受或拒绝、实际请求/尝试、结果/错误、估算与收费、终态观测、资源归零和隔离。使用可重放的协议数据与确定性时序，不断言私有函数、文件布局或恰好相同的内部调用顺序。
- **先失败后实现**：先写可证明缺失行为的公共 E2E 并保留失败，再实现并重跑。遵循项目测试原则；不得实现后补单元测试。确需隔离测试时，必须先列出完整失败方式。
- **既有测试先例**：沿用 OpenAIImagesFixtures、GoogleImagesFixtures 的协议夹具模式，OpenAIImagesEditingHooks 的最终请求守卫模式，OpenAIImagesResources 的限额/取消/释放模式，以及 XAIHostConcurrentIsolation、XAIHostDisableAndSnapshots 的租户与快照场景。复用产物和审计能力，测试 xAI 自身 wire 契约。

| 验收组 | 必须验证的外部行为 |
| --- | --- |
| 身份与授权 | 正确 image 身份可达；chat/classifier 跨入口拒绝；错误租户、型号、协议、endpoint 和白名单失败；应在凭据/HTTP 前拒绝的输入保持零凭据读取或零厂商尝试。 |
| 目录与快照 | 候选可发现但不授予调用权；能力/价格表非法时构造失败；PerImage 与 token 互斥；价格表独立复制与哈希变化；旧 token 目录编码含义保留。 |
| 选项与默认 | nil 默认、typed nil/null、错协议、缺省质量不注入 low、三档分辨率、生成数量边界、编辑 n=1、参考数边界、比例枚举与单图显式比例拒绝。 |
| 生成与编辑 | 生成、单图、多图的最终 JSON/端点正确；参考图顺序保持；内联 PNG/JPEG/WebP 验证；不发送 multipart 或聊天状态字段。 |
| Hook 最终守卫 | 修改合法内容成功；改型号/交付/端点类别/外部资源失败；已知超量载荷在昂贵处理前拒绝；raw/带类型容器/冻结字节重验；回调仅执行一次，重试 body 相同。 |
| 图片结果 | 多图完整有序发布；缺图、数量不符、重复/损坏 JSON、base64/MIME 冲突、未知格式和单图失败整组清空；MIME 缺失可按支持文件头识别。 |
| 拒绝与同步 | 明确审核拒绝、缺少审核标志、无明确原因的空结果分别处理；URL/异步/202 不触发下载、轮询或成功；2xx 后失败零额外尝试。 |
| 按张估算 | low/medium × 三档分辨率、生成多输出与编辑输入费；最终 Hook 选项生效；auto 区间上下界；事实不足保持未知；不能用请求 n 或假 token 补齐失败费用。 |
| 报告收费与 token | ticks 的缺失/null/零/合法正值/负数/小数/重复/溢出；USD 换算；已收费但 token 缺失；部分/完整 token 算术；reasoning/cache 不重复计入；失败保留合法消费事实。 |
| 生命周期与隔离 | 取消、各阶段超时、准入等待/释放、正文关闭、重试边界、同名不同租户配置、调用中撤销/轮换、输入与目录 mutation、结果与 Observer 副本隔离。 |
| 相邻回归 | OpenAI/Google 图片生成与编辑、旧 token 计价，以及 xAI Responses 的身份/endpoint/快照行为；共享改动再按实际影响扩大普通范围。 |

- 每轮公共 E2E 通过现有验证入口保留可核对且可重复的产物：最终请求/响应、结果与错误、Observer、目录版本/hash、图片、调用/资源统计、manifest 和 audit；有限必需场景必须实际执行，零匹配或父测试 PASS 不算完成。
- 若修改 live harness、资格消费者或 CLI，增加其现有进程边界的受控检查；无真实凭据、零厂商请求的受控/回放通过不能取得真实资格。
- 本组合 live 至少证明生成、单图编辑、多图顺序、拟声明选项以及真实用量/费用形状。首期设计上限和更多选项只有获得相应自身证据才能成为正式声明；不要求未经选择的笛卡尔积付费调用。
- 真实测试须固定自身身份、来源、二进制、目录和实际输入，使用本路线唯一凭据与新授权，保留原始捕获、整数 ticks、消费记录、失败和真实退出。超限或失败按有界策略停止，不自动扩大预算或覆盖失败。
- 验证阶段遵循普通 E2E→来源审查冻结→受影响 live→人工选择的 race→人工选择的压力→只读发布判定。race/压力按本次并发和大载荷影响选择；未选为 NOT_RUN，有依据的不适用为 NOT_APPLICABLE，二者不能混淆。
- 大载荷与压力候选包括已选多图输入、多输出响应、并发准入、取消、读取和字节限额；具体负载与预计耗时在实施后按真实修改面提交人工选择。本 spec 不自动启动这些阶段。

## Out of Scope

- Imagine 视频、语音、图生视频，以及 Responses 的 image_generation 托管工具。
- 1.0、quality/pro 旧型号与别名、其他地区 endpoint、其他服务等级或账号套餐推断。
- 编辑一次多输出；批量编辑 n>1 的输入费公式和自动补图。
- 单图编辑显式改比例；mask、透明背景、指定输出编码/压缩、seed、任意精确像素尺寸等缺乏本目标充分依据的能力。
- 公网图片下载、Files API、file_id、上传/托管输出、签名 URL 目标、deferred/polling/webhook 和后台持久作业。
- 新客户端、另一套网络/重试/授权/准入系统、通用插件架构、通用计费引擎、账单系统及逐 tick 无损公共货币接口。
- 将图片张数伪装成 token、将目录估算伪装成厂商收费、缺失价格填零、未知默认映射固定化和失败后隐式协议回退。
- 新用户体系、Secret Store、UI、数据库迁移、其他 Provider 的功能扩展和内部兼容 shim。
- 通过本次规格生成自动实施代码、调用付费 API、启动人工验证阶段、合并或发布。

## Further Notes

### 唯一当前状态

- 2026-10-29新指令：用户明确要求继续完成正式资格、race和压力，并指定代理127.0.0.1:7890。正在补本路线只读正式消费者与缺失的有限自身压力场景；尚未执行这三项新验证。先普通与复核冻结，再正式→race→压力；具体argv、必交case、有限负载、预计耗时及授权原件见[本轮范围](final-stages-plan-20261029-05.md)。现有六场景足以核验当前受限Builtin，无新增厂商事实需求，不新发付费grant或重跑。
- 2026-10-29：按用户“激活xAI目录”指令完成内置图片目录源码激活，工单01–06完成。完整规格仍部分完成：正式原生资格、人工race/压力与完整发布未完成。集成分支`codex/xai-imagine-integration`，不合并main、不push、不发布。
- 当前Builtin：image × xai × xai-images × grok-imagine-image-2.0，目录版本`2026-10-29.1`，hash `sha256:f44d07c6969346f056cc46d31fccbafb68827267910801c05bb314ac0c55348d`。生成最多2张、编辑输出1张、最多5参考图；resolution=1k/1.5k/2k，quality=low/medium/auto，显式比例仅1:1/16:9；单图编辑仍拒绝显式比例。候选仍独立保留10输出/16比例，不把候选上限导入Builtin；未覆盖完整笛卡尔积或视觉语义效果。
- 已交付：Client/HookedClient.GenerateImages文生图、单图/有序多图编辑，类型化选项、严格请求守卫、原子图片、MIME识别、取消/超时/准入/租户快照、按张估算及独立Reported Cost。本次只增加目录metadata/版本/快照、留存捕获回放和当前示例配置选择，产品协议、网络、凭据、重试、生命周期与计价单位不变。
- 目录价格：输入USD0.01/张；输出low×1k/1.5k/2k=0.04/0.05/0.06，medium=0.06/0.07/0.08，来源日期2026-10-29。auto继续区间估算，不从厂商收费反推固定质量。新旧快照核对其余117个型号全部字段相等；旧候选JSON/版本/hash与原测量所绑定字节保持。证明见`.evidence/barness-ai-xai-imagine/status/catalog-activation-diff.json`及[激活范围](catalog-activation.md)。
- 本次目录公共入口先红后绿：NewClient/Catalog默认发现、能力/价格/深复制、原生六场景逐份最终wire、7图hash/MIME、收费/ticks、token未报告及未开放能力预拒绝。原生捕获原件未改，必要协议夹具独立固定来源。初审1项P2（当前版本误绑历史激活版本）由同一实现者修复；两轴复核与配置增量复核均0遗留项，见[目录审查](review-catalog.md)。
- 中文评估相邻修复：目录激活后首次完整普通运行`20261029T015941.735420Z-oy93ipx2` FAIL/exit1，旧当前配置报`evaluation: task, model, dataset or price snapshot drift`，31个失败测试事件及1个包失败原件保持。新建三份v2配置，仅改变配置版本、目录版本/hash；v1字节、Plan.Load严格校验、历史消费、数据集、型号/价格和预算均保持。CLI默认与当前E2E选择v2。独立冻结98事件绿回执`20261029T020714.918159Z-ypaf148b`及本次最终受影响普通回执证明修复，失败明细见`.evidence/barness-ai-xai-imagine/status/20261029-catalog-active-closure/initial-full-failure.json`。
- 第二轮完整普通运行`20261029T021223.616298Z-3e5asmtb`保持FAIL/exit1、744.935s，仅TestToolRoundTrip/live-view-not-executable与父测试失败。共享PartialView按读取时刻返回消息，旧测试没有固定运行中时刻。确定性红复现与受控AfterChunk修复只改单个测试，保留原拒绝断言，新增Pending与释放后终态可执行对照；产品语义、provider/helper及fixture保持。失败原件见`.evidence/barness-ai-xai-imagine/status/20261029-catalog-active-closure/second-full-failure.json`，时序两轴复核及SOP范围依据见[目录审查](review-catalog.md)。未把这两次全包FAIL改成PASS；最终范围仅覆盖实际受影响测试及./ai全部实际测试，两个chineseeval包本身无测试，仅编译。
- 当前冻结普通验证（受影响范围）：PASS/exit0；实际argv：`python3 tools/verification/run.py --feature barness-ai-xai-imagine --stage e2e -- go test -json ./ai/e2e ./ai ./ai/examples/chineseeval ./ai/examples/chineseeval/cmd/chineseeval -run '^(TestXAI|TestBuiltinCatalogReleaseSnapshot$|Test.*Catalog|TestOpenAIImages|TestGoogleImages|TestChineseEvaluation|TestToolRoundTrip$|TestToolCallTruncation$|TestStreamLifecycle$|TestThinkingBudgets$|TestCostEstimate$)' -count=1 -timeout=20m`。墙钟127.374s，2058个PASS测试事件，全部1个原生bundle和audit有效。唯一主包1963个PASS case、0个NOT_RUN case；NOT_RUN不冒充普通必交通过，有限完整必需PASS集合固定在新索引。
- 当前普通原件：`.evidence/barness-ai-xai-imagine/e2e/20261029T023159.242880Z-w8ibgpxm/receipt.json`，SHA256 `1912803cdef936263a0d3dc14068822850a20f110388f5bfe4ca98f18146e40f`；主manifest `.evidence/barness-ai-xai-imagine/e2e/20261029T023159.242880Z-w8ibgpxm/bundles/20261029T023202.485682000Z/manifest.json`，SHA256 `5125e438060e6c1a73f4414bfe68139594b67cfc86ddfe66bcf104ac3beec99e`；audit SHA256 `8aa973100a15bf817985395637131e838cf84e427be4792583def923f020f5fa`。其他包的路径/hash、完整测试库存、实际selectedInputs见同一回执。
- 当前来源：最终普通实际受测提交`7d4edddddfcc3edb8485b6ed889ba3e2c3287e9b`；代码合入`0219357ce7f5e42b526cae784bb66b3a4c939986`。收尾仅Markdown及scratch不可变索引；当前2088文件完整来源与实测集合/hash相等，selectedInputs逐项相等，最终当前提交由新只读报告releaseCommit给出。与原真实测量`724e988132ac5bf6ac393db156e07081eb0e2427`相比15个目录/当前配置/测试夹具文件不同（含仅测试时序修正）；逐项分类见`.evidence/barness-ai-xai-imagine/status/20261029-catalog-active-closure/source-delta-final.json`。
- 历史普通原件保持：原完整`./ai/...` PASS/exit0为`20261029T003310.813167Z-869dtg8y`（16510事件）；代理harness过程PASS为`20261029T011128.326343Z-huqny4b5`（29事件）。最早历史来源测试误用当前checkout的FAIL、定向红绿及后续激活/配置修复的失败全部保留。历史来源修正只固定218项旧Git字节，不改consumer/资格/白名单，详见[原审查](review.md)。旧当前状态完整快照`.evidence/barness-ai-xai-imagine/status/20261029-catalog-active-closure/prior-spec.md`及record-01/02/03保存完整先前状态。
- 原真实候选：PASS/exit0、mode=live、stage=candidate、complete=true、audit通过、formalQualified=false；实际6调用/6HTTP、8输入/7输出，全部HTTP200，无重试，87.489s，proxy=`http://127.0.0.1:7890`。收费原始ticks合计5100000000，厂商报告USD0.51；目录预留上界USD0.53，停止阈值USD0.60。六次token均未报告，保持UsageUnreported，不填零。7图字节/hash/顺序、金额与权威两条grant记录独立核对。
- 原候选回执`.evidence/barness-ai-xai-imagine/live/xai-images/20261029T012001.157782Z-n96g7rg4/receipt.json`，SHA256 `790df98d96f614557f0201280d0879f49ee08acb0a35531ec491ad5d0892e8b4`；完整授权、原始请求/响应/图片、整数ticks、Observer、二进制/来源与claim见[候选不可变索引](evidence/candidate-live-20261029-03.json)。首次直连1次ConnectTimeout15s、FAIL/exit1、其余5场景NOT_RUN及已消费授权保持。新代理执行授权不复用；本次激活没有新厂商调用或付费授权。
- live适用性：本次目录激活收紧至自身六场景已测边界，当前默认Client离线逐字节回放，不新增厂商事实需求，因此不重买live。原候选的stage/mode/formalQualified不改；正式原生资格仍NOT_RUN，候选成功与源码激活不等于正式支持。本次不授予其他Provider当前资格；原件和资格按路线/型号保留。
- race：NOT_RUN。整个功能仍影响共享准入/取消/快照/Observer与新路线并发；所需live完成后由人选择具体argv与范围。候选命令`python3 tools/verification/run.py --feature barness-ai-xai-imagine --stage race -- go test -race -json ./ai/e2e -run '^(TestXAIImagine|TestXAIImages)' -count=1 -timeout=20m`未获选择或执行。目录激活增量本身无共享状态变化，不替代整个功能的人工阶段。
- 压力：NOT_RUN。整个功能仍涉及5参考输入、10输出候选响应、多图字节上限、并发准入/取消及读取；race通过后另行选择有限负载、argv和预计耗时。本次未选择或执行，不标NOT_APPLICABLE。
- 当前不可变[激活索引](evidence/catalog-activation-20261029-04.json)和[发布索引](evidence/record-20261029-05.json)绑定实际原件。只读发布报告`.evidence/barness-ai-xai-imagine/release/20261029-05/release-report.json`，原生命令/退出见同级`20261029-05-command/command.json`。来源和当前普通验证通过；正式原生资格/race/压力缺少，release.pass=false。record-01/02/03/04与report-01/02/03/04保持，不重写历史结论。
- 构建与回滚：当前只读releasegate构建原件`.evidence/barness-ai-xai-imagine/build/20261029-catalog-active-verified/build.json`，exit0。没有新live/race/压力/发布执行。Host可禁用image Binding并更新版本；源码回滚同时恢复目录内容/版本/快照与当前示例配置选择，已固定调用按原快照结束或取消。无数据库迁移、协议回退或内部兼容层。两个实施工作树已通过应用归档，主checkout按应用限制保留为同步副本；全部已知原件在根`.evidence`完整保存。

### 依据与待证事实

- [详细设计](design.md)、[官方研究](research.md)、[ADR-0056](../../docs/adr/0056-barness-ai-xai-images.md)、[领域词汇表](../../GLOSSARY.md)。研究结论固定于 2026-10-29；本次整理不新增官方实测结论。
- [操作与 Binding 授权 ADR-0020](../../docs/adr/0020-barness-ai-model-operations-and-binding-authorization.md)、[同步原生图像 ADR-0022](../../docs/adr/0022-barness-ai-openai-images-unary.md)、[xAI 聊天边界 ADR-0054](../../docs/adr/0054-barness-ai-xai-responses.md)。图片接入补充操作边界，不迁移聊天协议。
- [测试流程 SOP](../../docs/agents/verification.md)、[本地 Issue tracker](../../docs/agents/issue-tracker.md)、[triage 名称](../../docs/agents/triage-labels.md)。本文件即发布到项目本地 tracker 的 spec。
- 官方资料对 1.5k 的指南列举、单图比例、quality schema、审核标志和响应型号存在差异；本 spec 采用已经确认的限制，保留来源差异，不以 SDK 某个字段代表所有 REST 版本。
- 输入/请求/输出的厂商字节硬上限、审核混合结果和返回档位的 presence 仍需自身证据。Host 有限资源策略始终生效；未知厂商硬限额不套用聊天接口数值。
- 旧 quality 型号在 2026-11-02 的迁移已纳入型号取舍，首期直接使用 2.0；auto 的当前映射只是带日期的官方事实。
- 后续实施如需扩大范围，应先更新本交付契约和相关决定，不能用待证事实静默增加能力或费用承诺。

## Comments

2026-10-29：通过异步问题向用户提出 Q1–Q3；独立继续事实调查和草案编写。未确认的建议不作为实施授权。

2026-10-29：用户确认 Q1 文生图＋参考图编辑、Q2 仅 2.0、Q3 纳入按张估算；已更新设计树和 GLOSSARY。发出第二轮 Q4–Q6，等待能力上限、默认行为及收费金额精度的选择。[ADR-0056](../../docs/adr/0056-barness-ai-xai-images.md) 保持 proposed。

2026-10-29：文档检查 `git diff --check` 实际退出 0；对研究、设计、当前状态、ADR、GLOSSARY 五份文档检查本地引用及代码围栏，实际退出 0，缺失引用和未闭合围栏均为 0。未运行产品测试或付费调用，不能据此声明能力已交付。

2026-10-29：用户确认 Q4–Q6 全部推荐项，已同步设计和领域术语。等待 Q7 对整体方案达成共同理解，ADR 暂保持 proposed。

2026-10-29：用户确认 Q7 整体设计共识。本轮研究与方案完成，ADR-0056 更新为 accepted；后续实施未开始，图片路线真实支持和发布均未取得证明。

2026-10-29：按用户调用的 to-spec 技能重新生成完整功能规格，Type 更新为 feature，保留 ready-for-agent、Q1–Q7 共识与此前 Comments。沿用已确认的公共 Client E2E 边界，本轮不重新访谈、不实施代码、不执行产品测试或付费调用。

2026-10-29：重新生成后的文档核对通过，实际退出 0：模板七个章节及 Comments 顺序正确，76 条用户故事连续且符合指定句式，实现决定无源码文件路径或代码片段，本地引用与格式有效；git diff --check 实际退出 0。此结果仅证明规格文档检查通过。

2026-10-29：用户原回复“选择1，请使用本地代理 127.0.0.1:7890”。按新范围一次性执行6个真实候选场景，全部PASS，进程退出0，厂商报告USD0.51。正式晋升、race、压力及发布未执行；完整事实只维护于唯一当前状态。

2026-10-29：用户要求“激活xAI目录”。受限图片条目进入Builtin；默认Client重放原真实六场景。当前受影响普通验证PASS/exit0，原两次全包FAIL保持；新v2中文评估配置保留旧v1与严格历史边界。完整事实仅维护于唯一当前状态。
