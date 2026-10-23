# 修订后六路线候选实际执行结果02

本文件是本次授权的不可变执行记录；唯一可变状态仍在[spec](../spec.md#唯一当前状态)。原始批准[02](../../../.evidence/barness-ai-kimi/authorizations/02/approval.json)来自用户对提案03的明确同意，六路线各≤9调用/9Attempt、零重试、各≤US$1.50、合计≤US$9；实际执行提交d291e9b5eba87f717d3eb3fcee56c95649806a2e，native来源22b841baaf4b6a61f982a29cef8ed4199bdb05748a413705944955911f406eea。与普通实测2ffa761的1351份执行来源和135项选定输入相等。

| 路线 | 候选结果 | 实际调用 | 原生退出 / 候选消费退出 |
| --- | --- | --- | --- |
| kimi-cn-chat | PASS | 9 | 0 / 0 |
| kimi-cn-responses | PASS | 9 | 0 / 0 |
| kimi-cn-messages | PASS | 9 | 0 / 0 |
| kimi-intl-chat | FAIL | 4 | 1 / 1 |
| kimi-intl-responses | NOT_RUN | 0 | — / — |
| kimi-intl-messages | NOT_RUN | 0 | — / — |

共31调用/31个HTTP Attempt，零重试；30项实际PASS，1项FAIL，23项计划用例未执行。四份已执行grant永久消耗；余两份ID未claimed，但整个批准范围已首个真实候选失败停止，不能自动续跑或改ID复用。预算估算预留US$2.2780420208321455，不是厂商账单，实际账单未知。正式验证、晋升、旧路线、race和压力均未执行。

## 首个真实候选失败

国际站Chat前三项PASS，第4项simple-complete返回HTTP429、rate_limit_reached_error；服务响应明确该账户organization max RPM为3，Retry-After=1。统一错误分类rate_limited；本次零重试，usage/cost保持unreported，未把缺用量当免费。Calls、Bodies、Events、Permits、Waiters最终全0。该路线剩余五项及国际站Responses/Messages各九项NOT_RUN。此失败是本次账户吞吐限额触发，发生在合法请求的服务端响应处；已有三条中国站完整候选通过，不改写为正式资格。

后续见[09](../issues/09-international-rate-limit.md)：先处理账户限额/调用节奏，再准备新的具体范围；源码不变时评估既有中国站证据适用性，若调整原生验证器而改变来源，则按原生规则重新评估所有受影响资格。不得放宽九项完整性、隐式重试或刷新旧pin。

## 执行阶段纠正

初始运行器在中国站Chat九项及regionalverify -qualify都exit0后，误调用了正式发布消费者consume_live.py/releasegate -live-bundle，后者正确拒绝candidate材料、exit1。qualifiedLive要求live-report.json及formal=true，candidate只应由-qualify判定；这是执行方案的阶段选错。原始拒绝及初始STOPPED摘要保留，不改成PASS。只读核对后，继续同一用户批准中尚未消耗的五份candidate grant，未增加中国站Chat调用，未改变来源、二进制、输入、上限或到期时间；在国际站Chat真正候选失败后停止。正式发布消费者仍待实际正式阶段。

## 原件入口

- [完整六路线合并索引及精确bundle/receipt/消费者哈希](../../../.evidence/barness-ai-kimi/paid-execution/02/scope-result.json)
- [全部实际断言、四份原生审计、八份账本原件副本核对](../../../.evidence/barness-ai-kimi/paid-execution/02/assertion-verification.json)
- [初始运行器原件](../../../.evidence/barness-ai-kimi/paid-execution/02/summary.json)与[阶段纠正诊断](../../../.evidence/barness-ai-kimi/paid-execution/02/stage-clarification.json)
- [剩余五路线原件](../../../.evidence/barness-ai-kimi/paid-execution/02/remaining/summary.json)

没有改运行时或测试源码，没有重跑普通/race/压力。全部捕获与原始grant/ID和scope账本均保留；未对未执行用例伪造断言。只读汇总初次将NOT_RUN误当成有实际断言文件的读取失败也独立保留，修正读取逻辑后核对成功，不影响任何原生结果。
