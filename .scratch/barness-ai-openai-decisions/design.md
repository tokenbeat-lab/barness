# OpenAI Decisions 分类：详细设计

Status: ready-for-agent
Updated: 2026-10-12
Baseline: 8ea61b8d70de6c84fe27be34dc7d21c5c40841d9

关联[spec](spec.md)、[验收矩阵](acceptance.md)、[实施计划](implementation-plan.md)、[研究](../pi-1.1.0-openai-decisions/research.md)、[ADR-0026](../../docs/adr/0026-barness-ai-openai-decisions-classification.md)。这是完整设计，不是已实现声明。若执行时基线已变，先检查接缝和既有行为，不能回退他人改动。

## 1. 目标、领域和 Module 责任

Host 拥有身份、授权配置、标注任务、业务阈值及结果处理；barness-ai 拥有一次 Logical Call 的协议转换与生命周期；Provider 拥有模型推断和真实计费。System One 是 API Protocol，Jev 是 Classifier Model。Decisions 与它共同使用 classifier Operation，但有各自的 Protocol Model。

唯一主要执行 Seam 是已确认的 Client.Classify / HookedClient.Classify。公共 Interface 除类型还包括默认值、权限、错误阶段、原生回调、结果发布和资源成本。两个 Adapter 的复杂 wire 处理藏在 Module 内，Host 不需要管理连接、Attempt 或解析数组。

| Module / 内部责任 | 拥有 | 不拥有 |
|---|---|---|
| Classify 公共入口 | scope、已有 shape/byte checks、不可变请求、统一终态 | 按 provider 猜测凭据、业务决策 |
| Catalog | 完整型号身份、输入/问题能力、价格、deep copy/hash | 在线下载、自动授权 |
| callRuntime / initialRequest / sendUnary | Binding/Credential 固定、准入、网络尝试、取消、有限 body、观测 | 题目/答案 wire schema |
| TypeSafe Adapter | 现有 state/questions map、noul、JSON 说明、舍入和 usage | Decisions 的 message/array/refusal |
| Decisions Adapter | input/questions array、最终 wire guard、答案和 usage 转换 | 新 HTTP client、业务 fallback、全局 key |
| Host 示例 / live / gate | 有界场景、状态、质量和审计证据 | 将合成或另一路线结果冒充自身真实 PASS |

依赖方向：Host → 公共 Interface → 操作 runtime + Catalog → 固定协议 Adapter → shared unary transport。内部 TypeSafe 与 Decisions 不互相调用；共用的是已验证的运行时与纯校验能力，没有动态注册或对外 Adapter Interface。

## 2. 公共契约

### 2.1 Request / Options

ClassifierRequest 保留 State 和 Questions，增加 Images []Image，JSON 为 images,omitempty；省略、null 或空数组均为无图片。每图复用已有 Image 的 data/mimeType 字符串字段，不增加公共 type/image_url；公共 decoder 拒绝非对象图片、null/非字符串字段、未知键和大小写别名，typed Go 值则在 Classify scope 校验空/无效值。State 仍必须是合法、无重复键的 JSON string/object/array，包括空字符串、空对象和空数组；图片也不免除 State 要求，图像单独判断可用空对象作为状态。图片顺序有语义；问题 map 无插入顺序承诺。

公共 JSON decoder 负责语义，不拥有 Host 的入站 I/O；云端 Host 在反序列化前对外部 body 自行设置有限读取/解码预算。Classify 的 scope 检查限制模块自己的复制/编码，不冒称能回收 Host 已经分配的巨大请求。

ChoiceQuestion / ScoreQuestion / BoolQuestion 的形状不变。Instructions 仍要求非空字符串/对象/数组；不新增 trim 或空白折叠。Choice 的 null criterion 仍是“名称本身，无额外说明”。Bool.Criteria 省略与显式 null 的旧规则保持。

公共 choice 1–255，score 2–10。保留现有 scope 阶段的上限拒绝、无凭据/无请求/无 Attempt 及错误优先级。OpenAI 的下限/字符串说明另在 capability 阶段加严。不能将公共 10 档声称为 API 上限；本期不扩 11 档。

