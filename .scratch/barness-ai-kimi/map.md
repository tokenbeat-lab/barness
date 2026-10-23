# Kimi implementation map

## Notes

集成分支：`codex/barness-ai-kimi-integration`。仓库使用本地 Markdown tracker，无 PR 关闭要求。工单详见 [issues](issues/)。协议可在共同基础接口约定下并行开发。

## Decisions so far

- 公共 Client E2E seam 已由 spec 确认。
- [01 身份/目录/成本已完成](issues/01-foundation.md) 与 [05 Host 装配已完成](issues/05-host.md)：六路线四入口、十二调用并发、共享成本与既有 Host/包回归通过；[最终普通回执](evidence/ticket-05.md#final-six-route-ordinary-closure)。
- 普通验证自行执行；live、race、压力遵循 SOP 独立授权。
- [02 Chat 已完成](issues/02-chat.md)：精确请求、工具/原生历史、成本、故障与资源的受影响普通回归通过；[不可变原件索引](evidence/ticket-02.md)。
- [04 Messages 已完成](issues/04-messages.md)：稳定 Messages、Bearer、thinking 来源、成本与资源回归通过；[不可变原件索引](evidence/ticket-04.md)。
- [06 资格基础设施已完成](issues/06-qualification.md)：六路线真实受控进程与只读 replay/成本/晋升边界通过；[原件与索引](evidence/ticket-06.md)。正式目录保持 inactive。
- 源码冻结与完整普通验收依赖所有实现及双轴审查修复；阶段性通过不等于六路线真实资格或 release.pass。

- [03 Responses 已完成](issues/03-responses.md)：原生 summary/opaque IDs、终态、成本与历史碰撞由自身公共 E2E 验证。
- [07 普通验收/双轴修复已完成](issues/07-acceptance.md)：六路线候选批准后的首个真实用例失败，按范围停止；只在 spec 唯一当前状态维护实际阶段事实。

- [08 真实计费报告缺项](issues/08-live-usage-reporting.md)：用户已确认cached_tokens可选默认0，并选择保守费用区间；08已resolved，实现及完整普通验收14045必需PASS、两轴0发现，保留旧失败事实。六路线实际资格与晋升仍依赖自身完整付费证据，不从研究或回放推定。

## Fog

- [07 自身真实资格与人工阶段](issues/07-acceptance.md)：修订后六路线候选具体范围待本次批准；正式资格、晋升与后续人工阶段仍依赖自身完整证据，不从回放推定。
