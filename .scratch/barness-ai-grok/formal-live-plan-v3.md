# Grok 正式 live：消费者修复后的冻结计划 v3

Status: proposed immutable contract — NOT_AUTHORIZED_AT_CREATION

本计划在既定17路线范围内重新准备一次正式验证；六条Kimi仍为候选/受控。原批次`gf-20261026T112558Z-14ec1226`执行两条后按规则停止：xAI正式18/18通过，OpenAI9/9真实场景通过但证据消费者拒绝；累计39调用/39HTTP，其余15条未运行。[原停止检查点](../../.evidence/barness-ai-grok/live-batches/gf-20261026T112558Z-14ec1226/stopped-checkpoint-v1.json)及所有授权/claim/失败保留，不续用旧授权。

## 修复与冻结来源

消费者原来直接解码含非空接口Content的领域Result，无法读取真实结果。`5f933af9`改为消费者边界的局部DTO，保留归属、终态、usage、预算与库存检查；`fd02b56a`补齐mandatory-history分支。公共AI调用与JSON合同未改变。12项公开Audit/compiled CLI回归通过，原实际RED和五类篡改/六类历史反例保留；[v10独立两轴复核](review/review-v10.md)均0项。

本计划软件pin`fd02b56aa6177106ae1ef348fec0c3a4920b8a57`、2054默认来源文件；独立有限清单[cases-v7](verification/cases-v7.json)为1469主case/121入口，119额外实际输入未变。完整受影响普通需重新实际通过，实际receipt与当前来源相等性由[唯一当前状态](spec.md#唯一当前状态)记录，本计划不自报未发生结果。

[机器计划](verification/formal-plan-v3.json)固定每路线完整model/SDK/endpoint、原生source、binary SHA、argv、非秘密配置与未激活grant模板。[构建原件](../../.evidence/barness-ai-grok/build/20261026T120807-decoder-freeze-yuu2ulfr/formal-build-receipt.json)与[适用性/模板补充](../../.evidence/barness-ai-grok/build/20261026T120807-decoder-freeze-yuu2ulfr/qualification-preparation-receipt.json)保存真实命令和数字退出。追加E2E测试不在原生构建依赖内：537原生文件、119读取输入及8个冻结binary逐项相等；Grok无Key正式入口NOT_RUN、0调用/0claim。模板id为空、expiry为零，不能执行付费。

## 一次执行边界

取得本次新明确授权后，以下17条串行，每路线只启动一次已冻结原生进程；任一执行失败/取消/超时或原生正式资格拒绝，立即停止后续路线，不自动重跑。每条在同一paid HEAD取得其正式原生资格与外层schema2消费者；MiniMax国际和地区路线保留自身既有只读资格流程。

仅所选路线从原仓库`.env`读取对应变量到该进程，不source文件、复制所有密钥或把凭据放入argv/来源/日志/哈希。每路线显式代理`http://127.0.0.1:7890`。[账户来源](verification/credential-sources-v2.json)不变；国际MiniMax标准API/M3资格已获维护者确认，不重复索取密钥或资格。

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


合计238必交场景。新批次最多444 logical calls /472 HTTP attempts（含预留），不包含旧批次已发生的39次；两个批次合计上界483 calls /511 attempts。每路线1800秒原生时限，串行原生上界8.5小时，构建/普通/消费另记。

- xAI：每次1024可见输出/8192输入预留，累计24576 requested output/196608 reserved input，零自动重试，累计响应估算停止阈值5 USD；此阈值不是账单硬上限。call65秒、finalization30秒。
- Generic六条：每次4096输出，32 calls内每场景最多一次原生环境重试，Client不自动重试；scene240秒/call180秒，无原生累计USD硬上限。
- Zhipu两条：每次4096输出、零自动重试，24calls/32attempt预留，call120秒/finalization20秒。
- ZAI两条、MiniMax CN三条：每路线98304 requested output/120000 reserved input、零自动重试，call120秒。
- 国际MiniMax三条：每路线40000 requested output/120000 reserved input；保留原生有限rerun，始终受20calls/24attempt约束，call120秒/finalization20秒，不恢复旧closed ledger。

未知usage保持未知；不自动换模型、API、Provider、账户或endpoint。每路线新grant/claim绑定实际来源、配置和预算，最长24小时；失败仍消耗该次授权。race、压力和正式发布不在本次付费授权范围，须在全部适用live有效后另行人工选择。

## 当前状态

本计划v3真实调用：NOT_RUN，未获新执行授权。旧批次停止、旧来源资格及未执行项分别保留。完整普通、实际消费者、race/压力和最终只读发布判定由唯一当前状态引用不可变原件；未满足前仍未正式支持，release.pass=false。
