# 工单 09 的公共失败方式

接缝由 issue 09 指定：共享 fixture runner → Client/HookedClient.GenerateImages →
受控 HTTP Provider；宿主解析器、Observer 与既有 body/permit/call/event 探针。
先扩展 runner 和写生成、多图、非法选项、错误绑定、未启用、缺图、坏图、拒绝及用量 fixture，
首次运行因 APIOpenAIImages、GenerateImages、ImagePolicy、ImageOptions 等尚未存在而编译失败。
随后写回调和资源 E2E，再实现协议；没有新增单元测试或内部测试接口。

需要防住的行为：

- 旧聊天绑定或错误 Provider/API 获得图像权限；未启用策略仍读凭据/发送。
- 非法数量/尺寸/质量/透明背景/格式/压缩/审核请求被发送；非法 Unicode 或超长提示被发送。
- 已知超长字符串、raw JSON、raw 指针、容器先复制或编码再检查。
- 回调更换型号、端点或操作，引入 streaming/partial images/user/编辑或外部资源/续接状态。
- 响应格式仍依赖回调前的 options；回调数量修改未冻结；重试重新执行回调。
- JSON 被聊天帧预算误杀，或总量越界后继续读；2xx 后失败被重放。
- 只取第一张图，缺图仍成功，坏图导致部分输出；宽松 padding/换行/MIME/文件头检查。
- 输出数量、单张或总图片字节超限；损坏/重复/尾随 JSON；URL 被下载。
- 图片校验失败丢掉用量；未上报与零混淆，缺明细仍标 complete；按总量/张数猜成本。
- 直接 Images 接受缓存费率；HTTP 内容策略拒绝仍被分类成调用者 invalid_request。
- 异步 Observer、结果与尝试共享模态指针，宿主修改结果污染审计记录。
- 取消、超时、回调异常/panic、坏 JSON 或读断后 body/许可未释放。
- 原生路线冒充 pi 差分或聊天目录纳入，未跑 live 就把真实型号标为支持。

专项先复现的真实遗漏：HTTP 400 content_policy_violation 得到 invalid_request/request；
目录接受 CacheReadText/CacheReadImage（含显式 null）；缺失/null total_tokens 未汇总已知总量。
新增公共 E2E 先失败，再分别修正为 upstream_error/request、构造拒绝和 partial 下的已知总量。
早期专项还发现测试预期的 /v1 前缀、独立乘积 IEEE 舍入、测试策略字段关系写错；
按已有绑定契约和独立数值预期修正 fixture，未为错误预期修改生产行为。

两轴复审另发现两个入口/协议边界遗漏：解析绑定期间仍引用调用者的指针选项；
mime_type 为 null 或空字符串时被当成未声明。先加公共 E2E，确认三个场景 RED，
再在入口复制选项并严格拒绝无效 MIME 声明；专项与复审均转为 PASS。
