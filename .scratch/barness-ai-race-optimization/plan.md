# barness ai race 验证优化实施方案

Status: implemented; release acceptance open
Date: 2026-10-20
Review baseline: `eae89d5a1170bc8b716e6ee9ab26d016b9bbd013`

Review update: 用户已批准按收窄后的方案实施；当前进展与实际验证以 [status.json](status.json) 为准。实际 race 仍单独等待人工选择。

本方案已获批准并完成实现与普通验证，实际结果见 [交付索引](verification.md)。按用户最新要求，race 改为最后由人工显式选择触发的独立步骤；开发与普通验证均不自动启动 race。保留正式发布的完整 race 要求，再依据实测决定单轮性能改造。既有复盘记录为 76 次启动、累计约 5 小时 9 分钟，其中 5 次全量约 3 小时 31 分钟，见[执行明细](../barness-ai-minimax/race-run-audit-2026-10-19.json)。

## 本次建议批准的范围

执行 A 至 E 的实现及普通验证：验证策略与技能调整、发布前置检查与失败停止、定向场景与时序覆盖核对、最小计时、稳定源码的普通离线回归。E 中的实际 race 在完成这些工作后单独等待人工选择；确认实施本方案不自动授权执行 race。F 先交付可取得的数据分析；需要 race 数据的部分随人工执行完成，性能代码改造另行审查。

本次保持以下交付要求：完整 `go test ./...`、完整 `go test -race ./...`、双 vet、冻结差分、压力、独立 race 证据、当前来源资格与逐包审计。新增操作不读取厂商凭据、不发起真实模型调用、不更新支持矩阵或历史证明。本次确认不能解释为新增付费预算。

## 目标执行顺序

```text
变更与影响范围确认
  → 定向普通 E2E（失败则修复）
  → 核对并发/时序测试覆盖，记录建议 race 范围（不执行）
  → 关键源码审查（在需要真实验收时先完成）
  → 执行来源稳定
  → 发布输入 preflight
  → 双 vet
  → 完整普通测试及其证据消费/审计
  → READY_FOR_RACE：普通验证完成，race NOT_RUN
  → 人工显式触发（也可暂不执行）
  → 复核普通阶段的来源/输入/回执仍有效
  → 完整 race
  → 原有发布判定及来源/输入复核
  → 固定范围收尾
```

任何阶段失败都保存已经取得的事实与阻断原因。默认停止尚未开始的昂贵阶段；显式诊断也只能继续当前阶段，不跨越人工 race 边界。人工暂不执行时，可以交付实现与普通验证结果，但必须标明 race NOT_RUN、发布未认证，不能关闭要求完整发布证明的验收项。

## A 验证策略与技能

拟修改：`AGENTS.md`、新增 `docs/agents/verification.md`、`docs/agents/issue-tracker.md`，以及仓库外 `/Users/cyber/.agents/skills/implement-spec/SKILL.md`。新增 ADR 记录分层验证与失败停止；编号实施时按当前目录确定。

职责分配：通用 implement-spec 只负责读取并遵循项目验证策略；barness 的测试范围、审查时机、race 人工触发和证据要求由项目策略定义，releasegate 实现阶段拆分。修改通用技能不是拆分 CLI 的技术前提，其作用是让使用该技能的 agent 遵守项目已有约束。

`AGENTS.md` 只增加一条触发指针：决定测试范围、准备付费验证或完整门禁时读取 verification 文档。详细规则只维护一处。

| 变更类型 | 开发阶段 | 最终阶段 |
| --- | --- | --- |
| Client 生命周期、队列、取消、共享状态、资源策略 | 受影响公共 E2E，登记并发/隔离/释放的建议 race 范围 | 稳定源码完整普通通过后，等待人工选择完整 race |
| Provider 协议编码、解析、历史和用量 | 定向协议 E2E，登记涉及的异步、续接和资源路径 | 稳定源码完整普通通过后，等待人工选择完整 race |
| 发布调度、证据消费者或测试工具 | 自身公共 CLI/进程测试与受影响消费者；并发范围单独登记 | 稳定源码完整普通通过后，等待人工选择完整 race |
| 纯文档、展示清单、审查记录 | 引用、格式、状态及来源相等性检查 | 无选定执行输入变化时引用已有测试事实，不重跑 race |
| 模块、SDK、构建标签或无法可靠限定的共享变化 | 扩大普通验证范围，并登记完整 race 的必要性 | 完整普通及原生来源检查；完整 race 人工触发 |

