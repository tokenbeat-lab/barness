# MiniMax 国际站 Host 示例

[English](README.md)。2026-10-18 当前状态：三协议和六个候选身份有离线机制验收，
**M3 正式支持仍待工单10**。M3.1 没有已确认的 M Plan 资格或有来源的正式价格，不能继承
M3 的支持结论。本示例不发布支持。[候选目录](../../internal/testkit/minimax/candidate.json)
使用明确标记的合成测试费率，不能展示为真实估价。正式推广必须先有三个 M3 身份各自的
真实验收、有来源的 USD 价格、适用源码/目录 Hash，再通过矩阵、增量发布合同和最终门禁。
[实际验收工单10](../../../.scratch/barness-ai-minimax/issues/10-formal-catalog-support.md)
负责收敛，M3.1 始终单列候选。

## 先授权用途，再装配

`minimaxhost.Bindings` 接受可信 Host 配置，不接受客户端自报身份，也不探测账户权益。
一次生成三份可独立撤销的 Binding。本示例的 standard 只授权 M3；明确有资格的 M Plan
候选只授权 M3.1。其他实际获准型号集合需要 Host 自己明确审阅和授权。

```go
bindings, err := minimaxhost.Bindings(minimaxhost.Authorization{
    ProviderID: ai.ProviderMiniMax, // explicit international service; China: ai.ProviderMiniMaxCN
    TenantID: "认证得到的租户", AccountScopeID: "真实厂商账户作用域",
    CredentialRef: "minimax-standard-vault-ref", Version: "config-17",
    Purpose: minimaxhost.Standard,
})
// 检查 err，再把三份绑定分别保存到租户作用域的 Resolver。
```

| 后缀 | API | 基址 | 最终资源 | 认证 |
| --- | --- | --- | --- | --- |
| chat | openai-completions | https://api.minimax.io/v1 | /v1/chat/completions | Bearer |
| responses | openai-responses | https://api.minimax.io/v1 | /v1/responses | Bearer |
| messages | anthropic-messages | https://api.minimax.io/anthropic | /anthropic/v1/messages | x-api-key |

BindingID 含 `minimax-standard-` 或 `minimax-m-plan-`。两种用途使用独立凭据引用和授权；
同租户、同真实账户、同用途的三协议可以在明确许可下共享一个 CredentialRef。AccountScopeID
表示真实厂商账户，不是 Key 版本，也不是客户端声明。不会从 Key 格式猜权益、使用 Coding
专用基址、切区/换 Key/换型号/换协议或把套餐用途转成按量余额。本地交互套餐不自动代表
云端生产授权。Messages 的 SDK `?beta=true`、版本头只是传输事实，不开启 Claude beta
header/body 特性，不需要 wire shim；禁止通过回调追加另一套 Bearer/x-api-key 凭据。

## 本地装配

复用 [localassembly](../localassembly/localassembly.go)：由本地程序选择唯一 KeySource，
并显式传入审阅过的 Catalog。库及 SDK 不读厂商默认环境 Key、dotenv、profile 或默认端点。
另一个协议也显式选另一份 Binding，重复同样装配；每个 Client 自有凭据快照。

```go
// reviewedCatalog 由可信 Host 配置加载和验证；未转正前必须明确标记候选用途。
binding := bindings[0] // 本示例首项是 chat；实际按 API 显式选择。
local, err := localassembly.Open(localassembly.Config{
    TenantID: binding.TenantID, Binding: binding, Catalog: &reviewedCatalog,
    Key: localassembly.KeySource{EnvVar: "HOST_MINIMAX_STANDARD_KEY"},
    Policy: localassembly.LocalPolicy(),
})
if err != nil { return err }
result, err := local.Client.Complete(ctx, local.Scope(), local.Target("MiniMax-M3"),
    ai.Request{Messages: []ai.Message{ai.UserText("用一句话回答。")}},
    ai.ChatOptions{MaxTokens: ai.Value(512)})
```

