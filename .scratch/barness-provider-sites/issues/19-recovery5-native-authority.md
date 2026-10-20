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


## Offline progress

已实现固定ID5/schema6、独立5实际回复/来源审查/新guard字节绑定、第四批25字段和先首18/18消费后才可用的三项firstRouteProof。只读真实关闭4核验通过，第四批compact sorted JSON摘要 `bad0fbddc2bfe1f2e14f8a0ff15940dccda93fbcaa8f54836d91cc99c063158a`；旧 recovery摘要4478ed…和thirdCohort摘要1f346a…保持。公共CLI最初RED拒绝缺失新flag；首轮81合成公共进程全部PASS、205.437s、exit0。另保留“旧4历史读取错误比当前源”“缺失exitCode或null被误当0”各公共RED，修复保持旧4活动关闭、历史只读及消费命令数字退出要求。完整最终普通回归与原件索引尚待完成；此段不代表新paid PASS、READY、race或release，当前状态仍见[规格](../spec.md)。