JSON、目录和证据数据不自动属于“纯文档”；参与运行、价格、追溯或消费者判断的数据，按实际语义确定验证范围。

以下五项写入 barness 的 `docs/agents/verification.md`，不作为通用技能的全局默认规则：

1. 行为代码继续失败先行；文档与清单使用一致性检查，不要求行为 TDD。
2. 读取项目验证策略，按变更影响选择开发阶段验证，不机械地在每票和每次合并后重跑全量。
3. 付费或长时验收之前完成关键边界整合审查，再冻结来源；最终审查仍保留。
4. 普通验证结束后明确报告 READY_FOR_RACE，列出范围、命令、预计耗时及来源。只有人类显式命令或明确聊天指令才能触发 race；方案确认、测试失败、合并、诊断和收尾均不隐含该授权。定向 race 也适用同一原则。
5. 为每阶段记录测试命令、实际结果、来源和已有证据指针；文档收尾不新增逐轮复制器。

implement-spec 仅拟作两处最小调整：

1. 步骤1在读取 spec 和工单时，同时读取仓库指示的项目验证策略；在选择及执行验证前遵循其中的验证范围、审查检查点、证据要求和授权边界。
2. 步骤4保留现有 TDD 默认流程；项目策略明确区分任务类型及验证方式时，以该策略为准，避免无条件的 TDD 条款覆盖项目规则。

当前技能没有要求每票或每次合并都运行全量/race；这部分不能归因为技能的硬性要求。其每票 worktree、merger、最终 code-review、frontmatter 和调用方式均不在本次拟改范围。通用技能不写入 READY_FOR_RACE、人工触发 race、本项目命令、型号或目录，也不为其他项目新增 race 的默认授权规则。仓库外技能的原文与改动 diff 单独交付，便于独立回退。

tracker 只补充验证计划/实际结果指针和唯一当前状态约定，不回写旧工单、旧失败或账本。

完成标准：通用技能只承接项目策略；barness 能明确决定普通验证范围，并在所有自动路径末尾停在 race 人工边界。每条项目规则均有触发条件和结束条件；未定义相应项目策略的其他仓库不因本次技能调整而继承 barness 的 race 规则。

## B 发布前置检查与默认失败停止

拟修改/新增：

- `ai/release/cmd/releasegate/main.go`：参数解析和装配。
- 同目录 `preflight.go`、`execution.go`：仅封装该 CLI 的前置收集与顺序执行，不建立通用工作流框架。
- `ai/internal/testkit/release/preflight.go`：复用现有判定，定义前置检查结果。
- `ai/e2e/release_gate_process_test.go`、拟新增 `release_gate_preflight_test.go`。
- 必要时在 CLI 目录增加先行的进程级隔离测试 `execution_test.go`。

CLI 拟增加阶段选择与两个控制参数，其他输入参数沿用现有接口：

```sh
# 只检查当前外部输入；不启动 vet/完整普通/race
go run ./ai/release/cmd/releasegate -preflight <现有输入参数>

# 第一阶段：前置、双 vet、完整普通及审计；停在 race 待执行
go run ./ai/release/cmd/releasegate -stage check -out <check-dir> <现有输入参数>

# 第二阶段：人工选择后，验证阶段一事实，实际执行完整 race 并作最终判定
go run ./ai/release/cmd/releasegate -stage race -from <check-dir> -out <race-dir>

# 显式收集第一阶段诊断：仍不自动执行 race
go run ./ai/release/cmd/releasegate -stage check -collect-all <现有输入参数>
```

执行入口缺省阶段为 check，绝不自动 race。`-preflight` 与 `-collect-all` 互斥，均不能与 stage race 混用；stage race 必须提供 from，不能重新指定一套不同的业务输入。新的执行阶段参数与既有 `-bundle`/`-race-bundle` 只读重评估入口互斥。原 bundle 重评估保持命令 NOT_RUN，不能用作新认证。

普通阶段通过输出 `status: READY_FOR_RACE`、`race: NOT_RUN`、`release.pass: false`。为保留原认证命令“退出0才可能完整通过”的语义，普通待执行状态退出2，实际失败退出1，全部正式条件通过后才退出0；独立 `-preflight` 的退出0仍只表示预检就绪。终端明确区分普通验证完成、等待人工与实际失败，调用方需读取阶段状态。