M Plan 候选必须先确认权益，再用 `Purpose: minimaxhost.MPlan`、独立账户/凭据引用、审阅过
的候选目录及明确来源如 `HOST_MINIMAX_MPLAN_KEY`。核心没有识别套餐或凭据用途的枚举，
用途授权属于 Host。`Catalog:nil` 仍使用正式内置目录，不会让未转正 MiniMax 可调用。
`AllowLoopbackHTTP` 和受控 transport 仅用于离线测试。local Open 会创建不可变、已启用的
装配并分配 local 凭据/版本；撤销本地路线时先停止接受新轮次并退役该装配。需要动态
Enabled/版本撤销时使用下面的版本化 Resolver，不能把 Enabled=false 传给 Open 当作禁用。

## 云端装配

共享一个 Client，Binding/Credential Resolver 均按可信租户查询。凭据返回值的
BindingVersion 要对应本次固定的绑定版本；检查归属、真实账户、Active 和版本，返回独立
值，不使用全局 Key。Identity 由认证结果产生；TenantID 不能取自请求 JSON。

```go
client, err := ai.NewClient(ai.Config{
    Policy: hostintegration.CloudInteractivePolicy(),
    Bindings: trustedBindings, Credentials: trustedCredentials,
    Catalog: &reviewedCatalog, Observer: metadataObserver,
})
// 检查 err。可信部署可统一收紧输出预算，不触碰已授权型号和思考默认。
bounded := client.WithHooks(ai.Hooks{OnPayload: func(_ context.Context,
    _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
    if p.ProviderID == ai.ProviderMiniMax {
        field := map[ai.API]string{
            ai.APIOpenAICompletions: "max_completion_tokens",
            ai.APIOpenAIResponses: "max_output_tokens",
            ai.APIAnthropicMessages: "max_tokens",
        }[p.API]
        if field != "" { p.Body[field] = 512 }
    }
    return ai.KeepPayload(), nil
}})
cloud := hostintegration.New(bounded, tenantSessionStore, requestid.New)
who := hostintegration.Principal{TenantID: authenticatedTenant, ActorID: authenticatedActor}
session, err := cloud.OpenSession(ctx, who)
// 检查 err；HTTP handler 还应检查自己的应用权限、请求 DTO 和下游连接。
result, err := cloud.Turn(ctx, who, hostintegration.TurnRequest{
    SessionID: session, BindingID: bindings[0].BindingID,
    ModelID: "MiniMax-M3", Prompt: "简短回答。",
}, downstreamSink)
```

`Client` 和 `HookedClient` 都满足 Host 示例实际消费的 ModelClient interface，继续支持
已有聊天/图像/分类入口。示例 MemoryStore 不代替生产存储；持久化存储必须先保证记录完整性
及租户/session 归属。Host 会在调用库前拒绝伪造租户、跨租户 session，传递取消，并向下游
发送有归属的 EventEnvelope。UI 应处理加载、流式局部、空状态、length/error、重试反馈和
可访问性。断开连接要取消自己的调用，不能影响其他租户。

## 六身份的思考语义

| 型号 × 协议 | 省略 | 显式支持 |
| --- | --- | --- |
| M3 Chat | 开启 | 二元 off/high，high 只表示打开 |
| M3 Responses | 关闭 | high 打开，原生 none 关闭 |
| M3 Messages | 关闭 | ThinkingEnabled adaptive/disabled；Effort 拒绝 |
| M3.1 Chat | 开启、默认 max | low/medium/high/xhigh/max，不可关闭 |
| M3.1 Responses | 开启、默认 max | 五档深度，不可关闭 |
| M3.1 Messages | 开启、默认 max | 五档 Effort，包括 full 只给 Effort；不可关闭 |

DefaultThinkingLevel 只是省略时选采样默认的标签，不会物化成 wire effort 或用户显式档位。
ThinkingSwitchOnly 与 RequiresThinking 独立：把 M3.1 收窄为 high-only 仍保留 effort；
把 M3 收紧为必须思考仍不获得深度控制。型号声明只能收紧已实现的厂商形状，不承诺实测
推理强度，也不自动迁移协议。三协议各自声明型号，不从一个条目运行时派生另外两个。

