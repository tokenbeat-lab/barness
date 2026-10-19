# MiniMax 交付状态

Delivery: PARTIAL
Updated: 2026-10-19
Integration branch: `codex/minimax-global`

2026-10-19 **原定十一组复验全部通过（11/11，测量来源01341及其相同选定源码）**：M3×3、GLM×3、文本×2、protocol114×1、视觉×2。GLM三路来源d235，各22/22；严格raw/known-usage/views发布全部exit0。其余八项来源b9c3c7c，每项自身真实报告及当前公开消费者全部通过。Messages `glm-5.3/cancel` 在新完整22场景报告中1146ms PASS，返回canceled/stream、aborted；partial usage不记免费。

本轮付费已关闭：新增189calls/186HTTPAttempts，上限216/252；未用27/66不转移。中国GLM直连，国际服务经127.0.0.1:7890。连原续期648/648为837/834；原M3候选49/45、正式39/36及更早GLM任务另列。新M3自身正式39/36保留独立账本；M3.1三路新无Key证明NOT_RUN0/0，未获MPlan资格。见[GLM原件](evidence/ticket-19/actual/glm/README.md)、[严格发布](evidence/ticket-19/actual/glm-publication/README.md)、[剩余八项与统一装配](evidence/ticket-19/actual/remaining/README.md)。

统一装配已审计17份自身支持输入，仅更新新7行，旧generic10行对象与时间不变；默认矩阵17PASS/0FAIL。两文本另经公开`release.ChineseEffectsEvidence`当前来源校验exit0。此次关闭严格来源证据缺口，证据/文档两轴审查已各0发现。**完整发布门禁已实际关闭但未通过，交付保持PARTIAL**：来源01341f8，双vet/普通/race退出0/0/0/1，CLI1、pass=false；普通12495PASS/audit0，race12494PASS/1FAIL/0NOT_RUN/audit1。唯一场景VQ12-failure-decisions-plain-deadline的失败包重放被拒；唯一审计发现是该包samples/v08/request.json零字节、unparseable。17自身支持输入与四项业务效果均PASS，前后1074来源及21输入树相等，0Key/0新增付费。原件位于`.evidence/minimax-current-eleven-gate-2`；[完整留存入口](evidence/ticket-19/final-gate/README.md)已独立有界读回、零留存发现。票19按复验及真实结果留存范围关闭，完整spec继续开放。

**修复合入后的当前来源资格为8/11**。视觉示例修复已由两轴复审零发现并精确合入`d02b5ad`：真实调用在解析归属前或读取凭据后取消时保留实测零Attempt证据；不再生成空请求JSON；已发送取消包的null/unset型号判断亦已修复。23个确定性回归的普通/race均23PASS、audit0，双vet0；实际测量a203+WIP的1040份编译/JSON/modsum字节与已提交d02相等，不冒充提交后重跑。原a203的null访问计数发现已失败先行补正，原失败及审查保留。

本次只改视觉示例，M3/GLM/文本八组的完整原生选定源码集合保持；protocol114与视觉两组需新来源证明，派生Jev支持同时待新证明。原矩阵17PASS仍为已测量来源的保存数据，不能替代这三组当前资格。[票21](issues/21-visual-source-renewal.md)正在离线准备固定三组补验：各最多12calls/12Attempts、共36/36、零重试，经127.0.0.1:7890。原216/252轮已关闭、未用27/66不转移；**尚无新增36/36付费授权**。最新完整门禁仍是01341的历史FAIL，尚未为d02执行新的完整门禁。票20的原始回归留存正在封存。

以下为历史检查点；原失败、旧输入0/11诊断、原hash/map与消费均保持，不用本轮结果改写历史。

2026-10-19 最新单场景复测：**Messages `glm-5.3/cancel` PASS**，来源 `6a24ba3`。中国区直连，861ms 收到非空 ThinkingDelta 后主动取消，返回 `canceled/stream`、HTTP200、`aborted`；1call/1HTTPAttempt、零重试，usage为partial、不记免费。当前用户指令单独授权的1/1已执行关闭，旧续期连历史647/647累计648/648。已立即更新[机器状态](evidence/ticket-09/international-delivery-status.json)与[本次证据](evidence/ticket-18/README.md)。原120001ms失败保留，原始wire request相同，四selector与默认矩阵未变。此PASS只关闭该场景的最新复测观察；全22场景正式Messages报告、严格来源资格、GLM续期/矩阵装配与完整发布门禁仍未通过。以下固定六项是原历史检查点。

已按指定路由完成获准的六项补充真实验收：中国区GLM直连，国际protocol114及两项视觉服务通过`http://127.0.0.1:7890`。五项collection通过，GLM Messages失败；新增102calls/102HTTPAttempts，本次旧能力续期连原545/545累计647/647。该固定合同全部关闭，未用额度不转移，没有后续付费轮次授权。

