# 04 — score/bool、混合拒答、完整用量与输入阶梯

Status: ready-for-agent
Type: task
Depends on: 03
Updated: 2026-10-12

## What to build

将单选切片补齐为完整文本分类：score/bool 原生编码/guard，三类型 mixed answer、混合与全拒答、score 分布/图例、严格 envelope 和完整 presence/费用。依据 [R08/R10–12](../spec.md)、[设计 §4/6/7](../design.md)、[OD13/27–39/41](../acceptance.md)。

## Ownership and context

负责 Decisions request/guard/response/usage 及 public result 接缝；现有 TypeSafe decoder 和舍入函数保持。公共 score 仍 2–10；不新增 raw wire/cached 明细到 Observer，不复制 Responses 的 cache 拆桶。协作不回退他人编辑。

## Failure scenarios first

- bool 变阈值判断或丢 criteria；score levels 重排或缺分布/legend。
- 一个 refusal 清空其他有效答案，全 refusal 当上游错误，导致重试/零费用。
- 按答案 index 回配，duplicates last-wins，未声明 choice/错 type/超范围数被接受。
- 只验证普通题忽略其他损坏题，或所有答案尚未校验就发布部分结果。
- TypeSafe .01 舍入例外全局套用；合法 Jev 回放因此被收紧。
- usage 缺/零混淆、重复计数变 complete、input 扣/重复加缓存、阶梯边界或全请求金额错误。

## Acceptance criteria

- [ ] 新功能先写 OD13/27–39 和 2xx malformed/refusal 不重试变体，再实现扩展。
- [ ] 三题型 + named refusals 的完整集合原子发布；合法 mixed/all refusal stop/nil，无 invented reason。
- [ ] ScoreAnswer.Probabilities/Legend 从完整 native values/labels 转换；有限数、概率和、期望容差及 label 精确检查。
- [ ] unique envelope/usage 可先登记，任何答案失败保留；ambiguous/invalid usage 不选最后值、不假 complete。
- [ ] Input 保全部报告、cache 桶零、total/reasoning presence 和内部一致性按设计；免费 output 也如实报告。
- [ ] .1/.2 输入费率与 272000 严格边界，图像 token 不从 bytes 自行估；模型价格/usage Cost 不套聊天。
- [ ] Decisions 的 additive response metadata 与 exact known field/semantic 检查独立，Jev 仍严格未知字段拒绝。
- [ ] Jev 原 JSON/choice/bool/score/舍入/用量 replay 全保持，03 的 choice/refusal 不回归。

## Verification and handoff

通过 public Classify 捕获三型与错误响应，记录原数值、用量 completeness、expected/actual cost 和最后发布状态。执行目标新 case、原 Jev 对照、常规 2xx no-retry 生命周期。向 05 交付完整文本 Adapter，尚未图片/live 的条目保持未实现/NOT_RUN。

## Comments

2026-10-12：待实施；Refusal 是有效分类结果，整体协议损坏仍为失败。
