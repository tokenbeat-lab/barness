# Ticket 03 — GLM Chat 工具与同轮推理证据

日期：2026-10-17（Asia/Shanghai）。批准接缝为 [spec T01](../../spec.md) 的公共 `ai.Client`，Host 执行工具并显式发起第二轮；本票只使用 controlled TLS Provider 与两个独立的合成 Host Catalog 型号。实现提交 `2769e46`；合并集成分支 `9ea22d0` 后的已提交重放源码为 `4975214`。

[失败方式](failure-modes.md) 在实现前记录；[索引](index.json) 固定每包 manifest SHA-256、源码提交、案例数、状态和脱敏审计。逐包请求、SSE 脚本、事件、Result、Attempt/Usage、错误和断言保存在主工作区 `<repo>/.evidence/barness-ai-glm/ticket-03/`；每次 TestMain 的 `audit.json` 均为零发现。脚本中的提示与工具数据全部合成，来源说明在 [tools.json](../../../../ai/e2e/testdata/glm-chat/tools.json)。未读取 `.env`，未运行外部或收费请求。

| 片段/检查 | 结果 |
| --- | --- |
| 两型号工具往返 red → green | 4 例缺 tool_stream → 4 PASS；非连续 index、ID 单独续片和多工具分片分别组装 |
| named/simple/final 控制 red → green | 78 例非法工具选择/保留式思考被发送 → 114 PASS（含往返） |
| 最终数值精度 red → green | 4 例微小小数被舍入接受 → 12 PASS |
| 采样数值精度 red → green | 12 例 model/profile/call 小数提前舍入 → 48 精度例 PASS；同包历史/错误 20 PASS |
| 快照与其余公共入口 | 8 PASS；header→payload→response、调用输入/回调 map/Partial Snapshot/Result 隔离 |
| pending 工具调用 | 2 PASS；Host 在参数尚未结束时拒绝执行，取消保留唯一终态与单 Attempt |
| 已提交完整 GLM Chat 重放 | **358 PASS**（原 02 的 166 + 本票 192） |
| 本票全部新增案例 race | **190 + 2 PASS** |
| 受影响 Chat/DeepSeek Chat/回调/采样回归 | **1999 PASS**；1 个 opt-in oracle guard 初次 NOT_RUN，随后独立 **1 PASS** |
| 冻结 pi OpenAI Chat / DeepSeek Chat | **321 PASS，零 pending**；不计为 GLM 原生路由差分 |
| go vet ./ai/...、git diff --check | PASS |

工具 Host 流程实际经过 `ValidateToolCall`→参数 Decode→`ToolResultText`→第二次 Client 调用，两个型号分别保留必要的 reasoning_content、call ID、结果顺序和最终回答。普通 assistant 不增加 DeepSeek 专用的空 reasoning_content。可信持久化记录必须重新 `TrustNativeState`；反序列化本身没有信任。未受信、跨型号/Provider/API/账户历史保留可见推理和有效工具配对，记录正确降级原因。

默认/auto 可用，full required/none/function/allowed_tools 与 simple none 拒绝；clear_thinking:false 或非法值拒绝，省略仍保留厂商默认 true。model/general、canonical profile、调用 SamplingParams 和最终 OnPayload 都受检查，合法 hook 可修复中间采样，授权错误保持优先级。所有发送前失败均为零 Provider 请求、零 Attempt。GLM 最终输出预算精确检查 JSON 数字，保留合法整数值 1.0/1e0，避免微小小数被浮点舍入后放行；旧 Provider 的 JavaScript 采样语义不变。

初始诊断包全部保留：`red-tools` 一度用了错误的工具结束摘要格式，正式行为 red 取 `red-tools-corrected`；`history` 最初误以为 q:3 不会被既有工具校验转为字符串，随后改用无法转为整数的 n 字段并通过，不修改既有校验语义。这两类诊断不被计作生产缺陷或业务 red/green 证据。

## 重放

```sh
BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-glm/ticket-03/replay" \
  go test ./ai/e2e -run '^TestGLMChat' -count=1
BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-glm/ticket-03/replay-race" \
  go test -race ./ai/e2e \
  -run '^TestGLMChat(Tool|SimpleTools|FinalToolControls|IncompleteToolCalls|PendingTools|OutputBudgetPrecision)' -count=1
go vet ./ai/...

# 已按冻结 provenance 装好 oracle 依赖后：
BARNESS_AI_PIDIFF=1 go test ./ai/e2e \
  -run '^TestPiDifferential$/(^openai-completions$|^deepseek$)' -count=1
BARNESS_AI_PIDIFF=1 go test ./ai/e2e \
  -run '^TestThinkingSamplingBaseline$/^oracle-guards$' -count=1
```

每个 manifest 的 replay 字段可重放单例；red 包表示实现前的观察，当前源码重放应为 green。完整真实 Chat 服务验证、标准 Key/账户资格、价格快照、正式目录与支持矩阵由后续票验收；本票没有改动这些发布声明。
