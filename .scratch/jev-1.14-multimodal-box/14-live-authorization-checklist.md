# 工单14真实验证缺口与授权预算清单

## 最新执行状态（2026-10-17）

用户已允许按根`.env`现有key真实调用直到整票通过。可用变量名按Provider为：
OpenAI `OPENAI_KEY`，Anthropic `ANTHROPIC_KEY`，Google `GEMINI_KEY`，
DeepSeek `DEEPSEEK_KEY`，TypeSafe `TYPESAFE_KEY`。不记录任何值、长度、前后缀或指纹。
本次总上限550 Calls/Attempts；Provider上限分别232、64、72、128、54。
先初轮248上限，再唯一全量恢复302上限；已用312次包含一次≤32次的同源码
DeepSeek Chat完整确认。没有第三个全量轮，不再自动补跑，之前TypeSafe102次属独立历史授权。

当前选定11支持及4效果包均通过严格核验，矩阵已整批原子更新。初轮Decisions空代理
0调用配置FAIL及恢复轮DeepSeek空推理12次FAIL保持原包，修复/确认使用新包。
当前Go内容`9a8ea45`；实际报告提交`fc8f02b`、pass=true、CLI退出0，10项门禁全PASS。
normal/race均9098 PASS、0 FAIL，各861 Pi及33 UC14；17个输入审计零发现，#14已resolved。
未出现仍需补充才能执行当前必交调用的授权阻断；账户/区域/privilege及金额硬限内省仍未知，
不编造Host permissions或美元熔断。完整授权、按Provider实际用量、错误原包及最终结论见
[本轮验收](14-remaining-real/verification.md)和[授权](14-remaining-real/authorization.md)。
以下是首次清单、历史续记和保持不变的各路线合同，不代表仍缺少本轮授权。

## 初次核对与历史续记

2026-10-17，核对基准 `8e13d23`，Go 内容为 `cc78cfd` 的完整600文件集合。
当前正式目录版本 `2026-10-17.1`，Hash 为
`sha256:ae3315ce504af2019ce6476f3b021093e13428a5992b3e5fc7d3208576eb49c7`。
本清单初次整理时只记录执行条件，真实支持与效果均为 NOT_RUN、零真实调用。
下列完整路线和预算仍是固定合同；之后的获准执行单独记录，不扩大原范围。

2026-10-17 后续用户指定根 `.env:TYPESAFE_KEY`，TypeSafe 四项已完成一轮：
Jev1.13协议3、中文24、Jev1.14正式协议12、正式视觉12，共51 Calls/Attempts、零重试，
低于该Provider原54次上限。两条正式支持PASS、两项效果COMPLETE，其他9条支持与
2项效果仍NOT_RUN；剩余原额度合计248次，尚未取得其他凭据及执行授权。
账户/区域/厂商权限仍UNVERIFIED，没有提供美元硬上限。详见
[本轮验证](14-typesafe-real/verification.md)及[授权边界](14-typesafe-real/authorization.md)。

后续normal发现UC13历史矩阵测试缺陷。用户明确追加同一key四项一轮、最多54次，
修复提交`541ef2d`后第二轮已取得51/51、零重试的当前证据；两轮累计102/102，
首轮不作为新源码的当前验收。旧normal失败和race中止事实保留；第二轮完整normal/race
各9098 PASS、0 FAIL，七项内部门禁PASS，17个输入审计零发现。原UC13缺陷已通过验收关闭。
整体仍BLOCKED（CLI退出1），因为下述其他真实路线和效果仍未运行。
其他9条支持及2项效果仍未获准执行，待授权原额度仍为248次，不从累计调用数扣算。

## 初次核对时的剩余验收（历史）

