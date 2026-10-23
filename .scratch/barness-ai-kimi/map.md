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

- [本次候选结果02](evidence/candidate-live-result-02.md)：31调用/31Attempt、零重试；中国站三路线通过。正式发布消费者曾误用于candidate，原拒绝及阶段纠正均保留。

- [60秒续验结果04](evidence/candidate-live-result-04.md)：国际站Responses/Messages各9/9及原生候选资格PASS，连同同源Chat，当前国际三路线候选完整通过。18调用均200，未触发额外重试；09/10 resolved。新输入完整普通14071必需PASS，实际状态仍只维护spec。

## Fog

- 新来源中国站三路线候选、六路线formal/晋升及人工阶段仍待07；11嵌套SSE错误诊断待triage。09/10已按60秒续验完整通过，旧失败事实保留。

- 2026-10-23：用户要求延长间隔并继续，09认领Kimi grant有界间隔实现与公开进程验证；新的国际三路线固定35秒、零重试，旧来源/授权原件保留。当前实际进度仅见spec。

- [35秒续验结果03](evidence/candidate-live-result-03.md)：Chat9/9及原生资格PASS；Responses第4项SSE引擎过载，13调用后停止。09实现与普通已通过，剩余资格由[10](issues/10-responses-engine-overload.md)待新范围；[11](issues/11-responses-stream-error-diagnostics.md)仅记录安全错误诊断待核对。实际状态仍只维护spec。
