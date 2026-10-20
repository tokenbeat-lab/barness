# 05 整合验证与迁移交付

Status: resolved
Type: task
Blocked by: 01, 02, 03, 04, 06

## Context

[规格](../spec.md)、[设计](../design.md)、[验证策略](../../../docs/agents/verification.md)。需求：R01–R20，RP01–RP20。

## Ownership

普通 E2E、适用旧服务回归、双 vet、关键边界审查和双轴 code-review；补齐迁移/回滚与当前状态。付费/race 无授权保持 NOT_RUN，正式验收保持开放。

## Acceptance

先在规格指定的公共 Client / Host / CLI 接缝建立失败场景，再实现；保留实际命令、退出码、时间、执行来源、工件及 Hash。先读取相关 ADR 和验证策略。不读取或输出真实秘密；不调用付费服务，不运行 race。按实际证明关闭实现，不用普通 replay 代替自身 live。

## Comments

2026-10-20：由 implement-spec 根据已提供规格建立；无 GitHub PR 关闭义务。各票完成后在 Answer 留实现/验证指针，整体状态唯一维护于 spec。

2026-10-20：主代理已留存旧服务普通基线：`.evidence/provider-sites/baseline/receipt.json`（exit 0；3376 场景：3367 PASS / 9 NOT_RUN；未启用差分/压力）。此票待前置实现合并后完成最终整合，不以该基线代替新功能结果。


## Answer

2026-10-20：工单 01–04、06 合并并完成关键边界整合；最终普通受测提交 `d4b63d9b2ecec3dc7fcbdadf3b849e31c83cd92e`。双轴原审查、修复、独立复审及旧期望迁移补审均保留于 [review](../evidence/review.md)；全部修复交同一 implementer，不把不同轴的发现分头改写。

完整普通显式打开差分与压力，实际 exit 0、墙钟 661.883s，主 E2E 原生 13,244 / 13,244 PASS；889 份差分记录 pending 0，32 份原生 audit 零发现。双 vet、CLI 构建、受控自身证据消费及旧 HTTP 工件审计通过；旧 wire 审计 3,077 / 3,077 PASS，仅应用 ADR-0049 的 v2 派生键迁移。依赖、冻结上游、旧支持行与 394 追溯项保持不变；新增 40 项不删旧 mandatory。实际数字、时间、来源清单、命令、失败原件与输入见 [最终验证索引](../evidence/verification.md) / [机器索引](../evidence/index.json)。

迁移、命中率影响、凭据撤销与回滚说明见 [地区指南](../../../docs/barness-ai/regional-providers.md)。历史证明不重导出、不刷新；精确来源拒绝仍由原生消费者判断。完整普通第一次失败及收集器错误原样保留，修订活动 golden 后才取得本次全量通过。

按验证策略，本票交付实现与普通结果；正式前置因缺自身资格和所选旧 live / 效果输入实际 exit 1，未启动正式后续命令。完整 feature 的唯一当前状态仍见 [spec](../spec.md#当前状态)，不随本票解决而宣称发布通过。人工 race 命令、完整范围、固定来源与预计耗时已在最终索引交接；本次未执行 race，未读取真实秘密或调用厂商。
