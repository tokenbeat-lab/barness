# MiniMax 交付状态

Delivery: PARTIAL
Updated: 2026-10-18
Integration branch: `codex/minimax-global`
Historical CN checkpoint: `codex/minimax-cn` / `4295ee8`

用户已明确将当前目标改为国际站。票11已完成Provider/端点/USD价格、证明身份与共享预算转换和国际站离线验证，10/09继续真实验收；原中国区离线通过和实际401保留为历史。**当前完整发布门禁未通过，spec、10和09保持未关闭。**

原CN检查点见[历史机器可读状态](evidence/ticket-09/partial-delivery-status.json)与[ticket10证据入口](evidence/ticket-10/README.md)。

| 范围 | 当前状态 | 证据与边界 |
| --- | --- | --- |
| 国际站三协议与六型号离线约束 | PASS，离线范围 | 8d56410完整MiniMax普通/race各1340/1340，定向普通/race各1372；审计0、双vet和双轴审查通过，见[收敛证据](evidence/ticket-10/international-convergence/README.md)。此前e9731a4全量无Key检查通过；不证明厂商接受修复后的请求。 |
| 原CN三协议与六型号离线约束 | PASS，历史离线范围 | 最新完整MiniMax普通/race均1307/1307；含冻结pi、压力、完整软件回滚，审计0发现。不能证明厂商实际接受请求。 |
| 原CN M3 Chat实际研究 | FAIL，历史范围 | 两次Logical Call、两次Attempt均HTTP401 `authorized_error`；没有实际型号或已知usage。依赖原生工具续接NOT_RUN。 |
| 国际站M3 Chat实际研究 | FAIL，原始研究范围 | HTTP200、实际M3；两次调用/两次Attempt在相等推理镜像处被本地误拒，usage未上报，依赖native NOT_RUN。 |
| 国际站M3 Responses实际研究 | FAIL，原始研究范围 | 文本/工具通过HTTP200；native续接本地编码失败，3calls/2attempts。 |
| 国际站M3 Messages实际研究 | PASS，仅研究范围 | 文本、思考工具及native续接3calls/3attempts通过；不能代替完整候选或正式支持。 |
| 国际站M3完整候选三身份 | 三项PASS，完整实际候选 | 来源7605332、各13calls/12attempts；三路推理/工具及持久化native续接通过，自己的实际origin与审计通过，见[精确汇总](evidence/ticket-10/international-candidates-7605332-summary.json)。 |
| 国际站M3正式三身份及目录激活 | 三项正式PASS；目录已激活 | 来源8900034，各13calls/12attempts，完整正式总39/36；目录2026-10-18.1-minimax/3718603f，原始报告/origin/源码/审计见[正式胶囊](evidence/ticket-10/international-formal-glm/README.md)。支持矩阵及完整门禁待完成。 |
| M3.1线上候选三身份 | 三项NOT_RUN | 未获M Plan用途/资格授权；零实际调用，不虚构公开价格，离线约束已完成。 |
| 原十四身份与四业务效果续期 | PARTIAL | 初轮GLM Chat/Messages各22/22 PASS，Responses22/22 FAIL、不完整且无seal，66/66消费全部保留。GLM捕获源码修复后，初轮报告不作当前支持；三路补正、其余身份与四效果仍待续期，额外预算未批准。 |
| 完整ordinary/race发布门禁 | NOT_RUN，未取得发布认证 | 已保留原ordinary失败/中断记录；46个可本地修复case的定向普通/race重验均通过，不能推出完整门禁PASS。 |

原两次实际调用后，另有一次仅本地语法诊断读取同一已选`MINIMAX_KEY`，网络请求为0，未输出Key值、长度或Hash；Bearer前缀、内部空白和控制字符标记均为false。此诊断不增加收费账本，实际消费仍为2次调用/2次Attempt；无Key验证进程各自保持未读凭据。

## 已保存的验证

国际站真实研究来源38b7098：[完整不可变捕获及来源索引](evidence/ticket-10/international-research-38b7098.json)保留148文件和454639字节归档，三个隔离启动器审计均通过。国际站认证有效；本轮缺口为协议兼容，不再等待账户修复。Chat未知用量不记为零账单；研究PASS也不授予正式支持。