OpenAIDecisionsOptions 是空的 Options 实现，归属 APIOpenAIDecisions。nil 和空值选协议默认；TypeSafeOptions 在 Decisions 上 invalid_request/capability，反向也相同。非 nil pointer 按现有封闭选项复制；typed nil 不得导致新路线 panic，按 invalid_request/capability 拒绝。协议没有 SimpleOptions 或 stream 入口。

### 2.2 Model / Capabilities

保留 ClassifierModel 现有字段与 keyed literal。新增：

| 字段 | JSON | 规则 / 零值 |
|---|---|---|
| ClassifierModel.Input []Modality | input,omitempty | nil/空表示旧 text-only；显式只允许 text 或 text+image，无重复，必须包含 text |
| ClassifierCapabilities.MinChoices int | minChoices,omitempty | 0 为既有最低 1；负数拒绝；有效最低不得大于 MaxChoices |
| DescriptionFormats []ClassifierDescriptionFormat | descriptionFormats,omitempty | nil/空为现有 string/object/array；枚举与 JSON 根类型一致，拒绝未知/重复 |
| MaxInputImages int | maxInputImages,omitempty | 0 不支持图片；负数拒绝；image 声明要求正值 |
| ImageMIMETypes []string | imageMimeTypes,omitempty | 无图片时须为空；图片型号显式声明首期支持格式的非空子集 |

ClassifierDescriptionFormat 底层为 string；公共常量 ClassifierDescriptionString="string"、ClassifierDescriptionObject="object"、ClassifierDescriptionArray="array"。DescriptionFormats 作用于 instructions 和非 null criterion 说明，表示公共 JSON 格式，不承诺结构化对象可原样成为任意 native label。Choices/Score 范围仍由既有模型字段收窄；声明比公共范围宽不扩张公共入口。

已知 TypeSafe × System One 的 Input 只能 text，实际请求限制仍 1–255 / 2–10，保持已有结构化说明。新 OpenAI × Decisions 目录必须显式 MinChoices>=2、DescriptionFormats=[string]；image 如声明，则 MaxInputImages<=128、MIME 为 PNG/JPEG/WebP 的已验收子集。旧 Jev 省略新增字段仍构造成功，返回和 JSON 保持省略态；不强制为每个旧目录插入默认字段，不用 deprecated shim。

NewClient / Catalog.Hash 共用校验。新增已知协议能力有约束，但不收紧旧 MaxChoices/score 元数据已有的构造接受范围；调用的有效范围仍是原公共范围与声明能力的交集。Input、DescriptionFormats、ImageMIMETypes、Kinds、Cost.Tiers 和三种目录返回边界各自深复制；修改任何新字段改变整个目录 Hash，旧字段和省略态不因默认解释而改写。发现不读凭据、不授予 Binding 权限。

### 2.3 Binding 权限

新增 AllowClassifierImages bool，默认 false。Binding 仍是可信 Host 经 resolver 提供的 Go 配置，本期不新增统一 Binding JSON codec 或改变旧字段序列化；Host 的持久化 DTO 可将权限映射为可省略的 allowClassifierImages，迁移由 Host 负责。只允许 OperationClassifier 的 Binding 开启，其他 Operation 为 invalid_request/binding。它不替代 AllowedModels、Enabled、AccountScope 或 Credential Snapshot。

图片授权要求 Model 支持 image、Policy.Classifier.Images 非 nil、Binding.AllowClassifierImages=true。已有 Jev 与聊天 Binding 没有该字段仍保持行为；仅更新 Catalog 的 Input 不会授予租户图片权限。回调后来添加图片也检查固定 Binding；不能通过 payload 加此开关。

### 2.4 结果和分类拒答

ClassifierResult 不增加外层状态。封闭 ClassifierAnswer 新增 RefusalAnswer 空值类型，其 JSON 为 {"type":"refusal"}；名称由 Answers map 的 key 表达，没有可编造的 reason/message。

