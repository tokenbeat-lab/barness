# GLM / MiniMax 中国站与国际站：站点配置还是独立 Provider

研究日期：2026-10-20（Asia/Shanghai）。本地代码基线：`68dec5af86ad7f2f0c2dc1188a0ca046dc948780`。
本文是方案研究与实施建议，不是已接受的 ADR，也不表示新增站点已经接入或通过正式验收。
仅阅读代码和公开一手资料；未读取凭据、调用付费推理 API 或改变运行时配置。

## 1. 建议

**推荐方案 2：将中国站和国际站作为不同 Provider；复用经过核验的协议实现，各自维护 Model 目录。**

理由不是“多一个 Provider 字符串比较方便”，而是 barness 的 Provider 已经表示实际模型服务契约，Model 身份已经包含 Provider。地区服务有独立账户、认证、价格或型号发布节奏时，直接使用现有身份维度即可表达差异。项目术语表已经明确规定：认证、账户或定价契约独立的地区服务须区分身份；同品牌不能互授原生续接权限和支持证明。[术语表](../../GLOSSARY.md)、[ADR-0001](../../docs/adr/0001-barness-ai-protocol-boundary.md)、[ADR-0020](../../docs/adr/0020-barness-ai-model-operations-and-binding-authorization.md)

用户希望“单独配置 Model”是决定性需求。同一个 wire 型号名，在两个站点需要能分别声明：是否存在、支持的协议、上下文与输出限制、思考档位、采样默认、兼容约束和价格。当前这些元数据属于 Model，Binding 只对目录内的型号授权；独立 Provider 自然形成两条不同的目录身份。[Model](../../ai/catalog_models.go)、[Binding](../../ai/binding.go)、[目录索引](../../ai/catalog_query.go)

方案 1 也可实现，但有两种不同成本：

- **轻量版**：只在 Binding 增加 Site / Region，切换 endpoint、account、credential。可用于连接配置，但不足以在同一 Client 中表达同名型号的不同元数据。
- **完整版**：Site 进入 Model 身份、查询、授权、原生历史来源、调用归属和支持证明。可以达到方案 2 的能力，但需要贯穿全链路的新维度。对目前这两个厂商而言，没有充分收益抵消这项成本。

以上是基于代码边界的设计推论，不是厂商对客户端内部架构的要求。

## 2. 厂商一手资料核验

公开厂商资料核验时间：2026-10-20 08:33–08:44（Asia/Shanghai）。动态价格和型号信息仅表示本次读取时的公开声明。下列 SDK base URL 不代表所有账户都具有调用资格。

### 2.1 服务入口

