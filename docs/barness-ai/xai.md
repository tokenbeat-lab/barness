# xAI Host 装配与有界验证

首期身份固定为 `chat × xai × openai-responses × grok-4.7`，全球目标为 `https://api.x.ai/v1`。独立候选目录继续保留；按维护者确认的 [ADR-0054](../adr/0054-barness-ai-xai-responses.md)，本未发布集成分支已准备最终 `BuiltinCatalog` 接线，但正式资格、人工门槛与发布仍未完成。目录源码接线、候选或 controlled PASS 都不构成正式支持。型号、价格和未证事实见 [spec](../../.scratch/barness-ai-grok/spec.md)、[研究](../../.scratch/barness-ai-grok/research.md)与 [候选快照](../../ai/internal/testkit/xai/candidate.json)。真实资格、人工 race、压力和发布保持各自未完成状态。xAI 运行时和 live 源码的加入也改变现有地区/Kimi/GLM 原生来源库存；旧路线历史证据只保留原来源事实，当前来源适用性与所需自身 live 明确等待后续审查，不继承旧 PASS。

## 接入现有 Host

沿用 Host 已有认证、授权、BindingResolver 与 CredentialResolver。调用者给出可信 tenant/actor，Host 选择该租户拥有的 Binding；请求内容不能选择其他租户或读取环境中的 Key。库不创建用户体系、Secret Store 或工具执行器。

在 Host 的现有 Catalog 配置中显式加载已审核的候选 JSON，固定其版本与哈希。`ai/internal/testkit/xai.CandidateCatalog()` 仅供仓库内候选 harness 使用；外部 Host 通过已有配置加载相同 JSON，不导入 internal 包，也不修改全局内置目录。Host 可收窄上下文、输出、模态和价格能力，不能创造 wire 档位或改写必需推理。

以下装配使用现有公开类型；`hostBindings`、`hostCredentials` 是 Host 已有 resolver，`reviewedCandidateJSON` 是固定配置字节，`selectedKey` 来自明确指定的 Secret Store 账户。把下述两个快照放入已有 resolver 的存储；本例字段不构成新的认证模块。

```go
var cat ai.Catalog
if err := json.Unmarshal(reviewedCandidateJSON, &cat); err != nil {
    return err
}
hash, err := cat.Hash()
if err != nil || cat.Version != expectedVersion || hash != expectedHash {
    return errors.New("xAI candidate catalog snapshot differs")
}
binding := ai.Binding{
    TenantID: tenantID, BindingID: "grok-global", Version: "grok-binding-1",
    Enabled: true, Operation: ai.OperationChat,
    ProviderID: ai.ProviderXAI, API: ai.APIOpenAIResponses,
    Endpoint: "https://api.x.ai/v1", AuthKind: ai.AuthAPIKey,
    AccountScopeID: accountScopeID, CredentialRef: "grok-account-key",
    AllowedModels: []string{"grok-4.7"},
}
credential := ai.Credential{
    OwnerTenantID: tenantID, CredentialID: binding.CredentialRef,
    Version: "grok-key-1", BindingVersion: binding.Version,
    AccountScopeID: accountScopeID, Active: true,
    APIKey: ai.NewSecret(selectedKey),
}
// Store binding and credential through the Host's existing versioned resolver
// configuration before creating the Client; never log either credential.
policy := &ai.ResourcePolicy{
    MaxRequestBytes: 1 << 20, MaxImageBytes: 512 << 10,
    MaxFrameBytes: 2 << 20, MaxToolJSONBytes: 64 << 10,
    MaxErrorBodyBytes: 16 << 10, MaxOutputBytes: 16 << 20,
    MaxQueuedEvents: 1024, MaxQueuedEventBytes: 8 << 20,
    MaxQueuedObservations: 128,
    MaxConcurrentPerTenant: 2, MaxConcurrentProcess: 8,
    MaxAdmissionWaiters: 16, AdmissionWait: time.Second,
    CallTimeout: 3 * time.Minute, ConnectTimeout: 15 * time.Second,
    ResponseHeaderTimeout: 2 * time.Minute, ReadIdleTimeout: 90 * time.Second,
    Image: &ai.ImagePolicy{MaxInputImages: 2, MaxOutputImages: 1,
        MaxOutputImageBytes: 1 << 20, MaxTotalOutputImageBytes: 1 << 20},
}
client, err := ai.NewClient(ai.Config{
    Catalog: &cat, Bindings: hostBindings, Credentials: hostCredentials,
    Policy: policy, Observer: hostObserver,
})
if err != nil { return err }
scope := ai.CallScope{TenantID: tenantID, ActorID: actorID,
    RequestID: globallyUniqueRequestID, JobID: jobID}
target := ai.Target{BindingID: binding.BindingID, ModelID: "grok-4.7"}
ctx, cancel := context.WithTimeout(parentContext, 3*time.Minute)
defer cancel()
result, err := client.Complete(ctx, scope, target,
    ai.Request{Messages: []ai.Message{ai.UserText("Compute 17 * 23.")}},
    ai.ResponsesOptions{MaxTokens: ai.Value(1024)})
```

