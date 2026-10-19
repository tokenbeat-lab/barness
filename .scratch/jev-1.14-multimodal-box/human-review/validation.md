# 示例、测试、探针、正式支持和发布门禁

固定版本`f5cf536`。这些模块负责“怎样使用、怎样证明”，与运行库行为分开审查。

## Host可执行示例 对应工单10

[示例使用说明](/Users/cyber/RestoX/barness/ai/examples/classifierimages/README.md:1)给出统一输入、受控/真实模式、两入口和撤销流程。它是CLI示例，不是新增Web UI。

| 起读位置 | 内容和解释 | 人工核对重点 |
|---|---|---|
| [main.go L25](/Users/cyber/RestoX/barness/ai/examples/classifierimages/main.go:25)、[inputs.go L13](/Users/cyber/RestoX/barness/ai/examples/classifierimages/inputs.go:13) | 组合路线×场景×plain/hooked；构造context、复用、State、Instructions、criteria、Box和拒绝输入。 | CLI默认行为和live选择；Decisions局部位置与Box应明确拒绝。 |
| [host.go L31](/Users/cyber/RestoX/barness/ai/examples/classifierimages/host.go:31)、[run.go L44](/Users/cyber/RestoX/barness/ai/examples/classifierimages/run.go:44) | Host resolver、租户/凭据、Binding和图片权限撤销、Observer等待、调用与结果记录。 | 新调用被撤销，已固定调用按快照完成；结束后body/permit/call归零。 |
| [fixtures.go L36](/Users/cyber/RestoX/barness/ai/examples/classifierimages/fixtures.go:36)、[interpret.go L35](/Users/cyber/RestoX/barness/ai/examples/classifierimages/interpret.go:35) | 生成合成图和裁剪/旋转几何；按最终图片身份解释图例，L86起把Box还原到原图坐标。 | Host显式记录搜索图归属；不能假定第一张ImageInputs就是Box所属图片；视图深复制不能修改原结果。 |
| [capture.go L30](/Users/cyber/RestoX/barness/ai/examples/classifierimages/capture.go:30)、[provider.go L27](/Users/cyber/RestoX/barness/ai/examples/classifierimages/provider.go:27)、[bundle.go L58](/Users/cyber/RestoX/barness/ai/examples/classifierimages/bundle.go:58) | 受控Provider与请求/响应捕获，结束写清单及审计。 | 合成凭据不是生产秘密；错误/拒绝同样有完整可重放证据。 |
| [公开进程E2E](/Users/cyber/RestoX/barness/ai/e2e/classifier_images_example_test.go:1) | 实际运行示例命令并检查结果、错误、回滚及产物。 | 例子能执行，而不只是README中的理想片段。 |

## 独立视觉效果工具 对应工单12

“HTTP成功+答案类型正确”不能证明“认得准”。新增visualeval按固定有标注图片另算效果。

