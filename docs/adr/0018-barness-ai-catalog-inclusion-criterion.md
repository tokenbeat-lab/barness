---
status: accepted
date: 2026-10-03
---

# barness-ai 内置目录的列入标准：遵守模型硬约束即可列入，可选协议特性不阻塞

ADR-0011 决策五、ADR-0013 决策二沿用工单 16 实现时定下的目录原则："compat 需要 barness 未实现行为的模型不列入"（`ai/catalog.go` 的 `BuiltinCatalog` 注释）。这条原则把 pi `compat` 中两类性质不同的开关混为一谈，结果是：Messages 与 Responses 协议本身能服务的模型，因为某项可选的 beta 优化未实现而整个不在目录中。维护者于 2026-10-03 决定按本 ADR 的标准重新划定。

## 决策一：按开关性质区分硬约束与可选特性

pi 模型数据中的 `compat` 开关分两类：

- **硬约束**：不遵守时请求会被厂商拒绝，或在可达场景中失败。例如 `supportsTemperature: false`、`forceAdaptiveThinking`、level map 中为 null 的等级。目录必须携带它们，adapter 必须遵守。
- **可选特性**：开启后是协议层的额外能力，通常靠 beta 实现；不开启时 pi 自己也有完整的回退路径，请求仍然有效，只失去缓存、可用性等方面的收益。

判断依据是冻结 pi 的实现：如果 pi 在开关关闭或条件不满足时会改用另一种合法的请求形状，该开关就是可选特性。

**列入标准**：模型所在的 Operation × Provider × API 已经实现，adapter 能遵守该模型的全部硬约束，就列入内置目录。可选特性未实现时，目录照常列入该模型，按 pi 中该特性关闭时的行为发送请求。

协议确实服务不了的模型仍不列入，例如 ADR-0012 决策八排除的 Interactions API、Live API、图片输出模型，以及只在 Responses 上提供的 pro 模型不进 Chat 目录（ADR-0013 决策二）。它们本来就符合本标准，不受影响。

## 决策二：现有被排除开关的归类

| 开关 | 协议 | 归类 | 依据（pi-ai 0.87.1） |
| --- | --- | --- | --- |
| `supportsMidConvoToolChanges` | Anthropic | 可选特性 | `buildParams` 在开关关闭、没有初始工具或存在同名工具重定义时，改发 `getCurrentTools` 的当前工具列表，不带 beta；barness 现有 adapter 就是这条路径 |
| `allowedFallbackModels` | Anthropic | 可选特性 | 不发送 `fallbacks` 时请求有效，只是主模型不可用时不会自动切换；工单 28 的授权过滤也可能把列表滤空 |
| `supportsAdditionalTools`、`supportsToolSearch` | Responses | 可选特性 | `resolveTranscriptTools` 在两者都关闭或存在非追加的工具变更时，改发当前工具列表，不锚定追加的工具 |
| `supportsMidConvoEffort` | Anthropic | **硬约束** | `buildParams` 对这类模型恒用 adaptive thinking 并带 `block_binding: {prefix_mismatch_behavior: drop_block}`，pi 注释说明：这样做是为了让前缀不匹配时丢弃 thinking 块，"instead of surfacing as persistent 400 responses"；同时从不发送 temperature（claude-fable-5-1 的数据没有 `supportsTemperature: false`，靠这一分支避免发送） |

`supportsMidConvoEffort` 的失败场景是可达的：barness 每次调用独立，host 可以在同一会话的不同轮次使用不同 effort 或 thinking 设置。所以开启它的 claude-fable-5-1、claude-opus-5、claude-opus-5-5 仍要等工单 27 实现后才能列入。

## 决策三：本次纳入的模型

- Anthropic × Messages：claude-opus-4-8、claude-fable-5。两者的硬约束（adaptive thinking、opus-4-8 的 `supportsTemperature: false`、level map）都已被现有 `ModelCompat` 覆盖。
- OpenAI × Responses：gpt-5.4、gpt-5.4-mini、gpt-5.4-pro、gpt-5.5、gpt-5.6-luna、gpt-5.6-sol、gpt-5.6-terra、gpt-6-astra、gpt-6-luna、gpt-6-sol。`supportsMidConvoSystemMessages` 与 `supportsExplicitPromptCacheMode` 已实现，阶梯价格已实现（ADR-0010）。
- OpenAI × Chat Completions：上一行中除 gpt-5.4-pro 以外的模型，沿用 ADR-0013 决策二"pro 模型只在 Responses 上提供"的规则。

实现、差分登记与真实冒烟见工单 34。

## 决策四：未实现的可选特性在差分中登记为扩展

差分运行器按 pi 自己的模型数据生成请求，所以 pi 对上述模型会发出 barness 不发的内容（占位工具与 tool-changes beta、`fallbacks` 与 server-side-fallback beta、`additional_tools`/`tool_search_*` 条目）。这些差异按本 ADR 登记为 `extension`，`cases` 只限于覆盖这些模型的用例，避免对其他场景放宽。可选特性以后实现时，删除对应登记。

