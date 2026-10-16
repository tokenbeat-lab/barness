# 06：在 Jev 单题说明中使用视觉样例

Status: ready-for-agent

**What to build:** Host 把一张图放进某题说明，其他题不收到该局部图片，并能用有序片段混合文字与图片。

**Blocked by:** 05：在 Jev State 原位置使用图片引用。

## Acceptance criteria

- [ ] 在修改实现前列出本票所有失败方式，先写经 Client.Classify / HookedClient.Classify 的公开 E2E 并确认红灯，再实施；不在实现后补单元测试。
- [ ] 贯通 Choice/Score/Bool Instructions 的根图片、嵌套结构与有序文字/图片数组；只在实现后声明 instructions 位置能力。
- [ ] 图片算有效非空内容，未被引用的空/null 说明仍遵守既有约束；文字容器的 DescriptionFormats 与图片位置能力分别校验。
- [ ] native 图片留在该题 Instructions，不提升到 State/context，其他题不出现这份局部引用；同图在两题使用分开计展开预算。
- [ ] 说明内普通文本不作字面替换；合法精确 Pointer 变为图片，缺失/重复/禁区/非 null 槽位与越权身份拒绝。
- [ ] Hook 改题名/说明/图像后按最终问题与出现次序验证结果和 ImageInputs；所有前后置授权与容量校验完整。
- [ ] 同一输入交给 Decisions 时凭据前拒绝，不把图片 stringify 成标签、丢弃或拆请求。
- [ ] 已有 State/context、文本说明和其他协议回归保持；不因本票把尚未实现的视觉 Criteria 宣称为可用。
- [ ] 结束时产出可重复、可核验的图片/fixture hash、输入、最终请求/响应、结果/错误、Attempt/Observer、断言、配置、完整性清单及脱敏审计；仅受控 HTTP 结果不得声明真实服务 PASS。

## Traceability

- 规格要求：R03、R04、R06、R08、R10、R11、R14。
- 验收组：T03、T04、T05、T06、T08、T10、T11。
- 用户故事：08、12、13、14、15、17、18、26、27、30、40。

## Scope and hand-off

- 本工单的拆分已获用户确认并发布；执行时须先满足全部阻塞依赖。本次发布不代表已实现、已验证或取得部署授权。
- 单票只声明其已经打通的能力，其余新能力仍明确拒绝；授权/容量/Hook 最终复核随功能同时交付，不推迟到整合票。
- 复用现有运行时、测试记录器、图片和 JSON 能力；新增依赖前核查已有依赖/类型定义，不创造无真实第二用途的公共接口。
- 本票通过 Host 显式候选目录/受控 Provider 证明新行为，不改变内置型号、正式支持或生产可用性结论。

## Comments

2026-10-16：用户确认拆分，按已批准的标题、交付和依赖发布；尚无实现或验证执行记录。
