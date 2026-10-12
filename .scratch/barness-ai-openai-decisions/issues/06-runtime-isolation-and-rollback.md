# 06: 证明故障、快照隔离与授权撤销行为

**What to build:** 多租户共享 Client 并发调用 Jev、Luna、chat 和 image 时，重试保持固定请求与授权，取消/故障释放资源；撤销新路线授权仅拒绝新调用，已 pin 调用按既有快照结束。

**Blocked by:** 03: 交付 score/bool 混合答案、分布与逐题拒答；04: 证明 Decisions 用量完整性与输入阶梯计费；05: 打通分类图片、三重授权与完整资源边界

**Status:** resolved
**Type:** task
**Source tickets:** 06, 09
**Requirements:** R07, R09, R10, R13, R14, R19, R20
**Acceptance cases:** OD40, OD41, OD42, OD43, OD44, OD45, OD46, OD58
**Updated:** 2026-10-13

## Responsibility

负责公共受控故障、混合操作隔离、snapshot/race/Observer 和撤销 E2E，按证据克制修复共享运行时接缝。不新建 Decisions runtime；压力与 Pi oracle 分属独立切片。协作时保留他人改动。

## Failure scenarios first

- Pi 默认重试或 504 差异改变所有协议；2xx 读取、解码、拒答或响应回调失败后重复收费。
- Host 在首次失败后换 key/目录/权限，后续 Attempt 重读；Hook 多次执行或 mutable 引用改变发送。
- 在 resolver、hook、图片验证、准入等待、body read 或退避取消后仍继续请求。
- 同名 Tenant/Binding/Model 结果、费用、观察归属串用；共享账户使授权合并。
- 大响应/错误 body 无界，Observer 拥塞影响结果或泄露正文。
- 禁用新 Binding、型号或图片授权引起自动 Jev/chat fallback，撤销改变在途快照。

## Acceptance criteria

- [x] 所有生命周期、隔离和撤销反例先由公共 Seam 表达，再修改运行时。
- [x] 默认零重试、Binding 显式有限 retry/header/backoff 正确；非 2xx 有界脱敏诊断及 x-request-id，504 差异留给自己的账本记录。
- [x] 2xx 后任何故障不重放；header/payload 各一次，response 仅在最终 2xx 一次；终态、usage 与资源正确。
- [x] connect/header/read-idle/call 超时和各阶段取消有准确 Code/Phase、Attempt、归属及 body/permit/waiter 归零。
- [x] MaxOutputBytes/MaxErrorBodyBytes/JSON 深度等有公共超限证据，不使用 SSE Frame/queue 限额代替 unary 限额。
- [x] Jev/Luna/chat/image 共享 Client、同名 Binding/Model、共享账户均隔离，race 无发现；mutation 在明确已复制 barrier 后进行。
- [x] retry 的 bytes/headers/key/catalog version/hash/image flag 固定，凭据或 Binding 后端修改入参不能改变快照。
- [x] Observer metadata-only、有界、异常不改变结果，失败包与成功包审计零发现。
- [x] disabled Binding、移除 classifier 型号白名单及撤销图片权限分别阻止新调用；已 pin 调用和已报告费用保持，无自动 fallback，其他路线继续可用。
- [x] 输出受控脚本、全部捕获、最终结果/Attempt、Observer、资源读数、race 命令和可重放撤销证据，附 SHA-256 manifest 与整包审计。

## Completion

本票证明运行时与回滚的实际行为，压力容量和 Pi parity 不凭这些 E2E 自动宣称通过。

## Comments

2026-10-12：用户批准 13 张草案并替换旧票。此票按新编号发布；下列原票评论按来源保留，ready-for-agent 不表示已实施或依赖已完成。

### 原票 06

2026-10-12：仅为新增路由扩展当前 oracle，不升级已经冻结的 Pi 版本。

### 原票 09

2026-10-12：计划 ready-for-agent 不等于已发布；本任务只写设计与实施工单。


### 本票实施（2026-10-13）

实现提交 `c9124ef`，断言补强 `fb30d25`。先写 [runtime-failures](../runtime-failures.md)
与公共 E2E，初次 14 case 的 6 FAIL 暴露 JSON error.message 诊断缺口，修复后全部通过。
生产只补有界、无歧义 JSON 诊断，沿用共享 HTTP 脱敏、x-request-id、retry 和资源生命周期；
未增加 Decisions runtime、默认重试或 fallback。

新增 63 个公共用例全部 PASS，复跑 9 个 Decisions 2xx usage/refusal/read/hook 反例及图片
验证取消，最终 73 个本票适用 case 在普通/full race 均 PASS。四路线共享 Client、跨租户
同名 Binding/Model、同账户/key 有同步并发 barrier、各自图片/答案/独立费用/Attempt/Observer；
Binding/Credential/Catalog 和回调引用在已复制后修改，重试字节、headers、key、版本/hash、
图片权限固定。三个撤销反例分别拒绝新调用，活动请求及重试按旧快照完成，已报告费用保留；
同账户下其他租户/路线继续可用，恢复授权由新调用读取。所有最终资源读数归零。

完整普通及最终显式 `-timeout 30m` 的完整 race 各 7,221 case：7,212 PASS、9 既有压力
NOT_RUN，各 815 个原有冻结差分 PASS、pending=0；编译检查、普通/live-tag vet 通过。
首次完整 race 用例通过，但收尾审计超过默认 go test 进程时限，在 11 分钟被终止；
该命令 FAIL 原样保留，另用普通 auditor 完成零发现审计，没有冒称为原命令 PASS。
最终重跑命令退出 0 且自身审计零发现，运行时策略/调用时限/断言未收窄。

两轴代码审查及证据补审已解决聚合 Observer 审计命名和历史 artifact 完整性措辞的问题，
最终 Standards 0、Spec 0。整包保留原捕获和 verdict；聚合 Observer 有字节相同、同 SHA 的
observations-* 审计视图，触发字段白名单。全部 SHA-256、case/source/脚本、资源/Observer、
实际执行收据、失败历史、整包审计和重放见 [运行时证据](../runtime-evidence/README.md)。

默认 0 / 冻结 Pi classifier 默认 2、504 的差异独立披露，不计为 Decisions parity。
本票只证明运行时与回滚；新 Decisions 差分、设计负载、live、正式目录/矩阵和发布门禁
继续由 07/08/09/10/13 交付，没有借本票改为 PASS。用户已有其他工作区改动保持。
