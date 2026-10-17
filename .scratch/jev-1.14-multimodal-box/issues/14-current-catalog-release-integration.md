# 14：完成当前目录下的整体发布验收与追溯

Status: claimed

**What to build:** 维护者取得当前目录/源码下完整且可追溯的验收结论，能明确区分本功能完成、历史阻断和是否可整体发布。

**Blocked by:** 09：验证混合图片调用的并发、资源和撤销；13：纳入 Jev 1.14 目录并建立独立正式支持记录。

## Acceptance criteria

- [x] 在修改实现前列出本票所有失败方式，先写经 Client.Classify / HookedClient.Classify 的公开 E2E 并确认红灯，再实施；不在实现后补单元测试。
- [x] 建立每条规格要求/用户故事/验收组到当前有效测试与证据的追溯，运行既有质量门禁、完整 normal/race、新图片混合负载及两公开入口回归。
- [x] 保持唯一冻结 Pi 基线；原 Jev/Decisions 投影照常比较，新引用/manifest/Box/严格校验以明确有限 case/path 登记扩展，不能扩大豁免或冒充差分 PASS。
- [ ] 正式目录 Hash 变化后按既有各路线自身有限预算和独立进程重新形成其当前所需真实/效果证据；不复用历史 Hash、候选或另一模型 PASS。
- [x] 支持矩阵可以独立核对 Jev 1.13/1.14 和 Decisions，不以新增模型替换旧型号身份；必交能力不能写 UNSUPPORTED 绕过。
- [ ] 所有包清单、审计、源码/配置/目录一致性、预算预留/使用、Observer 脱敏、结果原子性与资源释放完整；失败/NOT_RUN 保持真实结论。
- [x] 公开迁移、实验选择、格式/位置拒绝、图例/索引解释和版本化撤销说明与可执行例子一致；确认无公共 Assets、旧内部分支或兼容 shim。
- [x] 已有压力流完整性工单的验收未满足时，总体发布不能 PASS；分别报告新增功能已验证范围与外部阻断，不擅改/关闭原工单或父规格。
- [x] 仅在全部合同确实满足时更新相应交付记录；不执行未授权部署，不因预算耗尽、单次重跑或文档齐全而宣称完成。
- [x] 结束时产出可重复、可核验的图片/fixture hash、输入、最终请求/响应、结果/错误、Attempt/Observer、断言、配置、完整性清单及脱敏审计；仅受控 HTTP 结果不得声明真实服务 PASS。

## Traceability

- 规格要求：R01、R02、R03、R04、R05、R06、R07、R08、R09、R10、R11、R12、R13、R14、R15、R16、R17。
- 验收组：T01、T02、T03、T04、T05、T06、T07、T08、T09、T10、T11、T12、T13、T14、T15、T16。
- 用户故事：01、03、17、23、25、32、33、41、42、43、44、45、46、47、48、49、50。

## Scope and hand-off

- 本工单的拆分已获用户确认并发布；执行时须先满足全部阻塞依赖。本次发布不代表已实现、已验证或取得部署授权。
- 单票只声明其已经打通的能力，其余新能力仍明确拒绝；授权/容量/Hook 最终复核随功能同时交付，不推迟到整合票。
- 复用现有运行时、测试记录器、图片和 JSON 能力；新增依赖前核查已有依赖/类型定义，不创造无真实第二用途的公共接口。
- 执行本工单中的真实验证时，只使用显式获准的路线/凭据及有限预算，环境不足记录 NOT_RUN，不借其他任务凭据或无界重试。
- 外部发布条件：barness-ai-thinking-sampling #07：定位当前发布压力流完整性失败（总体 PASS 前须满足其既有验收；不代为关闭）；该条件影响总体 PASS，不伪装成无关实现票的技术依赖。

## Comments

2026-10-16：用户确认拆分，按已批准的标题、交付和依赖发布；尚无实现或验证执行记录。


2026-10-17：当前 main 已提交发布整合及审查修正（最后 Go 源码提交 `4310a53`）。
公开 Classify/CLI 专项 33 个用例通过；包括两正式视觉路线、语义篡改拒绝、完整追溯、
设计负载消费和原 #07 状态转换。Standards / Spec 独立复核均无剩余代码发现。
完整 normal/race 各执行 9,094 用例，均为 9,093 PASS、1 FAIL；两阶段 UC14 33/33、
图片设计负载两入口及各 861 个 Pi 差分通过，主包审计均 0 发现。实际结果见
`../14-verification.md`、`../14-evidence.json`；中间失败与主动中止均保留。

本票保持 needs-info、未 resolved：本次没有获准路线凭据与收费预算，11 路线及各自
文本/视觉效果预检均 NOT_RUN、真实 Calls/Attempts 为 0；原压力流完整性 #07 的
既有因果验收也未满足。需要各路线明确授权和有限预算下的当前源码/目录证据，并由
原 #07 自身完成既有验收后再运行完整门禁。未修改或关闭原工单、父规格/ADR，也未部署。