| 响应 | Answers | StopReason / error |
|---|---|---|
| 全部普通合法答案 | 最终问题完整名称集 | stop / nil |
| 普通答案与 refusal 混合 | 完整名称集，拒答用 RefusalAnswer | stop / nil |
| 全部 refusal | 每个名称都有 RefusalAnswer | stop / nil |
| 任一协议损坏 | nil；不得发布其他题的半成品 | error / protocol,response |
| 取消 / 截止 / 资源错误 | nil，唯一上报用量及元数据保留 | 既有 aborted/error 与对应 Code/Phase |

nil error 的新路线语义是“协议响应完整且可解释”，不是每题都有 Choice/Score/Bool。迁移示例必须做封闭 type switch 处理 refusal，业务转人工/重问由 Host 执行。TypeSafe native refusal 仍是未知类型 protocol 错误，不能全局放宽 Jev decoder。

ScoreAnswer 保留既有 Score/Confidence/Probabilities/Legend；Decisions 将 native value 的 0..n-1 映射为整数键，将匹配的 label 编码为 JSON string 图例。Refusal 没有概率、confidence 或 score；BoolAnswer 不新增 confidence。

## 3. 生命周期和数据流

~~~mermaid
flowchart TD
  A[Host: scope / target / request / options] --> B[既有 scope、shape、bytes + 新图片策略检查]
  B --> C[复制公共输入；解析和授权 Binding / Model]
  C --> D[协议能力及图片授权；固定 Credential / Catalog]
  D --> E[固定协议分派、header 和 native payload]
  E --> F[Hooks；最终 native guard 与不可变记录]
  F --> G[sendUnary：准入、有界尝试、超时与取消]
  G --> H[OnResponse；bounded body]
  H --> I[先登记唯一 usage / ResponseModel]
  I --> J[整体校验答案、最后发布]
  J --> K[统一结果、元数据、Attempt、Observer 和释放]
~~~

保持 beginCall 的 scope/deadline/Operation 归属。一份状态、图片和全部命名问题是一个 Logical Call，不按问题种类/图片拆请求或内部并发凑答案；依赖前题答案的后题由 Host 发起下一次调用。公共已有限额/shape 检查仍先于 resolver；有图片时先按 known lengths 检查，后复制图片 slice、RawMessage 与各问题 map/slice。Resolver 回调不能再改入口快照。

resolve 先核对操作、Provider×API 路由、AllowedModels 和完整目录身份；capability 检查之后才读取 Credential。新增 fixed switch 执行 TypeSafe 或 Decisions，直接删除原来的唯一 TypeSafe 分派假设。pin 固定凭据/Binding/目录以及重试策略，调用期间不重新读取。

header/payload/response 仍按既有 runtime 契约执行；onPayload 一次，重试发送最终相同字节。2xx 的许可持有到读取、解码、验证、成功发布或失败关闭，所有路径 close body/release。MaxFrameBytes、事件队列、assembler 不参与分类 JSON；MaxOutputBytes 限制完整解码前 body。

## 4. Decisions 请求转换和 native guard

### 4.1 初始转换

以 POST <Binding.Endpoint>/decisions 发送，官方 base 是 https://api.openai.com/v1；Content-Type=application/json，Bearer 仅来自已固定 Credential，沿用现有 user-agent/header 变换和 transport。wire model 为已授权 Model.ID，绝不发送 classifier: 前缀或从 chat 型号推断新 ID。

- State string 解码为原文；object/array 用保留原始数字拼写的 JSON compact，不经 map[string]float64。不是对 state 再加引号；嵌套文字仍按 JSON 编码。
- 无图片：input 是字符串；有图片：单个 user 消息，content 顺序为一个 input_text 与有序 input_image，URL 仅为 data:<MIME>;base64,<data>。不自动读取、拉取或上传任何外部资源。
- choice：value 是公共字符串 key；null 或空字符串说明省略 description，其他说明必须 JSON string。score：按数组顺序构造 levels 的 label，不重新排序，也不合并重复文本 label。
- bool：type=predicate；criteria 存在时依次加非空说明，格式保持 Pi 的双换行及 “True means: ” / “False means: ” 两行，不用两个 boolean choice 代替。
- 问题/候选 map 使用冻结 JSON 对象的确定枚举顺序：规范 JS array-index 名称先按数值，其余按既有冻结键序，避免 Go 随机 range。公共 map 不承诺用户插入顺序；额外数字精度等 Pi 投影差异有限登记。
- body 初期只允许 model/input/questions；不发送流、temperature、reasoning、cache/service tier、metadata、timeout 或 safety_identifier。认证/endpoint 来自 Binding，不读 OPENAI_API_KEY 或代理环境。

