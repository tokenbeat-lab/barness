# 07：接通 Jev 视觉判定标准并还原公共图例

Status: claimed

**What to build:** Host 用图片定义 Choice、Bool 或 Score 标准，Score 返回可解释的公共图例引用，完全不依赖 Jev 私有 key。

**Blocked by:** 05：在 Jev State 原位置使用图片引用。

## Acceptance criteria

- [ ] 在修改实现前列出本票所有失败方式，先写经 Client.Classify / HookedClient.Classify 的公开 E2E 并确认红灯，再实施；不在实现后补单元测试。
- [ ] 贯通 Choice 单候选说明、Bool true/false 说明和 Score 各等级内的图片节点，保持标准集合和等级次序；三种全部完成后声明 criteria 能力。
- [ ] Choice 未引用 null 保持无说明含义，Score/Bool 只允许明确的图片 null 槽位；容器、题型字段与候选/等级名称不能成为目标。
- [ ] 图片留在本题本标准的值位置；独立计 source 与实际出现预算，共享素材不增加本来不应接收图片的题目作用域。
- [ ] native legend 先严格符合最终 frozen Criteria，再把图片还原为 null 与 LegendImageReferences；Pointer 相对 Legend 根，索引指向最终 ImageInputs。
- [ ] Hook 换等级/改名/换图后结果引用仍对应最终 MIME/hash；没有私有 key/图片 bytes 泄漏，不推断入口索引或 Provider 资源归属。
- [ ] 图例少项/错项/重复/改图/分布损坏导致整次 Answers 与 ImageInputs 不发布，同时保留唯一合法 usage。
- [ ] 无图图例、Jev 有界舍入、宽 JSON、Choice 单候选及 Decisions score/拒答规则保持；Decisions 对任何局部标准图片凭据前拒绝。
- [ ] 结束时产出可重复、可核验的图片/fixture hash、输入、最终请求/响应、结果/错误、Attempt/Observer、断言、配置、完整性清单及脱敏审计；仅受控 HTTP 结果不得声明真实服务 PASS。

## Traceability

- 规格要求：R03、R04、R06、R08、R10、R11、R14、R15。
- 验收组：T03、T04、T05、T06、T08、T10、T11、T13。
- 用户故事：09、10、11、12、13、14、15、17、18、26、27、30、39、40、41、42、44。

## Scope and hand-off

- 本工单的拆分已获用户确认并发布；执行时须先满足全部阻塞依赖。本次发布不代表已实现、已验证或取得部署授权。
- 单票只声明其已经打通的能力，其余新能力仍明确拒绝；授权/容量/Hook 最终复核随功能同时交付，不推迟到整合票。
- 复用现有运行时、测试记录器、图片和 JSON 能力；新增依赖前核查已有依赖/类型定义，不创造无真实第二用途的公共接口。
- 本票通过 Host 显式候选目录/受控 Provider 证明新行为，不改变内置型号、正式支持或生产可用性结论。

## Comments

2026-10-16：用户确认拆分，按已批准的标题、交付和依赖发布；尚无实现或验证执行记录。

