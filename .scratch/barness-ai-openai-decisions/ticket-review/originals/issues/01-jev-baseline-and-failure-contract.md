# 01 — 固定 Jev 行为基线与公共失败合同

Status: ready-for-agent
Type: task
Depends on: none
Updated: 2026-10-12

## What to build

在修改生产实现前固定现有 System One/Jev 的可重放公共行为，建立 Decisions 的失败目录和追溯。依据 [R02](../spec.md)、[设计 §9](../design.md)、[OD01/03/09/16/34/50](../acceptance.md)。

## Ownership and context

负责现有 Jev E2E fixture、catalog/default、strict envelope、native JSON、舍入、callbacks/lifecycle/diff 的基线证据，以及新增 public regression case。生产 Adapter、正式目录、历史证据不修改。协作时不要回退他人编辑；用现有 Classify / HookedClient.Classify Seam。

## Failure scenarios first

- 合法单候选、结构化说明、bool optional criteria 被新的统一约束拒绝。
- score 11 / choice 256 / malformed state 从 scope 移到 capability，提前访问 resolver 或 Credential。
- 舍入接受范围、duplicate 拒绝、最高概率并列、case-alias 或 raw JSON 精度发生改变。
- 回调验证入口旧题、重试重读 Host、usage 在结果错误后丢失。
- 原目录/Binding 省略新增字段时不能构造，或将旧 hash/来源暗改成新证据。

## Acceptance criteria

- [ ] 固定基线 commit、源码/fixture/hash、原公共契约和当前实际状态，不以历史 PASS 代替本次执行。
- [ ] 保留并运行既有 P07/相关 unary E2E；若增加旧行为断言，先证明现状，保存失败历史。
- [ ] 新 regression 比较 Code/Phase、resolver/credential/Attempt/request 数、原 wire、用量、舍入、终态和资源。
- [ ] 为 OD01/03/09/16/34/50 提供稳定映射，后续工单不得删除或改成 UNSUPPORTED。
- [ ] 确认冻结 oracle 已是唯一 Pi 1.1.0，不重复迁移或重新在线生成模型数据。

## Verification and handoff

运行目标 Jev、Catalog、unary 生命周期与固定差分；记录准确命令、raw fixture、请求/响应、结果、观测、assertions、manifest/audit。红例区分现有缺陷和新功能未实现，不修改生产代码来使本工单“全绿”。交付可供 02–09 对照的基线指针。

## Comments

2026-10-12：计划创建，尚未执行基线或修改生产实现。
