# 15 — 离线修复后的最小来源重验

Type: task
Status: resolved
Updated: 2026-10-19
Blocked by: 21

当前结论：本票已关闭；以下记录保留其形成时的状态，最终依据见文末[Answer](#answer)。

## Context

票13固定18个实际进程全部关闭，新增245calls/245Attempts；此前300/300保留，累计545/545。其余15项collection全部PASS，GLM Chat/Messages PASS、Responses19PASS/3FAIL；完整门禁报告false，后续离线缺口由票14修复。

既有基础spec要求SDK/型号/adapter改动重跑受影响组合；direct HTTP报告不会仅因无关Go文件变化自动过期。外层launcher完整来源冻结约束该次执行与取Key，不能推导为每一native报告的永久失效范围。票14若仅修改release工具、离线测试与夹具，且拟复用报告涉及的runtime、SDK、型号、目录、live生产者保持原实际来源，则原10项generic报告与原2份文本效果可按既有规则复用，原报告/commit/pins/timestamps全部不动。Protocol114及两份视觉报告有严格动态源码集合，仍须新实际证明；三GLM须新同源完整成功才可严格续期。

[独立Standards政策复核](../evidence/ticket-13/postgate-source-policy-review.json)与[独立Spec政策复核](../evidence/ticket-13/postgate-minimal-scope-spec-review.json)分别核对该边界。此前按外层1047文件集合推断全部18重跑的保守方案不作为额外收费依据；不新增来源豁免、不缩减旧map、不刷新旧Hash。

## Fixed scope — approved and closed (original execution contract)

只准备唯一`release-repair-validation-1`固定6进程：三路GLM各22场景，各最多24calls/32Attempts；protocol114最多12/12，视觉Jev与Decisions各最多12/12。计划102calls，硬上限新增108calls/132HTTPAttempts；在已知545/545之上合并最大653/677。所有GLM单次输出仍最多4096token，原inventory与默认零重试保持；GLM按[票16](16-glm-host-deadline-alignment.md)把专属Host响应头90秒对齐到原单调用120秒上限，旧90秒失败不改写。

中国区GLM直连；国际protocol114及视觉服务经`http://127.0.0.1:7890`，不继承代理或跨区回退。M3自身263文件来源若保持不重跑；M3.1没有M Plan资格仍NOT_RUN0/0。旧未用额度、旧审批、旧intent不能开启本合同，禁止自动新轮次。

该GLM专属装配调整在冻结前单独失败先行验证与双轴复审，明确属于新验三GLM的来源变化；不能把拟复用输入相关来源不变误写为全部GLM生产者不变。

本节保留执行前的固定审批条件：人类明确批准新增108/132，且精确source/binary/plan及六项无Key检查经两轴审查后，才允许读取本配置Key并执行。该审批与固定六项现已全部执行关闭；不会以未用额度开启新轮次，实际结果见文末。

## Acceptance

新合同独立绑定39份已关闭历史收据的545/545，冻结新来源和binary，另pin拟复用native输入及既有适用性证明。旧launcher在修复树上必须仍拒绝旧freeze，不能通过绕guard或重算旧pins来复用报告。新增唯一canonical prefix，缺新审批、错误预算、场景/来源/输入漂移、并发/跨目录重启、旧审批复用、失败预算重置和额外轮次均拒绝。固定全部前置检查通过后仍为PENDING，不读取Key或建立付费intent。

三GLM必须同轮自身完整PASS才发布严格raw/known-usage/support views；任何FAIL保留且不拼接旧Chat/Messages。随后冻结新protocol/visual阶段，派生新Jev支持包。矩阵只原子导入新GLM3+新Jev共4份报告，保留原13rows的全部时间和证明；不能重导旧报告绕过FinishedAt单调要求。最终公开门禁仍接收全部17自身输入与4效果，原未受影响输入按原字节保留。

最后在最终源码真正执行新双vet、完整ordinary/race、差分和压力，持久化四项数字退出码和报告；票13旧全量bundle不能授予修复后认证。新实际失败不自动开启下一轮。全部证据保留有界复核与独立两轴审查，完成后清理本轮自建工作树。

## Comments

2026-10-18：14、16的已审查源码及留存已在 `7231fdad` 合入并按各自离线范围关闭。本票6进程CLI/合同仍在失败先行实现，无新增付费审批。新来源、binary、plan及六项实际无Key检查须就绪并两轴复核，才提交一次具体108/132预算问题。

2026-10-18离线实现完成：源码09228（1b64后合入parent4d83）及留存90ec已合入集成分支，13个Python工具文件变化，Go/runtime仍等已审查4bf。71项公共工具测试、双vet、语法/行数/差异检查通过；源与留存各两轴0发现。[证据](../evidence/ticket-15/README.md)明确测量HEAD4bf+WIP与后提交的1056文件字节等同，隔离合成消费不作真实支持。实际ROOT source/binary/plan及六项无Key检查仍待完成，新108/132尚未请求或批准，本票不关闭。


### 2026-10-18 固定六项关闭后的最新交付检查点

2026-10-18 最新固定六项真实验收已按人类新增108calls/132HTTPAttempts批准完成：中国区GLM直连，国际protocol114及两项视觉经127.0.0.1:7890。Chat、Responses及三项效果collection通过，Messages21PASS/1FAIL，新增102/102，连原545/545累计647/647。该固定合同全部关闭，未用6calls/30Attempts不转移，无自动下一轮或新增付费授权。严格GLM续期和新四报告装配拒绝，原selector与17身份矩阵SHA保持；原矩阵16PASS/1FAIL为旧Responses记录，最新Messages原始失败未导入。

来源7931的首次完整无Key公共CLI及post-guards均已结束：八项输入校验0，双vet/ordinary/race数字退出0/0/1/1，CLI1、报告pass=false。普通/race各12451项（12445PASS/6FAIL/0NOT_RUN），审计0、868真实PIDIFF PASS；29个pressure命名用例通过，其中固定规则选择15项。六项失败为GLM三协议正式导入/已报告用量检查；未采用旧bundle或绕过缺失Messages报告。

最终双轴审查发现的Messages最终payload形状及文档问题已由票17离线补正，源码2306917、文档295728a均复审零发现。44公共用例为42非法输入零Provider/Attempt拒绝、两合法业务JSON原值发送；定向普通/race各1663PASS/8NOT_RUN，同8项mandatory显式开关另各8PASS，双vet0。实际测量为ca67+WIP，1027编译/JSON/mod/sum文件等于2306，不声称提交后重跑。修复后M3/GLM/文本/protocol/视觉全部11组严格来源当前相等性不成立；历史证据保持，generic10按既有受影响规则另行核对不变。旧门禁不认证当前源码。

本票固定执行与留存范围完成，但完整验收未满足，保持ready-for-human。不得将本轮108/132授权推给修复后的新来源或另开付费轮次。见[当前交付](../delivery-status.md)、[六项原件](../evidence/ticket-15/new-actual/glm/README.md)、[全量门禁](../evidence/ticket-15/new-actual/final-gate/README.md)及[最终审查](../evidence/ticket-09/final-closure/README.md)。

### 2026-10-19 单场景复测更新

用户单独要求重测Messages `glm-5.3/cancel`，票18独立1/1真实诊断PASS：来源6a24ba3，861ms收到非空ThinkingDelta后取消，canceled/stream与aborted，HTTP200。新增1/1、零重试，连647/647累计648/648；原固定六项与超时失败仍封存。没有拼接22场景、发布formal报告或修改四selector/矩阵；本票完整当前来源验收与发布门禁缺口仍在，状态不变。见[票18结果](../evidence/ticket-18/README.md)。

2026-10-19 当前补充结论：本票原有限轮次的失败、预算与旧拒绝保持历史。票19按新的直接批准执行十一组当前来源复验，GLM三路完整各22/22 PASS（Messages cancel1146ms），严格raw/known-usage/views发布及17份支持输入装配均exit0，默认矩阵17PASS/0FAIL。此前Messages/来源资格阻断已由新独立证据补足；完整发布门禁与最终审查另待完成。见[票19原件](../evidence/ticket-19/actual/remaining/README.md)。

2026-10-19 当前收尾检查点：票21新增36/36已闭合，三项各12/12 PASS；M3/GLM/文本八组完整原生来源集合不变，当前严格资格11/11。来源78843新完整门禁四命令0/0/0/0、CLI0、报告true；普通/race各12495PASS/audit0，最终源码及实际结果两轴均0发现。精确591323候选已实际exit0原子发布，仅Jev一行改变、其余16行对象与时间不变。原01341门禁FAIL及所有原有限授权轮次的失败、消费和未用额度保持；当前仅完整有界归档公开读回、最终文档审查及工单关闭收尾。见[当前交付](../delivery-status.md)及[当前实际结果](../evidence/ticket-21/actual/README.md)。


## Answer

原固定六项102/102、Messages120001ms失败及未用6/30保留，不转移旧授权。票18单场景861ms PASS及票19完整三路各22/22 PASS分别保留；票19十一组与票21三组独立新增批准完成当前来源补证。来源78843的新完整发布门禁四项实际命令0/0/0/0、CLI0、报告true；普通/race主包各12495PASS、0FAIL、0NOT_RUN、audit0，17份自身支持与4项业务效果PASS。严格当前来源11/11；五份归档、公开读回及固定证据两轴复核均通过，精确e2cf证据已合入8a1163b。默认591323候选实际发布，只更新Jev、保留其他16行对象和时间。自建工作树已移除、分支和中央原件保留，主仓三个安装入口均exit0，1078来源及七项原门禁hash保持。

续期累计873calls/870HTTPAttempts；原阶段、原失败、预算和未用额度不改写，所有付费进程关闭。中国区GLM直连，国际站经127.0.0.1:7890。见[最终闭合](../evidence/ticket-21/final-closure/README.md)及[当前交付](../delivery-status.md)。