## Considered Options

- 维持原原则，等工单 26–28 全部完成再列入：可选优化成了列入门槛，Provider 的协议本来能服务的模型被长期挡在目录外。
- 把所有开关都视为可选，五个 Anthropic 模型全部列入：`supportsMidConvoEffort` 关系到可达的 400 失败，列入就等于在目录里承诺了做不到的行为。
- 按开关性质区分（采用）。

## Consequences

- 目录的含义从"adapter 实现了 pi 对该模型的全部行为"变为"adapter 能为该模型构造有效请求"。可选特性是否实现，由 `docs/barness-ai/differences.md` 与差分账本说明，不再由模型是否在目录中来表达。
- 工单 26、28 改为不阻塞列入的优化工单；工单 27 仍是三款托管强度模型的列入条件，且不再依赖 26。
- 新模型以后按本标准判定：硬约束能用现有 `ModelCompat` 表达时，只需加一条数据；出现新的硬约束时才需要改 adapter。
- pi OpenAI 数据中另有 17 个模型（gpt-4o 系列、gpt-4-turbo、gpt-4.1-nano、o1、o1-pro、o3-pro、gpt-5.x-chat-latest、gpt-5.2-pro、gpt-5.3-codex 系列、gpt-5.4-nano、gpt-realtime-2.1）从未按任何原则评估过，只是早期没有选入。它们是否列入需按本标准逐个核对硬约束，不在本 ADR 范围内。

## 取代关系

取代 ADR-0011 决策五中"不列入需要原生中途工具变更、托管中途 effort 或服务端回退模型的模型"一句，以及维护者决定第 5 条中的工单依赖顺序；该决策中已列入的模型与 `ModelCompat` 字段不变。ADR-0012 决策八、ADR-0013 决策二维持不变。

## 实现（工单 34，2026-10-03）

- 目录升为 `2026-10-03.1`，列入决策三的 12 个模型（Responses 10 个、Messages 2 个），Chat 上另有 9 个条目。新增的 compat 只有已有的 `SupportsMidConvoSystemMessages`、`SupportsExplicitPromptCacheMode`（gpt-5.6/gpt-6 系列）与 opus-4-8 的 `SupportsTemperature=false`；`ModelCompat` 没有新字段。
- Chat 的排除规则从硬编码两个 ID 改为按 pi 数据中的模型 ID 判定：ID 以 `-pro` 结尾的 OpenAI 模型只在 Responses 上提供（gpt-5-pro、gpt-5.4-pro、gpt-5.5-pro）。
- 全部内置模型的名称、reasoning、输入、上下文、输出上限、level map、所携带的 compat 与价格，由 `TestCatalogInclusion/builtin-models-are-pi's` 在差分开启时与冻结 pi 的模型数据逐字段比较（runner 的 `models` 入口），不再靠人工核对。
- 差分新增 9 条 `extension` 登记（Anthropic 5 条、Responses 4 条），`cases` 限于 `catalog.json` 中带工具或 fable-5 的用例。Chat 与不涉及可选特性的用例与 pi 一致。
- 真实冒烟（2026-10-03，提交 5bb12f4，账户 prod-anthropic、prod-openai）：三个组合全部通过，没有环境重试。
  - `tool-changes-*` 在 claude-opus-4-8、claude-fable-5 上通过，gpt-5.4、gpt-6-sol 在 Responses 和 Chat 上都通过。流程是：多轮工具调用，中途移除 get_time、加入 convert_temperature，模型调用新加入的工具后作答。厂商接受了不带可选特性的请求：当前工具列表、原地的 system 消息，没有占位工具、beta、`fallbacks` 或 `additional_tools`。
  - gpt-5.4-pro 在 Chat 上被拒绝，HTTP 404 "This is not a chat model and thus not supported in the v1/chat/completions endpoint"，确认了决策三的排除。
  - 已列入的模型都留在目录中，结果已合并进支持矩阵。

## 实现（工单 27，2026-10-03）

托管强度按 ADR-0019 实现后，claude-fable-5-1、claude-opus-5、claude-opus-5-5 列入目录 `2026-10-03.2`，携带 `SupportsMidConvoEffort`、`SupportsMidConvoSystemMessages`、adaptive 与 level map，opus-5、opus-5-5 另带 `SupportsTemperature=false`。可选特性 `supportsMidConvoToolChanges` 仍不携带。它在差分中的扩展登记按决策四处理，限于工单 27 带工具的用例。真实冒烟（2026-10-03）通过：厂商接受了三款模型对话中途从 high 到 low 的强度变更，详见 ADR-0019。

## 1.0.0 基线迁移（工单 03，2026-10-08）

