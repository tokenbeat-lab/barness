# 11: 交付 Host 迁移示例与独立拒答评估流程

**What to build:** Host 可显式装配 Jev 或 Luna、发现输入能力并处理封闭拒答结果；固定中文任务在独立 route/config 下产生可校验的完整报告，fixture 能证明拒答和取消流程，旧 Jev 用法保持。

**Blocked by:** 03: 交付 score/bool 混合答案、分布与逐题拒答；04: 证明 Decisions 用量完整性与输入阶梯计费；05: 打通分类图片、三重授权与完整资源边界

**Status:** resolved
**Type:** task
**Source tickets:** 08
**Requirements:** R03, R04, R10, R18, R19
**Acceptance cases:** OD55, OD56
**Updated:** 2026-10-13

## Responsibility

负责本地/云端示例和中文评估 CLI/config/report/metrics/verifier。使用宿主自带目录完成离线流程，不等待正式目录或 key；真实效果由下一票执行。不把阈值、业务 fallback 或转人工放入协议库。协作时保留他人改动。

## Failure scenarios first

- nil error 后无分支断言 ChoiceAnswer，拒答 panic 或被算 false/正确类别。
- 默认变 Luna，结构化说明或 TypeSafeOptions 被静默转换。
- Config/report 只靠 ModelID 识别路线，实际调用用错 key/provider/API。
- refusal 从总分母删除或被算调用失败；全拒答丢状态、花费或对 null 指标解引用。
- 当前 catalog 被强加给旧 Jev 报告，导致历史核验无法使用记录时的快照。

## Acceptance criteria

- [x] 先写 CLI/process E2E 的错 route、错 Hash、预算、混合/全拒答、usage 缺省、取消和旧默认反例，再扩展。
- [x] 示例展示 nil/协议 Options、string 说明共同子集及不兼容输入拒绝，封闭 Answer type switch、三图片授权条件与 Host 处理责任。
- [x] 旧 Jev 命令/default 和原 schema/配置含义保留；新 route 明确 Operation/Provider/API/Model，非法 route 在 key 读取前拒绝，不加 deprecated shim。
- [x] 数据集保持固定四类 24 样本，route/config/dataset/catalog/code Hash 与实际 ResponseModel 各自记录；不把 alias 权重称为固定，也不改写授权 ModelID。
- [x] refusal 独立状态，进入全任务分母且不计正确；正常、拒答、失败、未执行的总数恰好覆盖全部样本。
- [x] 全拒答且完整可核验时允许 COMPLETE、总准确率 0；无普通答案的条件准确率、ECE/Brier 等为 null，拒答率 1，消耗照实保留，不因空分箱 panic。
- [x] usage partial/unreported 如实保留并按明确完整性规则判定，不伪造完整报告；协议完成、任务完整与质量分别呈现。
- [x] 普通答案准确率、混淆、confidence/chosen probability 各自校准、延迟和费用独立，空 bin 不插值，不共享 Jev threshold。
- [x] 24 Call/Attempt、零重试、串行、600 秒总预算，取消不发后续请求；fixture、NOT_RUN 和 live 身份不能互换。
- [x] 旧 Jev bundle 用自身 recorded snapshot 继续核验且结论不重写；历史核验不赋予它当前发布适用性。
- [x] 输出公开命令、两个 route 的独立 fixture 报告及 verifier 反例、manifest/audit；真实质量不从 fixture 推断。

## Completion

无需外部 key 即可完成示例与报告流程。正式当前效果由独立实际执行票证明，历史报告保持自身日期和 Hash。

## Comments

2026-10-12：用户批准 13 张草案并替换旧票。此票按新编号发布；下列原票评论按来源保留，ready-for-agent 不表示已实施或依赖已完成。

### 原票 08

2026-10-12：协议支持与标注任务质量是独立结论，拒答不是协议故障。

### 2026-10-13 — 实施完成

显式 Jev/Luna route、Host 自带目录与配置、schema 2 拒答/条件指标/用量与独立协议状态、
有界真实传输捕获、公共 Client 无网络重放、完整 manifest/audit 与历史 recorded Jev snapshot
均交付。原默认命令、schema 1 含义和数据集 bytes 保持；实际 ResponseModel 不改写授权 ModelID。
零 Attempt 取消、双 request-ID header、全拒答/null 指标、coherent fixture-to-live 与 captures/usage
篡改反例全部通过。独立双轴审查无剩余问题，全部新模块小于 500 行。

97 项专项 E2E PASS，完整 race 退出 0、6,494 PASS/1,050 原有 opt-in NOT_RUN；普通/live-tag vet
通过。开启冻结 Pi 的普通全量：7,356 PASS、187 NOT_RUN、一项既有 100ms timeout 终态竞争 FAIL，
原单项复测 PASS；失败保留并另建 `.scratch/barness-ai-timeout-ordering/issues/01-call-and-read-idle-deadline-order.md`，
没有放宽原断言或修改协议库。所有运行及交付包审计无发现。

[公开命令与最终独立报告、反例、审查及 SHA-256](../host-tooling-evidence/README.md)。
真实质量仍为 NOT_RUN，本票未读取外部 key 或执行真实效果任务，交由票 12 独立完成。