这些限额是短对话、最多两张小图与小型函数工具的装配示例，部署前按自身设计负载调整。请求和响应里的密文、工具参数、重复终态均计入字节限额；不能只按可见答案估算 frame。所有调用必须带取消范围，Stream/StreamSimple 消费完或显式 `Close`，失败路径也释放正文和 Admission。

Full 支持省略 effort 与 low/medium/high/xhigh；省略保留厂商默认。Simple 使用现有规范档位映射。`max_output_tokens` 仅限制可见输出，不是推理、工具参数或账单硬上限。客户端默认不自动重试；需要重试时由已有显式策略限制初始请求，不能在流开始后重放。

## 历史、图片与工具

每轮传递完整 Transcript。Host 持久化 assistant 内容、原始 Thinking.Signature、工具 ID 和 `message.NativeState.Envelope()`。JSON 解码不能恢复信任；Host 先检查存储所属 tenant 与完整性，再单独调用：

```go
trusted, err := ai.TrustNativeState(scope, storedEnvelope)
if err != nil { return err }
restoredMessage.NativeState = trusted
```

不要把客户端提交的 envelope 当成可信来源。同账户、同 Provider/API/型号才能原样续接；跨账户或型号会按既有合同降级。保留可见文本和工具配对，处理 `Metadata.NativeStateDowngrades`，不能把密文转给另一家服务。

Host 声明普通函数工具，收到完整成功的 `toolUse` 后逐个用 `ai.ValidateToolCall(result.Message, index, declaredTools)` 验证，再按 Host 自身权限执行。把原始 `ai.ToolCall.ID` 交给 `ai.ToolResultText` 或 `ai.ToolResultMessage.ToolCallID`，不要拆分、截断或自造 wire ID。下一轮顺序包含原 user、完整 assistant、每个 ToolResultMessage，再调用同一目标。并行工具逐 ID 配对，不按名称或返回顺序猜测。只支持 auto/required/none，不使用具名强制选择。

图片只传 `ai.Image{Data: base64Bytes, MimeType: "image/png"}` 或 JPEG，来自本次已授权内容。工具结果图放在该调用的 `ToolResultMessage.Content`，保留文本、PNG/JPEG 顺序。Host 不把公网 URL、file ID 或上传资源当作内嵌图片。

`SessionID` 是缓存提示，经已有 tenant/provider/account/session 派生后发送；原值不发给厂商。none 省略提示，short/long 不承诺 TTL 或缓存命中。`store:false` 保持无状态续接，不等同团队级 ZDR。

## 用量、成本与回滚

从 `Metadata.Attempts` 读取每次报告完整性与 `CostEstimate`。resolved 使用精确快照金额，bounded 的 Host 消费上界 Maximum；partial/unreported 必须保持未知，不能当免费。Reasoning 已包含在 Output，不能再次累加。目录按总输入含缓存达到 200000 时应用整请求长上下文价格；旧报告的价格快照不可覆盖。Observer 只记录元数据，默认不记录正文、密文、工具参数或 Key。

