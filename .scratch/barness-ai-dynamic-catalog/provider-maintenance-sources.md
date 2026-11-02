# barness 模型目录：维护源政策与历史研究

状态：正文为本轮确认后的现行维护源政策。附录保留 2026-10-29（Asia/Shanghai）的来源研究、固定提交和当时结论，不表示已重新核查实时外部数据。

现行设计与决策见 [design.md](design.md)、[discussion.md](discussion.md) 和 [ADR-0058](../../docs/adr/0058-barness-catalog-maintenance-and-isolated-loading.md)。models.dev 指定接口的 schema、字段语义及映射研究见 [models-dev-research.md](models-dev-research.md)；早期 Pi 背景研究见 [research.md](research.md)。

## 1. 现行维护源政策

**models.dev 的 [catalog.json?type=all](https://models.dev/catalog.json?type=all) 是主要维护源；barness 保留目录定义的维护权。** 官方模型文档、公告、定价页和 Provider model API 用于补充缺失信息及核对冲突。上游数据形成候选，经过 barness 映射、人工确认和受控 E2E 后，才进入本地生效目录；每个程序的有效目录由逐模型加载结果决定。

本轮明确取代 2026-10-29 的“官方摘录是唯一主要底稿，models.dev 只能辅助”建议。采用主维护源不等于把其 schema、缺省值或自动合并结果直接当作 barness 的运行契约。出现冲突时记录来源和判断理由，由维护者决定，不能按来源优先级静默覆盖已有人工结论。

| 来源 | 现行职责 | 边界 |
| --- | --- | --- |
| models.dev `catalog.json?type=all` | 模型发现、模型信息映射和批量价格更新的主要输入 | 固定本次取回的数据与来源摘要；原始字段不能越过 barness 映射和确认直接生效 |
| 官方模型文档、公告、定价页、model API | 补充缺失模型与字段，核对协议、价格、生命周期和上游冲突 | model API 往往是账户可见集，不能单独证明全球型号下架或缺失价格 |
| barness 已确认的映射、人工字段与人工覆盖 | 决定四维身份、协议适配与 barness 特有语义，记录选择和理由 | 自动同步保留人工覆盖；需要变更时显式展示差异并再次确认 |
| Pi、LiteLLM、OpenRouter、Vercel 等其他来源 | 保留历史架构参考和必要的交叉核对 | 不作为本期自动维护的另一套主源；网关价不能直接填入官方 Provider |

## 2. 候选映射与人工确认

维护对象覆盖 barness 支持的全部 Provider，身份保持 `(Operation, Provider, API, ID)`。中国站与国际站分别维护。models.dev 的 Provider、实验室、模型和 serving 记录与 barness 路线的对应关系需要明确记录，不能仅按展示名、品牌或模型族合并。

维护过程中保留以下关系，具体存储形状以设计稿为准：

- 本次使用的 models.dev 原始数据、来源 URL 与摘要，以及对应的模型和 serving 来源定位。
- 上游字段到 barness 字段的映射；有明确对应关系的字段按已确认规则转换。
- 无法直接推出的 barness 特有字段、缺失字段、来源冲突和人工覆盖，记录人工确认的值及理由。
- 每次操作的候选差异、确认结果与变更记录。批量更新也不得静默覆盖人工字段。

价格缺失保持未知，不得补成 0。币种、计费单位、上下文阶梯、服务等级和地区条件必须按 barness 的语义确认；不能把无法表达的条件价默认为一条普通费率。上游 `missing`、未出现在某次列表、`deprecated` 或未来停服日期都只形成提示，不能自动删除既有模型；模型下架是显式维护操作。

## 3. 本期范围与生效条件

本期实现维护 CLI、本地 JSON 加载和受控 E2E 合约复用。CLI 支持按 Provider/model 新增候选、辅助确认字段、版本管理、价格批量更新和模型下架；操作有变更记录，`help`、`dry-run` 等功能供人和 AI 检查计划后再执行。

新增候选通过实际受测代码的受控 E2E 后，才能进入生效目录。受控 E2E 证明的是候选满足该测试覆盖的既有合约，不自动宣称真实厂商调用、账户可见性和全部可选能力均已验证。需要代码适配时先适配并重新检查，未来与程序配套发布；不维护模型最低程序版本，也不实施程序版本过滤。模型级加载或使用错误保留诊断信息，并在适当时提示检查目录或尝试升级，不因此阻断其他有效模型；文件级错误、逐模型状态和严格未知字段处理见设计稿。

运行时只消费 barness 维护后的本地 JSON，不直连 models.dev 或其他外部目录。线上目录发布、下载、缓存、启动检查和五秒更新预算均延期，`barness.dev` 的线上分发定义另行设计。本期的候选确认与本地生效不代表线上发布服务已实现。

## 4. 初次迁移基准

初次迁移保留 [catalog-snapshot.json](../../ai/release/catalog-snapshot.json) 中现有 118 条四维模型身份及已确认语义，不因上游覆盖不完整而缩减。当前有 6 条记录没有准确对应的 models.dev serving 来源；这些记录保留现有定义并明确标记来源缺口，后续人工补充或确认，不伪造映射、不因缺失而下架。

本文件只规定维护源及确认边界，不重新定义现有 Catalog 字段。当前 `{hash, catalog}` 形状、逐模型加载状态、候选测试合约和迁移明细由设计稿及 models.dev 研究说明。

## 附录 A：2026-10-29 历史来源研究

以下保留当时的事实、引用和研究局限。本文没有重新请求这些实时接口。附录中“建议”“不适合”“事实底稿”等措辞属于历史分析；现行规范以正文为准。原先第 5、6 节围绕“官方底稿优先”的架构与取舍建议已被正文取代，重复的政策性论证不再保留。

固定提交：models.dev `140ccedf1a61e06064ce609c6d09f9d2b93480a7`（`dev`，2026-10-29T06:32:19Z）；Pi `e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a`；LiteLLM `ec7162aea2079c4616d9eda5f0738956d59a3102`（`main`，2026-10-29T07:23:11Z）。这些提交描述当时接口，不能替代本轮 `catalog.json?type=all` 研究。

当时核对的一等 Provider 来自 [ai/binding.go](../../ai/binding.go)：`openai`、`anthropic`、`google`、`deepseek`、`xai`、`zhipu`、`zai`、`minimax`、`minimax-cn`、`moonshot`、`moonshot-cn`、`typesafe`。内置目录的组装见 [ai/catalog.go](../../ai/catalog.go)。

### A.1 上游对照表

以下是 2026-10-29 的来源比较记录。表中“适合性”是当时的判断，不能替代本文现行政策；models.dev 的维护源优先级已由本轮确认改为主来源。

| 来源 | 实际有什么 | 缺少什么 / 会错在哪里 | 许可证 | 当时的适合性判断（历史） |
| --- | --- | --- | --- | --- |
| models.dev v2 | 按 Provider 分文件的 TOML：模态、限额、能力、USD/百万 token 价格、上下文阶梯、区域/服务等级倍率、生命周期 `alpha/beta/deprecated`。实验室文件与托管文件分开。中国站与国际站是不同 Provider 目录。 | 价格可选；区域价是倍率条件，不是 barness 的 Provider 身份。推理控制变更可要求人审，只改价可以自动合并。公开 v1 API 的价格字段是扁平数字。 | MIT，须保留版权声明 | 当时仅作为候选与交叉核对；本轮已改为主维护源 |
| Pi 生成器 | 构建时拉取上游，写成 Pi 自己的 Provider JSON。覆盖层很小。 | 缺失费率变 0；缺失限额变 4096。只映射上下文阶梯，地区/服务等级/时间价没有对应项。证明的是 Pi，不是 barness。 | MIT | 可借鉴结构；不要导入它生成的价格 |
| LiteLLM `model_prices_and_context_window.json` | 一个大 JSON：每 token 价格、上下文、`mode`、能力布尔、`deprecation_date`、部分条目的 `source` 与 `supported_regions`。 | 默认从浮动的 `main` 拉取。键把厂商、Bedrock 区域价和微调价混在一起。`sample_spec` 用 0.0 做字段示例，不是“缺失即免费”的证明。未在源码里找到“缺价填 0”的规则。 | 仓库根与非 `enterprise/` 内容为 MIT；GitHub 许可证字段是 `NOASSERTION`，因为文件头把 `enterprise/` 分开 | 可作价格交叉核对；不要当 schema 或身份源 |
| OpenRouter `GET /api/v1/models` | 网关价（USD/token）、上下文、模态、`supported_parameters`、推理档、`expiration_date`。`"0"` 在它自己的文档里表示该计费项免费。 | 价是 OpenRouter 的价。默认只返回文本输出。不区分 barness 的中国站/国际站。Pi 把缺失和负数价格收成 0，把“免费”和“未知”捏在一起。 | 文档未把这份响应授予可再发布的数据许可证；本次未把它当可 vendor 的数据库 | 仅当需要核对“网关上有什么” |
| Vercel AI Gateway `GET /v1/models` | 无鉴权的网关模型表：ID、上下文、价格、推理控制。同一模型可以有多个托管方。 | Pi 只把它用于 `vercel-ai-gateway` 这一家，并在转换时把缺价变成 0。不是 OpenAI/Anthropic 的官方价，也不按 barness 的区域身份拆开。 | 未把响应声明为可 vendor 的目录许可证 | 否，对 barness 一等 Provider 不如 models.dev |
| 官方列表 API + 官方定价页 | 各家实际返回的 ID，以及该页写明的限额、模态、生命周期。价格在单独的定价页，xAI 的列表 API 例外地带价格，但单位不是 barness 的美元/百万 token。 | 见附录 A.1.6。多数列表没有价格、没有硬约束全文、没有 barness 协议名。列表往往是账户可见集。 | 各厂商文档的使用条款各自成立；摘录进 Git 是维护证据，不是获得了再分发官方目录的授权。发布物仍必须是 barness 审核后的定义 | 当时作为主要事实底稿；本轮改为补充与冲突核对 |

#### A.1.1 models.dev

**事实：** 数据是仓库里的 TOML。实验室事实在 `models/<lab>/<id>.toml`，某家 API 怎么托管在 `providers/<id>/models/`。文件名即模型 ID。当前文件要求 `schema = 2`，schema 在 `packages/core/src/schema-v2.ts`。网站和 API 由这些文件生成。[README](https://github.com/anomalyco/models.dev/blob/140ccedf1a61e06064ce609c6d09f9d2b93480a7/README.md)

**事实：** v2 模型类型包括 `chat`、`image`、`video`、`embedding`、`decision`、`transcription`、`speech`、`realtime` 等。价格是 USD；token 价按每百万 token。`cost` 在模型对象上是可选的。价格数字允许 0，不允许负数。条件价的键包含 `context`、`service_tier`、`region`、`date`、`holiday` 等。`status` 只有 `alpha`、`beta`、`deprecated`，另有可选 `deprecation_date`。这不是 Anthropic 那种 `active/deprecated/retired` 加可修订的 `retires_at`。[schema-v2.ts 类型与价格](https://github.com/anomalyco/models.dev/blob/140ccedf1a61e06064ce609c6d09f9d2b93480a7/packages/core/src/schema-v2.ts#L12-L25)、[价格与 status](https://github.com/anomalyco/models.dev/blob/140ccedf1a61e06064ce609c6d09f9d2b93480a7/packages/core/src/schema-v2.ts#L462-L768)

**事实：** 贡献者规则写明，token 价必须是美元/百万 token，不得把人民币、欧元等直接当成美元；换算要把汇率和日期写在文件头注释里。区域差价进 `[[cost.multipliers]]`，不是另一个 Provider。[AGENTS.md 价格](https://github.com/anomalyco/models.dev/blob/140ccedf1a61e06064ce609c6d09f9d2b93480a7/AGENTS.md#L237-L243)

**事实：** 它把中国站和国际站分成不同目录，而不是在一个 OpenAI 兼容网关里用 region 字段区分。本次读到的 `provider.toml`：`openai` 指向 OpenAI 文档；`azure` 是另一家；`minimax` 的 API 为 `https://api.minimax.io/anthropic/v1`，`minimax-cn` 写明中国站已从 `api.minimaxi.com` 迁到 `api.minimax.cn`；`moonshotai` 为 `https://api.moonshot.ai/v1`，`moonshotai-cn` 为 `https://api.moonshot.cn/v1`；`zai` 文档指向 `docs.z.ai`、API 为 `https://api.z.ai/api/paas/v4`；`zhipuai` 文档指向 `docs.bigmodel.cn`、API 为 `https://open.bigmodel.cn/api/paas/v4`。[openai provider.toml](https://github.com/anomalyco/models.dev/blob/140ccedf1a61e06064ce609c6d09f9d2b93480a7/providers/openai/provider.toml)、[minimax-cn](https://github.com/anomalyco/models.dev/blob/140ccedf1a61e06064ce609c6d09f9d2b93480a7/providers/minimax-cn/provider.toml)、[moonshotai-cn](https://github.com/anomalyco/models.dev/blob/140ccedf1a61e06064ce609c6d09f9d2b93480a7/providers/moonshotai-cn/provider.toml)、[zai](https://github.com/anomalyco/models.dev/blob/140ccedf1a61e06064ce609c6d09f9d2b93480a7/providers/zai/provider.toml)、[zhipuai](https://github.com/anomalyco/models.dev/blob/140ccedf1a61e06064ce609c6d09f9d2b93480a7/providers/zhipuai/provider.toml)

**事实：** 谁在改。README 把它定义为社区贡献的开放数据库，并供 opencode 内部使用。同步工作流每小时第 17 分跑一次（`17 * * * *`），也会手动触发。变更写成 PR。自动合并函数把“安全”定义成：新建或删除的模型文件不超过阈值，且推理元数据的变化要么没有、要么落在一份允许自动过的网关名单里。OpenAI、Anthropic 不在那份名单。只改价、不改推理元数据时，推理检查直接跳过。因此已有模型的价格更新可以在无人审阅的情况下被标成 safe 并自动 squash 合并。[sync-models.yml](https://github.com/anomalyco/models.dev/blob/140ccedf1a61e06064ce609c6d09f9d2b93480a7/.github/workflows/sync-models.yml#L1-L7)、[自动合并](https://github.com/anomalyco/models.dev/blob/140ccedf1a61e06064ce609c6d09f9d2b93480a7/packages/core/src/sync/auto-merge.ts#L6-L106)、[safe 则合并](https://github.com/anomalyco/models.dev/blob/140ccedf1a61e06064ce609c6d09f9d2b93480a7/.github/workflows/sync-models.yml#L152-L153)

**事实：** API 不是另一份手工表。`dev` 分支每次 push 都部署。公开入口包括 `https://models.dev/api.json` 和实验性的 `https://models.dev/experimental/v2.0/api.json`。Worker 从已构建的静态资源读这些 JSON。[deploy.yml](https://github.com/anomalyco/models.dev/blob/140ccedf1a61e06064ce609c6d09f9d2b93480a7/.github/workflows/deploy.yml)、[worker 路由](https://github.com/anomalyco/models.dev/blob/140ccedf1a61e06064ce609c6d09f9d2b93480a7/packages/function/src/worker.ts#L102-L109)

**事实：** 它对 OpenAI 列表的态度本身就否定了“列表即目录”。同步模块 `skipCreates: true`、`deleteMissing: false`，注释写明 `/v1/models` 是账户范围，含旧型号和内部型号，没有权威生命周期；API 多出来的 ID 要求手写元数据，不能自动建文件。[openai.ts](https://github.com/anomalyco/models.dev/blob/140ccedf1a61e06064ce609c6d09f9d2b93480a7/packages/core/src/sync/providers/openai.ts#L54-L71)

**事实：** 许可证是 MIT，版权 2025 models.dev。复制实质性部分必须带版权声明和许可文本。许可不担保正确性。[LICENSE](https://github.com/anomalyco/models.dev/blob/140ccedf1a61e06064ce609c6d09f9d2b93480a7/LICENSE)

**推断：** 它适合做发现和 diff 的输入，因为身份拆分和字段比网关列表齐。它不适合做 barness 的事实终局，因为价格可以无人合并，区域是计费条件而不是 Provider，而且 v2 的条件价进不了 Pi 那种单一费率结构。

#### A.1.2 LiteLLM 价格表

**事实：** 规范文件仍在仓库根目录 `model_prices_and_context_window.json`，该提交上约 3.0 MB。`sample_spec` 用 0.0 和示例字符串说明字段：`input_cost_per_token`、`output_cost_per_token`、`max_input_tokens`、`max_output_tokens`、`mode`、能力布尔、`deprecation_date`、`supported_regions`，以及按次/按会话的附加费用。后面的真实条目是各厂商和路由的具体数字，部分条目带 `source` URL。[文件头](https://github.com/BerriAI/litellm/blob/ec7162aea2079c4616d9eda5f0738956d59a3102/model_prices_and_context_window.json#L1-L40)

**事实：** 运行时默认 URL 是浮动分支，不是某个提交：

`https://raw.githubusercontent.com/BerriAI/litellm/main/model_prices_and_context_window.json`

可用 `LITELLM_LOCAL_MODEL_COST_MAP=True` 或离线模式改读包内备份。拉取失败或模型数异常收缩时回退本地备份。模型不在表里时，查询最大输出的路径会抛出“isn't mapped yet”，并指向这个 GitHub 文件，而不是返回一个编造的限额。[默认 URL](https://github.com/BerriAI/litellm/blob/ec7162aea2079c4616d9eda5f0738956d59a3102/litellm/__init__.py#L411-L414)、[回退](https://github.com/BerriAI/litellm/blob/ec7162aea2079c4616d9eda5f0738956d59a3102/litellm/litellm_core_utils/get_model_cost_map.py#L1-L9)、[未映射](https://github.com/BerriAI/litellm/blob/ec7162aea2079c4616d9eda5f0738956d59a3102/litellm/utils.py#L6122-L6131)

**事实：** 同一文件里的 `bedrock/<region>/<model>` 键和 `ft:` 键被源码标明只用于计价，不应当成可调用模型名单。别名如果和已有键冲突会被跳过并打警告。[只用于计价](https://github.com/BerriAI/litellm/blob/ec7162aea2079c4616d9eda5f0738956d59a3102/litellm/__init__.py#L720-L742)、[别名冲突](https://github.com/BerriAI/litellm/blob/ec7162aea2079c4616d9eda5f0738956d59a3102/litellm/litellm_core_utils/get_model_cost_map.py#L583-L627)

**事实：** 许可证文件写明 `enterprise/` 若存在则另有许可证；该目录以外的内容是 MIT，版权 2023 Berri AI。价格 JSON 在仓库根，不在 `enterprise/`。GitHub API 对该仓库报告的 SPDX 是 `NOASSERTION`，与这段分段许可一致。MIT 允许带版权声明地复制，不提供正确性担保。[LICENSE](https://github.com/BerriAI/litellm/blob/ec7162aea2079c4616d9eda5f0738956d59a3102/LICENSE)

**推断：** 表更新靠向 `main` 提交。本次看到的最新提交信息是修正 Fireworks 上 MiniMax-M3 的视觉标记。没有在 CONTRIBUTING 里找到“改价必须附官方来源并经维护者逐条确认”的规则。因此它的延迟可以很短，schema 却是一个会继续加键的扁平字典，区域信息也不是 barness 的 Provider 身份。

#### A.1.3 OpenRouter

**事实：** `GET /api/v1/models` 的模型对象必填字段包括 `id`、`canonical_slug`、`name`、`created`、`pricing`、`context_length`、`architecture`、`top_provider`、`per_request_limits`、`supported_parameters`。`pricing` 的各数字是 USD/token、每次请求或每个单位。`context_length` 是最大上下文。`architecture` 有输入/输出模态。`expiration_date` 是可能下架的日期，空表示没有。`top_provider.max_completion_tokens` 是当前置顶端点的输出上限。默认不传 `output_modalities` 时只返回文本输出；图像和 `decisions` 要显式查询。不传分页参数时一次返回全表。[列表文档](https://openrouter.ai/docs/api/api-reference/models/list-all-models-and-their-properties.md)、[模型指南](https://openrouter.ai/docs/guides/overview/models.md)

**事实：** 指南写明，价格对象里的 `"0"` 表示该功能免费，不是“未知”。[Pricing Object](https://openrouter.ai/docs/guides/overview/models.md)

**推断：** 这是网关自己的价和路由，不是 OpenAI 或月之暗面的官方价。barness 若用它填 `openai` 或 `moonshot` 的目录，会把网关价写成厂商价。

#### A.1.4 Vercel AI Gateway 与 Pi 的关系

**事实：** Pi 的来源目录把 Vercel 列成与 models.dev、OpenRouter 并列的抓取器，不是 models.dev 的替代品。`vercel-ai-gateway` 这家 Provider 直接使用网关抓取结果，只对少数型号改输出上限。[README 来源表](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/ai/catalog/README.md)、[Provider 文件](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/ai/catalog/providers/vercel-ai-gateway.ts)

**事实：** Pi 请求 `https://ai-gateway.vercel.sh/v1/models`。转换函数把每个费率从“每 token”乘到“每百万 token”；解析失败时用 0。分类器路径把缓存读写固定写成 0。[vercel-ai-gateway.ts 转换](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/ai/catalog/sources/vercel-ai-gateway.ts#L35-L50)

**事实：** Vercel 自己的文档说该 URL 不需要 API key，返回模型 ID、上下文、价格和推理控制；同一模型可以由多个 provider 托管，价格和性能可以不同。模型字符串形如 `creator/model-name`。[Models and Providers](https://vercel.com/docs/ai-gateway/models-and-providers)

**推断：** 对 barness 已经按官方身份维护的 Provider，Vercel 比 models.dev 更不合适：它多一层网关价，而且不保留 `minimax` 与 `minimax-cn` 这种身份。它只说明 Pi 的网关 Provider 从哪来。

#### A.1.5 许可证会不会挡住 vendor

不会因为 MIT 而禁掉“把摘录放进私有维护仓库”。models.dev 与 Pi 都是 MIT。LiteLLM 价格 JSON 落在其 MIT 范围内。三份 MIT 都要求保留版权声明和许可文本，并写明按原样提供、无正确性担保。

这不构成把它们的响应交给 barness 客户端执行的理由。当时 Q1 讨论的是发布权，不是版权障碍；本轮仍由 barness 决定可生效的定义。OpenRouter 与 Vercel 的公开 JSON 本次没有看到一份“可把响应再发布成你的目录”的许可证；即便有，它们仍然是网关价。

#### A.1.6 barness 各 Provider 的官方列表和定价页

2026-10-29 重读 OpenAI、Anthropic、Gemini 的官方文档，并补上其余一等 Provider。没有发送带 key 的请求。列表“不返回”的判断以该页写出的响应字段为准。

| Provider | 官方列表 | 官方价格 | 列表有、且可直接当事实的 | 列表没有的 |
| --- | --- | --- | --- | --- |
| OpenAI | [GET /v1/models](https://developers.openai.com/api/reference/resources/models/methods/list)：`id`、`created`、`object`、`owned_by`、可选 `shutdown_date`。要 key。文档把能力定义成列出组织有权访问的模型 | [Pricing](https://developers.openai.com/api/docs/pricing)：按每百万 token，并区分标准、Batch、Flex、Fast 等 | ID、创建时间、可选停服日期 | 价格、上下文、模态、协议参数 |
| Anthropic | [GET /v1/models](https://platform.claude.com/docs/en/api/models/list)，文档示例版本头 `2026-10-21`：ID、展示名、`created_at`、能力集合、`max_input_tokens`、`max_tokens`、`line`、`lifecycle`、`deprecated_at`、`retires_at`。默认可不含 `retired`；可按 workspace 过滤 | [Pricing](https://platform.claude.com/docs/en/about-claude/pricing) | 限额、一批能力、生命周期。`retires_at` 过期仍以 `lifecycle` 为准 | 价格。不能从 ID 推断 `line` |
| Google Gemini | [models.list](https://ai.google.dev/api/models)：名称、基础模型 ID、版本、说明、输入/输出 token 上限、支持的方法、`thinking`、temperature/topP/topK。`topK` 为空表示不允许发送该参数。要 key，有分页 | [Pricing](https://ai.google.dev/gemini-api/docs/pricing) 与 [Models 指南](https://ai.google.dev/gemini-api/docs/models) | 上下文限额、方法、采样约束、是否 thinking | 价格。指南写明 2.5 只对曾经用过的用户开放，同时写明尚未弃用、会继续服务 |
| xAI | [GET /v1/models](https://docs.x.ai/developers/rest-api-reference/inference/models)：认证 key 可见的模型。含 ID、`created`、`context_length`、文本 token 价、长上下文价和阈值、图像 token 价、图像按张价、`capabilities.reasoning_effort` | [Pricing](https://docs.x.ai/developers/pricing) 与 [Models](https://docs.x.ai/developers/models)：美元/百万 token 的表 | ID、上下文、推理档，以及一组价格数字 | 价格单位是“每 1 亿 token 的美元美分”，图像矩阵另有更小单位。这不是 barness 的美元/百万 token，不能直接拷贝数字。列表仍是该 key 可见集 |
| Kimi 国际 `moonshot` | [GET https://api.moonshot.ai/v1/models](https://platform.kimi.ai/docs/api/list-models)：`id`、`created`、`owned_by`、`context_length`、`supports_image_in`、`supports_video_in`、`supports_reasoning`。`platform.kimi.ai` 与 `platform.kimi.com` 的 key 不可混用 | [定价说明](https://platform.kimi.ai/docs/pricing/chat) 与 [模型列表](https://platform.kimi.ai/docs/models)。本次 markdown 里的价格表格是空的，所以不引用具体数字。页面写明按每百万 token，K3 缓存写按 TTL 另计 | ID、上下文、三类能力布尔 | 价格、输出上限、停服状态的结构化字段。停服说明在模型页正文里 |
| Kimi 中国 `moonshot-cn` | [GET https://api.moonshot.cn/v1/models](https://platform.kimi.com/docs/api/list-models)：字段与国际站同形，服务器是 `api.moonshot.cn` | [中国站定价说明](https://platform.kimi.com/docs/pricing/chat)。同样，本次 markdown 价格表为空 | 同国际站列表 | 同国际站。文档明确两边 key 不能互用，所以不能合成一个 Provider |
| MiniMax 国际 `minimax` | [GET https://api.minimax.io/v1/models](https://platform.minimax.io/docs/api-reference/models/openai/list-models)：`id`、`object`、`created`、`owned_by` | [Pricing Overview](https://platform.minimax.io/docs/pricing/overview)，美元/百万 token，并区分标准与优先 | ID 与创建时间 | 价格、上下文、模态、输出上限 |
| MiniMax 中国 `minimax-cn` | 请求 `platform.minimaxi.com` 的列表文档时落到 `platform.minimax.cn`。OpenAPI `servers` 为 `https://api.minimax.cn`，字段与国际站同形。[中国站列表](https://platform.minimax.cn/docs/api-reference/models/openai/list-models) | [中国站定价](https://platform.minimax.cn/docs/pricing/overview)，单位是元/百万 token | ID 与创建时间 | 同国际站。货币与国际站不同，不能共用一条价格 |
| DeepSeek | [GET /models](https://api-docs.deepseek.com/api/list-models)：`id`、`owned_by`、`name`、`context_window`、`max_output_tokens`、输入/输出模态、effort 档、按协议的 `api_capabilities`。该页正文没有 price 字段 | 英文 URL `/quick_start/pricing` 本次返回的却是“Your First API Call”文档。中文页 [模型 & 价格](https://api-docs.deepseek.com/zh-cn/quick_start/pricing) 成功，单位是元/百万 token，并区分北京时间高峰/空闲 | ID、上下文、输出上限、模态、effort、协议能力差异 | 价格，以及高峰/空闲这种 barness 现有费率结构没有的时间条件 |
| Z.AI 国际 `zai` | 在 [docs.z.ai 的 API 目录](https://docs.z.ai/llms.txt) 里没有看到 List Models。目录里有 Chat Completion、图像、视频等调用接口 | [Pricing](https://docs.z.ai/guides/overview/pricing)：全部美元。GLM-5.3 为输入 $1.4、缓存输入 $0.26、输出 $4.4 / 百万 token | 不能从“目录里没看到”推出接口不存在 | 未见到列表响应字段。价格只在定价页 |
| 智谱中国 `zhipu` | 在 [docs.bigmodel.cn 文档索引](https://docs.bigmodel.cn/llms.txt) 里，模型信息是 [模型概览](https://docs.bigmodel.cn/cn/guide/start/model-overview)，不是一个 List Models API。概览给上下文和最大输出，并指向定价页 | [API 定价](https://docs.bigmodel.cn/cn/guide/start/pricing)：除特别说明外为元/百万 token。GLM-5.3 为输入 8、输出 28、缓存命中 2 | 概览页上的上下文和最大输出是文档事实，不是列表 API 事实 | 未见到与 Z.AI 分开的机器列表。货币与 Z.AI 不同；同一展示名不能合成一个 Provider |
| TypeSafe | [GET /v1/models](https://docs.typesafe.ai/api/list-models)：账户可用的 `name`、`description`、`release_date` | [Models](https://docs.typesafe.ai/models)：`jev-1.14.0` 与 `jev-1.13.0` 的每百万 token 价格、默认 64k/32k 上下文，以及长上下文头把两者提到 262k | 名称与发布日期 | 价格和上下文。别名 `jev-latest` 会移动，响应里的 `model` 才是版本 ID |

**事实：** 智谱与 Z.AI 的同名模型不是同一份价。2026-10-29 的官方页上，GLM-5.3 在 Z.AI 是美元 1.4 / 4.4，在智谱是人民币 8 / 28，每百万 token。MiniMax-M3 的国际站折扣价是美元，中国站是人民币。这是“中国站和国际站必须分开”的直接证据，不是风格偏好。[Z.AI 定价](https://docs.z.ai/guides/overview/pricing)、[智谱定价](https://docs.bigmodel.cn/cn/guide/start/pricing)、[MiniMax 国际](https://platform.minimax.io/docs/pricing/overview)、[MiniMax 中国](https://platform.minimax.cn/docs/pricing/overview)

**推断：** 官方列表适合发现“这个账户此刻能看见哪些 ID”以及该 schema 明确给出的限额和能力。它不适合单独生成可执行目录。xAI 虽然返回价格，单位换算本身就是 barness 语义，必须留在人工覆盖层。

### A.2 Pi 的真实维护模型

**事实：** `node catalog/generate.ts` 在生成时抓取上游，再写内置数据和 pi.dev 要用的 JSON。`sources/models-dev.ts` 把 models.dev v2 实验 API 归一化。OpenRouter、Vercel AI Gateway、NVIDIA NIM、Radius 有自己的模块。`providers/<id>.ts` 决定模型从哪来、用哪个 API 和 base URL、兼容默认、纠错，以及上游没有的模型。`families/` 写跨托管方仍成立的家族事实。README 写明优先去 models.dev 改正错误数据，本地纠错保持短小并说明原因。[README](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/ai/catalog/README.md)

**事实：** 归一化阶段保留“缺失价格就是 undefined”。下一步 `modelsDevCost` 的注释是 “Missing rates are free.”，实现把输入、输出、缓存读、缓存写的空值收成 0。`modelsDevModels` 再把缺失的 `contextWindow` 和 `maxTokens` 收成 4096，除非调用方另给默认。没有工具能力的模型会被丢掉。地区、服务等级和时间条件价没有 Pi 的对应结构，只有上下文阈值变成阶梯。[缺失仍是 undefined](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/ai/catalog/sources/models-dev.ts#L69)、[收成 0](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/ai/catalog/sources/models-dev.ts#L213-L219)、[默认 4096](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/ai/catalog/define.ts#L100-L132)

**事实：** OpenAI 的覆盖层不是全文。它从 models.dev 的 `openai` 映射到 Responses，删掉 `gpt-5.6` 这个 API 不接受的别名，把一批长上下文型号的窗口改回短上下文阈值，并手写 `gpt-5-chat-latest` 和 Decisions 分类模型。注释指向 OpenAI 官方文档。[openai.ts](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/ai/catalog/providers/openai.ts#L20-L121)

**事实：** 中国站在 Pi 里也是另一家 Provider，但基址可以和官方当前文档不一致。`minimax-cn` 仍请求 `https://api.minimaxi.com/anthropic`。models.dev 该提交的 `minimax-cn` 已写迁到 `api.minimax.cn`。MiniMax 中国站 OpenAPI 的 `servers` 也是 `https://api.minimax.cn`。[Pi minimax-cn](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/ai/catalog/providers/minimax-cn.ts)、[models.dev minimax-cn](https://github.com/anomalyco/models.dev/blob/140ccedf1a61e06064ce609c6d09f9d2b93480a7/providers/minimax-cn/provider.toml)、[中国站列表 OpenAPI](https://platform.minimax.cn/docs/api-reference/models/openai/list-models)

**事实：** OpenRouter 转换同样把缺失、非法和负数基础价收成 0。注释说明 OpenRouter 用 -1 表示路由器型号的价格取决于最终选中的模型；Pi 的处理是落成 0，而不是保留“未知”。[openrouter-api.ts](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/ai/src/providers/openrouter-api.ts#L36-L55)

**历史设计比较（该维护形状已被本轮取代）：** 2026-10-29 的 barness 设计写作“每家一份完整定义”，当时与 Pi 的比较认为它更严。严在发布闸门：人必须看过价格、协议和删除，缺失费率不能变成免费。**推断：** 这份严格性买到的是审查，不是“人手再敲一遍官方已经返回的字段”。Pi 的 OpenAI 覆盖层只改基址、窗口、个别上限，并补上上游没有的型号；注释把其中几项指回 OpenAI 文档。把整份定义再抄一遍，还会把 Pi 仍写着的 `api.minimaxi.com` 这类过期基址一起抄进来。

**推断：** 正确性要买在三个规则上，而不是买在重复录入上：缺价就拒绝发布该条价格；列表消失不等于删除；中国站和国际站的摘录不能合并。

### A.3 当时未核查的事项

- 没有用任何厂商 key 调用列表或推理 API。账户可见集与文档示例是否一致，没有当场验证。
- Kimi 国际站和中国站定价页的 markdown 导出里，价格表的表头和表体是空的。本文不引用 Kimi 的具体数字。模型页上的停服日期来自正文，不是来自列表字段。
- DeepSeek 英文路径 `/quick_start/pricing` 返回的文档 ID 是 `your_first_api_call`，不是价格页。价格事实取自同日成功打开的中文页。
- 智谱与 Z.AI 的已读 API 目录里没有 List Models。这只说明索引中没有该页，不证明不存在未编入索引的接口。
- 没有逐条读 models.dev 的全部同步器，也没有读完 `translate.ts`，所以没有证明 v2 的区域倍率一定不会进入 v1。已核对的是 README 列出的 v1 价格字段本身是扁平数字。也没有核对 LiteLLM JSON 里每个 barness 型号的数字。LiteLLM 源码里没有找到“缺价填 0”；`sample_spec` 的 0.0 只是字段示例，不能解释成运行时填零规则。
- 没有请求 OpenRouter 或 Vercel 的实时 JSON。字段以当日官方文档为准。
- 没有审 Pi 的 NVIDIA NIM 与 Radius 抓取器。README 已说明它们是独立来源；barness 没有对应的一等 Provider。
- TypeSafe 的价格和默认上下文来自 Models 页，不来自 `GET /v1/models` 的 schema。
