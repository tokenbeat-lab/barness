# 18 分类预算须隔离模型操作身份

Status: claimed
Type: task

## Context

阻塞[07最终验收](07-final-acceptance.md)，当前验收状态仅见[规格](../spec.md)。恢复4的第18项OpenAI Decisions实际exit1，17项PASS、后六项NOT_RUN；原失败、父UNKNOWN、完整预留和关闭原件保持。此次仅离线诊断和最小修复，不授予新付费、Key、grant/claim或race。

## Required investigation

在公共Client/CLI进程与合成transport建立和重复最小RED，再提出三个可证伪假设并逐项验证。沿当前真实最终wire与固定价格合同，区分同名型号的Operation/Provider/API；不得绕过final-wire guard、变更生产价格/协议/模型/来源选择器，或用账本0/0宣称整次历史零消费。

覆盖本计划所有操作/服务/协议及其最终wire形状，尤其OpenAI Decisions、TypeSafe/Jev、四effects和图片。修复后通过受影响公共进程普通回归，保留真实argv/UTC/数字退出及工件；核对Regional/MiniMax/GLM执行来源影响，不把旧PASS自动当新资格。恢复5及新付费计划另列范围，本票不实现。

## Evidence

原实际目录：`.evidence/provider-sites/final-acceptance-20261020/recovery4/live/openai-decisions/20261020T195220.781600000Z`。自身报告9Calls/9 recorderHTTP且全0ms transport；guard ledger原件0/0。该差异须解释，不能改写任何原件或计数。离线诊断与RED/GREEN另存 `.evidence/provider-sites/final-acceptance-20261020/decisions-budget-diagnosis-dev/`。

## Offline findings and implementation

三个先验假设是同名chat价格污染、input-only分类缺少max output、账本配置/额度不合法。公共Client RED重复exit1；同账本TypeSafe input-only可发送，排除后两项作为共性原因。仅外部diagnostic overlay保留实际trace：found=true、InputRate0.25、OutputRate0.75、OutputReserve0；Decisions目录自身应为最高tier输入0.2、输出0。同ID的chat目录先被累加，分类分支只重置found/outputReserve而未清价格，导致合法无output请求被decline。

最小修复在价格累加前互斥选择classifier/image/chat目录，保持每目录原有最高tier/cache语义。没有改变生产协议、Host/SDK、模型、价格、wire上限、guard开关或selector，没有新增shim。公共进程锁住Decisions0.2/0与零输出合同，原cap/错policy/失败不退款/Anthropic长TTL/图片工程预留回归保留。

八条已PASS generic的全部84原final payload经公共Client/Hook/recorder/guard/合成transport，原guard基线与修后逐份method/URL/requestBody及完整owned USD ledger bytes相等。另六地区＋三M3＋三GLM十二路线183原最终请求同样精确相等，但严格执行来源包含guard，因此它们仍须新当前资格。全计划尚未成功的七流程均启用guard并完整受控成功：Decisions9、TypeSafe3、Jev图像协议12、文本两路各24、视觉两路各12；受控效果保持RealQuality NOT_RUN，不能冒充真实效果。

[永久开发索引](../../../.evidence/provider-sites/final-acceptance-20261020/decisions-budget-diagnosis-dev/index.json)、[84份精确差异](../../../.evidence/provider-sites/final-acceptance-20261020/decisions-budget-diagnosis-dev/compatibility-eight.json)、[诊断与范围](../../../.evidence/provider-sites/final-acceptance-20261020/decisions-budget-diagnosis-dev/diagnosis.md)保留实际命令、退出、源、所有RED与装配中间失败。原停止2032 pins逐项一致。GLM/M3/Regional均含guard，没有不改执行源码而合法恢复此阻断的方法；禁止绕guard、重写价格或刷新旧证明。独立双轴review及新有限付费方案不在本次离线成功范围，Status继续claimed；唯一当前状态见[规格](../spec.md#当前状态)。

## Integrated review follow-up

恢复5整合66f3e50的独立Spec初审0；Standards两项P2由单一implementer先建立持久公共RED，再作最小修复：已知typed嵌套字段的alias/冲突不再覆写失败审计，合法描述扩展仍允许；canonical intent收尾读取复用现有非阻塞有界JSON读取。定向六原生preflight场景及两个FIFO场景均已GREEN。没有修改排除的actual approval疑点、guard字节/价格/SDK/生产协议或历史原件。

[统一修复永久索引](../../../.evidence/provider-sites/final-acceptance-20261020/recovery5-review-fix-dev/index.json)记录精确源码、RED/中间失败、最终普通组合与退出；[控制器原48/48事实](../../../.evidence/provider-sites/final-acceptance-20261020/recovery5-controller-dev/index.json)为319.891s、clean66来源，不冒充本次新源码结果。票据保持claimed等待最终整合审查/各自验收边界；实际冻结、根检查与新增付费状态见[准备索引](../../../.evidence/provider-sites/final-acceptance-20261020/recovery5-preparation/index.json)，唯一功能当前状态见[规格](../spec.md#当前状态)。截至本次离线修复阶段，新US36/19进程尚无实际批准，READY/race NOT_RUN、release.pass=false；旧28.80与UNKNOWN4不退款。后续实际状态由准备索引记录。


## Root keyless artifact follow-up

ebb98e6的184原生/50 Python普通PASS与双轴0/0原件保留；首次根完整无Key检查仍实际exit1，三条预算pipeline的process/assertions共六HOME audit发现，原目录封存不覆盖。公共进程先重现真实checkout位于合成HOME的同六失败，再仅alias各自证据根，保留完整stdout/argv/route/session与环境审计。四个预算测试仍绑定精确Git来源，任意新test漂移继续拒绝；guard/价格/wire与原84兼容证明不改。

[本次永久开发索引](../../../.evidence/provider-sites/final-acceptance-20261020/recovery5-home-audit-fix-dev/index.json)关联RED/GREEN、84请求精确不变与受影响普通范围，不能代替新根完整检查；后续实际状态只见[规格唯一当前状态](../spec.md#当前状态)及其准备索引。本票和07继续claimed，不把无Key工件修复当作新增付费资格。
