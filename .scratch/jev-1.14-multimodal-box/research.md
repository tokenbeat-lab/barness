# Jev 1.14 多模态输入与 Box：协议研究及 barness-ai 初步设计

研究日期：2026-10-16（Asia/Shanghai）。范围：公开协议、官方文档、官方 SDK；本次没有使用凭据或调用收费推理 API。
下文“事实”由一手来源支持；“推断/建议”是 barness 的工程判断；“未确认”表示公开资料不足，不能视为服务保证。

## 1. 结论与发布状态

- **事实**：2026-10-14 发布 Jev 1.14；图片能力同时开放给 `jev-1.13.0` 和 `jev-1.14.0`，Box 仅支持 1.14。图片和 Box 均为实验能力，请求须显式携带 `TypeSafe-Danger-Zone-Images: true`。[发布公告][release]
- **事实**：这次支持的是静态图片输入，未开放音频/视频；Box 是现有 System One 请求中的新问题类型，返回一个框，不是图像生成操作。[图片][images]、[Box][box]
- **事实**：截至研究日，`jev-latest → jev-1.13.0`，`jev-preview → jev-1.14.0`；公告计划发布一周后迁移 latest。应冻结版本 ID，不能将 latest 当作当前已支持 Box 的稳定标识。[型号][models]
- **事实**：官方明确实验格式、答案、限制可能随版本变化。1.14 宣称主要改善多语言表现，公告未给可直接迁移到 barness 中文任务的效果保证。[发布公告][release]、[Box][box]

## 2. 原理：能证实到哪一层

**事实：结构化决策接口。** System One 以一个 `state` 配合多个类型化问题，返回值和概率；官方称 Jev 采用新架构、并行 sampler 及 RLCD（Reinforcement Learning for Calibrated Decisions），而非逐 token 生成自由文本。概率的校准是统计性质，不保证某个答案正确。[首发说明][launch]、[AI primer][primer]

**事实：可组合的图像内容。** 顶层 `assets` 保存图像，内容中的普通字符串引用在服务端变成图像；`state` 图像供所有问题共享，`instructions`/`criteria` 图像仅进入所属问题。Choice、Score、Noul 可用图像定义判定标准；Box 扩展输出空间为几何位置。[图片][images]

**事实：Box 的数值语义。** OpenAPI 将每条边解释为该边预测位置的期望；输出不包含完整位置分布。由此只能读取四个期望坐标，不能恢复概率热图、多个候选框或联合几何分布。[OpenAPI][openapi]、[Python 生成 schema][py-wire]

**推断**：可以把接口理解为“图文状态 + 独立的类型化判定/定位”。但公开 SDK 只实现素材收集、引用序列化、上传和响应解码，不能证明视觉编码器结构、patch 大小、融合方式、坐标 head、loss 或采样细节；本研究不把 CLIP、特定 ViT 或标准目标检测流水线写成 Jev 事实。[JS 素材实现][js-assets]、[Python 请求构建][py-endpoints]

## 3. HTTP 与 assets 精确协议

### 3.1 入口及请求模型

```http
POST https://api.typesafe.ai/v1/systemone
Authorization: Bearer <API_KEY>
TypeSafe-Danger-Zone-Images: true
Content-Type: application/json
```

**事实**：请求必填 `model: string`、`state: string|object|array`、`questions: map<string,Question>`（至少一题）；可选 `assets: map<string,Asset>`，提供时至少一个条目。问题/答案通过同名 key 对应，问题 ID 不参与推理。新增图片无需改为聊天 messages。[API][api]、[OpenAPI][openapi]

| Asset 字段 | HTTP 含义 |
| --- | --- |
| `type` | 可省略；提供则只能为 `"image"` |
| `data_base64` | 图像文件的标准 Base64；带 `=` padding，无换行/空白、无 `data:` 前缀 |
| `format` | Base64 方式必填：小写 `png`、`jpg`、`jpeg`；后两者同义 |
| `part` | multipart 图像 part 的名称；此方式 `format` 可省略 |

