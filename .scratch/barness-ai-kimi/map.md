# Kimi implementation map

## Notes

集成分支：`codex/barness-ai-kimi-integration`。仓库使用本地 Markdown tracker，无 PR 关闭要求。工单详见 [issues](issues/)。协议可在共同基础接口约定下并行开发。

## Decisions so far

- 公共 Client E2E seam 已由 spec 确认。
- [01 身份/目录/成本已完成](issues/01-foundation.md) 与 [05 Host 装配已完成](issues/05-host.md)：六路线四入口、十二调用并发、共享成本与既有 Host/包回归通过；[最终普通回执](evidence/ticket-05.md#final-six-route-ordinary-closure)。
- 普通验证自行执行；live、race、压力遵循 SOP 独立授权。
- [02 Chat 已完成](issues/02-chat.md)：精确请求、工具/原生历史、成本、故障与资源的受影响普通回归通过；[不可变原件索引](evidence/ticket-02.md)。
- [04 Messages 已完成](issues/04-messages.md)：稳定 Messages、Bearer、thinking 来源、成本与资源回归通过；[不可变原件索引](evidence/ticket-04.md)。
- 源码冻结与完整普通验收依赖所有实现及双轴审查修复；阶段性通过不等于六路线真实资格或 release.pass。

## Fog

- 六路线实际资格与晋升依赖自身付费证据，不从研究或回放推定。
