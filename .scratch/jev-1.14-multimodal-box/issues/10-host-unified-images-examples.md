# 10：提供 Host 统一图片输入与结果处理示例

Status: resolved

**What to build:** Host 可运行例子切换两协议、处理视觉图例和 Box，并在受控环境重放成功、拒绝及撤销流程。

**Blocked by:** 07：接通 Jev 视觉判定标准并还原公共图例；08：通过 Classify 获取 Box 与独立存在性答案。

## Acceptance criteria

- [x] 在修改实现前列出本票所有失败方式，先写经 Client.Classify / HookedClient.Classify 的公开 E2E 并确认红灯，再实施；不在实现后补单元测试。
- [x] 提供可执行 Host 示例：同一 PNG/JPEG context 公共输入仅替换 Target 与对应 Options 即调用 Jev/Decisions；不要求宿主填写 assets/private key/native URL。
- [x] 示范显式复用/排序、State/说明/标准图片；Decisions 无法表达的局部输入通过预期的拒绝反馈展示，不自动改写含义。
- [x] Host 正确解释最终 ImageInputs 和图例引用，处理 Hook 换图、索引变化、nullable confidence、refusal 和独立存在性答案。
- [x] 坐标示例按实际图片方向/尺寸/裁剪转换，说明多图目标关联是 Host 责任；不把示例业务阈值设成库默认。
- [x] 默认运行通过可控 Provider 和合法合成 fixture 重放，不需要真实凭据；真实模式显式 opt-in 且使用 Host 注入配置，不自动寻找其他 key。
- [x] 同步公共使用说明、能力/预算/撤销与新结果字段，术语和拟议 ADR 状态保持一致；不修改父规格或宣称已生产可用。
- [x] 产出可复放完整输入、请求/响应、结果、错误、图像/hash、断言与审计清单，证明例子实际跨 Classify 运行。
- [x] 结束时产出可重复、可核验的图片/fixture hash、输入、最终请求/响应、结果/错误、Attempt/Observer、断言、配置、完整性清单及脱敏审计；仅受控 HTTP 结果不得声明真实服务 PASS。

## Traceability

- 规格要求：R01、R02、R03、R05、R07、R11、R12、R15、R16、R17。
- 验收组：T01、T02、T03、T06、T09、T11、T12、T16。
- 用户故事：01、02、03、04、05、07、08、09、10、11、17、18、19、21、22、25、34、35、36、37、38、39、40、47、48、50。

## Scope and hand-off

- 本工单的拆分已获用户确认并发布；执行时须先满足全部阻塞依赖。本次发布不代表已实现、已验证或取得部署授权。
- 单票只声明其已经打通的能力，其余新能力仍明确拒绝；授权/容量/Hook 最终复核随功能同时交付，不推迟到整合票。
- 复用现有运行时、测试记录器、图片和 JSON 能力；新增依赖前核查已有依赖/类型定义，不创造无真实第二用途的公共接口。

## Comments

2026-10-16：用户确认拆分，按已批准的标题、交付和依赖发布；尚无实现或验证执行记录。


2026-10-16：依赖 07/08 已 resolved。实现可执行
[Host 统一图片示例](../../../ai/examples/classifierimages/README.md)：同一 PNG/JPEG 公共输入
只替换 Target/Options 接通 Jev 候选与 Decisions；默认 48 场景跨公开 Classify/HookedClient.Classify
与受控 HTTP 重放，覆盖显式复用/排序、局部图拒绝、最终身份与图例、Box null/零置信度、
独立存在性、真实旋转/裁剪坐标逆变换、Hook 换图、refusal、撤销、预算和取消。
真实模式须显式 Host 配置与命名 key 来源，无自动发现或 fallback。

[失败清单](../10-failure-modes.md)与公开 E2E 先于实现，保留缺失命令首红和跨 Classify
执行后的图例视图行为红；没有实现后补单元测试。实现 `773588d`；
[双轴审查](../10-code-review.md)发现并在 `5d2d6ac` 修正一处视图共享图例 map/字节，
Standards 复审确认无剩余发现，Spec 0。

最终 `go test ./... -count=1` exit 0：7,890 PASS / 1,050 显式 opt-in NOT_RUN；
本票普通及实际示例进程 race 各 52/52 PASS，go vet 通过。
[验证与重放](../10-verification.md)和[完整性清单](../10-evidence.json)保留完整输入/配置、
最终请求/响应、结果/错误、Attempt/Observer、资源归零、原图/处理图/hash、断言与脱敏审计，
两份最终父包与两份嵌套 Host 包均零发现。受控 PASS 不声明真实协议、视觉质量或生产可用；
内置目录/支持矩阵/冻结 Pi、父规格和 proposed ADR 状态未改。