**事实**：每张图选 Base64 或 part；同时出现时 API 优先 `part`、忽略 `data_base64`，实际格式从 bytes 检测。不存在公开的 `image_url`、Data URL、远程 URL、`file_id` 或 `image_id` 协议字段。[图片][images]、[OpenAPI][openapi]、[Python schema][py-wire]

JSON 方式的示意（`<标准Base64>` 是占位说明，不是可调用图片）：

```json
{
  "model": "jev-1.14.0",
  "state": {"photo": "<img:cf134309-3d8a-41e4-9be9-75fc2ecaf995>"},
  "questions": {
    "visible": {"type": "noul", "instructions": "Is there a mug in `photo`?"},
    "mug": {"type": "box", "instructions": "The coffee mug in `photo`"}
  },
  "assets": {
    "<img:cf134309-3d8a-41e4-9be9-75fc2ecaf995>": {
      "type": "image", "format": "png", "data_base64": "<标准Base64>"
    }
  }
}
```

该示例根据公开 schema 组合，未发送；同批 Noul 与 Box 互相独立，Noul 不是 Box 的条件执行开关。[OpenAPI][openapi]、[并行问题模式][fanout]

### 3.2 引用替换及作用域

**事实**：引用是字面量匹配，不是 JSONPath、变量语法或必须带特定分隔符的模板语言；引用可作为完整字符串值，也可嵌入更长文本，嵌套 object/array 同样生效。[图片][images]

1. 完整值匹配优先：值恰好等于某引用时，整个值成为该图像。
2. 较长字符串按从左到右匹配，匹配不重叠；`cat` 会误中 `category`，建议每图每请求随机 UUID。
3. 只替换内容字符串值，对象键不替换，包括 question ID、Choice 选项名和普通对象键。
4. 每个 asset 必须在 `state`、问题 `instructions` 或 `criteria` 至少出现一次；未引用返回 400。
5. 文件只上传一次，但每次引用都再次插入模型输入，并计入 expanded request 限制。
6. `state` 的图供所有题读取；题内素材只供该题。共享图放 state 后用反引号路径指代，可减少重复。
7. 响应回显 criteria 时保留原引用文本，例如 Score `legend` 不会变成二进制、Base64 或 Image 对象。

上述均为[图片协议][images]的明文规则；JS 的 `SentContent`/`ScoreLegend` 类型也明确图像内容回显为字符串。[JS 类型][js-types]

**未确认**：前缀重叠引用（如 `ab` 与 `abc`）在同一位置的优先级、匹配 Unicode 归一化与否，以及多轮递归替换均无明文保证。实现不应凭 map 遍历顺序猜测；实验调用宜使用互不包含的高熵引用。

### 3.3 multipart 推荐格式

**事实**：官方推荐 multipart，避免 Base64 约 1/3 的体积膨胀。首 part 必须叫 `request`、承载 JSON（`application/json`），之后才是图片文件 part。[图片][images]

下面示例要求本地 `photo.png` 满足第 4 节尺寸规则，未在本次研究运行：

```sh
curl https://api.typesafe.ai/v1/systemone \
  -H "Authorization: Bearer $TYPESAFE_API_KEY" \
  -H "TypeSafe-Danger-Zone-Images: true" \
  -F 'request={
    "model":"jev-1.14.0",
    "state":{"photo":"<img:cf134309-3d8a-41e4-9be9-75fc2ecaf995>"},
    "questions":{
      "visible":{"type":"noul","instructions":"Is there a mug in `photo`?"},
      "mug":{"type":"box","instructions":"The mug in `photo`"}
    },
    "assets":{
      "<img:cf134309-3d8a-41e4-9be9-75fc2ecaf995>":{"type":"image","part":"asset-0"}
    }
  };type=application/json' \
  -F 'asset-0=@photo.png;type=image/png'
```

**事实**：part 名非空且唯一，`request` 保留。每个文件 part 必须恰好被一个 asset 的 `part` 使用；遗漏、多余或多 asset 共享 part 均 400。图片 part 的 MIME 只允许 `image/png`、`image/jpeg`、`image/jpg`、`application/octet-stream` 或省略。重用同一图应重复内容引用，不能共享 part。[图片][images]

