# MiniMax 国际 Messages：只补 thinking-on、复用原13项

日期：2026-10-28。范围为设计与只读核验，尚未实施，未获付费执行授权。建议新增**1次逻辑调用、1次HTTP Attempt、0次重试/场景rerun，拟议USD0.01上限**。原13项不重买；原失败包保持FAIL。数据与精确原件指针见同名.evidence的scope-and-budget-v2.json。

## 1. 已有事实与最小缺口

原包为`.evidence/barness-ai-grok/live/minimax-messages/20261027T000249.194767Z-krdog6id/bundles/20261027T000249.211139000Z`，原formal/live、complete=false、producer exit1。15条Call记录、14个唯一场景中，thinking-control首次失败已有合法rerun恢复，唯一未恢复的是thinking-on。旧失败和已消费批准不改写。

[前轮13项适用性](../provider-final-readiness/status.md)已验证source/SDK/实际M3目录投影、全包seal/audit、原授权、权威ledger和逐项语义；[当前最小缺口](../provider-final-readiness/minimal-capture.md)不授予执行权限。新消费者实现后须再离线消费这些原件，不能把分析overlay当生产资格入口。

身份固定：`minimax-messages / minimax / anthropic-messages / MiniMax-M3 / standard`，官方端点`https://api.minimax.io/anthropic`；实际POST为`/v1/messages?beta=true`。原账户/地区别名为MINIMAX/global，tenant=live-smoke、accountScope=live-minimax-messages；未来Host必须确认真实账户仍与原证据一致，别名本身不是账户证明。仅standard API，不转M Plan、CN或M3.1。

原Host Request紧凑JSON为308字节，**实际HTTP正文389字节**；此前清单的308指Host输入，不是wire。按最终wire的`max(2048, bytes+1024)`仍只需预留2048。原793字节SSE前缀已收到非空Thinking，随后读取失败；底层read error未留存，不能断言SDK/代理/厂商/网络谁有缺陷。

## 2. 一个固定单场景入口

新增精确入口`^TestMiniMaxThinkingOnSupplement$`，`-test.count=1 -test.timeout=180s`。唯一profile为`minimax-messages-thinking-on/v1`；由**完整精确校验后的grant.scope.scenes=[thinking-on]**确定；profile只是固定CLI/报告标签，不新增Scope.Profile字段，不新增任意场景环境变量或子测试过滤入口。

保留`Identity.Purpose=standard`的账户语义，profile仅表示验收范围。复用现有M3 CRT任务、stream=true、thinking.type=adaptive、max_tokens=2048，不简化问题、不追加热身/健康厂商调用，不提高输出。当前普通受控证明已经能重现相同389字节wire。

实施时复用`minimaxSession.execute`和现有recorder/gate：

1. 验证固定身份、candidate阶段、单项单项场景形状/baseline pin、冻结source/build/binary/SDK/目录/价格、账户和新授权；精确入口/count/timeout不符在claim/HTTP前拒绝。只读取已选服务的Host注入凭据。
2. 按现有Claim永久占用新ID/scope；使用该claim下独立ledger。预留一次2048输入和2048输出；不选择旧ledger/renewal/round。
3. 解码后的场景形状只用于选择固定校验器，必须完整Check通过后才确定计划和预算；未知集合拒绝。在Ledger.Reserve追加前检查本scope逻辑调用/累计input/output上限；在HTTP gate追加Attempt前检查最终wire、USD和Attempt上限。最终wire超出2048预留时零发送拒绝，不涨cap、不记第二次reservation。
4. SDK初始请求重试、环境重试、tool-capability rerun、RequestRerunnable初始请求rerun均为0；即便服务未返回HTTP、429/5xx或流内中断，也不追加调用。首次失败/取消/超时即停止；不增加等待、退避或账户调度器。
5. 调用最多120s，进程180s，内部deadline为进程上限减现有20s收尾余量并受授权有效期约束。沿用安全错误/HTTP/原始响应诊断，不为补回旧read error重构公共日志/recorder。

新报告保留真实candidate/live或controlled身份，`kind=minimax-thinking-on-supplement`，Expected/Calls仅一项。`outcome=PASS`只表示所选单项成功；**complete始终false**，不写完整14项live-report，不填13个虚构PASS/NOT_RUN。现有完整VerifyCapture仍拒绝单项或原失败包。