| 起读位置 | 内容和解释 | 人工核对重点 |
|---|---|---|
| [data.go L131](/Users/cyber/RestoX/barness/ai/examples/visualeval/data.go:131)、[request L170](/Users/cyber/RestoX/barness/ai/examples/visualeval/data.go:170)、[dataset.json](/Users/cyber/RestoX/barness/ai/examples/visualeval/testdata/dataset.json:1)、[experiment.json](/Users/cyber/RestoX/barness/ai/examples/visualeval/testdata/experiment.json:1) | 固定12样本、任务/图片hash、前8公共可比；Jev后4为local/Box，Decisions后4为supplemental context。 | 后4不是同能力对照；标注、输入图片、ROI/预处理是否相符。 |
| [generate.go L18](/Users/cyber/RestoX/barness/ai/examples/visualeval/testdata/generate.go:18)、[scenes.json](/Users/cyber/RestoX/barness/ai/examples/visualeval/testdata/scenes.json:1)、[pins.json](/Users/cyber/RestoX/barness/ai/examples/visualeval/testdata/pins.json:1) | 生成源图和送入模型的PNG/JPEG，固定几何、场景和hash。 | 24张source/提交图片不能混为24个评测样本；全量文件见逐文件索引。 |
| [authorization.go L37](/Users/cyber/RestoX/barness/ai/examples/visualeval/authorization.go:37)、[main.go L57](/Users/cyber/RestoX/barness/ai/examples/visualeval/main.go:57)、[host policy L38](/Users/cyber/RestoX/barness/ai/examples/visualeval/host.go:38) | 固定官方路线和专门授权，controlled/live/formal分开；有限时间、Calls/Attempts、图片和题数。 | 未授权为NOT_RUN；不能靠改endpoint、型号或candidate目录冒充formal/live。 |
| [run.go L119](/Users/cyber/RestoX/barness/ai/examples/visualeval/run.go:119)、[assess L217](/Users/cyber/RestoX/barness/ai/examples/visualeval/run.go:217)、[finalize L252](/Users/cyber/RestoX/barness/ai/examples/visualeval/run.go:252) | 串行固定样本、预留预算，保存每个请求/响应、结果和Observer；分别标失败、取消、refused和版本漂移。 | 不挑成功样本、不因预算不足将缺样本标COMPLETE；合规执行与分数分开。 |
| [IoU L62](/Users/cyber/RestoX/barness/ai/examples/visualeval/metrics.go:62)、[score L87](/Users/cyber/RestoX/barness/ai/examples/visualeval/metrics.go:87)、[groupMetrics L143](/Users/cyber/RestoX/barness/ai/examples/visualeval/metrics.go:143) | 分类混淆、存在性准确率/Brier/误报、Box IoU、低存在性抑制与raw absent boxes各自统计。 | absent仍返回Box要记录，不能只报消费后误报0；缺答案不能缩小固定分母美化准确率。 |
| [sourceFiles L77](/Users/cyber/RestoX/barness/ai/examples/visualeval/bundle.go:77)、[verify L21](/Users/cyber/RestoX/barness/ai/examples/visualeval/verify.go:21)、[replaySample L337](/Users/cyber/RestoX/barness/ai/examples/visualeval/verify.go:337) | 完整源码/输入/文件集合校验，重算指标并用捕获响应走公开Classify重放。 | 拒绝未列文件、symlink、删capture、假live、篡改预算或分数；随机Jev私有key只在受限重放比较中归一。 |

当前真实结果见[最终效果与门禁记录](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/14-remaining-real/verification.md:84)：两中文24/24，视觉每路线12样本COMPLETE；Jev Box present2/absent2、MeanIoU约0.712374，消费后absent误报0，但raw absent boxes=2。固定小合成任务未设置新的质量通过线，也不能推论生产数据准确率。

## 候选与正式协议探针 对应工单11和13

| 起读位置 | 改了什么 | 人工核对重点 |
|---|---|---|
| [授权配置 L48](/Users/cyber/RestoX/barness/ai/live/multimodal_config_test.go:48)、[Jev candidate](/Users/cyber/RestoX/barness/ai/live/testdata/multimodal/jev.json:1)、[Decisions candidate](/Users/cyber/RestoX/barness/ai/live/testdata/multimodal/decisions.json:1) | 每个进程只选择一条固定路线、型号和候选目录；拒绝其他凭据、endpoint override，九字段授权与12次上限。 | 候选目录不能自行证明正式支持；dedicatedCredential仅代表隔离注入，不证明厂商账户权限。 |
| [场景 L47](/Users/cyber/RestoX/barness/ai/live/multimodal_scenarios_test.go:47)、[断言 L153](/Users/cyber/RestoX/barness/ai/live/multimodal_assertions_test.go:153) | Jev有6类场景×plain/hooked，覆盖默认/显式context、复用、局部图、criteria、Box存在性；Decisions覆盖其自身支持与拒绝边界。 | 检查最终native内容、ImageInputs、结果、usage、Attempts和Observer，不只检查status=200。 |
| [入口 L22](/Users/cyber/RestoX/barness/ai/live/multimodal_harness_test.go:22)、[源码固定 L84](/Users/cyber/RestoX/barness/ai/live/multimodal_harness_test.go:84)、[执行 L212](/Users/cyber/RestoX/barness/ai/live/multimodal_harness_test.go:212) | candidate与formal入口分开；正式用BuiltinCatalog；固定完整ai Go源码实际字节，执行前预算预留、过程限时和资源收尾。 | 源码变化连测试文件也会使formal协议包过期；新文件集合也必须比较，不能只校验旧列表。 |
| [最终传输预算 L19](/Users/cyber/RestoX/barness/ai/live/multimodal_transport_test.go:19)、[终态报告 L96](/Users/cyber/RestoX/barness/ai/live/multimodal_report_test.go:96)、[TestMain L45](/Users/cyber/RestoX/barness/ai/live/main_test.go:45) | 实际发送时再次校对Calls/Attempts/题数/引用等额度；清单、审计完成后发布最终报告。 | 崩溃/未完成审计必须保留未完成事实；不能有report PASS但缺case或capture。 |
| [candidate进程测试](/Users/cyber/RestoX/barness/ai/live/multimodal_process_test.go:20)、[formal语义篡改测试](/Users/cyber/RestoX/barness/ai/live/jev_formal_tamper_test.go:20)、[FIFO拒绝测试](/Users/cyber/RestoX/barness/ai/live/jev_formal_fifo_unix_test.go:15) | 独立进程覆盖wrong key/model、预算、时间、终态、清单和真实输入语义篡改。 | 重新算Hash的恶意包也须拒绝；非regular文件不能阻塞读取。 |

