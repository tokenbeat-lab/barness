# Ticket 04 — MiniMax Messages

2026-10-18，公共 Client / 外部 loopback Provider 的合成协议证据；非 live 通过或正式价格。
实现提交 `e769362`；集成 tip 已合入，最终验证在 `f68f38d`。候选版本为
`minimax-six-candidate-2026-10-18.2-synthetic-prices`。所有库和测试源文件 ≤500 行。

最终结果：138 个 Messages/元数据 cases PASS；相同 138 cases race PASS；297 个原厂
Anthropic、GLM Messages 和 managed-effort 回归 cases PASS；core 与 vet PASS。
每包包含冻结 manifest、原始合成请求/SSE、事件、Result/Attempt、断言和脱敏 audit；用量
场景另有 Observer 生命周期。归档复制后核对了所有 3933 个 manifest 文件摘要。

```sh
go test -v ./ai/e2e -run '^TestMiniMax(Messages|ThinkingSwitchOnlyMetadata)' -count=1
go test -race -v ./ai/e2e -run '^TestMiniMax(Messages|ThinkingSwitchOnlyMetadata)' -count=1
go test -v ./ai/e2e -run '^Test(Anthropic|GLMMessages|ManagedEffort)' -count=1
go test ./ai
go vet ./ai/...
(cd .scratch/barness-ai-minimax/evidence/ticket-04 && shasum -a 256 -c SHA256SUMS)
```

归档内 `assertions.json` 的 replay 可单独复跑该 case。`green-final.tar.gz`、
`race-final.tar.gz`、`regression.tar.gz` 分别保存上述三组最终完整包；对应日志给出命令
结果。manifest 中的 UTC 目录时间是北京时间 2026-10-18 的运行。

失败先行记录（每项 `red-NN.log` / `red-NN.tar.gz`，原始失败保留）：

| Red | 失败切片 | 最终公共验证 |
| --- | --- | --- |
| 01 | omission 注入 disabled；adaptive 变成 budget/display；effort-only 丢失；Temperature 被吞 | ThinkingWire |
| 02 | 未接入具名控制读凭据/运行 callback/发送 | NamedGuards |
| 03 | 精确预算、mandatory off、原厂扩展、cache/forced tool/image 穿透最终 body | FinalGuards |
| 04 | 函数参数内 file_id/type:file 业务数据误认为资源 | BusinessFields |
| 05 | 索引/类型/终态非法顺序被接受 | Sequence |
| 06 | 完整 start input 丢失；未证实 input 与片段混合被接受 | StartInput、Tools |
| 07 | 用量派生 Claude 1h cache 状态/价格 | Usage |
| 08 | 必要正文/增量字段缺失/null 被当空值 | MalformedContent |
| 09 | 未定义 null 语义的 top_p 被当合法值 | FinalGuards |
| 10 | 未闭合工具 JSON 被修复为成功 | IncompleteToolJSON |
| 11 | Host 收紧不支持 Temperature 后仍发送 | ModelNarrowsSampling |
| 12 | mandatory binary 误认为支持深度 | RequiredBinaryIsNotDepth |
| 13 | 按 map 推断深度拒绝 high-only M3.1 effort；缺少显式二元元数据 | NarrowedDepth、ThinkingSwitchOnlyMetadata |

其余最终 cases 覆盖 M3/M3.1 全入口 omission、五档 effort、预算 1/上限、Temperature
0/2/null、top_p 0/1、auto/none、缓存提示、metadata、sampling maps 忽略、schema 同名
属性、compat 不能激活 Claude 字段、四公共入口工具往返、签名片段、多工具独立 JSON、
持久化信任/账户/租户/型号降级、unsigned/empty/redacted，以及完整/部分/零/未报告/中断
usage 与 Observer。请求 capture 核验资源路径、单一 x-api-key、SDK `beta=true` query 和
`Anthropic-Version: 2023-06-01`；query 不被当成 beta 能力。

`ThinkingSwitchOnly` 与 `RequiresThinking` 分别声明二元语义和不能关闭；支持 map 可
进一步收紧，但不改变这两种语义。值参加 JSON/Hash/Client 自有快照，零值保持原厂行为；
M3 的三个独立候选都声明 true。详见 ADR-0040。候选目录是可信 Host 声明，adapter 不按
ModelID 硬编码已知型号事实。

真实缺口保留给工单 07：M3 自身签名/工具续接与实际块序、start input/partial_json 组合、
是否合法返回 unsigned/empty/redacted、实际诊断 ID 位置。M3.1 需自己的账户资格与真实
证据，不能继承 M3。当前仅可信签名原生续接；unsigned 可见文本降级，opaque 结果保留且
未证实的同来源原生回放明确拒绝。完整 start input 后追加 delta 目前明确拒绝，没有借
Claude/GLM 样本假定 MiniMax 组合语义。自动缓存不承诺 TTL 或主动写缓存；测试费率和
candidate Hash 不构成正式支持。真实错误、重试和资源隔离由后续 05/07 收敛。

来源/fixture SHA 和每包 case 数量见 `provenance.json`；fixtures 在
`ai/e2e/testdata/minimax-messages/`，所有内容和签名均为合成值。
