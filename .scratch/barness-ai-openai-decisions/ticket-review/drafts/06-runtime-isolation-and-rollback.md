# 06: 证明故障、快照隔离与授权撤销行为

**Status:** needs-triage
**Publication:** draft，待用户批准后转 ready-for-agent；尚未发布到 issues。
**Blocked by:** 03: 交付 score/bool 混合答案、分布与逐题拒答；04: 证明 Decisions 用量完整性与输入阶梯计费；05: 打通分类图片、三重授权与完整资源边界
**Source tickets:** 06, 09
**Requirements:** R07, R09, R10, R13, R14, R19, R20
**Acceptance cases:** OD40, OD41, OD42, OD43, OD44, OD45, OD46, OD58
**Updated:** 2026-10-12

**What to build:** 多租户共享 Client 并发调用 Jev、Luna、chat 和 image 时，重试保持固定请求与授权，取消/故障释放资源；撤销新路线授权仅拒绝新调用，已 pin 调用按既有快照结束。

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

- [ ] 所有生命周期、隔离和撤销反例先由公共 Seam 表达，再修改运行时。
- [ ] 默认零重试、Binding 显式有限 retry/header/backoff 正确；非 2xx 有界脱敏诊断及 x-request-id，504 差异留给自己的账本记录。
- [ ] 2xx 后任何故障不重放；header/payload 各一次，response 仅在最终 2xx 一次；终态、usage 与资源正确。
- [ ] connect/header/read-idle/call 超时和各阶段取消有准确 Code/Phase、Attempt、归属及 body/permit/waiter 归零。
- [ ] MaxOutputBytes/MaxErrorBodyBytes/JSON 深度等有公共超限证据，不使用 SSE Frame/queue 限额代替 unary 限额。
- [ ] Jev/Luna/chat/image 共享 Client、同名 Binding/Model、共享账户均隔离，race 无发现；mutation 在明确已复制 barrier 后进行。
- [ ] retry 的 bytes/headers/key/catalog version/hash/image flag 固定，凭据或 Binding 后端修改入参不能改变快照。
- [ ] Observer metadata-only、有界、异常不改变结果，失败包与成功包审计零发现。
- [ ] disabled Binding、移除 classifier 型号白名单及撤销图片权限分别阻止新调用；已 pin 调用和已报告费用保持，无自动 fallback，其他路线继续可用。
- [ ] 输出受控脚本、全部捕获、最终结果/Attempt、Observer、资源读数、race 命令和可重放撤销证据，附 SHA-256 manifest 与整包审计。

## Completion

本票证明运行时与回滚的实际行为，压力容量和 Pi parity 不凭这些 E2E 自动宣称通过。
