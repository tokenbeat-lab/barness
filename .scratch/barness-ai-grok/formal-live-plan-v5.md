# Grok 正式 live 冻结计划 v5

Status: PROPOSED / NOT_AUTHORIZED / NOT_RUN

本计划在18推理探针与19受控账本隔离修复后，重新准备既定17路线的一次正式验证；六条Kimi仍候选/受控，不连带晋升。新源尚无付费授权。[唯一当前状态](spec.md#唯一当前状态)保存各阶段实际事实。

## 已完成的软件修复与来源

- 18仅给MiniMax-M3 thinking-on换固定有界三同余计算任务，继续要求该场景自身实际非空推理；模型、14场景/ID、预算、重试、消费者不变。原真实零推理FAIL保留，不改为PASS或UNSUPPORTED。[正反例索引](evidence/minimax-reasoning-probe-v1.json)。
- 19让FreshLedger测试使用独立私有Git common namespace，保留真实HEAD/对象/源码worktree与内部同namespace重启。v13发现的exec.Cmd复用问题已通过真实第二进程/新PID与runID/原生exit1/精确已消耗拒绝/零新增请求/四文件digest补齐；[v2索引](evidence/controlled-grant-isolation-v2.json)明确撤回旧v1错误证明，原件不改。
- 原验证选择偏差的3个synthetic ID、6 guards及3 ledgers全部原样保留，无真实Key或厂商请求；[偏差核验](../../.evidence/barness-ai-grok/diagnosis/controlled-store-v1/receipt-v2.json)。[v14独立两轴](review/review-v14.md)均0项。
- 软件pin `d816aacb4538f735b76df7bdacb686b621aef382`，2062默认来源；[有限清单](verification/cases-v11.json)1522主必交/127入口，另有GLM32与MM18六项辅助。119既有执行输入及单列原生合同文档未变。完整普通实际结果以新receipt及当前状态为准，不预报通过。
- [真实构建](../../.evidence/barness-ai-grok/build/20261026T203229-probe-freeze-boskjyeq/formal-build-receipt.json)4个native+4个CLI及[模板/17条无Key无grant启动](../../.evidence/barness-ai-grok/build/20261026T203229-probe-freeze-boskjyeq/qualification-preparation-receipt.json)全部PASS、0厂商调用/0claim。Regional/Generic、xAI、M3原生来源改变，GLM342原生文件map未变；全局stage来源/输入变化仍需17条同一冻结范围的实际证明，旧7条仅保留历史。二进制与原始go version -m/argv/退出均保存。

## 待人明确选择的失败处理

两种选择都仅一次固定17路线、串行、每路线只启动一次；不重跑失败路线、不发新一轮grant、不换Provider/API/model/account/endpoint。所有适用路线仍须正式资格有效才可发布，失败没有支持资格。

1. **收集所有独立路线结果（建议）**：某条发生已确定的native执行或资格失败时保留FAIL，继续其余未运行路线一次，尽量在同批收齐问题。来源/输入变化、凭据隔离异常、取消或无法可信确认的采集/进程状态立即全批停止。这样无需每遇到单条已知失败便另买整批。
2. **任一失败即停**：沿既有边界，任一执行失败/取消/超时/原生资格拒绝后，剩余NOT_RUN。

此处只是新提案；原v4及其补充的失败即停边界已经执行，不据本计划自动继续旧权限。取得具体选项的原始人类回复后才可建新17 grant/config。新授权/输入加入后，先取得完整匹配普通，再启动付费。race、压力或正式发布不在本次范围。

## 路线与预算

| 顺序/路线 | 仅本机密钥来源 | 场景 | calls / HTTP尝试上限 |
| --- | --- | --- | --- |
| 1 xai-responses | .env:GROK_KEY | 18 | 24 / 24 |
| 2 openai-responses | .env:OPENAI_KEY | 9 | 32 / 32 |
| 3 deepseek-responses | .env:DEEPSEEK_KEY | 9 | 32 / 32 |
| 4 openai-chat | .env:OPENAI_KEY | 11 | 32 / 32 |
| 5 deepseek-chat | .env:DEEPSEEK_KEY | 12 | 32 / 32 |
| 6 zhipu-chat | .env:ZHIPU_KEY | 24 | 24 / 32 |
| 7 zai-chat | .env:ZAI_KEY | 20 | 24 / 24 |
| 8 minimax-chat | .env:MINIMAX_KEY | 14 | 20 / 24 |
| 9 minimax-cn-chat | .env:MINIMAX_CN_KEY | 10 | 24 / 24 |
| 10 anthropic-messages | .env:ANTHROPIC_KEY | 12 | 32 / 32 |
| 11 zhipu-messages | .env:ZHIPU_KEY | 24 | 24 / 32 |
| 12 zai-messages | .env:ZAI_KEY | 20 | 24 / 24 |
| 13 minimax-messages | .env:MINIMAX_KEY | 14 | 20 / 24 |
| 14 minimax-cn-messages | .env:MINIMAX_CN_KEY | 10 | 24 / 24 |
| 15 google-gemini | .env:GEMINI_KEY | 7 | 32 / 32 |
| 16 minimax-responses | .env:MINIMAX_KEY | 14 | 20 / 24 |
| 17 minimax-cn-responses | .env:MINIMAX_CN_KEY | 10 | 24 / 24 |

238完整必交 = **237 PASS +1固定已声明UNSUPPORTED/零请求**（OpenAI Chat推理历史）；该能力继续不支持，不伪造PASS；其他全部必交和xai-origin-history仍必须PASS。模型、SDK、endpoint、账户/地区别名、完整场景、原生限额及精准flags由[机器计划](verification/formal-plan-v5.json)逐条冻结。国际MiniMax标准API/M3资格事实已确认，凭据来源不重复索取。

新批最多444逻辑calls/472 HTTP尝试（含预留），历史241 calls/240 HTTP另计，合计最多685/712。每native进程1800秒，17串行原生时限上界8.5小时（不是预计耗时）；普通/构建/消费者另记。

- Grok保留24calls/24HTTP、零自动重试，每次1024可见输出/8192输入预留；累计24576 requested output/196608 reserved input，响应估算累计停止阈值5 USD，不是账单硬上限。
- 其余路线沿原生token预算及原有限重试规则，无USD硬上限；unknown usage保持unknown，不补零、不承诺账单封顶。
- 每次仅所选路线对应.env Key，显式代理http://127.0.0.1:7890；不source整文件或将凭据放argv/来源/日志/哈希。只在授权后逐路线读取所选值。
- collector在仓库根采集来源/输入，实际native统一系统env -C /Users/cyber/RestoX/barness/ai/live +绝对冻结binary+原flags；grant环境为其同文件绝对路径。launcher SHA与真正producer SHA分别记录。
- 各路线先取得自身原生资格；地区保留regionalverify -formal，国际MiniMax保留keyless minimaxverify -formal -publish，然后全部保存编译releasegate的schema2只读消费者。它不代表对外发布。

原模板ID空、expiry零，不是执行权限；新grant最长24小时/单次claim，失败仍消耗该次权限。本提案没有沿用v4治理输入例外或旧grant，后续重新运行需再明确选择。全部17实际资格有效后，race和压力仍分别由人选择，release.pass=false，当前仍未正式支持。