### 4.2 Hook 后有效子集

model 必须等于授权型号。input 只接受字符串，或恰好一个 user 消息、一个首位 input_text 和之后有序 inline images；message type 可省略或等于 message。image 只包含 type/image_url；detail、file ID、非 user role、tool/audio/item reference、多消息等不在本期支持子集。

questions 必须是非空数组，数量符合 Policy；name 非空且唯一。choice 为字符串值的唯一候选数组，至少 2 个，说明只为 string；predicate 只有 string instructions；score 的 levels 为 2–10 的有序 label/string 可选 description。回调可以改变名称、题目种类/顺序/说明、state text 和 inline 图片，只要所有约束成立。

将 guard 成功的 wire body 解码为私有 frozen decisions request，保存最终名称/kind、候选、score labels 和图片信息；该记录只存在本次调用，不引用回调对象或入口问题。解析答案依据此记录，不能把 native input 反解析为旧 State 来恢复原始问题。

### 4.3 失败归类

| 情况 | Code / Phase | credential / Attempt / 发送 |
|---|---|---|
| 原公共形状/政策限额 | 既有 invalid_request 或 resource_limit / scope | 0 / 0 / 0 |
| 初始格式/题目/模型模态不匹配 | invalid_request / capability | 0 / 0 / 0 |
| Binding 不允许图片 | tenant_denied / capability | 0 / 0 / 0 |
| 回调改 model、权限/认证/目标/外部资源 | tenant_denied / request | 可能已固定凭据 / 0 / 0 |
| 回调 schema、类型、未知普通字段 | callback_failed / request | 可能已固定凭据 / 0 / 0 |
| 回调资源限额 | resource_limit / request；沿用既有具体 limit | 可能已固定凭据 / 0 / 0 |
| 响应不合契约 | protocol / response | 保留唯一 usage / 已有 Attempt / 不重试 |

authority 字段包含 endpoint/api_key/headers/provider/api/operation/credential/tools 及图片授权开关。外部图片 URL/file ID 表示越权资源请求，为 tenant_denied/request；合法支持子集外的 detail/safety_identifier 是 callback_failed/request。Header auth/account/target 保护沿用现有 bound hooks。

## 5. 图片授权、限制与容量

Policy.Classifier.Images 是新的 ClassifierImagePolicy，字段 MaxImages int、MaxTotalImageBytes int64；nil 关闭。正的有限值构造时验证，总图预算不超过 MaxRequestBytes；全局 MaxImageBytes 继续限制每图的解码字节。Policy.Image 开启或关闭不改变分类图片权限。

初始入口顺序：计数/原始长度/解码长度上界与总请求下界 → 既有状态问题检查 → 图片 MIME/strict standard base64 → 深复制。禁止 CR/LF、URL/path、错误 padding 与 MIME/file header 不符。复用已有 pure base64/file header 验证，不复用 GenerateImages 的编辑/mask/16 张/20 MiB 限制；不分配完整 raster，不自行补 MIME 或 resize。

首期格式是 image/png、image/jpeg、image/webp。安全校验只证明编码和格式 header 合法，不承诺损坏文件可由厂商完整解码；真实文件另由自己的 live 验收。单图 decoded bytes、decoded bytes 总和、base64 及 data URL/JSON 完整编码成本独立计量；所有乘法/加法要检查溢出。

最终请求限制是 Host policy、Model capability、OpenAI 128 上限的交集。Policy 超限 resource_limit/scope，模型/API 能力超限 invalid_request/capability；Hook 后在 request 阶段使用同样完整边界。Guard 必须遍历 user message 中所有图片，不能只检查公共 Images 或第一张图。

## 6. 答案、数值和响应扩展

