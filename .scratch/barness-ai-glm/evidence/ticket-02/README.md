# Ticket 02 — GLM Chat 公共 Client 证据

日期：2026-10-17（Asia/Shanghai）。接缝：[spec T01](../../spec.md) 已确认的公共 `ai.Client` 四个入口与 `WithHooks`，本地 controlled Provider / 可信 Host。首次实现提交 `2c291dd`；合入当时最新集成分支后的已提交重放源为 `89069c811f0849403b5c8481717192c8134fa8f2`。此后工单关闭只增加证据索引和状态，不改代码。

[失败清单](failure-modes.md) 先于实现写出；[完整索引](index.json) 固定每包 manifest SHA-256、代码提交、案例数及结果。原始请求、SSE script、事件/Envelope、Result、Attempt、错误、候选目录/Hash 与断言在 `<repo>/.evidence/barness-ai-glm/ticket-02/`，均在主工作区保存，未提交大体积运行产物。每次 TestMain 完成后无条件审计脱敏，均无发现。

| 片段/检查 | 结果 |
| --- | --- |
| text red → green | 16 个旧 OpenAI 形状失败 → 16 PASS |
| canonical thinking red → green | 3 个非法 full 档位被错误发送 → 24 PASS |
| final thinking/采样/元数据 red → green | 17 个违规最终请求或缺失字段失败 → 27 PASS |
| 默认 enabled thinking 的用量完整性 red → green | 5.2 缺 reasoning 被误报完整 → 9 PASS（含 preflight） |
| GLM 输出预算 red → green | 60 个非法实际 wire 预算/字段失败 → 162 PASS（当时全部 GLM） |
| Provider 上限 red → green | Host 放大目录错误扩大 schema 上限 → 166 PASS |
| 120 秒候选装配 red → green | 2 个沿用旧 180 秒策略失败 → 2 PASS |
| 已提交最终 GLM 重放 | **166 PASS** |
| 已提交 controlled live 装配重放 | **2 PASS**；合成 transport、零外部网络 |
| 受影响 Chat/DeepSeek Chat/采样/回调/环境回归 | 2070 PASS；1 个需 opt-in 的 oracle guard 初次 NOT_RUN，随后独立重放 **1 PASS** |
| 可信回调授权/隔离补充 | 48 PASS |
| GLM race | 76 PASS（新增无共享状态的预算 guard 前；预算另由最终 166 例覆盖） |
| 冻结 pi OpenAI Chat + DeepSeek Chat | **321 PASS，零 pending**；GLM 自身不计为原生 pi 路由差分通过 |
| go vet ./ai/...、git diff --check | PASS |

回归和冻结差分在加入纯 GLM 预算 guard 前完成；新增 guard 仅 `ProviderZhipu` 路径执行，之后按增量范围重放全部 GLM 与 vet。没有重复扩大全量门禁，也没有用候选结果变更正式支持状态。

## 重放

```sh
BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-glm/ticket-02/replay" \
  go test ./ai/e2e -run '^TestGLMChat' -count=1
BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-glm/ticket-02/replay-live-assembly" \
  go test -tags live ./ai/live -run '^TestGLMCandidateAssemblyControlled$' -count=1
go vet ./ai/...

# 现有冻结依赖已按自身 provenance 安装时：
BARNESS_AI_PIDIFF=1 go test ./ai/e2e \
  -run '^TestPiDifferential$/(^openai-completions$|^deepseek$)' -count=1
BARNESS_AI_PIDIFF=1 go test ./ai/e2e \
  -run '^TestThinkingSamplingBaseline$/^oracle-guards$' -count=1
```

每个 manifest 的 replay 可重放单例。red 包保存实现前的历史观察；当前提交重放相同场景应为 green。初始编译问题在进入行为 red 取证前修正，不计为功能失败证据。

## 结论边界与后续

两型号独立 Chat Catalog 身份、text-only、131072 schema 输出上限、1000000 barness 保守上下文和来源见 [候选目录与标准 Key Host 示例](../../../../ai/internal/testkit/glm/README.md)。价格为合成测试值，不是厂商报价。没有读 `.env`、没有运行付费调用、没有新增内置型号或正式支持矩阵条目。根任务保存的独立 Messages/Responses 探索不被本票计为 Chat 服务通过。

`newGLMCandidateClient` 与 legacy harness 仅共享 Client 装配；候选策略为 120 秒，legacy 调用仍使用原策略。工单 11 负责双开关、单组合 Key、24 Logical Calls / 32 Attempts 的发送前预留、4096 token/次和 1800 秒进程上限，以及三协议独立正式报告。工具往返/clear_thinking guard 属于 03；其他协议数据必须独立编写；完整价格/正式目录/发布门禁仍由后续票完成。
