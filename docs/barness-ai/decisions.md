# Decisions 公共契约与 Host 迁移

正式 `BuiltinCatalog()` 包含 `(classifier, openai, openai-decisions, gpt-6-luna)`。
同名聊天型号是另一身份；Host 的 Binding 固定操作、Provider、协议和 AllowedModels。
无日期 `gpt-6-luna` 是服务 alias；CatalogVersion/Hash 固定元数据和价格，不固定服务权重，
因此真实报告保留时间和 ResponseModel。三种问题类型均支持，公共 score 范围仍为 2–10
个有序等级，下标从零开始；922000 目录上下文及厂商潜在更大 score 范围不表示本项目实测容量。

| 边界 | Jev / System One | Luna / Decisions |
| --- | --- | --- |
| Options | nil 或空 TypeSafeOptions | nil 或空 OpenAIDecisionsOptions |
| State | JSON string/object/array | 字符串转原文，对象/数组转保留数字精度的 compact JSON 文本 |
| 问题说明 | 保留 JSON 字符串/对象/数组及既有单选 null 规则 | Instructions / Criteria 仅 JSON 字符串 |
| choice | 保留单候选行为 | 至少两个候选 |
| score | 2–10，保留 Jev 有界舍入 | 2–10，分布/期望容差 1e-6，不共享舍入例外 |
| 普通答案 | ChoiceAnswer / ScoreAnswer / BoolAnswer | 同左，score 分布和字符串 legend 保留 |
| 拒答 | 未知 native refusal 拒绝 | RefusalAnswer，混合与全拒答为 stop/nil error |
| 图片 | text-only | inline PNG/JPEG/WebP，按 context 引用排序和复用，无 URL、文件 ID 或自动 resize |

`ClassifierRequest.Images` 必须同时满足三项：`ClassifierPolicy.Images` 的正数
MaxImages/MaxTotalImageBytes，ClassifierModel.Input/ImageMIMETypes/MaxInputImages 的能力，
以及 `Binding.AllowClassifierImages=true`。省略默认不授权。GenerateImages 的 Image 子策略
不启用分类图片。每图 decoded 受 MaxImageBytes、总 decoded 受分类图片子策略、完整含
base64/data URL 的 native JSON 受 MaxRequestBytes；支持上限为 128 张，实际取更小的
Host/型号限制。State 仍必填，图片任务可用空对象。构造器复制配置，不能用共享 map 暗改快照。

最终 OnPayload 后会重新检查模型身份、问题集、整个 input/content 树、资源与图片权限，
随后冻结请求与答案验证依据。合法地改名、增删题或说明会改变最终期望答案；越权目标、模型、
身份或图片权限不能通过 Hook 获得授权。回调每 Logical Call 一次，重试保持最终 bytes、
headers、key、目录价格与 Binding/Credential 版本；没有另一协议的隐式 fallback。

迁移 Host 时先保持 Jev 输入与 Options 原样，新增独立 Luna Binding，再把说明明确转换为
JSON 字符串并验证至少两个 choice。答案 type switch 必须处理 `RefusalAnswer`；nil error
不表示每题有普通答案。缺少/多余/重复/损坏答案整体清空 Answers，以 protocol/response
失败；合法逐题拒答保留其他普通答案。文本/图片初始输入、回调与结果的可重复例子见
[迁移宿主](../../ai/examples/classifiermigration/README.md)。

