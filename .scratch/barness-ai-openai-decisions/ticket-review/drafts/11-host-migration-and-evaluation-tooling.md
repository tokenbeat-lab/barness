# 11: 交付 Host 迁移示例与独立拒答评估流程

**Status:** needs-triage
**Publication:** draft，待用户批准后转 ready-for-agent；尚未发布到 issues。
**Blocked by:** 03: 交付 score/bool 混合答案、分布与逐题拒答；04: 证明 Decisions 用量完整性与输入阶梯计费；05: 打通分类图片、三重授权与完整资源边界
**Source tickets:** 08
**Requirements:** R03, R04, R10, R18, R19
**Acceptance cases:** OD55, OD56
**Updated:** 2026-10-12

**What to build:** Host 可显式装配 Jev 或 Luna、发现输入能力并处理封闭拒答结果；固定中文任务在独立 route/config 下产生可校验的完整报告，fixture 能证明拒答和取消流程，旧 Jev 用法保持。

## Responsibility

负责本地/云端示例和中文评估 CLI/config/report/metrics/verifier。使用宿主自带目录完成离线流程，不等待正式目录或 key；真实效果由下一票执行。不把阈值、业务 fallback 或转人工放入协议库。协作时保留他人改动。

## Failure scenarios first

- nil error 后无分支断言 ChoiceAnswer，拒答 panic 或被算 false/正确类别。
- 默认变 Luna，结构化说明或 TypeSafeOptions 被静默转换。
- Config/report 只靠 ModelID 识别路线，实际调用用错 key/provider/API。
- refusal 从总分母删除或被算调用失败；全拒答丢状态、花费或对 null 指标解引用。
- 当前 catalog 被强加给旧 Jev 报告，导致历史核验无法使用记录时的快照。

## Acceptance criteria

- [ ] 先写 CLI/process E2E 的错 route、错 Hash、预算、混合/全拒答、usage 缺省、取消和旧默认反例，再扩展。
- [ ] 示例展示 nil/协议 Options、string 说明共同子集及不兼容输入拒绝，封闭 Answer type switch、三图片授权条件与 Host 处理责任。
- [ ] 旧 Jev 命令/default 和原 schema/配置含义保留；新 route 明确 Operation/Provider/API/Model，非法 route 在 key 读取前拒绝，不加 deprecated shim。
- [ ] 数据集保持固定四类 24 样本，route/config/dataset/catalog/code Hash 与实际 ResponseModel 各自记录；不把 alias 权重称为固定，也不改写授权 ModelID。
- [ ] refusal 独立状态，进入全任务分母且不计正确；正常、拒答、失败、未执行的总数恰好覆盖全部样本。
- [ ] 全拒答且完整可核验时允许 COMPLETE、总准确率 0；无普通答案的条件准确率、ECE/Brier 等为 null，拒答率 1，消耗照实保留，不因空分箱 panic。
- [ ] usage partial/unreported 如实保留并按明确完整性规则判定，不伪造完整报告；协议完成、任务完整与质量分别呈现。
- [ ] 普通答案准确率、混淆、confidence/chosen probability 各自校准、延迟和费用独立，空 bin 不插值，不共享 Jev threshold。
- [ ] 24 Call/Attempt、零重试、串行、600 秒总预算，取消不发后续请求；fixture、NOT_RUN 和 live 身份不能互换。
- [ ] 旧 Jev bundle 用自身 recorded snapshot 继续核验且结论不重写；历史核验不赋予它当前发布适用性。
- [ ] 输出公开命令、两个 route 的独立 fixture 报告及 verifier 反例、manifest/audit；真实质量不从 fixture 推断。

## Completion

无需外部 key 即可完成示例与报告流程。正式当前效果由独立实际执行票证明，历史报告保持自身日期和 Hash。