当前国际站来源e9731a4：MiniMax普通/race各1314 PASS，live-tag race66 PASS；两种vet、精确来源prepare/check及无Keylive通过（328 PASS/62 NOT_RUN，另有辅助结果）。三份M3.1无资格NOT_RUN各为0calls/0attempts，未读Key或建立付费intent。双轴源码审查均0发现，完整捕获、逐文件Hash与重放入口见[ticket11](evidence/ticket-11/README.md)。这是国际站离线检查点，不替代自身actual/formal或完整发布认证。

完整MiniMax离线套件：[普通](evidence/ticket-10/minimax-ordinary-1205300.json)43.000秒、[race](evidence/ticket-10/minimax-race-1205300.json)152.736秒，均1307个case通过；冻结pi差分、压力与完整软件回滚均运行。46个修复后的定向旧case[普通](evidence/ticket-10/repaired-ordinary-1205300.json)/[race](evidence/ticket-10/repaired-race-1205300.json)分别46/46。[无Key live和两种vet](evidence/ticket-10/offline-checks-1205300.json)通过。来源1205300及其后矩阵数据修正的适用范围以10的原始manifest、审计、命令和归档Hash为准，不给历史捕获刷新来源。

原CN检查点无Key live的主manifest为323 PASS/62 NOT_RUN，另有两个辅助PASS；进程正常结束不意味着未执行的真实场景已通过。[历史CN M3.1三份NOT_RUN](evidence/ticket-10/m31-not-run-1205300-summary.json)及关键来源/子捕获保存在[有明确选择范围的无Key胶囊](evidence/ticket-10/postauth-keyless-1205300.tar.gz)。

旧9481a6a完整ordinary的测试主体已结束，[原manifest](evidence/ticket-10/whole-ordinary-failed-9481a6a.json)为12304 PASS/52 FAIL；进程在后续阶段因已知harness修复被停止，exit=-15，不能声称完整普通套件通过。独立redaction audit随后完成，0发现；这是独立派生审计，不修改原退出状态。[计时调查](evidence/ticket-09/whole-ordinary-timing-investigation.md)没有支持E03或精确小数解析死循环的证据。52个失败中的46个本地可修复case已定向重验，其余六个GLM来源校验需真实续期，不能放宽适用性规则。

## 历史证据完整性

原完整失败bundle已在主工作区本地保留，110MB归档不纳入Git；ticket10的[有限索引](evidence/ticket-10/whole-ordinary-failed-9481a6a.json)与[完整文件Hash清单](evidence/ticket-10/whole-ordinary-failed-9481a6a-files.json.gz)记录其位置和SHA。原131505个文件未改写，归档额外保留一个21字节的派生审计文件。

98236个声明的artifact引用全部匹配，零缺失、零artifact不符。另349项fixture来源归属Hash与脱敏后的artifact文件Hash不同，二者按既有语义分别记录；不能合并两种Hash后推断捕获被篡改。[独立只读复核](evidence/ticket-09/whole-ordinary-artifact-integrity.md)通过单case来源和登记替换的逐字节复现解释了该区别，见[版本与来源](evidence/ticket-09/whole-ordinary-artifact-integrity-provenance.json)。该完整性结论不改变原52个FAIL或中断退出。

复核脚本记录的是调查时路径。可将保留归档解压到临时目录，并把脚本的`bundle`指向该目录后只读复核；不需要Key、厂商调用或重跑测试。

## 后续冻结前检查点

国际站实际wire收敛已合并至366fb7f：完整普通/race各1340 PASS，定向普通/race各1372 PASS、双vet与双轴审查0发现，1024个生成来源文件与8d56410字节相同。[原始收敛归档](evidence/ticket-10/international-convergence/README.md)保留red/green与原始失败，不将修复视作服务接受。

独立后续顺序审计发现：激活M3使当前目录Hash改变，旧GLM正式回放却装配当前目录，导致续期前全量无Key检查自阻塞。[顺序与具体边界](evidence/ticket-10/international-postcandidate-sequence.md)交同一实现者先作失败先行修复：历史回放只能使用经验证原始快照并保持replay-only，当前实际正式/导入/门禁继续严格校验当前目录。该修复完成前不启动完整候选，避免已消费的一次性证据因源码改变而过期。