先使用独立 envelope 保留唯一 usage 和 model；重复字段不 last-wins，重复 usage 不成为权威值。只有一个完整 JSON object，body/深度/字段大小受既有限制，不能返回半截有效答案。

answers 是数组，以 name 匹配，不依赖返回顺序。发送的每题都有非空 name，所以 null/缺 name 无法回配，为 protocol；名称必须唯一且与最终集合严格相等。type 仅 choice/score/predicate/refusal；合法拒答跳过普通数值要求，但不能带可执行或 fabricated 普通答案。

普通答案规则：

- choice 必须字符串且在最终候选内，probability entries 唯一、完整；每项及 confidence 有限且在 0..1，概率和绝对容差 1e-6。选择须属于最高概率项，允许并列，不归一化、不替换 choice。
- predicate probability 必须明确上报、有穷且在 0..1；转 BoolAnswer，名称和类型匹配最终问题。
- score 分布必须完整覆盖整数 0..n-1，每项 label 与最终 label 精确匹配。confidence、概率在 0..1，score 在 0..n-1；和及加权期望绝对容差 1e-6。保留原值，TypeSafe 的两位小数规则不适用。
- 如真实 OpenAI wire 证明有限舍入特例必要，先记录失败回放，再提出 Decisions 专属有界规则及 ADR 补充；不得自动放宽容差或修改 Jev。

Decisions 请求 schema 是封闭的。响应已定义字段使用精确大小写、重复检测、类型和数值检查；额外 vendor 非语义字段可忽略但仍受总字节和重复键约束，大小写别名不能覆盖已定义字段。未知 answer discriminator 或会影响已定义答案语义的字段不忽略。TypeSafe 继续原有严格 decoder，不扩大其未知字段容忍。

ResponseModel 仅保存唯一合法上报值，不覆盖 Metadata.ModelID；无原生 ID 不伪造。x-request-id 来自响应 header；hook failure/body failure/答案错误的唯一已报告用量保留，缺 body 不能推断免费。

## 7. Usage 和 Estimated Cost

不复用 TypeSafe 或 Responses 的 wire usage parser；复用 Usage/UsageCost/Attempt/Price Snapshot。Decisions Input=input_tokens，Output=output_tokens，CacheRead/CacheWrite=0，TotalTokens 为合法报告或已知 Input+Output。Reasoning 的 unset 与明确 0 区分；不把 cache 明细加入 canonical 桶，不新增正文型 Observer 数据。

| 情况 | Reporting / 处理 |
|---|---|
| usage 省略/null | unreported，零兼容值、Reasoning unset；合法答案仍可返回 |
| 各已知字段部分省略/null | 保留唯一有效计数，partial；总数未报按已知项 checked sum |
| 计数及细节全部明确出现，包括 0 | complete；所需字段为 input/output/total、cached/cache_write、reasoning |
| usage 重复、负数、小数、非有限/溢出、错误类型 | protocol/response；ambiguous/invalid usage 为 unreported，不选最后值 |
| 计数内部矛盾 | protocol/response；不输出自相矛盾的完整用量 |

检查 known cache counts 非负、合计不超过已知 Input；reasoning 不超过已知 Output；全 input/output/total 同时已报时总数须一致。cache details 缺失不是免费证明；它只影响 reporting，不扣 Input。token 字段只接受非负整数且不溢出 int64。

费用只按独立 Catalog 的 input 费率估算：Input<=272000 用 .1，Input>272000 全输入用 .2 USD/百万；其他费用桶为 0。边界至少 271999/272000/272001，图像 token 计入 Provider 报告 Input；不由图片张数或 bytes 推导 token。输出非零仍可报告而 output Cost 为零。

区域 premium、缓存统计与实际账单一致性是自己真实核验事项，初期目录注明全球标准处理估算来源；模型/价格 Hash 不代表账单。当前共享账户 Scope 仍固定；不从 endpoint 字符串猜区域费率。真实非零缓存统计没有证据时报告“未观察”，不为获得它重复造流量。

## 8. 错误、重试、取消与观测