人工触发来自实际命令或聊天中的明确指令；清单里自写 `humanApproved: true` 不能成为授权。阶段二不从输入报告执行任意命令，race 的包范围、选项和资源边界仍由工具固定。

前置检查复用 `CheckReleasePlan`、`CheckSnapshot`、`CheckClassifierExtensions`、已有 live/support-matrix 校验、`ChineseEffectsEvidence`、`VisualEffectsEvidence` 与审计。只检查目前能成立的外部前置条件，不伪造尚未生成的普通/race bundle。

前置检查覆盖：

1. trace/ledger/catalog/模块身份、固定映射与完整必交集合。
2. 17 份自身支持及四项业务效果的必要输入、身份、状态、当前来源和审计。
3. 输入的完整性、读取上限、路径及原有消费者要求；不得只检查“文件存在”。
4. 失败原因分类为输入缺失、完整性、资格/来源、映射/快照等，不把旧 PASS 当作当前适用性。

输出一个 `preflight-report.json`，字段包括 `scope: preflight`、`ready`、检查明细和来源。`ready=true` 只表示可以进入后续阶段，不是发布结论。独立 preflight 的结论不会被后续认证直接缓存引用，认证进程重新检查自己的输入。

默认 check 顺序为 `preflight → vet → vet-live → ordinary → ordinary 的公共消费/审计 → READY_FOR_RACE`。任一步失败即停止后续启动；全部通过也不执行 race。

阶段二从阶段一目录读取固定结构的来源、构建工具身份、实际命令回执、输入计划与原生 bundle，复核完整性、原生消费者及当前适用性；不能只接受 `passed: true`。来源、依赖、相关测试或选定输入变化时拒绝沿用，先重新完成受影响普通验证。纯文档变化依完整来源集合相等性判断。阶段二成功执行 race 后，组合两阶段实际证据再调用原 Evaluate；第一阶段原件不覆盖，第二阶段用新目录保存结果。

继续使用 `release.Command` 与 `Evaluate` 的现有语义：未执行命令不伪装成成功条目，`commandGate` 本来就会将其判为 not run。增加一份小型 `execution-report.json`，逐阶段记录 `NOT_RUN/RUNNING/PASS/FAIL`、实际数字退出码、开始/结束、耗时和 `blockedBy`；不复制整套原生证据。

普通/race bundle 未生成或无法消费时也输出 `pass=false` 的可读发布报告，说明错误和缺失项。缺失输入不得填充合成成功 case。参数尚不能解析或输出目录不可写时，只能返回非零与 stderr，不承诺不可写目录中仍有报告。

输入在前置检查、消费和最终收尾仍按现有严格读取与原生来源规则核对；前置检查不能免除后续读取时的校验。来源或输入发生变化，认证失败。

各实际阶段开始前也核对本次冻结的编译相关来源与选定输入，变化时停止后续启动；不能仅凭 HEAD 相同或文件时间相同判断未变。collect-all 只对可解析的 check 请求继续收集，参数冲突、不可写输出或无法装配请求时仍直接拒绝。调用者和文档中的原自动全量入口同步更新；历史冻结控制器只保留，不修改为新执行授权。

### 先行验证场景

先写失败清单与测试，再改 CLI：

| 场景 | 必须观察到的结果 |
| --- | --- |
| 默认模式缺少支持输入、陈旧来源、错误 mapping、坏 inventory | 前置拒绝；启动 vet/普通/race 次数均为0 |
| 独立 preflight 全部就绪 | ready=true；没有完整执行；不能作为 release PASS |
| vet 或 live-vet 失败 | 后续普通/race NOT_RUN，保留真实退出码 |
| 普通测试失败 | race 启动次数0，发布false |
| 普通 exit0，但证据缺失/不完整/审计失败 | race 启动次数0，发布false |
| 普通全部通过，未人工触发 | READY_FOR_RACE，退出2，race启动0次，发布false |
| 只有整体方案确认、文档或合并变化 | 不启动定向或完整race |
| stage race 缺 from、普通失败或目录被篡改 | 拒绝，race启动0次 |
| 两阶段之间源码/依赖/相关测试/选定输入变化 | 拒绝旧阶段一，不能认证混合来源 |
| race 失败或实际出现检测器报告 | 保留失败，不能发布 |
| 全部阶段与必交消费者通过 | 每条命令恰好执行一次，仍经原 Evaluate 判定 |
| collect-all 的可装配前项失败（含前置资格失败） | 后续check命令确实执行；始终不自动race，原失败仍保留；不安全或不可读取输入仍直接拒绝 |
| 子进程启动失败、取消或写报告失败 | 非零退出；已知事实不丢失，未完成不能算PASS |
| 使用 bundle 重评估或冲突参数 | 不能替代执行；参数组合明确拒绝 |