[#14](issues/14-current-catalog-release-integration.md)仍有两项未勾选；TypeSafe部分已执行，
以下全合同还需其他9条支持与2项效果的当前真实包：

| 缺口 | 完成所需证据 |
| --- | --- |
| 当前目录及源码下各路线真实支持和效果 | 11个独立正式支持输入，另有2文本、2正式视觉效果包；旧Hash、候选、其他型号和controlled包不能替代 |
| 实际执行包的完整性和一致性 | 对每包核对源码/配置/目录/固定输入、预算预留及实际使用、捕获重放、Attempt/Observer、资源归零、清单和审计；实际服务故障保留FAIL |

初次核对基准的normal/race各9,098 PASS、0 FAIL，两种vet、P0、861个冻结Pi差分及输入审计均通过。
原压力#07及续修01/02已完成验收并关闭，不再是当前待修缺陷。
[最新完整验收](../barness-ai-release-verification/verification.md)和
[封存回执](../barness-ai-release-verification/evidence.json)保持原样。

旧发布报告仍为pass=false，并保留启动时#07未关闭的snapshot。这是历史快照；
取得当前真实包、合并支持矩阵后，仍须在新目录重跑完整发布命令，产生新的总体结论。
#14继续needs-info；父规格和ADR的交付状态只能依据新的完整结果更新。

首轮真实矩阵更新后，完整门禁normal出现UC13历史矩阵测试FAIL（其余9,097 PASS），
race在用户追加授权后主动中止。测试把一次迁移的历史保留断言施加于后续合法重测；原失败保留，
[续修03](../barness-ai-release-verification/issues/03-jev-matrix-history-current-runs.md)记录诊断和源码变化对真实包有效性的影响。

## 正式支持路线

以下是仓库当前固定合同中的型号和上限，不代表实时服务可用性或实际账户权限。
表中Call上限均为每路线一次独立进程，故障与允许的重试也消耗同一额度。

| 路线 | 必须覆盖的型号 | Call上限 | 重试 | 授权配置 |
| --- | --- | ---: | --- | --- |
| openai-responses | gpt-5-mini、gpt-5.4、gpt-6-sol | 32 | 每场景最多1次环境故障重试 | A |
| anthropic-messages | claude-haiku-4-5、claude-opus-4-8、claude-fable-5、claude-fable-5-1、claude-opus-5、claude-opus-5-5 | 32 | 每场景最多1次 | A |
| google-gemini | gemini-3.8-flash | 32 | 每场景最多1次 | A |
| openai-chat | gpt-5-mini、gpt-5.4、gpt-6-sol；另有gpt-5.4-pro预期拒绝探针 | 32 | 每场景最多1次 | A |
| deepseek-responses | deepseek-flash | 32 | 每场景最多1次 | A |
| deepseek-chat | deepseek-flash、deepseek-v4-pro | 32 | 每场景最多1次 | A |
| typesafe-classifier | jev-1.13.0 | 6 | 每场景最多1次环境故障重试，计入6次 | A |
| openai-images | gpt-image-2.5-sunburst-2026-09-08 | 4 | 每场景最多1次环境故障重试，计入4次 | A |
| google-interactions-image | gemini-nano-banana-2.1 | 4 | 每场景最多1次环境故障重试，计入4次 | A |
| openai-decisions | gpt-6-luna | 12 | 0 | B |
| typesafe-classifier-jev-1.14.0 | jev-1.14.0，协议12场景覆盖plain/hooked | 12 | 0 | C |

A路线的官方Endpoint和API如下；仅允许所选服务、型号和原场景，不自动替换型号或跨路线fallback。
聊天场景包括文本stream/complete、工具往返、取消、推理历史、图片输入及各自附加场景。
既有明确可选能力按原消费者规则处理；必交能力不得新标UNSUPPORTED来跳过。

| 路线 | Endpoint | API |
| --- | --- | --- |
| openai-responses / openai-chat / openai-images | https://api.openai.com/v1 | openai-responses / openai-completions / openai-images |
| anthropic-messages | https://api.anthropic.com | anthropic-messages |
| google-gemini / google-interactions-image | https://generativelanguage.googleapis.com/v1beta | google-generative-ai / google-interactions |
| deepseek-responses / deepseek-chat | https://api.deepseek.com | openai-responses / openai-completions |
| typesafe-classifier | https://api.typesafe.ai/v1 | typesafe-system-one |

A类授权需要按路线明确账户/区域别名、专用凭据来源及所列型号/API许可。
进程设置`BARNESS_AI_LIVE=1`、`BARNESS_AI_LIVE_COMBO=<路线>`、
`BARNESS_AI_LIVE_ACCOUNT_ALIAS=<账户和区域别名>`；唯一凭据变量为
`BARNESS_AI_LIVE_KEY_`加路线名大写并将`-`替为`_`，例如
`BARNESS_AI_LIVE_KEY_OPENAI_RESPONSES`。凭据值只通过获准的本机秘密来源注入。

六聊天路线各最多32 Calls、每Call最多4096输出token；公共Call timeout 180s，
每次场景执行context 240s。A类没有统一独立的业务进程收尾deadline，Host还需明确
整进程总时限和收尾安排，不能将通用Go测试超时当作完整终态保证。
1.13分类另限累计10题、每Call State≤128KiB，包含既有有界context错误探针。
两图片路线各最多请求4张、每Call 1张；OpenAI为1024×1024，Google为1K/1:1、store=false。
OpenAI先完成必需JSON编辑，再执行生成/可选mask；Google必需生成和单参考编辑。
所有Binding的SDK内部重试为0；A类最多1次的是harness场景重试。

B类使用`BARNESS_AI_LIVE_COMBO=openai-decisions`及
`BARNESS_AI_LIVE_KEY_OPENAI_DECISIONS`，独立入口`^TestFormalDecisions$`。
官方服务为`https://api.openai.com/v1/decisions`。默认还需：

- `BARNESS_AI_LIVE_ACCOUNT_ALIAS`、`BARNESS_AI_LIVE_REGION_ALIAS`；
- `BARNESS_AI_LIVE_PERMISSIONS_FILE`指向Host记录，字段为`accountAlias`、`regionAlias`、
  `dedicatedTestKey=true`、`model=gpt-6-luna`、`permissions=["POST /v1/decisions"]`、
  正数`spendLimitUSD`、非未来`verifiedAt`及`verificationReference`。

源码也允许显式`BARNESS_AI_LIVE_ALLOW_UNKNOWN_ACCOUNT=1`且不提供permissions文件，
将账户/区域/权限记为unknown；若采用此分支，执行授权需明确接受未核验状态。
Alias和Host记录都不等于厂商权限内省，`spendLimitUSD`字段也不是金额熔断器。
本路线限12 Calls/12 Attempts、累计24题、4MiB State、8图/4MiB解码图片；
Call≤30s、进程180s含10s收尾、零重试，九个必交场景不能UNSUPPORTED。

C类使用独立的多模态进程，不能用通用COMBO替代：
`BARNESS_MM_ROUTE=jev`、`BARNESS_MM_FORMAL=1`、`BARNESS_MM_LIVE=1`、
唯一`BARNESS_MM_KEY`和`BARNESS_MM_AUTHORIZATION`，入口`^TestMultimodalFormal$`。
授权JSON须恰有九字段：`route=jev`、`model=jev-1.14.0`、
`service=https://api.typesafe.ai/v1/systemone`、`accountAlias`、`regionAlias`、
`authorizationReference`、`dedicatedCredential=true`、`maxCalls=12`、`maxAttempts=12`。

C类协议累计≤32题、24 sources/32 occurrences、8MiB源图、16MiB展开、4MiB State，
每Call native body≤4MiB、Call≤30s、进程180s含10s收尾、零重试。
正式支持还需下面的Jev视觉效果包；封装为一个正式支持输入，效果调用不重复预算。

## 独立效果任务

| 任务 | 路线与服务 | 每轮Call / Attempt | 其他上限 | 凭据及授权 |
| --- | --- | --- | --- | --- |
| 中文Jev | jev-1.13.0，https://api.typesafe.ai/v1/systemone | 24 / 24 | 累计24题；零重试；串行；Call15s/工作600s | D |
| 中文Luna | gpt-6-luna，https://api.openai.com/v1/decisions | 24 / 24 | 同上，独立24样本 | D |
| 正式视觉Jev | jev-1.14.0，https://api.typesafe.ai/v1/systemone | 12 / 12 | 零重试；Call30s/工作225s/进程240s | E |
| 正式视觉Decisions | gpt-6-luna，https://api.openai.com/v1/decisions | 12 / 12 | 同上，独立12样本 | E |

中文固定任务为24条四类中文请求，不升级为1.14任务。视觉固定任务为12张合成图片，
公共可比子集仅前8张；Jev额外覆盖局部图片与Box，Decisions保持其context能力。
每视觉进程上限48题、24 occurrences、8MiB源图、16MiB展开，实际Jev28题/16次图片、
Decisions24题/12次图片。分类、存在性、误报和IoU独立报告；不设置准确率通过线，
不重跑样本挑高分。COMPLETE要求固定任务、实际终态、用量及完整性满足原规则，分数照实保留。

D类使用`BARNESS_AI_CHINESE_EVAL=1`和账户/区域组合别名
`BARNESS_AI_CHINESE_EVAL_ACCOUNT_ALIAS`。Jev只注入
`BARNESS_AI_CHINESE_EVAL_TYPESAFE_KEY`，Luna只注入
`BARNESS_AI_CHINESE_EVAL_OPENAI_KEY`；执行参数显式`-route jev|luna -live`。
它不读取B类permissions JSON，授权人仍需明确任务、路线及这一轮有限预算。

E类唯一凭据为`BARNESS_VISUAL_KEY`；分别执行`-formal -mode live -route jev|decisions`，
`-authorization`指向各自正式目录授权JSON。必须恰有十字段：`route`、`model`、完整`service`、
上面的`catalogHash`、`accountAlias`、`regionAlias`、`reference`、
`dedicatedCredential=true`、`maxCalls=12`、`maxAttempts=12`。
视觉alias/reference仅允许字母数字和`._-`，长度2–80，不接受`@`；不要原样复制中文任务alias。
真实视觉执行不使用环境代理，也没有与其他harness等价的显式proxy参数。

## 总额和仍需填写的授权项

完整一轮最大额度为：六聊天192 + Jev1.13协议6 + 两图片8 + Decisions协议12 +
Jev1.14协议12 = **230 Calls**；四效果任务48 + 24 = **72 Calls**；合计**302 Calls**。
按现有Binding零内部重试且每Call最多一个Attempt，HTTP Attempts也不高于302。
预算覆盖所有允许的故障重试；这是上限，正常执行量通常较低。Jev正式封装/验证、
支持矩阵合并、离线重放及最后完整门禁不额外调用真实服务。

按Provider汇总上限：OpenAI116、Anthropic32、Google36、DeepSeek64、TypeSafe54 Calls。
这是各路线额度相加；共享同一账户时还需明确账户总额度及各进程分配。

调用数、token、题数、图片和时间额度不等于美元硬上限。当前通用harness没有统一金额熔断；
B类的`spendLimitUSD`只验证Host记录为正数，不按累计费用自动停机。各账户需要提供可执行的
金额限额控制（例如服务侧可强制的项目限额）；当前无法据302 Calls推导精确账单。

每条路线或任务的批准记录应明确：

- 本清单中哪些具体行获准执行，包含其完整型号集合、API、固定任务和一轮次数；
- 使用哪个账户/区域、secret引用或本机变量映射，是否已核验权限及所允许的未知状态；
- Calls/Attempts及上述题数、token、图片上限，整进程时限和失败后停止规则；
- 每行及账户合计的金额上限、币种、约束机制，所有进程共享账户时如何分配额度；
- 是否允许既有显式loopback HTTP代理；A/B/C用`BARNESS_AI_LIVE_PROXY`，D用`-proxy`，E无此选项；
- 授权引用和有效执行窗口。新增一轮或失败后重新采集需仍在明确授权额度内，不自动循环补跑。

每进程只注入该路线的获准key。先无凭据构建，再以清理后的环境运行所选完整入口，
保留TLS校验与固定官方Endpoint；无环境代理、其他key或endpoint override。
不读取或借用其他任务的.env；本清单不复用历史授权来触发当前付费执行。

## 验证和发布顺序

1. 固定本轮源码/目录/配置和输出目录，确认上述授权、凭据来源及服务侧额度。
2. 分别执行11支持路线和4效果任务，保留所有失败、NOT_RUN、实际消耗及原包。
3. 用严格verifier验证效果包。将Jev1.14协议和同轮正式Jev视觉包封装并`-require-live`核验。
4. 用各自当前正式报告合并支持矩阵；Jev1.13、Jev1.14和Decisions身份独立，不改写旧包。
5. 将11个`-live`输入及4个效果输入交给新一轮完整releasegate，重跑normal/race、差分、压力和审计。
   新报告读取原#07已resolved的当前快照；旧报告不改写。满足全部合同后再更新#14及交付记录。

Jev封装及独立核验（均无真实服务调用）：

```sh
go run ./ai/live/cmd/jevformal -protocol NEW_JEV_PROTOCOL -effects NEW_JEV_VISUAL \
  -out NEW_JEV_SUPPORT -require-live
go run ./ai/live/cmd/jevformal -verify NEW_JEV_SUPPORT -require-live
go run ./ai/examples/visualeval -verify NEW_FORMAL_VISUAL -require-live
go run ./ai/examples/chineseeval/cmd/chineseeval -route jev -verify NEW_JEV_TEXT
go run ./ai/examples/chineseeval/cmd/chineseeval -route luna -verify NEW_LUNA_TEXT
```

全部路径使用新包，真实视觉两路线分别严格验证。完整发布命令及参数模板见
[发布指南](../../docs/barness-ai/current-catalog-release.md)。
此前零凭据收集脚本会清除key，只适合NOT_RUN/受控核验，不能直接用来执行真实路线。

## 依据

- [固定路线、型号与附加场景](../../ai/live/combos_test.go)、[进程与通用限额](../../ai/live/main_test.go)、[重试计账](../../ai/live/session_test.go)。
- [当前必交11路线](../../ai/release/traceability.json)、[当前矩阵](../../ai/live/support-matrix.json)。
- [Jev正式协议与封装](../../ai/live/testdata/jev-formal/README.md)、[协议授权实现](../../ai/live/multimodal_config_test.go)。
- [Decisions权限记录](../../ai/live/decisions_config_test.go)、[正式入口](../../ai/live/decisions_formal_test.go)。
- [中文效果](../../ai/examples/chineseeval/README.md)、[Jev配置](../../ai/examples/chineseeval/testdata/jev-route-v1.json)、[Luna配置](../../ai/examples/chineseeval/testdata/luna-route-v1.json)。
- [视觉效果](../../ai/examples/visualeval/README.md)、[授权实现](../../ai/examples/visualeval/authorization.go)、[固定预算](../../ai/examples/visualeval/testdata/experiment.json)。