## 独立正式支持封装与矩阵 对应工单13

| 起读位置 | 改了什么 | 人工核对重点 |
|---|---|---|
| [jevformal CLI L28](/Users/cyber/RestoX/barness/ai/live/cmd/jevformal/main.go:28) | 将新正式协议包与独立效果包复制为自包含正式支持包；也支持离线-verify和-require-live。 | 必须新目录；不能只靠协议或另一模型PASS；verify不能同时生产新证据。 |
| [安全文件读取 L23](/Users/cyber/RestoX/barness/ai/internal/testkit/jevformal/files.go:23)、[封存 L134](/Users/cyber/RestoX/barness/ai/internal/testkit/jevformal/files.go:134) | 有界读取/目录遍历、复制、完整文件清单和审计。 | 文件总数/总大小先限额；非regular、父级symlink、额外/缺失文件均需拒绝。 |
| [CheckInputs L102](/Users/cyber/RestoX/barness/ai/internal/testkit/jevformal/verify.go:102)、[ReplayProtocol L88](/Users/cyber/RestoX/barness/ai/internal/testkit/jevformal/replay.go:88) | 以当前BuiltinCatalog和完整源码集合独立验证，复放12个正式场景而非相信提交包自报。 | L148–180严格源码当前性；L203以后独立协议verifier和效果verifier；真实模式必须两包同为live。 |
| [独立身份 L13](/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/jev.go:13)、[NewJevReport L35](/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/jev.go:35)、[VerifyJevReport L62](/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/jev.go:62) | 从已验证包派生schema3报告，新增typesafe-classifier-jev-1.14.0独立路线。 | 1.13继续保持typesafe-classifier身份；逐字比较派生报告和原始capture/预算，controlled不能合并为真实支持。 |
| [JSON读取 L73](/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/file.go:73)、[exactFields L117](/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/file.go:117)、[SaveMatrix L37](/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/file.go:37)、[合并校验 L237](/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/matrix.go:237) | 新Jev报告纳入严格字段/身份检查与矩阵合并，保持原子写入及陈旧报告拒绝；identity.go也纳入新路线。 | 审查模型归属、未知字段、重复键、后续合法重测和所有其他行原样保留。 |
| [当前支持矩阵](/Users/cyber/RestoX/barness/ai/live/support-matrix.json:1)、[ADR0031](/Users/cyber/RestoX/barness/docs/adr/0031-jev-model-specific-formal-evidence.md:1) | 独立正式1.14支持记录，最终更新11条当前真实路线。 | 时间戳/来源是实际本轮包；不能从原先1.13记录继承1.14能力。 |

## 当前目录发布门禁 对应工单14

