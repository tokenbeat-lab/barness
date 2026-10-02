---
status: accepted
date: 2026-10-02
---

# barness-ai 用量完整性按尝试记录，价格快照以目录版本与哈希标识

spec I10 要求 Usage 与成本对齐冻结 pi，同时在 Result/观测元数据中区分未上报、部分上报、完整上报，记录价格版本，且不修改兼容消息的数字字段、不把多次尝试合计反写到消息、不把零值解释为免费。spec 没有规定完整性挂在哪里、"完整"对 Responses 意味着什么、失败响应自带的用量如何处理，以及价格版本的形态。工单 15 实现时定下以下几点，维护者于 2026-10-02 确认（决策三按维护者意见改为与 pi 一致）。

## 决策一：兼容 Usage 照搬 pi，完整性与每次尝试的用量放在 `Attempt` 上

`Usage` 增加 `cost`（始终存在）、可选的 `reasoning` 与 `cacheWrite1h`（用 `Nullable` 区分未设置与 0），数值与 pi 逐位一致：pi 只在 completed/incomplete 终态计算成本，只在报告了 usage 时写入 `reasoning`；消息初始为零值。

另在 `Attempt` 上增加 `UsageReporting`（`unreported`/`partial`/`complete`）与该次尝试自己的 `Usage`（已计价）。未取得响应的尝试（HTTP 错误、连接失败、准入后失败）为 `unreported`、用量为零值；读取流的那次尝试记录终态报告的内容（`response.failed` 除外，见决策三）。消息 Usage 等于读取流那次尝试的 Usage，从不跨尝试求和。

### Considered Options

- 在 `Usage` 内加完整性字段：会改变兼容消息的形状，差分出现 barness 独有字段，违反"不修改兼容消息"。
- 在 `CallMetadata` 上只放一个调用级完整性：无法表达重试路径上各次尝试分别的情况。
- 每次尝试记录（采用）：Result 的 `Metadata.Attempts` 与 Observer 的 AttemptFinished 记录同一份值。

### Consequences

- Responses 只在终态响应里报告 usage，失败的尝试（HTTP 层）永远是 `unreported`；目前无法出现"两次尝试都上报用量"，所以"不求和"由构造保证并由重试场景断言消息用量等于最后一次尝试。
- 宿主做成本对账应读取 `Attempts`，而不是消息 Usage。

## 决策二：Responses 的"完整"= 五个始终定义的计数都在

缺失或 null 的计数一律按 0 计（与 pi 的 `|| 0` 一致）。`complete` 要求 `input_tokens`、`output_tokens`、`total_tokens`、`input_tokens_details.cached_tokens`、`output_tokens_details.reasoning_tokens` 都存在且非 null（0 也算）；usage 存在但缺其中任何一个为 `partial`（缺失的按 pi 的 `|| 0` 读作 0，成本可能偏差）；usage 缺失或为 null 为 `unreported`。`cache_write_tokens` 不作要求：它较晚才进入 OpenAI 的 schema，许多响应没有它，缺失读作没有缓存写入。

### Considered Options

- 连 `cache_write_tokens` 也要求：现有真实响应形态几乎都会变成 `partial`，信号失去意义。
- 只要 usage 存在就算完整：缺少明细时无法区分缓存读与普通输入，成本会被高估却被标为完整。

## 决策三：`response.failed` 携带的用量与 pi 一样丢弃

pi 不接收失败响应的 usage，消息 Usage 保持零值。首期与 pi 一致：该尝试同样不记录这份用量，标为 `unreported`（而不是零值的 `complete`），所以至少不会被读成"已完整上报且免费"。

### Considered Options

- 计价后记在尝试上（实现初稿）：Provider 已计费的失败请求在元数据中可见，但超出 pi 行为。维护者决定先与 pi 一致，后续再优化，见工单 25。

### Consequences

- Provider 对失败响应收取的费用在 barness 元数据中不可见；宿主若需对账，只能依据厂商账单。工单 25 落地后移除此限制。

## 决策四：价格版本 = 目录版本 + 内容哈希，挂在 `CallMetadata`

`Model` 增加 `Cost`（与 pi 模型数据同形，含 tiers）。`Catalog.Hash()` 为目录 JSON 编码的 SHA-256（`sha256:<hex>`），Client 构造时计算一次；调用解析后 `CallMetadata.CatalogVersion`/`CatalogHash` 标出计价所用的快照。同一调用的各次尝试共用同一快照，因此放在调用级，Observer 的尝试记录通过其 `Call` 同样可见。

内置目录的版本与哈希固定在 E2E fixture 中：内容改变而版本未改会使测试失败。构造 Client 时拒绝负数或非有限的价格（`ConfigError{Field: "Catalog"}`），否则 NaN 成本会使 Result 无法 JSON 编码。

### Consequences

- 内置目录升为 `2026-10-02.1`：补齐所有模型价格，并加入 `gpt-5.5-pro`（唯一可列出且带阶梯价格的模型；pi 中其他带 tiers 的模型需要 additional_tools 或 tool search）。

## 决策五：成本规则逐位移植，1h 写入以隔离测试对照 pi

`calculateCost` 的运算顺序原样保留，每个乘积单独舍入（Go 允许把乘加融合为 FMA，可能与 JavaScript 末位不同）。Responses 的 service tier 调整按 pi：`response.service_tier ?? options.serviceTier`，flex ×0.5、priority ×2（gpt-5.5 为 ×2.5）；与 pi 一样依据调用选项而非 onPayload 改写后的请求体。

Responses 从不报告 1h 缓存写入，也没有多档 tiers 的模型，这两条规则由 `TestCostEstimate` 直接对照冻结 pi 的 `calculateCost`（`BARNESS_AI_PIDIFF=1`），同一测试还核对内置价格与 pi 模型数据一致。Anthropic adapter（工单 16）接入后应由 E2E 覆盖 1h 写入，届时这一隔离测试中对应用例可删除。

## 维护者决定（2026-10-02）

1. 决策二：采纳；没有的计数（如 `cache_write_tokens`）计为 0，且不影响 `complete`。
2. 决策三：失败响应的用量先与 pi 一样丢弃，记录到尝试上的优化见工单 25。
3. 决策四：采纳；价格快照放在调用级。
