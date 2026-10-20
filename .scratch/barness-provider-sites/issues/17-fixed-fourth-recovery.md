# 17 固定恢复4的完整关闭历史与有限续验

Status: claimed
Type: task

## Context

[16来源修复](16-zai-responses-native-source.md)、[07最终验收](07-final-acceptance.md)、[规格R18/R20](../spec.md)。本票仅准备具体可审阅的下一固定授权/采集机制；无新的付费批准，不读Key、不请求厂商或执行race。

## Fixed history

恢复3已按约停止，不能重启原grant/nonce、删claim或退款。[停止索引](../../../.evidence/provider-sites/final-acceptance-20261020/recovery3-stopped/index.json)SHA `b3fcbe0d37dfd03a3892a3be9025bdc2ffe8cb1a20e43e971115cf6f909e8320`固定1012原件；[关闭seal](../../../.evidence/provider-sites/final-acceptance-20261020/closed-cohort-3/seal.json)SHA `1579ac7775a87c3c027381f67f68af77fa6403fef576de4e5a9130fba5f05c24`。实际378/374、claimed564/612、整笔12.60USD、UNKNOWN3永久保留。旧15原生事实、closed1、closed2、ID3九个原件（含三nonce、75条ledger，39/36）、五canonical3进程的4PASS/1FAIL全部只读绑定。

## Required design

固定ID4/schema5，旧ID3永久拒绝活动使用；新实际human、完整来源/工具/二进制、source-review与新24明确绑定。历史闭环校验原grant/report/原索引与原commit，不要求已经过期的旧source等于当前，也不把旧PASS刷新成当前资格。只提取已有真实用例的共同校验接缝，不引入无限轮次注册表或调用方自选任意身份。

16同步shared消费者后，三M3及六formal必须重采，GLM245逐项相等才保留。固定新24为M3三、地区formal六、generic十、Jev协议一及四效果，最大506/518；跨轮领取上限1070/1130。先ZAI Responses验证当前阻断，其余23保持恢复3相对顺序；其PASS后才颁发本轮新native4 grant也可，但所有范围必须预先由同一实际批准授权。任一进程FAIL/UNKNOWN/额度耗尽即停，无新研究或自动完整进程重跑。

最终33逻辑角色仅选六历史candidate＋三保留GLM＋新24；R2/R3/R4实际记录共46（17＋5＋24），三个旧失败/UNKNOWN不得替代或移除。首轮已闭事实另行保留，不混入该46口径。

原final-wire资金policy2/Go guard与GLM245不改；validationRound/nativeRenewal/operator/human/claims均真实4。金额按[预算研究](../../../.evidence/provider-sites/final-acceptance-20261020/recovery3-zai-responses-diagnosis/budget-feasibility/budget-feasibility.md)：仅Google interactions图片从4.50降为2.60，其余恢复3分配不动；新增17.30＋已占12.60＝29.90≤30。正常两固定请求保守预留上界2.592832；任意第三次已发送尝试的固定项就至少3.052032，因此必须明确这是operational硬停止额度，不保证已发送重试分支完成，不能删既有场景/重试/模型/wire限制或当作账单保证。

所有国际站继续显式走127.0.0.1:7890，中国站和DeepSeek直连；九服务资格及新READY后一次完整race沿用已有批准，不重复询问。

## Acceptance

先公共CLI/进程RED，覆盖伪旧PASS/FAIL/计数/父UNKNOWN、关闭索引与九native原件变动、旧3活动入口、漏人类新批准、错资金policy/预算、重复claim/不同out/并发/nonce/FIFO/路径/源码漂移拒绝；保留合法旧历史消费和当前实际资格的差别。新public流程完整24→33/46→23+4输入，Jev protocol唯一配对延期及最终强制verify不弱化。独立双轴复审清零后冻结实际源/新二进制/工具及全部输入，完成无Key检查，模板保持PENDING；由根对具体计划申请一次新增paid范围。取得新真实READY后才执行已有批准的唯一完整正式race。

## Offline native implementation progress