后续源码de05193已修复上述循环及完整激活时发现的当前目录版本、Host夹具和MiniMax双目录选择问题；续期工具将只更新四份当前目录归属文档与生成快照。定向live67 PASS、续期CLI11/11 PASS；独立完整激活克隆的全量无Keylive为331 PASS/62未选厂商NOT_RUN/0 FAIL。该来源的MiniMax完整普通/race各1340 PASS，双轴审查0发现，见[审查](evidence/ticket-10/international-preflight-review-de05193.json)；激活后的公共E2E全集已结束：12378 PASS/11 FAIL，exit1、审计0；六项为待真实续期的GLM来源校验，另外三项发布叶子场景及聚合失败来自完整配置Hash耦合目录归属，GLM回滚另为私有克隆软链接设置问题。原始失败不改写，归档待整理。这些均不构成真实候选或正式支持。

最新df342b9保留固定中文任务的独立业务Hash，只排除已由Plan.Load严格校验的两项目录字段；24个定向边界case通过，包含指令/超时协同篡改拒绝、原发布正例和回滚设置复验。[双轴源码审查](evidence/ticket-10/international-business-boundary-review-df342b9.json)均0发现。该结果只解决已复现的额外失败，不能将原11 FAIL改标为完整ordinary PASS；六项GLM来源时效仍须真实续期，后续最终发布门禁仍必交。

## 待继续工作

1. [票11](issues/11-international-target.md)已完成国际站身份、三端点、USD价格、frozen pi与离线复验；沿用户指示使用单一MINIMAX_KEY做国际站标准M3三协议验收，不再等待中国区修复，不推断M3.1资格。
2. 票10先以已保存实际wire失败E2E修复Chat镜像推理及Responses本地native编码。随后对新的已提交源码重新prepare、无Key检查和精确来源审查；旧9481a6a批准不适用于新HEAD。随后使用独立国际站研究intent与自身前驱继续三候选真实验收；保留原CN401及所有失败，不将其作为跨区native前驱，不覆盖既有intent。完整候选三路已通过；候选阶段累计49/120调用、45/144 Attempts（49reserved），包含历史消费；Chat槽17calls/16attempts、Responses16/14、Messages16/15。正式阶段已39/60调用、36/72实际Attempts（39reserved）。不得重置账本或降低验收。
3. 三候选各自通过后才作数据目录激活，获取三份自身正式报告；再续期原十四身份和四业务效果，运行完整ordinary/race/压力/pi、两种vet、逐包脱敏和回滚门禁。
4. 对固定基点`ecfd4b4`完成最终独立Standards/Spec审查及必要修复重验，全部必交通过后才关闭spec和10/09。

历史、受控或回放结果均不转成actual/formal PASS。历史追溯表继续保留其原来源和状态；当前完整交付判断只依据实际适用证据。

2026-10-18当前实际收集已封存：[完整候选](evidence/ticket-10/international-complete-candidates/README.md)和[正式/GLM原始轮次](evidence/ticket-10/international-formal-glm/README.md)。两份便携校验器均通过，全量原始失败保留；六份M3实际origin已由公共验证器验证并登记。GLM Responses失败同时包含4096预算用尽、无正文、512KiB捕获截断和17.7MB事件文件；原22/22计入消费。拟离线准备一次GLM专属补正，不修改M3执行来源或正式断言，不自动授予新预算。

[票12离线检查点](evidence/ticket-12/README.md)已完成GLM专属捕获/重放容量及用量结算修复：最终普通/race各4 PASS/audit0，辅助CLI35 PASS，两种vet通过；两轴最终0 open。三份现有M3正式公共验证均PASS，各263执行文件不变；集成1,033生成文件与已审查6ff字节相同。拟一次GLM corrective-1，固定66场景、额外最多72calls/96Attempts，每次4096输出；本次续期累计GLM138/162、其他旧能力302/302不变、旧能力合计440/464。此前任务历史另行保留。该方案等待明确人类批准，尚无新真实凭据读取、厂商调用或canonical corrective intent。
