---
status: accepted
date: 2026-10-08
---

# barness-ai 多类型模型操作与 Binding 授权

pi-ai 1.0.0 区分聊天、图像与分类；宿主需要发现不同型号而不把一种能力的授权扩展到另一种。
型号身份采用 `(Operation, Provider, API, ID)`，操作只取 `chat`、`image`、`classifier`。
目录使用独立的 `Model`、`ImageModel`、`ClassifierModel`，各自声明身份、能力与价格；保留原有
`Model` 的聊天字段与 keyed struct literal，避免嵌入公共结构破坏宿主源码契约。

Binding 固定一种操作，其白名单只在该操作、Provider、协议的目录中解释。入口在读取凭据前核对
绑定操作，然后核对目录与白名单交集。未知非空操作为 `invalid_request/binding`；入口不匹配为
`tenant_denied/capability`，两者均不读凭据、不发送请求、不创建 Attempt。授权仍按租户解析绑定；
凭据快照仍遵循 ADR-0003，不因操作字段而重新解析或更换 key。

Binding.Operation 的零值规范化为 chat，是公共契约中对失败关闭零值原则的显式例外：旧宿主的绑定
只曾授权聊天，解释为 chat 不扩大权限。Enabled 与 Credential.Active 仍以零值拒绝。
持久化的旧绑定无需批量改写；新操作必须显式启用。回退时禁用对应绑定，不回退到聊天 adapter。

## 目录与归属

- NewClient 拒绝重复完整身份、矛盾模态或能力、非法分类范围、负数与非有限价格、已声明模态缺少
  费率，以及 Responses/Chat 型号默认 samplingParams 的保留字段。显式零费率合法，unset/null 不等于零。
- Client 持有深复制目录和完整身份索引；`Client.Catalog`、三类 Lookup 与列表均返回独立副本。
  目录发现不读取凭据、不授予调用权限。三类型号与价格都参加整个目录 JSON 的 SHA-256。
- `CallAttribution.Operation` 从入口收到调用时即设置；CallStarted 与未解析错误同样记录 chat。
  Provider、协议、型号、账户与版本只在凭据快照一致后设置。operation 是固定枚举元数据，已审阅
  并加入 Observer 脱敏审计白名单；不包含请求正文。

## 分阶段交付

工单 04 交付目录、绑定边界与现有四个聊天入口，目录版本升为 `2026-10-08.2`。
内置目录仍只有已验收的聊天型号；宿主可构造其他类型的元数据，发现能力不代表调用路线已实现。
Classify、GenerateImages 及各自的反向跨入口拒绝在工单 05–14 的实际调用切片验收，届时才引入
共同运行时与按操作分派的 unary adapter。禁止先添加空 adapter、兼容 shim 或未验收的内置型号。

## Considered Options

- 用聊天 Model 承载全部操作：会暴露无意义的聊天预算与 compat，且图像价格无法表达缺失模态费率。
- 把身份字段嵌入公共结构：破坏现有宿主的 keyed struct literal，收益不足以承担契约变化。
- 未设置操作时放行全部操作：旧聊天绑定意外获得图像与分类授权，不符合最小权限。

修订 ADR-0001 的模型身份与调用边界、ADR-0018 的目录包含维度；价格快照延续 ADR-0010，
聊天用量与消息投影不变。公共 E2E 覆盖目录与聊天入口，并产出可重放快照和审计证据。
