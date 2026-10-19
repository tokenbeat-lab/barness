# MiniMax 交付状态

Delivery: PARTIAL
Updated: 2026-10-19
Integration branch: `codex/minimax-global`

2026-10-19 十一组当前来源复核已实际执行：旧输入当前正式资格0/11；逐组来源拒绝及缺失报告见[无Key复核](evidence/ticket-19/precheck/README.md)。原场景历史结果及票18单场景PASS不改写。独立、有界的续期工具与M3持久授权已集成，源码fbbdea4经Standards/Spec独立复审各0发现；矩阵历史报告重导缺陷已失败先行修复。来源0bb的六项无Key前置检查已关闭：5通过、live测试因新受控场景的子进程输出含本机路径触发脱敏审计而失败；[原失败](evidence/ticket-19/preflight-first/README.md)保留。HOME诊断修复a257已合入且双轴复审各0发现，80项Python工具回归PASS，HOME及外层原件有限读回均0；最终来源d235的[新冻结六项无Key检查](evidence/ticket-19/preflight-final/README.md)全部通过，新216/252已由直接回复批准，GLM三路各22/22正式PASS，共66/66；Messages取消场景1146ms通过，新增原始证据见[本轮GLM](evidence/ticket-19/actual/glm/README.md)，[严格raw/known-usage/views发布](evidence/ticket-19/actual/glm-publication/README.md)全部公开验证通过；其余八项仍待执行。本阶段0Key读取、0收费请求，新预算216/252已批准且开始执行；[源码审查](evidence/ticket-19/reviews/README.md)不授予正式资格。

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

spec、10、09及真实GLM验收仍未关闭。当前没有完整发布认证。此前300/300与区域路由轮245/245、旧全量失败及中国区401均按各自原轮保留；M3候选49/45、正式39/36与更早独立GLM任务另列。未知或部分usage不记免费。

[两轴审查及留存入口](evidence/ticket-09/final-closure/README.md)提供原始命令、全文件hash、有界无Key复核及失败分支；大型归档保留在主仓库`.evidence/minimax-retained/`。自建worktree已实际移除，分支保留以便恢复；见[清理记录](evidence/ticket-09/final-closure/owned-worktree-cleanup.json)。用户原有session-analysis与其他功能human-review目录保持。