原有目录的已携带字段与 1.0.0 数据完全相同；模型数据文件哈希因 schema 6、
非聊天类型与新增型号变化。目录升为 `2026-10-08.1`，新增四种聊天型号：

| 型号 | 硬约束与现有实现 | 未实现可选特性的处理 |
| --- | --- | --- |
| gpt-6.1-sol（Responses / Chat） | 完整 level map、上下文预算、阶梯价格；原有中途 system 与显式缓存模式 | grammar 无声明入口；additional_tools / tool search 关闭时用当前工具列表 |
| gpt-daybreak-blue-latest、gpt-daybreak-red-latest（Responses / Chat） | 完整 level map、显式缓存模式；只加数据，无新 adapter 分支 | grammar 无声明入口 |
| claude-sonnet-5-5（Messages） | ADR-0019 已实现托管 adaptive effort、block_binding、逐轮强度；忽略 temperature；minimal 被 clamp 为 low | 不启用原生工具变更 beta，使用当前工具列表 |

上述 fallback 是已有批准的领域规则。未扩大差分账本中任何 extension 的 cases；
新增型号的 full/simple 无工具控制场景与全部内建型号字段参加 1.0.0 差分。
本次不追补 0.87.1 时未评估的 17 个 OpenAI 型号（前文 Consequences），也不把
非聊天类型、Interactions、Live 或需要托管 computer use 的 Google 型号放进聊天目录。
新增支持不改变 Binding 白名单，宿主仍需逐个授权。

## 多类型目录（工单 04，2026-10-08）

ADR-0020 将包含准则的维度补为操作 × Provider × 协议。聊天目录仍只含聊天型号，
能接收图片不等于提供图像生成授权；图像、分类目录用各自的能力与价格描述，完整身份可跨操作并存。
目录类型可供宿主自带配置，内置条目必须等实际 adapter 与该路线自己的验收证据齐备后才纳入。
工单 04 不把尚未交付的图像或分类型号写入内置目录，Google Interactions 型号仍不进聊天目录。
内置目录版本升为 `2026-10-08.2`，聊天条目的字段、价格和六条路线保持不变。

## 工单 08：首个内置分类型号（2026-10-08）

TypeSafe adapter 已实现三种问题及全部硬约束，并取得自己的官方 live 证据后，
目录 2026-10-08.3 仅纳入 jev-1.13.0。能力：choice 上限 255、score 2–10 个有序等级、
bool；上下文 64000（另有 state + 最长问题 32000 的厂商约束，Host 不本地计 token）。
名称 Jev 保持与冻结 pi 的共同字段一致；固定 ID 和官方输入费率是已登记扩展。
jev-latest / jev-preview 不列入。公共目录测试与官方来源 fixture 对照，
自己的真实混合/单选/上下文错误 fixture 经 Client 回放；聊天条目与授权白名单不变。

## 工单 09：原生图像生成（2026-10-08）

原生 OpenAI Images 明确豁免冻结 pi 的聊天目录字段比对，以自己的扩展路线 fixture、能力/费率来源和 live 证据验收。TestCatalogNativeImageExemptions 断言登记。工单 09 仅交付宿主自带目录的生成协议；内置图像型号仍由工单 11 纳入，未提前宣称首批型号已支持。

详见 [ADR-0022](0022-barness-ai-openai-images-unary.md)。

## 工单 11：首个内置图像型号（2026-10-08）

原生 OpenAI Images 的 JSON 编辑是纳入前必须通过的自身 live gate。固定快照
`gpt-image-2.5-sunburst-2026-09-08` 在本组合的公共 Client 上通过 JSON 编辑、生成、
mask 后纳入目录 `2026-10-08.4`。仅保留验证的一张参考图/输出、1024×1024、low/medium
质量、mask 和透明背景；PNG/JPEG/WebP 均实测。官方允许更大数量、分辨率及其他质量，
本次不扩大声明；未来扩大须以独立来源/本组合证据换目录版本。moving alias 不纳入。
输入保真度的 Sunburst 允许值仍未明确且未测试，空能力列表要求省略，不误标为
厂商 UNSUPPORTED。官方分模态费率、抓取日期、完整真实 usage 与可重放 fixture
验证价格快照；直接 Images API 不含缓存输入费率或按张派生附加费。

详见 ADR-0022 与 [工单 11 证据](../../.scratch/barness-ai-pi-1.0/openai-images-live-evidence/README.md)。

## 工单 12：Google 图像生成协议（2026-10-08）

Google × Interactions 图像登记为 P09 扩展与原生目录比对豁免；离线场景明确断言 routes
与全部 fixture 的 pi 差分跳过。此次仅服务宿主自带目录，内置目录版本仍为 2026-10-08.4，
不把协议实现或 Google 聊天的 live 通过当成图像已支持。工单 14 自己的模态/能力/价格与
真实冒烟验收是首次包含前提；参考图由工单 13 交付。