外部 CLI E2E 复用现有 `releaseProcess`。完整正向调度不能靠伪造正式厂商证明建立：必要时对私有阶段执行函数做进程级隔离测试，使用真实子进程夹具观察顺序、退出和未启动行为；先穷举失败方式再实现。该夹具只验证编排，明确标记受控结果，不能生成正式发布证明。生产 CLI 不增加跳过 preflight 的测试开关。

## C 建议 race 范围与确定性时序

不新增测试选择器框架。verification 文档维护按风险选择的既有测试入口示例，作为人工选择定向诊断时的依据，不自动运行。最终正式覆盖仍为人工触发的完整 race。

MiniMax/共享生命周期的起始集合包含：

- `TestMiniMaxSharedTenantRoundTrips`、`TestMiniMaxSharedConcurrentRelease`、`TestMiniMaxSharedAdmissionIsolation`。
- `TestMiniMaxSharedCancellation`、`TestMiniMaxSharedObserverFailures`、`TestMiniMaxSharedFrozenRetries`。
- `TestEventQueueLimits`、`TestEventQueueReleasesAsRead`、`TestProtocolTimeouts`。
- 共享逻辑变化时追加相应旧 Provider 的取消、Settlement 与隔离入口，不能用 MiniMax 自身通过替代。
- 容量策略变化时运行原压力集合和 `TestMiniMaxSharedDesignLoad`，保留相应 opt-in，不能把未启用的压力测试写成通过。

视觉取消直接复用当前已修复的 `TestVisualUnsentCancellation`、`TestVisualUnsentAfterPin`、`TestVisualSentCancellationBeforeDecode` 及既有 public verify/取消计量场景。先核对覆盖，已覆盖的阶段不重复添加测试。

若发现真实缺口，先通过现有受控 HTTP/Transport/Host hook 和 channel 建立失败场景，再修复。每个新增时序必须证明：到达了哪个阶段、何时取消、实际 CredentialReads/HTTPAttempts、唯一终态、资源释放，以及保留包可以公开回放。沿用既有有限阶段语义，不发明新的测量值；不用 sleep 猜测到达阶段，也不添加仅供测试使用的生产开关。

拟改文件限实际缺口涉及的测试/夹具，例如 `ai/e2e/minimax_shared_*_test.go`、`ai/e2e/visual_evaluation_*_test.go`、`ai/examples/visualeval/*_test.go`。若核对发现已有覆盖充分，本阶段仅记录入口与理由，不为取得新 RED 人为破坏已有实现。

完成标准：局部变更有明确的建议并发验证集合；取消边界在普通 E2E 中即可确定性覆盖，不依赖自动 race。人工未选择 race 时明确保留 NOT_RUN。

## D 最小计时与诊断

拟新增 CLI 同目录 `testevents.go`，复用 B 的执行路径。普通/race 的实际 Go 参数增加 `-json`，保留 `-count=1`、现有完整包范围、20m/75m 包级限时与差分/压力开关。

只流式保存 `Time/Action/Package/Test/Elapsed` 等元数据到同一份事件文件；原测试输出继续沿现有日志渠道输出。执行报告记录实际 argv，不能把新增 `-json` 的执行写成历史旧 argv；旧收据和冻结控制器保持原样。

执行器只保留有界失败尾部，避免现有 bytes.Buffer 保存全部输出后才截断。JSON 解析、输出消费和子进程等待均必须有界；遇到格式或写入错误，保留该错误、结束/回收进程并拒绝认证，不能悄悄丢弃事件却声称计时完整。

本轮统计命令总耗时、包耗时、顶层/子测试耗时、普通/race 放大比例和两者的包尾部差额。父子测试时间重叠、并行包重叠，不能相加当作墙钟。现有 TestMain 的 Finish/Audit 未各自计时，因此不能把包尾部差额直接命名为“审计耗时”。