| 起读位置 | 改了什么 | 人工核对重点 |
|---|---|---|
| [CLI L46](/Users/cyber/RestoX/barness/ai/release/cmd/releasegate/main.go:46)、[完整执行 L111](/Users/cyber/RestoX/barness/ai/release/cmd/releasegate/main.go:111) | 加入独立正式视觉参数、扩展表快照、原压力状态；完整模式执行两种vet、normal、race，再聚合11支持/2文本/2视觉。 | -bundle仅评估旧包，不能当新执行；race保留原场景时限，整套命令允许更长审计时间。 |
| [混合负载消费 L18](/Users/cyber/RestoX/barness/ai/internal/testkit/release/multimodal.go:18)、[最终capture语义 L23](/Users/cyber/RestoX/barness/ai/internal/testkit/release/multimodal_capture.go:23) | 独立消费UC09两入口四路并发产物，核查输入、最终请求/响应、答案/usage、资源和Observer。 | 被重新Hash的错误模型、改答案、缺hooked、漏capture、伪造资源归零都要失败。 |
| [视觉必交 L18](/Users/cyber/RestoX/barness/ai/internal/testkit/release/visual.go:18)、[race必交 L8](/Users/cyber/RestoX/barness/ai/internal/testkit/release/race.go:8)、[live清单 L13](/Users/cyber/RestoX/barness/ai/internal/testkit/release/live_manifest.go:13) | 视觉任务独立于协议与中文；race仍要完整验收；所有声明场景对齐case与审计。 | 60秒有限的独立视觉核验子进程；不能把candidate/controlled或错误目录的真实包当current formal。 |
| [RequiredCases L62](/Users/cyber/RestoX/barness/ai/internal/testkit/release/trace.go:62)、[追溯求值 L170](/Users/cyber/RestoX/barness/ai/internal/testkit/release/trace.go:170)、[固定UC要求 L16](/Users/cyber/RestoX/barness/ai/internal/testkit/release/plan.go:16)、[追溯表](/Users/cyber/RestoX/barness/ai/release/traceability.json:1) | 每条要求对应完整有限case集合；缺一个case不能靠其他一个PASS顶替整个要求。 | 逐条核对R01–R17/T01–T16与映射，尤其root-null、权限、撤销、图例和Box。 |
| [有限扩展 L12](/Users/cyber/RestoX/barness/ai/internal/testkit/release/extensions.go:12)、[扩展case/path登记](/Users/cyber/RestoX/barness/ai/release/classifier-extensions.json:1) | Pi 1.1.0不具备的新图片/Box能力登记为独立扩展，固定规范JSON Hash。 | 这不是扩大Pi差分豁免，不能把新扩展写成差分PASS；原冻结oracle未改。 |
| [原压力工单状态 L13](/Users/cyber/RestoX/barness/ai/internal/testkit/release/blocker.go:13)、[整体求值 L263](/Users/cyber/RestoX/barness/ai/internal/testkit/release/gate.go:263)、[报告 L14](/Users/cyber/RestoX/barness/ai/internal/testkit/release/report.go:14) | 原#07只有自身resolved才解除，结合所有snapshot、命令、证据给整体PASS。 | 集成门禁不擅自关闭原票；保留原票原文Hash；历史FAIL报告不回写。 |
| [目录快照](/Users/cyber/RestoX/barness/ai/release/catalog-snapshot.json:1)、[live聚合 L74](/Users/cyber/RestoX/barness/ai/internal/testkit/release/live.go:74)、[中文效果](/Users/cyber/RestoX/barness/ai/internal/testkit/release/evaluation.go:18) | 新目录及11行支持当前性、独立中文效果和摘要纳入现有门禁。 | 目录Hash变化后的旧包不能继续被选为current。 |

## 公开E2E推荐对照表

测试函数的精确声明行与所有其他测试文件见changes.json中的symbols；逐文件索引列出实际修改的全部行块。以下入口能串起主要边界。

