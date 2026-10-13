# 01: 固定 Jev 公共行为与可重放回归基线

**Status:** needs-triage
**Publication:** draft，待用户批准后转 ready-for-agent；尚未发布到 issues。
**Blocked by:** None (can start immediately)
**Source tickets:** 01
**Requirements:** R02, R14, R15, R19
**Acceptance cases:** OD01, OD03, OD09, OD16, OD34, OD50
**Updated:** 2026-10-12

**What to build:** 维护者在新增协议前能通过公共 Classify 重放现有 Jev 的请求、结果、错误顺序及资源释放，并得到后续切片可比较的固定证据。这是前置行为基线；不引入推测性 Adapter 抽象。

## Responsibility

负责 Jev 公共回归场景、受控 Provider 脚本及基线证据。当前唯一 oracle 为 Pi 1.1.0；核对实际 HEAD 与设计基线，记录偏移原因，不把历史 PASS 当本次执行。协作时保留他人改动。

## Failure scenarios first

- 单候选、宽 JSON 说明、空状态、可选 bool criteria 被新增限制收窄。
- 公共 choice 256 / score 11 / 坏 state 的拒绝延后到 Binding 或凭据解析。
- Jev 精确字段、概率并列、舍入和原数字精度改变，Hook 换题改为验证入口旧题。
- 历史目录、fixture、oracle 或报告被重新生成后冒充原基线。

## Acceptance criteria

- [ ] 在生产变更前运行已有 Jev、目录、回调、unary 和官方 wire replay；补充断言先验证现状，记录已有失败。
- [ ] 固定单候选、结构化说明、bool、score 2–10、duplicate/case-alias 拒绝及有界舍入的合法与非法对照。
- [ ] 拒绝顺序同时断言 Code/Phase 与 resolver/credential/Attempt/request 次数；scope 拒绝不访问 resolver。
- [ ] 固定原 wire、usage、ResponseModel、归属、Hook 次数和 body/permit 释放，后续断言不得删除或收窄。
- [ ] 验证唯一 Pi 1.1.0 pins、来源及模型数据 Hash；已有差分或 gate 缺陷保留真实状态并关联后续责任。
- [ ] 产出含 fixture、捕获、结果、Attempt、Observer、assertions、版本、SHA-256 manifest、审计和准确重放命令的独立基线包。

## Completion

仅完成基线核验，不以修改生产代码使本票全绿。保留当前失败与原因；存在阻断后续行为比较的问题时明确列出，不能用该基线宣称 Decisions 已支持。
