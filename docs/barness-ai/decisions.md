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
| 图片 | text-only | 仅有序 inline PNG/JPEG/WebP，无 URL、文件 ID 或自动 resize |

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
