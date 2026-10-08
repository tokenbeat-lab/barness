---
status: accepted
date: 2026-10-08
---

# barness-ai 以有界 unary 直连 OpenAI Images

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
