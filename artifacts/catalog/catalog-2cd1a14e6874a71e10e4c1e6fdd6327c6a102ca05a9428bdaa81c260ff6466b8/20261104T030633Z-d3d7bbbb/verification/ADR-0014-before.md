---
status: accepted
date: 2026-10-02
---

# barness-ai DeepSeek × Responses：复用 Responses adapter，按 Provider 区分协议能力，登记为差分外扩展

工单 21 接入 DeepSeek × Responses（P05）。spec I4 要求它复用 Responses adapter，但 Provider 身份、能力配置、endpoint 与 key 和 OpenAI 完全独立；请求不得自动沿用 OpenAI 的服务端续接或存储字段，历史通过完整输入回放；它属于冻结 pi 路由之外的扩展，以官方协议 fixture 证明，不计为 pi 差分通过。spec 没有规定能力配置放在哪里、模型数据从哪里来、调用者显式请求不受支持的能力时怎么办，以及"登记为扩展"落在哪里。本 ADR 记录实现时的做法，维护者于 2026-10-02 确认（见文末）。

## 背景：DeepSeek 文档与研究证据

DeepSeek 的 Responses 指南（https://api-docs.deepseek.com/guides/responses_api/，2026-10-02 阅读）：base URL `https://api.deepseek.com`，只有 `deepseek-flash`；流事件为 `response.created/in_progress`、`output_item.added/done`、`content_part.added/done`、`reasoning_text.delta/done`、`output_text.delta/done`、`function_call_arguments.delta/done`、`completed/incomplete/failed`；`previous_response_id`、`conversation`、`store`、`include`、`prompt`、`truncation`、`service_tier`、`prompt_cache_key`、`prompt_cache_retention`、`stream_options`、`background`、`metadata` 等**静默忽略**；`reasoning` 只支持 `effort`（`summary` 接受但不生成）；工具只有 `function`；输入的 `reasoning` 条目以其纯文本并入 assistant 消息；usage 报告 `cached_tokens` 与 `reasoning_tokens`。指南没有列出 effort 的取值、reasoning 条目的字段与错误体；这些按 OpenAI Responses 的 schema（指南声明与之兼容）编写 fixture，并在 fixture 的 `source` 中注明。

研究探针（harness `.scratch/pi-ai-go/sdk-validation`）在真实服务上验证过：`reasoning_text` 事件、effort `none` 与 `low`、强制函数调用（含 `strict: true`）、不带 `previous_response_id` 的完整历史。`incomplete`/`failed` 的真实触发与服务端拒绝字段没有实测，只有离线断言（spec 研究追溯）。

## 决策一：协议能力按 Provider 区分，不另建 adapter

`responsesCapabilities`（`ai/responses_capabilities.go`）列出 Responses 核心之外的四项服务端能力：声明无状态存储（`store: false`）、加密推理（`include: ["reasoning.encrypted_content"]`）、提示缓存（`prompt_cache_key`、保留期字段与 session 亲和头）、服务等级（`service_tier` 请求与按响应等级计价）。OpenAI 四项都有（pi 的行为不变）；DeepSeek 一项都没有，所以它的请求里这些字段一个也不出现，session id 不派生缓存键。`previous_response_id` 与 `conversation` 本来就从不发送，samplingParams 与 payload 回调也不能加入（`responsesPayloadReferences`，ADR-0005 的回调授权），对两家都一样。历史始终是完整输入：同模型的 reasoning 条目原样回放（DeepSeek 并入 assistant 消息），工具调用与结果照 Responses 编码。

能力按 binding 的 Provider 而不是模型 compat 确定：它是厂商端点的事实，对该厂商所有模型相同；宿主自带的目录也不能给一个会忽略这些字段的厂商打开它们。只有 DeepSeek 一个真实用例，因此是一个 Provider 判断加一个小结构，不建可配置的能力注册表；出现第二个 Responses 兼容厂商且能力组合不同时再考虑。

其余 Provider 相关行为沿用 Responses adapter 既有的 pi 规则：HTTP 错误前缀为 `deepseek API error`；目标 Provider 不在 pi 的 OpenAI 工具调用 Provider 中，跨模型（含未担保降级）回放的工具调用 id 整体净化为一段；中途关闭流的终态文本与 P01 相同。

### Considered Options

- 新增模型 compat 标志（如 `supportsStore`）：pi 的 Responses compat 没有这些字段，DeepSeek 的条目又只有一个模型；放在模型上会让宿主目录可以对 DeepSeek 打开它们，或在 OpenAI 自定义模型上意外关闭。
- 独立的 DeepSeek Responses adapter：违反 spec "不复制整套生命周期"，差异只在请求字段与计价。
- 按 Provider 的能力结构（采用）。

