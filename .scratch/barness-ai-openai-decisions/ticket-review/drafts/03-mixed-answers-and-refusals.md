# 03: 交付 score/bool 混合答案、分布与逐题拒答

**Status:** needs-triage
**Publication:** draft，待用户批准后转 ready-for-agent；尚未发布到 issues。
**Blocked by:** 02: 打通 Decisions 文本单选与最终请求闭环
**Source tickets:** 04
**Requirements:** R02, R04, R08, R09, R10, R11, R19
**Acceptance cases:** OD13, OD22, OD25, OD27, OD28, OD29, OD30, OD31, OD32, OD33, OD34, OD39, OD41
**Updated:** 2026-10-12

**What to build:** Host 在一次文本 Logical Call 中提交 choice/score/bool，得到完整命名答案、score 分布与图例；混合或全拒答仍成功，任何协议损坏使整次 Answers 为空。

## Responsibility

负责 score/predicate 请求与最终 guard、完整答案转换和封闭公共结果。用量计数及费率专项由计费切片拥有，保持已有唯一合法用量的保留路径。TypeSafe 舍入和严格 decoder 独立。协作时保留他人改动。

## Failure scenarios first

- bool 被变为阈值判断，true/false 说明丢失，score levels 被排序或合并。
- 返回数组按 index 回配；一次 refusal 清空其他答案或触发错误重试。
- 多/缺/重复题或字段、错误 kind、错 score label/index 与概率损坏被修补或部分发布。
- Decisions 套用 Jev 两位小数例外；TypeSafe 未知 native refusal 被误接受。
- Hook 改题型或 levels 后仍验证原题；全拒答缺少最终问题的完整名称。

## Acceptance criteria

- [ ] 先写三题型、合法并列、乱序结果、混合/全拒答及所有损坏响应 E2E，再实现。
- [ ] bool 映射 predicate；固定 true/false 拼接规则，省略/空说明对照；score 保留有序重复 label 和公共 2–10 范围。
- [ ] Hook 可在已实现类型间合法改名、改种类、重排候选与 levels；完全依据最终冻结记录校验，原对象不被修改。
- [ ] 普通与拒答 Answers 完整原子发布；refusal 无编造原因，混合和全拒答 stop/nil，2xx 不重放。
- [ ] 缺/多/重复名称、case-alias/重复字段、错 discriminator、非字符串 choice、非法数值整体 protocol/response。
- [ ] score 概率覆盖整数 0..n-1、labels 精确，概率和及期望绝对容差 1e-6；保留原数值、Probabilities/Legend，无归一化。
- [ ] 已知答案语义严格，benign additive metadata 有界忽略；拒答夹带矛盾普通答案与非法语义字段有明确拒绝证据。
- [ ] 任何答案错误保留唯一合法 usage；TypeSafe 原拒答拒绝、宽 JSON、图例与有界舍入 replay 不变。
- [ ] 从两个公共入口生成请求、结果、错误、用量、Attempt、断言、manifest 与审计证据。

## Completion

完整文本题型与拒答闭环可单独演示；不宣称已完成图片、usage presence 专项或真实服务验证。
