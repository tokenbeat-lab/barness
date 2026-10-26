# Grok 正式 live：17 路线待冻结计划 v2

Status: proposed immutable contract — NOT_AUTHORIZED_AT_CREATION

本计划准备维护者已选择的 Grok 自身与16条已发布旧目标。六条Kimi保持候选，只保留受控回归。此前候选v3授权已经消耗，不适用于本计划；本文件及已提供的 `.env` 来源均不授权调用。

## 执行边界

14与15软件已合入并完成两轴0项复核；无凭据构建、完整来源与非激活模板已准备，详见[不可变机器计划](verification/formal-plan-v2.json)。最终完整普通验证另由当前状态与新receipt记录。待维护者明确授权后，每路线只启动一次所选正式原生进程，按下表顺序串行；任一路线失败、取消、超时或原生正式消费拒绝即停止后续路线。失败仍消耗该次授权，不自动重跑。每完成一路线，在同一paid HEAD取得其独立原生资格和外层schema2只读消费者原件。

所有路线拟显式使用已验证可用的本机代理 `http://127.0.0.1:7890`；此选择仍属本次待批准范围。仅所选路线从原仓库 `.env` 读取一个指定变量到进程环境；不 `source .env`，不把文件或密钥加入argv、来源快照、日志或哈希。账户别名和地区标签按[来源清单](verification/credential-sources-v2.json)，地区标签说明固定endpoint范围，不声称已证明账户资格。实际服务/账户拒绝保留为失败，不扩大权限或自动换模型/API。维护者另已确认国际MiniMax为标准API账户且有MiniMax-M3权限，只有该账户配置 `purpose=standard` 与 `entitlement=confirmed`；此事实确认仍不授予本次调用权限。

## 完整原生范围及上限

数量以最终审查后的原生模板为准，冻结前必须复核；完整自身能力与新增 `xai-origin-history` 均为必交。表内所有数值都是现有原生限额，不是账单承诺。

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

合计238必交场景，最大444 logical calls / 472 attempts；Zhipu的attempt预算为预留上限，零自动重试下实际HTTP不超过其calls。场景内工具续接可能使用多次调用，不能把场景数当作请求数。

- xAI：每次1024可见输出、8192输入预留；累计24576 requested output /196608 reserved input，零自动重试，响应估算累计停止阈值拟5 USD；此阈值不是账单硬上限。call65秒、finalization30秒。
- Generic六条：每次4096输出，32 calls内保留原生每场景最多一次环境重试，单次Client不自动重试；scene240秒、call180秒。不虚构原生没有的累计输入或USD硬上限。
- Zhipu两条：每次4096输出、零自动重试，call120秒、finalization20秒。
- ZAI两条、MiniMax CN三条：每路线累计98304 requested output /120000 reserved input、零自动重试，call120秒。
- 国际MiniMax三条：每路线累计40000 requested output /120000 reserved input；保留原生有限rerun规则且始终受20 calls/24 attempts约束，不恢复历史closed ledger，call120秒、finalization20秒。

所有原生进程均使用明确1800秒test timeout；17条串行进程上界8.5小时，消费/构建另记，不把父子耗时相加。账户资格及厂商能力由实际响应与原生消费者判定，未知usage继续保持未知。新grant每条独立领取、最长24小时有效；原始人类授权、grant、claim及所有失败/取消/未知材料保留。

## 冻结与产物（尚待完成）

- 最终软件pin `421efc23f82ada2dd356dc9c8bb6cf0aef0cf5ab`；完整source/file map、Builtin2026-10-26.1及hash、SDK、精确models/endpoint见[机器计划](verification/formal-plan-v2.json)。
- 两轴[v7复核](review/review-v8.md)均0项；有限1457case/120入口为[cases-v4](verification/cases-v5.json)，119个额外实际输入列于[普通输入](verification/ordinary-inputs-v2.json)。最终完整普通receipt待执行，不由本不可变计划自报PASS。
- 四种冻结producer与四种只读CLI的实际argv/exit/binary hash均已保存[构建回执](../../.evidence/barness-ai-grok/build/20261026T091716-formal-freeze-99enr2ih/formal-build-receipt.json)；无Key xAI正式精确入口NOT_RUN、零调用，不产生真实资格。
- 每路线完整机器scope、非秘密配置与非激活grant模板见机器计划；id为空、expiry为零，本次原始人类付费授权仍缺少。尚未签发实际调用权限。
- 真实live与原生消费：NOT_RUN。
- race、压力、正式发布：NOT_RUN，须在适用live通过后另行人工选择。

最终执行统一使用 `tools/verification/run.py --feature barness-ai-grok --stage live --route ROUTE --input <全部冻结非秘密输入> -- <该路线冻结二进制> -test.run=<精确入口> -test.timeout=1800s`。所有固定输入采用相同集合及内容，消费者在paid HEAD捕获；不覆盖历史candidate pin，也不在最终门槛之后修改软件继续复用旧资格。

原v1计划与所有旧pin保留；v2仅更新E2E证据布局后的外层来源与有限清单。四类native source、实际payload/SDK/价格及所有冻结binary均未变，见[适用性核验](../../.evidence/barness-ai-grok/native-build-applicability-v2.json)。原完整普通M3包269测试已通过，按未变代码范围沿用；本次完整复验限定三个实际受影响包，主E2E与32case辅助GLM均须实际原生消费通过。