**事实**：SDK 将同一 Image 放在多个资产别名时，为每个别名上传独立 part；这是“一素材多次引用”与“一份 bytes 多资产”的区别。[JS encodeAssets][js-assets]、[Python with_assets][py-endpoints]

## 4. 图像限制、预处理、上下文和计价

以下限制均为研究日[图片文档][images]的边界，包含端点；超过界限拒绝，不自动缩放；任意一图失败则整次请求无部分答案。

| 资源 | 限制 |
| --- | --- |
| 图片格式 | PNG/JPEG 静态图；APNG 拒绝 |
| 单文件 | ≤16 MiB；Base64 按解码后计算 |
| 每条边 | 256–1568 px |
| 总像素 | 78,400–2,458,624；256×256 不合格，280×280 合格 |
| HTTP body | ≤32 MiB，含 JSON/Base64/multipart framing，不含 headers |
| 全部独立图片 | 服务处理/重新编码后合计 ≤32 MiB，可能大于上传文件总和 |
| Expanded request | 引用展开后，文本/问题/结构/重复插入图片合计 ≤32 MiB |
| 引用 key | 每个 ≤1024 UTF-8 bytes，全部合计 ≤65,536 bytes；非空且非全空白 |

**事实**：准备图片时应先将像素转正、缩放至范围内，必要时放大或补边；API 去除 EXIF、色彩 profile、注释等元数据，但不应用 EXIF orientation。常见风险是本地显示直立、上传后实际横倒；官方建议标准 8-bit baseline JPEG，渐进 JPEG 先重存，GIF/WebP/HEIC/视频先转换/抽帧。[图片预处理][images]

**事实**：gzip（含 x-gzip）或 zstd 仅允许一层；zstd window ≤8 MiB，32 MiB 同时约束压缩体和解压体，压缩不能规避容量。[API][api]

| 上下文 | 默认 | 显式 Long Context |
| --- | --- | --- |
| per-request | state + **所有问题之和** ≤64k tokens | ≤262k |
| per-question | state + 任一单题 ≤32k tokens | ≤262k |

**事实**：两种限制同时生效；官方称 state 摄入一次，问题并行独立评估，问题数没有单独上限（仍受 token/body 限制）。`TypeSafe-Danger-Zone-Long-Context: true` 对 1.13/1.14 都生效，和 Images header 分别开启；长上下文更慢且准确率下降。[型号][models]

**事实**：两版本均 $0.042/百万输入 tokens（$42/十亿）；输出 tokens 免费。图像计入 `usage.input_tokens`、上下文和既有组织限流，与文本同输入单价，无额外图像收费。官方尚未公布图像 token 估算公式。[型号][models]、[图像 tokens][images]

**未确认**：没有公开“分辨率 → tokens”的公式、patch 规格、图片乘数或可靠预估接口。不能套用其他 Provider 的图像算法；同一引用多次展开会重复图像输入，但其精确 token/费用关系仍以实际 reported usage 为准。

## 5. Box 请求、响应及几何语义

**事实**：Box wire 只有 `type: "box"` 和 `instructions`；没有 `criteria`、`image_id`、`target_asset`、`max_boxes`。指南要求说明“找什么、在哪张图”，支持按 state 路径指明搜索图，题内另放图像可作待匹配目标的视觉样例。[Box][box]、[Python Box 输入][py-questions]

**事实**：OpenAPI/两 SDK 允许 instructions 省略或 null，指南则要求描述目标；OpenAPI `required` 仅含 `type`。有效语义是否能完全省略说明不能从 schema 推定；barness 可继续采用自身非空说明约束，并标明这是主动缩窄。[OpenAPI][openapi]、[JS Question][js-types]、[Python 输入][py-questions]

示意响应（数值为示例，不是本次观测）：

```json
{
  "model": "jev-1.14.0",
  "answers": {
    "visible": {"type": "noul", "noul": 0.96},
    "mug": {
      "type": "box",
      "box": {"top": 0.25, "bottom": 0.82, "left": 0.44, "right": 0.84},
      "confidence": null
    }
  },
  "usage": {"input_tokens": 290, "output_tokens": 21}
}
```

