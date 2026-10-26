# 本组合真实资格与捕获回放

Type: task
Status: needs-info
Blocked by: 04, 07, 08, 09, 10

## Contract

[Spec](../spec.md)；[设计](../design.md)；[验证 SOP](../../../docs/agents/verification.md)。

R21/R24；E17。冻结来源后列出精确场景、账户/地区、预算及所有受影响路线，取得本次明确授权才执行付费验证。保留非零 reasoning、续接、工具及图片捕获；先失败后接入回放，源码变化需重新冻结。

## Ownership

本路线真实证据、捕获夹具与资格记录。

## Verification

通过 spec 已确认的公共 Client / Catalog 接缝，先保留失败场景，再实现并保留独立 green attempt；遵循 SOP，不能隐式运行 live/race/压力。当前事实统一见 [spec 当前状态](../spec.md#唯一当前状态)。

## Answer

自身候选真实18/18已通过，24调用/24HTTP、零自动重试、审计0、采集器exit0；[完整索引](../evidence/candidate-live-v3.json)。正式消费者仍exit1/pass=false。剩余脱敏捕获回放、原生资格及其他22受影响路线；未完成正式资格，不关闭工单。

## Comments

- 2026-10-26：由 implement-spec 从已就绪规格建立依赖工单；实现不授予后续人工或付费验证权限。

- 2026-10-26：候选实现、两轴复核及冻结完整普通验收完成；仍需本次账户/地区/费用阈值与明确付费授权。初始可选择xAI自身18场景/24调用；其余22受影响聊天目的路线分别待计划/授权。[实际当前状态](../spec.md#唯一当前状态)；[待授权范围](../next-stage-plan.md)。

- 2026-10-26：两次明确授权均已实际消耗；直连超时后用户另批准代理运行，后者15场景通过、JPEG夹具像素被上游拒绝，已建立10修复。原生消费者均实际exit1、pass=false，未产生正式资格；[失败汇总](../evidence/live-attempts-v1.json)。

- 2026-10-26：计划v3一次明确授权已实际完成并消耗，18场景全部PASS、已知响应估算上界0.031036 USD；用户/工具PNG和JPEG均通过。下一步以完整捕获准备离线回放和原生资格接缝；不自动启动其他付费路线或人工阶段。
