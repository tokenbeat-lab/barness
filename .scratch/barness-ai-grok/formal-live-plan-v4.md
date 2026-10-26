# Grok 正式 live：原生已声明限制修复后的冻结计划 v4

Status: proposed immutable contract — NOT_AUTHORIZED_AT_CREATION

本计划重新准备既定17路线的一次正式验证，六条Kimi保持候选/受控。v2批次39次、v3批次65次正式调用分别保留，合计104次，不计入新批上限。v3前三条正式资格通过，第4条OpenAI Chat的推理历史依既有合同为UNSUPPORTED/零调用，资格层错误要求正调用/PASS而拒绝；后13条未运行。[原停止索引](../../.evidence/barness-ai-grok/live-batches/gf3-20261026T131257Z-6a9ba85b/stopped-checkpoint-v1.json)及全部真实原件/授权/claim保留，不续用其剩余grant。

## 修复与冻结

生产`268faeba`将通用报告验证归属到现有supportmatrix合同，只允许固定OpenAI Chat × gpt-5-mini × reasoning-history的原生已声明限制；精确原因、零调用/HTTP/重试/资源/请求ID，以及仅assertions原件均校验。该能力继续UNSUPPORTED、无新的LastPassedAt或推理回放能力；伪造PASS亦拒绝。其他必交场景和全部xai-origin-history仍必须PASS。Generic六条当前固定目录分支审计见[原生合同](../../docs/barness-ai/legacy-history-qualification.md)。

`b77859c0`补完整11-scene native bundle的公共compiled CLI回归，三份真实RED证明原严格规则、报告接线、原件接线均可复现故障，61/61定向case GREEN/audit0；所有合成材料整体release=false，不产生真实资格或权限。[v12独立两轴复核](review/review-v12.md)均0项。