- **事实**：`type`、`box`、`confidence` 都必填；`confidence` 是 number 或 null，null 表示未提供置信信息；`box` 本身不可 null，四个 number 边均必填。[OpenAPI][openapi]、[Python 响应][py-responses]、[JS 类型][js-types]
- **事实**：左上角 `(0,0)`，右下角 `(1,1)`；top/bottom 相对高度，left/right 相对宽度，范围 [0,1]；指南承诺 left<right、top<bottom。转换像素时各横坐标×宽、纵坐标×高；crop 或 resize 的坐标还原由应用完成。[Box][box]
- **事实**：每题只返回一个框；没有 not-found，目标缺失也产生四边。应同批问存在性 Noul，由 Host 决定何时忽略框；置信度不是存在概率，官方未提供推荐阈值。[Box][box]
- **事实**：答案不含对应 image 引用；多图场景靠 instructions 自然语言/路径指定搜索图，响应没有机器可验证的图像归属字段。[Box][box]、[OpenAPI][openapi]
- **推断**：严格响应校验需自行检查有限数、范围和正面积，不能仅依赖 schema；OpenAPI 只写 number，未编码 minimum/maximum 或边关系。类型正确也不能证明定位正确，更不能用“总会返回框”推断物体存在。

## 6. 错误与 SDK 实际实现

| HTTP | 新能力相关含义 |
| --- | --- |
| 400 | 缺实验 header；坏文件/格式/Base64/引用/multipart；尺寸低于最小值；`max_tokens_exceeded` |
| 413 | 文件、边长/面积上限、body、处理后图片合计、展开体或引用字节超限 |
| 415 | 不支持/多层 Content-Encoding |
| 422 | schema 校验失败；`detail[]` 中有 loc/msg/type，可能包含原始 input/ctx |
| 401/429/529 | 认证失败/组织限流/服务过载 |

状态依据[图片错误][images]、[API 错误][api]；422 结构见[生成 schema][py-wire]。错误可能回显输入，不能把厂商错误正文直接进入租户外日志。barness 继续由 Binding Retry 控制有界重试，不因 SDK 默认策略扩增尝试。

**事实**：JS 0.7.0（Node≥22）和 Python 0.7.5 于 2026-10-14 加入图片/Box；JS 接收 bytes/Blob，Python 接收 bytes/path/PIL，SDK 接口的“图片对象嵌入 JSON”先序列化成随机 `<image:32hex>` 引用并收集上传文件，不是把语言内存对象直接发给服务器。[JS changelog][js-changelog]、[Python changelog][py-changelog]、[素材实现][js-assets]

**事实**：JS `withAssets` 跟踪插值素材，Python `attach_assets` 维持显式作用域；两 SDK 最终发 multipart，首 request 后各 asset-N 文件 part。SDK 不会因出现 Image/Box 自动添加实验 header，调用者仍须显式传 headers。[JS client][js-client]、[Python endpoints][py-endpoints]、[官方图片集成测试][py-integration]

**事实**：JS 素材去重/未引用预检扫描已序列化 JSON，并非服务端值级替换的完整复刻；素材只做签名识别，未识别类型标 octet-stream 交给服务端。Python 的 from_base64 可忽略空白后解码，发送的却是 multipart bytes，不能据此放宽 HTTP `data_base64` 的规则。[JS assets][js-assets]、[Python assets][py-assets]、[Python 素材测试][py-asset-tests]

## 7. 一手资料之间的不一致与未确认项

