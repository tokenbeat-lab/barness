# 04：接通 Jev JSON 上下文图片闭环

Status: resolved

**What to build:** Host 用已有公共图片与 context 引用，在明确选择实验后通过 Jev 判断并获得最终图片元数据。

**Blocked by:** 01：先重构 TypeSafe 原生快照，保持文本行为；03：让 Host 核对 Decisions 最终图片身份。

## Acceptance criteria

- [x] 在修改实现前列出本票所有失败方式，先写经 Client.Classify / HookedClient.Classify 的公开 E2E 并确认红灯，再实施；不在实现后补单元测试。
- [x] 采用显式 Host 候选目录中的固定 Jev 1.14，新增 ExperimentalImages 的实际传递；本票只声明 context，尚不声明局部位置/Box。
- [x] private key 经有界碰撞检查且重试前固定；一个来源一个 native JSON asset，严格 PNG/JPEG format 与 Base64，不接受公共 Assets/data URL。
- [x] context 按引用次序进入 state 包裹，原 State 子树和用户同名 state/images 字段保留；无图文本 wire 不变，不改自然语言路径。
- [x] 图片策略、位置/MIME/数量能力、Binding 图片授权和实验选择缺一即凭据前拒绝；header 重复/矛盾、实验越权及 Long Context 注入拒绝。
- [x] 贯通静态格式、APNG、尺寸/面积、decoded 单文件、key 与整 body 限制；Jev 专属限制不改变 Decisions WebP 和已有边界。
- [x] Hook 后只接受已支持 native JSON 子集：合法资产增删/换图/重排重新冻结，unused/part/URL/substring/混淆 key 拒绝，身份与授权保持固定。
- [x] 最终 Jev ImageInputs 按确定实际使用次序建立，同 asset 复用一项；错误原子性、usage/价格和 Observer 脱敏保持。
- [x] 不改内置目录和支持矩阵，不把受控 Provider 结果标记真实 PASS；与旧 Jev 文本/Decisions 回归一起产证。
- [x] 结束时产出可重复、可核验的图片/fixture hash、输入、最终请求/响应、结果/错误、Attempt/Observer、断言、配置、完整性清单及脱敏审计；仅受控 HTTP 结果不得声明真实服务 PASS。

## Traceability

- 规格要求：R04、R06、R07、R08、R09、R10、R11、R14、R15。
- 验收组：T01、T02、T06、T07、T08、T09、T10、T11、T13。
- 用户故事：01、02、03、04、05、18、19、21、22、23、24、25、26、27、28、30、31、40、42。

## Scope and hand-off

- 本工单的拆分已获用户确认并发布；执行时须先满足全部阻塞依赖。本次发布不代表已实现、已验证或取得部署授权。
- 单票只声明其已经打通的能力，其余新能力仍明确拒绝；授权/容量/Hook 最终复核随功能同时交付，不推迟到整合票。
- 复用现有运行时、测试记录器、图片和 JSON 能力；新增依赖前核查已有依赖/类型定义，不创造无真实第二用途的公共接口。
- 本票通过 Host 显式候选目录/受控 Provider 证明新行为，不改变内置型号、正式支持或生产可用性结论。

## Comments

2026-10-16：用户确认拆分，按已批准的标题、交付和依赖发布；尚无实现或验证执行记录。



2026-10-16：完成固定 Host 候选 `jev-1.14.0` 的实验 JSON context 图片闭环。
实现提交 `5615b02`，设计复核修正 `7bfec2e`、`724a0a7`；所有代码提交到当前 main。
State 独立包裹，原内容/精确数字保留；一个来源一个私有 asset，重试固定 key/bytes/header。
PNG/JPEG/APNG/尺寸/decoded/body/key、来源/引用预算及图片权限/实验选择在正确边界复核；
Hook 采用独立 native freeze，合法省略 asset type、换图/增删/重排及最终身份有公开 E2E。
未开放局部位置、视觉图例或 Box，未改内置目录、正式支持矩阵和冻结 Pi。

先写[失败清单](../04-failure-modes.md)，各切片确认公开 E2E 红灯再改实现，无实现后补单元测试。
独立双轴审查发现 4 项 P2 设计偏差，均先补红灯后修正，最终 Standards / Spec 均 0 未解决项。
完整套件 exit 0：6,938 PASS / 1,050 显式 opt-in NOT_RUN；相关 race 463/463 PASS；
本票 137 场景在两包中全部 PASS；go vet 通过。两份脱敏审计均零发现，独立核验
53,591 个产物 Hash、最终图片身份、失败原子性及源文件清单。

首轮完整验证记录的一项既有 100ms timeout 竞争断言失败完整保留；10 次单项及后续完整
验证均通过，另建[needs-triage 记录](../../unary-timeout-priority/issues/01-unary-timeout-priority.md)
保留调查条件，未在本票扩改共享 timeout 运行时。
[验证、重放与最终 manifest](../04-verification.md)、[审查](../04-code-review.md)、
[完整性清单](../04-evidence.json)收录 fixture/图片 hash、输入、最终请求/响应、结果/错误、
Attempt/Observer、配置、释放断言和审计。真实服务 **NOT_RUN**，仅声明受控 HTTP 协议闭环。
