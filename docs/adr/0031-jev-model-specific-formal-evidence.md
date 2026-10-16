---
status: accepted
date: 2026-10-17
---

# 固定 Jev 1.14 使用独立正式证据身份

同一 TypeSafe 协议的 1.13 文本报告不能代表 1.14 图片与 Box，因此保留
`typesafe-classifier` 的固定 1.13 validator 和历史行，新增
`typesafe-classifier-jev-1.14.0`。目录 `2026-10-17.1` 保留旧型号并新增四位置
PNG/JPEG 与 Box；图片仍需实验选择、独立策略和 Binding 授权，moving alias 不纳入。

采用 schema 3 已有的逐路线报告结构，为新身份增加可选 `jev` 证据成员，同时更新
producer/consumer，而不改写旧行、旧报告或新增兼容执行路径。合并时验证原始正式
协议包和独立效果包，核对当前目录 Hash、源码、完整必交能力、预算、最终捕获、
Observer、审计及完整性清单；仅凭 report 自报 PASS 或共享 Adapter 不构成支持。

候选协议和效果先独立真实执行并经严格验证，见[纳入回执](../../.scratch/jev-1.14-multimodal-box/13-admission.json)。
正式目录仍需在自身 Hash 下重采两份证据；保留原包而不将候选重命名为正式。
固定简单几何图任务不代表生产准确率，Box confidence 不代表存在性，账户/区域/权限
未核验也不得冒称已核验。其他路线历史 PASS 不能自动成为新 Hash 的 PASS。

这是 ADR-0020/0021 模型能力与正式支持身份的扩展。ADR-0030 的统一引用投影和
ADR-0025/0026 的整体状态不因本票改变；父规格及整版发布仍由后续整合验收决定。
具体预算、复放、回滚和当前原始证据见[正式验证指南](../../ai/live/testdata/jev-formal/README.md)。
