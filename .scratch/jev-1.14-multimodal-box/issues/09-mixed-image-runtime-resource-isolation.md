# 09：验证混合图片调用的并发、资源和撤销

Status: resolved

**What to build:** 共享 Client 在真实混合分类负载下隔离租户、执行有限资源与重试，并在取消、错误和撤销后释放资源。

**Blocked by:** 07：接通 Jev 视觉判定标准并还原公共图例；08：通过 Classify 获取 Box 与独立存在性答案。

## Acceptance criteria

- [x] 在修改实现前列出本票所有失败方式，先写经 Client.Classify / HookedClient.Classify 的公开 E2E 并确认红灯，再实施；不在实现后补单元测试。
- [x] 设计有限且声明清楚的 mixed 图片负载，实际覆盖两协议、重复 context、局部说明/标准、Box 与混合答案；normal/race 采用相同外部契约。
- [x] 相同 Binding 名、来源索引和题名的两租户并发不串图；配置/调用参数/返回值变更不改变已 pin 的内容与价格。
- [x] 引用放大、Base64/framing、Hook 构建、输出及单/总图预算在不同失败阶段准确生效；已知超限先于复制，不靠放大限额或减载换 PASS。
- [x] 覆盖发送前/读取中/等待许可/重试等待取消与 deadline，所有路径 body/许可/Observer 终态释放，内存与资源用量有可重复记录。
- [x] 默认零重试；显式有限重试逐 Attempt 保持字节/header/key/凭据/目录不变，Hook 每 Logical Call 一次，2xx 解码失败不再发送。
- [x] 逐项验证禁用 Binding、移除型号或撤销图片权限只影响新调用，已 pin 调用按原快照结束；不降级为文本/其他 Provider。
- [x] Observer 审计不含图片、框、private key、Pointer、hash 或错误正文；失败 history 和 usage 不丢失。
- [x] 只修复本票 E2E 确认的相关边界；既有其他发布缺陷保留原编号/状态，不以重复运行成功替代根因验收。
- [x] 结束时产出可重复、可核验的图片/fixture hash、输入、最终请求/响应、结果/错误、Attempt/Observer、断言、配置、完整性清单及脱敏审计；仅受控 HTTP 结果不得声明真实服务 PASS。

## Traceability

- 规格要求：R06、R07、R08、R09、R10、R13、R14、R15、R16。
- 验收组：T08、T09、T10、T13、T14、T15、T16。
- 用户故事：23、24、26、27、28、29、31、32、33、41、42、43、44、45、46、48、50。

## Scope and hand-off

- 本工单的拆分已获用户确认并发布；执行时须先满足全部阻塞依赖。本次发布不代表已实现、已验证或取得部署授权。
- 单票只声明其已经打通的能力，其余新能力仍明确拒绝；授权/容量/Hook 最终复核随功能同时交付，不推迟到整合票。
- 复用现有运行时、测试记录器、图片和 JSON 能力；新增依赖前核查已有依赖/类型定义，不创造无真实第二用途的公共接口。

## Comments

2026-10-16：用户确认拆分，按已批准的标题、交付和依赖发布；尚无实现或验证执行记录。

2026-10-16：依赖 07、08 已 resolved；实现与审查修复见 `c6bdfcd` 至 `e429932`，
测试接缝沿用已确认的公开 Classify/HookedClient。[失败清单](../09-failure-modes.md)
先于实现；所有修复均保留公开 E2E 红灯/审查复现，不补实现后单元测试。修复 Jev
Hook 引用放大的提前计数，覆盖普通 JSON 的类型、字段提升/省略、引号、数字键、
Base64 与 Unicode 表示；保留最终独立授权/资源/协议检查。

共享 Client 的两租户、双协议四调用负载覆盖重复 context、局部说明/标准、Box 与
独立存在性答案、快照/价格、分阶段容量、取消/重试/撤销、Observer 和完整资源释放。
[独立双轴审查](../09-code-review.md)最终 Standards 0、Spec 0 剩余问题。
全量 `go test ./... -count=1` exit 0：7,838 PASS / 1,050 显式 opt-in NOT_RUN；
最终相关 race 207 PASS；本票两份最终包均 191/191 PASS；`go vet ./...` 通过。
两份脱敏审计零发现。[验证与重放](../09-verification.md)及
[完整性清单](../09-evidence.json)封存最终产物和 12 份失败历史，独立核对图片身份、
固定价格、请求/响应、结果/错误、Attempt/Observer、断言、配置及资源归零。

两次全量保留等值调用/空闲时限竞争：原 `P10-OD42`，以及新 deadline fixture 的同类
症状。`f97648f` 将本票 fixture 明确为先到期的 250ms 宿主 deadline，保持负载与
aborted 断言，不修改共享超时实现；原 timeout-ordering / unary-timeout-priority
工单维持 needs-triage，原因仍待核查，后续 PASS 不构成修复结论。
真实服务协议、视觉质量和正式支持仍 NOT_RUN；目录、支持矩阵与冻结 Pi 未改变。
