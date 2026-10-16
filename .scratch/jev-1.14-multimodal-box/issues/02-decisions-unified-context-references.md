# 02：通过统一引用调用 Decisions 上下文图片

Status: resolved

**What to build:** Host 可用同一 Images 来源和显式 context 引用调用 Decisions，得到正确顺序、复用和发送前拒绝。

**Blocked by:** None (can start immediately)。

## Acceptance criteria

- [x] 在修改实现前列出本票所有失败方式，先写经 Client.Classify / HookedClient.Classify 的公开 E2E 并确认红灯，再实施；不在实现后补单元测试。
- [x] 新增公共 ImageReferences 的严格编解码、JSON 往返、深复制及引用计划；imageIndex 必须出现，零合法，负/小数/越界/缺失值、未知字段和重复键拒绝。
- [x] 省略/null/空引用与既有 Images 顺序相同；显式声明完整使用来源，按引用次序保留重复使用，不按 hash 合并或按来源下标重排。
- [x] 按公共契约解析 value 的位置与 null 槽位以作结构检查；RFC 转义、数字对象键、数组下标、重复/不存在/禁区目标和普通 null 均有公开断言。
- [x] Decisions 只发送 context，存在合法局部用途则整体 invalid_request/capability，零凭据读取、零 Attempt；不得忽略或提升作用域。
- [x] native 保留唯一 user、State text-first、随后 data URL 图片的次序；新引用不改变旧 State/问题转换、MIME 子集、refusal 或用量规则。
- [x] 贯通 context 位置能力的目录校验/克隆/查询/Hash，省略默认不改旧快照；尚未实现的 Jev 图片与局部位置不声明支持。
- [x] 来源与引用展开预算都在复制前/最终编码边界生效，子策略省略继承有限上限；Decisions 128 和型号限制按展开数取交集，额外复用不能绕过预算。
- [x] 本票即可验证权限、原生 payload Hook 加图与顺序修改后的完整复核，不把基本授权或容量检查留给后续加固。
- [x] 结束时产出可重复、可核验的图片/fixture hash、输入、最终请求/响应、结果/错误、Attempt/Observer、断言、配置、完整性清单及脱敏审计；仅受控 HTTP 结果不得声明真实服务 PASS。

## Traceability

- 规格要求：R01、R02、R03、R05、R06、R08、R14、R16。
- 验收组：T01、T02、T04、T05、T06、T08、T09。
- 用户故事：01、02、03、04、05、06、12、13、14、15、16、17、20、26、27、28、29、46、47。

## Scope and hand-off

- 本工单的拆分已获用户确认并发布；执行时须先满足全部阻塞依赖。本次发布不代表已实现、已验证或取得部署授权。
- 单票只声明其已经打通的能力，其余新能力仍明确拒绝；授权/容量/Hook 最终复核随功能同时交付，不推迟到整合票。
- 复用现有运行时、测试记录器、图片和 JSON 能力；新增依赖前核查已有依赖/类型定义，不创造无真实第二用途的公共接口。

## Comments

2026-10-16：用户确认拆分，按已批准的标题、交付和依赖发布；尚无实现或验证执行记录。


2026-10-16：已实现并完成双轴审查。统一 ImageReferences 编解码/计划、Decisions context 顺序与复用、value 槽位结构检查及发送前能力拒绝、目录位置能力与深复制、来源/展开预算及最终 Hook 复核均已交付。运行时代码提交 `15b66fd6d9ffbbf1572ddc000890bebe3a73861b`。

验证：`go test ./... -count=1`、112 个新增场景的 race、`go vet ./...` 全部通过；完整证据为 6,717 PASS / 1,050 显式 opt-in NOT_RUN，新增场景 112/112 PASS。两份 bundle 共 45,059 个 artifact hash 核验通过、脱敏审计零发现。真实服务 NOT_RUN；Jev 图片和局部位置仍拒绝，不扩大本票支持范围。

记录：[失败清单](../02-failure-modes.md)、[验证与重放](../02-verification.md)、[双轴审查](../02-code-review.md)、[完整性清单](../02-evidence.json)。
