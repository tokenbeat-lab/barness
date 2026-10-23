# 国际站候选有界退避续验范围05（用户直接授权）

用户最新指令：**拉大间隔时间重试，如有类似失败请做相同处理**。本次授权取代上一范围在同类临时厂商失败后停止整个范围的限制；其他失败仍停止。原始字节见`.evidence/barness-ai-kimi/authorizations/04/approval.json`。既有Chat完整9/9证据保留，本次只续验Responses和Messages。

不修改SDK、native源码、二进制或错误归一化；使用已审查来源`2f1513067bedf1431f1e702b6b24e222543f22bf992e01e7bd64fba3399dd1b4`、binary`00b595e39dd242ead429123c45b49e217648ec4b56d3e002cc28f9767e1c4ff6`。同一来源已有14071必需普通PASS；新的grant/执行输入先作新冻结普通和输入集合核对。KIMI/GLOBAL，仅所选KIMI_KEY。

每个新native进程前/每次调用前等待60000ms，完整固定九项、9调用/9HTTP Attempt，每次4096输出上限、120秒调用窗、进程1800秒；不在同一Client调用或已开始的流内自动重放。每次失败仍消耗grant，原始FAIL与NOT_RUN保留。下一次是新grant、新请求ID，从头完整重跑九项，不拼接部分捕获。

每路线最多4个完整进程、两路线最多8个进程，理论≤72调用/72HTTP Attempt。每进程USD预留上限1.50，**上一轮已预留US$0.964002＋本轮所有进程已实际预留USD累计+下个grant全额上限必须≤4.50**才可启动，单进程绝不返还已预留额度；未知预算/缺原件即停。新一轮额外预留至多US$3.535998；该总上限包含上一轮Chat与失败Responses，不另增一份4.50额度。不是实际账单。

整体开始 `2026-10-23T03:15:29Z`，绝对截止 `2026-10-23T05:15:29Z`（上海2026-10-23T13:15:29+08:00），最多2小时含准备/普通验收。首轮原生等待60秒；同路线第2/3/4轮须在前轮完全结束后分别冷却120/240/480秒，再执行新grant自带的首项60秒等待。Agent按本次有限清单逐轮选择执行，不建立常驻调度或恢复控制器。

只允许实际HTTP429/503结构化限流/过载，或HTTP200 SSE error/response.failed中的已识别限流/过载代码；缺额度/余额、认证、取消、到期、协议/用量/成本、工具缺失、资源/审计、证据不全、未知错误和预算不足均不自动续跑。每次相似失败必须先保留并核验完整原件、资源归零、原生失败退出、自己的-qualify失败，再冷却；不能把服务端错误记为成功。响应自由文本不进入公共SDK诊断。

| 路线 | 轮次 | 前次结束后额外冷却 | 本grant截止UTC | 新grant SHA256 |
| --- | --- | --- | --- | --- |
| kimi-intl-responses | 1 | 0秒 | 2026-10-23T05:15:28Z | e21a0d4f81a762893c3247b5f61dc1c418b67626bbedd67fd5d9fee33c5db60d |
| kimi-intl-responses | 2 | 120秒 | 2026-10-23T05:15:27Z | 44695b7f206467179c69642fed2fea5df2621b053ef1eafa783ffb91c79d282b |
| kimi-intl-responses | 3 | 240秒 | 2026-10-23T05:15:26Z | 7666a5e1e5537abf489030c77afe384845025e27cc67e7f2f6c2c4e979ae5164 |
| kimi-intl-responses | 4 | 480秒 | 2026-10-23T05:15:25Z | 34b60fef457b03e318124fabc8c9da7631bdcc1b27bc27ba54c82c5ad5ff7dc8 |
| kimi-intl-messages | 1 | 0秒 | 2026-10-23T05:15:24Z | 445ec780a8b9584b37f69a9dd329360416034e9f42a81b01fef1527b00aef050 |
| kimi-intl-messages | 2 | 120秒 | 2026-10-23T05:15:23Z | 78ea545d3a8cb96f1abca81261c7e7e326027baa6dce662568af794d12fe3aff |
| kimi-intl-messages | 3 | 240秒 | 2026-10-23T05:15:22Z | fb6f6040862b1ed2ff918846b09735f7a12ed5ccfafab445e4cff97457cfbfa4 |
| kimi-intl-messages | 4 | 480秒 | 2026-10-23T05:15:21Z | ec6ed7c29caee147b3ca193f155953a4b798a9d6cc74036cbc66bea7023934d0 |

每份grant是这次用户批准的独立尝试，ID和明确期限不同；旧grant/账本完全保留，不仅靠更换ID复用同一已消耗scope。原生ledger会独立拒绝重复ID和scope。每次执行只运行精确TestRegionalCandidate，候选只交regionalverify -qualify；不调用formal消费者，不执行中国站、已完成Chat、正式验证/晋升、旧路线、race或压力。

精确范围JSON、fresh8grants与命令见`.evidence/barness-ai-kimi/preparation/05/`；单进程执行器`.evidence/barness-ai-kimi/paid-execution/04/execute_one.py`，每次仅执行一个授权条目，自身没有重试循环。具体选择、退避与原件均在本轮独立记录中留存。