| 范围 | 已关闭的实际结果 | 依据 |
| --- | --- | --- |
| 最新GLM三协议，来源7931 | Chat、Responses各22/22 PASS；Messages21PASS/1FAIL；66calls/66Attempts | [新GLM原始证据](evidence/ticket-15/new-actual/glm/README.md) |
| 新protocol114、visual Jev、visual Decisions，来源7931 | 三项各12/12 PASS，共36/36 | [新效果原始证据](evidence/ticket-15/new-actual/effects/README.md) |
| GLM严格raw/views续期 | 实际exit1，四个原selector字节不变 | [新GLM原始证据](evidence/ticket-15/new-actual/glm/README.md) |
| 新四报告装配 | Jev派生0、校验0，候选导入1；无默认矩阵导入 | [装配失败原始证据](evidence/ticket-15/new-actual/assembly/README.md) |
| 默认17身份矩阵 | 仍是上一轮16PASS/1FAIL（GLM Responses）；最新Messages失败尚未导入 | [原矩阵](../../ai/live/support-matrix.json) |
| 原国际M3三协议自身formal，来源8900034 | 历史三项PASS，39calls/36Attempts，本轮未重跑 | [原正式证据](evidence/ticket-10/international-formal-glm/README.md) |
| M3.1三协议线上候选 | 三项NOT_RUN0/0；没有获准MPlan资格 | [原候选状态](evidence/ticket-13/actual-legacy/README.md) |

GLM Messages的`glm-5.3/cancel`在收到可观察内容之前到达120秒调用期限，实际duration120001ms；这不能证明成功执行了取消，也不能定位厂商或内部网络原因。其raw formal报告与完整性原件保留，`live-report.json`未生成。严格续期与装配拒绝均保留实际退出和错误，不舍入时间、不合成支持报告、不替换旧Messages或Jev报告。新Responses通过不改写上一轮90秒响应头超时或R1缺正文的历史。

来源7931的完整公开发布CLI与最终输入post-guards均已完成：八项自身输入校验数字0；vet/live-vet/ordinary/race分别0/0/1/1，CLI exit1，报告pass=false且与实际命令一致。普通与race主E2E均为12451项（12445PASS、6FAIL、0NOT_RUN），审计各0；868个真实PIDIFF、29个名称含pressure的用例均PASS，其中固定PRESSURE规则选择15项。六个失败均为GLM三协议正式导入或已报告用量检查。完整stdout另外确认失败包仅ai/e2e；live/trace等最终门禁也保留导入拒绝后的真实失败，不冒充通过。未采用旧bundle，门禁不读Key/不新增付费。见[完整门禁原件](evidence/ticket-15/new-actual/final-gate/README.md)。

[票17](issues/17-final-payload-review-gaps.md)离线修复最终Messages的消息容器、角色、内容块及函数工具必需字段校验，保留业务JSON/schema数据。源码2306917、文档295728a均已独立双轴0发现，原最终审查的Standards1项/Spec2项源码与文档问题闭合。44公共用例中42非法输入零Provider/零Attempt拒绝，两合法业务输入原值发送；定向普通/race各1663PASS、8NOT_RUN，另以显式压力/差分开关对同8个mandatory用例各补普通/race8PASS，审计均0、双vet0。原NOT_RUN与命令窗口重叠不改写。实际测量为ca67+WIP，1027个编译/JSON/mod/sum文件等于已提交2306；没有声称提交后运行或完整1056来源冻结。票17原始留存及新负向夹具构造器的有限读取/解压/哈希问题也已失败先行补正，最终双轴复审零发现；原夹具错误/失败GREEN/依赖缺失等分支见[完整离线历史](evidence/ticket-17/README.md)。

原M3、GLM、文本、protocol及视觉的严格来源集合均包含本次最终payload修复涉及的runtime文件。修改后，原报告仍可在其测量来源复核历史结果，但不能认证修复后的当前源码。全部原hash、map、selector、catalog、矩阵时间和授权保持；source7931全量门禁也不授予后续17源码认证。generic10已按原受影响组合规则核对SDK/目录/型号/live producer及非MiniMax路径不变；这不依赖缺map豁免。最终合入来源共1117个文件（ai源码/数据及go.mod/go.sum）与2306字节相等，完整新增/删除集合也已核对，原严格map仍不具当前资格。

历史检查点当时spec、10、09及真实GLM验收未关闭。当前真实GLM与十一组来源资格已通过，spec、10、09仍待完整发布认证和最终审查。此前300/300与区域路由轮245/245、旧全量失败及中国区401均按各自原轮保留；M3候选49/45、正式39/36与更早独立GLM任务另列。未知或部分usage不记免费。

[两轴审查及留存入口](evidence/ticket-09/final-closure/README.md)提供原始命令、全文件hash、有界无Key复核及失败分支；大型归档保留在主仓库`.evidence/minimax-retained/`。自建worktree已实际移除，分支保留以便恢复；见[清理记录](evidence/ticket-09/final-closure/owned-worktree-cleanup.json)。用户原有session-analysis与其他功能human-review目录保持。