full 传对应协议的 ChatOptions / ResponsesOptions / AnthropicOptions，错误协议选项拒绝。
示例显式 MaxTokens=512。M3 开启用 Chat/Responses ReasoningEffort=ThinkingHigh，或
Messages ThinkingEnabled=true；关闭用 Chat SamplingParams `thinking:{type:disabled}`、
Responses SamplingParams `reasoning:{effort:none}`，或 Messages ThinkingEnabled=false。
M3.1 拒绝 off/none、full minimal；Messages 只给 EffortLow 也不会漏掉 effort。
SimpleOptions.Reasoning 留空保持服务默认；M3 minimal/low 向上夹到 high；M3.1 minimal
夹到 low。没有新增 simple off 枚举，Messages 不使用 Claude budget 算法。

Complete/Stream 使用 full options，CompleteSimple/StreamSimple 使用 SimpleOptions；
WithHooks 四入口均可用。每个 Logical Call 都要 requestid.New，包括工具续接和手动重试。
Stream 消费 Next/Event/Envelope 后取 Result，并始终 Close；Complete 消费同一流式实现。
Host 工具循环可以复用 toolloop.Loop：只执行已结束且通过 ValidateToolCall 的调用，完整
追加 assistant 思考/正文/调用及配对结果，再开始新轮次。length/error/aborted 的局部工具
不能执行。循环必须有有限轮数，不把工具执行塞进协议模块。

## 请求子集、回调和费用

首期是文本/工具子集，不表示厂商本身没有多模态能力。未声明图片按既有占位/历史转换处理，
回调不能插入多模态容器。工具只有默认/auto/none；required/any/具名强制、禁并行、strict、
托管工具及被厂商忽略的停止/资源控制均拒绝。服务层级只支持省略/standard，priority 拒绝。
Messages 不接入显式 budget/display/interleaved/beta/block_binding/eager/browser、top_k、
stop_sequences；adaptive 期间仍保留显式 Temperature。Chat/Messages temperature [0,2]、
top_p [0,1]；Responses 两者先用 (0,1]，零值文档冲突尚未收敛。既有具名 null/省略规则保持，
不能把未知 null 当数字。最终预算是精确整数 1..型号上限，不抬到 16；采样合并后执行 payload
回调，回调后不重选分档。Messages 忽略所有 SamplingParams 和分档 map。

Hook 顺序是 headers→payload→response，每个逻辑调用一次，不随初始重试重复。不能改变
授权身份、认证、目标、型号或原生资源引用，最终 body 再校验。非法具名输入在读凭据前拒绝，
不能指望后续 hook 修复。默认零重试；显式有限重试也不会重放已开始的流或自动切账户/用途。

Chat/Messages 使用自动缓存；Messages short/long 不发 cache_control、不承诺 TTL。
Responses 只有 SessionID 且 retention 非 none 时派生租户/账户/session 路由键；没有 session
或 none 就不发。short/long 不保证期限/命中；不发原始 session 或 OpenAI 亲和头。路由键不等于
厂商缓存存储隔离承诺，原生缓存状态不要进日志。

Usage 缺失/部分/中断不等于已上报零。Messages 总输入是 Input+CacheRead+CacheWrite；
Chat/Responses 从输入总量拆出缓存部分；Reasoning 属于 Output，不重复累计或计费。
CostRates 是 USD/百万 tokens，UsageCost 是 USD；CatalogVersion/Hash 标识固定价格快照。
M3 标准价格采用有日期来源的国际站原生美元快照；官方只写512k，512000是保守项目估算策略，
取得可靠整数定义后再修订。未知缓存写入
或 M3.1 价格不能填零或复制；当前候选合成费率不是真实按量/M Plan 报价。取消只停止本地
工作，不保证远端停算或免计费。

## 信任与诊断

