---
status: accepted
date: 2026-10-08
---

# barness-ai 以有界 unary 直连原生图像协议

工单 09–10 交付 OpenAI × openai-images × image 的同步生成与 JSON 编辑操作。采用 Client 的唯一 HTTP
客户端和已验证的 unary 生命周期：授权、凭据快照、Binding.Retry、逐次准入、回调、
用量和终态观测。公开请求为提示加有序参考图；输出使用独立封闭块，整组通过后才发布。
无参考图走 generations，有参考图走 edits；内置型号、实际费率、JSON 编辑接受性与真实
支持声明由工单 11 验收。

## 直连的理由与替代方案

SDK 生成端点可用，但同一 Images 协议的下一切片要求有界 JSON 编辑。已安装的
openai-go v3.66.0 将编辑图片表示为 `io.Reader`/`[]io.Reader`，mask 也是 reader；
[ImageEditParams / MarshalMultipart](https://github.com/openai/openai-go/blob/v3.66.0/image.go#L967)
在 bytes.Buffer 中编码整个 multipart，再返回完整字节。
[NewRequestConfig](https://github.com/openai/openai-go/blob/v3.66.0/internal/requestconfig/requestconfig.go#L195)
先执行此编码，后应用 RequestOption，无法在 SDK 读入前以宿主预算限制它。
此次已直接核查项目安装的模块源码，未新增依赖或自行实现 multipart。

继续 SDK 生成、另做编辑路径会拆开同一协议的最终请求守卫与结果校验；无界 multipart
不满足资源契约。因此按 spec 的 SDK 例外，仅此 Provider × 协议直连。将来 SDK 提供
可在读取前设限的 JSON 编辑与同等回调/读取能力时，重新评估；不因 SDK 更新自动换路。

[官方生成参考](https://developers.openai.com/api/reference/resources/images/methods/generate)
文档化 JSON 生成、png/jpeg/webp、质量、透明背景、压缩、审核和 usage 明细。
[编辑参考](https://developers.openai.com/api/reference/resources/images/methods/edit)
列有 JSON 内联资源，但编辑是否可用仍由工单 11 自己的 live 证据确认。
运行时不回退 multipart，也不读取任意文件或下载资源 URL。

## 边界、容量与原子发布

`GenerateImages` 与 HookedClient 同步返回 ImagesResult，不产生 delta，不接受 SimpleOptions。
ImageOptions 为协议封闭集合；nil 取协议默认，显式 null 选项拒绝，压缩零保留为零。
ImagePolicy 为 nil 时在 scope 阶段拒绝；四项容量须为正，单张 ≤ 总图像字节 ≤
MaxOutputBytes，构造深复制。旧聊天策略无需新增字段。

生成请求仅允许 model、prompt、n、size、quality、background、output_format、
output_compression、moderation；编辑另允许 images、mask、input_fidelity。提示以 Unicode
字符计 1–32000，数量取 1–10 与型号/策略上限的交集。参考图使用既有 Image，严格标准
base64 且文件头符合 png/jpeg/webp MIME；有序 slice 和 mask 在任何宿主 resolver 前独立复制。
编辑仅发 JSON：images 为有序 `{image_url: data URL}` 数组，mask 为相同资源对象。
端点在回调前由入口参考图确定；生成不能被回调改为编辑，编辑不能删除全部参考图。
原始 JSON、带类型的 map/slice 也重验；重复资源字段、显式 null mask、未知字段拒绝。

每个完整 URL ≤20971520 字符，参考图 ≤min(16, MaxReferenceImages, MaxInputImages)，mask
另占宿主 MaxInputImages 与 MaxImageBytes，不占协议的 16 个参考图名额。入口先检查全部
已知长度、base64 尺寸估算和请求总预算，再复制或遍历 base64；回调普通容器先限数量与
URL 大小，包括普通指针、结构体与带类型的容器。raw JSON 使用项目已安装的
jsonparser v1.1.2 返回字节视图；先探测第一个超额下标，逐条校验资源，转义 URL 用固定
缓冲计长，仅保留前缀与 padding，不分配完整字符串。自定义 JSON/TextMarshaler 的
表示不可预测，沿用 ADR-0002 交给编码器并在冻结后重验，不能按底层 slice/map 猜容量。
mask 要求参考图、
型号 Mask 能力且与第一张同尺寸。`DecodeConfig` 在 base64 reader 上仅读图片头，不分配
栅格。标准库提供 png/jpeg；项目已有依赖不提供 WebP 配置解码，采用 Go 官方
[golang.org/x/image/webp](https://pkg.go.dev/golang.org/x/image/webp) v0.45.0，保持原有 x/text、
x/sync 版本；不自行实现通用图像解码。

尺寸、质量、透明背景和输入保真度由目录能力授权。InputFidelity 的空列表要求省略，且仅
编辑允许；gpt-image-2 必须省略，2.5 的允许值仍由工单 11 的真实证据确定，不按 ID 猜测。
CustomSizes 从布尔开关改为可选 ImageSizeConstraints，含 EdgeMultiple、MaxAspectRatio、
MaxEdge、MinPixels、MaxPixels 与 Source（来源和日期）；所有正值/区间/来源在目录边界
校验并深复制。numeric Sizes 也不能绕过已声明的约束。未声明 CustomSizes 时只允许 Sizes。
[官方图像指南](https://developers.openai.com/api/docs/guides/image-generation) 的 GPT image 2
示例是 16、3、3840、655360、8294400；本切片使用带来源的合成能力，不纳入真实型号。
Qualities 空列表要求省略质量。生成与编辑都按冻结后的最终请求验证全部组合。

回调前检查已知输入和请求长度；回调后先检查已知 map 长度，再编码、按 MaxRequestBytes
检查、独立解码并重做选项校验。改型号和引入外部资源、操作或授权字段为 tenant_denied；
无效 schema/选项为 callback_failed。回调仅执行一次，重试复用冻结 body。

成功体仅受总输出限制，忽略聊天帧容量；2xx 后回调、读取、解码或图片校验失败均不重放。
响应逐条读取，保留有界输出记录与用量；不收集完整响应体，逐图临时解析的 raw JSON
不跨图片保留。base64 先估算字节，再以固定缓冲严格遍历，仅读文件头，不保留解码图片。
拒绝非规范 padding、换行、URL、缺图、无效 MIME/文件头以及重复/损坏 JSON。
数量须符合最终 n；任一图片失败清空全部内容，已登记用量保留。

格式依次取响应 output_format、冻结请求格式、协议默认 png。声明格式与文件头必须一致。
x-request-id 为 ProviderRequestID；协议没有响应 ID/型号，二者不填。不保存 revised prompt。
HTTP 内容策略拒绝为 upstream_error/request，2xx 内拒绝为 upstream_error/response，
损坏完成响应为 protocol/response，超限为 resource_limit；错误正文不进入 Observer。

## 用量、价格与证据

三种操作共用 Usage 与 Attempt.UsageReporting。Modalities 可选，四个 Nullable 分项保留
未报告与零的差别；结果、尝试、Observer 各自复制。聊天 JSON 不变。总量与两组明细
齐全为 complete；缺项为 partial；usage 缺失/null 为 unreported。总量未报告时只加已知
输入与输出，同时保持 partial。先登记用量，再验证图片。

ImagePricing 按文本/图片 token 逐项单独舍入乘积后相加；缺项只估已知部分，不依据张数、
分辨率或总 token 猜价格。直接 Images 禁止缓存输入费率，非负有限费率由目录校验。
模态计数进入审计白名单；提示、base64、revised prompt、响应正文不进入观测。

冻结 pi 1.0.0 只路由 OpenRouter 图像；原生 OpenAI Images 在差分账本 routes 登记为
扩展，每个 P08 fixture 有显式 pidiffSkip。原生图像目录有显式豁免，不计为聊天型号比对
或 pi 差分通过。P08、P0 和追溯映射分别登记；离线证据包含版本、目录快照、可回放场景
和脱敏审计。本工单 09–10 的合成能力与价格不构成真实型号已支持的证据。

## 工单 11：自身 JSON 编辑真实接受性与目录（2026-10-08）

[编辑参考](https://developers.openai.com/api/reference/resources/images/methods/edit)、
[图像指南](https://developers.openai.com/api/docs/guides/image-generation)、
[型号页](https://developers.openai.com/api/docs/models/gpt-image-2.5-sunburst)
于 2026-10-08 重新核对，固定日期快照、JSON inline images/mask、质量和格式均有来源。
输入保真度只明确旧型号，Sunburst 不推导允许值，初次实测省略该参数。

首个真实请求是 JSON `/v1/images/edits`，HTTP 200；其后生成和带 mask JSON 编辑
也 HTTP 200，无 multipart 或协议回退。真实结果记录原生 request ID、usage、图片
校验与 bounded captures，独立审计零发现。目录 `2026-10-08.4` 纳入固定快照。
首批能力有意保守：一张参考图与输出，1024×1024，low/medium，mask、透明背景；
png/jpeg/webp 分别实测。更大能力只保留官方事实，不构成当前支持声明。

标准处理的美元/百万 token 费率为输入文本 5、输入图片 8、输出图片 30；型号只输出
图片，文本输出不计费，显式 0 保存该事实。直接 Images 无缓存输入费率，不加按张
派生价格。实测三个请求的 input/output 与四项模态明细齐全，为 complete，估算总额
0.029431 美元；partial 仅价已知分项、unreported 不代表无消耗，由受控 E2E 保持。

支持矩阵及 release trace 现覆盖 P08，自身报告才可更新；编辑/生成必交，mask 可选
但须明确厂商证据。宿主探测目录与内置声明分开，编辑尚未实测时不提前纳入。
相关后续扩充由工单 17 发布门禁与后续型号评估追踪。

[证据、归一化约束与复跑命令](../../.scratch/barness-ai-pi-1.0/openai-images-live-evidence/README.md)。

## 工单 12：Google Interactions 生成（2026-10-08）

选择 Interactions 是按已确认设计建立新图像操作：类型化 response_format 与显式模态用量
符合边界；默认服务器存储则由固定 store=false、同步、inline 和最终请求字段白名单收敛。
[官方概览](https://ai.google.dev/gemini-api/docs/interactions-overview) 推荐新项目使用它，
仍支持 generateContent；图像路线独立，既有聊天不迁移（ADR-0012）。
[v1beta 参考](https://ai.google.dev/api/interactions-api) 的裸型号 ID 与图像配置是本路线的契约；
v1 的型号/前缀差异不通过自动回退隐藏，将来切换须独立评审并更新 fixture。

复用已交付 unary HTTP 生命周期，不引入读取环境/自建客户端的 SDK，不复制重试、准入或
传输设施；仅协议 DTO、最终守卫、响应步骤解析与用量映射独立。Endpoint 必须为 v1beta
基址，不允许 query/fragment/userinfo，API key 仅在 x-goog-api-key。GoogleImagesOptions
为 AspectRatio/ImageSize，按型号列表校验；工单 12 交付时入口与回调均拒绝参考图，工单 13 在同一入口扩展（见下）。
最终 JSON 同时受宿主与 20 MB 限额约束（工单 13 明确为十进制字节）；固定项被改/删除、续接或外部资源为 tenant_denied，
未知/重复/无效 schema 为 callback_failed。回调一次，重试复用最终 body。

响应按步骤与内容块增量解析，不收集整份正文/整步 JSON。忽略非输出步骤内容；所有
model_output 的文本/图片保序，严格验证整组图片后才发布，不承诺精确数量。
completed 才成功；failed/cancelled 为 upstream_error，非终态/incomplete/续读令牌为 protocol。
无图但有 errors 诊断为 upstream_error；不下载 URI，内联同时出现 URI 也失败。
响应 id/型号为独立诊断，不成为续接状态或覆盖授权型号。解析或图片失败仍保留唯一合法用量。

官方示例 7/20/22/49 对应 Input=7、Output=42、Reasoning=22、TotalTokens=49。
Google Input 保留总输入（CacheRead 为另记的诊断总量），不套聊天的缓存扣除规则。
四个总量与两组 text/image 明细齐全为 complete；缺项 partial，缺 usage 为 unreported。
[价格页](https://ai.google.dev/gemini-api/docs/pricing) 将文本与思考按同一输出费率收费，
因此即使目录声明仅输出 image，也要求显式 OutputText 费率；不能把 missing 当免费。
原始 Modalities 不改写；输出/思考总量均报告时，计价视图按完整输出明细之和判断思考是否已包含，未包含才加到文本。
不一致计数拒绝；缺完整关系时只估已知部分并标 partial，不猜缓存模态或按张价格。
来源 fixture 明确记录两个合成包含变体；真实形状由工单 14 确认并修订来源证据。

P09 所有 fixture 显式 pidiffSkip，routes/P0/trace 及目录豁免共同登记，不算 pi 差分通过。
Observer 沿用模态计数白名单且有 P09 独立审计证据。许可持有到响应关闭和完整验证结束，
2xx 后失败不重放。此切片只开放宿主自带目录，未增加内置型号或真实支持声明。

## 工单 13：同一入口的有序参考图编辑与完整守卫（2026-10-08）

公开 ImagesRequest 继续为 prompt 加有序 ReferenceImages，而非任意厂商 Content/Step。
[官方图像指南](https://ai.google.dev/gemini-api/docs/image-generation) 支持交错输入，公开契约
有意收敛为一段提示加图片，避免厂商协议模型进入领域边界；不模拟 OpenAI mask，不增加
厂商存储、工具、外部资源或额外执行方式。这项差异随 P09 原生扩展在 ledger/differences 登记；
将来真实需要交错输入或额外 MIME 时单独评审类型与验证 fixture，而不增兼容路径。

v1beta input 先 `{type:text,text:prompt}`，再按原顺序发送 `{type:image,mime_type,data}`。
沿用公共 Image 的严格 base64、png/jpeg/webp MIME/文件头验证与独立 slice 快照；字符串
在 Go 中不可变，无需复制大图字符串。协议参考图硬上限为 14，取宿主 MaxInputImages、
型号 MaxReferenceImages 与协议交集。指南保留首批 Nano Banana 2.1 最多 10 个物体与 4 个
角色的说明；总数校验不尝试分类参考图语义。首批分辨率 1K/2K/4K，不以其他型号的 512
扩张能力，最终尺寸和比例依授权目录而非型号名判断。

工单 13 将总请求上限明确为 **20 MB = 20,000,000 字节**；工单 12 的 20 MiB 实现已收紧。
提示、MIME、全部 base64 和 JSON 开销都计入。入口仅编码剥去图片数据的小型信封，按合法
base64 无转义的长度补回准确字节；超过协议或更小宿主预算时在凭据读取前拒绝。已知
回调 map/slice/指针/带类型容器/raw JSON 先检查数量、单图、总量
（含提示转义、MIME 和信封开销），带标签的嵌入字段使用共享资源字段选择，普通 []byte
按 base64 膨胀估算，自定义 map 键不重复调用但其普通值仍先检查；raw base64 转义用固定
缓冲计长，无比例副本。自定义 marshaler 的表示和 IsZero 省略语义无法预知，按 ADR-0002 在冻结后校验；
保留可取地址字段的指针编码方法，不额外执行宿主方法。
回调可在同一模型授权内添加参考图；其后独立解码最终字节、重验提示、MIME/base64、
单图、数量、大小及比例，不重新读旧选项。无效体为 callback_failed、越权为 tenant_denied、
容量超限为 resource_limit；固定项删除/放宽、每个拒绝字段和 URI 均有独立 E2E 证据。

生成与编辑共享响应解析、完整有序输出、原子失败、保留用量、逐次准入与释放语义。
重复场景同时验证两种输入，2xx 后从不重放；输出数量只是上限，不承诺精确 N 或补发请求。
P09 新追溯项覆盖编辑、输入预算与回调最终输入；离线证据包含版本、回放 fixture、脱敏
观测、资源归零和既有路线回归。真实型号/能力、价格与此组合 live 仍由工单 14 验收。
