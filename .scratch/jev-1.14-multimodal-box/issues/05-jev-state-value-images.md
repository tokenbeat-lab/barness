# 05：在 Jev State 原位置使用图片引用

Status: resolved

**What to build:** Host 将图片放入结构化 State 的明确值位置，保留业务结构、精确数值和共享作用域。

**Blocked by:** 04：接通 Jev JSON 上下文图片闭环。

## Acceptance criteria

- [x] 在修改实现前列出本票所有失败方式，先写经 Client.Classify / HookedClient.Classify 的公开 E2E 并确认红灯，再实施；不在实现后补单元测试。
- [x] 接通 State value 的成功链路，包括嵌套 object/array 及明确引用的 State 根图片；更新真实位置能力而非只放宽目录。
- [x] 仅已解析 null 槽位成为图片，未引用 null 与普通字符串/对象键不变；图片复用仅填对应私有 key，数字拼写与候选/问题名不变。
- [x] 无 context 不包裹 State；混合 context/value 时在包裹前解释公共 Pointer，源对象字段不被覆盖，顺序与可见范围明确。
- [x] 重复 value 目标、根/容器误用、坏 RFC escape、数组非法下标、未使用来源与不存在目标按 scope 合同拒绝；不能凭模板外观识别图片。
- [x] Jev 共享 State 图片不因问题数重复计出现；source、引用展开、State bytes 与最终 native body 都执行各自预算。
- [x] 最终 Hook guard 扫描实际图片值和用途，重建 ImageInputs；substring、对象键/普通文本中的 private key 和未使用资产拒绝。
- [x] Decisions 对同一个合法 State value 请求仍在 capability 整体拒绝；未实现的题内位置不被元数据误宣称支持。
- [x] 结束时产出可重复、可核验的图片/fixture hash、输入、最终请求/响应、结果/错误、Attempt/Observer、断言、配置、完整性清单及脱敏审计；仅受控 HTTP 结果不得声明真实服务 PASS。

## Traceability

- 规格要求：R01、R02、R03、R04、R06、R08、R10、R11、R14。
- 验收组：T03、T04、T05、T06、T07、T08、T10、T11。
- 用户故事：05、07、12、13、14、15、16、17、18、19、26、27、30、40。

## Scope and hand-off

- 本工单的拆分已获用户确认并发布；执行时须先满足全部阻塞依赖。本次发布不代表已实现、已验证或取得部署授权。
- 单票只声明其已经打通的能力，其余新能力仍明确拒绝；授权/容量/Hook 最终复核随功能同时交付，不推迟到整合票。
- 复用现有运行时、测试记录器、图片和 JSON 能力；新增依赖前核查已有依赖/类型定义，不创造无真实第二用途的公共接口。
- 本票通过 Host 显式候选目录/受控 Provider 证明新行为，不改变内置型号、正式支持或生产可用性结论。

## Comments

2026-10-16：用户确认拆分，按已批准的标题、交付和依赖发布；尚无实现或验证执行记录。

2026-10-16：阻塞工单 04 已 resolved，完成 State value 图片闭环。当前 main 上实现
`3a6fb7b`、fixture 补全 `a79cbe3`、审查修正 `6793760`。无 context 保持业务 State
位置，混合引用先解析公共 Pointer 后包裹；根图片、嵌套 object/array、转义/数字/空键、
普通 null/模板文本和精确数字均有公共 E2E。最终 native 扫描按实际全值资产引用重建用途、
出现预算和 ImageInputs；未开放题内位置/Box，Decisions 仍在 capability 整体拒绝 value。

先列[失败方式](../05-failure-modes.md)，公开 Classify E2E 红灯后逐片实施，无实现后补
单元测试。独立双轴审查 Standards 0；Spec 发现 1 个 P2：从无图请求 Hook 加图时混淆
state/images 业务对象，先加两条红灯再修复；最终两轴均 0 遗留，见[审查](../05-code-review.md)。

最终完整套件 exit 0：7,067 PASS / 1,050 显式 opt-in NOT_RUN；相关 race 592/592 PASS；
本票 129 场景在全量和 race 中均 PASS；go vet 通过。两份脱敏审计均零发现，完整性
清单核验 57,291 个产物 hash、最终 decoded 文件身份、失败原子性、共享观测、资源释放
及源文件清单。[验证与重放](../05-verification.md)、[完整性清单](../05-evidence.json)
保留合成图片/fixture hash、输入、最终请求/响应、结果/错误、Attempt/Observer、配置和断言。
真实服务与视觉质量 **NOT_RUN**；未改内置目录、支持矩阵、冻结 Pi 或正式支持结论。