回滚时先禁用该租户 Binding 并更新版本，阻止新调用。已固定的调用按原快照结束，或由 Host 取消；轮换凭据时保持 AccountScopeID 代表真实账户。保留历史、envelope、旧目录与验证原件，不自动切到其他厂商，不添加内部兼容层。本功能无数据库迁移。

## 候选 live 的精确合同

唯一入口是编译后的 `-test.run=^TestXAICandidate$`，一个进程只运行本路线。配置复用现有通用转发字段：`BARNESS_AI_LIVE_COMBO=xai-responses`、`BARNESS_AI_LIVE=1`、唯一 `BARNESS_AI_LIVE_KEY_XAI_RESPONSES`、账户/地区别名、可选显式 `BARNESS_AI_LIVE_PROXY` 与 `BARNESS_AI_LIVE_PERMISSIONS_FILE`。账户/地区别名是 Host 声明，不是从 Key 推断的厂商权益证明。拒绝其他路线、凭据和 endpoint override。无 switch/key/grant 产出 NOT_RUN、零 vendor HTTP；错误身份/目录/来源/预算/别名/过期或重复授权失败。

候选已取得自身完整真实验证，事实与原件由 [唯一当前状态](../../.scratch/barness-ai-grok/spec.md#唯一当前状态)维护；该次授权已消耗，不能转用为正式授权。以后每次付费运行须先按 [SOP](../agents/verification.md) 固定当前来源、构建、目录、SDK 与完整场景，取得新的明确授权，保存原始 grant，并用工具的 `--input` 逐文件固定本次授权及所有实际参与输入；凭据只通过该路线专用环境字段注入，不进入 argv 或输入快照。grant 为严格 JSON，schema=1，字段如下：

| 字段 | 固定事实或人工选择 |
| --- | --- |
| id, accountAlias, regionAlias | 本次新 ID 及人工认可的非敏感别名，ASCII 字母/数字/下划线/连字符，1–64 字符 |
| mode, stage | live、candidate |
| combo, provider, api, endpoint, model | xai-responses、xai、openai-responses、全球 URL、grok-4.7 |
| sourceSha256, binarySha256, catalogVersion, catalogHash, sdk | 本次冻结二进制输出的机器事实，逐项相等 |
| expiresAt | 未来不超过 24 小时的 UTC 时间 |
| expected | 下表 18 个完整 ID，原顺序，不删减 |
| maxCalls, maxAttempts | 24、24；重试为 0 |
| maxRequestedOutputTokens, maxReservedInputTokens | 24576、196608（每调用 1024 可见输出、8192 输入预留） |
| maxUSD | 人工本次明确选择的有限正金额；无默认值 |

source 来自现有完整运行时/harness 源码库存，另绑定独立候选 JSON；binary 是实际执行文件 SHA-256。首次无开关运行的候选报告可列出这些机器事实，但不能取得授权。冻结构建通过 `-ldflags=-X=github.com/tokenbeat-lab/barness/ai/live.xaiBuildSource=<sourceSha256>` 注入来源，再为实际二进制生成授权；paid 运行必须使用该精确二进制。构建命令也须隔离 GOENV/GOFLAGS 和所有 live/race/压力配置，保留 argv/退出、二进制与 SHA。最终执行必须传 `-test.timeout=1800s`，过程时限 1800 秒，其中保留 30 秒给最终证据；每调用 65 秒。任何失败即停止后续调用，不自动重买。

maxUSD 是按每次实际报告的 Maximum 累加的停止阈值，未知/越界立即失败并停止；不能承诺单次厂商账单硬封顶。可见预算也不是总费用上限。若操作者需要不可超支的财务硬限制，必须在既有账户侧配置并在本次批准范围中明确，不能仅靠本地停止逻辑推断。

| ID | 固定调用数与外部事实 |
| --- | --- |
| entry-stream / entry-complete / entry-stream-simple / entry-complete-simple | 各 1，四种公共入口及完整生命周期 |
| effort-default / effort-low / effort-medium / effort-high / effort-xhigh | 各 1，缺省 effort 不改写，各显式档位；high 要求非零 reasoning |
| encrypted-continuation | 2，JSON 持久化、Host TrustNativeState、完整 reasoning item 语义相等回放、无降级 |
| tools-auto / tools-required | 各 2，合法参数、Host 校验、完整工具往返 |
| tools-none | 1，无工具调用 |
| tools-parallel | 2，两次独立调用，反序工具结果仍按原 call/item 对关联 |
| user-png / user-jpeg | 各 1，完整内嵌图片 |
| tool-result-image | 2，调用配对及 PNG/JPEG 工具结果投影 |
| cache-usage | 2，固定有界重复、派生缓存 key、合法零或非零 cache，不刷命中 |

候选运行的编码请求上限 128 KiB，单图 4 KiB，frame 256 KiB，完整响应 512 KiB；输入 token 预留在实际报告后核对，越界即停止，不承诺厂商硬限额。每次调用保存完整受限 HTTP 捕获、事件、Result、终结 Observer，核对原始与规范 usage 算术及实际返回型号。截断捕获不能通过。成功、失败或取消均在 common Git 的 `barness-xai-grants/<id>.json` 以排他写消费 claim；不能删 claim 重用。候选只写 `xai-candidate-report.json`，不写通用正式 support report 或支持矩阵。

普通 E2E `TestXAILiveHarnessProcess` 用新 HOME 与最小环境编译 live 二进制，只运行精确子进程；controlled 模式无 live 开关、无真实凭据且底层 transport 为合成响应，报告 `mode=controlled` 与 `vendorHTTPAttempts=0`。它保留原生子进程文件作为普通证据，正式消费者仍拒绝。`BARNESS_XAI_CONTROLLED` 不进入工具的 paid 环境转发白名单。

## 未发布分支的正式验证入口

最终源码准备完成并审查冻结后，正式运行只接受编译后的 `-test.run=^TestXAIFormal$ -test.timeout=1800s`。它使用完整当前 `ai.BuiltinCatalog()` 的版本/hash，唯一全球 xAI 路线及型号、同样18场景/24调用、零重试和显式USD停止阈值。grant的 `stage` 必须是 `formal`，不能把候选grant、目录hash或已消费claim复用到正式入口；凭据和环境转发字段不新增。

正式元数据写入 `xai-formal-report.json`（kind=xai-formal、stage=formal），同时保存 `xai-execution-source.json` 的完整原生来源文件/hash库存。`sourceSha256` 是当前库存哈希，`buildSourceSha256` 是实际冻结构建通过现有ldflags注入的来源，`binarySha256` 是实际执行文件hash；付费运行必须逐项匹配。无开关/Key/grant仍为NOT_RUN和零厂商HTTP。

完整正式进程还写通用 `live-report.json` 投影供现有 `releasegate -live-bundle` 消费；该投影不是另一套资格算法。预算InputTokensUsed包含缓存输入，OutputTokensUsed是total−input并已包含reasoning，不能按可见1024预算截断。正式流事件原件保留具体字段并补显式 `type` 供独立生命周期核对。只有live、formal、完整、自身审计通过的原件才能取得自身正式资格；controlled结构通过仍标记formalQualified=false，并被正式消费者拒绝。

正式入口在最终报告与通用投影完成后重新写入并同步原生manifest，使其真实结束时间包含最终报告；manifest写入失败也按晚期故障处理。晚期写入、目录同步或审计失败会撤下通用成功投影，保留自身FAIL报告、原报告与实际预算及非零退出。控制故障目标和失败原件继续保留，不以重跑覆盖。正式验证及独立race/压力必须各自明确授权/选择；本分支不会自动合并main、push或发布。
