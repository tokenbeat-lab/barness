# 19 固定恢复5的原生有限续验与完整关闭历史

Status: claimed
Type: task

## Context

[规格](../spec.md)、[17固定恢复4](17-fixed-fourth-recovery.md)、[18预算修复](18-classifier-budget-operation-isolation.md)。恢复4实际停止，旧18父进程（17PASS/1UNKNOWN）、三nonce与75条ledger不可覆写；新轮尚无付费授权。本票只实现离线原生边界。

## Fixed design

固定 `provider-sites-final-acceptance-5` / schema6；保留旧 recovery 与 thirdCohort 原事实，新增 fourthCohort。旧4按原始08源码/grant/report/Git archive互证，不与当前源码比较；旧4活动入口关闭。新source/preflight零Key，actual human5必须独立批准378 Calls/414 HTTP、累计1358/1454、旧28.8＋新6.9≤36USD；所有guard源码字节明确绑定新SHA，资金selector仍为2。三M3固定60/72、每路线20/24、4096/40000/120000；先新ZAI Responses18+原生consumer成功，才可发行grant5。历史保留不是自动刷新来源资格。

## Failure ways before implementation

公共CLI先RED：缺新flags/实际human、PENDING、复用4回复或ID/预算/plan/source/review/guard指纹不符；schema1/3/4/5携带fourthCohort；旧4活动或借旧ledger；关闭4的seal/stop/index/file集合漂移、重复或别名JSON、越界路径/symlink/FIFO/超大文件；18父进程漏项或增项、17PASS伪造、UNKNOWN抹除、失败9/9与guard0/0误写、未开始六claim；三已用nonce缺失/更改、75链/39-36/旧source更改；首PASS未完成就领grant。新路由重复nonce、跨worktree/进程、并发、预算上限与取消后重启均拒绝。

## Verification

先公共CLI/独立进程RED，再实现GREEN及受影响旧authority回归、目标vet。只用合成owned fixture，保留实际argv/UTC/数字退出与失败原件；不读真实Key、不请求厂商、不执行race、不创建真实新授权/grant/claim。离线结果不关闭最终验收。


## Offline implementation and ordinary result

已实现固定ID5/schema6、独立5实际回复/来源审查/新guard字节绑定、第四批25字段和首18/18消费后才可用的三项firstRouteProof。只读真实关闭4核验通过，第四批compact sorted JSON摘要 `bad0fbddc2bfe1f2e14f8a0ff15940dccda93fbcaa8f54836d91cc99c063158a`；旧 recovery摘要4478ed…和thirdCohort摘要1f346a…保持。源325文件的完整集合与实际Go facts的执行摘要 `7b6c6ecb85a38575bffd27508091d36a78298f220e3c35530051b900e7dcc886`逐项相等，包含预算修复后的estimate新SHA db93f4…。

最终集成源码 `5078258a1ea29aaf8becbc8421dc0b68ef1eeafb` 构建持久minimaxlive.test与minimaxverify并保留go version -m；完整原生普通包259测试PASS，七个child专用入口按设计跳过，包exit0、Go墙钟554.861s、外层实际555.553587s。新5的84个独立公共CLI/进程全部exit0，包含三路线75行39/36、旧4活动拒绝而历史只读、source/guard/human/review/plan/预算、nonce/并发/跨worktree、首route proof和关闭4篡改拒绝。目标vet实际exit0；真实历史-source0、无新grant的-preflight1和冲突flags2均符合预期。

[原生开发永久索引](../../../.evidence/provider-sites/final-acceptance-20261020/recovery5-native-dev/index.json)保留原始flag缺失RED、旧4历史误比当前source RED、缺失/null exitCode误当0 RED、所有中间结果及最终binary/argv/UTC/数字退出/完整源副本。旧08/补强前组合回归被最终507替代后，按根要求精确SIGTERM其单独进程树，实际退出-15并标记ABORTED_SUPERSEDED；它不替代最终PASS。

本票Status仍claimed，等待整合审查及新5实际授权、真实资格与最终十项门禁。以上均零Key/零厂商调用，未创建真实5human/grant/claim；READY、race仍NOT_RUN，release.pass=false。唯一当前状态见[规格](../spec.md)。