## 3. 独立授权与预算

| 项目 | 拟议单次上限/约束 |
| --- | --- |
| 路线/模型/账户用途 | 国际minimax-messages / MiniMax-M3 / standard |
| 场景 | thinking-on，保留原CRT请求和adaptive thinking |
| 逻辑调用 / HTTP Attempt | 1 / 1 |
| SDK / 环境 / 场景rerun | 0 / 0 / 0 |
| 最终wire输入预留 | max(2048, wire bytes+1024)，累计≤2048 |
| 单次及累计requested output | 2048，含推理及可见输出，不重复加预算 |
| call / process | 120s / 180s；count=1，无额外首等要求 |
| 拟议USD | 0.01；仅本路线单项，不与Kimi USD0.60合并借用 |
| 有效期 | 执行前另签，建议30分钟，现有最长24小时不变 |

价格依据仓库冻结`minimax-global-m3-standard-2026-10-18-direct-usd`和ADR-0042，不是实时报价。M3低阶USD/M输入0.30、输出1.20，最高已知阶输入0.60、输出2.40：

- 2048输入+2048输出按低阶为`USD0.003072`。
- 保守取最高已知阶，完整预留`2048*0.60/1e6 + 2048*2.40/1e6 = USD0.006144`，低于拟议上限USD0.01。
- 缓存命中不降低前置预留；未知缓存写价不按0免费处理，继续沿用正写价未知拒绝资格规则。未知/不完整用量、失败/取消仍保留全额预留与已消费批准，不能以费用下界释放权限。

这是Host按冻结已知费率的保守预留，不是厂商账单硬封顶；单调用token/HTTP约束与失败停止实际执行。未来授权前须核对实际standard账户、型号/服务等级/价格/输入仍适用；价格、未知收费项或payload越界时先离线重估，不自动提高预算、换priority或涨token。若本次再次截断/断流，没有第二次请求额度。

复用现有`livegrant.Grant`、Claim/Consume和`.git/barness-live-process-grants`的ID/scope永久占用，不新建批准平台/账本/跨轮权限。最小必要数据变化是在既有Scope末尾追加两个omitempty字段：`BaselineManifestSHA256`、`MaxUSD`；既有Scenes字段已经绑定单项，不重复增加Profile字段。它们全部进入现有规范scope哈希；MiniMax专用构造器逐字段固定1/1/0/2048/2048/180/0.01及原baseline SHA。

旧字段零值仍省略，保留旧JSON顺序/keys，原grant/claim字节不变；实施必须用真实旧批准及原时刻验证该性质。通用层只做有界结构校验：process时限由固定1800改为合法正值≤1800，VerifyEvidence按已验证scope时限核对。每个原路线的scope构造仍精确1800及原预算，不能由这个结构变化缩短/扩大其批准；MiniMax新构造器只接受上述固定profile。不给通用层增加场景注册表或厂商分派。

未来processgrant仅增加固定`-profile minimax-messages-thinking-on/v1`与原`-baseline-manifest-sha256`的合法MiniMax模板组合；Kimi的purpose合同独立，source模式保持只读/排他。普通14项grant与单项grant不能交叉启动。**当前本轮不生成可执行模板、不签ID/expiry、不claim、不读凭据。**

## 4. 固定13+1只读组合

每次新建一个仅含`minimax-thinking-composition.json`的目录，类型固定`minimax-messages-thinking-composition`。仅两个非嵌套引用：baseline为整份旧失败bundle，supplement为未来单项bundle；分别pin manifest/report/seal和真实producer回执SHA。不复制胶囊、不改原报告、不追溯恢复轮次、不按成功结果自动挑多个包。

MiniMax专用接口拟为`supportmatrix.VerifyMiniMaxThinkingComposition(dir, formal)`。身份/原件/授权/ledger/逐项语义留在minimaxlive，组合与当前资格映射留在supportmatrix；releasegate沿用既有inspect/live-consumer/record，只增加专用分派。Kimi保留自己的合同，不引入通用N+M组合引擎。

baseline检查必须是明确的13项贡献合同，**不能吞掉VerifyCapture错误后随意捞PASS行**：

