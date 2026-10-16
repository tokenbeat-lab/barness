# 08：通过 Classify 获取 Box 与独立存在性答案

Status: resolved

**What to build:** Host 同批询问位置和存在性，得到严格类型化的 Box，并能区分 null confidence、零值及缺失字段。

**Blocked by:** 06：在 Jev 单题说明中使用视觉样例。

## Acceptance criteria

- [x] 在修改实现前列出本票所有失败方式，先写经 Client.Classify / HookedClient.Classify 的公开 E2E 并确认红灯，再实施；不在实现后补单元测试。
- [x] 新增 sealed BoxQuestion/BoxAnswer 的公共 JSON、clone、容量/能力检查与独立 native 转换；只对明确支持的 Jev 型号声明 box。
- [x] 四边必须出现、有限、归一化且正面积；confidence 必须出现，null 与数字零区分，缺失/越界/错误类型一律拒绝。
- [x] Box 可与既有三题型同批；每 Box 至少能看到 context/State/自身 Instructions 的图，只有其他题的局部图不满足前提。
- [x] 无可见图的原请求在 scope 拒绝；Hook 后丢失图为 callback_failed/request；图片权限和实验 header 不因 box 被绕过。
- [x] 合法框及低/负存在性答案仍独立返回，库不阈值化、不条件执行、不虚构 not-found/image_id/存在概率；多图关联与像素变换由 Host 处理。
- [x] Hook 在已授权型号支持的题型之间合法变更后按最终题集验证；一个坏框清空全部答案/图片元数据，合法 usage 保留。
- [x] Jev 1.13 或 Decisions 请求 Box 均在 capability 凭据前拒绝，不能改走其他型号或协议；已有普通答案与 refusal 回归保持。
- [x] 结束时产出可重复、可核验的图片/fixture hash、输入、最终请求/响应、结果/错误、Attempt/Observer、断言、配置、完整性清单及脱敏审计；仅受控 HTTP 结果不得声明真实服务 PASS。

## Traceability

- 规格要求：R03、R06、R07、R10、R11、R12、R14、R15。
- 验收组：T03、T06、T09、T10、T11、T12、T13。
- 用户故事：25、30、34、35、36、37、38、40、41、42。

## Scope and hand-off

- 本工单的拆分已获用户确认并发布；执行时须先满足全部阻塞依赖。本次发布不代表已实现、已验证或取得部署授权。
- 单票只声明其已经打通的能力，其余新能力仍明确拒绝；授权/容量/Hook 最终复核随功能同时交付，不推迟到整合票。
- 复用现有运行时、测试记录器、图片和 JSON 能力；新增依赖前核查已有依赖/类型定义，不创造无真实第二用途的公共接口。
- 本票通过 Host 显式候选目录/受控 Provider 证明新行为，不改变内置型号、正式支持或生产可用性结论。

## Comments

2026-10-16：用户确认拆分，按已批准的标题、交付和依赖发布；尚无实现或验证执行记录。



2026-10-16：依赖 06 已 resolved。实现 sealed BoxQuestion/BoxAnswer、必填 nullable
Confidence、严格归一化正面积四边和独立 native 转换；每题必须看到 context/State/自身
Instructions 图片。原请求无可见图在 scope 拒绝，Hook 丢图在 request 拒绝；最终题型
和图片重新验证，保留既有授权、实验 header、预算及原子 usage 语义。存在性 Bool 与
合法框独立返回，Host 决定阈值、搜索图归属和像素变换。仅 Host 显式固定候选
`jev-1.14.0` 可声明 box；内置目录、正式支持和冻结 Pi 未改。

先列[失败方式](../08-failure-modes.md)，公共 JSON E2E 双入口先红，随后写完整 typed
E2E 再实现（typed 编译红单独记录）。实现 `443fc30`；[双轴审查](../08-code-review.md)
Standards 0、Spec 0。全量 exit 0：7,647 PASS / 1,050 显式 opt-in NOT_RUN；相关 race
exit 0：2,436 PASS / 225 opt-in NOT_RUN；本票两份最终包均 246/246 PASS；go vet 通过。

[验证与重放](../08-verification.md)和[完整性清单](../08-evidence.json)独立核验
83,434 个最终产物 hash，保留图片/fixture、输入、最终请求/响应、结果/错误、Attempt/
Observer、断言、配置和资源释放；两份脱敏审计零发现。首个红灯历史另存并核验。
真实服务协议、视觉质量及正式支持仍 NOT_RUN；受控 HTTP PASS 不扩大支持结论。
