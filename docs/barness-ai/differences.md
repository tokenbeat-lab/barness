# barness-ai 差异登记

本文登记 barness-ai 相对冻结基线 pi-ai `0.87.1`（commit `898ab804050730e9dcefb4443875d5a932aa6a32`）的全部可观察差异。spec Testing Decisions §5 只允许三种处置：**已修复**、**本规范定义或后续明确批准的扩展**、**待处理**；待处理阻断对应协议发布。逐条判定的权威记录是差分账本 [`ai/e2e/testdata/pidiff/ledger.json`](../../ai/e2e/testdata/pidiff/ledger.json)（每条带协议、路径、适用用例、处置与依据）；本文按主题归纳它，并补上不经差分比较的租户扩展。

发布门禁（`go run ./ai/release/cmd/releasegate`）读取账本：存在 `pending` 决定、任一差分用例有待处理发现、或扩展路径未登记，差分门禁即失败。

## 1. 账本概况

| 协议 | 扩展 | 已修复 | 待处理 |
| --- | --- | --- | --- |
| openai-responses（P01） | 57 | 12 | 0 |
| anthropic-messages（P02） | 38 | 0 | 0 |
| google-generative-ai（P03） | 31 | 0 | 0 |
| openai-completions（P04、P06） | 37 | 0 | 0 |

“已修复”条目只记录曾出现并已修复的差异；同一差异再次出现即为回归，按待处理计。扩展路径（`routes`）：DeepSeek × Responses。

## 2. 规范新增决策

| 决策 | 差异 | 依据 | 验收 |
| --- | --- | --- | --- |
| D1 显式有限资源策略 | pi 无字节、队列、并发与时限策略；barness 构造时必须提供有限策略，超限以 `resource_limit` 结束（限额内与 pi 一致；超限的帧/输出/工具 JSON/错误体差分登记为用例级扩展），超限的尝试不重试 | spec I9；ADR-0002、ADR-0007、ADR-0008 | E08、D1-construct-*、E08-policy-pressure-* |
| D2 不一致配置快照直接失败 | 原研究允许重新解析；barness 在两次解析冲突时发请求前失败（`credential_unavailable`/consistency），不在同一逻辑调用内重解析 | spec I3；ADR-0003 | E07 snapshot-*、*-between-reads |

## 3. 扩展接入路径

**DeepSeek × Responses（P05）**：冻结 pi 只经 openai-completions 路由 DeepSeek。barness 复用 Responses adapter，按 Provider 确定能力：不发送 `store`、`include`、提示缓存字段与亲和头，显式服务等级在发送前拒绝，历史以完整输入回放，推理流为 `response.reasoning_text.delta/done`。以官方协议 fixture（`ai/e2e/testdata/deepseek-responses`）离线证明，**不计为 pi 差分通过**，登记于账本 `routes`（维护者决定 2026-10-02，ADR-0014）。

DeepSeek × Chat Completions（P06）是 pi 自身的路由，全部场景进入差分（ADR-0015）。

## 4. 已批准的扩展（按主题）

| 主题 | 差异 | 依据 |
| --- | --- | --- |
| HTTP 栈默认头 | Accept-Encoding、Accept-Language、Connection 等 Node undici 与 Go net/http 的默认值不同；不属于 pi 或协议逻辑 | 维护者决定 2026-10-01（各协议同） |
| SDK 运行时指纹与宿主描述头 | barness 发送自己的 `User-Agent: barness-ai`；不发送 SDK 描述宿主 OS、架构与运行时版本的 `X-Stainless-*` 头（pi 的 SDK 发送） | 维护者决定 2026-10-01/02；工单 31；ADR-0016 决定 7 |
| live partial | pi 的 `partial` 是共享的持续更新视图；barness 以受同步保护的 `PartialView` 提供快照，事件中的 partial 比较按此投影。两边序列化 partial 的时机不同（pi 单线程处理完整个 chunk 后才序列化事件），所以 barness 的视图可能领先或落后于 pi 在同一事件上的视图，但不会落后于正在读取的事件 | spec I8；pi types.ts；工单 24（Gemini usage 方向） |
| 工具参数原始 JSON | barness 在每个工具调用上保留 Provider 原始参数文本（`RawArguments`） | spec I6 |
| 错误文本脱敏（安全差异） | pi 原样复制 Provider 错误体；barness 脱敏其中 key 形态的文本 | spec I9；ADR-0009 |
| 运行时错误文本 | pi 报告 JavaScript `JSON.parse` 与 undici `terminated`/AbortError 文本；barness 报告对应的 Go 文本 | 各协议账本条目 |
| 原生状态担保 | 无可信封套的原生状态按跨模型规则降级（pi 原样回放） | spec I6；ADR-0001 |
| 缓存/亲和标识派生 | 模块构造的 prompt cache key 与亲和头按租户与账户派生，不直接发送裸 session ID | spec I9 |
| 资源限额 | 超过 MaxFrameBytes、MaxOutputBytes、MaxToolJSONBytes、MaxErrorBodyBytes 的用例以 `resource_limit` 结束；超限错误体（429/5xx）不重试而 pi 会重试 | D1；ADR-0007 |

## 5. 不进入差分、单独断言的租户扩展

以下行为不塞入 pi golden，由离线 E2E 单独断言（spec Testing Decisions §5）：租户/调用/尝试归属与事件信封（E10）、Code/Phase 错误分类（E02、E07）、用量完整性与价格快照标记（E11，ADR-0010）、资源限制与准入（E08，ADR-0007/0008）、并发与同步访问（E01、E06）、Observer 记录与脱敏（E09，ADR-0009）、绑定上的重试策略（E05，ADR-0006）、托管工具放行（E04，ADR-0005）。

## 6. 记录在案的有意差异与已知限制

| 项目 | 内容 | 依据 / 跟踪 |
| --- | --- | --- |
| `timeoutMs: 0` | barness 视为未设置；pi 会立即超时 | ADR-0008 维护者决定 |
| 失败响应携带的用量 | 与 pi 一样丢弃，尝试标为未上报 | ADR-0010 决策三；工单 25 |
| Gemini 回调 | 回调看到 REST 请求体而非 pi 的 SDK 参数；不调用 onResponse（与 pi 同） | ADR-0012 |
| 流式工具参数解析成本 | 每个 delta 全量重解析（与 pi 同），CPU 随参数大小平方增长；示例策略据此收紧 `MaxToolJSONBytes` | 工单 32 |
| 未纳入目录的模型 | 需要中途工具变更、托管推理强度或服务端备用模型的 Anthropic 模型，以及需要 additional_tools/tool search 的 OpenAI 模型不列入内置目录 | 工单 26–28；ADR-0011 |
| SDK 响应体兜底 | SDK 在调用方 context 结束时不关闭恰好到达的响应体，barness 以中间件兜底关闭 | 工单 29 |
| 真实冒烟确认的 DeepSeek 行为（2026-10-02） | thinking 加强制工具选择被 400 拒绝；提示缓存字段被接受；Chat 的 usage 在 finish_reason chunk 内；high/max 推理等级被接受；错误体为 `error{code,message,param,type}`，key 只显示后四位；厂商请求 id 取 `x-ds-trace-id`（DeepSeek 不发送 `x-request-id`）；关闭思考时缺推理计数视为完整上报。P05/P06 fixture 已按实测形状修正 | 工单 23、33；ADR-0014/0015/0016 |