| 工单 | 先看公开测试 | 主要证明 |
|---|---|---|
| 01 | [NativeSnapshot](/Users/cyber/RestoX/barness/ai/e2e/typesafe_native_snapshot_test.go:16)，[NativeGuard](/Users/cyber/RestoX/barness/ai/e2e/typesafe_native_snapshot_test.go:100) | Hook后独立native问题、旧文本行为、封闭结构与结果原子性。 |
| 02 | [ReferencesOrder](/Users/cyber/RestoX/barness/ai/e2e/decisions_references_test.go:25)，[BeforeCopy](/Users/cyber/RestoX/barness/ai/e2e/decisions_references_budgets_test.go:97)，[JSON](/Users/cyber/RestoX/barness/ai/e2e/decisions_references_json_test.go:10) | 排序/重复、显式完整声明、复制前预算、JSON字段精确性。 |
| 03 | [Ownership](/Users/cyber/RestoX/barness/ai/e2e/decisions_image_inputs_test.go:64)，[Hooks](/Users/cyber/RestoX/barness/ai/e2e/decisions_image_inputs_hooks_test.go:23) | 最终图片身份、重排/换图、结果所有权及失败清空。 |
| 04 | [ContextProjection](/Users/cyber/RestoX/barness/ai/e2e/jev_context_test.go:11)，[Preflight](/Users/cyber/RestoX/barness/ai/e2e/jev_context_security_test.go:42) | context信封、asset复用、权限与实验选择、文件及资源边界。 |
| 05 | [StateProjection](/Users/cyber/RestoX/barness/ai/e2e/jev_state_test.go:11)，[StateFinalHooks](/Users/cyber/RestoX/barness/ai/e2e/jev_state_hooks_test.go:13) | 根/嵌套null投影、精确数字、Hook新增同名业务对象不误判。 |
| 06 | [InstructionsProjection](/Users/cyber/RestoX/barness/ai/e2e/jev_instructions_test.go:10)，[RemoveContext](/Users/cyber/RestoX/barness/ai/e2e/jev_instructions_context_hooks_test.go:12) | 题内作用域、混合内容、删除context后保留局部使用。 |
| 07 | [CriteriaProjection](/Users/cyber/RestoX/barness/ai/e2e/jev_criteria_test.go:11)，[LegendFinalHook](/Users/cyber/RestoX/barness/ai/e2e/jev_criteria_legend_test.go:95) | 三类标准图片、图例先验证后恢复、最终索引。 |
| 08 | [BoxVisibility](/Users/cyber/RestoX/barness/ai/e2e/jev_box_scope_test.go:10)，[InvalidAnswers](/Users/cyber/RestoX/barness/ai/e2e/jev_box_answers_test.go:49) | 可见性、缺失/null/0、四边/置信度、错误整批拒绝。 |
| 09 | [DesignLoad](/Users/cyber/RestoX/barness/ai/e2e/multimodal_runtime_load_test.go:33)，[BorrowedAccounting](/Users/cyber/RestoX/barness/ai/e2e/multimodal_runtime_budget_test.go:208)，[Revocation](/Users/cyber/RestoX/barness/ai/e2e/multimodal_runtime_retry_test.go:102) | 两租户×两Provider同时四调用；64MiB分配/增长阈值；重试快照、撤销和Hook复杂编码。 |
| 10 | [Host命令E2E](/Users/cyber/RestoX/barness/ai/e2e/classifier_images_example_test.go:1) | 可运行例子、拒绝/撤销/图例/Box坐标。 |
| 11 | [CandidateProcess](/Users/cyber/RestoX/barness/ai/live/multimodal_process_test.go:20) | 显式授权、有限预算、真实/controlled隔离和失败产物。 |
| 12 | [FixedWorkflow](/Users/cyber/RestoX/barness/ai/e2e/visual_evaluation_test.go:45)，[Integrity](/Users/cyber/RestoX/barness/ai/e2e/visual_evaluation_integrity_test.go:99) | 固定任务评分、篡改/取消/预算/完整源码拒绝，捕获重放。 |
| 13 | [FormalCatalog](/Users/cyber/RestoX/barness/ai/e2e/jev_formal_catalog_test.go:1)，[MatrixHistory](/Users/cyber/RestoX/barness/ai/e2e/jev_formal_matrix_test.go:17)，[Tamper](/Users/cyber/RestoX/barness/ai/live/jev_formal_tamper_test.go:20) | 正式能力、型号独立、合法重测、陈旧报告原子拒绝。 |
| 14 | [CurrentRelease](/Users/cyber/RestoX/barness/ai/e2e/current_catalog_release_test.go:18)，[LoadImport](/Users/cyber/RestoX/barness/ai/e2e/current_release_load_test.go:15)，[VisualCorruption](/Users/cyber/RestoX/barness/ai/e2e/current_release_visual_test.go:36)，[IncompleteTrace](/Users/cyber/RestoX/barness/ai/e2e/current_release_trace_test.go:13) | 完整UC14追溯、两入口负载、独立视觉、历史阻断状态转换和语义篡改拒绝。 |

## 完整门禁暴露的四项续修

