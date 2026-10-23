# 候选付费执行结果04：60秒间隔完成剩余国际路线

按用户“拉大间隔时间重试，如有类似失败请做相同处理”和[范围05](candidate-live-review-05.md)执行，[原始授权04](../../../.evidence/barness-ai-kimi/authorizations/04/approval.json)绑定八份有限新grant。实际仅用两份第1轮grant，国际站KIMI/GLOBAL的Responses、Messages各自从头完整九项；未拼接旧进程的前三项。原生调用前等待60000ms，包含首项；同类失败后的120/240/480秒冷却没有触发。

| 路线 | 本次结果 | native / 自身-qualify退出 |
| --- | --- | --- |
| 国际站Responses | 第1轮9/9 PASS，含此前过载的simple-complete | 0 / 0 |
| 国际站Messages | 第1轮9/9 PASS | 0 / 0 |
| 国际站Chat | 同native来源前轮9/9 PASS，本轮没有付费重跑 | 前轮0 / 0 |

18调用/18HTTP Attempt均HTTP200；18项全部PASS，未再出现429或流内过载。SDK重试0、额外整轮重试0。该结果证明本次有限验证在60秒间隔下完成，不能据此推定厂商配额或容量已永久改变。此前失败、原始请求/响应、grant和账本全部保留，见[结果03](candidate-live-result-03.md)。

本轮预算预留US$1.330650，连同上一范围US$0.964002合计US$2.294652，低于批准的US$4.50；这不是实际账单，实际费用未知。两份使用过的grant永久消耗；四份可信ID/scope账本与保留副本逐字节及SHA一致，其余六份grant未使用，本范围完成关闭。

## 原件与独立核对

- [执行汇总](../../../.evidence/barness-ai-kimi/paid-execution/04/summary.json)：实际提交92d4f63fbcfbaa26f029931b327e0dffc47e99f4，包含每进程argv、真实数字退出、时间、授权、完整来源和原件SHA。Responses于03:43:33.972469Z至03:53:24.239450Z执行；Messages于2026-10-23T03:53:40.394183+00:00至2026-10-23T04:03:37.359671+00:00执行。两路线分别完整消费，没有把candidate送正式发布消费者。
- [逐项HTTP及账本库存](../../../.evidence/barness-ai-kimi/paid-execution/04/postprocess-inventory.json)：18项断言全部有效、资源bodies/calls/events/permits/waiters全0，原始seal库存、manifest全部SHA一致，两audit零发现。四事件时序逐项核对，首项以授权startedAt锚定，逐调用实际间隔最小60003.497ms。
- [Responses第1轮](../../../.evidence/barness-ai-kimi/paid-execution/04/runs/kimi-intl-responses-1/summary.json)与[Messages第1轮](../../../.evidence/barness-ai-kimi/paid-execution/04/runs/kimi-intl-messages-1/summary.json)：各自精确receipt、bundles/<run-id>、manifest/audit/report Hash、原生-qualify与账本保留指针。
- [完整普通核对07](../../../.evidence/barness-ai-kimi/acceptance/07/verification.json)：真实exit0、14071必需原生ID全部PASS，214重负载NOT_RUN，14943 Go PASS节点，墙钟444.042秒。1359源码/603选定输入和85/6/5消费目录集合前后一致；主及五辅助audit通过，断言非空且有效。
- [执行器审查](retry-execution-review-01.md)：Standards0/Spec0、10项公开CLI评估通过；额度、未知混合错误、协议/取消/超时和缺失/损坏原件不得转为可重试。先红后绿的失败原件保留。本轮没有修改SDK错误归一化或新增SDK重试。

native source仍为2f1513067bedf1431f1e702b6b24e222543f22bf992e01e7bd64fba3399dd1b4，binary仍为00b595e39dd242ead429123c45b49e217648ec4b56d3e002cc28f9767e1c4ff6；因此国际站Chat已有完整当前来源候选仍适用。旧中国站三路线候选属于前一native来源，未在本轮重买；当前来源中国站候选、六路线formal、目录晋升、旧路线、race和压力均未执行，release.pass=false。

[09](../issues/09-international-rate-limit.md)及[10](../issues/10-responses-engine-overload.md)的国际候选续验目标已完成；[11](../issues/11-responses-stream-error-diagnostics.md)保持needs-triage。唯一可变状态见[spec](../spec.md#唯一当前状态)，新不可变索引为[07](../../../.evidence/barness-ai-kimi/records/07/record-final.json)。本次没有新建worktree。
