# barness-ai 对标 pi-ai 1.0.0：实施设计

Updated: 2026-10-07

本文指导 [spec](spec.md) 的编码实施。spec 规定要做什么和验收标准，本文规定怎么做：每个工作包落在哪些文件和函数、目标行为的精确规则、要移植的 pi 代码、线上请求和响应的形状、要先写的失败场景。两者冲突时以 spec 为准，并回来修正本文。

本文由以下材料合并而成，所有关键事实都已追溯到一手来源并在文中直接引用：同事的增量设计底稿（2026-10-05）、对底稿的评审（2026-10-07），以及评审时对 pi 源码、barness 源码、厂商官方文档的三份核验记录。底稿里被评审判定为错误或过期的内容已在本文中改正，不再单独保留。

本文中的 Go 片段是接口轮廓，用来固定形状和命名，不是可直接编译的实现。

## 1. 固定基线与证据规则

| 对象 | 固定版本 | 如何引用 |
| --- | --- | --- |
| barness | `fdb075b5858966fbefd455d92386e2092f5a146c` | 本文的 barness 行号均对应此提交；实施时以函数名定位，行号仅供参考 |
| 当前冻结 oracle | npm `@earendil-works/pi-ai@0.87.1`，gitHead `f07218c4d4bbc12bef056a7058c3dd49dfe41abe` | [PROVENANCE](../../ai/internal/testkit/pioracle/node/PROVENANCE.md) |
| 目标 oracle | pi tag `v1.0.0`，commit `a13d35a742c6ef8462812a28fbe1d8c8b7431c32`，npm `@earendil-works/pi-ai@1.0.0` | GitHub 永久链接，形如 `https://github.com/earendil-works/pi/blob/v1.0.0/packages/ai/...` |
| pi 版本窗口 | 0.99.0、0.99.1、0.99.2、1.0.0 | [CHANGELOG @ v1.0.0](https://github.com/earendil-works/pi/blob/v1.0.0/packages/ai/CHANGELOG.md) 第 3–80 行 |
| openai-go | 已安装 v3.66.0；最新 v3.71.1（2026-10-02） | `https://github.com/openai/openai-go/blob/v3.66.0/...` |
| 厂商文档 | 2026-10-07 在线状态 | 各节给出官方 URL |

证据规则：

- pi 事实一律以 tag `v1.0.0` 的源码为准，不以本地工作树为准。
- 厂商事实会过期。实施时把采用的请求/响应形状、型号能力与价格存为 fixture，fixture 的 `source` 字段写官方 URL 与抓取日期。
- 本文标注"live gate"的事项，在对应 combo 的真实冒烟通过前不得写进支持目录。

## 2. 现状地图

实施会改动的现有构件。行号对应 `fdb075b`。

| 关注点 | 位置 | 现状 |
| --- | --- | --- |
| 调用执行 | [call.go](../../ai/call.go) `Client.run`（第 31 行）、`execute` | 顺序为 scope → binding → capability/options → credential → snapshot → admission → send；`execute` 直接构造聊天 `assembler` |
| adapter 注册 | [adapter.go](../../ai/adapter.go) `registry()`（第 56–65 行） | `map[API]adapter`；`adapter` 接口为 `simpleOptions`、`historyRules`、`headers`、`stream`，是聊天专用的 |
| Options 接口 | [options.go](../../ai/options.go) 第 11–25 行 | `api()`、`clone()`、`validate()`；`execute` 在第 97–99 行检查 `options.api() != binding.API` |
| 预检 | [preflight.go](../../ai/preflight.go) `resolveBinding`、`authorizeModel`（第 60 行）、`checkSnapshot`（第 98–122 行） | 不支持的 `AuthKind` 在第 43–45 行以 `invalid_request`/`PhaseBinding` 拒绝 |
| Binding | [binding.go](../../ai/binding.go) 第 68–102 行 | 零值失败关闭：`Enabled`、`Active` 为零即拒绝，`RetryPolicy` 为零不重试 |
| 目录 | [catalog.go](../../ai/catalog.go)，470 行 | `Model` 第 19–37 行；`clone` 第 426 行；`Hash` 第 440 行；`validate` 第 454 行只检查价格；`lookup` 第 463 行取首个匹配；五个内建数据构造闭包在第 158、184、215、296、360 行 |
| 用量 | [usage.go](../../ai/usage.go) | `Usage`、`UsageCost`、`UsageReporting`、`ModelCost`；`TotalTokens` 注释写"the provider's own total"（第 24 行） |
| 响应体限额 | [limits.go](../../ai/limits.go) `limitBody`（第 96–105 行）、`streamBody`、`errorBody`（第 154–172 行）、`base64Size`（第 84–91 行）、`checkImages`（第 61–82 行） | 所有 2xx 都包成 `streamBody`，同时按帧与总输出计量；没有空行的 JSON 是一整帧 |
| 资源策略 | [policy.go](../../ai/policy.go) `ResourcePolicy`（第 10–86 行） | 全是标量；`NewClient` 以 `*cfg.Policy` 值拷贝（[client.go](../../ai/client.go) 第 112 行），今天是完整拷贝 |
| 重试 | [retry.go](../../ai/retry.go) `initialRequest`、`delay`、`serverDelayMs`、`msDuration` | 无法解析的 `retry-after` 返回 `NaN, true`，`msDuration(NaN)` 为 0，立即重试 |
| 错误 | [errors.go](../../ai/errors.go) `Code`、`Phase`、`codeForStatus`（第 144–158 行） | Phase 九个值：scope、binding、capability、credential、consistency、admission、request、stream、event_queue |
| 回调 | [hooks.go](../../ai/hooks.go)、[hooks_run.go](../../ai/hooks_run.go) `boundHooks.payload` | 请求体必须是 JSON 对象，回调后以 `marshalJS` 重新编码 |
| 观测 | [observer.go](../../ai/observer.go) `callFinished`（第 214–215 行） | 读 `res.Message.StopReason`、`res.Message.Usage` |
| 归属 | [result.go](../../ai/result.go) `CallAttribution`（第 11–23 行）、`CallMetadata` | 没有操作维度 |
| 直连设施 | [gemini.go](../../ai/gemini.go)、[transport.go](../../ai/transport.go) `watchedTransport`、[http_failures.go](../../ai/http_failures.go)、[openai_sdk.go](../../ai/openai_sdk.go) `requestIDHeaderOf`、[request.go](../../ai/request.go) `isImageMediaType` | Gemini 已用直连 HTTP，unary 路线复用同一套 |
| Responses 解析 | [responses.go](../../ai/responses.go) `responsesParser.open`（第 166–185 行）、`itemDone`（第 270–311 行）、`finalize`（第 316–347 行） | `open` 直接覆盖 `p.slots[outputIndex]`；`itemDone` 删除 slot；`finalize` 不检查未完成工具调用 |
| 服务等级 | [responses_usage.go](../../ai/responses_usage.go) `serviceTierMultiplier`（第 70–82 行） | 只有 flex、priority |
| Anthropic 用量 | [anthropic_stream.go](../../ai/anthropic_stream.go) `anthropicUsage.start`（第 343–355 行）、`delta`（第 357–367 行） | `start` 读 `ephemeral_1h`，`delta` 不读 |
| samplingParams | [options.go](../../ai/options.go) `SimpleOptions.resolve`（第 93–101 行）；[responses_options.go](../../ai/responses_options.go) `applySamplingParams`（第 203 行）；[chat_request.go](../../ai/chat_request.go) 第 156 行、[responses_request.go](../../ai/responses_request.go) 第 182 行 | 只有 simple 入口合并型号默认值；full 入口只应用调用级 |
| 差分 oracle | [pioracle/node/runner.mjs](../../ai/internal/testkit/pioracle/node/runner.mjs)、`provenance.json`、[pioracle](../../ai/internal/testkit/pioracle/) 的 `oracle.go`、`compare.go`、`ledger.go` | runner 按 `api/*` 模块的 `stream`/`streamSimple` 运行，校验版本与模型数据哈希 |
| 差分账本 | [pidiff/ledger.json](../../ai/e2e/testdata/pidiff/ledger.json) | `decisions` 与 `routes`；DeepSeek × Responses 在 `routes` |
| fixture runner | [scenario_fixture_test.go](../../ai/e2e/scenario_fixture_test.go) `fixtureProtocol`、`fixtureScenario` | 以 SSE 为中心：`sse func(...)`、`textMarker`、`Events` |
| 追溯 | [traceability.json](../../ai/release/traceability.json) | `scenarios` 正则、`p0`、`differential.required/routes`、`liveCombos` |
| 支持矩阵 | `ai/internal/testkit/supportmatrix/matrix.go` 第 207、212–213 行；[support-matrix.json](../../ai/live/support-matrix.json) | 以 combo 名为键，校验 provider/API 一致 |
| live | [combos_test.go](../../ai/live/combos_test.go)、[main_test.go](../../ai/live/main_test.go) 第 19–29 行 | combo 列表固定；预算 32 次调用、4096 输出 token，按聊天设计 |
| 审计 | `ai/internal/testkit/audit/audit.go` 第 239、254–266 行 | live 响应头白名单与 Observer 字段白名单 |
| 目录包含测试 | `TestCatalogInclusion/builtin-models-are-pi's`（[ADR-0018](../../docs/adr/0018-barness-ai-catalog-inclusion-criterion.md) 第 67 行） | 逐字段比较内建型号与冻结 pi 模型数据 |

## 3. 工作包与顺序

| 顺序 | 工作包 | 前置 | 交付物 |
| --- | --- | --- | --- |
| 1 | G0 基线迁移（第 4 节） | 无 | 1.0.0 oracle、新 provenance、账本改写、目录比对；门禁通过 |
| 2 | G 聊天五项（第 5 节） | G0 | 五项行为与 fixture；ADR-0006/0010/0011 修订 |
| 3 | A 类型与授权（第 6 节） | G | 目录拆分、三类型号、Binding.Operation、预检顺序、归属 |
| 4 | B 共同运行时（第 7 节） | A | 调用运行时抽取、按操作分派、unary 入口、PhaseResponse |
| 4 | C 资源与计价（第 8 节） | A | unary reader、子策略、Usage.Modalities、ImagePricing |
| 5 | T 验收设施（第 9 节） | B、C | unary fixture、classify oracle 入口、追溯、账本 routes、矩阵、live、审计 |
| 6 | D TypeSafe（第 10 节） | T | P07 |
| 7 | E OpenAI 图像（第 11 节） | D | P08 |
| 8 | F Google 图像（第 12 节） | E | P09 |
| 9 | H 发布证据 | 全部 | 九条 combo 证据、目录快照、压力报告 |

每个工作包先写出对应小节"失败场景"中的 fixture，再写代码（AGENTS.md 测试原则）。每个工作包单独合入、单独通过门禁。

## 4. G0：基线迁移到 pi-ai 1.0.0

### 4.1 为什么只保留一个 oracle

[ADR-0004](../../docs/adr/0004-barness-ai-frozen-pi-oracle.md) 规定"更换冻结版本即更换兼容基线：先改规范，再重做等价核验、更新 `provenance.json` 与分类账"，runner 会拒绝版本或模型数据哈希不符的副本。并行两个 oracle 需要第二个 node 目录、runner 与 provenance，长期维护违背工程原则 10。0.87.1 → 1.0.0 在 barness 四条聊天协议上的可观察差异，按 CHANGELOG 逐条核对就是第 5 节的五项，迁移成本可控。

### 4.2 步骤

1. **替换发布件。** `pioracle/node/package.json` 固定 `@earendil-works/pi-ai@1.0.0` 与对应的 `@earendil-works/pi-telemetry`；`npm install --ignore-scripts` 生成新 lockfile；记录 tarball integrity、npm gitHead、`dist/providers/data/` 每个文件的 sha256 与模型数据哈希，写入 `provenance.json` 与 PROVENANCE.md 的表格。
2. **等价核验。** 按 PROVENANCE.md 现有的"Why the npm release is the frozen commit"步骤，确认发布件 gitHead 与 tag `v1.0.0` 在 `packages/ai`、`packages/telemetry` 的差异，并把冻结源码填入发布数据后构建，与发布件逐字节比较。结果写入 PROVENANCE.md。
3. **runner 适配。** 1.0 的入口变化如下，均已在 tag 源码核对：
   - `api/*` 模块仍导出 `stream`、`streamSimple`（如 [openai-responses.ts](https://github.com/earendil-works/pi/blob/v1.0.0/packages/ai/src/api/openai-responses.ts) 第 127、238 行）。
   - `providers/{openai,anthropic,google,deepseek}.models` 仍存在。
   - 根入口仍导出 `normalizeContext`（来自 `utils/transcript.ts`）与 `calculateCost`（来自 `models.ts`，签名改为接受 `AnyModel`）；`api/simple-options` 仍导出 `adjustMaxTokensForThinking`。
   - 根入口不再导出 `images-models.ts` 与 `models-store.ts`；runner 不用它们。
   - 1.0 的 `getModels` 只返回聊天型号，聊天入口对非聊天型号运行时抛 `ModelsError`（CHANGELOG 0.99.0 第 47–54 行）。runner 从 `*.models` 目录按 `api` 过滤，不受影响；但过滤时需排除非聊天类型，因为 1.0 的生成目录同一文件里分别导出 `*_MODELS`、`*_IMAGE_MODELS`、`*_CLASSIFIER_MODELS`。
   - 为 T 工作包新增 `classify` 入口，见第 9.2 节。
4. **改写账本。** 运行全量差分（`BARNESS_AI_PIDIFF=1`）。第 5 节五项会产生新发现，逐条把账本条目或 fixture 期望改成 1.0 行为，`decision` 写"原行为 → 1.0 行为 → 依据"。其余发现必须为 0；若出现第 5 节以外的差异，先查 CHANGELOG 是否漏记，再由维护者决定。
5. **目录比对。** 用 1.0 模型数据跑 `TestCatalogInclusion/builtin-models-are-pi's`。价格变化与新型号按 ADR-0018 处理：硬约束未实现的型号不列入，列入的型号字段与 pi 一致。内建目录升版本，`go run ./ai/release/cmd/releasegate -write-snapshot` 重写 `catalog-snapshot.json`。
6. **文档。** 修订 ADR-0004 的版本与 provenance 段落；[differences.md](../../docs/barness-ai/differences.md) 首段的基线版本与 commit；spec 的兼容基线段。
7. **删除旧副本。** 迁移合入后，0.87.1 的 lockfile 与 provenance 不保留。

### 4.3 失败场景

- runner 遇到版本或模型数据哈希不符的副本时拒绝运行。
- 全量差分在五项以外有任何 pending 时门禁失败。
- 内建型号与 1.0 数据有未登记的字段差异时目录包含测试失败。

## 5. G：聊天五项行为

五项都是 0.87.1 下有意的对齐行为，迁移后改为 1.0 行为。每项都会改动参与差分的 fixture，必须与 G0 同一基线下验证。

### 5.1 无法解析的 Retry-After

**现状。** `serverDelayMs`（[retry.go](../../ai/retry.go)）对 `retry-after-ms` 只要不是 NaN 就采用（含 ±Infinity）；对 `retry-after` 先 `jsParseFloat`，再 RFC 9110 日期，都不行就返回 `NaN, true`；`delay` 把 NaN 交给 `msDuration`，结果为 0，即立即重试。注释写"As in pi"，[ADR-0006](../../docs/adr/0006-barness-ai-retry-policy-on-binding.md) 第 24 行把自由格式日期立即重试记为已知差异，fixture `retry-after-unparseable`（[responses/retry.json](../../ai/e2e/testdata/responses/retry.json) 第 123–129 行，`sleepsMs:[0]`，无 `pidiffSkip`）锁定该行为。

**pi 1.0**（[provider-retry.ts:51-67](https://github.com/earendil-works/pi/blob/v1.0.0/packages/ai/src/utils/provider-retry.ts#L51-L67)）：

```ts
const retryAfterMs = error.headers?.get("retry-after-ms");
if (retryAfterMs) {
  const value = Number.parseFloat(retryAfterMs);
  if (Number.isFinite(value)) return validateServerRetryDelayMs(value, maxRetryDelayMs, error.message);
}
const retryAfter = error.headers?.get("retry-after");
if (retryAfter) {
  const seconds = Number.parseFloat(retryAfter);
  const delayMs = Number.isNaN(seconds) ? Date.parse(retryAfter) - Date.now() : seconds * 1000;
  if (Number.isFinite(delayMs)) return validateServerRetryDelayMs(delayMs, maxRetryDelayMs, error.message);
}
const exponentialDelay = Math.min(0.5 * 2 ** retryIndex, 8) * 1000;
return exponentialDelay * (1 - Math.random() * 0.25);
```

**改法。** 只改 `serverDelayMs`，`delay` 的结构不动：

- `retry-after-ms`：`jsParseFloat` 结果必须有限（非 NaN、非 ±Inf）才返回 `ms, true`，否则继续看 `retry-after`。
- `retry-after`：`jsParseFloat` 不是 NaN 时取 `s*1000`（可能是 ±Inf）；是 NaN 时取 HTTP 日期差；都没有时为 NaN。结果有限才返回 `ms, true`，否则返回 `0, false`，由 `delay` 走指数退避。注意 `retry-after: Infinity` 在 pi 1.0 中 `seconds` 为 Infinity、不会再尝试日期，直接退避；Go 端保持相同顺序。
- 负的有限值（过去的日期）照旧返回，`msDuration` 记为 0，与 pi 的 `abortableSleep(Math.max(0, ms))` 一致。
- 更新 `serverDelayMs` 与 `parseHTTPDate` 的注释：自由格式日期在 barness 无法解析时走退避，而 pi 的 `Date.parse` 可能解析成功并等待；这仍是已知差异，但方向从"立即重试"变为"退避"。

**fixture。** `retry-after-unparseable` 的期望改为退避延迟：`maxRetries:1` 时第一次退避为 `500ms × (1 − jitter×0.25)`，在可替换时钟下 jitter 固定，`sleepsMs` 写确定值。新增 `retry-after-ms-infinity`（`retry-after-ms: Infinity` 与有效 `retry-after: 2` 同时出现，期望等待 2000 ms）和 `retry-after-infinity`（期望退避）。这些用例都参与差分。

**ADR。** 修订 ADR-0006 第 24 行。

### 5.2 Responses fast 服务等级

**pi 1.0**（[openai-responses.ts:387-400](https://github.com/earendil-works/pi/blob/v1.0.0/packages/ai/src/api/openai-responses.ts#L387-L400)）：`flex` 为 0.5；`priority` 与 `fast` 为 `model.id === "gpt-5.5" ? 2.5 : 2`；其他为 1。等级取 `response?.service_tier ?? options.serviceTier`（[openai-responses-shared.ts:579-582](https://github.com/earendil-works/pi/blob/v1.0.0/packages/ai/src/api/openai-responses-shared.ts#L579-L582)）。Codex adapter 没有 `fast`，与 barness 无关。

**改法。** `serviceTierMultiplier` 的 `case "priority":` 改为 `case "priority", "fast":`。等级来源逻辑不变。

**fixture。** 沿用 pi 修复提交 `a6ca86102` 的测试用例：请求 `priority` 响应 `fast` 为 2 倍；请求与响应都是 `fast` 为 2 倍；`gpt-5.5` 的 `priority` 为 2.5 倍。型号用 barness 内建目录里有的 Responses 型号。

**ADR。** 修订 [ADR-0010](../../docs/adr/0010-barness-ai-usage-reporting-and-price-snapshot.md) 决策五的倍率表。

### 5.3 Anthropic 增量中的 1 小时缓存写入

**pi 1.0**（[anthropic-messages.ts:843-849](https://github.com/earendil-works/pi/blob/v1.0.0/packages/ai/src/api/anthropic-messages.ts#L843-L849)）：`message_delta` 的 usage 带 `cache_creation.ephemeral_1h_input_tokens` 时，`output.usage.cacheWrite1h` 被**覆盖**为该值；不带时保持原值。注释说明 Vercel AI Gateway 会在增量里给出 TTL 分项，SDK 只在 `message_start` 上声明它。

**改法。** `anthropicUsage.delta` 在 `a.take(w)` 之后加：`w.CacheDetails != nil && w.CacheDetails.Ephemeral1h != nil` 时 `a.u.CacheWrite1h = Value(*w.CacheDetails.Ephemeral1h)`。`anthropicWireUsage` 已有该字段。`reporting()` 不变：该明细缺失仍读作没有，不影响 `complete`。

**fixture。** `message_start` 报 1h=0、`message_delta` 报 1h=N，期望 `cacheWrite1h` 为 N、成本按 1h 费率；两次 delta 先 N 后 M，期望为 M 而不是 N+M；delta 不带该字段时保持 start 的值。

**ADR。** 修订 [ADR-0011](../../docs/adr/0011-barness-ai-anthropic-messages-adapter.md) 决策三：写明增量会覆盖 1h 明细，完整性规则不变。

### 5.4 型号默认 samplingParams

**pi 1.0**：修复提交 `c01f687e5` 把合并从 `simple-options.ts` 移到 adapter。`simple-options.ts` 第 29 行改为只传 `options?.samplingParams`；`openai-responses.ts` 第 381–382 行、`azure-openai-responses.ts` 第 348 行、`openai-completions.ts` 第 999 行为 `Object.assign(params, model.samplingParams, options?.samplingParams)`。[types.ts:200-206](https://github.com/earendil-works/pi/blob/v1.0.0/packages/ai/src/types.ts#L200-L206) 说明只有 OpenAI 兼容 adapter 应用它，其他 API 忽略。

**改法。**

- `SimpleOptions.resolve` 删除与 `m.SamplingParams` 合并的分支，只保留调用级的 `o.SamplingParams`。
- `chat_request.go` 与 `responses_request.go` 的 `applySamplingParams(out, opts.SamplingParams)` 改为先应用 `model.SamplingParams`、再应用 `opts.SamplingParams`。这样 full 与 simple 入口都在同一处合并一次，顺序为命名字段 → 型号默认 → 调用级。
- 保留字段检查：现在 `checkReservedKeys` 只在 Options 的 `validate` 中检查调用级 map。型号默认值来自目录，也必须检查：在目录校验（第 6.2 节）中对 Responses/Chat 型号的 `SamplingParams` 执行同一个保留字段检查，构造时拒绝。
- 更新 `ChatOptions.SamplingParams` 与 `ResponsesOptions.SamplingParams` 的注释（[chat_options.go](../../ai/chat_options.go) 第 24–28 行、[responses_options.go](../../ai/responses_options.go) 第 21–25 行）。
- Anthropic、Gemini 没有该字段，不变。内建目录目前没有型号设置 `SamplingParams`，只影响宿主自带目录。

**fixture。** 现有 `simple-sampling`（[chat/options.json](../../ai/e2e/testdata/chat/options.json) 第 4955 行附近）保持通过；新增 full 入口版本：`modelPatch` 设型号默认，`options` 设部分覆盖，期望请求体为合并结果；新增调用级覆盖型号默认的同名键；新增目录中带保留字段的型号构造失败。Responses 与 Chat 各一组。

### 5.5 Responses 未完成工具调用

**pi 1.0**（[openai-responses-shared.ts:760-777](https://github.com/earendil-works/pi/blob/v1.0.0/packages/ai/src/api/openai-responses-shared.ts#L760-L777)）：先检查是否收到终态事件；然后若 `output.stopReason === "toolUse"`，遍历全部 `toolCall` 块，块上仍有 `partialJson` 或 `customInput` 即抛 `OpenAI Responses stream completed with an unfinished tool call: ${name} (${id})`。两个暂存字段只在对应的 `output_item.done` 中删除（同文件第 720、734 行附近），所以"暂存字段仍在"等于"done 没有到"。

**现状。** `finalize` 没有任何检查。`open` 对重复的 `output_index` 直接覆盖 slot，`itemDone` 处理后删除 slot，所以只检查剩余 slot 会漏掉被覆盖的块。

**改法。** 不依赖 slot，按块追踪：

- `responsesParser` 增加 `unfinished map[int]toolCallRef`，键为内容块下标，值为工具名与 id（id 即 `item.CallID+"|"+item.ID`，与 `toolCallStart` 一致）。
- `open` 的 `function_call` 分支登记 `s.index`。
- `itemDone` 的 `function_call` 分支在 `toolCallEnd` 之后删除 `i`。
- `finalize` 在 `stop == StopReasonStop && p.out.hasToolCall()` 分支中，设置 `StopReasonToolUse` 之前，若 `unfinished` 非空，按内容块下标从小到大取第一个，返回 `newError(CodeProtocol, PhaseStream, "OpenAI Responses stream completed with an unfinished tool call: "+name+" ("+id+")")`。消息文本与 pi 一致，供差分比较。
- 现有的终态事件检查（`readResponsesStream` 第 126–128 行附近）保留在前，顺序与 pi 相同。

**fixture。** 工具调用没有 `output_item.done` 就 `response.completed`；两个调用共用一个 `output_index`，第一个被覆盖后只有第二个收到 done；缺失 `output_index`；正常完成的对照组。全部参与差分。

## 6. A：类型与授权

### 6.1 型号类型

```go
type Operation string

const (
    OperationChat       Operation = "chat"
    OperationImage      Operation = "image"
    OperationClassifier Operation = "classifier"
)

// Model 不变，表示聊天型号。

type ImageModel struct {
    Provider     ProviderID        `json:"provider"`
    API          API               `json:"api"`
    ID           string            `json:"id"`
    Name         string            `json:"name"`
    Input        []Modality        `json:"input"`  // text、image
    Output       []Modality        `json:"output"` // text、image
    Capabilities ImageCapabilities `json:"capabilities"`
    Pricing      ImagePricing      `json:"pricing"`
}

type ImageCapabilities struct {
    MaxReferenceImages int      `json:"maxReferenceImages"`
    MaxOutputImages    int      `json:"maxOutputImages"`    // 1 表示不承诺多张
    Sizes              []string `json:"sizes,omitempty"`    // OpenAI 的 WIDTHxHEIGHT 或 auto
    ImageSizes         []string `json:"imageSizes,omitempty"` // Google 的 1K/2K/4K
    AspectRatios       []string `json:"aspectRatios,omitempty"`
    Mask               bool     `json:"mask,omitempty"`
    TransparentBackground bool  `json:"transparentBackground,omitempty"`
    InputFidelity      []string `json:"inputFidelity,omitempty"` // 空表示必须省略
}

type ClassifierModel struct {
    Provider      ProviderID             `json:"provider"`
    API           API                    `json:"api"`
    ID            string                 `json:"id"`
    Name          string                 `json:"name"`
    ContextWindow int                    `json:"contextWindow"`
    Capabilities  ClassifierCapabilities `json:"capabilities"`
    Cost          ModelCost              `json:"cost"`
}

type ClassifierCapabilities struct {
    Kinds         []string `json:"kinds"`         // choice、score、bool
    MaxChoices    int      `json:"maxChoices"`    // System One: 255
    MinScoreLevels int     `json:"minScoreLevels"` // 2
    MaxScoreLevels int     `json:"maxScoreLevels"` // 10
}

type Catalog struct {
    Version          string            `json:"version"`
    Models           []Model           `json:"models"`
    ImageModels      []ImageModel      `json:"imageModels,omitempty"`
    ClassifierModels []ClassifierModel `json:"classifierModels,omitempty"`
}
```

- 三个类型各自声明身份字段，不嵌入公共结构。把 `Model` 的 `Provider`、`API` 改为嵌入结构的提升字段，会让 keyed struct literal 编译失败；仓库内受影响的是 `catalog.go` 五个构造闭包，仓库外宿主未知。
- `omitempty` 保证只有聊天型号的目录 JSON 不变；但 `Hash` 的输入会变，见第 6.2 节。
- 内部用一个私有描述符 `modelKey{op, provider, api, id}` 统一索引与重复检查，不导出。

### 6.2 目录拆分与校验

`catalog.go` 先按职责拆分，再加新类型。建议：

| 文件 | 内容 |
| --- | --- |
| `catalog.go` | `Catalog`、公开查找与列表、`BuiltinCatalog` |
| `catalog_models.go` | `Model`、`ModelCompat`、`ImageModel`、`ClassifierModel` 与能力类型 |
| `catalog_builtin.go` | 五个内建数据构造闭包与新增的图像、分类条目 |
| `catalog_index.go` | `modelKey`、索引构建、`clone`、`Hash`、`validate` |

校验（构造 Client 时执行，失败为 `ConfigError{Field: "Catalog"}`）：

- 同一 `modelKey` 出现两次即拒绝。现有 `lookup` 取首个匹配的行为删除。
- `ImageModel.Input`、`Output` 只能是 text、image，且 `Output` 必须含 image；能力与模态矛盾即拒绝，例如 `MaxReferenceImages > 0` 而 `Input` 不含 image。
- `ImageModel.Pricing` 对 `Input`、`Output` 声明的每个模态都必须有费率（第 8.4 节）。
- `ClassifierModel.Capabilities.Kinds` 非空，范围值合法。
- 所有价格非负且有限（沿用 `ModelCost.problem`）。
- Responses/Chat 型号的 `SamplingParams` 不含保留字段（第 5.4 节）。

`clone` 深复制新增的切片、能力结构与价格。`Hash` 对整个 `Catalog` 做 JSON 编码后 SHA-256，三类型号自然纳入；内建目录升版本，重写 `catalog-snapshot.json`。

查找与列表：现有 `Catalog` 的聊天查找与 `Models` 字段含义不变；新增私有 `lookupImage`、`lookupClassifier`，公开 `ImageModelsOf(provider, api)` 一类的列表方法按需加入，命名与现有公开方法保持一致。

### 6.3 Binding.Operation

```go
type Binding struct {
    // ... 现有字段
    // Operation is the one model operation calls through this binding
    // perform. The zero value means chat: a binding written before
    // operations existed could only have meant chat, so the default widens
    // nothing. It is the deliberate exception to this type's fail-closed
    // zero values (Enabled, Active). Any other unknown value is refused.
    Operation Operation
}
```

`resolveBinding` 在 `AuthKind` 检查旁加：`Operation` 为空规范化为 chat；不是三种之一即 `newError(CodeInvalidRequest, PhaseBinding, "binding operation is not supported")`。

### 6.4 预检顺序

`execute`（或第 7 节抽取后的共同运行时）在 `resolveBinding` 之后：

1. 入口给出期望操作（`Stream`/`Complete` 为 chat，`GenerateImages` 为 image，`Classify` 为 classifier）。
2. `binding.Operation != want` 即 `newError(CodeTenantDenied, PhaseCapability, "binding does not allow this operation")`，不读凭据、不发请求。
3. 按操作查 adapter：`c.adapters[adapterKey{op, binding.API}]`。
4. `authorizeModel` 按操作查对应类型目录，再与 `AllowedModels` 求交集。
5. Options 校验，含 `options.api() == binding.API`。
6. 凭据解析与 `checkSnapshot`，不变。

### 6.5 归属

`CallAttribution` 增加 `Operation Operation`（JSON `operation`）。由入口在 `run` 开始时填写，所以 `CallStarted` 观测就带它。`Resolved`、`ProviderID`、`API`、`ModelID`、`AccountScopeID` 仍在快照一致后填写。审计的 Observer 字段白名单增加 `operation`。

### 6.6 失败场景

同一 Client 混合三种操作；跨租户同名 binding；零值 binding 调用 `GenerateImages`、`Classify` 被拒且不读凭据；`Operation: "images"` 等拼错值被拒；同一型号 ID 同时出现在聊天与图像目录，聊天 binding 不能用它调图像；目录重复键、矛盾能力、缺模态费率、型号默认 samplingParams 含保留字段时构造失败；构造后修改 `cfg.Catalog` 不影响 Client。

## 7. B：共同运行时

### 7.1 抽取调用运行时

从 `Client.run`/`execute` 抽出操作无关的部分：

```go
// callRuntime is one logical call's operation-independent state: the
// pinned binding, credential, catalog snapshot, metadata and the initial
// request that owns attempts and admission.
type callRuntime struct {
    op       Operation
    scope    CallScope
    target   Target
    binding  Binding
    cred     Credential
    meta     *CallMetadata
    initial  *initialRequest
    header   http.Header
    hooks    boundHooks
    limits   byteLimits
}

// resolve runs scope → binding → operation → capability → credential →
// snapshot for op and returns the pinned runtime, or the failure. It does
// not send anything.
func (c *Client) resolve(ctx context.Context, op Operation, scope CallScope,
    target Target, hooks Hooks) (callRuntime, adapterSet, *Error)
```

- 聊天路径：`run` 调用 `resolve`，之后照旧做历史准备、`simpleOptions`、`assembler` 与 `ad.stream`。聊天的 `Result` 与事件不变。
- unary 路径：`runUnary` 调用 `resolve`，交给图像或分类 adapter，返回类型化结果。
- `observations.callStarted/callFinished` 改为接收操作无关的终态摘要：

```go
type callOutcome struct {
    StopReason StopReason
    Usage      Usage
    Err        *Error
}
```

聊天从 `res.Message` 取，unary 从各自结果取。

### 7.2 adapter 契约与注册

```go
type adapterKey struct {
    op  Operation
    api API
}

type imageAdapter interface {
    headers(call unaryCall) http.Header
    generate(ctx context.Context, call unaryCall, req ImagesRequest, opts ImageOptions) (imageOutput, *Error)
}

type classifierAdapter interface {
    headers(call unaryCall) http.Header
    classify(ctx context.Context, call unaryCall, req ClassifierRequest, opts ClassifierOptions) (classifierOutput, *Error)
}

// unaryCall is adapterCall without the chat history: endpoint, key, the
// resolved model of the call's operation, header, hooks, initial request
// and limits.
type unaryCall struct { /* ... */ }
```

- 现有 `adapter` 接口与 `adapterCall` 不变，作为聊天 adapter。
- registry 改为三张显式表，或一张 `map[adapterKey]any` 再按操作断言；推荐三张表，避免运行时断言：`chat map[API]adapter`、`image map[API]imageAdapter`、`classifier map[API]classifierAdapter`。
- 新增常量：`APIOpenAIImages = "openai-images"`、`APIGoogleInteractions = "google-interactions"`、`APITypeSafeSystemOne = "typesafe-system-one"`、`ProviderTypeSafe = "typesafe"`。
- 首期不开放动态注册。

### 7.3 公开入口

```go
func (c *Client) GenerateImages(ctx context.Context, scope CallScope, target Target,
    req ImagesRequest, opts ImageOptions) (ImagesResult, error)

func (c *Client) Classify(ctx context.Context, scope CallScope, target Target,
    req ClassifierRequest, opts ClassifierOptions) (ClassifierResult, error)
```

`HookedClient` 提供同名方法。`ImageOptions`、`ClassifierOptions` 都是 `Options` 接口的实现集合：`OpenAIImagesOptions`、`GoogleImagesOptions`、`TypeSafeOptions`。为了让签名有类型约束，定义：

```go
type ImageOptions interface { Options; imageOptions() }
type ClassifierOptions interface { Options; classifierOptions() }
```

入口在接收时 `clone` 请求与选项（与 `newCall` 相同），先按已知长度检查限额再复制（第 8.1 节）。

### 7.4 PhaseResponse

`errors.go` 增加 `PhaseResponse Phase = "response"`，用于 unary 响应体的读取、解码与校验失败。同步更新 [contract.md](../../docs/barness-ai/contract.md) 第 4 节的 Phase 列表与错误表。HTTP 状态分类沿用 `codeForStatus`：401/403 为 `upstream_auth`，429 为 `rate_limited`，5xx/408/409 为 `upstream_error`，其他 4xx 为 `invalid_request`。

### 7.5 发送与重试

unary adapter 用 Gemini adapter 已经使用的直连方式：

1. 构建请求体（JSON 对象）。
2. `hooks.payload` 执行一次，得到最终请求体；回调后重新校验（第 11.5、12.6、10.5 节各自的规则）。
3. 用最终请求体对 `MaxRequestBytes` 检查。
4. `initial.send` 发送；重试只由 `Binding.Retry` 决定，复用冻结的请求体；每次尝试有独立 AttemptID 与准入许可。
5. 2xx 时执行一次 `hooks.response`，再以 unary reader 读完整个 body（第 8.2 节），再解码、校验。
6. 解码前先 `initial.usage(reporting, usage)` 登记用量，校验失败也保留。
7. 成功响应之后的任何失败都不重放。

许可在 adapter 返回、body 已关闭后释放，沿用 `defer initial.done()`。

### 7.6 失败场景

取消、`ConnectTimeout`、响应头超时、读空闲超时、`CallTimeout`、准入拒绝；200 后读断不重放；坏 JSON 不重放；所有失败路径关闭 body 并释放许可（用现有 probe 断言）；`OnPayload` 只执行一次；重试时每个 attempt 归属正确；`PhaseResponse` 在错误中出现且 Phase 为空时 `errors.Is` 匹配仍按 Code。

## 8. C：资源与计价

### 8.1 输入检查

- `ImagesRequest.ReferenceImages`：数量对 `min(ImagePolicy.MaxInputImages, model.Capabilities.MaxReferenceImages, 协议硬限制)`；每张 `base64Size` 对 `MaxImageBytes`；MIME 用 `isImageMediaType`。mask 计入输入图片数量与 `MaxImageBytes`。
- `ClassifierRequest`：问题数对 `MaxQuestions`；`State` 字节数对 `MaxStateBytes`；每个问题的 instructions 与 criteria 编码后字节数对 `MaxQuestionBytes`。
- 先用已知长度检查，再深复制，避免先复制超限数据。
- 回调后对最终请求体重新执行同样的检查（回调可能插入图片或问题）。
- `MaxImageBytes` 的注释（[policy.go](../../ai/policy.go) 第 18–20 行）改为覆盖聊天历史中的图片、参考图与 mask。

### 8.2 unary 响应读取

`limitBody` 增加一个参数说明响应类型：

```go
type bodyKind int

const (
    bodySSE bodyKind = iota
    bodyJSON
)

func (l byteLimits) limitBody(res *http.Response, kind bodyKind)
```

- `bodySSE`：现有 `streamBody`，不变。
- `bodyJSON` 的 2xx：复用 `errorBody` 的总字节计量逻辑，把它参数化为 `totalBody{limit, field}`，`field` 为 `MaxOutputBytes`；非 2xx 仍为 `MaxErrorBodyBytes`。
- 超限返回 `limitExceeded`，由 `http_failures` 现有逻辑归为 `resource_limit`，Phase 为 `response`。
- 现有 SDK 中间件 `limitBodies` 调用处传 `bodySSE`。

### 8.3 子策略

```go
type ResourcePolicy struct {
    // ... 现有标量字段
    // Image enables GenerateImages; nil disables it.
    Image *ImagePolicy
    // Classifier enables Classify; nil disables it.
    Classifier *ClassifierPolicy
}

type ImagePolicy struct {
    MaxInputImages           int
    MaxOutputImages          int
    MaxOutputImageBytes      int64
    MaxTotalOutputImageBytes int64
}

type ClassifierPolicy struct {
    MaxQuestions     int
    MaxStateBytes    int64
    MaxQuestionBytes int64
}
```

- `validate`：子策略非 nil 时每个字段必须为正；`MaxTotalOutputImageBytes ≤ MaxOutputBytes`，因为 base64 后的输出图片总在响应体里。
- `NewClient` 深复制两个指针，保持 `Config` "构造后修改原值不影响 Client" 的承诺（[client.go](../../ai/client.go) 第 25–26 行）。
- 子策略为 nil 时，对应入口在 scope 阶段以 `invalid_request` 失败："the client policy does not enable image generation"。
- `localassembly`、`hostintegration` 的示例策略各加一份图像与分类子策略，数值附适用负载与依据，并由压力场景验证（ADR-0002 的要求）。

### 8.4 用量

三种操作共用 `Usage` 与 `Attempt.UsageReporting`，不新增平行的用量结构或第二个上报状态。理由：平行结构会让通用汇总把"未上报加零值"误读为没有消耗；[ADR-0010](../../docs/adr/0010-barness-ai-usage-reporting-and-price-snapshot.md) 已规定零值不表示免费、partial 表示缺失项按 0 计；pi 自己也用聊天 `Usage` 表示图像用量（[openrouter-images.ts:167-198](https://github.com/earendil-works/pi/blob/v1.0.0/packages/ai/src/api/openrouter-images.ts#L167-L198)）。

```go
// ModalityUsage splits an image operation's tokens by modality. Each count
// is unset when the provider did not report it.
type ModalityUsage struct {
    InputText   Nullable[int64] `json:"inputText,omitzero"`
    InputImage  Nullable[int64] `json:"inputImage,omitzero"`
    OutputText  Nullable[int64] `json:"outputText,omitzero"`
    OutputImage Nullable[int64] `json:"outputImage,omitzero"`
}

// Usage 新增：
//   Modalities *ModalityUsage `json:"modalities,omitempty"`
```

聊天不设置 `Modalities`，JSON 不变，pi 差分投影不受影响。`TotalTokens` 注释改为"厂商报告了总量时取厂商值，否则为分项之和"，与 Anthropic 已有的计算方式一致。

各协议映射与完整性见第 10.6、11.6、12.7 节。

### 8.5 图像计价

```go
// ImagePricing is an image model's price in US dollars per million tokens,
// by modality.
type ImagePricing struct {
    InputText   Nullable[float64] `json:"inputText,omitzero"`
    InputImage  Nullable[float64] `json:"inputImage,omitzero"`
    OutputText  Nullable[float64] `json:"outputText,omitzero"`
    OutputImage Nullable[float64] `json:"outputImage,omitzero"`
    CacheReadText  Nullable[float64] `json:"cacheReadText,omitzero"`
    CacheReadImage Nullable[float64] `json:"cacheReadImage,omitzero"`
    // Source is the official pricing URL and the date the rates were read.
    Source string `json:"source"`
}
```

费用计算：

```text
Cost.Input  = InputText×Modalities.InputText + InputImage×Modalities.InputImage   （缓存读取另计入 CacheRead）
Cost.Output = OutputText×Modalities.OutputText + OutputImage×Modalities.OutputImage
Cost.Total  = Input + Output + CacheRead
```

- 每个乘积单独计算再相加，与 ADR-0010 决策五对舍入的要求一致。
- 目录校验保证声明的模态都有费率，所以完整上报一定能算出费用。
- 缺少某个模态分项时，该分项按 0 计，`UsageReporting` 为 partial，与 ADR-0010 对 partial 的定义一致；不按总量与某一费率猜测。
- OpenAI 型号不设缓存读取费率：官方说明缓存输入价"doesn't apply to direct Images API requests, including /v1/images/edits"（[图像指南](https://developers.openai.com/api/docs/guides/image-generation)）。
- 官方"每张图约多少钱"只是派生值，不叠加计价。分辨率与 token 的对应因型号而异，只写进能力说明，计价只依据厂商报告的 token。
- 实施时从官方页面读取费率，填 `Source`；本文中的数字只作示例。

### 8.6 失败场景

JSON 响应大于 `MaxFrameBytes`、小于 `MaxOutputBytes` 时成功；大于 `MaxOutputBytes` 时在读取中失败，`resource_limit`/`response`；输入图片数量、单张、mask 超限；输出图片数量、单张、总量超限；base64 膨胀导致的估算超限；回调插入大图或多余问题被拒；子策略为 nil 时入口被拒；构造后修改原子策略不影响 Client；用量缺分项为 partial 且费用只含已知分项。

## 9. T：验收设施

### 9.1 unary fixture

`fixtureScenario` 已有 `Replies []fixtureReply`（状态加原始 body），`provider.Reply` 已能返回任意 JSON body。新增：

- `fixtureScenario.Entry` 新值 `generateImages`、`classify`。
- `fixtureScenario.Request json.RawMessage`：unary 入口的请求，按入口解码为 `ImagesRequest` 或 `ClassifierRequest`。
- `fixtureExpect` 增加 `images`、`answers` 断言字段；现有 `requests`、`attempts`、`stopReason`、`error`、`usage` 断言沿用。
- `fixtureProtocol.sse` 对 unary 协议为 nil；runner 遇到 unary 入口时只走 `Replies`。

新目录：`ai/e2e/testdata/typesafe`、`ai/e2e/testdata/openai-images`、`ai/e2e/testdata/google-interactions`。

### 9.2 classify 的 oracle 入口

TypeSafe 的请求编码与基本答案转换进入 1.0 差分。`runner.mjs` 新增：

```js
"typesafe-system-one": {
  api: "@earendil-works/pi-ai/api/typesafe-system-one",
  catalogs: [{ models: "@earendil-works/pi-ai/providers/typesafe.models", catalog: "TYPESAFE_CLASSIFIER_MODELS" }],
},
```

入口 `classify` 调用该模块导出的 `classify(model, context, options)`（[typesafe-system-one.ts](https://github.com/earendil-works/pi/blob/v1.0.0/packages/ai/src/api/typesafe-system-one.ts)），记录发出的请求与返回的 `ClassifierResult`。比较时：

- 只运行 pi 能表达的输入子集：`state` 为对象、`instructions` 为字符串、choice criteria 为字符串。
- barness 的 `ClassifierResult` 投影到 pi 形状：去掉 score 的 `probabilities`、`legend`，`ResponseModel` 不参与比较。
- 不比较 `usage.cost`：pi 把直连 `jev-latest` 定价为 0（[generate-models.ts:2674-2698](https://github.com/earendil-works/pi/blob/v1.0.0/packages/ai/scripts/generate-models.ts#L2674-L2698)），barness 用官方价。
- pi 的 classifier 默认重试 2 次（[system-one-shared.ts:220](https://github.com/earendil-works/pi/blob/v1.0.0/packages/ai/src/api/system-one-shared.ts#L220)）；runner 显式传 `maxRetries` 与 barness 的 binding 一致，这是配置位置差异，不登记。
- 型号 ID：pi 目录只有 `jev-latest`，barness 只列 `jev-1.13.0`。差分用 `modelPatch` 把 pi 侧 id 改为 `jev-1.13.0`，比较发出的 `model` 字段。

### 9.3 账本、追溯与目录包含

- **账本 `routes`**：增加 `{provider: "openai", api: "openai-images"}` 与 `{provider: "google", api: "google-interactions"}`，`classification: "extension"`，`decision` 写"pi 1.0 只有 openrouter-images，没有该原生路线；以官方协议 fixture 离线证明，不计为差分通过"，引用本设计与 ADR-0014。这两个目录的 fixture 全部带 `pidiffSkip`，并各有一个离线用例 `P08-E11-registered-as-extension`、`P09-E11-registered-as-extension` 断言已登记，照 `P05-E11-registered-as-extension` 的写法。
- **账本 `decisions`**：TypeSafe 的扩展逐条登记：更宽的 state/instructions/criteria、score 分布与图例、严格答案校验、版本化型号 ID、目录价格。
- **traceability.json**：`scenarios` 增加 `"P07": "^(PIDIFF-)?P07-"`、`"P08": "^P08-"`、`"P09": "^P09-"`；`p0` 加入三者；`differential.required` 加入 P07；`differential.routes` 加入 P08、P09；`liveCombos` 加入 `typesafe-classifier`、`openai-images`、`google-interactions-image`；新增 T/C/H 条目对应本设计的验证组。
- **目录包含测试**：`TestCatalogInclusion/builtin-models-are-pi's` 增加豁免：`ImageModels` 中 `openai-images`、`google-interactions` 的型号不与 pi 比较；`ClassifierModels` 只比较 pi 拥有的字段（`api`、`contextWindow`），`id` 与价格差异按扩展登记。

### 9.4 支持矩阵与 live

- 矩阵行增加 `operation` 字段；`matrix.go` 的一致性检查同时校验 operation、provider、API。六个聊天行的 `operation` 为 chat。
- `combos_test.go` 新增三个 combo，`spec` 为 P07–P09；`sdk` 为 `directHTTP()` 的变体，注明本设计。
- `main_test.go` 的预算按操作分开：图像按张数与分辨率（首期每 combo 不超过 4 张、最大 1K/1024×1024），分类按问题数。
- live 场景：
  - P07：三种问题混合一次；单选一次；故意的 422（空 questions）一次。
  - P08：生成一张；JSON 编辑一张（live gate）；带 mask 的编辑一张（型号支持时）。
  - P09：生成一张；带一张参考图编辑一张；请求断言 `store:false` 且省略 `delivery`，响应断言内联（工单 14 / ADR-0023）。
- 只验证可读图片、格式、尺寸与数量约束和协议结果，不做像素断言。

### 9.5 审计

- Observer 字段白名单增加 `operation` 与 `usage.modalities.*`。
- live 响应头白名单增加 `x-typesafe-request-id`。
- 断言 Observer 记录不含 prompt、base64、state、answers 或错误正文。

## 10. D：TypeSafe System One

### 10.1 绑定与端点

`ProviderTypeSafe` × `APITypeSafeSystemOne` × classifier；endpoint `https://api.typesafe.ai/v1`；`POST {endpoint}/systemone`；请求头 `Authorization: Bearer <key>`、`Content-Type: application/json`（[API](https://docs.typesafe.ai/api)）。

直连 HTTP：TypeSafe 只有官方 Python 与 JS SDK，没有 Go SDK，属于 spec §5 的例外条款；pi 同样直连。

### 10.2 公共类型

```go
type ClassifierRequest struct {
    State     json.RawMessage // JSON string、object 或 array
    Questions map[string]ClassifierQuestion
}

type ClassifierQuestion interface{ classifierQuestion() }

type ChoiceQuestion struct {
    Instructions json.RawMessage            // string、object 或 array
    Criteria     map[string]json.RawMessage // 值为 string、object、array 或 null
}

type ScoreQuestion struct {
    Instructions json.RawMessage
    Criteria     []json.RawMessage // 有序等级，下标即分值
}

type BoolQuestion struct {
    Instructions json.RawMessage
    Criteria     *BoolCriteria
}

type BoolCriteria struct {
    True, False json.RawMessage
}

type ClassifierAnswer interface{ classifierAnswer() }

type ChoiceAnswer struct {
    Choice        string
    Probabilities map[string]float64
    Confidence    float64
}

type ScoreAnswer struct {
    Score         float64
    Confidence    float64
    Probabilities map[int]float64 // 可选
    Legend        map[int]string  // 可选
}

type BoolAnswer struct {
    Probability float64 // "是"的概率
}

type ClassifierResult struct {
    Answers       map[string]ClassifierAnswer
    ResponseModel Nullable[string]
    Usage         Usage
    StopReason    StopReason
    ErrorMessage  string
    Metadata      CallMetadata
}

type TypeSafeOptions struct{} // 首期为空
```

- `ClassifierQuestion` 的 JSON 解码按 `type` 字段分派：`choice`、`score`、`bool`。拒绝未知类型与字段组合错误。
- JSON 数字以 `RawMessage` 或 `json.Number` 保留精度。
- 与 pi 的差别（[types.ts:633-689](https://github.com/earendil-works/pi/blob/v1.0.0/packages/ai/src/types.ts#L633-L689)）：pi 的 `state` 只接受 `JsonObject`，`instructions` 只接受 `string`，choice criteria 是 `Record<string, string>`，score 答案只有 `score` 与 `confidence`。

### 10.3 线上请求

```json
{
  "model": "jev-1.13.0",
  "state": {"ticket": "订单三天未发货"},
  "questions": {
    "intent":  {"type": "choice", "instructions": "客户意图", "criteria": {"refund": null, "track": "查询物流"}},
    "anger":   {"type": "score",  "instructions": "情绪强度", "criteria": ["平静", "不满", "愤怒"]},
    "urgent":  {"type": "noul",   "instructions": "是否紧急", "criteria": {"true": "需立即处理", "false": "可排队"}}
  }
}
```

- `BoolQuestion` 编码为 `type: "noul"`，与 pi 的 `wireRequest` 一致（[system-one-shared.ts:148-158](https://github.com/earendil-works/pi/blob/v1.0.0/packages/ai/src/api/system-one-shared.ts#L148-L158)）。
- 问题键不进入模型推理，只用于对应答案（官方原文："The key is not sent to the underlying model"）。
- bool criteria 的线上形状以实施时的官方 API 页为准，fixture 记录来源。

### 10.4 线上响应与映射

```json
{
  "model": "jev-1.13.0",
  "answers": {
    "intent": {"type": "choice", "choice": "track", "probabilities": {"refund": 0.1, "track": 0.9}, "confidence": 0.82},
    "anger":  {"type": "score", "score": 1.05, "confidence": 0.7, "probabilities": {"0": 0.0, "1": 0.95, "2": 0.05}, "legend": {"0": "平静", "1": "不满", "2": "愤怒"}},
    "urgent": {"type": "noul", "noul": 0.95}
  },
  "usage": {"input_tokens": 296, "output_tokens": 20}
}
```

| 线上 | barness |
| --- | --- |
| `model` | `ResponseModel`（不覆盖授权的 `ModelID`） |
| choice `choice`/`probabilities`/`confidence` | `ChoiceAnswer` |
| score `score`/`confidence`/`probabilities`/`legend` | `ScoreAnswer`，分布与图例的键转为 int |
| noul `noul` | `BoolAnswer.Probability` |
| 响应头 `x-typesafe-request-id` | `Attempt.ProviderRequestID` |

### 10.5 校验

请求（构造时与回调后各一次，回调后以最终请求体为准）：

- 至少一个问题；问题键非空。
- instructions 是 JSON string、object 或 array 且非空。
- choice criteria 1–255 项（`MaxChoices`）；score criteria 2–10 项；bool criteria 必须有 true 与 false。
- `ClassifierPolicy` 的字节与数量限制。
- 回调不能改 `model`：改了即 `tenant_denied`；回调产出无法解码的请求即 `callback_failed`。

官方 [OpenAPI](https://api.typesafe.ai/openapi.json) 比文字说明宽松（score `minItems:1`、无 255/10 上限、instructions 可为 null），这些规则以文字说明为准，只属于本 adapter。

响应（先登记用量，再整体校验）：

1. 最终请求的每个问题键都有答案，且 `type` 与问题类型一致（bool 对 noul）；不允许多出未请求的键。
2. 所有概率与 confidence 有限且在 `[0,1]`。
3. choice：`choice` 属于最终 criteria 的键；`probabilities` 的键集与 criteria 相同；和在 `1 ± 1e-6` 内（OpenAPI 原文为 "sum to approximately 1"，容差在 live 证据后可调）；`choice` 是最高概率项之一。
4. score：`score` 在 `[0, 级数−1]`；提供了分布时键集为 `0..级数−1`、和在容差内，且 `score` 与 `Σ i·p_i` 在容差内一致；提供了图例时键集与分布相同。
5. 任何一条不满足：`newError(CodeProtocol, PhaseResponse, ...)`，`Answers` 为空，用量保留。不重新归一化。

pi 只检查数字有限与形状（[system-one-shared.ts:63-121](https://github.com/earendil-works/pi/blob/v1.0.0/packages/ai/src/api/system-one-shared.ts#L63-L121)），严格校验登记为扩展。

### 10.6 用量与计价

| 线上 | Usage |
| --- | --- |
| `input_tokens` | `Input` |
| `output_tokens` | `Output` |
| 二者之和 | `TotalTokens` |

两个计数都存在为 complete，缺一个为 partial，没有 `usage` 为 unreported。pi 同样先登记用量再解析答案（[system-one-shared.ts:227-230](https://github.com/earendil-works/pi/blob/v1.0.0/packages/ai/src/api/system-one-shared.ts#L227-L230)）。

目录 `Cost`：Input 为官方价，Output 为 0。[models](https://docs.typesafe.ai/models) 页写明 `jev-1.13.0` "Charged per input token. Output tokens are free"，价格为每百万输入 token $0.042。

### 10.7 限制与运维事实

| 项 | 内容 | 处理 |
| --- | --- | --- |
| 上下文 | 每请求 64k token；state 加最长问题合计 32k token | `ContextWindow = 64000`；barness 不在本地计 token，按字节限制兜底，超限由厂商返回 422 |
| 速率 | 100K token/s、80 请求/s，官方注明会动态调整 | 不在库内限速；429 由 `Binding.Retry` 决定 |
| 错误码 | HTTP 参考列出 401、422、429、529；403 只见于官方 Python SDK 的异常列表 | 按 `codeForStatus`：401/403 upstream_auth、422 invalid_request、429 rate_limited、529 upstream_error |
| retry-after | 429/529 可能带 | 由第 5.1 节的规则读取 |
| 语言 | 英文效果最好，CJK 需自行测试；不做客户微调 | 业务效果另用标注集评估 |
| 别名 | `jev-latest`、`jev-preview` 当前都指向 `jev-1.13.0`，会随版本移动 | 首期只列 `jev-1.13.0` |

### 10.8 失败场景

三种问题混合；bool 与 noul 往返；有序 score；缺答、多答、`type` 不匹配；choice 不在 criteria 中；分布键集不符；概率越界、NaN；分布和偏离；score 超出等级；score 与分布不一致；401、403、422、429、529；答案错误但用量保留；回调改问题后按最终集合校验；回调插入超限问题被拒；回调改 model 被拒；`ResponseModel` 与授权 ID 不同时都保留。

## 11. E：OpenAI Images

### 11.1 绑定、端点与为什么直连

`ProviderOpenAI` × `APIOpenAIImages` × image；endpoint `https://api.openai.com/v1`；无参考图 `POST /images/generations`，有参考图 `POST /images/edits`；`Authorization: Bearer <key>`。

直连 HTTP 适用 spec §5 的例外条款。已安装的 openai-go v3.66.0 中，`ImageEditParams` 的图片与 mask 是 `io.Reader`（[image.go](https://github.com/openai/openai-go/blob/v3.66.0/image.go) 第 967–1065、1087 行），`MarshalMultipart` 先写进 `bytes.NewBuffer(nil)`（第 1066–1081 行），`NewRequestConfig` 在应用请求选项之前调用它（[requestconfig.go](https://github.com/openai/openai-go/blob/v3.66.0/internal/requestconfig/requestconfig.go) 第 195–221 行）。v3.71.1 新增了不缓冲文件内容的 `MarshalMultipartTo`，但 `ImageEditParams` 仍是 `io.Reader`，默认请求路径仍先调用 `MarshalMultipart()`，也没有 JSON 编辑路径。SDK 无法在读入前限制请求大小，所以仅此路线直连；聊天 SDK 路线不变。先例是 [ADR-0012](../../docs/adr/0012-barness-ai-gemini-developer-api-adapter.md)。

### 11.2 请求

生成：

```json
{
  "model": "gpt-image-2.5-sunburst-2026-09-08",
  "prompt": "一张俯拍的餐厅露台照片",
  "n": 1,
  "size": "1024x1024",
  "quality": "medium",
  "output_format": "png"
}
```

编辑：

```json
{
  "model": "gpt-image-2.5-sunburst-2026-09-08",
  "prompt": "将背景改为餐厅露台",
  "images": [{"image_url": "data:image/png;base64,<内联图片>"}],
  "mask": {"image_url": "data:image/png;base64,<内联 mask>"},
  "output_format": "png"
}
```

**JSON 编辑是 live gate。** [edit 参考](https://developers.openai.com/api/reference/resources/images/methods/edit) 的 "Body Parameters JSON" 列出 `images: array of {file_id, image_url}` 与 `mask: {file_id, image_url}`，"Provide exactly one of image_url or file_id"。但该页与 [图像指南](https://developers.openai.com/api/docs/guides/image-generation) 的所有 `/v1/images/edits` 示例都是 multipart（`-F image[]=@…`）；指南里的 `image_url` JSON 片段属于 Responses API 的 `image_generation` 工具。所以 P08 的第一个 live 场景就是 JSON 编辑。未通过前编辑能力不进支持目录，运行时不切换到 multipart。

只生成内联 data URL。`file_id`、`http(s)` URL 在构造时与回调后都拒绝。

### 11.3 选项与校验

```go
type OpenAIImagesOptions struct {
    N                 Nullable[int]
    Size              Nullable[string] // "auto" 或 WIDTHxHEIGHT
    Quality           Nullable[string] // low、medium、high；2.5 型号另有 xhigh、max；auto
    Background        Nullable[string] // transparent、opaque、auto
    OutputFormat      Nullable[string] // png、jpeg、webp
    OutputCompression Nullable[int]    // 0–100，仅 jpeg/webp
    Mask              *Image
    InputFidelity     Nullable[string] // high、low
    Moderation        Nullable[string] // low、auto
}
```

| 规则 | 来源 |
| --- | --- |
| `prompt` 1–32000 字符 | edit/generate 参考 |
| `n` 1–10，且不超过 `MaxOutputImages` | 同上 |
| 参考图最多 16 张（GPT image 型号） | edit 参考 |
| 每个 `image_url` 不超过 20971520 字符（约 15 MiB 解码字节） | edit 参考 `maxLength 20971520` |
| `Mask` 要求至少一张参考图；GPT image 的 mask 是提示性的，作用于第一张参考图 | 指南 |
| `Background: transparent` 要求 `OutputFormat` 为 png 或 webp | edit 参考 |
| `OutputCompression` 只在 jpeg/webp 时允许 | 同上 |
| `Size` 的任意 `WIDTHxHEIGHT` 只对 gpt-image-2 与 2.5 型号开放：边长为 16 的倍数，宽高比在 1:3 到 3:1，最大 3840×2160 | 同上 |
| `InputFidelity`：gpt-image-1/1.5 可为 high/low；gpt-image-1-mini 只能 low；gpt-image-2 必须省略；2.5 型号以 live 证据为准 | 同上 |
| 不暴露 `stream`、`partial_images`、`response_format`、`user`、`style` | `response_format`、`style` 已标为 retired 型号专用 |

型号相关的规则写进 `ImageCapabilities`，adapter 只读能力，不硬编码型号名。

### 11.4 响应映射

响应字段只有 `created, background, data, output_format, quality, size, usage`（[ImagesResponse](https://github.com/openai/openai-go/blob/v3.66.0/image.go) 第 748–775 行与参考页一致）。

- 遍历全部 `data[]`，每条 `b64_json` 映射为一个图片块，顺序保留。
- MIME 由响应的 `output_format`（缺失时用最终请求的 `output_format`，再缺失为 png）决定，并用解码后的文件头校验。
- base64 严格校验；单张与总量对 `ImagePolicy`。
- 不设置 `ResponseID`、`ResponseModel`：响应没有 `id` 与 `model`。
- `x-request-id` 记为 `ProviderRequestID`（`requestIDHeaderOf`）。
- 不保留 `revised_prompt`：参考页写明 "Not returned by GPT image models"。
- `data` 为空或全部无效：`protocol`/`response`。

### 11.5 回调

回调后重新解码最终请求体并检查：`model` 未变；只含第 11.3 节允许的字段；`images[]`、`mask` 只有内联 data URL；数量与大小限额；生成请求不含 `images`，编辑请求含 `images`（端点在回调前已选定，回调不能改变它）。

### 11.6 用量

| 线上 | Usage |
| --- | --- |
| `input_tokens` | `Input` |
| `output_tokens` | `Output` |
| `total_tokens` | `TotalTokens` |
| `input_tokens_details.text_tokens`/`image_tokens` | `Modalities.InputText`/`InputImage` |
| `output_tokens_details.text_tokens`/`image_tokens` | `Modalities.OutputText`/`OutputImage` |

`usage` 缺失为 unreported；有 `input_tokens`、`output_tokens`、`total_tokens` 与输入明细为 complete；缺输出明细为 partial。参考页仍写 usage "For gpt-image-1 only"，而指南对 GPT Image 2.5 要求读取 usage，所以按存在与否处理。

### 11.7 首批型号

`gpt-image-2.5-sunburst-2026-09-08`。它是 edit 参考默认型号 `gpt-image-2.5-sunburst` 的日期快照，固定快照 ID 以便证据可复现。费率以 [图像指南](https://developers.openai.com/api/docs/guides/image-generation) 的 "GPT Image 2.5 costs" 段为准，录入时填 `Source`。live gate 通过后才列入；其他型号按 ADR-0018 逐个评估。

### 11.8 失败场景

生成与编辑的路径和认证头；JSON 参考图与 mask；`image_url` 超长；`file_id`、远程 URL 被拒；mask 无参考图被拒；透明背景配 jpeg 被拒；不支持的 `InputFidelity` 被拒；全部 `data` 条目映射；MIME 与文件头不符；坏 base64；`data` 为空；400 内容拒绝；usage 缺失与缺输出明细；回调加 `stream` 或 `file_id` 被拒。

## 12. F：Google Gemini Interactions

### 12.1 为什么选 Interactions

| 方面 | Interactions | 扩展现有 generateContent adapter |
| --- | --- | --- |
| 官方定位 | "recommended for all new projects"；"all new models, multimodal capabilities, tools, and agentic features will launch on the Interactions API"（[overview](https://ai.google.dev/gemini-api/docs/interactions-overview)） | "now considered legacy"，"remains fully supported"；[图像页](https://ai.google.dev/gemini-api/docs/generate-content/image-generation) 仍以 `responseModalities:["TEXT","IMAGE"]` 文档化同批型号 |
| 图像配置 | 类型化的 `response_format`：宽高比、尺寸、MIME、交付方式 | `imageConfig` |
| 用量 | 显式模态分项，thought token 单独列出 | 现有 usageMetadata |
| 现有资产 | 无 | REST adapter、错误格式、ADR-0012 fixture |
| 代价 | 默认服务端存储；需拒绝的字段多；2026-05/06 刚有破坏性变更（[changelog](https://ai.google.dev/gemini-api/docs/changelog) 2026-05-06） | 新型号可能不再上线 |

决定选 Interactions：图像型号更替很快，底稿完成次日首批型号即被弃用，新型号只承诺在 Interactions 上线。现有 generateContent 聊天 adapter 不迁移；ADR-0012 决策八把图像输出型号排除出聊天目录的决定继续有效。

### 12.2 版本固定 v1beta

[v1 参考](https://ai.google.dev/api/interactions-api-v1) 的型号枚举使用 `models/` 前缀，且不含 `gemini-3-pro-image`、`gemini-3.1-flash-lite-image`、`gemini-nano-banana-2.1`；[v1beta 参考](https://ai.google.dev/api/interactions-api) 使用裸 ID，主图像指南也用 v1beta。不做 v1/v1beta 自动回退；将来切换 v1 是一次单独的协议配置与 fixture 变更。

### 12.3 请求

```http
POST https://generativelanguage.googleapis.com/v1beta/interactions
x-goog-api-key: <key>
Content-Type: application/json
```

```json
{
  "model": "gemini-nano-banana-2.1",
  "store": false,
  "input": [
    {"type": "text", "text": "将产品放在餐厅露台上"},
    {"type": "image", "mime_type": "image/png", "data": "<base64>"}
  ],
  "response_format": {"type": "image", "aspect_ratio": "16:9", "image_size": "2K"}
}
```

- `input` 顺序：先 prompt 文本，再参考图，与 `ImagesRequest` 的收敛一致。
- 只要求图像时 `response_format` 为单个 image 项；官方说明默认同时返回文本与图像，指定图像格式后只返回图像（[图像指南](https://ai.google.dev/gemini-api/docs/image-generation?hl=en)）。首期只请求图像。

```go
type GoogleImagesOptions struct {
    AspectRatio Nullable[string] // 1:1、2:3、3:2、3:4、4:3、4:5、5:4、9:16、16:9、21:9、1:8、8:1、1:4、4:1
    ImageSize   Nullable[string] // 512、1K、2K、4K，按型号能力
}
```

不模拟 OpenAI 的 mask。

### 12.4 固定项与拒绝项

adapter 构造时写入固定项；回调后重新解码最终请求体，固定项被改或出现拒绝项即 `tenant_denied`（扩大授权）或 `callback_failed`（无法解码）。

| 类别 | 字段 |
| --- | --- |
| 固定 | `store: false`；`response_format.type: "image"`；省略 `response_format.delivery`（ADR-0023）；同步执行 |
| 拒绝 | `previous_interaction_id`、`background`、`agent`、`tools`、`environment`、`webhook_config`、`continuation_token`、`service_tier`、`stream`；`input[].uri`；任何 `delivery` 字段（含 inline/uri/null） |

官方说明 `store=false` "is incompatible with background execution and prevents using previous_interaction_id"（[overview](https://ai.google.dev/gemini-api/docs/interactions-overview)）；默认存储时付费层保留 55 天。

### 12.5 硬限制

| 项 | 限制 | 来源 |
| --- | --- | --- |
| inline 请求总大小 | 20 MB，含文本与全部内联字节 | [image-understanding](https://ai.google.dev/gemini-api/docs/image-understanding) |
| 参考图 | 最多 14 张；`gemini-nano-banana-2.1` 为最多 10 个物体加 4 个角色 | [图像指南](https://ai.google.dev/gemini-api/docs/image-generation?hl=en) |
| `image_size` | `gemini-nano-banana-2.1` 支持 1K、2K、4K，不支持 512 | 同上 |
| 输出数量 | "The model won't always follow the exact number of image outputs" | 同上；`MaxOutputImages` 不承诺精确 N |
| 输入 MIME | png、jpeg、webp、heic、heif、gif、bmp、tiff | v1beta 参考 `ImageContent.mime_type` |

20 MB 是 `MaxInputImages × MaxImageBytes` 与 prompt 合计的上界，在构造时与回调后按最终请求体字节数检查。

### 12.6 响应映射

| `status` | 处理 |
| --- | --- |
| `completed` | 解析输出 |
| `failed`、`cancelled` | `upstream_error`/`response`，诊断信息写入 `ErrorMessage` |
| `in_progress`、`requires_action`，或带 `continuation_token` | `protocol`/`response`；同步路线不续读 |

- 遍历全部 `steps` 中 `type == "model_output"` 的 `content`，text 与 image 按出现顺序映射为输出块。不使用只返回最后一张图的 `output_image` 便利属性，官方说明交错输出时它"will not capture all parts"。
- 图片块取 `data` 与 `mime_type`；若出现 `uri` 而没有 `data`，按 `protocol` 失败（请求已固定 inline，出现即厂商异常）。
- `id` 写入 `ResponseID`，仅供诊断，不转为可续接状态。
- `model` 若出现，写入 `ResponseModel`。
- 没有任何有效图片：`protocol`/`response`；被安全策略拦截且有诊断时为 `upstream_error`。

### 12.7 用量

| 线上 | Usage |
| --- | --- |
| `total_input_tokens` | `Input` |
| `total_output_tokens + total_thought_tokens` | `Output` |
| `total_thought_tokens` | `Reasoning` |
| `total_tokens` | `TotalTokens` |
| `total_cached_tokens` | `CacheRead` |
| `input_tokens_by_modality[]` 的 text、image | `Modalities.InputText`、`InputImage` |
| `output_tokens_by_modality[]` 的 text、image | `Modalities.OutputText`、`OutputImage` |

官方示例为 input=7、output=20、thought=22、total=49（v1beta 参考），说明 thought 不含在 `total_output_tokens` 里。加进 `Output` 后保持 `Usage.Reasoning` "已包含在 Output 内"的现有定义，并满足 7 + 42 = 49。四个 total 与两组模态分项齐全为 complete，缺任意一项为 partial。

thought token 的计价：pricing 页把 "text and thinking" 列在同一个输出费率下，所以 thought 按 `OutputText` 费率计入 `Cost.Output`。实施时若 `output_tokens_by_modality` 已含 thought，则不重复计算；以 live 证据确认后写入 fixture。

### 12.8 首批型号

`gemini-nano-banana-2.1`（2026-10-06 GA）。不列入：`gemini-3.1-flash-image` 于 2026-10-06 弃用，官方原文 "Migrate to gemini-nano-banana-2.1"；`gemini-2.5-flash-image` 已弃用，[pricing](https://ai.google.dev/gemini-api/docs/pricing?hl=en) 写 2026-10-02 关闭而 [deprecations](https://ai.google.dev/gemini-api/docs/deprecations) 表写 2027-03-15，官方自相矛盾；Imagen 4 已关闭，且是独立协议。

费率以 pricing 页为准。示例：`gemini-nano-banana-2.1` 图像输出为每百万 token $30；分辨率与 token 的对应为 1K 1120、2K 1680、4K 3780。`gemini-3.1-flash-image` 的 4K 为 2520 token，两者不能共用分辨率表。

### 12.9 失败场景

v1beta 路径与 `x-goog-api-key`；请求中 `store:false` 存在、`delivery` 省略且回调不能加入；每个拒绝字段各一个用例；`input[].uri` 被拒；最终请求超过 20 MB 被拒；参考图超过型号上限被拒；不支持的 `image_size` 被拒；多个 `model_output` 步骤的交错文本与图片顺序；只有文本没有图片；五种 status 与 `continuation_token`；图片块只有 `uri`；usage 等式与 partial；回调改 `model` 被拒。

## 13. 回调的共同边界

- 两条图像路线与分类路线都沿用 `TransformHeaders`、`OnPayload`、`OnResponse`；`Payload` 与 `ResponseInfo` 增加 `Operation` 字段。
- `OnPayload` 之后，以最终请求体为准重新执行第 10.5、11.5、12.4 节的检查；输出 MIME、输出配置与计价都以最终请求为准，不再读回调前的 Options。
- 回调可以在允许范围内修改字段，例如 `output_format`；不能改变端点、操作、授权型号，不能引入外部引用或续接状态。
- 无效回调结果为 `callback_failed`；扩大授权为 `tenant_denied`。
- unary 路线在取得 2xx 响应后执行一次 `OnResponse`，再读取响应体。Gemini 聊天路线仍不调用 `OnResponse`（[ADR-0012](../../docs/adr/0012-barness-ai-gemini-developer-api-adapter.md) 决策四）。

## 14. pi 0.87.1 → 1.0.0 增量的处置汇总

窗口内的变化（[CHANGELOG @ v1.0.0](https://github.com/earendil-works/pi/blob/v1.0.0/packages/ai/CHANGELOG.md) 第 3–80 行）：

| 增量（版本） | 处置 |
| --- | --- |
| 统一 image/classifier 类型、`generateImages`、`classify`、按类型查找（0.99.0） | 本设计第 6–7 节 |
| 模型数据 schema 3 → 6；动态 overlay 接受全部类型（0.99.0） | 借鉴类型键，不移植 store |
| TypeSafe Jev（0.99.0） | 第 10 节 |
| Jev 经 OpenRouter、Cloudflare、Vercel、OpenCode（0.99.0） | 不纳入 |
| `llama-cpp-classify`（0.99.0） | 不纳入，保留 classifierAdapter 接缝 |
| `onProviderStreamEvent`（0.99.0） | 后续独立增量；若做，必须是可信同步 Hook |
| `AssistantMessage.thinkingLevel`（0.99.0）；`NestedToolCalls`（窗口内 types.ts 新增，提交 `8562bcf66`，不在 CHANGELOG） | 宿主与 Agent 范围 |
| 型号默认 samplingParams（0.99.0）、fast 计价（0.99.0）、1h 增量（0.99.0）、未完成工具调用（0.99.0）、Retry-After（0.99.2） | 第 5 节 |
| ChatGPT 登录（0.99.0）、Anthropic 身份联合（0.99.2）、copy-code 登录（1.0.0）、订阅超限不重试（0.99.0） | 不纳入 |
| `pi-ai/models` 轻入口（0.99.2） | 不移植 |
| Anthropic `strict:"prefer"` 关键字降级（0.99.2）、grammar 工具 `ctc_` 回放（1.0.0） | barness 没有 strict 与 grammar 工具，无作用对象 |
| Mistral、Z.AI、OpenCode、Copilot 专属修复；新型号（Claude Sonnet 5.5、GPT-6.1 Sol 等） | Provider 不纳入；型号随 G0 按 ADR-0018 评估 |
| `models.all.json` 数组目录（同 ID 每类型一次） | 与第 6.1 节"同一 ID 可在不同操作下各出现一次"一致 |

以下能力在 0.87.1 或更早已存在，不属于本窗口：图像生成与 `ImagesModels`（0.74.1、0.80.0，0.87.1 oracle 已含 openrouter-images）、grammar/custom tools（0.80.4）、strict prefer/require（0.82.0）、tool search（0.84.2）、托管推理强度与 `providerThinkingLevel`（0.85.0，barness 已实现，ADR-0019）、Anthropic 中途工具变更（0.86.0）、`allowedFallbackModels`。

## 15. 文档与 ADR 修订清单

| 文件 | 修订 | 工作包 |
| --- | --- | --- |
| ADR-0004 | 基线 1.0.0、新 provenance | G0 |
| ADR-0006 第 24 行 | 无法解析的 Retry-After 改为退避 | G |
| ADR-0010 决策五 | 增加 fast；用量部分增加 Modalities 与图像计价 | G、C |
| ADR-0011 决策三 | 增量中的 1h 明细覆盖更新 | G |
| ADR-0001 | 单次生成轮次扩展为单次模型操作 | A |
| 新 ADR：多类型模型操作与 Binding 授权 | 操作维度、四维身份、零值例外、分派、PhaseResponse | A、B |
| ADR-0002/0007/0008 | unary reader、子策略深复制、许可持有到解析结束 | C |
| ADR-0005/0009 | 新路线拒绝字段、Operation 元数据、unary OnResponse 时机 | B |
| 新 ADR：原生图像与 TypeSafe unary 协议 | 直连例外、Interactions 取舍、v1beta、单一用量轴与模态计价 | D、E、F |
| ADR-0012 | 新图像路线与聊天路线并存 | F |
| ADR-0014 | `routes` 增加两条图像路线 | T |
| ADR-0016/0017 | 新 combo、operation 字段、live 预算、门禁输入 | T |
| ADR-0018 | 非聊天型号的包含准则与豁免 | T |
| [contract.md](../../docs/barness-ai/contract.md) | 新入口、Phase 列表、零值规则、子策略 | B、C |
| [differences.md](../../docs/barness-ai/differences.md) | 基线版本、TypeSafe 扩展、图像扩展路线 | G0、D、E、F |
| [README](../../docs/barness-ai/README.md) | 支持矩阵 | H |
| [GLOSSARY](../../GLOSSARY.md) | 模型操作、图像型号、分类型号、unary 调用；修订 barness-ai 与 Support Matrix | A |

## 16. 开放事项与 live 确认点

| 事项 | 当前假设 | 确认方式 |
| --- | --- | --- |
| OpenAI `/images/edits` 是否接受 JSON 请求体 | 接受，参考页有文档 | P08 首个 live 场景；不接受则编辑能力不列入 |
| `gpt-image-2.5-sunburst` 对 `InputFidelity` 的要求 | 未文档化 | P08 live；结果写入能力 |
| Google 首批型号的 usage 是否把 thought 计入模态分项 | 不计入 | P09 live；写入 fixture |
| TypeSafe bool criteria 的线上形状 | `{"true": ..., "false": ...}` | 实施时读官方 API 页，P07 live |
| choice/score 分布和的容差 | `1e-6` | P07 live 后按观测调整 |
| Google `gemini-2.5-flash-image` 关闭日期矛盾 | 不列入，不影响实施 | 无 |
| 厂商价格 | 本文数字只作示例 | 实施时读官方页面，填 `Source` |