Decisions 的 usage presence 独立验证：没有报告不等于免费；已唯一验证的 usage/cost
在坏答案或合法拒答后保留。Input 包含缓存读写，canonical cache buckets 为零，避免重复计费；
输出与 reasoning 仍保留但估价为零。按 classifier 目录 input-only 费率计算，input >272000
使用整个请求的高档费率，不把同名聊天价格带入。规则及错误后保留边界见
[用量契约](contract.md#decisions-用量与估价工单-04)。

初始授权拒绝发生在绑定/能力阶段；Hook 图片越权为 tenant_denied/request；损坏的最终
payload 为 callback_failed/request；结果错误为 protocol/response。取消、时限、2xx 后不重试、
准入和资源释放沿用[运行时契约](decisions-runtime.md)。Observer 仅元数据，不含正文、图片、
答案、秘密或错误文本；业务阈值与转人工由 Host 自己决定。

回滚由 Host 发布新的 Binding 版本：停用 Enabled、移除 AllowedModels 中的 Luna，或撤销
AllowClassifierImages。前两者拒绝新的对应调用；撤销图片权限仍允许已授权文本。已 pin
调用按原快照及有限重试结束，历史 Usage/Estimated Cost 不撤回。不改走 Jev/chat，保留原
证据和失败结果。撤销/恢复的六个入口与 Hook 场景是 normal/race 发布各自必交证据。

十组合发布需要当前 schema 3 的各自报告、七个冻结 Pi 差分子集（含 P10）和原三条
Extension Route、旧混合及分类图片设计负载、Jev/Luna 独立效果与逐包审计。
历史 Jev 使用 recorded snapshot 仍可单独核验；当前门禁另要求固定任务/config/catalog、
当前执行源码完整集合和真实 COMPLETE。完整全拒答的准确率为 0，条件准确率/ECE/Brier
为 null，消耗如实保留，不设置业务准确率门槛。


## 统一图片引用

`Images` 是唯一来源集合；`ImageReferences` 省略、null 或空数组时，按来源顺序各使用一次
context。非空数组是完整声明，每个来源必须至少被使用一次。相同 bytes 的不同来源保持
独立，重复引用也不去重。例如：

```go
req.Images = []ai.Image{png, jpeg}
req.ImageReferences = []ai.ImageReference{
    {ImageIndex: 1, Placement: ai.ImagePlacementContext},
    {ImageIndex: 0, Placement: ai.ImagePlacementContext},
    {ImageIndex: 1, Placement: ai.ImagePlacementContext},
}
```

最终唯一 user 消息先发 State text，再发 JPEG、PNG、JPEG 三个 data URL block。
公共 JSON 每条引用要求显式整数 `imageIndex`（零合法）及 `placement`；未知字段、重复键、
大小写别名、负数、小数或缺失下标拒绝。context 必须省略 `pointer`。

公共 `value` 引用用非空 RFC 6901 JSON 字符串 `pointer` 指向现有 null 内容槽位：State、
单题 Instructions，或单个 Choice/Score/Bool criterion 的根或后代。支持 `~0`/`~1`、
数字对象键和规范数组下标；不创建路径，不解释字符串中的特殊标记，不改变普通 null。
Questions、Criteria 集合及 type/图片/引用/授权字段不能作为目标。被明确引用的 State 或
Instructions 根可为 null；未引用时仍按既有格式和非空规则检查。

此票只实现 Decisions context：任何合法 value 用途（包括同时具有 context 引用）都整体
以 `invalid_request/capability` 在凭据读取前拒绝，坏引用为 `invalid_request/scope`。
Jev 图片及局部位置仍未开放；不会丢弃局部用途或提升为共享图片。

`ClassifierCapabilities.ImageLocations` 当前仅可显式声明 `context`，且只限 Decisions 图片
型号。省略时旧图片型号仍解释为 context，文本型号没有图片位置；查询不写回默认值，旧目录
Hash 不变。未知、重复、文本型号位置及未实现位置声明在 Catalog.Hash/NewClient 处拒绝。

`ClassifierImagePolicy.References` 可配置 `MaxReferences` 和 `MaxExpandedImageBytes`，显式值
必须都为正数；nil 时分别继承 MaxImages 与 MaxTotalImageBytes 的有限上限。来源仍受原数量、
单图和总 decoded 预算约束；展开预算按每次使用计算。Decisions 展开数还须不超过
`min(128, ClassifierModel.Capabilities.MaxInputImages)`，MaxInputImages 也继续限制来源数。
已知超限在复制前拒绝，最终编码体再次检查 MaxRequestBytes。Hook 后原生图片 block 按
展开用途重检这些预算及权限，既有来源预算不会错误地收紧 Host 显式配置的复用额度。

## 最终图片身份

成功的 `ClassifierResult.ImageInputs`（JSON `imageInputs`）按最终 Decisions 图片 content
顺序列出 `ClassifierImageInput{MimeType, SHA256}`，每项仅含规范 `mimeType` 和 decoded
原文件的 64 位小写十六进制 `sha256`。hash 覆盖完整文件，不能用 Base64、data URL 或
图片头的 hash 替代。重复 block 保留不同条目，即使其 MIME/hash 相同。

Hook 可以合法换图、增删或重排；Host 应核对返回的 MIME/hash，不能假定结果下标等于
`Images` 来源下标。上述 JPEG、PNG、JPEG 示例产生三个相应身份条目。结果记录由 Host
独立持有，不包含图片 bytes、厂商 key 或资源 ID；hash 不是访问授权，也不是跨调用引用。

元数据在最终权限、格式与预算复核后，以固定缓冲流式计算并检查取消；全部答案合法及
调用未取消时才随 Answers 一起发布。合法 RefusalAnswer 也可带 ImageInputs。坏答案、
坏 usage、回调错误、取消或其他失败不发布两者；已验证 usage 与 ResponseModel 的保留
顺序延续原合同。文本调用（包括 Hook 删除全部图片后）省略 `imageInputs`。

元数据记录的编解码拒绝缺失/null/重复/未知/大小写别名字段、非规范 MIME 和非规范 hash。
内容 hash 只返回本次调用的租户，不进入 Observer、Attempt 或共享诊断。此字段不扩展
Jev 图片、局部位置、Box、目录能力或真实服务支持结论；协议证据见
[工单 03 验证记录](../../.scratch/jev-1.14-multimodal-box/03-verification.md)。
