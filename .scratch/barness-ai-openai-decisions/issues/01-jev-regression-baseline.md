# 01: 固定 Jev 公共行为与可重放回归基线

**What to build:** 维护者在新增协议前能通过公共 Classify 重放现有 Jev 的请求、结果、错误顺序及资源释放，并得到后续切片可比较的固定证据。这是前置行为基线；不引入推测性 Adapter 抽象。

**Blocked by:** None (can start immediately)

**Status:** resolved
**Type:** task
**Source tickets:** 01
**Requirements:** R02, R14, R15, R19
**Acceptance cases:** OD01, OD03, OD09, OD16, OD34, OD50
**Updated:** 2026-10-12

## Responsibility

负责 Jev 公共回归场景、受控 Provider 脚本及基线证据。当前唯一 oracle 为 Pi 1.1.0；核对实际 HEAD 与设计基线，记录偏移原因，不把历史 PASS 当本次执行。协作时保留他人改动。

## Failure scenarios first

- 单候选、宽 JSON 说明、空状态、可选 bool criteria 被新增限制收窄。
- 公共 choice 256 / score 11 / 坏 state 的拒绝延后到 Binding 或凭据解析。
- Jev 精确字段、概率并列、舍入和原数字精度改变，Hook 换题改为验证入口旧题。
- 历史目录、fixture、oracle 或报告被重新生成后冒充原基线。

## Acceptance criteria

- [x] 在生产变更前运行已有 Jev、目录、回调、unary 和官方 wire replay；补充断言先验证现状，记录已有失败。
- [x] 固定单候选、结构化说明、bool、score 2–10、duplicate/case-alias 拒绝及有界舍入的合法与非法对照。
- [x] 拒绝顺序同时断言 Code/Phase 与 resolver/credential/Attempt/request 次数；scope 拒绝不访问 resolver。
- [x] 固定原 wire、usage、ResponseModel、归属、Hook 次数和 body/permit 释放，后续断言不得删除或收窄。
- [x] 验证唯一 Pi 1.1.0 pins、来源及模型数据 Hash；已有差分或 gate 缺陷保留真实状态并关联后续责任。
- [x] 产出含 fixture、捕获、结果、Attempt、Observer、assertions、版本、SHA-256 manifest、审计和准确重放命令的独立基线包。

## Completion

仅完成基线核验，不以修改生产代码使本票全绿。保留当前失败与原因；存在阻断后续行为比较的问题时明确列出，不能用该基线宣称 Decisions 已支持。

## Comments

2026-10-12：用户批准 13 张草案并替换旧票。此票按新编号发布；下列原票评论按来源保留，ready-for-agent 不表示已实施或依赖已完成。

### 原票 01

2026-10-12：计划创建，尚未执行基线或修改生产实现。

### 本票实施（2026-10-12）

实际 HEAD 与设计 Baseline 同为 `8ea61b8d70de6c84fe27be34dc7d21c5c40841d9`，原源码/fixture/oracle/历史证据 Hash 无变化。既有回放 542 PASS、2 NOT_RUN；显式 Jev 差分 12 PASS、pending 0；新增公共对照 98 PASS，artifact 验证先红后绿并拒绝缺包/篡改/重写。完整开启 Pi 的测试套件退出 0，6,227 E2E case：6,218 PASS、9 NOT_RUN，815 差分均 PASS；相关 race 394 PASS，两种 vet 与整包脱敏审计通过。

独立包、版本、原始捕获/判定、Hash 和准确命令见[基线说明](../jev-baseline/README.md)及[最终验证收据](../jev-baseline/validation.json)。历史源码重建仅核验原证明，没有冒称本次重建；历史 release/live/effect FAIL 保持，[压力缺陷](../../barness-ai-thinking-sampling/issues/07-pressure-stream-integrity.md)仍由原票负责。本次没有阻断 Jev 离线比较的问题；不表示 Decisions 已实现或发布 gate 已通过。后续工单不得删除、收窄本包及既有 Jev 断言。
