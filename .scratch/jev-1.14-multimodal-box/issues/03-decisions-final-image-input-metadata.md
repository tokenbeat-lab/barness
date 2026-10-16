# 03：让 Host 核对 Decisions 最终图片身份

Status: resolved

**What to build:** Host 从成功结果得到最终图片 MIME/hash，能识别 Hook 换图及重复图片，失败时不消费半成品元数据。

**Blocked by:** 02：通过统一引用调用 Decisions 上下文图片。

## Acceptance criteria

- [x] 在修改实现前列出本票所有失败方式，先写经 Client.Classify / HookedClient.Classify 的公开 E2E 并确认红灯，再实施；不在实现后补单元测试。
- [x] 新增 ImageInputs 结果契约和严格编码：每项是最终 native 图片的规范 MIME 与 decoded 文件 SHA-256，只有成功结果发布。
- [x] 按最终 Decisions 图片 content 次序记录；同 bytes 的多个 block 可以产生相同 hash 的不同项，不虚构与原请求索引相等。
- [x] Hook 合法换图/增删/重排后元数据对应实际发送内容；检查完整 data URL、格式、权限与预算，非法输入不创建网络 Attempt。
- [x] 坏答案、坏 usage、取消及其他失败不发布 Answers 或 ImageInputs，唯一已验证 usage 与 ResponseModel 按原合同保留。
- [x] 文本调用省略新字段；元数据深复制、无图片 bytes/vendor key，hash 不进入 Observer 或共享诊断。
- [x] 元数据计算有界并尊重取消；不引入图片持久化、全局缓存或跨调用引用。
- [x] 结束时产出可重复、可核验的图片/fixture hash、输入、最终请求/响应、结果/错误、Attempt/Observer、断言、配置、完整性清单及脱敏审计；仅受控 HTTP 结果不得声明真实服务 PASS。

## Traceability

- 规格要求：R11、R13、R14、R15。
- 验收组：T10、T11、T13、T15。
- 用户故事：30、31、40、41、42、44、46。

## Scope and hand-off

- 本工单的拆分已获用户确认并发布；执行时须先满足全部阻塞依赖。本次发布不代表已实现、已验证或取得部署授权。
- 单票只声明其已经打通的能力，其余新能力仍明确拒绝；授权/容量/Hook 最终复核随功能同时交付，不推迟到整合票。
- 复用现有运行时、测试记录器、图片和 JSON 能力；新增依赖前核查已有依赖/类型定义，不创造无真实第二用途的公共接口。

## Comments

2026-10-16：用户确认拆分，按已批准的标题、交付和依赖发布；尚无实现或验证执行记录。


2026-10-16：已完成最终 Decisions ImageInputs 契约与严格编解码。仅在所有答案合法且调用未取消时发布最终 content 次序的 MIME/decoded 文件 SHA-256；重复 block 保留独立项，Hook 换图、增删、重排按最终内容复核。有限策略、权限、取消、重试、usage/ResponseModel 和 Observer 隔离保持，未开放后续票能力。

实现提交 `b6175d705591f4477263e42f313b5811c62c4704`。先写[失败清单](../03-failure-modes.md)，公开 E2E 身份及 codec 确认红灯后实现；未新增单元测试。`go test ./... -count=1`（6,801 PASS / 1,050 显式 opt-in NOT_RUN）、89 场景 race 和 `go vet ./...` 均通过。新增场景 84/84 PASS，完整证据 47,432 个 artifact hash 与两包脱敏审计零发现。真实服务 NOT_RUN，仅声明受控 HTTP 闭环。

[验证与重放](../03-verification.md)、[双轴审查](../03-code-review.md)、[完整性清单](../03-evidence.json)。规范审查发现验证链接尚未落盘，已补齐；规格审查零发现，最终无未解决项。
