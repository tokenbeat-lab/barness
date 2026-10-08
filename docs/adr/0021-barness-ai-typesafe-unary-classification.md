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
  每个问题再检查完整编码字节；最终请求受 MaxRequestBytes 约束。首期 instructions 为非空 JSON string，
  criteria 为 JSON string 或 null，选项 1–255，状态为 JSON string/object/array。
  结构化说明与 score/bool 留给 [工单 06](../../.scratch/barness-ai-pi-1.0/issues/06-typesafe-mixed-questions-and-parity.md)。
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