## 决策二：调用者显式请求不受支持的能力时

只拒绝显式请求：`ResponsesOptions.ServiceTier` 在 DeepSeek 上以 `invalid_request`（phase request）在发送前失败——发出去会被静默忽略，而按等级计价会让成本估算错误。session id 与缓存保留期是缓存提示，对 DeepSeek 不派生任何字段也不报错；`reasoningSummary` 照发（DeepSeek 接受但不生成）。samplingParams 与 payload 回调对 `previous_response_id`、`conversation`、`store`、`prompt_cache_key` 的既有拒绝对 DeepSeek 同样生效（`store: true` 与非派生缓存键由回调授权拒绝）。

## 决策三：内置目录

目录 `2026-10-02.5` 新增 DeepSeek × Responses 的 `deepseek-flash`：取 pi `deepseek.json`（sha256 `549a7ddb…4d0d`）中该模型的名称、reasoning、level map、输入（文本与图片）、价格、上下文与输出上限，API 换为 `openai-responses`。pi 为它写的 compat 描述 Chat Completions（store、developer 角色、`max_tokens`、thinking 格式），只保留 `supportsStrictMode`（Responses 函数工具同样接受 `strict`，探针实测过）。level map 也是 Chat 的：Responses 的 effort 取值 DeepSeek 没有列出，实测只有 `none` 与 `low`；`high`/`max` 是否被接受由工单 23 的真实冒烟确认。`deepseek-v4-pro` 只在 Chat 上提供，不列入。

## 决策四：扩展的登记

差分账本 `ai/e2e/testdata/pidiff/ledger.json` 新增 `routes`：每项是冻结 pi 没有路由的 Provider × API，分类只能是 `extension`，须写处理决定与引用。DeepSeek × Responses 登记在此；它的 fixture（`ai/e2e/testdata/deepseek-responses`）全部带 `pidiffSkip`，不产生 `PIDIFF-P05-*` 用例，离线 E2E 的 `P05-E11-registered-as-extension` 断言这两点。这样差分之外的例外都在同一份经评审的账本里，而不是只写在测试注释中。

## Consequences

- OpenAI 的请求不变（全量差分 0 pending）；`responsesBody.Store` 改为可省略的指针。
- DeepSeek 文档更新时，改 `responsesCapabilitiesOf` 与 P05 fixture，并记录差异（spec I4 末段）。
- DeepSeek 的 reasoning 条目、函数调用条目 id 与错误体形状是按 OpenAI schema 的合理推断；真实服务若不同，以工单 23 的冒烟结果修正 fixture 与本 ADR。
- 中途关闭流的错误文本仍写着 "OpenAI Responses"（协议名，而非 Provider），与 P01 一致。

## 维护者决定（2026-10-02）

1. 决策一：采纳。协议能力按 Provider 区分而非放在模型 compat 中；DeepSeek 不发送 `store`、`include`、提示缓存字段与亲和头。
2. 决策二：采纳。显式 `ServiceTier` 在 DeepSeek 上以 `invalid_request` 拒绝；session id 与缓存保留期静默不派生；`reasoningSummary` 照发。
3. 决策三：采纳，附条件。`deepseek-flash` 取 pi 的 Chat 数据换为 `openai-responses`；level map 中未经 Responses 实测的 `high`/`max` 由工单 23 的真实冒烟确认，不被接受的级别改为 null 并换新目录版本。
4. 决策四：采纳。差分账本的 `routes` 是扩展接入路径的登记处。

## 真实冒烟后的处理（2026-10-03，工单 33）

首次真实冒烟确认 high、max 推理等级被接受（决策三的条件未触发，目录不变）。P05 fixture 按实测形状修正：UUID 形式的 id、`call_00_…` 调用 id、推理条目带 `encrypted_content`、消息条目带 `phase`、401 错误体只显示 key 后四位并附 request_id。厂商请求 id 取 `x-ds-trace-id`（DeepSeek 不发送 `x-request-id`）。`auth-refused` 场景已删除。

## 工单 09：原生图像生成（2026-10-08）

差分账本 routes 新登记 OpenAI × openai-images（image）。P08 每个 fixture 显式跳过冻结 pi，独立证明原生 JSON 协议；不伪称 OpenRouter 图像差分或聊天差分通过。

详见 [ADR-0022](0022-barness-ai-openai-images-unary.md)。

## 工单 12：Google 原生图像扩展（2026-10-08）

差分账本 routes 新登记 Google × google-interactions（image）。P09 fixture 显式
pidiffSkip，独立证明自身原生协议，不产生或计为 pi 差分通过。与 OpenAI 图像及 DeepSeek
Responses 使用同一扩展登记规则，详见 [ADR-0022](0022-barness-ai-openai-images-unary.md)。