| 观察 | 结论 / 对 barness 的影响 |
| --- | --- |
| Models 表仍标 Input=Text；System One 页甚至写不支持图片 | 文档滞后；以 1.14 公告、Images 专页、OpenAPI、SDK 交叉确认图片能力。[models][models]、[旧概述][system-one] |
| Box 指南仅描述 number confidence；OpenAPI/两 SDK 允许 null | 保留“字段必填、值可 null”，不得把 null 改成 0。[box][box]、[py][py-responses]、[js][js-types] |
| OpenAPI Box 示例用 `v14_hot_mode` 且 PNG 是 16×16 | 内部型号名/不合公开尺寸的陈旧示例；不可照抄成有效 fixture。[OpenAPI][openapi] |
| OpenAPI instructions nullable/optional，指南称必填说明 | 请求采用已知可解释、非空的子集；后续用自身 route 证据决定是否扩大。[OpenAPI][openapi]、[Box][box] |
| SDK usage 允许 null；Python 可缺省，OpenAPI 仍 required integer | 用量完整性不得把缺失/null 当 0；需重新核对 native decoder 接受边界。[JS 类型][js-types]、[Python 响应][py-responses] |
| OpenAPI 未编码尺寸、坐标/关系、Box criteria 禁止及诸多语义限额 | schema 不是完整验证器；请求/响应还需显式领域校验。[OpenAPI][openapi]、[图片][images] |
| Score 既有漂移：OpenAPI minItems=1、SDK JS≥2、指南 2–10 | 图片扩展不能顺手放宽 barness 已冻结的题目约束。[Python schema][py-wire]、[JS questions][js-questions]、[API][api] |

待验证清单：图像 tokens 公式与复用计费细节；重叠引用 tie-break；仅题内含图能否满足 Box 的所有服务端条件；1.13+Box 的精确错误形状；零/负面积与边界置信 wire；格式标识和 bytes 不一致时行为；各种图片错误的稳定 code/usage presence；多图搜索图误指时响应；中文定位/存在判定实际质量。**这些未在本次运行收费 API 验证，不能标记支持矩阵通过。**

## 8. barness-ai 统一图片与引用设计（2026-10-16 修订）

**设计状态：proposed，尚未实施。** 用户明确要求图片及引用归一为 barness-ai 定义。本节取代最初“Images 给 Decisions、Assets 给 Jev”的草案；协议事实仍见第 1–7 节，完整交付以 [design.md](design.md) / [spec.md](spec.md) 为准。规格已按本地 issue tracker 标为 ready-for-agent；该标签不表示已经实现。

### 8.1 当前接缝与领域结论

现有 [ClassifierRequest](/Users/cyber/RestoX/barness/ai/classifier.go:11) 已有 State、Questions、Images；[TypeSafe 路径](/Users/cyber/RestoX/barness/ai/typesafe.go:78) 只支持原来的文本 wire，[Decisions 路径](/Users/cyber/RestoX/barness/ai/openai_decisions_request.go:41) 把有序图片转换成 user content。[目录图片检查](/Users/cyber/RestoX/barness/ai/classifier_capabilities.go:21) 尚只允许 Decisions；新增 Jev 能力需要修改真实转换和最终 guard，不能只改目录。

继续使用 Classify 和现有租户运行时。共同语义是对一份状态回答具名类型化问题；Box 可与其他问题同批，不新增模型操作。[领域术语](/Users/cyber/RestoX/barness/GLOSSARY.md) 已明确 Input Image、Image Reference、Image Placement、Bounding Box；这描述设计语言，不作运行时支持声明。

### 8.2 公共定义

| 定义 | 统一契约 |
| --- | --- |
| 图片来源 | 仅 ClassifierRequest.Images，复用 Image{Data,MimeType}；没有公共 Assets |
| 图片引用 | 新增 ImageReferences；每条含必填 imageIndex、placement，以及 value 独有的 pointer |
| context | 所有问题共享的有序图片，按引用声明次序使用 |
| value | RFC 6901 JSON Pointer 指向公共请求中的现存 null 槽位；槽位与引用共同成为图片节点 |
| 默认简写 | 引用省略/null/空时每张图按来源顺序各作一次 context |
| 完整声明 | 非空引用只采用已声明用途；每个来源须被使用，禁止重复 value 目标和授权字段位置 |
| 复用 | 同一来源可以多次引用，重复出现计展开预算；同 bytes 不自动合并来源 |
| 结果引用 | Score 图例图片还原为 null 与 LegendImageReferences；索引指向最终 ImageInputs MIME/hash 元数据 |