default Binding.Retry=0；显式重试继续使用共享运行时，包括 x-should-retry、408/409/429/5xx、retry-after 与有界等待。不新增 Decisions 默认 2 重试，不为了 Pi 的 504 特例复制/修改全局 retry。该差异进入 limited ledger。对 504/其他非 JSON gateway 错误保留状态/request ID，给调用者固定可读诊断，不回显 HTML；JSON error 的明确 message 经有界读取和脱敏后可保留。Observer 不含错误正文或秘密。

CallTimeout 包含 resolver/hook/encode/read/validate；发送沿用 Connect/ResponseHeader/ReadIdle。解析前后都检查 context，图片验证流式检查取消。2xx 后任何故障都不再尝试；重试中的最终 body/credentials/catalog 固定，header/payload hook 一次，response hook 只在最终 2xx 一次。

Observer 保留 CallStarted/AttemptStarted/AttemptFinished/CallFinished 和 Operation 身份。拒答不是协议失败，观察记录为成功 stop；不记录拒答正文、数量明细或画像。Host 从类型化结果形成业务统计；异常路径记录诊断 Code/Phase/HTTPStatus/request ID 和 Reporting。

## 9. Jev 兼容合同和明确取舍

| 必保面 | 断言 |
|---|---|
| 方法/JSON | 旧 keyed declarations、未新增图片的新请求、nil/TypeSafeOptions、closed questions 和 result JSON 继续成立 |
| 输入 | 字符串/对象/数组说明、null choice、空状态、单候选、bool criteria 可省略/提供 |
| 容量/错误优先级 | 公共 1–255 / 2–10 scope 检查保持；模型进一步限制仍 capability；不改变旧 Code/Phase/Attempt/credential 次数 |
| wire | state/questions map、noul、原始 numeric JSON、没有 images/input/Decisions 参数 |
| callback | 固定顺序、合法最终换题、严格未知字段、授权不可扩大、回调失败归类 |
| 响应 | 精确名称/选项/图例、duplicates 拒绝、最高概率并列、类型与概率范围 |
| 舍入 | 已有 Jev .01 网格的有界一致性接受；高精度/越界反例继续拒绝 |
| usage/runtime | 现有输入费率、output/cache Cost 为零、原 Reporting 和 ResponseModel、snapshot、retry、release、Observer |
| 配置默认 | 未加 Input/MinChoices/DescriptionFormats/图片权限的旧 Host 自带目录与绑定仍能运行 |

研究建议把 2–10 从通用 validator 移出，但这样会改变旧 scope 拒绝顺序；本期保留公共范围并明确命名为 barness 支持范围。不存在跨协议补偿性 shim 或为了扩大新 API 临时延后旧输入错误。未来实际需要 11 档时另行评估公共契约和错误顺序。

“同公共方法”不等于任意合法 Jev 请求可以只改 Target。结构化说明和 TypeSafeOptions 在新 API 上正确拒绝；跨协议共同子集是 string 说明、至少两个 choice、2–10 score、bool、无图或获授权的 Luna 图片。

## 10. 来源、候选、目录与发布

来源 pins 保持 Pi 1.1.0 / abe508e1b89912adde45528136c3221eb69acdd7 / 既有 model-data hash。Decisions 是该冻结版本原生路线，不标为没有 upstream 的 Extension Route；新增信息/严格校验是 limited differential extensions。必要的回调时机/故障归类、header 脱敏/Host 扩展也逐 case/path 登记：barness OnResponse 在成功 header 后、body 读取前，保留既有 unary 契约，不为 Pi 在 JSON 解析后的回调顺序改变 Jev。

候选身份：classifier/openai/openai-decisions/gpt-6-luna，Name=GPT-6 Luna，ContextWindow=922000；Input=text,image，choice Min=2 Max=255，score Min=2 Max=10，DescriptionFormats=string；图片上限 128、格式 PNG/JPEG/WebP。能力来源区分 Pi 快照、官方 API 128/choice、barness 2–10/格式支持子集和自身 live；不是声称测过 922K 或所有图片数量。