- 重新核验整个旧包的seal/audit、所有15条Call、source/SDK/所用型号、原时刻授权、真实ledger和producer原件，保留原预算及原FAIL/exit1。
- 14项标准Plan保持不变；13个贡献场景固定为Plan减thinking-on。唯一未恢复失败只允许thinking-on；thinking-control首次错误及相邻rerun-1按原规则核验并保留，贡献选择其已成功的case。额外失败、未知/缺失/重复场景拒绝。
- 原失败thinking-on保留在库存与excludedFailures中，没有覆盖贡献；无成功终态的前缀不能充当新一项。原13项也不改写成一次新13-call测量。
- 复用现有逐项检查，独立核对当前请求编码及响应解释；如涉及请求重建，工具历史只从本包实际tool-call Result及同tenant/account封套恢复。xAI历史仍按既有固定夹具/降级合同；取消只验证已观察前缀与原时序，预检保持零Attempt，不能用捕获payload覆盖掩盖当前编码差异。

supplement检查固定profile、base manifest pin、mode/live或controlled、candidate、单项成功及1/1/零重试/金额/时限、当前wire/响应/Result/Observer/用量/资源、实际退出0和独立claim/ledger；单独不给完整资格。

旧producer应证明**原失败退出1**而不是强求PASS/0；新producer必须成功/0。核对原真实argv，包括旧`env -C ... <binary>`和新精确入口/count/180s、原执行二进制、日志/数字退出/source.files+selectedInputs（两域若重复必须一致）。不能改旧回执满足新入口约束。

两个贡献集合必须不交叠且并集严格等于当前14项。输出是新当前资格：保存两次measurement原source/commit/time/stage/budget/exit，14个scene到各自bundle/manifest/case/artifact SHA的映射，以及原排除失败；没有伪造统一测量时间、统一grant或“14项同进程PASS”。相同caseID由measurement和manifest区分，旧失败MM07-thinking-on不会被覆写。

同路线/Provider/API/端点/model/standard用途、账户/地区/tenant/account scope、SDK和相关型号事实必须相容；未审产品或SDK变化单列受影响场景。旧formal+新candidate可以作当前正式资格输入，但两份均须真实live，当前M3正式型号事实匹配，原stage不改；两个measurement的mode必须一致：live+live才可取得真实资格，controlled+controlled只能得到机制结论；混合mode或replay输入拒绝。无需额外整组formal实测。

外层继续`consume_live.py` schema2采集一个releasegate进程。输入闭包覆盖描述符、两包**完整**库存（包括旧失败和control初次错误）、两producer/日志/数字退出/二进制、原source/selected-input字节证明与可信ID/scope/ledger。账本位置来自受信Git common dir，不能由描述符提供替代品。

发布record的live阶段只引用新单项实际producer成功回执；14项必需case通过组合coverage证明，旧exit1作为baseline历史事实保留，不冒充新阶段PASS。只读发布CLI不再嵌套启动资格命令。

## 5. 最小实现范围

| 位置 | 必要变化 |
| --- | --- |
| minimaxlive新增小的thinking profile文件、process_scope.go | 固定Messages/M3/standard/candidate、一个Plan、Scope两个可选字段及精确caps、base pin、USD纯计算；账户purpose仍standard |
| livegrant/grant.go、evidence.go | 两个可选范围字段进入规范哈希、受控duration结构边界；原schema1/keys保持；不改Claim/Consume算法和存储 |
| ai/live/minimax_config、process_grant、main、harness/finish | 校验后选固定profile、精确入口/count/180s、内部deadline、零rerun、单项complete=false；正常14项入口不放宽 |
| minimaxlive/ledger、ai/live/minimax_capture | 从已验证scope约束追加reserve/Attempt前的上限与最终wire/USD；复用原锁/append/sync/链，单请求USD记录入case预算原件，不加USD累计账本 |
| minimaxlive/verify、verify_call及专用组合小文件 | 抽取已有完整性/源/ledger检查供固定13项及1项使用；只新增受限profile，不放宽完整VerifyCapture。新预算artifact仅在单项profile允许 |
| supportmatrix及release现有inspect/live_inputs/consumer/record分派 | 一个MiniMax专用组合和传递输入闭包，保留两次原测量；可复用成熟文件/哈希/退出采集原语，不把厂商失败语义泛化 |
| processgrant | 只增加固定profile+baseline模板参数，继续不授予权限，无新CLI/脚本控制器 |
| ADR-0055、测试SOP及MiniMax工具说明 | 实施前补允许旧失败测量中13项贡献与新1项的受限规则；本设计不修改已接受ADR或当前门禁 |

