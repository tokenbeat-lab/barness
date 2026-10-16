# 07：接通 Jev 视觉判定标准并还原公共图例

Status: resolved

**What to build:** Host 用图片定义 Choice、Bool 或 Score 标准，Score 返回可解释的公共图例引用，完全不依赖 Jev 私有 key。

**Blocked by:** 05：在 Jev State 原位置使用图片引用。

## Acceptance criteria

- [x] 在修改实现前列出本票所有失败方式，先写经 Client.Classify / HookedClient.Classify 的公开 E2E 并确认红灯，再实施；不在实现后补单元测试。
- [x] 贯通 Choice 单候选说明、Bool true/false 说明和 Score 各等级内的图片节点，保持标准集合和等级次序；三种全部完成后声明 criteria 能力。
- [x] Choice 未引用 null 保持无说明含义，Score/Bool 只允许明确的图片 null 槽位；容器、题型字段与候选/等级名称不能成为目标。
- [x] 图片留在本题本标准的值位置；独立计 source 与实际出现预算，共享素材不增加本来不应接收图片的题目作用域。
- [x] native legend 先严格符合最终 frozen Criteria，再把图片还原为 null 与 LegendImageReferences；Pointer 相对 Legend 根，索引指向最终 ImageInputs。
- [x] Hook 换等级/改名/换图后结果引用仍对应最终 MIME/hash；没有私有 key/图片 bytes 泄漏，不推断入口索引或 Provider 资源归属。
- [x] 图例少项/错项/重复/改图/分布损坏导致整次 Answers 与 ImageInputs 不发布，同时保留唯一合法 usage。
- [x] 无图图例、Jev 有界舍入、宽 JSON、Choice 单候选及 Decisions score/拒答规则保持；Decisions 对任何局部标准图片凭据前拒绝。
- [x] 结束时产出可重复、可核验的图片/fixture hash、输入、最终请求/响应、结果/错误、Attempt/Observer、断言、配置、完整性清单及脱敏审计；仅受控 HTTP 结果不得声明真实服务 PASS。

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


2026-10-16：依赖 05 已 resolved，本票实现 `53edb35`。Host 候选 Jev 1.14
贯通 Choice 单候选、Bool true/false 与 Score 有序等级的根/嵌套图片；投影保持本题
本标准位置，来源与实际出现预算独立。最终 native Hook guard 重建实际资产与索引，
Score Legend 先严格匹配最终 Criteria，再还原 null 与相对 Legend 的公共引用。
合法改名/换等级/换 MIME 与图片均以最终 hash 为准；坏图例/分布整体不发布结果，
唯一合法 usage 保留，2xx 不重放。Decisions 局部图片仍在凭据前整体拒绝。

[失败清单](../07-failure-modes.md)先于实现；公开 Classify/HookedClient.Classify
红灯后实施，无实现后补单元测试。[独立双轴审查](../07-code-review.md) Standards 0、
Spec 0。完整 `go test ./... -count=1` exit 0：7,401 PASS / 1,050 显式 opt-in
NOT_RUN；相关 race 934/934 PASS；本票 184 场景在两次最终运行均 PASS；go vet 通过。
两份脱敏审计零发现，[完整性清单](../07-evidence.json)核验 66,881 产物 hash、
最终图片身份/图例引用、原子失败、共享观测与资源释放。[验证与重放](../07-verification.md)
保留 fixture/图片 hash、输入、最终请求/响应、结果/错误、Attempt/Observer、断言及配置。
真实服务/视觉质量 NOT_RUN；内置目录、支持矩阵、冻结 Pi 和正式支持结论未改。
