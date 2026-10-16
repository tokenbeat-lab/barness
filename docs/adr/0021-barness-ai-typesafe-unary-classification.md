---
status: accepted
date: 2026-10-08
---

# TypeSafe 分类采用独立 unary 协议路径，共享调用运行时

工单 05 交付 Client.Classify 与 HookedClient.Classify 的单选切片。分类提交一份状态和完整问题集合，
一次逻辑调用不拆分、不并发凑答案。TypeSafe 没有官方 Go SDK，采用直连 HTTP：
typesafe × typesafe-system-one × classifier、Bearer、POST {Binding.Endpoint}/systemone。
线上形状依据 [官方 API](https://docs.typesafe.ai/api) 与
[OpenAPI](https://api.typesafe.ai/openapi.json)（2026-10-08 核对）；fixture、型号和费率均为合成数据。
真实型号、价格与支持声明由 [工单 08](../../.scratch/barness-ai-pi-1.0/issues/08-typesafe-live-and-catalog.md) 验收。

已有聊天和新的分类共同使用 scope/deadline、绑定操作与目录白名单授权、凭据一致快照、尝试重试、
准入及终态观测。聊天历史、SimpleOptions 映射、原生状态、assembler 与事件队列留在聊天路径。
操作和协议固定分派；分类目前只有一个协议，直接实现，不增加动态注册或尚无第二个用例的 adapter 接口。

## 取舍与公共边界

- 不把分类伪装成聊天或 SSE：ClassifierResult 直接给出类型化答案、用量、ResponseModel、终态和元数据。
  输出 JSON 以 MaxOutputBytes 总字节限制；MaxFrameBytes 只约束 SSE。读取、解码与答案校验使用 response 阶段。
- Classify 使用既有封闭 Options 接口，与聊天入口相同地在 capability 阶段拒绝协议不匹配选项。
  nil 和空 TypeSafeOptions 表示协议默认；SimpleOptions 不实现该接口，不能传入。
  这是对实施设计建议的 marker 子接口的克制调整，保留工单要求的运行时错协议拒绝行为。
- Classifier 子策略 nil 禁用；启用时三项容量均为正，构造时深复制。调用在复制前检查已知输入大小，
  每个问题再检查完整编码字节；最终请求受 MaxRequestBytes 约束。工单 06 扩展为非空 JSON string/object/array instructions，
  choice criteria 是 JSON string/object/array/null（1–255），score 是有序 string/object/array（2–10），
  bool criteria 可省略、提供时必须同时有 true/false string/object/array。状态是 JSON string/object/array。
  构造与回调后均验证形状、内容、有限字节与问题数，再冻结最终问题；上下文 token 超限留给厂商 422。
- 可信回调沿用 headers → payload → response。入口给回调 scope、Payload 与 ResponseInfo 填入 Operation，
  scope 中宿主提交的 Operation 会被覆盖，不授予权限。回调后的最终请求重新解码、校验并冻结；
  按最终问题集校验答案。无效请求为 callback_failed，改变模型、认证或添加授权字段为 tenant_denied。
- 成功响应之后从不重放。body 在所有路径关闭，许可持有到读取、解码和整体答案校验完成。
  默认不重试；绑定显式允许的初始请求重试复用冻结的请求、凭据和目录快照。
- 先登记 usage，再检查答案类型、精确问题/选项键集、有限概率与 confidence、分布和的 1±1e-6 容差、
  choice 属于最高概率项（允许并列）。重复响应字段、答案键或概率键拒绝。任何答案失败时整体 Answers 为空，明确上报的用量保留。
  重复 usage 或 token 字段的用量存在歧义，以 unreported 拒绝，不采用最后值覆盖。
  input/output 均存在为 complete，缺项为 partial，无 usage 为 unreported；只按目录输入费率估价。
- 响应型号只放 ResponseModel，不替换授权 ModelID；厂商 ID 从 x-typesafe-request-id 读取。
  Observer 复用既有有界队列，只含操作、归属、尝试、分类和用量，不含状态、问题、答案或错误正文。
  关闭该分类绑定或禁用子策略即可停止新调用，不回退到聊天；已发请求的费用仍需按厂商账单核对。

离线 P07 场景、公共接缝的失败记录和复现命令见
[工单 05 验证记录](../../.scratch/barness-ai-pi-1.0/typesafe-choice-evidence/README.md)。

## 混合问题与差分（工单 06，2026-10-08）

公共问题与答案的封闭接口新增 ScoreQuestion/ScoreAnswer 与 BoolQuestion/BoolAnswer。
公共 bool JSON 使用 bool/probability，协议边界转换为 noul/noul，不把厂商术语泄漏给宿主。
BoolCriteria 的指针表示说明可省略；这符合官方 API 和 pi 的可选 criteria，提供时两种说明均校验。
ScoreQuestion 的有序 Criteria 下标就是分值，模型能力再限制等级数；合法回调可替换问题名、类型和顺序。
所有答案只依据最终冻结的问题集合校验，不能依据入口的旧集合，也不能改写入口可变数据。

score 的期望须在等级范围；提供分布时必须是完整 0..n−1 键集、每项有限且在 0–1、
和及期望的绝对容差均为 1e-6；图例必须伴随分布，完整覆盖相同等级并与最终 Criteria 相等。
三种答案的任何错误都以 protocol/response 整体清空 Answers，先登记的明确用量不丢失、不归一化。

实施设计曾建议 Legend 为 map[int]string，但[官方 OpenAPI](https://api.typesafe.ai/openapi.json)
（2026-10-08）允许图例值为对象和数组，且它是请求 criteria 的回显。采用 map[int]json.RawMessage
才能保留宿主结构化等级说明及精确数值；强制字符串会拒绝本工单已允许的正常响应。
对照图例时对象键序和数值书写形式无关，数组有序；数值按十进制系数/指数比较，不做 float64 舍入或指数展开。
脱敏审计器也使用 json.Number，继续对完整 JSON 文档执行所有秘密、header 和元数据规则。
空 instructions（空字符串/对象/数组）拒绝；描述内部嵌套值保留任意合法 JSON，重复对象键拒绝。

冻结 pi 1.0.0 oracle 新增 classify 入口与 TYPESAFE_CLASSIFIER_MODELS；只运行 pi 可表达的对象状态、
字符串说明及字符串单选说明。modelPatch 将 pi 的 jev-latest 固定为 jev-1.13.0，maxRetries 必须显式等于 Binding。
基础答案投影只移除 score 的 probabilities/legend 和双方 usage.cost；ResponseModel 与宿主 Metadata
不属于 pi 结果，由离线 E2E 断言。pi 的其他结果字段保留比较（含映射 timestamp）；请求不删字段。
更宽 JSON 输入、score 详情、严格答案校验、版本 ID 和输入计价逐项登记到差分账本，并限定用例和路径。
Node fetch / Go net/http 的六个 transport header 差异按各自字段登记，不豁免整个 header 或结果。

分类目录只比较共同字段 type/provider/api/name/contextWindow；id、cost 是上述具名扩展；
pi 的 baseUrl 由 Binding 配置承接，input=[text] 为分类的隐含输入模态，没有聊天输入配置。
目录测试显式断言这两个 pi 字段，并在 pi 出现未登记新字段时失败；本轮目录仍为合成宿主目录。
官方内建 Jev 型号与真实费率交付条件属于工单 08，本轮不能凭合成调用标记 live 通过。

P07 离线、差分和扩展记录进入 P0/需求追溯，详见
[工单 06 证据](../../.scratch/barness-ai-pi-1.0/typesafe-mixed-evidence/README.md)。

两轴审查补充：encoding/json 默认把结构体 JSON 键按大小写不敏感匹配，可能让 Noul/Score
覆盖同名小写字段。封闭协议记录在解码前按精确 JSON tag 验证字段；问题名、选项名及原生说明对象
仍保留自己的任意键。bool 的省略 criteria 与显式 null 分开解码，只有省略表示无额外说明。
公开入口的 case-alias / case-collision 与 callback-null-criteria 先红后绿复现并验证。

## 工单 08：真实 wire 与有界舍入（2026-10-08）

[官方型号页](https://docs.typesafe.ai/models)、[API](https://docs.typesafe.ai/api) 与
[OpenAPI](https://api.typesafe.ai/openapi.json) 重新抓取并记录日期和 SHA-256。
线上确认 true/false criteria、版本化 model、x-typesafe-request-id、两计数完整 usage；
上下文 guard 为 400 max_tokens_exceeded，usage 缺失；422 真实形状仍未确认（ADR-0016）。
首批目录与自己的混合/单选/有界错误探针证据见[工单 08](../../.scratch/barness-ai-pi-1.0/typesafe-live-evidence/README.md)。

一份官方 200 wire 的 score=1.32，probabilities=.01/.67/.32，其加权期望为 1.31。
原 1e-6 期望校验拒绝了真实响应；公共离线回放先失败，再修正。
这是依据线上两位小数的有界兼容推断，厂商文档没有承诺舍入算法。
概率和仍要求 1±1e-6，单项/键集/图例/range 规则不变，不归一化原值。
只有 score 和所有概率均在 .01 网格上时，允许每项 ±.005（并裁入 0–1）
对应一个总和为 1 的潜在分布，其期望区间必须与 score±.005 相交；
按剩余质量向低/高等级分配求极值。高精度值维持原 1e-6；超出区间整体拒绝。
潜在风险是厂商真实精度可能并非独立舍入；更多精度的有效反例出现时重开
[工单 08](../../.scratch/barness-ai-pi-1.0/issues/08-typesafe-live-and-catalog.md)，
以该路线的真实证据调整，不能扩大为任意差异容差。移除条件是厂商提供一致高精度值。

回滚：禁用 classifier Binding 或 ClassifierPolicy 停止新调用；恢复旧目录快照时
显式移除型号授权。schema 2 矩阵按 Git 回滚整体数据，拒绝把旧报告再合并进新 schema。

## 第二个 classifier 协议的文本闭环（2026-10-12）

Decisions 转换在独立 `openai_decisions*.go` 中，保留 TypeSafe 的 native map/noul、宽说明、
舍入及 usage 语义。公共 Classify 复用固定 Binding/Credential、Hooks 与 sendUnary，按 API 固定分派。
Decisions 无图 wire 仅 model/input/questions；State string 转原文，对象/数组保留数值拼写 compact，
问题及候选由冻结 map 的确定键序转换为 native array。初始和最终 native 请求均校验完整题目字节与总量。
合法 Hook 换题后以最终独立 wire 记录校验答案，不把 wire input 反解成旧公共请求。

新增 RefusalAnswer 的公共 JSON 仅 type=refusal，合法拒答为 stop/nil error；它要求宿主显式处理
新的封闭答案 variant，因此保留逐题拒答，避免把有效拒答混入协议失败。TypeSafe native refusal
依然 protocol/response；任何损坏 Decisions 答案整体清空，同时保留唯一基础 usage 和输入估价。
新 usage 暂保守为 partial，完整 presence/details 与价格边界由工单 04 交付。
当前仅支持至少两个候选、string 说明的文本 choice，未实现的 score/bool/message/image 在凭据前
或最终 Hook guard 拒绝。后续切片直接扩充该协议真实实现，不新增 shim 或占位 Module。

该切片按[工单 02](../../.scratch/barness-ai-openai-decisions/issues/02-decisions-text-choice.md)
交付，公开 E2E 与固定 Jev 回归见[文本证据](../../.scratch/barness-ai-openai-decisions/text-choice-evidence/README.md)。
它部分落实 proposed ADR-0026，不把整份图片/发布设计标记 accepted；停用 Binding 或移除白名单即可阻止新调用。

## Decisions 文本混合答案与逐题拒答（工单 03，2026-10-12）

Decisions 增加 ordered score levels 和 predicate 请求；公共 bool 两侧非空说明按固定双换行、
True means / False means 行拼入 instructions，空说明省略。score 标签按原数组次序保留，允许重复；
公共 2–10 范围与型号能力取交集，回调也必须满足原授权型号的 kinds 和范围。
最终 guard 分 kind 检查封闭 schema，拒绝夹带其他 kind 的字段，即使字段为空或 null。
Hook 可以合法改名、互换三种 kind、重排候选与等级；冻结记录独立拥有最终 wire 数据。

完整答案按最终 name 集合匹配后原子发布；混合和全拒答均为 stop/nil，拒答只有 RefusalAnswer。
score 必须完整覆盖整数 0..n−1，逐项 label 精确匹配最终 levels；公共 Probabilities 与 Legend 保留，
Legend 为独立 JSON string。分布和与期望绝对容差固定 1e-6，不共享 Jev 的百分位舍入特例。
bool 保留 0..1 概率，任何业务阈值由 Host 决定。任何坏答案使整次 Answers 为空，唯一合法 usage 保留。

Decisions 响应允许有界 additive vendor metadata，选择避免非语义扩展破坏合法响应；代价是未来新增
语义字段须显式纳入规则。已知字段精确大小写、重复键、跨 kind 答案字段、伪造拒答原因仍拒绝；
MaxOutputBytes 限制完整响应。TypeSafe 请求/响应继续使用独立严格 decoder，不扩大 Jev 的未知字段容忍。
相关公共 E2E 同时覆盖两入口的乱序、重复 label、容差内外、损坏后 usage 与单 Attempt，
并保留 Jev 原拒答拒绝、宽 JSON、图例和真实舍入 replay。

本切片使用合成 Host 目录和本地受控 HTTP；图片、usage presence 专项与真实服务验证仍由后续票拥有。
新调用回滚仍通过禁用 classifier Binding 或移除型号白名单，已发送调用按固定快照完成并保留费用。
本票固定请求、响应与红绿历史见[混合答案证据](../../.scratch/barness-ai-openai-decisions/mixed-answer-evidence/README.md)。

## TypeSafe 独立 native 快照（2026-10-16，统一图片工单 01）

将原先“Hook 后回解为 ClassifierRequest，再复用公共校验”的内部路径直接删除。
TypeSafe 现在独立拥有原生请求、noul/choice/score 题目与 criteria；最终严格解码同时建立
独立快照，答案只使用该最终 native 题集。公共 JSON decoder 只接受公共 bool，
不再携带原生模式。该分层与既有 Decisions 的 native 所有权一致，为后续不同图片投影保留边界。

选择让 native schema、数量与格式能力独立校验，代价是协议自身显式维护这些规则；
以纯 JSON 描述校验和格式匹配复用稳定语义，避免把协议模型重新包装成 Host 类型。
序列化前的借用值容量/authority 检查继续限制 Hook 数据，最终仅一次 native freeze；
不增加 adapter 接口、动态注册、shim 或第二个语义 guard。

这是等价重构：保留 TypeSafe 对 callback 子容量的原有 callback_failed/request 分类，
总请求仍为 resource_limit/request；题目预算仍包含 canonical JSON 的 HTML 转义成本。
保持原错误优先级、精确数字和图例、有界评分舍入、usage/ResponseModel、2xx 不重放及资源释放。
先冻结公开成功与失败基线再实现；[失败清单与证据](../../.scratch/jev-1.14-multimodal-box/native-snapshot-failures.md)
记录两个既有 Classify 入口、请求捕获、Attempt/Observer 与完整性核验。

本票不增加图片/Box，不改变 Jev 1.13 wire、目录 Hash、冻结 Pi 或支持矩阵；ADR-0030
仍是 proposed，后续能力逐票验收。关闭 Binding/型号授权仍阻止新调用；代码回滚直接恢复
本票前实现，没有数据迁移或双路径切换，已经发送的 usage 继续保留。

## Jev 1.14 实验 context 图片（2026-10-16，统一图片工单 04）

在 Host 显式候选目录中仅接通固定 `jev-1.14.0` 的 PNG/JPEG context 图片。
目录必须声明图片输入、context 位置、MIME 和有限数量；绑定图片授权与 Classifier 图片策略
同时生效，调用还须选择 `TypeSafeOptions.ExperimentalImages`。选择本身不授予权限；
Long Context、Hook 改实验选择、重复或矛盾 header 都拒绝。内置目录、Hash 和支持矩阵不变。

context 投影为 native `state={state:<原 State>,images:[<私有 key>...]}`，保留原 State 中
同名字段及精确数值，不改写问题里的自然语言路径。一个来源对应一个 native JSON asset，
重复引用保留出现次序但不重复上传；原始字符串与对象键经过有界碰撞检查。上传 body 与 decoded
引用展开分别计预算，不能套用 Decisions 每次引用都生成 Base64 block 的上传估计。
绑定解析前只使用候选目录放宽 body 估计，若同名模型跨协议存在，解析后重新应用实际协议约束；
这个估计不参与授权。无图 TypeSafe wire、旧文本容量/错误顺序和 Decisions WebP 边界保持。

最终 native freeze 只接受该 context 投影及全值引用，拒绝 part、URL、substring、重叠 key、
unused assets 和局部图片。原生 asset 可合法省略 type，存在时必须为 image；Builder 总是显式输出。
型号数量限制 Jev 来源，出现次数由 References 策略限制；最终 State 预算包含生成 key 和包裹结构。Hook 可合法增删/换图/重排，最终 PNG/JPEG 格式、APNG、尺寸、文件、
来源、展开、key 和 body 预算重新校验；对普通容器、RawMessage 和具名 struct 的已知图片长度
在序列化前限流，未知自定义 encoder 仍受最终 body 与 schema 约束。Jev 的已知文件限制来自
[官方 Images 文档](https://docs.typesafe.ai/images)（2026-10-16）；本地 decoded 预算不代表
厂商处理后的体积、精确 token 或准确率。

成功的 ImageInputs 只返回最终图片 MIME 与 decoded 文件 SHA-256，按 context 首次实际使用
建立顺序，同 asset 复用一项，不按相同 bytes 合并独立 asset。全部答案合法且调用未取消后，
它与 Answers 同时发布；Observer 和共享诊断不接收内容、私有 key 或 hash。来源、实验选择、
headers、bytes 与凭据在重试前固定，2xx 后坏答案不重放。撤权只阻止新调用，已 pin 请求完成；
代码回滚直接撤回本票实现，无数据迁移。

这是 ADR-0030 拟议方案中的窄实施切片，不改变其 proposed 状态。State value、题内图片、
图片图例、Box 与真实服务/效果/正式支持仍由后续工单验收。
[工单 04 验证记录](../../.scratch/jev-1.14-multimodal-box/04-verification.md)保存失败方式、
公开 E2E、最终请求/响应、Attempt/Observer、复现命令、完整性清单和脱敏审计。

## Jev State value 图片（2026-10-16，统一图片工单 05）

固定 Host 候选 `jev-1.14.0` 可显式声明 `ImageLocations:[state]` 或 `[context,state]`。
公共 `value` Pointer 在任何包裹之前解析；只替换明确引用的现存 null 值，包含 State 根、
嵌套 object/array、转义键和数字对象键。无 context 时 native State 保持该业务树位置；
混合时先填值，再包裹为工单 04 的独立 state/images 结构。普通 null、字符串、模板样式
对象、候选/问题名和 JSON 数字拼写保留；同一来源仍只上传一个 asset。

最终 freeze 按实际全值 key 扫描 State，依次校验用途、出现/展开预算及未使用资产，按对象
键排序、数组原序构造首次使用的 ImageInputs；context 包裹中 images 先于 state 遍历。
Hook 移动/增删/换图后不复用旧 Pointer 或来源索引。对象键、substring 和题内私有引用拒绝。
生成 key 与包裹在分配引用展开前计入 State 预算；source、occurrence 和 native body 独立计量。

包裹拓扑是 Adapter 的投影元信息，不是从业务字段名推断的协议标记。已有图片的 Hook
保留该拓扑，可修改其中所有内容和资产；业务对象恰好只有 state/images 字段仍是业务树。
Hook 从无图请求首次加入图片时，没有既有投影：型号支持 State 时完整扫描业务树；仅
context 型号才识别严格的 native context envelope，且 images 必须均指向实际 assets。该选择不扩大目录位置
能力；调用授权与实验选择仍固定。这个约束避免普通业务对象和原生包裹混淆，不引入公共标记。

Decisions 含任意合法 value 时仍在 capability 整体拒绝。题内 Instructions/Criteria、图片
图例和 Box 未开放；内置目录、正式支持与 ADR-0030 proposed 状态不变。回滚直接撤回本票
代码，无持久化变更。[工单 05 验证记录](../../.scratch/jev-1.14-multimodal-box/05-verification.md)
给出公开 E2E 与独立核验产物；受控 HTTP 不能证明真实服务或视觉质量 PASS。

## Jev Instructions value 图片（2026-10-16，统一图片工单 06）

固定 Host 候选 `jev-1.14.0` 现在可显式声明 `instructions`；Choice、Score、Bool
说明中的明确 null 槽位及根图片保持题内位置。沿用公共图片来源和精确 Pointer，
不把局部图片提升到 State/context，不替换普通文字。同来源只上传一个 asset，跨题
出现分别消耗展开预算；完整题目预算在展开前包含生成 key 的 HTML 转义长度。

根图片按位置能力校验，普通 object/array 说明仍按 DescriptionFormats 校验。这避免
为图片虚构 null 文字格式，代价是公共计划与 native 实际全值引用各自识别根图片。
两者保持独立所有权；共享的仅是纯 JSON 格式判断和已有内容投影语义。
最终 native 扫描先 State 后按问题键排序遍历 Instructions，按首次实际使用重建
ImageInputs；不信任入口 Pointer 或旧问题。Criteria、身份键、对象键和 substring
中的私有 key 仍拒绝。Hook 合法增删换题/换图由最终快照决定答案与图片身份。

含 assets 的 Hook 容量失败统一 resource_limit/request；无图已有子容量失败分类
保持。绑定、实验选择、来源、格式与最终容量的前后校验持续生效。Decisions 对
局部图片在凭据前整体拒绝。Criteria 图片、Box、真实服务与视觉效果仍未验收；
内置目录、支持矩阵和 ADR-0030 proposed 状态不变。撤销授权影响新调用，回滚本票
代码无数据迁移。[公开说明](../barness-ai/jev-instruction-images.md)与
[工单验证](../../.scratch/jev-1.14-multimodal-box/06-verification.md)记录范围与可重放证据。
