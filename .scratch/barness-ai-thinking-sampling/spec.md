# barness-ai 按 thinking level 映射采样参数

Status: ready-for-agent
Updated: 2026-10-11
Type: feature

本文件定义交付契约。基于 `39ab6a5` 的公共采样实现与单一 1.1.0 oracle 来源/差分已完成；完整发布验证尚未通过，当前 live/effect 失败及一次未确认原因的压力失败均保留实际状态，尚未正式交付。逐项结果见[实际证据](evidence.md)。

配套：[详细设计](design.md)、[验收矩阵](acceptance.md)、[实施顺序](design.md#8-实施顺序)、[ADR-0025](../../docs/adr/0025-barness-ai-thinking-level-sampling.md)。依据：[研究第 9 节](../pi-1.1.0-thinking-sampling/research.md#9-barness-ai-接入映射与建议尚未实现)。

## Problem Statement

barness-ai 目前仅支持模型通用 `SamplingParams` 和调用级 `SamplingParams`。同一模型在关闭思考或不同思考档位下需要不同默认值时，Host 必须在每次调用前自行计算；直接按请求档位或厂商字符串选择，又会与模块既有的模型能力夹取产生偏差。

Pi 的 `samplingParamsByThinkingLevel` 将这项配置放在模型元数据中，按模型支持的 canonical thinking level 选择默认值。该特性始于 1.0.2，1.1.0 已包含。barness 的既有兼容基线是 1.0.0，不能在未迁移 oracle、目录数据与发布证据的情况下宣称已兼容 1.1.0。

“实际采用的档位”在此仅指客户端依据模型元数据求得的 canonical level。它不表示最终请求 JSON 中的厂商值、服务端默认强度或模型实际消耗的推理 token；full 入口与回调覆写尤其需要明确这条边界。

## Solution

在 Chat Model 中增加可选的模型级 `SamplingParamsByThinkingLevel`，由既有采样请求构造路径消费。按既有模型能力规则夹取输入档位，再将通用默认、该档默认、调用覆盖逐 key 浅合并，最后执行既有 `OnPayload` 回调和授权校验。

首期覆盖 OpenAI、DeepSeek 各自的 Chat Completions 与 Responses 路由，保持 Provider 与 API Protocol 的独立身份。Anthropic Messages、Google Gemini 原生聊天路由继续忽略采样 map；图像生成和分类操作不获得该字段。

以公共 Client E2E 验证请求、失败、隔离和生命周期，通过一个冻结 pi-ai 1.1.0 oracle 判定适用路线的差分。DeepSeek Responses 继续使用自己的扩展路线证据。基线迁移、目录快照、账本及正式发布证据作为同一交付处理。

## User Stories

1. As a Host developer, I want general sampling defaults on a Chat Model, so that calls share a consistent starting configuration.
2. As a Host developer, I want sampling defaults per canonical thinking level, so that reasoning modes can use different parameters.
3. As a Host developer, I want an off profile, so that omitted reasoning can select defaults for the non-thinking mode when supported.
4. As a Host developer, I want per-call sampling overrides, so that a single call can override model defaults.
5. As a Host developer, I want named option precedence documented, so that I can choose the correct override mechanism.
6. As a Host developer, I want unsupported levels clamped using model capabilities, so that sampling follows the supported canonical level.
7. As a Host developer, I want mapped provider strings kept separate from canonical keys, so that different levels can share a wire value without losing their defaults.
8. As a Host developer, I want off-null models to select their lowest supported profile, so that omitted reasoning follows the same selection rule.
9. As a Host developer, I want non-reasoning models to select off, so that supplied reasoning does not select an inapplicable profile.
10. As a Host developer, I want undeclared xhigh and max levels to fall back consistently, so that extended levels require explicit model support.
11. As a Host developer, I want missing profiles to retain general defaults, so that sparse configuration remains useful.
12. As a Host developer, I want empty profiles to add no overrides, so that an empty object has a predictable meaning.
13. As a Host developer, I want zero, false and null preserved, so that explicit values are not mistaken for missing values.
14. As a Host developer, I want nested objects replaced as a whole, so that merge behavior matches pi.
15. As a Host developer, I want Stream and Complete to use the same defaults, so that delivery mode does not change my request.
16. As a Host developer, I want full and simple entries to retain their respective reasoning semantics, so that existing protocol behavior remains compatible.
17. As a Host developer, I want Responses summary-only calls to start selection at medium, so that summary configuration matches pi.
18. As a Host developer, I want DeepSeek thinking switches handled by its existing adapter, so that profile selection does not alter its wire contract.
19. As a Host developer, I want WithHooks to retain the same sampling behavior, so that attaching trusted callbacks does not create another configuration path.
20. As a Host developer, I want OnPayload to run after sampling resolution, so that a trusted final payload override remains effective.
21. As a Host developer, I want post-sampling reasoning overrides to avoid reselection, so that request construction has a single deterministic direction.
22. As a Host developer, I want invalid profile keys rejected at configuration boundaries, so that typos fail before model traffic.
23. As a Host developer, I want malformed profile JSON rejected even in inactive levels, so that accepted catalog snapshots remain serializable.
24. As a tenant administrator, I want reserved protocol fields rejected in every applicable profile, so that model metadata cannot change authorized routing or hosted resources.
25. As a tenant administrator, I want catalog profile maps and JSON bytes copied, so that one caller cannot mutate another caller's configuration.
26. As a tenant administrator, I want retries to reuse one authorized payload, so that later configuration changes cannot alter an ongoing logical call.
27. As a tenant administrator, I want concurrent calls isolated, so that tenants sharing model identifiers cannot affect each other's defaults.
28. As a Host developer, I want profile changes reflected in catalog hashes and versions, so that requests and prices remain attributable to a reproducible snapshot.
29. As a Host developer, I want existing configurations without profiles to retain their request behavior, so that adoption is incremental.
30. As a Host developer, I want native Anthropic and Gemini routes to keep ignoring these maps, so that unsupported protocol parameters do not leak onto their requests.
31. As a maintainer, I want one verified 1.1.0 oracle, so that differential results have one unambiguous baseline.
32. As a maintainer, I want published model data verified before copying defaults, so that catalog values have a reproducible upstream source.
33. As a maintainer, I want differential findings classified individually, so that unsupported routes and unrelated changes cannot be hidden as parity.
34. As a maintainer, I want repeatable public Client evidence, so that failures can be reproduced without calling private helpers.
35. As a maintainer, I want explicit release and rollback criteria, so that an incomplete migration cannot become the published contract.
36. As a maintainer, I want examples to distinguish illustrative values from provider recommendations, so that configuration mechanics do not imply vendor approval.

## Implementation Decisions

### R01 — 配置属于 Chat Model

新增可选模型级字段 `SamplingParamsByThinkingLevel`，JSON 名为 `samplingParamsByThinkingLevel`；使用既有 ThinkingLevel key 和逐参数原始 JSON 值表示。仅接受 `off/minimal/low/medium/high/xhigh/max` 七个 canonical key。调用级仍只有 `SamplingParams`，不增加调用级分档 map，不扩大 SimpleOptions 的 Reasoning 输入枚举；调用方以省略表示未请求思考。

Host 负责组装 Catalog。本特性不新增 models.json 解析器、modelOverrides 组合器或配置继承服务；传入 Client 的模型元数据已是 Host 选定的最终配置。

### R02 — 选档只依据模型能力

复用既有 supported-level 和 clamp 规则，顺序为 off、minimal、low、medium、high、xhigh、max。已支持的请求档位保持 canonical key；不支持时先向上再向下寻找，全部不支持时保留 pi 的 off 兜底。非推理模型只选择 off；普通档位默认支持，null 标记不支持；xhigh/max 只有显式非 null 映射才支持。

采样 key 不由 wire 字符串反查。off 映射为 none 仍选择 off；off 为 null 则向上寻找最低支持档。不推断厂商的默认强度，不依据响应 token 重新选档。

### R03 — 输入档位保持入口语义

Chat full 以显式 ReasoningEffort 或 off 起算；Responses full 以显式 ReasoningEffort 起算，只有 ReasoningSummary 时以 medium 起算，否则以 off 起算。Simple 路径保留既有显式 reasoning 夹取及 off 转为省略 effort 的行为；省略 reasoning 时采样仍从 off 夹取。非推理模型最终选 off。

Full 的 wire reasoning 继续按既有原始选项编码，不因本特性统一夹取；因此 full wire low 可以与采样 high 并存，summary-only wire medium 可以与另一个采样档并存。这个兼容行为必须写入使用说明与 E2E，不静默修正。

### R04 — 固定合并顺序

完整优先级为：具名请求字段 → 模型通用 SamplingParams → effective canonical level 的分档 map → 调用 SamplingParams → OnPayload。合并只覆盖顶层 key，嵌套对象整体替换；保留 null、0、false、数组和其他有效 JSON 值。null 是请求值，不是删除命令；空分档无覆盖，缺分档无覆盖，缺全部采样配置不增加请求字段。

Go 在请求构造边界完成一次采样应用，SimpleOptions 转换不预先物化模型默认，不复制 pi 的内部两次合并。四个公共入口必须产生等价的可观察采样结果，保持传入选项和目录不变。

### R05 — 协议范围明确

启用范围为已实现的聊天路由：OpenAI × Responses、OpenAI × Chat Completions、DeepSeek × Responses、DeepSeek × Chat Completions。采样能力由 API adapter 决定，具体路由仍由现有注册与 Binding 授权决定。Anthropic Messages、Google Gemini 忽略通用和分档采样 map；不新增 Azure、Codex、OpenRouter 或自定义 Provider 注册能力。

DeepSeek Responses 仍是 barness 扩展路线。可复用同一采样语义，但不得把模拟该协议的 pi 请求称为这条路线的原生 pi 差分通过。

### R06 — 配置边界校验

新增 map 的全部档位都校验 canonical key、对象形状和每个原始 JSON 值，包括当前未选中的档位和忽略采样的原生聊天协议。外层缺失或空 map 有效；档位对象为空有效，档位对象为 null 无效；档位内参数值为 JSON null 有效。

对消费采样 map 的 API，每档执行既有协议保留键校验，不新增厂商参数白名单或温度区间。忽略 map 的协议只校验新元数据的结构，不套用其他协议的保留键。既有通用 map 的校验范围保持不变。Client 构造和 Catalog.Hash 使用同一规则；错误包含安全的模型位置、档位和问题类型，不回显参数值。

### R07 — 授权与回调保持原边界

分档默认不能改变模型身份、路由、消息或厂商托管资源授权。OnPayload 在采样应用后按既有受信回调机制执行，完成后的请求仍经过已有授权检查。调用 map 或回调修改 reasoning 字段后，不重新选择采样档位，不增加反馈循环。

不新增 Observer 正文字段、采样值日志或厂商默认值探测。验收中的选档期待值和请求值只写入现有受控合成证据包。

### R08 — 快照与并发隔离

Client 构造快照、Catalog 的模型查询与枚举、Client 返回目录都深拷贝外层 map、每档 map 和原始 JSON bytes。每次逻辑调用固定模型与选项快照；采样组装不修改调用方数据。合法的调用后修改、返回值修改和跨租户并发不得串扰；不承诺调用方同时写入正在被读取的 Go map 是安全用法。

采样和回调在初次发送前完成一次，授权后冻结请求字节，重试复用；不在每次尝试或配置热更新时重新选档。沿用既有取消、限额、失败终态及资源释放语义。

### R09 — 目录数据与版本

新字段纳入既有确定性 Catalog.Hash；新增两层 map 的 key 插入顺序不影响哈希，参数值变化影响哈希，保留既有原始 JSON 值的哈希处理。任何实际内置目录内容变化升级 Catalog.Version，同步正式目录和价格快照。内置默认值只从核验后的 1.1.0 发布数据迁入已支持型号；不编造推荐值，不依据源码 tag 缺生成数据而假定所有型号都没有默认值。

模型列入仍遵守现有硬约束标准。本特性不自动增加型号、操作或能力；若目标数据使既有型号出现新硬约束，必须明确解决或登记阻断，不能静默忽略或悄悄删型号。

### R10 — 单一 1.1.0 兼容基线

目标冻结 pi-ai 1.1.0，commit 为 `abe508e1b89912adde45528136c3221eb69acdd7`。先证明官方 npm 发布件与固定源码的等价性、取得生成模型数据及 provenance，再更新唯一 oracle 和基线合同。不保留第二套 1.0.0 oracle，不将本特性登记为长期 1.0.0 扩展。

迁移对全部既有差分子集做增量清点，保留仍成立的既有扩展边界。新发现必须登记为 fixed、明确规范支持的 extension 或 pending；未解释差异默认 pending，不能用宽泛扩展放行。分阶段实现时允许分支上的已登记预期失败，正式切换与发布要求零 pending 和完整回归证据。其他 1.1.0 新功能只有在既有承诺的兼容迁移必需时才作最小处理，不以基线升级扩大产品范围。

### R11 — 发布和回滚以完整证据判定

同步公共契约、差异说明、兼容主规范、ADR、oracle provenance、账本、追溯和发布基线校验。新需求加入既有发布合同，不能删除旧的 P0、差分、压力、live 或业务效果要求以降低门槛。目录变化使历史 live 证据不能冒充当前目录通过，按既有九组合合同补齐。

回滚恢复完整先前版本的实现、Catalog、oracle lock、provenance 和账本，不增加内部旧路径 shim 或双写。Host 在选择旧版本时同时撤回新字段配置；本次无数据库迁移或对外 HTTP 契约变更。

### R12 — 交付文档与工单

用户文档提供完整 Host 配置示例、优先级、canonical/wire 区别、full/simple 边界和协议范围；演示数值标为合成值。实施拆成发布件核验、基线迁移、目录元数据、请求组装、隔离差分、发布文档六个依赖明确的工单，每项保留失败场景、执行命令、证据索引和未执行说明。

## Testing Decisions

用户已确认采用现有公共 Client E2E 边界：Stream、Complete、StreamSimple、CompleteSimple 及 WithHooks，经 Host 替身与受控 Provider 捕获最终请求，比较终态和请求 JSON，并产出可重放证据。无需上层 Agent 流程，不以私有 helper 单测替代集成验收。

先按验收矩阵写失败场景和 E2E，再实现对应行为；对真正新增的公开纯边界，优先在 Client/Catalog 公共入口验证。发布件重建与哈希检查是可重复来源核验，不是第二套采样业务测试。

沿用 TestOptions、TestChatOptions、TestDeepSeekChatOptions 的选项载荷场景，TestModelSamplingRejects 的保留键边界，TestCatalogSnapshot/TestCatalogInclusion 的目录隔离与字段比较，以及 TestTrustedCallbacks/TestProtocolCallbacks 的回调顺序；差分进入既有 TestPiDifferential 与各协议子集，不新增独立测试体系。

测试覆盖选档、合并、入口差异、协议边界、非法元数据、快照深拷贝、回调授权、重试冻结和多租户并发。适用路线与唯一 1.1.0 oracle 比较全部请求字段和既有结果语义；扩展路线独立断言并显式跳过原生差分。历史基线不能被当作新增用例的正确 golden。

每条新增需求映射到用例与实际 artifacts，证据保留 fixture、模型配置、请求、响应脚本、终态、断言、差分及哈希，执行既有脱敏审计。离线、差分、race、live 和发布门禁各自报告 PASS/FAIL/NOT_RUN；来源研究中的 16 条 probe 不是本特性已通过的 E2E。

完整场景和适用规则见[验收矩阵](acceptance.md)。完成特性的最低条件是全部新增公共 E2E 通过、全部既有差分子集零 pending、race 与既有完整发布合同通过。

## Out of Scope

- 按任务自动调参、性能推荐、从服务端反馈推断实际思考强度或新增采样参数白名单。
- 调用级分档 map、models.json 加载器、modelOverrides、配置继承或新 Provider 注册。
- Azure Responses、Codex Responses、OpenRouter 及未实现的协议或操作。
- 改写 full reasoning 夹取语义、Anthropic Managed Effort、Gemini budget/level、采样覆写后的 usage 推理策略。
- 自动扩充内置型号、提供缺乏冻结来源的默认值、实施全部 pi 1.1.0 新能力。
- Agent 循环、工具执行、上层 UI、数据库迁移或兼容内部旧实现的双路径。

## Further Notes

本规范中的 1.1.0 基线迁移、全档位结构校验和 Go 单次请求边界合并是本方案的设计选择；此前研究不构成实现或发布证据。本次来源与差分核验后已更新候选基线说明，保留 ADR-0004 的历史 1.0.0 事实；正式交付仍须 B01–B04 全部通过。

第一方依据：[pi 的采样 helper](https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/api/simple-options.ts#L24-L64)、[能力夹取](https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/models.ts#L1222-L1254)、[官方配置文档](https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/coding-agent/docs/models.md#L101-L125)。详细出处与已验证范围见[研究报告](../pi-1.1.0-thinking-sampling/research.md)。

已确认：用户于 2026-10-11 选择“沿用现有公共 Client E2E”。已完成公共功能与单一 oracle 实施，实际执行结果分别记录，不以离线通过替代完整发布。

已确认：用户于 2026-10-11 批准补齐聊天终态 durationMs，并允许独立运行耗时数值差异。最终字段必须为非负整数、与唯一终态相等；差分只对有限 case 的该字段数值及既有非终态共享视图时机登记扩展。具体边界见 ADR-0025 durationMs。审查基准固定为用户确认的 39ab6a5。