本阶段不改 `ai/internal/testkit/evidence/evidence.go` 的原生 manifest 或库存格式。若数据表明证据尾部是主要热点，再在 F 中提出精确的额外测量点。

## E 普通验收交付与人工 race

执行前完成 A–D 的源码和先行场景审查、输入影响检查。开发过程只跑必要普通验证。稳定后实际执行双 vet、完整普通并交付待 race 清单，至此停止自动工作；完整 race 单独等待人工选择。所有实际测试使用无厂商凭据环境，保留差分/压力及独立 bundle。人工选择后按固定范围运行，任何重跑均说明原因并遵循其明确授权范围，不因失败自动扩大次数。

优化代码本身可能使旧真实证据的来源过期：视觉 `sourceFiles()` 收集所有非测试 Go，发布 CLI 变化也会命中；外层 paid launcher 又冻结更宽的 Go/JSON/Python 集合。实施时先对完整原生 selector 集合做影响清单，区别“构建/测试通过”和“新版本已经获得完整发布资格”。

如果既有 live 输入仍具当前资格，第一阶段产出待执行状态，第二阶段人工运行后取得完整 verdict。如果它们过期，新 preflight 必须拒绝；仍可独立执行普通 Go 离线回归验证工具改动，race 继续由人工决定，结果只认证实际已运行范围。受影响组合续期另列范围、预算和批准；本计划不授权付费，不缩小 selector、不刷新旧 hash、不拼接旧报告。

产物保持最小集合：优化源码与审查、前置/执行报告、事件元数据、普通/race 原生 bundle、审计和一个最终索引。沿用既有留存能力，不为本次优化新增逐轮归档器/读取器/复制器。

## F 性能结论与后续改造边界

先按普通运行列出热点、子进程相关场景和包尾部差额；人工完成 race 后再补放大比例与对比。未选择 race 时不为了性能分析自动补跑。

只有取得证据后才选择下一步：重复构建则考虑在同一来源下复用测试进程的现有二进制；重复大夹具则考虑一次生成后只读引用；Finish/Audit 显著则先增加精确分段计时，再优化重复解析。保持每条跨信任边界校验、取消回归、独立 race 证据和完整 case 覆盖。

本次不批准提前实现内容寻址存储、通用测试调度平台、磁盘验证缓存、`!race` 排除、缩小最终全仓集合、提高默认并行或降低场景超时。这些选择需要针对实际热点给出后续设计。

## 提交顺序与完成判据

| 顺序 | 变更集 | 验收方式 |
| --- | --- | --- |
| 1 | A 项目策略、ADR、通用技能最小衔接 | 检查文档、协议、共享取消、发布工具案例；race规则只在项目内定义，通用技能遵循项目策略 |
| 2 | B 前置与执行停止，含先行测试 | 公共 CLI 拒绝与必要进程级隔离场景；未通过前置/普通时 race 启动0次 |
| 3 | C 现有覆盖核对与必要时序补充 | 普通E2E及待选race集合；不存在新增实际缺口则不增加重复测试 |
| 4 | D 计时，含先行事件/错误消费场景 | 事件字段、真实数字退出与输出完整；计时失败不冒充完整 |
| 5 | E 普通验收与 F 初步性能结论 | 双vet/完整普通，race待人工；人工选择后补齐实际结果与分析 |

新增必交 case 与 traceability、固定 approvedMappingHash 在同一变更集内同步验证，避免再次产生“测试已改、映射未改”的收尾缺口。单文件保持不超过500行；新增代码限于现有 release/evidence 使用边界，不新增依赖。

每个变更集可以独立 review/revert。回退 CLI 恢复旧调度即可，不涉及数据库迁移、支持矩阵回写或厂商预算；新生成的历史结果继续保留，不能删除失败记录。技能通过其独立原文/diff 回退。

可承诺的行为指标：人工触发前所有自动路径的 race 启动数为0，包含普通全部通过和 collect-all；race未执行时完整发布false；来源变化不能沿用旧普通结果；新增审查材料不引发新工具开发。单轮 49 分钟降到多少、总耗时降低百分比，待实际热点测量后再确定。

用户已确认实施本方案；当前事实与未完成项只维护在 status.json，race 的实际执行仍单独等待人工明确选择。