软件pin`b77859c0a7b94b7ef874cd8a52196f544d0e01e1`、默认来源2057文件，[有限静态清单](verification/cases-v9.json)1518必交/125入口。119执行输入未变，修改的合同文档另列为规划输入。最终四包完整普通实际结果由[唯一当前状态](spec.md#唯一当前状态)和新receipt记录，本计划不自报未发生结果。

[机器计划](verification/formal-plan-v4.json)冻结每路线model/SDK/endpoint/完整scenes/expected outcomes、原生source、binary SHA、argv及非秘密配置。[构建原件](../../.evidence/barness-ai-grok/build/20261026T141636-unsupported-freeze-i5jh1jou/formal-build-receipt.json)和[模板准备](../../.evidence/barness-ai-grok/build/20261026T141636-unsupported-freeze-i5jh1jou/qualification-preparation-receipt.json)均实际PASS；无Key Grok正式入口NOT_RUN/0调用/0claim。追加3个E2E文件不在原生编译/来源范围内，[逐项适用性](../../.evidence/barness-ai-grok/build/20261026T141636-unsupported-freeze-i5jh1jou/native-build-applicability-v4.json)核对537原生文件、119执行输入和8个image字节相等。相比付费f952，四类native source均改变，不借旧资格或旧M3 source不变说明完成当前验收。

## 一次执行边界与原生范围

本次需新明确授权后，17路线串行，每条只启动一次冻结native进程；任一真实执行失败/取消/超时或原生资格拒绝即停止后续，不自动重跑。每路线在同一paid HEAD先取得自身资格，再保存外层schema2只读消费者。MiniMax国际保留既有keyless post-exit publisher；地区路线保留其自身formal qualifier。

每次仅持有所选路线对应.env key，显式代理`http://127.0.0.1:7890`；不source整个文件或把凭据加入argv/来源/日志/哈希。[来源清单](verification/credential-sources-v2.json)沿用；国际MiniMax标准API/M3权限事实已确认，不重复索取。

| 顺序/路线 | 账户/地区与本机来源 | 完整场景 | 单进程 calls / attempts 上限 | 精确 test.run |
| --- | --- | --- | --- | --- |
| 1 xai-responses | GROK/global `.env:GROK_KEY` | 18 | 24 / 24 | `^TestXAIFormal$` |
| 2 openai-responses | OPENAI/global `.env:OPENAI_KEY` | 9 | 32 / 32 | `^TestLive$/^openai-responses$` |
| 3 deepseek-responses | DEEPSEEK/global `.env:DEEPSEEK_KEY` | 9 | 32 / 32 | `^TestLive$/^deepseek-responses$` |
| 4 openai-chat | OPENAI/global `.env:OPENAI_KEY` | 11 | 32 / 32 | `^TestLive$/^openai-chat$` |
| 5 deepseek-chat | DEEPSEEK/global `.env:DEEPSEEK_KEY` | 12 | 32 / 32 | `^TestLive$/^deepseek-chat$` |
| 6 zhipu-chat | ZHIPU/cn `.env:ZHIPU_KEY` | 24 | 24 / 32 | `^TestGLMFormal$` |
| 7 zai-chat | ZAI/global `.env:ZAI_KEY` | 20 | 24 / 24 | `^TestRegionalFormal$` |
| 8 minimax-chat | MINIMAX/global `.env:MINIMAX_KEY` | 14 | 20 / 24 | `^TestMiniMaxFormal$` |
| 9 minimax-cn-chat | MINIMAX_CN/cn `.env:MINIMAX_CN_KEY` | 10 | 24 / 24 | `^TestRegionalFormal$` |
| 10 anthropic-messages | ANTHROPIC/global `.env:ANTHROPIC_KEY` | 12 | 32 / 32 | `^TestLive$/^anthropic-messages$` |
| 11 zhipu-messages | ZHIPU/cn `.env:ZHIPU_KEY` | 24 | 24 / 32 | `^TestGLMFormal$` |
| 12 zai-messages | ZAI/global `.env:ZAI_KEY` | 20 | 24 / 24 | `^TestRegionalFormal$` |
| 13 minimax-messages | MINIMAX/global `.env:MINIMAX_KEY` | 14 | 20 / 24 | `^TestMiniMaxFormal$` |
| 14 minimax-cn-messages | MINIMAX_CN/cn `.env:MINIMAX_CN_KEY` | 10 | 24 / 24 | `^TestRegionalFormal$` |
| 15 google-gemini | GOOGLE/global `.env:GEMINI_KEY` | 7 | 32 / 32 | `^TestLive$/^google-gemini$` |
| 16 minimax-responses | MINIMAX/global `.env:MINIMAX_KEY` | 14 | 20 / 24 | `^TestMiniMaxFormal$` |
| 17 minimax-cn-responses | MINIMAX_CN/cn `.env:MINIMAX_CN_KEY` | 10 | 24 / 24 | `^TestRegionalFormal$` |



合计238必交场景：**237项要求PASS，1项固定原生声明UNSUPPORTED且零请求**（OpenAI Chat推理历史）。不支持不是成功能力；不得用FAIL/NOT_RUN/其他UNSUPPORTED代替必交成功，也不得把该项伪造为PASS。

新批次最多444 logical calls /472 HTTP attempts（含预留），历史104次正式调用另计，合计上界548 calls /576 attempts。所有native进程1800秒，串行原生时限上界8.5小时；普通/构建/消费者另记。

- Grok：24calls/24HTTP，零自动重试；每次1024可见输出/8192输入预留，累计24576 requested output/196608 reserved input，响应估算累计停止阈值5 USD（不是账单硬上限）；call65秒/finalization30秒。
- Generic六条：32calls/32HTTP，每次4096输出、每scene最多一次原生环境重试、Client不自动重试；scene240秒/call180秒，无原生USD硬上限。
- Zhipu两条：24calls/32attempt预留，每次4096输出、零自动重试，call120秒/finalization20秒。
- ZAI两条及MiniMax CN三条：每条24calls/24HTTP、98304 requested output/120000 reserved input，零自动重试，call120秒。
- 国际MiniMax三条：每条20calls/24attempt、40000 requested output/120000 reserved input，保留原生有限rerun且不恢复历史closed ledger；call120秒/finalization20秒。

所有token/响应估算限额不是账单承诺，未知usage保持未知；不自动换模型/API/Provider/账户/endpoint。每条新grant最长24小时、一次claim，失败仍消耗该次权限。模板ID为空、expiry为零，当前未授权。race、压力与正式发布不在该付费授权范围；全部适用live有效后再按序人工选择。

当前本计划v4实际调用NOT_RUN；来源变化需要新匹配普通与付费权限，旧批次及未执行项不改pin。未完成前release.pass=false、仍未正式支持。