固定原生ID4/schema5已在07807a1交付；公开facts精确导出第三批21字段摘要1f346af96e867d5f5fb60621eac454cedd71c4ce415a900ae58e825742c50f21，旧recovery字段、15/closed1/closed2及真实closed3全部只读闭合。旧3活动入口永久拒绝；旧source按原grant/report/commit/archive互证，不能要求其与当前相等或复活旧资格。实际新4回复还须使用不同于旧3的requestReference。

[原生永久索引](../../../.evidence/provider-sites/final-acceptance-20261020/recovery4-native-dev/index.json)SHA9ce2967938173eea9d845caf1e8409b426f41a0036fd4abafd19d4dbf78263f7固定55公共进程场景PASS（实际墙钟121.614878s）、持久binary/argv/UTC/数字退出码、目标vet、旧回归和全部中间失败。覆盖三路线新39/36、篡改/旧活动/UNKNOWN/nonce/并发/跨worktree/旧回复复用等拒绝；在该原生离线验证阶段，真实4人类记录和grant均不存在，无真实Key/调用/账本写入或race。Status继续claimed；离线实现与复审的后续固定事实见下节，唯一当前状态见[规格](../spec.md)。

## Offline controller and final review progress

[控制器259工件](../../../.evidence/provider-sites/final-acceptance-20261020/recovery4-controller-dev/index.json)与[集中修复138工件](../../../.evidence/provider-sites/final-acceptance-20261020/recovery4-review-fix-dev/index.json)分别保留原RED/GREEN、实际argv/UTC/数字退出和源副本；原259指纹逐项未变。最终受影响公共组合50/50 PASS，204.167s、exit0，覆盖旧R2/R3及新R4的完整24→33/46→27inputs→23行发布流程。

Standards两项（policy/原批准指纹非阻塞有界读取、删除内部纯转发facade）与Spec一项（首claim/Key前完整原生历史校验及proof绑定）均由单一implementer在81e79b3集中修复。首条只读原生facts与后续grant发行采用不同独占回执；grant仍须首18/18及当前consumer PASS。最终[Standards0/Spec0双轴索引](../../../.evidence/provider-sites/final-acceptance-20261020/recovery4-preparation/review/index.json)SHA `8ca4aab7dcc940c481dc2fc17a27fb146029f6c8a6527ebad194ce794b9295dc`。此次修复未改Go/资金guard/GLM，执行集合GLM245/MiniMax320/Regional414保持。

后续授权/运行、实际冻结及根检查状态见[准备当前状态（CURRENT_STATE）](../../../.evidence/provider-sites/final-acceptance-20261020/recovery4-preparation/index.json)及[具体计划](../../../.evidence/provider-sites/final-acceptance-20261020/recovery4-preparation/plan.md)，不以本段离线结果代替实际来源/十binary/工具/检查回执。截至81e79b3离线复审结束，实际新付费授权4/grant4均无，无真实Key/厂商调用或新race；该阶段READY/race NOT_RUN、release.pass=false，原历史数字/原件/预算/资格边界保持，Status继续claimed。

## Recovery4 actual stop and budget correction

恢复4实际08a3d10完成六地区formal及三M3自身完整PASS，前17进程成功；第18项OpenAI Decisions在finalbudget前发送估算失败，后六项按约未启动。原报告9Calls/9 recorderHTTP与USD guard0/0并存，不改写计数或退款。停止2032原件和closed4、跨轮native591/584、claimed980/1040、整笔28.80USD与UNKNOWN4全部保留。

[18分类预算隔离](18-classifier-budget-operation-isolation.md)离线最小修复进入三类严格来源，既有GLM245/MiniMax320/Regional414成功不能作为修后当前资格。新预算E2E加入后Regional416；没有改变任何来源选择器。八generic的保留须额外绑定原件与逐份final-wire/ledger差异证明。上述只记录本票历史和依赖，唯一当前状态见[规格](../spec.md#当前状态)；本票保持claimed，新付费机制/授权与最终门禁另行完成，READY/race未运行、release.pass=false。
