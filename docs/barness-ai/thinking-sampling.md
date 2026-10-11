# Thinking-level sampling / 思考档位采样

The implementation branch uses the single verified pi-ai **1.1.0** baseline.
The complete release remains pending until its nine live routes and independent
Chinese evaluation pass. Source proof and actual gate evidence are indexed in
[implementation evidence](../../.scratch/barness-ai-thinking-sampling/evidence.md).

`Model.SamplingParamsByThinkingLevel` (`samplingParamsByThinkingLevel` in JSON)
is optional metadata for chat models. Its seven keys are `off`, `minimal`, `low`,
`medium`, `high`, `xhigh`, `max`. There is no call-level profile map and no new
`SimpleOptions.Reasoning="off"` input. Omit reasoning to start selection at off.

Requests apply named fields, model `SamplingParams`, the effective canonical
profile, call `SamplingParams`, then `OnPayload`. This is one top-level merge at
the request builder: objects and arrays replace whole values. `null`, `0` and
`false` survive. Missing/empty profiles do not erase general defaults; no maps
means no additional fields. Empty profile objects are valid; null objects,
noncanonical keys and invalid RawMessage bytes fail at NewClient and Catalog.Hash,
including unused profiles. Native Anthropic/Gemini validate profile structure but
ignore both sampling maps. Consumers are OpenAI/DeepSeek × Responses/Chat;
DeepSeek Responses retains its explicit extension-route status.

The canonical profile is clamped using model capability, independently of wire
reasoning. Ordinary unset levels are supported, null levels unsupported; xhigh
and max need an explicit non-null mapping. Search upward first, then downward,
then off. Nonreasoning models use off. Different canonical profiles can share
one wire value. Profile presence never declares capability.

Full calls retain their existing un-clamped wire effort. Simple calls keep their
existing clamp. Responses full summary-only starts profile selection at medium;
omitted effort starts at off, even when off=null forces the sampler upward.
DeepSeek Chat retains its native enabled/disabled thinking rules. Sampling/callback
reasoning overrides happen after selection and never trigger reselection.

## Complete Host configuration example

The functions below compile as one Go file. The Host passes its existing trusted
binding and credential resolvers. The binding must authorize tenant, account and
`gpt-5` on OpenAI Responses at the official endpoint; credentials come only from
the Host. The finite LocalPolicy is the existing documented local load example.
All sampling values here are **synthetic demonstrations, not vendor recommendations**.

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/tokenbeat-lab/barness/ai"
    "github.com/tokenbeat-lab/barness/ai/examples/localassembly"
    "github.com/tokenbeat-lab/barness/ai/examples/requestid"
)

func configuredClient(bindings ai.BindingResolver, credentials ai.CredentialResolver) (*ai.Client, error) {
    catalog := ai.BuiltinCatalog()
    for i := range catalog.Models {
        model := &catalog.Models[i]
        if model.Provider != ai.ProviderOpenAI || model.API != ai.APIOpenAIResponses || model.ID != "gpt-5" {
            continue
        }
        model.ThinkingLevelMap.Low = ai.Null[string]()
        model.ThinkingLevelMap.Medium = ai.Null[string]()
        model.SamplingParams = map[string]json.RawMessage{
            "temperature": json.RawMessage(`0.5`),
            "top_p": json.RawMessage(`0.95`),
        }
        model.SamplingParamsByThinkingLevel = map[ai.ThinkingLevel]map[string]json.RawMessage{
            ai.ThinkingLevel("off"): {"temperature": json.RawMessage(`0.7`)},
            ai.ThinkingHigh: {"temperature": json.RawMessage(`0.2`), "top_k": json.RawMessage(`20`)},
        }
    }
    return ai.NewClient(ai.Config{
        Catalog: &catalog, Policy: localassembly.LocalPolicy(),
        Bindings: bindings, Credentials: credentials,
    })
}

