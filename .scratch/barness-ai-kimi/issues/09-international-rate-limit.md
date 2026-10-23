# 09 — 国际站候选受账户 RPM 限额阻塞

Type: task
Status: claimed
Blocked by: 10

## Context

- [唯一当前状态](../spec.md#唯一当前状态)
- [实际结果02与原件](../evidence/candidate-live-result-02.md)
- [SOP](../../../docs/agents/verification.md)

## Scope

国际站Chat前三项成功，第4项被服务端HTTP429拒绝；本次账户响应明确organization max RPM=3，Retry-After=1。按零重试和首失败停止合同，国际站Chat剩余五项、Responses及Messages各九项未执行。三条中国站完整candidate及其原生资格均已通过；尚无六路线formal资格。

## Resolution paths

1. 由账户维护者处理该账户RPM配额及其他共享消耗，在源码/目录/SDK/二进制不变前提下，按原生规则评估中国站已通过证据，并准备国际站新grant/具体预算范围。
2. 若需调整原生验证器调用节奏，先明确有界间隔与取消/预算/超时合同，通过公共进程E2E先失败后实现、审查及冻结普通验收；新native来源须重新构建并按规则评估所有六路线资格，不能直接套用旧捕获。

不得取消必要用例、把429当成功、启用未批准重试或改旧grant继续。新的付费范围、正式验证、晋升、race和压力均须各自明确授权；本票不授予这些权限。

## Acceptance

确定并落实限额/调用节奏方案后，按SOP提供适用来源与具体预算的独立批准。国际站三路线九项及自身候选消费全部通过；随后另审晋升和正式资格。始终保留本次失败和未执行事实，实际状态只在spec维护。

## Comments

2026-10-23：依据已批准六路线提案03执行，真实候选首失败后关闭本范围。本次共31调用/31Attempt、零重试；不自动购买下一轮。

2026-10-23：用户明确要求延长间隔并继续剩余验证，选择Resolution path2。认领Kimi-only grant.callIntervalMs0..60000、可取消等待及独立时序消费者，先公开进程红绿、审查和新冻结普通；随后按每国际路线35秒间隔、9调用/9Attempt、零重试、≤US$1.50范围执行新来源。旧失败与grant不复用，中国站旧来源候选适用性另判定。

2026-10-23：有界间隔实现、425项定向、双轴复审及14071必需完整普通已PASS；国际站Chat完整9/9、-qualify退出0，上次429项通过。Responses第4项出现新的HTTP200内engine overloaded，按范围停止，详见[结果03](../evidence/candidate-live-result-03.md)。本票剩余完整国际资格未满足，由[10](10-responses-engine-overload.md)接续外部阻塞；没有自动新一轮。

2026-10-23：用户明确要求“拉大间隔时间重试，如有类似失败请做相同处理”，已直接授权[范围05](../evidence/candidate-live-review-05.md)。原生60秒/新grant完整九项；同类临时失败可120/240/480秒冷却后续轮，SDK不在已开始流内重放。最多4轮/路线及2小时，合计预留4.50包含上轮0.964002；其他失败/额度/未知/证据问题停止。新普通与原始执行由spec唯一状态记录。
