# xAI Imagine 官方接口研究

查阅日期：2026-10-29（Asia/Shanghai）。状态：官方资料研究完成；未调用付费 API，未取得本项目的真实调用支持证明。本文区分官方事实、文档冲突和接入建议。barness-ai 的本地模块改动方案由同目录的设计文档承接。

## 结论

Imagine 同时覆盖图像和视频。图像生成与编辑存在独立 REST endpoint，适合接入 barness-ai 的 Image Model 操作。推荐直接选择 `grok-imagine-image-2.0`，以同步 JSON 请求和 `b64_json` 响应作为首期候选路径。视频、Responses 内的 `image_generation` 托管工具、文件托管与 deferred 作业是独立能力，不因接入 image 操作而自动获得支持。[Imagine 总览](https://docs.x.ai/developers/model-capabilities/imagine)、[图像 REST](https://docs.x.ai/developers/rest-api-reference/inference/images)、[生成指南](https://docs.x.ai/developers/model-capabilities/images/generation)

存在四处必须显式处理的差异：xAI 编辑使用 JSON；分辨率与宽高比分开指定；质量影响价格且 `auto` 在生成和编辑中选择不同；Provider 直接报告的费用不同于目录估算成本。[编辑指南](https://docs.x.ai/developers/model-capabilities/images/editing)、[生成指南](https://docs.x.ai/developers/model-capabilities/images/generation)、[费用跟踪](https://docs.x.ai/developers/cost-tracking)

## 1. 型号及临近迁移

| 型号 | 官方现状及接入影响 |
| --- | --- |
| `grok-imagine-image-2.0` | 当前推荐的 Text/Image → Image 型号；支持生成与编辑。[型号页](https://docs.x.ai/developers/models/grok-imagine-image-2.0) |
| `grok-imagine-image` | 1.0；仍在价目表，11 月迁移不影响它。不能把它默认为 2.0 的别名。[迁移指南](https://docs.x.ai/developers/migration/imagine-image-quality-nov-2) |
| `grok-imagine-image-quality` | 2026-11-02 退役；之后相同 slug 重定向到 2.0、固定 `quality: low`。不是请求参数 `quality` 被移除。[迁移指南](https://docs.x.ai/developers/migration/imagine-image-quality-nov-2) |
| `grok-imagine-image-pro` | 已转向 quality slug，11 月再随其转到 2.0 low。首期不建议主动引入该旧别名。[迁移指南](https://docs.x.ai/developers/migration/imagine-image-quality-nov-2) |

迁移指南称响应 `model` 表示实际服务型号，但当前 REST 图片响应 schema 没有该字段。建议解析时允许读取已返回的实际型号，并保留“未报告”，不要以请求型号伪造实际服务型号。[迁移指南](https://docs.x.ai/developers/migration/imagine-image-quality-nov-2)、[REST Markdown](https://docs.x.ai/developers/rest-api-reference/inference/images.md)

## 2. Wire 路径与输入

| 项目 | 已核对的官方契约 |
| --- | --- |
| 服务与认证 | `https://api.x.ai`，`Authorization: Bearer <API key>`。[REST 总览](https://docs.x.ai/developers/rest-api-reference/inference) |
| 文生图 | `POST /v1/images/generations`，JSON。[REST](https://docs.x.ai/developers/rest-api-reference/inference/images) |
| 图片编辑 | `POST /v1/images/edits`，JSON。OpenAI SDK 的 `images.edit()` 发 multipart，官方明确不支持该调用方式。[编辑指南](https://docs.x.ai/developers/model-capabilities/images/editing) |
| 单输入 | `image: {url: ...}`；官方示例另带 `type: image_url`，schema 未要求此 type 字段。[编辑指南](https://docs.x.ai/developers/model-capabilities/images/editing)、[REST Markdown](https://docs.x.ai/developers/rest-api-reference/inference/images.md) |
| 多输入 | `images: [{url: ...}, ...]`，与 `image` 互斥；2.0 最多 5 张。顺序有意义，输出默认沿用第一张的比例。[多图编辑](https://docs.x.ai/developers/model-capabilities/images/multi-image-editing) |
| 内容来源 | 公网 URL、base64 data URI、xAI Files `file_id`；多图允许混用。[编辑指南](https://docs.x.ai/developers/model-capabilities/images/editing)、[多图编辑](https://docs.x.ai/developers/model-capabilities/images/multi-image-editing) |
| 媒体类型 | image source schema 指明 JPEG、PNG、WebP；`url` 与 `file_id` 互斥。[官方 OpenAPI 固定版本](https://github.com/xai-org/xai-sdk-ts/blob/51bfc686c74310226d601e6037fa6c2e562244f7/spec/openapi.json) |

接入建议：Host 已提交的 Input Image 字节在 adapter 内编码为 data URI，避免为接入而增加公网图片下载、上传或跨调用 Files 资源生命周期。该建议不表示 xAI 不支持 URL 或 file ID。

示意请求（方案示例，未执行）：

```json
{
  "model": "grok-imagine-image-2.0",
  "prompt": "把画面中的椅子改成蓝色",
  "image": {"type": "image_url", "url": "data:image/png;base64,..."},
  "resolution": "1k",
  "quality": "low",
  "response_format": "b64_json"
}
```

## 3. 参数与边界

| 参数 | 官方事实 | 接入建议或未定事实 |
| --- | --- | --- |
| `n` | generation 为 1–10，省略为 1；edit schema 有 `n`，但无上下限。[生成指南](https://docs.x.ai/developers/model-capabilities/images/generation)、[官方 OpenAPI](https://github.com/xai-org/xai-sdk-ts/blob/51bfc686c74310226d601e6037fa6c2e562244f7/spec/openapi.json) | 不把 generation 的范围直接声明成 edit 已证契约；编辑批量数量与多输入 × 多输出组合待本路径确认。 |
| `resolution` | REST enum 为 `1k`、`1.5k`、`2k`，省略为 `1k`。[REST Markdown](https://docs.x.ai/developers/rest-api-reference/inference/images.md)、[官方 OpenAPI](https://github.com/xai-org/xai-sdk-ts/blob/51bfc686c74310226d601e6037fa6c2e562244f7/spec/openapi.json) | 不能用 `1024x1024` 等 OpenAI size 值直接替代；不是精确像素大小承诺。 |
| `aspect_ratio` | 生成默认 `auto`；支持 `1:1`、`3:4`、`4:3`、`9:16`、`16:9`、`2:3`、`3:2`、`9:19.5`、`19.5:9`、`9:20`、`20:9`、`1:2`、`2:1`、`21:9`、`5:2`、`auto`。[生成指南](https://docs.x.ai/developers/model-capabilities/images/generation) | 多图编辑可显式覆盖；单图编辑另见下方证据冲突。 |
| `quality` | 仅 2.0：`low`、`medium`、`auto`；省略为 `auto`，当前生成选择 low，编辑选择 medium，按实际选择计费。[生成指南](https://docs.x.ai/developers/model-capabilities/images/generation) | 不把省略当 low；是否固定质量属于 Host 产品决策。 |
| `response_format` | 默认 `url`；可用 `b64_json`。[生成指南](https://docs.x.ai/developers/model-capabilities/images/generation) | 首期固定 b64 可直接满足字节结果，减少下载步骤。 |

文档差异的处理：

1. **1.5K**：生成指南只列 1K/2K，但 REST enum、型号页和价格表都有 1.5K。可把 1.5K 纳入候选参数，证据以 REST schema 和价格表为准。[REST Markdown](https://docs.x.ai/developers/rest-api-reference/inference/images.md)、[型号页](https://docs.x.ai/developers/models/grok-imagine-image-2.0)
2. **单图编辑比例**：固定官方 TS OpenAPI 的 `EditImageRequest.aspect_ratio` 描述要求只用于多图编辑，单图不应设置；最新 REST Markdown 仅列 enum，未重复该约束。保守候选方案是单图编辑遇显式比例时拒绝，不静默忽略；这不是所有 API 版本均已证实的事实。[官方 OpenAPI](https://github.com/xai-org/xai-sdk-ts/blob/51bfc686c74310226d601e6037fa6c2e562244f7/spec/openapi.json)、[REST Markdown](https://docs.x.ai/developers/rest-api-reference/inference/images.md)
3. **默认质量**：官方 Python SDK 的 2026-08-18 changelog 仍称缺省 medium，当前生成指南及 TS 类型说明为 auto。REST 接入应采用当前指南，不把旧 changelog 当当前默认值。[Python changelog](https://github.com/xai-org/xai-sdk-python/blob/main/CHANGELOG.md)、[TS 类型](https://github.com/xai-org/xai-sdk-ts/blob/51bfc686c74310226d601e6037fa6c2e562244f7/src/types.ts#L130)、[生成指南](https://docs.x.ai/developers/model-capabilities/images/generation)
4. **quality schema**：官方 OpenAPI 请求 schema 未列 quality；官方指南、迁移指南及 TS 手工类型补充一致确认该参数。不能以 schema 缺失推断参数不支持。[TS 类型](https://github.com/xai-org/xai-sdk-ts/blob/51bfc686c74310226d601e6037fa6c2e562244f7/src/types.ts#L130)、[迁移指南](https://docs.x.ai/developers/migration/imagine-image-quality-nov-2)

## 4. 响应、审核与字节

REST 结果为 `data[]`。每项可含裸 base64 `b64_json`、`url`、可空 `mime_type`；默认 URL 是临时地址，官方未在已查阅图像文档中给出固定 URL 生存期。不得把 deferred 结果保留 24 小时误当图片 URL 必然可用 24 小时。[REST Markdown](https://docs.x.ai/developers/rest-api-reference/inference/images.md)、[生成指南](https://docs.x.ai/developers/model-capabilities/images/generation)

官方 TS SDK说明输出通常 JPEG，部分质量可能产生 PNG；所以不能固定写入 `image/png` 或只按请求期望格式填 MIME。[官方 TS README](https://github.com/xai-org/xai-sdk-ts#image-generation)

接入建议：严格解码 base64，使用图片文件头核对实际 MIME。`mime_type` 缺省时从文件头识别 JPEG/PNG/WebP；类型矛盾、未知格式、空图片和超出宿主限制均明确失败。识别 MIME 是本地转换策略，不是对 Provider 响应字段的保证。

生成指南展示 Python SDK 的 `respect_moderation` 检查；官方 Python 源码在图片内容缺失且该标志为 false 时抛出审核错误，在内容缺失但未标记审核失败时抛出不同错误；这是 gRPC SDK 证据，不能推为 REST 形状保证。REST OpenAPI `GeneratedImage` 未声明该字段。应解析实际返回的明确审核失败信息，但不要要求 REST 必须含该标志，也不要仅凭 `data: []` 就声称“已证明审核拒绝”。空结果必须可区分于成功产图。[官方 Python 源码](https://github.com/xai-org/xai-sdk-python/blob/60188edfd4fadbd5312c3974dfa4343421af14ca/src/xai_sdk/image.py#L90)、[官方 OpenAPI](https://github.com/xai-org/xai-sdk-ts/blob/51bfc686c74310226d601e6037fa6c2e562244f7/spec/openapi.json)

## 5. Usage 与费用

### Provider 报告与目录估算分开

`usage.cost_in_usd_ticks` 是 Provider 报告的该次请求实际计费金额，已含适用折扣。`1 USD = 10^10 ticks`；`400000000 ticks = $0.04`。它应保留为独立的 Provider 报告事实，不混入 barness-ai 的目录 Estimated Cost。[Cost Tracking](https://docs.x.ai/developers/cost-tracking)

REST usage 可以缺省/为空；token 字段可以为空。官方 TS media usage 映射保留 null，并独立增加由 ticks 换算的 `cost_usd`，不是把缺失用量归零。[官方 TS usage 映射](https://github.com/xai-org/xai-sdk-ts/blob/51bfc686c74310226d601e6037fa6c2e562244f7/src/usage.ts)

| 官方 token 字段 | 含义 |
| --- | --- |
| `input_tokens` | 文字与输入图片 token 合计。 |
| `input_tokens_details.text_tokens/image_tokens` | 文字提示词与输入图片分项。 |
| `input_tokens_details.cached_tokens` | text_tokens 的子集，不能再加一次。 |
| `output_tokens` | 重写提示词文字、推理及产图 token 合计。 |
| `output_tokens_details.text_tokens/reasoning_tokens/image_tokens` | 三类输出分项。 |
| `total_tokens` | 输入与输出合计；不可从图片张数伪造。 |

这些字段定义来自 [REST 完整 schema](https://docs.x.ai/developers/rest-api-reference/inference/images.md)。它们说明 token 报告能力，不等于图像按 token 计费；公开价格依然按张及质量/分辨率分档。

### 2026-10-29 价目表快照

单位为美元/张；这些是研究时的公开费率，不能当未来固定费率。[官方价格](https://docs.x.ai/developers/pricing)

| 型号 | 输入图片 | 输出 1K | 输出 1.5K | 输出 2K |
| --- | ---: | ---: | ---: | ---: |
| 2.0 low | 0.01 | 0.04 | 0.05 | 0.06 |
| 2.0 medium | 0.01 | 0.06 | 0.07 | 0.08 |
| 1.0 | 0.002 | 0.02 | 未列 | 0.02 |
| quality（11/2 将迁移） | 0.01 | 0.05 | 0.06 | 0.07 |

公开表没有明确展开 `n>1` 编辑时输入图费用是否按每个输出重复计入；在取得该路径证据前，不把批量编辑总额公式写成精确合同。公开价目表也不能替代 Provider ticks 反映的折扣后金额。建议优先保存报告金额；目录估算明确其适用范围。[官方价格](https://docs.x.ai/developers/pricing)、[费用跟踪](https://docs.x.ai/developers/cost-tracking)

## 6. 同步、deferred 与托管输出

官方指南称图像默认在本次请求完成；`deferred: true` 立即返回 request_id，然后轮询 `GET /v1/images/{request_id}`。它只支持 URL 响应，结果保留 24 小时；pending 为 202，done 和 failed 都可为 200，失败要读 body.error。[生成指南](https://docs.x.ai/developers/model-capabilities/images/generation)

但固定官方 TS SDK 的公开类型移除了 deferred，注释称查询接口尚未 live。这与当前指南存在时序差异，未经过付费调用不能推断账户已可用。首期同步 b64 路径不会依赖该冲突。[TS 类型](https://github.com/xai-org/xai-sdk-ts/blob/51bfc686c74310226d601e6037fa6c2e562244f7/src/types.ts#L143)

`storage_options` 可存到 Files 并返回 file_output；`output.upload_urls` 可要求 Provider PUT 到每张图对应的签名 URL，要求 URL 响应。它们引入持久资源与额外目标授权，不建议夹带进本次字节型 image 适配。[官方 TS README](https://github.com/xai-org/xai-sdk-ts#image-generation)、[REST Markdown](https://docs.x.ai/developers/rest-api-reference/inference/images.md)

## 7. 错误、配额及尚未证明的能力

官方错误页列出 400/401/403/404/415/422/429 等状态，分别涵盖参数/认证/权限/型号或目标/媒体类型/字段格式/限流；429 需要调用方处理限流。具体错误文本不可作为稳定码硬编码。[Debugging Errors](https://docs.x.ai/developers/debugging)

官方 2.0 型号页显示 6 RPS，rate-limits 页面按 tier 列出 6/12/25/50/100 RPS，且 Imagine 配额提升另需联系官方。应用应采用实际绑定账户限制，不能将网页数值变成公共无限重试策略。[型号页](https://docs.x.ai/developers/models/grok-imagine-image-2.0)、[Rate Limits](https://docs.x.ai/developers/rate-limits)

已查阅资料未能充分确认以下事实，应保留为待证项：

- edit 的 `n` 精确范围，以及多输入与多输出组合的完整约束。
- 单图编辑能否覆盖比例；显式 `auto` 与省略是否完全等价。
- 单图编辑省略 resolution 后的实际像素尺寸；schema 写 1K，但实际比例与尺寸需以响应图片为准。
- 输入字节上限、总请求上限和输出字节最坏值；不得套用聊天图片的“20MiB/不限张数”到编辑接口。
- REST 审核失败是否始终含稳定机器码、是否按图返回 respect_moderation，以及混合成功/失败的实际形状。
- 返回 model、quality、resolution 是否始终存在；不能从缺省值倒推实际服务档位。
- mask、透明背景、指定输出编码、seed、任意精确像素尺寸等能力没有取得本次目标型号充分官方证据，首期不得声明支持。

上述未知项不改变同步图片生成/基本编辑的官方路径结论；它们限定候选支持范围，后续由项目自身的真实调用证据补齐。
