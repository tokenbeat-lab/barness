# models.dev 维护源研究：模型定义、映射与迁移覆盖

整理日期：2026-11-01（Asia/Shanghai）。状态：研究完成；供[当前设计](design.md)与实施讨论引用，不代替产品契约或测试证据。

本报告研究指定端点 [`https://models.dev/catalog.json?type=all`](https://models.dev/catalog.json?type=all)。只使用公开目录、官方文档和源码，以及本地既有 [catalog-snapshot.json](../../ai/release/catalog-snapshot.json)。未请求真实 Provider 模型 API，未发起模型推理，未运行产品测试或构建。此前的 [Pi 背景研究](research.md)保留为历史资料。

## 1. 结论与证据边界

1. **可以将 models.dev 作为主要维护源，但不能将其响应直接替换为 barness Catalog。** 名称、能力、限制和部分 token 价格可自动映射；协议适配、推理档位接线、图片参数与计费，以及 barness 特有能力仍需确认。
2. **指定 URL 当前提供 v1 兼容输出。** 上游用 `schema = 2` 的 TOML 维护模型，再投影到 v1。不能把 `schema-v2.ts` 的所有字段当作该端点实际返回的字段。[翻译实现](https://github.com/anomalyco/models.dev/blob/514482b88d15d86772042f826231dbf3bc82fabe/packages/core/src/translate.ts#L132)
3. **模型本身与某 Provider 提供的服务型号是两类数据。** 应以 `providers[provider].models[model]` 的 offering 为映射主记录；`models[canonicalID]` 只补充通用资料，不能提供其他托管商的价格或路由。[官方说明](https://github.com/anomalyco/models.dev/blob/514482b88d15d86772042f826231dbf3bc82fabe/README.md#adding-model-metadata)
4. **首次迁移必须先无损保留 118 条现有记录。** 112 条聊天记录有同 Provider、同 ID 的源 offering，但仍存在 API 与字段差异；3 条图片、3 条分类记录均没有准确的服务来源覆盖，不能由上游重新生成后覆盖当前定义。详见第 7 节。

源码依据固定到官方仓库提交 [`514482b88d15d86772042f826231dbf3bc82fabe`](https://github.com/anomalyco/models.dev/commit/514482b88d15d86772042f826231dbf3bc82fabe)。现场响应与该提交的定义相符，**没有证据证明线上部署恰为这个提交**。端点本身没有返回部署提交或目录版本。

本次会话曾对指定 URL 作无认证只读 GET，并在内存解析，未将下载内容落盘。留存的首次响应观测为：

| 项目 | 实际观测 |
| --- | --- |
| 根对象 | `providers` 与 `models` 两个集合 |
| Provider 数 | 235 |
| canonical model 数 | 617 |
| 原始响应字节数 | 6,513,163 |
| 原始响应 SHA-256 | `8c1a562dbe98e75ddf36ac064dfa3d55bce8d812d75af3f3c67506a27a674a35` |
| 精确抓取时间 | 工具记录未固定为可复核的时间字段；不补造时分秒 |

该摘要仅标识首次观测，不代表现在重新请求仍会得到相同字节。第 7 节的 118 条覆盖核对使用了后续独立 GET，该次未计算摘要、未与首份响应比对，也未保留原始字节；不能将覆盖结果绑定到上面的精确摘要，或宣称现有报告留存了可重放的完整输入。本报告没有为落盘再次拉取新快照。模型参数是 models.dev 的记录，本次未逐项向厂商验真。未来维护 CLI 须按设计固定实际参与映射的输入与摘要。

## 2. 指定端点的实际结构

```ts
type Catalog = {
  providers: Record<ProviderID, Provider>;
  models: Record<CanonicalModelID, ModelMetadata>;
};
```

`models["openai/gpt-5-mini"]` 表示底层模型资料；`providers.openai.models["gpt-5-mini"]` 表示 OpenAI 提供的服务型号。同一底层模型可以由多个 Provider 提供不同服务限制和价格。offering 的 `canonical_model_id` 可缺省：现场 OpenAI `gpt-5-mini` 未带该字段，但 canonical 集合存在 `openai/gpt-5-mini`。不能把关联缺省当成模型不存在。[生成逻辑](https://github.com/anomalyco/models.dev/blob/514482b88d15d86772042f826231dbf3bc82fabe/packages/core/src/generate.ts#L33)

### 2.1 Provider

| 字段 | 类型与必填性 | 含义 |
| --- | --- | --- |
| `id` | string，必填 | models.dev 的 Provider ID |
| `name` | 非空 string，必填 | 显示名称 |
| `env` | string[]，至少一项，必填；元素仅要求 string | 其 SDK 认证使用的环境变量名称；不是凭据值，源 schema 未要求元素非空 |
| `npm` | 非空 string，必填 | AI SDK 包名；不是 barness API 枚举 |
| `api` | string，可选 | 基础 URL；缺省可能由对应 SDK 决定 |
| `doc` | 非空 string，必填 | Provider 文档地址 |
| `models` | Record<string, Model>，必填 | 以服务型号 ID 为键的 offering 集合 |

`api` 的有无受 SDK 规则约束，例如 `@ai-sdk/openai-compatible` 要求它，OpenAI 和 Anthropic 可缺省。不能因字段缺省就认为该 Provider 没有 API。[Provider 定义与校验](https://github.com/anomalyco/models.dev/blob/514482b88d15d86772042f826231dbf3bc82fabe/packages/core/src/schema.ts#L366)

现场 OpenAI Provider（仅移除 `models` 集合）：

```json
{
  "id": "openai",
  "env": ["OPENAI_API_KEY"],
  "npm": "@ai-sdk/openai",
  "name": "OpenAI",
  "doc": "https://platform.openai.com/docs/models"
}
```

### 2.2 Provider offering

下表针对 `providers[provider].models[model]`。标记“必填”表示固定源码的输出定义要求，不等于每个字段适用于每种模型。

| 字段 | 类型与必填性 | 含义及解释边界 |
| --- | --- | --- |
| `id` | string，必填 | Provider 接受的服务型号 ID |
| `type` | 枚举，可选 | 模型类型，见下文 |
| `name`、`description` | 非空 string，必填 | 名称与描述 |
| `family` | 模型家族枚举，可选 | 分组资料，不作为 barness 适配规则的唯一依据 |
| `attachment` | boolean，必填 | v1 是否接受非文本输入，不表达 MIME 或尺寸限制 |
| `reasoning` | boolean，必填 | 是否支持推理 |
| `reasoning_options` | 判别联合数组，条件必填 | 推理控制方式，见第 3 节 |
| `tool_call` | boolean，必填 | 工具调用能力，不代表 strict mode |
| `structured_output` | boolean，可选 | 结构化输出能力，不代表严格工具调用 |
| `temperature` | boolean，可选 | 是否支持 temperature 参数 |
| `interleaved` | `true` 或 `{field}`，可选 | 推理相关兼容声明；`field` 为 `reasoning_content` 或 `reasoning_details` |
| `knowledge` | 日期 string，可选 | 知识截止日期 |
| `release_date`、`last_updated` | 日期 string，必填 | 型号日期，不是整份目录版本 |
| `modalities` | `{input: Modality[], output: Modality[]}`，必填 | 粗粒度输入/输出模态 |
| `open_weights` | boolean，必填 | 权重是否开放 |
| `limit` | 对象，必填 | `context`、`output` 必填，`input` 可选；均为非负数，表示 token 限制 |
| `status` | `alpha/beta/deprecated`，可选 | 上游生命周期记录；不能自动触发 barness 下架 |
| `cost` | 对象，可选 | 基础 token 成本及上下文阶梯，见第 4 节 |
| `provider` | 对象，可选 | 本型号覆盖的 `npm/api/shape/body/headers`；各成员可选 |
| `experimental.modes` | 模式名到对象的映射，可选 | 模式可含自身 `cost` 与请求 `body/headers` 覆盖 |
| `canonical_model_id` | string，可选 | canonical 资料关联 ID |

日期允许 `YYYY-MM` 或 `YYYY-MM-DD`。`Modality` 为 `text/audio/image/video/pdf`。可选字段通常是省略，不是 `null`；`limit.output=0`、`cost.output=0` 是现场存在的值，不得当成缺失。[Model、日期和限制定义](https://github.com/anomalyco/models.dev/blob/514482b88d15d86772042f826231dbf3bc82fabe/packages/core/src/schema.ts#L108)

v1 类型为 `chat/image/video/embedding/reranking/decision/transcription/speech/realtime`。过滤器把缺省 `type` 当作 `chat`，默认请求排除 `decision`；`type=all` 才包括它。但上游 v2 的 `audio` 历史兼容为无 `type` 输出；`v2_only` 型号则被排除。因此无 `type` 只适合生成聊天候选，不能直接确认。[过滤规则](https://github.com/anomalyco/models.dev/blob/514482b88d15d86772042f826231dbf3bc82fabe/packages/core/src/filter.ts#L1)、[类型兼容例外](https://github.com/anomalyco/models.dev/blob/514482b88d15d86772042f826231dbf3bc82fabe/packages/core/src/translate.ts#L73)

### 2.3 canonical metadata

canonical 的必填字段只有 `id/name/description`。可选字段包括 `type/family/attachment/reasoning/tool_call/structured_output/temperature/knowledge/release_date/last_updated/modalities/open_weights/limit/license/links/weights/benchmarks`。它不包含 offering 的价格、推理控制或服务路由。引用、权重和评测适合展示/溯源，不进入现有模型请求定义。[metadata 定义](https://github.com/anomalyco/models.dev/blob/514482b88d15d86772042f826231dbf3bc82fabe/packages/core/src/schema.ts#L210)

## 3. 推理控制与 null

`reasoning_options` 可以包含：

| `type` | 其他字段 | 解释 |
| --- | --- | --- |
| `toggle` | 无 | 提供推理开关控制 |
| `effort` | `values` 数组 | 允许的 effort 值 |
| `budget_tokens` | 可选 `min/max` | 推理 token 预算范围；v1 允许 `min >= -1`、`max >= 0`，同时存在时 `min <= max` |

v1 effort 值允许 `null/none/minimal/low/medium/high/xhigh/max/default`。已读官方定义与说明**没有明确解释 `null` 的 wire 语义**；该提交的 v2 effort 枚举实际不含 `null`。结论只能是“v1 允许 null，具体 wire 语义待人工确认”，不能与 barness `thinkingLevelMap` 中“此档位不支持”的 `null` 等同。[v1 推理定义](https://github.com/anomalyco/models.dev/blob/514482b88d15d86772042f826231dbf3bc82fabe/packages/core/src/schema.ts#L23)、[v2 effort 定义](https://github.com/anomalyco/models.dev/blob/514482b88d15d86772042f826231dbf3bc82fabe/packages/core/src/schema-v2.ts#L136)

`reasoning=true` 要求 `reasoning_options` 存在，但可以是空数组；`false` 时不允许该字段。不能据空数组推导完整档位，也不能把 `none`、省略参数、关闭开关与预算特殊值擅自合并。barness 的统一档位到协议值仍需单独确认。

## 4. 成本单位与 v1 信息损失

v1 `cost` 中 `input/output` 必填，其余 `reasoning/cache_read/cache_write/input_audio/output_audio` 可选，值为非负数。字段统一表达 **USD / 百万 token**。缺省表示该响应未提供该项，`0` 表示明确零值；`cost` 缺省不代表免费。[v1 成本说明](https://github.com/anomalyco/models.dev/blob/514482b88d15d86772042f826231dbf3bc82fabe/README.md#L204)

上游 v2 可表达 token、character、image、megapixel、second、minute 单位；token/character 为每百万，其他为每单位。它还提供条件覆盖、乘数、不同缓存 TTL 和缓存存储价格等。[v2 成本定义](https://github.com/anomalyco/models.dev/blob/514482b88d15d86772042f826231dbf3bc82fabe/packages/core/src/schema-v2.ts#L430)

指定 v1 端点只输出部分信息：媒体专属成本和字符计费不会冒充 token 价格；service tier、区域等条件通常不进入基础成本；缓存写价若按 TTL 列表维护只取第一项。图片条目可能有文本 token 成本，也可能完全无 `cost`，两者都不足以构造 barness 的完整图片计费定义。

现场 `gpt-6.1-sol` 成本原样如下：

```json
{
  "input": 2,
  "output": 10,
  "cache_read": 0.1,
  "cache_write": 2.5,
  "tiers": [{
    "input": 4,
    "output": 15,
    "cache_read": 0.2,
    "cache_write": 5,
    "tier": {"type": "context", "size": 272000}
  }],
  "context_over_200k": {
    "input": 4,
    "output": 15,
    "cache_read": 0.2,
    "cache_write": 5
  }
}
```

`context_over_200k` 是兼容别名，实际门槛以上述 `tiers[].tier.size` 为准，不得重复计算。转换会将上游 `gt/gte` 都压到 `size`，因此仅靠该 JSON 无法证明 barness `inputTokensAbove` 的严格 `>` 边界；迁移时保留原定义，新导入时核对原始条件或官方说明。[成本投影](https://github.com/anomalyco/models.dev/blob/514482b88d15d86772042f826231dbf3bc82fabe/packages/core/src/translate.ts#L261)、[兼容别名](https://github.com/anomalyco/models.dev/blob/514482b88d15d86772042f826231dbf3bc82fabe/packages/core/src/generate.ts#L285)

## 5. 真实服务型号裁剪

以下示例保留现场键名，仅移除与说明无关的字段。

OpenAI `gpt-5-mini`：

```json
{
  "id": "gpt-5-mini",
  "name": "GPT-5 Mini",
  "reasoning": true,
  "reasoning_options": [{"type": "effort", "values": ["minimal", "low", "medium", "high"]}],
  "tool_call": true,
  "structured_output": true,
  "temperature": false,
  "modalities": {"input": ["text", "image"], "output": ["text"]},
  "limit": {"context": 400000, "input": 272000, "output": 128000},
  "cost": {"input": 0.25, "output": 2, "cache_read": 0.025}
}
```

Google `gemini-3-pro-image`：

```json
{
  "id": "gemini-3-pro-image",
  "type": "image",
  "name": "Nano Banana Pro",
  "reasoning": true,
  "reasoning_options": [{"type": "effort", "values": ["low", "high"]}],
  "modalities": {"input": ["text", "image"], "output": ["text", "image"]},
  "limit": {"context": 65536, "output": 32768},
  "cost": {"input": 2, "output": 120},
  "canonical_model_id": "google/gemini-3-pro-image"
}
```

这里的 `output: 120` 是 token 价格，不能转成单张图片价格。现场 XAI 图片 offering `grok-imagine-image` 等没有 `cost`。Anthropic `claude-sonnet-5-5` 使用同一 offering 结构，记录 text/image/pdf 输入、推理 effort 和缓存读写价；不需要为该 Provider 另造一套元数据 schema。

OpenCode 提供的 `jev-1.13`：

```json
{
  "id": "jev-1.13",
  "type": "decision",
  "reasoning": false,
  "tool_call": false,
  "structured_output": true,
  "modalities": {"input": ["text", "image"], "output": ["text"]},
  "limit": {"context": 64000, "output": 0},
  "provider": {"npm": "@ai-sdk/typesafe-ai", "api": "https://opencode.ai/inference/systemone/v1"},
  "cost": {"input": 0.042, "output": 0},
  "canonical_model_id": "typesafe/jev-1.13.0"
}
```

decision 输出模态在 v1 被投影为 `text`，不能因此归类 chat。这条成本属于 OpenCode 服务，不是 TypeSafe 直连的价格证明。

## 6. 多 API 与地区标识

上游 v2 可维护多 API，v1 只投影第一项。`provider.shape` 只有 `responses/completions`；API path 及其他协议能力无法完整保留。`npm` 加 shape 可提供默认协议线索，不能证明 barness 当前全部 API 组合，尤其不能据此删掉同型号的其他 API 条目。[默认 API 投影](https://github.com/anomalyco/models.dev/blob/514482b88d15d86772042f826231dbf3bc82fabe/packages/core/src/translate.ts#L213)

现场地区标识如下；按现有地区身份分别关联，不按品牌合并：

| barness | models.dev | models.dev 基础地址 |
| --- | --- | --- |
| `moonshot` | `moonshotai` | `https://api.moonshot.ai/v1` |
| `moonshot-cn` | `moonshotai-cn` | `https://api.moonshot.cn/v1` |
| `zai` | `zai` | `https://api.z.ai/api/paas/v4` |
| `zhipu` | `zhipuai` | `https://open.bigmodel.cn/api/paas/v4` |

其他已有 Provider 同名匹配；`typesafe` 例外，见下节。公共源的 `env/body/headers` 不应直接进入租户 Binding 或凭据配置；来源资料与业务装配边界保持独立。

## 7. 118 条初始迁移覆盖

核对输入为 [catalog-snapshot.json](../../ai/release/catalog-snapshot.json)，目录版本 `2026-10-29.1`，外层记录哈希 `sha256:f44d07c6969346f056cc46d31fccbafb68827267910801c05bb314ac0c55348d`。该哈希是快照声明的既有 Catalog 身份，不冒充文件原始字节摘要。

| 现有集合 | 记录数 | 准确来源核对结果 |
| --- | --- | --- |
| `models` | 112 | 采用上述三项 Provider 别名后，全部命中同服务 Provider、同 model ID；对应 71 个去除 API 维度后的 Provider+ID，尚不证明 API/能力相同 |
| `imageModels` | 3 | 全部缺少准确同 Provider offering |
| `classifierModels` | 3 | 两条服务 Provider 不存在，一条同 ID 命中 chat，均缺少准确分类服务来源 |

聊天记录覆盖分布：OpenAI 51、Anthropic 16、Google 17、DeepSeek 3、Zhipu 6、Z.AI 6、MiniMax 国际/中国各 3、Moonshot 国际/中国各 3、XAI 1。OpenAI 51 是 27 个 Responses 条目加 24 个 Completions 条目，不能理解为 51 个去重型号。

| 缺少准确来源的条目 | 观测与处理 |
| --- | --- |
| openai / openai-images / `gpt-image-2.5-sunburst-2026-09-08` | 同 Provider offering 和 canonical 均无此精确 ID；保留现有来源，不自动改名 |
| google / google-interactions / `gemini-nano-banana-2.1` | 仅有同名 canonical；不能借用其他 Provider 的服务价格 |
| xai / xai-images / `grok-imagine-image-2.0` | 仅有同名 canonical；`grok-imagine-image` 等不是已确认别名 |
| typesafe / typesafe-system-one / `jev-1.13.0` | `typesafe` serving Provider 不存在；canonical 或其他 Provider offering 不构成直连来源 |
| typesafe / typesafe-system-one / `jev-1.14.0` | 同上 |
| openai / openai-decisions / `gpt-6-luna` | 同 Provider/ID 对应 chat offering，类型与 API 不符，不能复用其输出成本 |

即使聊天型号命中也不能直接覆盖。例如快照 `gpt-6.1-sol.contextWindow=272000`，现场上游 `limit.context=1050000`；快照 OpenAI classifier `gpt-6-luna.cost.output=0`，同名 chat offering 为 `0.5`。它们都是实质行为差异，应由维护者确认，不能把源刷新当作无损迁移。

## 8. 映射建议与当前决策衔接

| 分类 | 字段/操作 | 建议 |
| --- | --- | --- |
| 直接候选 | offering `id/name/reasoning`、已支持的 `modalities.input` | 校验类型和模型用途后映射；初迁移保留旧值，差异另列 |
| 名称转换 | `limit.context/output` → `contextWindow/maxTokens`；snake_case token 成本 → camelCase | 数值含义明确时转换，缺省不补零 |
| 类型转换 | `decision` → classifier、`image` → image | 先确认协议与业务操作；其他类型不自动塞进 chat |
| 条件转换 | `cost.tiers` → `inputTokensAbove` 阶梯 | 核对严格边界；忽略重复兼容别名，不忽略无法表达的条件 |
| 显式关联 | Provider ID、服务型号、可选 canonical ID | 保留地区；一个源 offering 可关联多个 barness 四维身份 |
| 人工确认 | `api`、`thinkingLevelMap`、`compat`、图片/classifier `capabilities`、完整图片 `pricing` | 依据 barness 代码、Provider 官方文档/API 或既有人工定义；记录来源与理由 |

需要维护的是 Provider 对照、源 offering 关联、全局字段转换规则，以及每条记录的人工确认/覆盖。没有必要把所有直接字段为每个型号再配置一遍；但不能用 canonical ID 替代 serving Provider，也不能只用 Provider+ID 消除 `(Operation, Provider, API, ID)`。

本次讨论后的当前政策以 [design.md](design.md) 和 [ADR-0058](../../docs/adr/0058-barness-catalog-maintenance-and-isolated-loading.md) 为准：models.dev 是主要源，Provider 官方文档与模型 API 是补充依据；外部 AI 调用确定性维护 CLI，CLI 本身不调用模型。候选模型经**当前受测代码的受控 E2E**检查，未通过时继续修复代码或定义；受控结果不宣称厂商实测、真实可用性或历史版本资格。只维护模型 catalog 内容版本，不增加最低 barness-ai 版本、`schemaVersion` 或旧软件兼容矩阵。本期产品范围为维护 CLI、候选受控检查和本地 JSON 加载；当前仅更新文档，尚未实施，线上发布及在线更新另议。

本次文档更新只核对来源、字段、示例、覆盖计数和引用。产品 E2E、厂商 live、race、压力及发布判定均未执行；本报告不产生上述验证资格。