func demonstrate(ctx context.Context, client *ai.Client, tenantID, bindingID string) error {
    scope := ai.CallScope{TenantID: tenantID, RequestID: requestid.New()}
    target := ai.Target{BindingID: bindingID, ModelID: "gpt-5"}
    req := ai.Request{Messages: []ai.Message{ai.UserText("Give a short answer.")}}
    full := ai.ResponsesOptions{
        ReasoningEffort: ai.ThinkingLow,
        SamplingParams: map[string]json.RawMessage{"temperature": json.RawMessage(`0`)},
    }
    // Profile high contributes top_k=20; call temperature=0 wins.
    // Full wire effort stays low. The Host owns result handling.
    result, err := client.Complete(ctx, scope, target, req, full)
    if err != nil { return err }
    elapsed, _ := result.Message.DurationMs.Get()
    fmt.Println("durationMs", elapsed)
    scope.RequestID = requestid.New()
    // Simple wire effort is high; the selected profile remains high.
    _, err = client.CompleteSimple(ctx, scope, target, req, ai.SimpleOptions{
        Reasoning: ai.ThinkingLow,
        SamplingParams: map[string]json.RawMessage{"temperature": json.RawMessage(`0`)},
    })
    return err
}
```

NewClient takes independent copies of every map and JSON byte slice. Catalog
lookup/list/client-return values also own their copies. Modify inputs only after
handoff, never concurrently with constructor reads. Each logical call freezes
its options/model; authorized request bytes and callbacks are resolved once,
then reused across bounded initial-request retries. Final authorization and byte
limits still apply after callbacks. Observer carries no sampling values.

All chat success/error/abort results and their terminal event expose nonnegative
integer `durationMs`, rounded from monotonic elapsed time from call setup through
settlement, including hooks and retries. A pending PartialView can lack the field
until it settles. Independent Go/pi execution times differ; finite ledger scopes
permit only numbers and existing shared-view timing, while final presence and
terminal/result equality remain mandatory assertions.

## 中文说明

上述示例可直接作为 Go 文件编译。Host 提供既有受信的 Binding/Credential resolver，
必须授权对应租户、账户、协议与型号；不从普通请求或环境中隐式取凭据。示例采样数值是
合成值。模型的通用默认与分档默认都是配置元数据，调用只提交 SamplingParams 覆盖。

顺序是具名选项 → 通用默认 → 有效 canonical 档默认 → 调用覆盖 → OnPayload。
只浅合并顶层字段，null/0/false 保留；对象和数组整体替换。canonical 顺序是
 off/minimal/low/medium/high/xhigh/max，普通档缺映射表示支持，null 表示不支持，
xhigh/max 需要显式非 null 映射。不可用时先向上、再向下、最后 off；非推理型号用 off。

本例 low/medium 不支持，所以 sampler 用 high。Full 仍发送 low，Simple 发送 high；
Responses 只有 summary 时从 medium 选档，省略 effort 时从 off 选档。off=null 可能向上
选择 profile，却不自动增加 effort。回调和调用 map 改 reasoning 后不会重新选参。
同一 wire 值也可以对应不同 canonical profile，不按 wire 值反查档位。

消费范围只有 OpenAI/DeepSeek 的 Responses/Chat；原生 Anthropic/Gemini 忽略两个采样
map，但新 map 的全部档位仍校验结构。消费协议还逐档校验既有保留键。空 outer map、
空 profile 对象和参数值 null 有效；null profile 对象、坏 key/JSON 无效。构造与 Hash
使用同一校验。所有公开快照深拷贝三层，冻结重试字节，回调后的授权和限额仍生效。

当前核验的已支持内置型号没有这两种默认，不填入演示值。正式发布必须使用当前目录
hash 的九组合真实报告及独立效果包，不能用历史 PASS 替代失败。回滚应整体恢复实现、
目录、oracle lock/provenance、账本和发布合同；Host 同时撤回新字段配置。没有数据库迁移，
也没有旧内部路径 shim 或双写。
