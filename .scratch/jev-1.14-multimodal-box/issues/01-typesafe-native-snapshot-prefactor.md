# 01：先重构 TypeSafe 原生快照，保持文本行为

Status: resolved

**What to build:** Host 的既有 Jev 文本请求、合法 Hook 换题和混合答案通过独立 native 快照运行，为新增图片保留清晰边界。

**Blocked by:** None (can start immediately)。

## Acceptance criteria

- [x] 在修改实现前列出本票所有失败方式，先经 Client.Classify / HookedClient.Classify 冻结当前成功与失败行为并确认回归基线，再重构；行为等价重构允许回归先绿，不为制造红灯改变已有语义，也不在实现后补单元测试。
- [x] 先冻结当前文本、混合答案、宽 JSON、单候选、精确数字、有界评分舍入及错误优先级的公开回归；不把本票改造成新增图片。
- [x] 最终 TypeSafe 请求及答案验证依据独立拥有 native 数据，移除回解为公共请求的内部路径；不留 shim、双 guard 或动态注册。
- [x] Hook 合法换题、改名、重排后按最终题集验答案；越权模型/认证/目标及不合法原生字段仍拒绝，精确字段名与重复键规则保持。
- [x] 单题损坏清空全部答案，合法 usage/ResponseModel 按既有语义保留；2xx 后不重试，取消和 body/许可释放保持。
- [x] Jev 1.13 文本 wire、目录内容/Hash 和冻结 Pi 基线保持；新分层不反向依赖 Host 或 Agent。
- [x] 结束时产出可重复、可核验的图片/fixture hash、输入、最终请求/响应、结果/错误、Attempt/Observer、断言、配置、完整性清单及脱敏审计；仅受控 HTTP 结果不得声明真实服务 PASS。

## Traceability

- 规格要求：R10、R13、R14、R15、R16。
- 验收组：T10、T13、T14、T16。
- 用户故事：30、32、33、41、42、47。

## Scope and hand-off

- 本工单的拆分已获用户确认并发布；执行时须先满足全部阻塞依赖。本次发布不代表已实现、已验证或取得部署授权。
- 单票只声明其已经打通的能力，其余新能力仍明确拒绝；授权/容量/Hook 最终复核随功能同时交付，不推迟到整合票。
- 复用现有运行时、测试记录器、图片和 JSON 能力；新增依赖前核查已有依赖/类型定义，不创造无真实第二用途的公共接口。
- 本票是有公开行为回归的先行重构；代码机械重组直接删除旧内部分支，不使用 expand/contract 兼容层。

## Comments

2026-10-16：用户确认拆分，按已批准的标题、交付和依赖发布；尚无实现或验证执行记录。发布时明确先行等价重构以既有公开行为基线为准，避免为制造红灯改变功能。

2026-10-16：已完成本票文本等价重构。起点 `6b893c4`，生产实现 `d918842`。
先记录[失败清单](../native-snapshot-failures.md)，在生产代码未改动时冻结 702 PASS 的公开回归
和 12 PASS 的 Jev/Pi 差分（另 180 NOT_RUN）；32 个补充 E2E 旧/新实现均通过，没有实现后补单元测试。
独立 native 请求/题目拥有 State、criteria 与最终题集，删除 native→公共请求及 public decoder 的 native 模式；
canonical 题目字节预算、错误优先级、精确数字、舍入和原子答案语义保持，ADR-0021 已同步。

最终完整套件执行一次，7,468 PASS / 187 NOT_RUN，861 个固定 Pi 差分全部 PASS；相关 race
448 PASS / 178 NOT_RUN，普通/live vet、Node 检查及历史固定包验证均通过。Standards 与 Spec
两轴独立审查均 0 项发现。完整测试后的证据汇总误读嵌套示例 manifest 曾退出 1，`90157fa`
修复汇总范围后复用原产物完成审计/封存；该失败保留在证据中，完整测试没有重跑。

目录仍为 `2026-10-13.1` / `sha256:faae843ac61fea6e3ae59bb75165dc97fcf6aa409dabfe0a2f7c9dccbba12029`；
目录源码、TypeSafe fixtures、支持矩阵、冻结 Pi 和历史 Jev 包逐文件 Hash 未变。
[证据、复现与完整性核验](../native-snapshot-evidence/README.md)包含请求/响应、结果/错误、Attempt/Observer、
配置源码/Hash、资源断言、清单及零发现脱敏审计。本票无图片输入，图片 Hash 项不适用，文本 fixtures
各自保存 SHA-256。Live 为 NOT_RUN；仅声明受控 HTTP 文本行为等价，图片/Box 与真实效果仍由后续票验收。

