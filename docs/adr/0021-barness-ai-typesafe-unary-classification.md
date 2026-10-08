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
