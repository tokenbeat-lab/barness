# 08 — Host 拒答迁移与独立中文分类评估

Status: ready-for-agent
Type: task
Depends on: 07
Updated: 2026-10-12

## What to build

新增 Luna Host 装配/输入能力/typed refusal 示例；用相同固定中文合成数据形成独立 route/config/report，保留 Jev 默认用法和历史 bundle 校验。依据 [R18](../spec.md)、[设计 §2/9/10](../design.md)、[OD55/56](../acceptance.md)。

## Ownership and context

负责 local/cloud example 与 chineseeval 的参数选择、明确 route/config/report 身份和完整性；旧 CLI 默认 Jev，不改历史报告含义，不给协议库添加阈值/人工路由。两真实协议每进程独立 key、account alias 和预算，协作不回退他人编辑。

## Failure scenarios first

- nil error 后无分支断言 ChoiceAnswer，refusal 导致 panic 或误分类。
- 新 route 自动接受 TypeSafeOptions / structured instructions，或 example 将它们静默 stringify。
- 同数据集新配置误读 TypeSafe key/型号/目录，两个报告互相验证通过。
- refusal 从 accuracy 分母删除、失败样本漏掉、全拒答 COMPLETE 被冒称准确率通过。
- 修改 Jev 默认/历史 verifier 或覆盖原结果，以统一配置 schema 名义丢旧证据。

## Acceptance criteria

- [ ] 先写完整性/错路由/拒答/预算的 public process/E2E 反例，再实现 Host 扩展。
- [ ] 示例展示三个图片授权条件、JSON说明差异、nil/协议 Options 和封闭 Answer type switch。
- [ ] 明确单独 Luna command/config 的 route identity；Jev 旧命令/default、旧报告验证继续通过，无 deprecated CLI shim。
- [ ] dataset/config/catalog/code/hash 与实际 response 型号固定；合法拒答有状态，纳入总样本分母，不算正确答案。
- [ ] accuracy、conditional accuracy、confusion、refusal/失败/未执行、confidence/chosen probability、ECE/Brier、延迟与消耗分开。
- [ ] 两 route 的真实报告和 keys 独立，Luna 不共享 Jev threshold；空 bin/未执行指标为 null、不插值、不预设业务门槛。
- [ ] 24 调用/Attempt、零重试、串行、600 秒总预算；取消停止后续调用，真实预算和审计可核验。

## Verification and handoff

先 fixture 流程证据与旧 Jev verifier/default 对照；然后独立实际 Luna 评估，保留 Jev 当前适用报告。合成测试/NOT_RUN 不称为真实质量结论。向 09 提交两个路线的独立证据及旧报告兼容结果，不将效果写入支持矩阵。

## Comments

2026-10-12：协议支持与标注任务质量是独立结论，拒答不是协议故障。
