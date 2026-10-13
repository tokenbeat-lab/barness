# Host 从 Jev 显式接入 Luna

本地 `localassembly.OpenOperations` 接收明确的 tenant、Binding、按 CredentialRef 指定的
KeySource、自带 Catalog 和资源策略；云端 `hostintegration.New` 接收已装配 Client，并从
认证的 Principal 派生 CallScope。两种路径继续使用同一 `Client.Classify`，不会自动换 Provider。
可执行 HTTP E2E 覆盖下面的共同子集和拒绝路径：

```sh
go test ./ai/e2e -run '^TestClassifierHostMigration$' -count=1
```

## 先选择身份，发现输入能力

```go
jev := ai.Binding{TenantID: tenant, BindingID: "jev", Enabled: true,
    Operation: ai.OperationClassifier, ProviderID: ai.ProviderTypeSafe,
    API: ai.APITypeSafeSystemOne, Endpoint: "https://api.typesafe.ai/v1",
    AccountScopeID: "jev-account", CredentialRef: "jev-key", AllowedModels: []string{"jev-1.13.0"}}
luna := ai.Binding{TenantID: tenant, BindingID: "luna", Enabled: true,
    Operation: ai.OperationClassifier, ProviderID: ai.ProviderOpenAI,
    API: ai.APIOpenAIDecisions, Endpoint: "https://api.openai.com/v1",
    AccountScopeID: "luna-account", CredentialRef: "luna-key", AllowedModels: []string{"gpt-6-luna"}}
model, ok := client.Catalog().LookupClassifier(luna.ProviderID, luna.API, "gpt-6-luna")
if !ok { return errors.New("Host 未授权该型号") }
// 查看 model.Input、Capabilities.DescriptionFormats/MinChoices/MaxInputImages/ImageMIMETypes。
// Jev 的旧省略态表示 text-only、最低 1 个候选、string/object/array 说明。
```

本地把这两份 Binding 和各自的 `KeySource{EnvVar: ...}` 交给 OpenOperations，并把同一
自带目录显式传入 Catalog。云端 resolver 按认证租户解析这两份 Binding，分别返回与
TenantID、CredentialRef、BindingVersion、AccountScopeID 匹配的 Credential Snapshot。
Binding Version/AuthKind/凭据有效性由相应 Host 装配；入站 TenantID 只能是待核对的 claim。
不得将某条路线的 key 作为另一条的环境 fallback。生产 HTTP handler 在 JSON 解码前限制 body
大小，并将入站 DTO 转成独立请求，不把共享可变对象或认证声明直接当作可信身份。

## 共同输入与选项

```go
input := ai.ClassifierRequest{State: json.RawMessage(`"请处理退款"`),
    Questions: map[string]ai.ClassifierQuestion{"intent": ai.ChoiceQuestion{
        Instructions: json.RawMessage(`"按当前处理目标分类"`),
        Criteria: map[string]json.RawMessage{
            "billing": json.RawMessage(`"支付或退款"`),
            "other": json.RawMessage(`"其他目标"`),
        },
    }}}
// nil 选择目标协议默认值；显式值仅适用于各自协议。
result, err := client.Classify(ctx, scope, ai.Target{BindingID: "luna", ModelID: "gpt-6-luna"}, input, nil)
// Luna: ai.OpenAIDecisionsOptions{}；Jev: ai.TypeSafeOptions{}。
```

本地 scope 来自 `operations.Scope()`；云端调用
`h.Classify(ctx, principal, hostintegration.ClassificationRequest{Target: target, Input: input, Options: options})`。
用户取消上下文原样传递。Jev 的结构化 Instructions/Criteria 不变；Luna 仅接受 string 或允许的
null 说明、最低两个 choice。把 `TypeSafeOptions{}` 或 object/array 说明直接交给 Luna 会在发送
前返回 typed error；Host 应明确修改输入或保持 Jev，不做静默字符串化或选项转换。

图片须同时满足型号 image 能力、`Binding.AllowClassifierImages = true` 和
`Policy.Classifier.Images = &ai.ClassifierImagePolicy{MaxImages: 1, MaxTotalImageBytes: 4096}`。
还要明确全局 `MaxImageBytes` 每图限额和 `MaxRequestBytes` 完整 JSON 限额、正确 MIME/base64。
这些示例值只针对一张不超过 4 KiB 的测试图片。少任意一个条件都会拒绝；Jev text-only
不会丢弃图片后继续发送。云端权限只能由租户授权配置开启，不能来自入站 body 或 Hook。

## nil error 后仍按封闭类型分支

```go
if err != nil {
    // 用 ai.Error 的 Code/Phase 和 Metadata 定位；已知 Usage 仍可能存在。
    return err
}
for name, answer := range result.Answers {
    switch a := answer.(type) {
    case ai.ChoiceAnswer:
        // 转成 Host 的类别视图，Confidence 与 Probabilities[a.Choice] 各自保留。
        _ = a.Choice
    case ai.ScoreAnswer:
        // 保留期望分、最终 rubric 和分布。
        _ = a.Score
    case ai.BoolAnswer:
        // 是 yes 的概率；阈值归业务所有。
        _ = a.Probability
    case ai.RefusalAnswer:
        // 保存该题拒答状态；Host 决定提示、重问、转人工及其额外预算。
        // 不把它算 false、默认类别、低置信分或协议错误。
    default:
        return fmt.Errorf("Host 缺少问题 %q 的类型化结果", name)
    }
}
```

一个 refusal 不清除其他普通答案；全 refusal 仍可 stop/nil error。库不执行 fallback 或转人工，
也没有拒答原因文本可供 Host 编造。云端返回自己的视图 DTO，并按 Principal 的租户归属隔离
结果；不要将模型内容或错误正文写到共享日志。业务报告中的拒答统计来自结果类型，Observer
仍仅记录调用/Attempt 元数据。撤销 Luna Binding 或图片权限阻止新调用；不自动重路由到 Jev。