本次完整门禁还保留 normal cloud 压力 ENOBUFS、race P07 stop reason 不一致，及
Decisions NOT_RUN 声明场景缺少清单 case 的拒绝。后两项按 needs-triage 记录在
`.scratch/barness-ai-release-verification/issues/`，没有擅改原验收或放宽校验。
因此“所有包清单/审计一致”仍未勾选，真实/效果证据条件也未勾选；工单不宣称完成。


2026-10-17 续修：用户授权修复Decisions清单与normal/race失败，并按原#07验收定位压力。
完整新执行f29703b（Go内容cc78cfd）normal/race各9,098PASS、0FAIL，两种vet、P0、
差分、聚合脱敏门禁PASS，17个输入审计零发现；各33个UC14及861个Pi比较通过。
Decisions逐场景NOT_RUN清单已修复并被严格消费者接受，unary取消归属与受控夹具缓冲已修复。
原#07在重复自然因果反馈、回归和本版完整受控验收后由其自身关闭；状态前后原文/Hash和
验收后转换单独保留，启动时未关闭的旧snapshot及原report.pass=false均不改写。
真实11路线及4效果包仍NOT_RUN、Calls/Attempts0，本票保持needs-info，真实/效果证据
及整体完整性验收项仍不宣称全部完成。详见
[续修验收](../../barness-ai-release-verification/verification.md)与
[回执](../../barness-ai-release-verification/evidence.json)。

2026-10-17 TypeSafe真实验证：用户明确指定根`.env:TYPESAFE_KEY`，按清单仅执行该Provider
四项一轮。Jev1.13协议3场景PASS（3次）、中文24样本COMPLETE且24/24正确；
Jev1.14正式协议12场景PASS、正式视觉12样本COMPLETE，两项均严格require-live核验。
总计51 Calls/51 Attempts、零重试，低于原54次上限；正式封装、效果清单、审计和
支持矩阵两条独立身份均已核对，其他9矩阵行不变。未读取或注入其他Provider凭据。
账户/区域/厂商权限仍未内省，没有编造美元硬上限或权限验证。
其他9条支持、2项效果仍NOT_RUN，本票继续needs-info，原两项整体验收不提前勾选。
完整无凭据门禁normal出现唯一UC13历史矩阵测试FAIL（9,097 PASS、1 FAIL），
该断言错误拒绝后续合法重测，单例复现和独立Spec核对均已完成；race继续收齐。
本轮结果、续修03和剩余248次待授权额度见
[TypeSafe验证](../14-typesafe-real/verification.md)。

2026-10-17追加：用户明确授权修复UC13测试并用同一key再跑四项一轮、最多54次。
旧门禁在已有normal确定失败后主动中止，race未完成，不追认PASS。`541ef2d`修复
一次迁移与后续真实重测的边界，原before未改，after取自首次纳入提交；公开CLI
验证1.13仅更新自身、其他十行完整保留、陈旧报告原子拒绝，单例RED/GREEN已保留。
修复后第二轮真实四项再获51 Calls/Attempts、零重试，严格verifier/正式封装均通过，
矩阵提交`4d191fa`，两轮累计102次。新源码当前Box平均IoU为0.6840596383，
不使用首轮较高分替换；独立两轴复核无阻断发现。第二轮完整无凭据门禁仍在运行，
真实缺口与两项整体验收继续保持未完成，其他Provider仍NOT_RUN。

2026-10-17最终收尾：第二轮完整normal/race各9098 PASS、0 FAIL，每阶段861Pi与33UC14
均PASS；两种vet、P0、差分、快照和聚合脱敏七项内部门禁PASS，17个输入审计零发现。
UC13矩阵历史续修03按自身验收关闭；原#07以resolved当前快照通过，旧失败及中止未改写。
实际发布报告pass=false、CLI退出1，剩余live/traceability/required-artifacts三项阻断均来自
未获授权的其他9条真实支持和中文Luna/正式视觉Decisions两项效果。TypeSafe本轮缺口已验证，
本票继续needs-info，两项整体验收保持未勾选，不更新父规格/ADR完成状态、不部署。

2026-10-17 新授权：用户允许按根`.env`现有Provider key执行剩余9路线和2项效果直到
整票通过，只报告变量名。`OPENAI_KEY`、`ANTHROPIC_KEY`、`GEMINI_KEY`、
`DEEPSEEK_KEY`、`TYPESAFE_KEY`可用；账户/区域/厂商权限仍UNVERIFIED，金额硬限未知。
本轮初248、最多一次修复后全量302恢复、合计550 Calls/Attempts上限。
初轮实际120次、零重试：8支持及2效果完成；Decisions可选空代理配置缺陷0调用FAIL，
由续修04按三个公开进程E2E先RED后修复，实际旧失败保持原样。
当前本票进入claimed，已固定恢复源码并先跑无凭据harness；仅在全部当前真实/效果、
清单审计及完整发布门禁通过后完成剩余两项验收。详见`../14-remaining-real/authorization.md`。
