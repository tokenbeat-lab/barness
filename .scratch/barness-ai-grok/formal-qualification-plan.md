# Grok 正式资格：冻结顺序选择

Status: needs-info

## 已确定事实

- [候选真实v3](evidence/candidate-live-v3.json)已18/18 PASS、24调用、零自动重试，formalQualified=false；正式消费者实际exit1。
- `ai/internal/testkit/release/record_live.go:59`要求正式report的catalog版本/hash等于当前BuiltinCatalog；`chat_qualification.go`还要求当前Git提交。
- 真实回放E2E/夹具不进入xaiSource，却改变外层验收来源；资格器、live harness和目录接线进入原生来源。旧候选不能改pin为正式。
- [spec R21](spec.md#r21--候选真实资格与正式目录)要求候选不自动激活BuiltinCatalog，并要求晋升具备完整证据和适用SOP。06工单按字面把目录激活放在人工门槛之后；该顺序与现有消费者的当前Builtin约束冲突，激活后来源变化又会使旧门槛失效。

## 建议：区分未发布源码接线与正式发布

沿用已有regional promotion的候选原件独立消费及后续正式报告思路：

1. 离线完成真实脱敏回放、xAI专属原生资格合同、正式harness和最终目录接线，在未发布的`codex/barness-ai-grok`分支形成可审查的最终实现；候选成功本身不触发任何目录修改或正式支持声明。
2. 两轴审查、冻结最终源码/目录版本/hash/二进制/实际输入，完整普通验收。
3. 以该最终冻结来源另取正式live的明确授权和新grant，执行xAI自身正式场景及其他受影响路线各自的授权范围；不能复用本次候选授权，也不自动付费。
4. live有效后再分别取得人工race与适用压力选择，在同一冻结来源完成。
5. 全部Release Gate通过后才接受/发布正式晋升；此后不再改目录或运行源码。不自动合并main、push或发布。

该建议将“正式晋升”解释为接受/发布，而允许其最终源码先在未发布分支准备。这是需要明确的合同解释，不能悄悄声称满足了“任何Builtin源码变更都必须在全部门槛之后”的字面顺序。完整通过前release.pass始终为false。

## 保留字面隔离的影响

若要求未发布分支中的Builtin函数也必须一直排除xAI，则当前正式消费者不能沿用，需要重新设计晋升前资格合同及目录激活后的来源适用性；不得放宽未知路线、借用旧报告或伪造当前pin。这是另一项协议设计，不能通过文档改名消除。

## 离线实现边界

- 公共Client回放工作可独立继续，不依赖上述选择。保留原始捕获/hash，移除auth/cookie/追踪标识，对密文和opaque ID做一致替换，明确回放不产生live资格。
- xAI拥有自身identity、完整能力集、用量算术、24次调用上限和grant/claim/source/binary/catalog检查；不借regional身份、不扩展不相关路线。
- 候选、controlled/replay与formal分别判断；失败/缺场景/截断/来源漂移/非零退出均拒绝正式资格。
- 不新增代际控制器、构建标签、兼容shim或自动发布流程。