工单先用合成宿主目录做离线，自己的固定 candidate catalog 做真实证据；live 不允许 endpoint override，candidate 只在真实官方 endpoint 使用明确版本/Hash。通过协议/图片/精度/价格来源验收后再升内置目录版本。正式版本/Hash 与 candidate 不同，必须重跑正式 Decisions live；既有正式 Hash 适用性重验，不能直接把 candidate 报告合并为正式 PASS。

无日期 gpt-6-luna alias 是有来源的显式目录决策；不制造日期型号。元数据价格固定与服务权重可变分别记录，质量报告固定实际 ResponseModel 和执行日期。禁用 Binding / image flag / catalog allowlist 可撤销新调用。

Support Matrix 增第十组合 openai-decisions-classifier，使用该 combo 独立 report validator / expected capability set；Jev 的身份/验收器不被更名或放宽。发布 traceability 固定十组合，拒绝旧 schema/旧 Hash/缩小能力集伪造通过；按新 schema 迁移既有记录时保留其历史状态，不重写时间为当前通过。

本次矩阵及新报告 schema 升为 3。检入矩阵从 schema 2 一次性迁移，保留九行所有历史结论/时间/Hash，并新增 NOT_RUN 行；schema 2 报告不直接 merge 进 schema 3。旧 bundle 仍是历史证据，正式十组合报告以当前 schema/版本生成，回滚用原 schema 的完整 Git 快照，不建立双写或运行时迁移 shim。

候选 live 最多 12 Logical Call、12 Attempt、24 个问题、4 MiB 总状态字节、8 张累计图片、4 MiB decoded 总图片，进程 180 秒、单调用 30 秒、Binding 零重试；预算是本项目测试值。独立中文任务最多 24 调用/Attempt、零重试、串行、总 600 秒；额外服务容量/大输入探针单独 opt-in、最多 2 调用/1 MiB 状态/60 秒，不用于正常 complete 判定。价格长阶梯用离线明确 usage 验证，不主动发 272K+ token 请求。

Live 必交正常文本 choice/score/bool、混合题、图片三格式和 response/usage/ID/归属/取消。拒答输入不必保证线上每次触发；其语义由 public fixture E2E 必交，live 若自然出现按合法拒答记录。不要为了得到 refusal/429/504 循环制造危险内容或故障。坏 key 仅使用合成假值一次、独立预算；无法确认的错误形状注明未确认。

## 11. 实施位置、文档和验收链接

路径只用于定位，责任按上述 Module 定义。执行时核对现状；每文件<=500 行，复杂协议转换按单一职责拆分。

| 责任 | 当前接缝 / 计划位置 |
|---|---|
| Request / Options / Answer | ai/classifier.go、classifier_questions.go、classifier_answers.go、classifier_json.go |
| 能力 / 默认 / 克隆 | ai/catalog_operations.go、catalog_validate.go、catalog_snapshot.go |
| Binding / 分派 | ai/binding.go、call_runtime.go、classifier.go |
| 原生协议 | 保留 ai/typesafe*.go；新增 openai_decisions*.go 的 orchestration/request/guard/response/usage |
| 资源 | classifier_policy.go、classifier_payload.go、policy.go；复用纯 image validation，新增 classifier 图片 guard |
| 公共测试和证据 | ai/e2e、ai/internal/testkit/provider、evidence、probe、pioracle |
| 发布 | ai/live、internal/testkit/supportmatrix、ai/release 和 catalog-snapshot/traceability |
| Host 效果 | ai/examples/chineseeval 及其独立配置/报告，保留既有 Jev 默认 |

协议相关的新字段在 JSON decoder、clone、Hash、查询、README、contract、差分投影和 audit 中同步。已有 Method/metadata/Observer 类型不为图片和拒答另起执行入口。accepted ADR 的历史不覆写；实施时追加 ADR-0020/0021/0018 的引用与交付记录，ADR-0026 由 proposed 转 accepted 只在实装与证据条件满足时执行。

复用[acceptance](acceptance.md)的稳定 case ID 和[implementation-plan](implementation-plan.md)的依赖顺序。原 Jev fixture/diff/官方 wire replay 全量保持；任何真实精度、schema、格式与设计不合先记录有界反例和影响，不吞错、不随意宽容。