value 允许 State、题内 Instructions、Choice / Score / Bool 单个判定说明内的值。普通文字和未被引用的 null 保持原义；不会把 UUID、URL、特殊对象或宿主自定义字符串解释为图片。句中图片使用有序内容数组，首期不提供字符串偏移或字面模板匹配。具体 JSON / 类型形状见设计第 2 节。[JSON Pointer 标准](https://www.rfc-editor.org/rfc/rfc6901.html)

### 8.3 两协议映射

| 公共含义 | Jev | Decisions 本期 |
| --- | --- | --- |
| context | 包裹后的 native state={state:原状态,images:有序私有引用} | 唯一 user 消息，先 State text 再有序 input_image |
| State value | 原槽位换成私有 key，服务端展开 | capability 拒绝 |
| Instructions value | 保留题内图片及作用域 | capability 拒绝 |
| Criteria value | 保留具体候选/等级/Bool 说明位置 | capability 拒绝 |
| 图片编码 | private assets 的 format / data_base64；未来可换 multipart | 内部 data URL |
| 重复引用 | 一个 asset，多处出现 | 每次出现一个图片 block |
| Box | 新增封闭问题和答案 variant | capability 拒绝 |

Jev 私有 key 由 Adapter 生成并检查碰撞，只填明确位置；不在公共字符串内执行 find-and-replace。原生 substring 支持仍是协议事实，但本期最终 Hook guard 也收窄到 whole-value。Jev context 包裹是明确的结构投影，不覆盖原对象中的 state/images 字段；Host 依赖公共字段与图片次序，不填写私有包裹路径，公共契约也不承诺 native 绝对路径稳定。两 Provider 对自然语言解释的效果分别验证。

Decisions 保持已有 compact JSON / 字符串转换、题目转换、两候选下限、2–10 评分及文字说明约束。有任何局部引用就整次在凭据前拒绝，不能把它塞入全局 context。128 上限按实际展开图片 block 计。[官方 Decisions 指南](https://developers.openai.com/api/docs/guides/decisions)

### 8.4 授权、资源和结果

- 资源策略、型号位置/MIME/数量能力、Binding.AllowClassifierImages 三方交集授权；TypeSafe 还需显式 ExperimentalImages 选择。Long Context 首期不开放。
- 复用单图/所有来源 bytes 和数量预算；新增 References 子策略控制出现次数和累计 decoded 展开 bytes。省略子策略时继承既有有限上限，不能得到无界展开。
- Jev 静态 PNG/JPEG 与官方尺寸/16 MiB 单文件/32 MiB body 等约束独立校验；Decisions 既有 WebP 支持不扩散给 Jev。Host 负责旋转、缩放、转码和裁剪。
- 每协议独立冻结最终 native 问题、图片与引用；Hook 换图/换题后按最终请求验答案。重试固定 bytes、headers、key、凭据和目录；2xx 后不重放。
- 图例先严格对照最终 Criteria 再转回公共 null 槽位和引用；最终 ImageInputs 只返回 MIME/hash，不回显 private key 或图片 bytes。Hash 不是权限、不是 Observer 标签。
- Box 四边有限、[0,1] 且正面积；confidence 必填可 null。没有 not-found、存在概率或机器可证明的多图归属，Host 另问存在性并管理坐标使用。
- 坏答案使全部 Answers / ImageInputs 不发布，唯一合法 usage 保留；价格按 reported input_tokens 计一次。Observer 不携带图片、框、引用路径、hash、正文或错误文本。

### 8.5 交付和决策记录

用户已确认以现有 Client.Classify / HookedClient.Classify 为同一个公共验收接缝，连接可控 HTTP Provider，验证映射、结果、授权与资源。先列失败方式、先写 E2E，再实施；结束必须有可复放图片、请求/响应、结果、断言、清单与审计。设计及规格已覆盖错误、取消/重试/并发隔离、图例来源、Box 与旧文本/Decisions 回归。

[拟议 ADR-0030](/Users/cyber/RestoX/barness/docs/adr/0030-barness-ai-unified-classifier-image-references.md) 同步公共来源、位置、协议投影及取舍，明示对 ADR-0020/0021/0026 的扩展；既有 proposed 记录不改变状态。未落地的公共 Assets 草案直接撤销，既有 Images 默认契约继续成立，没有 shim 或并行实现。

本轮没有实施、收费 API、目录/支持矩阵/Pi 更新。后续先用固定 jev-1.14.0 候选验证；图片协议、容量、定位 IoU、目标缺失误报及中文效果分别出证据。Jev context 包裹的效果、多图归属、处理后容量和 token 公式仍未真实确认，不能复用旧版本/另一协议 PASS。回滚沿用 Binding 版本化撤销，仅阻止新调用，已 pin 快照和历史 usage 保留。

## 9. 来源与可复查版本

所有网页于 2026-10-16 访问；动态网页无不可变发布 hash。OpenAPI 为 3.1.0，`info.version=0.2.0`；该版本号不能区分服务所有语义变化。本次读取 OpenAPI 原始 21,610 bytes 的 SHA-256 为 `7914b47cfb1787c0d13c47e6f29b5f33852b00fa0c37986e24ff0ca244fc0e8b`。[OpenAPI][openapi]
SDK 以官方 GitHub 当日浅克隆固定到：JS `b844cd1aca12855a5d24a2df52a1401c86e7d37a`（package 0.7.0），Python `f36e68549ceca3191e5791e5f99ea1baa1a258b5`（pyproject 0.7.5）；下方代码链接均使用 commit 固定。

[release]: https://typesafe.ai/blog/jev-1-14
[launch]: https://typesafe.ai/blog/introducing-system-one-models-and-jev
[primer]: https://docs.typesafe.ai/introduction/machine-learning-primer
[system-one]: https://docs.typesafe.ai/concepts/system-one
[images]: https://docs.typesafe.ai/images
[box]: https://docs.typesafe.ai/primitives/box
[models]: https://docs.typesafe.ai/models
[api]: https://docs.typesafe.ai/api
[openapi]: https://api.typesafe.ai/openapi.json
[fanout]: https://docs.typesafe.ai/patterns/fan-out
[js-changelog]: https://docs.typesafe.ai/sdk/javascript/changelog
[py-changelog]: https://docs.typesafe.ai/sdk/python/changelog
[js-types]: https://github.com/typesafe-ai/typesafe-sdk-js/blob/b844cd1aca12855a5d24a2df52a1401c86e7d37a/src/types.ts
[js-assets]: https://github.com/typesafe-ai/typesafe-sdk-js/blob/b844cd1aca12855a5d24a2df52a1401c86e7d37a/src/assets.ts
[js-client]: https://github.com/typesafe-ai/typesafe-sdk-js/blob/b844cd1aca12855a5d24a2df52a1401c86e7d37a/src/client.ts
[js-questions]: https://github.com/typesafe-ai/typesafe-sdk-js/blob/b844cd1aca12855a5d24a2df52a1401c86e7d37a/src/questions.ts
[py-wire]: https://github.com/typesafe-ai/typesafe-sdk-python/blob/f36e68549ceca3191e5791e5f99ea1baa1a258b5/src/typesafe_sdk/_schemas/models.py
[py-questions]: https://github.com/typesafe-ai/typesafe-sdk-python/blob/f36e68549ceca3191e5791e5f99ea1baa1a258b5/src/typesafe_sdk/_core/question_types.py
[py-responses]: https://github.com/typesafe-ai/typesafe-sdk-python/blob/f36e68549ceca3191e5791e5f99ea1baa1a258b5/src/typesafe_sdk/_core/response_types.py
[py-endpoints]: https://github.com/typesafe-ai/typesafe-sdk-python/blob/f36e68549ceca3191e5791e5f99ea1baa1a258b5/src/typesafe_sdk/_core/endpoints.py
[py-assets]: https://github.com/typesafe-ai/typesafe-sdk-python/blob/f36e68549ceca3191e5791e5f99ea1baa1a258b5/src/typesafe_sdk/_core/assets.py
[py-asset-tests]: https://github.com/typesafe-ai/typesafe-sdk-python/blob/f36e68549ceca3191e5791e5f99ea1baa1a258b5/tests/test_assets.py
[py-integration]: https://github.com/typesafe-ai/typesafe-sdk-python/blob/f36e68549ceca3191e5791e5f99ea1baa1a258b5/tests/test_integration_images.py
