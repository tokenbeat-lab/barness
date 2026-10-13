# 09 — 十组合发布合同、文档与回滚

Status: ready-for-agent
Type: task
Depends on: 08
Updated: 2026-10-12

## What to build

把新 route 的 P10、来源/能力/价格、差分、资源和效果纳入固定发布合同，保留九旧组合与 Jev 要求；更新公共文档、ADR 交付记录与独立回滚验证。依据 [R17/R19/R20](../spec.md)、[OD53/57/58](../acceptance.md)。

## Ownership and context

负责 releasegate/traceability/catalog snapshot/matrix/evaluation 输入核验、public README/contract/differences 和 ADR 引用。文档描述最终实现和实际证据；不改写历史 PASS、不自动发布、不回退他人改动。

## Failure scenarios first

- gate 仍只要求九组合或只收 Jev 效果，新条目只写映射没有实际证据。
- 删 P10/cases/能力/某旧 route 能让 gate 绿；candidate/旧 hash/缩小 expected 也通过。
- source/hash/assertion/audit 缺失，NOT_RUN/FAIL 被当 PASS；当前未解决的旧 gate 缺口被隐藏。
- 禁用新 Binding/图片权限后绕道 Jev/chat，活动调用重读新配置或用量丢失。
- 文档和 status 声称已支持，而实际分支/field/错误契约或真实结果不一致。

## Acceptance criteria

- [ ] 先写 gate/public artifact 的删项/错 hash/错 route/预算/报告反例和 rollback E2E，再改合同。
- [ ] 固定十组合、P10 必交和所有 R01–R20 追溯由 actual artifacts 判定；旧 Jev/P07 等条目不删除。
- [ ] gate 接收并按身份独立核验 Jev 与 Luna 效果，保留既有单 Jev 的历史核验路径；新发布必须两份当前适用证据。
- [ ] 正常/race/vet/live-tag-vet/固定差分/压力和十组合真实报告/audit 适用性满足，先前未完成项照实呈现。
- [ ] 文档说明公共 score2–10、raw JSON格式、图片默认/三条件、refusal、presence/cost、回调、错误、迁移和 rollback。
- [ ] ADR-0026 只有实际实施与证据成立才 accepted；旧 ADR 追加关联/交付，不覆写 Jev 历史。
- [ ] rollback 禁 Binding/flag/型号授权停止新调用，已 pin 调用仍按 snapshot；不自动 fallback，旧 Jev/chats 继续可用。
- [ ] 正式目录/价格版本、哈希、固定 pins、manifest 与 source refs 完整，所有缺证据项阻止发布完成。

## Verification and handoff

运行现有完整 releasegate Interface；如为双效果报告增加明确参数，先红后绿并记录真实命令，不能靠注释宣称 gate 已检查。打包 release-report/json/md、fixed contract、traceability、snapshots、matrix、全部证据与脱敏审计。最终交付列实际 PASS/FAIL/NOT_RUN 与回滚结果；无完整 gate 不标本功能完成。

## Comments

2026-10-12：计划 ready-for-agent 不等于已发布；本任务只写设计与实施工单。