harness当前493行，实施时把新固定profile/检查放在独立小文件，并把必要既有场景构建移到合适模块，所有文件≤500行。不能把大if块继续塞入原文件。

不改公共Client、协议adapter、SDK、模型目录或原价格；不增加任意case列表、组合DAG、插件注册器、调度器、账户节流服务、新claim算法、恢复控制器或历史兼容层。standard14项及原grant合同保留，不新增fallback。

## 6. 先失败再实现的验证清单

未来实施先公共Client/CLI/子进程RED，再实现并GREEN：

1. 真实13项逐项复用、一个受控新项组成14项；原包仍FAIL、旧exit1不改。新单包1项/旧13项均拒绝完整资格，受控组合不升级真实。
2. 用真实布局的source.files+selectedInputs、旧env启动argv和旧到期时间验证producer；缺原件/错binary/数字退出/来源或两域冲突均拒绝。
3. profile/scene/model/Provider/标准用途/账户/地区/SDK/source/binary/基包SHA/entry/count/timeout/expiry/USD任一错误发送前拒绝；full与single授权不能交叉启动。
4. 一次实际HTTP，禁止初次请求、工具或流内rerun；第二次reserve/Attempt、超输入/输出/USD在追加/发送前拒绝，取消和失败仍消耗claim。
5. 复用已留793字节断流/EOF/完整控制，检查当前wire、Thinking、成功文本/终态、usage/费用、Observer和资源；截断/未知写价/缺用量/协议错误均不能认证成功。
6. 跨包tool前序、跨tenant/account、重复/缺项、原失败伪装PASS、任意重封/额外文件、路径逃逸/symlink、传递库存或可信ledger改动均拒绝。
7. 复用现有controlled全14入口，形成13项成功、thinking-on稳定受控失败且原生真实exit1的完整baseline；再启动独立单项成功/exit0进程，组合验证14项。可追加control初错后合法恢复的受控脚本。无需另建13项采集器或fixture标签，不得删行/删文件伪造baseline。
8. 公共既有路线授权/消费者回归必须证明新增omitempty字段不改变旧canonical scope/claim key；各旧scope仍为原1800和精确预算。

随后跑完整受影响普通包、相关CLI/工具过程及vet，冻结真实source/input/binary；所有合成grant/claim写前断言私有Git namespace并核对公共存储前后无变化。race/压力按实际claim接入/取消影响列人工范围，不由实施批准启动。

## 7. 避免牵动其他Provider

原13项仅因实际产品/SDK/相关型号改变才重评；消费者改动优先离线解释完整原件。没有新的厂商事实缺口就不重买其他路线，不恢复Grok17路批次。

实现可能改变共享livegrant/原生来源哈希：这需要在最终冻结后重建受影响采集二进制，并更新尚未批准的Kimi六份机器范围引用。**重建和重新生成待批准提案不等于旧54项或其他Provider资格失效**，不刷新旧grant/expiry/账本，也不重跑整组付费验证。Kimi自身9+1语义、6次/USD0.60范围独立。

完成条件依次为：离线实现/受控普通通过→冻结单项机器范围→人另行明确批准一次/USD0.01和自身账户→签一个新grant→只采thinking-on→只读13+1取得当前资格。任一步失败保留原件并停；失败后的新执行需要独立新批准。目录无新增推广需求，整体race/压力/发布仍独立收口。

## 8. 本轮边界

本轮仅设计、预算算术、原件/引用与来源/存储只读核验。没有实现上述模块，没有生成未来supplement引用、可执行scope/grant或审批有效期，没有读取凭据、claim、厂商请求、race、压力、目录激活、暂停恢复或清理。实施普通、真实单项、组合正式资格均为NOT_RUN；既有普通/13项适用性保留原引用。