| 续修 | 代码起读位置 | 原因、行为变化与回归 |
|---|---|---|
| Decisions NOT_RUN清单 | [notRun分支 L154](/Users/cyber/RestoX/barness/ai/live/decisions_harness_test.go:154)、[formal进程测试](/Users/cyber/RestoX/barness/ai/live/decisions_formal_process_test.go:22) | 原先整体Skip导致9个声明场景没有case；现在逐个记录原因并Skip，仍0调用，不伪造capture。工单[续修02](/Users/cyber/RestoX/barness/.scratch/barness-ai-release-verification/issues/02-decisions-notrun-case-inventory.md:1)。 |
| deadline归属竞态 | [armAttemptTimer L141](/Users/cyber/RestoX/barness/ai/timeouts.go:141)、[公开回归 L17](/Users/cyber/RestoX/barness/ai/e2e/unary_deadline_order_test.go:17) | 较早Call deadline回调延迟，较晚读闲timer误抢终态。现在只启动更早的Attempt限额；两种先后和两入口都测试。工单[续修01](/Users/cyber/RestoX/barness/.scratch/barness-ai-release-verification/issues/01-unary-race-timeout-stop-reason.md:1)。 |
| Darwin受控压力夹具ENOBUFS | [Accept L29](/Users/cyber/RestoX/barness/ai/internal/testkit/provider/write_trace.go:29)、[有界native复现](/Users/cyber/RestoX/barness/.scratch/barness-ai-release-verification/native-pressure/reproduce.py:1) | 默认测试Provider为TCP设置64KiB发送缓冲，靠背压控制内核队列；不是修改生产SDK资源预算，也没有降低32路原设计负载或恢复错误续写。原#07[自身验收记录](/Users/cyber/RestoX/barness/.scratch/barness-ai-thinking-sampling/issues/07-pressure-stream-integrity.md:1)保留自然字节偏离因果。 |
| UC13历史断言与直连配置 | [历史快照 L40](/Users/cyber/RestoX/barness/ai/e2e/jev_formal_matrix_test.go:40)、[configureDecisionsProxy L102](/Users/cyber/RestoX/barness/ai/live/decisions_config_test.go:102) | 一次迁移保留历史与后来合法自身重测分开验证；可选代理为空则明确直连，非空仍严格loopback。对应[续修03](/Users/cyber/RestoX/barness/.scratch/barness-ai-release-verification/issues/03-jev-matrix-history-current-runs.md:1)和[续修04](/Users/cyber/RestoX/barness/.scratch/barness-ai-release-verification/issues/04-decisions-direct-transport.md:1)。 |

FIFO首次2秒超时根因没有内部栈证明，旧FAIL仍保留；同源码独立用例及完整harness复核通过。DeepSeek空reasoning的原FAIL也保留，新完整路线确认13次后才选新包，未修改Go或断言。

## 真实执行、预算及封存脚本

| 起读位置 | 责任 | 人工核对重点 |
|---|---|---|
| [首阶段选定secret L106](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/14-typesafe-real/run.py:106)、[旧轮封存 L269](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/14-typesafe-real/seal.py:269) | 单Provider两轮各51次的历史执行及保留；第二轮在UC13源码修复后重采。 | 旧失败normal、中止race、旧整体false没有追认通过。 |
| [源码guard L65](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/14-remaining-real/run.py:65)、[选定变量 L75](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/14-remaining-real/run.py:75)、[隔离执行 L101](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/14-remaining-real/run.py:101) | 按用户选定的.env变量逐行读取，单进程单key；固定源码，有限预算、intent和进程组收尾。 | 不source dotenv、不继承其他key/ambient proxy、不打印秘密、不自动重发收费任务。 |
| [恢复guard L26](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/14-remaining-real/recover.py:26)、[固定jobs L83](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/14-remaining-real/recover.py:83)、[一次路线确认 L16](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/14-remaining-real/recheck.py:16) | 源码变化后一次完整恢复；同源码一次DeepSeek完整确认，均在Provider与任务原上限内。 | 312=120+179+13；550是上限，不能为凑满继续调用；原失败消耗仍计总数。 |
| [原始消耗 L103](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/14-remaining-real/seal.py:103)、[实际门禁 L137](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/14-remaining-real/seal.py:137)、[无调用verify L292](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/14-remaining-real/seal.py:292) | 从原始报告重算预算、完整producer集合、Git对象及源码路径、清单/树和实际命令退出码。 | 保留636346条目含失败；独立verify实际0；这里只核验字节和证据，不能替代生产重跑。 |
| [授权清单](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/14-live-authorization-checklist.md:1)、[本轮有限授权](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/14-remaining-real/authorization.md:1)、[最终双轴审查](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/14-remaining-real/code-review.md:1) | 变量名、Provider额度、unknown元数据、持续授权边界与审查结果。 | Calls/Attempts不是美元熔断；账户/区域/厂商权限内省依然未知，不伪造spendLimitUSD。 |

其他02–14的verify/seal脚本、原始RED/GREEN日志、工单和回执已逐一列在files-records.md；它们是证据生产/消费层，没有进入运行库公共API。