完整消息与 NativeStateEnvelope 一起保存。进程内 Result 已有可信来源；JSON 不能恢复信任。
Host 先核验存储完整性和租户/session，再用可信 scope 调 TrustNativeState。同租户/账户/
Provider/API/型号才能原生续接，不匹配时按已有规则降级并记录元数据。Chat 保留自身推理
字段；Responses 使用有来源的无损候选输入投影；Messages 当前只允许可信签名 thinking。
未验证 opaque/redacted 原生回放明确失败，不借用 GLM 无签名例外；每个实际续接形状仍需
工单10的自身真实证据。

用 errors.Is/As 检查 Error.Code/Phase；仅使用实际报告的 HTTPStatus/ProviderRequestID。
Host RequestID/AttemptID 单独保留，缺厂商 ID 就保持缺失，不从 Host ID 或其他厂商头臆造。
Observer 只有元数据，可进共享运维日志；Result、prompt、工具参数、思考、原生状态、原始
错误文本和秘密不能进共享日志。不要把无界 request/actor/job ID 当指标维度。

## 撤销、软件回滚和验收

只把选定 Binding.Enabled=false 并原子发布新版本：新调用零凭据读取/Attempt/HTTP，其他
租户、用途和路线继续可用。已发送调用固定原版本及已知消耗；恢复授权是新版本、新调用，
没有自动转投、双写或永久兼容 shim。

软件回滚独立恢复完整基线 `ecfd4b4d6ec54db30453f4b5112989e9fdbd0782`，tree
`b9798a4e96dfb08cbfcce45c182915589b42eed1` 的 7039 路径、blob 和模式，含运行时、目录、
价格、矩阵、差分、发布合同和依赖；随后验证旧路线可用、六个 MiniMax 身份不存在且零凭据/
请求。恢复发生在外部临时目录，结束后移除；保留用户根目录、调用 checkout 和已修改/未跟踪
文件。输出必须选择新路径：

```sh
python3 .scratch/barness-ai-minimax/evidence/ticket-08/rollback.py .evidence/minimax-rollback-review
go test ./ai/e2e -run '^TestMiniMax(Host|FullRollback)' -count=1
```

[工单08证据](../../../.scratch/barness-ai-minimax/evidence/ticket-08/README.md)含配置、完整
软件恢复及脱敏审计。它不续订历史真实支持，不绕过工单09/10最终发布门禁。

08证据保留历史中国区适用性；[11证据](../../../.scratch/barness-ai-minimax/evidence/ticket-11/README.md)记录当前国际站离线观察，不授予真实服务支持。


## Independent China service

`Authorization.ProviderID` is required: `ProviderMiniMax` keeps international
semantics; `ProviderMiniMaxCN` selects China with `api.minimax.cn`, independent
`minimax-cn-standard-*` BindingIDs and credential records. China standard M3 is a
new candidate; existing international qualification never authenticates China.
China M Plan is not granted by this example's international Preview entitlement.

`Endpoints` validates exact official service/protocol bases. Set
`TrustedGateway: true` only for explicitly reviewed nonofficial HTTPS/loopback
routes; Client loopback opt-in is separately required. No URL credentials, query,
fragment, official cross-market route or automatic endpoint fallback is accepted.

`minimaxhost.Credentials` validates its independent `CredentialRecord` (Provider,
Tenant, AccountScope, Purpose, allowed APIs, Active and both versions) against
trusted `Authorization` before invoking `ReadSecret(ctx, ref)`. Purpose/ownership
mismatch returns ErrAccessDenied without reading secrets; version conflict returns
ErrSnapshotConflict. Store record metadata independently of incoming Bindings.
The immutable example snapshot does not replace a Host backend's atomic updates.
Same service/account/purpose can share one tenant-owned record across its allowed
protocols; different services require separate records. Secrets are never inspected
to infer region or entitlement.

[Regional candidates and frozen prices](../../internal/testkit/regional/README.md)
retain nine new full identities, own source hashes, CNY conversion and incomplete
qualification. Software rollback first disables the new bindings and removes the
new catalog identities before an old process loads its catalog. No history rewriting
or sibling-service fallback. Ordinary proof: `go test ./ai/e2e -count=1 -run '^TestRegionalProvider'`.
