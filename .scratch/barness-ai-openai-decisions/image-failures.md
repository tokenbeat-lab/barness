# 工单 05：实现前的公共边界失败清单

测试 seam 已由工单指定为 `Client.Classify`、`HookedClient.Classify`、公共 JSON 与 `NewClient` / `Catalog.Hash` / 目录查询；使用本地 Provider 的最终 HTTP 捕获、Host 凭据读取计数和现有资源 gauges。

- Images 省略/null/[]，元素非对象/null，data/mimeType 缺失/null/非字符串、未知字段/大小写别名/重复字段；State 缺失/null/非法类型。
- Classifier nil / Images nil，GenerateImages 独立；政策非正总量/数量、总量超过整请求；错误操作打开 Binding 开关。
- text-only / Jev、缺 image capability/MIME/count，以及 Binding 默认 false；均零凭据/Attempt/HTTP。
- PNG/JPEG/WebP 原资产及顺序，坏 MIME/magic、非法字母/截断/padding/noncanonical padding/CRLF、URL/path/file ID/detail；不修补、不拉取。
- 单图 decoded、总 decoded、数量、公共 known bytes、完整 JSON 分别测上下边界；128/129 与更小 Host 优先级；checked arithmetic，巨大已知数据在复制和解码前拒绝。
- Hook 从文字添加图片，按 pinned Binding 授权；第二 message、非 user、text 非 first、坏 content/type/detail/外部资源、后位巨大图片/累计/问题名/自定义 JSON replacement 全量 guard。
- 同步 barrier 后改 caller Images/问题/政策/目录或 callback map/Binding，最终请求和重试字节不变；取消图片验证、唯一终态及 body/permit/waiter 归零。
- 旧 Jev 与 GenerateImages 对照；固定资产/MIME/hash、decoded/base64/whole JSON 数值、失败/成功捕获、SHA-256 manifest、整包 audit、重放命令留证。

这张票交付离线图片单选；设计负载、真实厂商与固定 Pi 新路线由工单 08/09/07 提供证据。