| 服务 | 通用 Chat Completions SDK base URL | 其他公开协议 SDK base URL | 直接依据 |
| --- | --- | --- | --- |
| 智谱 BigModel 中国站 | `https://open.bigmodel.cn/api/paas/v4` | Responses `https://open.bigmodel.cn/api/v1`；Messages `https://open.bigmodel.cn/api/anthropic` | [中国站 API Key 分类说明](https://docs.bigmodel.cn/cn/coding-plan/notice/api-key-management)、[Responses](https://docs.bigmodel.cn/cn/guide/develop/responses/introduction)、[Messages](https://docs.bigmodel.cn/cn/guide/develop/claude/introduction) |
| Z.AI 国际站 | `https://api.z.ai/api/paas/v4` | Responses `https://api.z.ai/api/v1`；Messages `https://api.z.ai/api/anthropic` | [API 介绍](https://docs.z.ai/api-reference/introduction)、[API Key Management Guide](https://docs.z.ai/devpack/notice/api-key-management) |
| MiniMax 中国站 | `https://api.minimax.cn/v1` | Responses 同 `/v1` base；Messages `https://api.minimax.cn/anthropic` | [OpenAI SDK](https://platform.minimax.cn/docs/api-reference/text-openai-api)、[Responses OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-responses.json)、[Anthropic SDK](https://platform.minimax.cn/docs/api-reference/text-anthropic-api) |
| MiniMax 国际站 | `https://api.minimax.io/v1` | Responses 同 `/v1` base；Messages `https://api.minimax.io/anthropic` | [OpenAI SDK](https://platform.minimax.io/docs/api-reference/text-openai-api)、[Responses OpenAPI](https://platform.minimax.io/docs/api-reference/text/api/openapi-responses.json)、[Anthropic SDK](https://platform.minimax.io/docs/api-reference/text-anthropic-api) |

这些是 SDK base URL，不应把 `/anthropic` 与最终 `/anthropic/v1/messages` 混为一谈。中国站 MiniMax 的索引仍出现 `platform.minimaxi.com` 链接，但当日 SDK 文档给出的调用主机是 `api.minimax.cn`；不要把历史域名自动演绎为另一独立市场 Provider。[CN 文档索引](https://platform.minimax.cn/docs/llms.txt)、[CN SDK](https://platform.minimax.cn/docs/api-reference/text-openai-api)

### 2.2 Model 与价格必须允许独立变化

**GLM 有已经发生的计价结构差异，不能靠统一汇率解决。** 中国站 GLM-5.1 公开输入/输出价格按输入 `<32K` 与 `≥32K` 分为 `6/24`、`8/28` 元/百万 tokens；国际站 GLM-5.1 当日表列为 `$1.4/$4.4` 每百万 tokens，未在该表列同样的两档。中国站公开价格表包含 GLM-5-Turbo，国际站价格页当前没有该条目。**这只说明所核验价格表的覆盖范围不同，不能据此断言国际站一定不可调用。** 国际站折叠的 Other Pricing 中也列有 GLM-4.7-Flash，不能将其误列为中国站独有。 [中国定价](https://docs.bigmodel.cn/cn/guide/start/pricing)、[国际定价](https://docs.z.ai/guides/overview/pricing)

MiniMax 两站当前文档都列出 `MiniMax-M3.1-Flash` 和 `MiniMax-M3`，并且都声明 Chat Completions 支持文本/图片/视频；这为复用协议实现提供依据，却不证明每个账户有同样资格、实际输出或限额。[中国 SDK](https://platform.minimax.cn/docs/api-reference/text-openai-api)、[国际 SDK](https://platform.minimax.io/docs/api-reference/text-openai-api)

| 型号/档位 | 中国站 CNY/百万 tokens：输入、输出、缓存读 | 国际站 USD/百万 tokens：输入、输出、缓存读 |
| --- | --- | --- |
| MiniMax-M3.1-Flash | `1.00 / 4.00 / 0.10` | `0.15 / 0.60 / 0.015` |
| MiniMax-M3 standard，≤512k 输入 | `2.10 / 8.40 / 0.42` | `0.30 / 1.20 / 0.06` |

上表直接使用各站价格，不推导固定汇率或跨站价格相等。M3 另有长输入及 priority 档位，此处只为架构比较保留短输入 standard 示例。[中国定价](https://platform.minimax.cn/docs/pricing/overview)、[国际定价](https://platform.minimax.io/docs/pricing/overview)

**与仓库 2026-10-18 历史研究的差异：** 当日新核验文档已使用 `MiniMax-M3.1-Flash`（没有 `-Preview`）并发布按量价格，不能照搬旧研究“M3.1 仅 Preview 且缺价”的结论；也不能反向把新名称、价格或资格写进旧证据。这说明 Model 数据还必须按时间维护版本；某日期、某市场的资料不能替代另一份目录的依据。[当前型号说明](https://platform.minimax.io/docs/guides/models-intro)、[当前价格](https://platform.minimax.io/docs/pricing/overview)

### 2.3 凭据有地区和用途两个不同边界

MiniMax 中国站 M Plan FAQ 的 401 排查明确要求将 CLI `region` 设为 `cn`（国内 API Key）或 `global`（海外 API Key），并确认区域与 Key 来源平台一致。这证明官方接入流程承认凭据所属市场；不能无声跨区域重用凭据。文档没有在此保证物理数据驻留位置。[中国站 M Plan FAQ：登录后仍报 401](https://platform.minimax.cn/docs/m-plan/faq)

MiniMax 两站均区分普通按量 API Key 与 Subscription Key：前者扣账户余额；后者使用 M Plan 额度/积分包，不能混用；生产场景官方建议按量付费。**这属于同一市场内的计费用途，不是“中国/国际”这个维度。** [国际 M Plan FAQ](https://platform.minimax.io/docs/m-plan/faq)、[中国 M Plan FAQ](https://platform.minimax.cn/docs/m-plan/faq)

GLM 两站均在 **2026-10-16** 发布 Key 分类说明。新建标准 Key 与 Coding Plan Key 的计费用途由 Key 类型决定，在支持的端点间切换地址不改变类型；旧 Compatible Key 仍依赖端点与账户是否曾有 Coding Plan：通用 Chat 标准端点走余额/普通资源包，而 coding Chat、通用 Messages/Responses 可随账户历史走套餐。**因此不能再笼统以 URL 或协议推断是否会扣 Coding Plan。** [中国 Key 分类](https://docs.bigmodel.cn/cn/coding-plan/notice/api-key-management)、[国际 Key 分类](https://docs.z.ai/devpack/notice/api-key-management)

### 2.4 必须保留的未知

本次所读官方来源没有确认 GLM 两站或 MiniMax 两站的登录账户、余额、套餐额度、资源 ID 或原生续接状态可以跨站通用/转移；也没有取得覆盖所有 key 类别的“绝不互通”承诺。准确结论是：**没有跨站互通保证；barness 应独立授权，默认不跨站复用。** 此为工程推论，不是伪造厂商声明。各站提供自己的 Key/余额入口，MiniMax 另外明确 Key 来源平台应与 region 一致。[Z.AI 认证](https://docs.z.ai/api-reference/introduction)、[BigModel Key](https://docs.bigmodel.cn/cn/coding-plan/notice/api-key-management)、[MiniMax 国际 Key](https://platform.minimax.io/docs/faq/about-apis)、[MiniMax 中国 region](https://platform.minimax.cn/docs/m-plan/faq)

本次没有验证真实账户资格、实时余额、服务物理驻留位置、延迟、输出质量或跨区续接可用性。同名型号、相似 schema、同一品牌均不足以证明这些结论。

### 2.5 上游做法：仅作工程参照

pi 当日 main 固定提交 `73f8de80bc87021375f8b525292ac07492253002` 的 KnownProvider 有 `minimax` 与 `minimax-cn`，还有 `zai` 与 `zai-coding-cn`；其 env 映射分别为 `MINIMAX_API_KEY`、`MINIMAX_CN_API_KEY`、`ZAI_API_KEY`、`ZAI_CODING_CN_API_KEY`。这为方式 2 提供真实上游先例，**但 `zai-coding-cn` 不能冒充通用中国标准 API，pi 的 Provider 划分不是厂商契约。** [pi 类型 L106–110](https://github.com/earendil-works/pi/blob/73f8de80bc87021375f8b525292ac07492253002/packages/ai/src/types.ts#L106)、[pi Key 映射 L102–106](https://github.com/earendil-works/pi/blob/73f8de80bc87021375f8b525292ac07492253002/packages/ai/src/env-api-keys.ts#L102)

models.dev 当日 dev 固定提交 `85afece4ffa4427326dd52abe4a1fc618b7cc04e` 维护 `zai`/`zhipuai` 和 `minimax`/`minimax-cn` 四份 Provider 配置，各有自己的 API 地址和型号目录。这是方案 2 的另一个先例。其某些地区 env 名相同、部分 doc 链接跨站，因此只参考独立身份/目录结构，不能整份照抄。 [zai](https://github.com/anomalyco/models.dev/blob/85afece4ffa4427326dd52abe4a1fc618b7cc04e/providers/zai/provider.toml)、[zhipuai](https://github.com/anomalyco/models.dev/blob/85afece4ffa4427326dd52abe4a1fc618b7cc04e/providers/zhipuai/provider.toml)、[minimax](https://github.com/anomalyco/models.dev/blob/85afece4ffa4427326dd52abe4a1fc618b7cc04e/providers/minimax/provider.toml)、[minimax-cn](https://github.com/anomalyco/models.dev/blob/85afece4ffa4427326dd52abe4a1fc618b7cc04e/providers/minimax-cn/provider.toml)

## 3. barness 已有边界

| 现状 | 对本问题的含义 | 代码 / 决策来源 |
| --- | --- | --- |
| `ProviderZhipu = "zhipu"` 明确指中国区；`ProviderMiniMax = "minimax"` 明确指国际站 | 现有 ID 已有确定地区语义，不能随着 URL 改变而重新解释 | [binding.go](../../ai/binding.go)，50–55 行 |
| 完整型号身份是 `(Operation, Provider, API, ID)` | 不同 Provider 可以使用相同 wire ID，仍各有元数据；同一完整身份重复会被拒绝 | [catalog_validate.go](../../ai/catalog_validate.go)，9–28 行；[ADR-0020](../../docs/adr/0020-barness-ai-model-operations-and-binding-authorization.md) |
| Binding 固定操作、Provider、协议、端点、账户、凭据引用及 AllowedModels | 多个账户、部署与使用资格已经有承载位置，无须每个账户都新增 Provider | [binding.go](../../ai/binding.go)，88–131 行 |
| 调用按 Binding 的 Provider / API 和请求 ModelID 查目录，先核对白名单再读凭据 | 两份不同 Binding 可以收窄可用型号，但不能重定义同一目录身份的参数与价格 | [call_runtime.go](../../ai/call_runtime.go)，70–100 行 |
| Model 拥有能力、Compat、默认思考、采样和 Cost；Catalog 是整个 Client 的快照 | 各站独立 Model 可以一起放在同一 Catalog / Client；无需为每站拆 Client | [catalog_models.go](../../ai/catalog_models.go)、[client.go](../../ai/client.go)、[catalog_snapshot.go](../../ai/catalog_snapshot.go) |
| Provider 能力决定自动字段和请求形状，Model Compat 只能在相应边界内限制 | 即使同协议，也不能给新地区直接套原生 OpenAI / Anthropic 默认值 | [responses_capabilities.go](../../ai/responses_capabilities.go)、[chat_compat.go](../../ai/chat_compat.go)、[anthropic_capabilities.go](../../ai/anthropic_capabilities.go) |
| 原生续接检查 Tenant、AccountScope、Provider、API、Model | 使用独立 Provider 可直接触发跨身份降级，保留可见内容和工具配对；不把同名模型当同一原生来源 | [native.go](../../ai/native.go)，14–25、98–104、130–147 行；[ADR-0001](../../docs/adr/0001-barness-ai-protocol-boundary.md) |
| CallAttribution 已带 Provider / API / Model / AccountScope | 方案 2 可使用现有归属字段区分站点，方案 1 若要求同等可观测性则需传播 Site | [result.go](../../ai/result.go)，5–25 行 |
| MiniMax 曾从中国区切换国际站，原 CN 原生前驱、价格、真实证明不能晋升国际站 | 仓库已有地区服务独立的实际案例，不能把新增 CN 支持实现成对旧证明的重新命名 | [ADR-0042](../../docs/adr/0042-barness-ai-minimax-international-target.md) |

**一个具体限制**：给两个 `zhipu` Binding 配不同 Endpoint 与 AllowedModels，可以控制某站展示和调用哪些型号；但同为 `chat / zhipu / openai-completions / 同名型号` 时，目录仍只有一份 `ContextWindow / ThinkingLevelMap / SamplingParams / Cost`。同名型号在各站不同，轻量 Site 无法解决。上述结论来自目录唯一性与查找逻辑，而不是假设两站今天一定存在每一种差异。[目录唯一性](../../ai/catalog_validate.go)、[调用查找](../../ai/call_runtime.go)

## 4. 方案 1：增加站点配置

### 4.1 只让站点影响连接

示意：`Binding{Provider: "glm", Site: "cn", Endpoint: ..., AccountScope: ..., CredentialRef: ...}`。这是提议形状，当前 Binding 没有 Site 字段。

可行之处：连接、账户和凭据的配置入口直观；若站点只是同一服务契约下的网络入口，型号事实相同且资源作用域明确，站点预设可以让用户少填 URL。已有 Binding 本身已经可以分别保存这些连接信息。[Binding](../../ai/binding.go)

不满足之处：Site 若不进入目录选择，两个站点仍共用同一 Model 元数据。全局 `Provider.Site` 更会把选择变成 Client 级状态，不适合共享 Client 中同时调用两站。把 Site 放到每个 Binding 可以避免全局切换，但仍不能改变目录键。[Client](../../ai/client.go)、[目录键](../../ai/catalog_validate.go)、[调用解析](../../ai/call_runtime.go)

可用但不推荐的绕路：Host 每站创建独立 Client，各自装配 Catalog，可以分别配置同名 Model；代价是 Host 负责分派 Client、整合目录和归属，资源与观测装配也需协调。给本地型号 ID 人工追加 `-cn` / `-intl` 则不能直接解决：当前 `Model.ID` 就是发给厂商的 `model` 字段，并没有独立的本地别名 / wire ID 映射。[Config.Catalog](../../ai/client.go)、[Chat 请求](../../ai/chat_request.go)、[Responses 请求](../../ai/responses_request.go)、[Messages 请求](../../ai/anthropic_request.go)

### 4.2 让站点成为完整身份维度

可以设计成 `(Operation, Provider, Site, API, ModelID)`，或者 Provider 下有多个 site catalog；两者都必须保证 Site 在所有边界中参与唯一识别，而不是只在 UI 中出现。

需要同时调整的面包括：

1. Model 数据、目录去重、Lookup / ModelsOf、快照 Hash 与版本；
2. Binding 解析、目录交集与协议能力选择；
3. AssistantMessage、NativeStateEnvelope、持久化回读和跨站降级规则；
4. CallAttribution、Hooks / Observer 及诊断元数据；
5. 支持矩阵、原生 fixture、真实证明身份与发布追溯。

上述范围由现有四维身份的使用位置推导，见[目录](../../ai/catalog_query.go)、[解析](../../ai/call_runtime.go)、[原生状态](../../ai/native.go)、[消息](../../ai/message.go)、[结果](../../ai/result.go)、[回调](../../ai/hooks.go)、[GLM 证明消费](../../ai/internal/testkit/supportmatrix/glm.go)、[MiniMax 证明消费](../../ai/internal/testkit/supportmatrix/minimax.go)。采用此方案还需正式修订 ADR-0020，处理旧持久化数据缺少 Site 时的明确归属；空 Site 不能解释成“任意站点”。

**适用条件**：未来存在大量同契约部署，且产品确实需要分别查询“厂商”和“部署站点”时，新增维度可能值得投入。即使如此，独立计费市场也不应仅凭地理名称推定同一服务契约。当前 GLM / MiniMax 两个实例不足以证明必须引入这套体系。

## 5. 方案 2：不同站点作为不同 Provider

### 5.1 身份建议

在保持现有明确含义的前提下扩充，无须为了命名对称重命名已用 ID：

| 服务 | 建议 Provider ID | 状态 | 显示名称建议 |
| --- | --- | --- | --- |
| GLM 中国标准 API | `zhipu` | 现有，保留原语义 | 智谱 / 中国站 |
| GLM 国际标准 API | `zai` | 拟新增，名称建议 | Z.AI / 国际站 |
| MiniMax 国际 API | `minimax` | 现有，保留原语义 | MiniMax / 国际站 |
| MiniMax 中国 API | `minimax-cn` | 拟新增为正式独立服务 | MiniMax / 中国站 |

`ProviderZAI`、`ProviderMiniMaxCN` 是可选 Go 常量名；本研究未添加它们。`minimax-cn` 即使在历史研究中出现过，也须作为独立新接入取得自身证明，不恢复旧失败捕获的支持资格。[当前常量](../../ai/binding.go)、[ADR-0042](../../docs/adr/0042-barness-ai-minimax-international-target.md)

各 Provider 的 ModelID 保留厂商实际要求的值。即使两站存在同名型号，下面两条**示意身份**也互不冲突：

```text
chat / zhipu / openai-completions / <厂商实际型号 ID>
chat / zai   / openai-completions / <同名或不同型号 ID>
```

两条记录各自声明完整元数据，不建立“复制中国站后覆盖少数字段”的隐式继承。独立记录避免某站更新时悄悄改变另一站；已有 GLM 目录也明确注明，型号与协议的数据重复是为独立演进而保留。[catalog_glm.go](../../ai/catalog_glm.go)、[Model](../../ai/catalog_models.go)

### 5.2 职责和数据流

```text
Host 按租户和业务权限解析 Binding
  → 固定 Provider + API + endpoint + account + credential reference
  → 在 Catalog 中定位 Operation + Provider + API + ModelID
  → 核对 AllowedModels 与型号约束
  → 取得并核对版本化 Credential Snapshot
  → 对应协议 adapter，应用该 Provider 自己已核验的协议行为
  → Result / Event / Usage 保留实际 Provider 与 AccountScope
```

这是对现有调用边界的沿用；不需要在公开 Target 中增加 Site，也不改变 Provider 与 API 分离的关系。[Target / Binding](../../ai/binding.go)、[call_runtime.go](../../ai/call_runtime.go)、[ADR-0001](../../docs/adr/0001-barness-ai-protocol-boundary.md)

推荐分工：Provider 表示独立服务；API 表示协议；Model 表示该服务与协议中的型号事实；Binding 表示某租户实际获准使用的连接、账户与型号范围。品牌分组可以在 Host 展示层用静态数据表达，例如把 `zhipu` / `zai` 同组，不必先引入运行时 `Vendor / ProviderFamily / SiteRegistry`。[现有领域词汇](../../GLOSSARY.md)

### 5.3 复用协议实现的边界

独立 Provider 不等于复制 SDK、HTTP 生命周期、流队列、重试和资源治理。共享的仍是已有 Chat / Responses / Messages adapter；新 Provider 只选择经过核验适用的请求、解析和能力行为。[Client adapter 装配](../../ai/client.go)、[Chat](../../ai/chat.go)、[Responses](../../ai/responses.go)、[Messages](../../ai/anthropic.go)

**不能只加常量和 JSON**。现在多处直接判断 `ProviderZhipu` / `ProviderMiniMax`，决定最终 payload 校验、SSE 解析、原生 reasoning 处理、用量和错误解释。新增 ID 若没有显式选择对应行为，可能落入 OpenAI / Anthropic 默认分支；国际 GLM 或中国 MiniMax 也不能因同品牌就全量沿用旧站行为。[能力默认](../../ai/responses_capabilities.go)、[Messages 默认](../../ai/anthropic_capabilities.go)、[Responses 分派](../../ai/responses.go)、[Chat 最终检查](../../ai/chat_request.go)、[历史转换](../../ai/responses_assistant.go)

实施时建议逐协议检查这些分支：有两个真实服务且契约一致的部分可以共享内部函数；差异留在明确分支或小型不可变配置中。共享实现选择不得改写真正的 `Model.Provider`、调用归属或历史来源。尤其不能先把 `minimax-cn` 规范化成 `minimax`，再做 Native State 信任比较——那会抹掉方案 2 要保留的安全边界。[nativeReplay](../../ai/native.go)、[CallAttribution](../../ai/result.go)

### 5.4 Provider 之外仍须守住的边界

1. **Provider ID 不是凭据校验器。** 当前 Credential 核对租户、引用、账户和配置版本，不带独立 Provider 字段；Host 必须把凭据引用绑定到明确服务契约，不能把另一站 key 作为失败重试的候选。[Credential](../../ai/binding.go)、[checkSnapshot](../../ai/preflight.go)
2. **Provider ID 不自动约束 Endpoint。** 当前核心只验证 HTTPS，测试情况下允许 loopback HTTP，没有按 Provider 核对官方域名。Host 的官方站点预设应验证所选服务与端点匹配；可信代理是显式配置，不能通过 URL 猜 Provider 或在失败后换站。[endpointAllowed](../../ai/preflight.go)、[Host 示例](../../ai/examples/minimaxhost/bindings.go)
3. **缓存键还依赖账户标识质量。** 当前派生键只包含 Tenant、AccountScope 和 Session，没有 Provider。采用方案 2 不会自动改变这一点；Host 应使用跨服务不冲突的账户作用域标识。若要求核心在相同账户别名下也显式区分站点，应另行把 Provider 纳入缓存命名空间并验证其影响；不要把本建议写成现有保障。[cache_key.go](../../ai/cache_key.go)
4. **价格必须各自取证，但不需要改变公共币种。** 当前 CostRates 和 UsageCost 统一用 USD；中国站人民币价应按明确日期、来源与汇率形成价格快照，不能直接当美元写入，也不能从国际站价格反推。套餐剩余额度、最终扣费和按 token 估算成本是不同概念。[usage.go](../../ai/usage.go)、[GLM 目录价格来源](../../ai/catalog_glm.go)、[ADR-0042](../../docs/adr/0042-barness-ai-minimax-international-target.md)
5. **网络路径单独配置。** 国际站名称不保证物理数据驻留，也不决定必须走代理。现有 Transport 是 Client 构建时固定的；Host 若需要不同网络路径，可提供按明确目标路由的受控 Transport 或分别装配 Client，不把路由策略藏到 Provider 命名里。[Config.Transport](../../ai/client.go)、[ADR-0046](../../docs/adr/0046-region-routed-validation-round.md)

## 6. 两方案对比

| 维度 | 方案 1：Binding 站点配置 | 方案 1：站点进入完整身份 | 方案 2：独立 Provider |
| --- | --- | --- | --- |
| 可实现性 | 高，连接配置已有基础 | 高，需完整设计与迁移 | 高，利用现有身份模型 |
| 同一 Client 同时调用两站 | 可以，但共享同一型号身份 | 可以 | 可以 |
| 同名 Model 独立能力 / 默认值 / 价格 | 不支持；只能拆 Client 或扩展目录 | 支持 | 直接支持 |
| 账户与凭据隔离 | Binding 可以表达，Host 必须正确装配 | Binding + Site 必须一致 | Binding + Provider 必须一致 |
| 原生历史跨站处理 | 当前没有 Site 检查，不能只换 URL 就宣称安全 | Site 必须进入来源检查 | 现有 Provider 检查可利用 |
| 型号新增、下架与各市场发布 | 白名单可收窄；元数据仍共用 | 独立 | 独立 |
| 协议实现复用 | 可以，需按站点选择行为 | 可以，需按站点选择行为 | 可以，需显式接入新 Provider 行为 |
| 公开模型与持久化契约影响 | 小，但不能完整满足需求 | 大，身份增加维度 | 相对小，仍需审核下游对新枚举的处理 |
| 正式支持证明 | 仍须按真实目标独立取证 | 仍须独立取证 | 仍须独立取证 |
| 长期主要成本 | 元数据与站点不一致的隐患 | 全链路增加 Site 维度 | 目录条目增加，已验证共同行为需有序复用 |
| 更合适的场景 | 同一服务的连接预设或等价入口 | 大量需独立管理的同契约部署 | 独立市场、账户、计费、型号事实的服务 |

比较是根据现有代码推导的定性判断；没有声称工期数字或性能优劣。两种完整方案都可达到目标，方案 2 的优势在于不用增加新身份轴即可满足此次差异化市场需求。[身份定义](../../docs/adr/0020-barness-ai-model-operations-and-binding-authorization.md)、[原生来源](../../ai/native.go)、[目录与授权](../../ai/call_runtime.go)

## 7. 若进入实施，建议的最小切片

本节是后续建议，未创建实施工单或修改 accepted ADR。

| 顺序 | 工作范围 | 完成条件 |
| --- | --- | --- |
| 1 | 记录新 ADR，确认四个 Provider 的精确定义与标准 API 范围 | 沿用 ADR-0020 四维身份；明确扩展 ADR-0042“当前只接入国际站”的范围，同时保留无回退、无证据互授原则 |
| 2 | 先在公共 Client 边界建立同名型号、跨站历史、错误端点 / 凭据与隔离失败场景 | 新行为有可重复的失败依据，再改生产实现 |
| 3 | 显式接入新 Provider 的协议能力与最终请求 / 解析分支 | 不落入未核验的默认能力；只复用已证明一致的部分 |
| 4 | 各自准备候选 Model 数据、Host Binding 和凭据用途声明 | 同名型号可独立查找、定价、配置和撤销；不复制另一站支持状态 |
| 5 | 运行受影响公共 E2E，保留脱敏请求、响应、归属、价格与降级工件 | 可重放并验证隔离和实际 wire；旧 Provider 相关行为回归通过 |
| 6 | 新身份各自取得获准范围内的真实证明，再晋升内置目录 / 支持矩阵 | 候选、协议实现、真实支持和完整发布状态分别有确切依据 |

与代码关联的改动集中在 `ai/binding.go`、`ai/catalog_*`、各协议能力 / 请求 / 解析分支、Host 示例及支持证明消费端；无需改 Target 的 BindingID / ModelID 入口形状。内部过时路径直接替换，不增加把新 ID 折回旧 ID 的兼容 shim。[Target](../../ai/binding.go)、[协议能力](../../ai/responses_capabilities.go)、[证明消费](../../ai/internal/testkit/supportmatrix/minimax.go)、[工程原则](../../AGENTS.md)

建议先验证的 E2E 结果：同一 Client / 同一型号名可命中两份不同元数据；禁用一站不影响另一站；跨站受信原生历史降级且工具关联完整；并发与重试固定各自 Binding / Credential 快照；服务错误不触发换站；最终 payload 使用实际型号 ID，价格与归属属于实际 Provider。适用范围由具体代码影响决定，不能凭研究文档宣称上述场景已通过。[验证策略](../../docs/agents/verification.md)

付费验证需另有身份、范围、资格与预算的明确授权；race 仅在人类明确选择相应范围时执行。本次只做文档引用和来源完整性检查，不运行这些验证。[verification.md](../../docs/agents/verification.md)

## 8. 限制和待核验项

- 公开文档能证明厂商声明，不能证明某个账户已经获得模型、协议或套餐资格；本研究没有使用账户验证。
- 同名型号不证明参数、缓存、原生状态或计费相同；两站文档相同也不等于实际服务行为相同。
- 标准 API、Coding Plan / M Plan 是接入资格与结算用途，不能只凭“中国 / 国际”二分决定。当前保留 Host 显式用途与型号授权；只有出现无法由这些边界表达的真实差异，才另行讨论是否需要新的服务身份。
- 本研究没有变更冻结 pi 差分基线。上游存在某个 Provider 或协议实现，不能自动把 barness 新地区路线计为已有差分 / 正式支持。[冻结基线](../../ai/internal/testkit/pioracle/node/PROVENANCE.md)、[Extension Route](../../GLOSSARY.md)
- 本文建议保持现有 ID 含义；若另行决定统一重命名，需先识别外部配置、持久化 transcript / envelope 与证据契约，再设计显式迁移，不能静默重解释历史数据。[原生封套](../../ai/native.go)、[ADR-0042](../../docs/adr/0042-barness-ai-minimax-international-target.md)

## 9. 研究验证记录

2026-10-20 08:44（Asia/Shanghai）完成纯文档检查：

- 已阅读本地 Provider / Model / Binding / Native State 相关定义、调用路径、能力选择、Host 示例和相关 ADR；代码基线见文首。
- 官方研究覆盖四个服务的公开端点、GLM 两站价格与 Key 分类、MiniMax 两站定价 / SDK / M Plan 文档，以及固定提交的 pi / models.dev 源码。综合时再次核对关键价格页和 Key 分类；修正了折叠价格表容易遗漏的 GLM-4.7-Flash 条目，未以价格表缺项断言模型不可用。
- Python 文档检查退出码 `0`：81 处本地引用（36 个独立目标）全部存在；36 处外部引用（26 个独立 URL）保留厂商或上游源码原地址；无行尾空白，代码围栏配对。这里的外部计数是引用完整性检查，不冒称每个 URL 在复核时都能成功响应。
- 仅新增本研究文件；未修改生产代码、目录数据、规格、accepted ADR、支持矩阵或历史证据。未执行 Go 测试、race、发布门禁或付费真实验证；这些均不是本次纯文档检查的通过项。
