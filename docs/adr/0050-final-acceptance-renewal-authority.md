---
status: accepted
date: 2026-10-20
---

# 最终验收使用独立且绑定完整前序的续期授权

新增地区 Provider 后需要对最终源码重新取得国际 MiniMax 的三份自身正式证明；[ADR-0047](0047-current-source-formal-renewal.md) 的首轮续期已关闭。采用第二个实际用例 `provider-sites-final-acceptance-1`，只承载国际 `minimax`、`standard`、`MiniMax-M3` 的 Chat / Responses / Messages。旧 ID 永久拒绝活动执行，仅保留历史报告和来源的只读验证；不覆盖旧 grant、账本、claim 或 launcher，也不提供调用方自选的新轮次入口。

新 schema 2 grant 绑定原正式 39 Calls / 36 Attempts 的 SHA / 尾链，以及已关闭首轮续期的完整固定 canonical 文件集合：8 份 authority 文件、3 份 launcher；同时包含原正式账本和原 3 份 launcher，共 15 份事实。原生导出完整路径到 SHA 集合；来源审查和人类批准共同绑定其摘要及计划，不能用提交 bundle 自带的 manifest 选择历史。新轮次仍限制为总计 60 / 72、每身份 20 / 24、单次输出 4096、每身份累计请求输出 40000 与输入预留 120000。前序预算及失败不会转移或重置。

父 launcher 在读取 `.env` 前使用自己的唯一 intent 和只读 `minimaxverify -renewal-preflight=<route>` 校验；该校验不能领取子进程的 claim。子进程 Host 在读取所选 Key 前调用 `BeginRenewal`，以同一 Git common owner 的独占锁领取带进程随机 nonce 的持久 route claim。首笔预留只允许该进程消费一次 pending 状态；崩溃、换进程、复制相同 session、部分写入和删除账本均不会恢复可用性。claim / 账本和目录均同步到磁盘；不确定写入保留占用，代价是失败后必须检查原始事实并另行批准有限轮次。

活动执行必须匹配冻结 HEAD 和完整执行来源；历史消费仅允许未选中数据提交后的完整集合相等性。调用所依赖的固定美元 guard 及其区域候选价格目录均进入 MiniMax 执行来源集合。新授权机制和离线合成测试不授予真实请求、账户资格或付费预算。实际批准、证据与发布状态只记录在[规格当前状态](../../.scratch/barness-provider-sites/spec.md)；[工单 09](../../.scratch/barness-provider-sites/issues/09-final-round-minimax-authority.md)保留实现与验证指针。

直接替换旧常量或重封旧 grant 会复活已消费的权限；无限轮次注册表则会扩大当时只有两个真实用例的授权面，因此均不采用。

## 首个候选批次关闭后的固定续验 2

首个地区候选批次发生真实失败，`provider-sites-final-acceptance-1` 的原生 M3 grant/账本/nonce 从未发放。固定活动入口改为 `provider-sites-final-acceptance-2`，旧1显式选择和无 selector 的 canonical 账本路径均永久拒绝。旧1只允许已封存的 operator authority 和空锁；任何新 grant、nonce、账本或其他 intent 都令新原生预检失败。

schema3在上述15份原生历史之外，单独绑定 canonical `closed-cohort-1/seal.json`。grant、人类批准和独立审查各自 pin 该摘要，并保留旧批次实际40/40、已认领52/52及1.5USD整笔占用。新33进程仅722/758、28.4USD，跨轮774/810、全额29.9USD，不增加原30USD。原生 M3自身60/72和原15份历史39/36×2不变。固定路径从 common Git owner解析，合成独立owner可形成自身封存；不把真实本机SHA常量写进SDK。

封存中的旧二进制以稳定文件流式散列验证，限制4096文件、64MiB/文件、512MiB总量，源码归档遍历限制8192条目、深度32；禁止符号链接和未pin的归档文件。该具体历史校验与一般运行时授权分离，避免为一次续验引入通用轮次框架。

## 已消费 ID2 失败后的固定恢复 3

ID2 的国际 M3 Chat 已真实执行13 Calls／12 Attempts并语义失败，单次 nonce 和25行 canonical账本均已消费；它不符合既有瞬态或工具恢复条件，也不能在每身份20 calls中容纳另一次完整13调用。采用新的固定 `provider-sites-final-acceptance-3` 与 schema4，而不是覆盖旧grant、删除claim、复活ID2或建立任意轮次注册表。schema4在原15份39/36×2历史和cohort1之外，另行绑定 common-owner 的 `closed-cohort-2/seal.json`、234项停止原件、ID2原grant/human/review/chat nonce及账本SHA/尾链；失败13/12独立记录，不能冒充正式PASS。17份旧R2进程经234停止索引、其绑定的131项早期索引和三份真实GLM续行记录逐层保留，两份旧UNKNOWN及六份被新来源替代的formal不改写。

执行授权与资金政策分开：validationRound、nativeRenewalId、operator authority、人类回复和canonical新claim都是真实ID3；`spendingGuardPolicyId=provider-sites-final-acceptance-2`仅选用原US$30 campaign已经实现、字节不变的final-wire USD guard。新3输出中的独立USD文件及对应env如实使用旧政策的2格式，不提供ID2执行权限；plan/approval/claim/command显式记录政策源码Hash、原US$30批准、新3批准与新USD路径。保持guard源码不变可保留完整245项相等的三份b384 GLM原件；改guard会令三GLM资格过期，必须另增72/96的明确调用范围。MiniMax原生授权及CLI来源仍改变，因此六regional formal必须完整重采，不能刷新旧source pins。已发生的六candidate晋升作为历史保留，不能代替新formal资格。

新固定24进程为三M3、六formal、十generic、Jev协议及四效果，最多506／518；前序456／492领取和303／302实际不退款，外层962／1010。US$10.70整笔保持，未来US$19.20，合计US$29.90≤原30；Anthropic1.80是硬停止额度，Images保持3.50，其余原额度不变，不能保证长签名、历史和有界重试全部分支完成或成为厂商账单上界。原生M3自身60／72及每身份20／24、4096、40000／120000不变。先完整验证已知阻断的M3 Chat，再按固定顺序逐进程推进；每个前项须成功并完成原生消费；唯一延期例外为已成功的protocol114等待自身独立visual-jev，最终由jevformal配对验证通过才授予Jev支持资格，不能以pending状态发布。任何FAIL/UNKNOWN均停止，没有自动整进程重跑。最终33有效角色来自6历史candidate＋3保留GLM＋24新进程，而41份R2/R3实际记录及首cohort全部保留。

新schema4 human/review先以既有StrictJSON校验整个JSON envelope，再独立投影所需字段；允许描述性扩展字段，授权字段只认可规范大小写；任何层次的重复字段及大小写别名均不能靠Hash固定或last-value解析消除歧义。此实现和独立审查不授予新的付费范围：实际新回复仍由根任务在具体源码、预算、binary、计划及无Key检查可审阅后取得，PENDING模板不能生成有效grant或读取Key。原九服务standard资格及新READY后一次完整race的选择继续只引用既有明确批准；实际状态仍以[唯一规格](../../.scratch/barness-provider-sites/spec.md)及[固定恢复工单15](../../.scratch/barness-provider-sites/issues/15-fixed-recovery-authority.md)为准。

## 已消费 ID3 关闭后的固定恢复 4

ID3 三个 M3 进程39/36及地区 Chat18/18已完成，ZAI Responses18/18中的固定原生来源断言真实失败；父UNKNOWN与整笔额度保留。新的固定 `provider-sites-final-acceptance-4` 使用schema5，保留schema4的 `recovery` 闭合第二批事实，另加独立 `thirdCohort`21字段。它绑定第三批seal、1012项停止原件、原生九份authority（含三nonce和75行账本）、五份父claim及真实4PASS/1FAIL；旧ID3选中或隐式canonical活动入口永久拒绝，旧nonce/grant/账本不重写，不复活剩余额度。

历史消费与当前资格必须分开：第三批旧grant、原生报告、完整source-at-start manifest与原commit的Git archive相互核验，其来源已因验收源契约修复过期，不能与当前执行来源比较来决定能否封存。封存通过只证明旧事实未被改写，不恢复旧PASS的当前正式资格。旧schema1/3/4显式拒绝新增thirdCohort字段；新schema5依然要求完整当前执行来源和冻结HEAD。共同接缝仅为canonical grant加载、固定caps和稳定有界文件读取，固定历史语义分别验证，不提供任意ID、轮次或配额注册表。

新增24各一次，最多506/518，累计领取1070/1130；旧实际378/374、领取564/612、整笔12.60USD和三UNKNOWN原样保留。未来17.30USD加旧12.60为29.90≤30，仅Google interactions图片由4.50改为2.60，其余恢复3分配不动。正常固定两次发送的保守界2.592832，但第三次已发送尝试固定项至少3.052032；2.60是可能让有界重试分支提前停止的硬cap，不是完整分支或厂商账单保证。原生三M3仍60/72、每身份20/24、4096、40000/120000；USD-policy2的Go guard及GLM245不改变。

先执行预先固定的ZAI Responses完整18场景并完成原生消费，成功后才发放同一新批准覆盖的native4grant；其余23保留恢复3相对顺序。每项FAIL/UNKNOWN或额度耗尽即停，无自动完整进程重跑。新human/review额外绑定第三批seal及历史摘要，nativehuman还绑定canonical新实际回复；PENDING、重复或大小写别名均不能领取nonce或读Key。最终33角色只选六历史candidate、三保留GLM及新24，46份R2/R3/R4实际记录及更早关闭事实完整保留。实现、合成测试和审查均不授予新增付费权限；实际状态仍见[唯一规格](../../.scratch/barness-provider-sites/spec.md)及[工单17](../../.scratch/barness-provider-sites/issues/17-fixed-fourth-recovery.md)。


## 已消费 ID4 关闭后的固定恢复 5

ID4 已真实完成十八个进程：十七成功，第十八个 OpenAI Decisions 在底层 transport 前被美元 guard 拒绝。原生 recorder 的九 Calls/九 HTTP attempts 与美元 guard 零 reservation/零 transport reservation分开保存；父 UNKNOWN 不改成 PASS，也不能由 guard0/0推断厂商账单。停止索引2032项与第四批seal、原生九 authority 文件、三已用nonce、75行39/36账本和十八父claim均只读校验，实际591/584、领取980/1040、28.80USD整笔与四UNKNOWN不退款。

采用固定 `provider-sites-final-acceptance-5` / schema6，旧 `recovery` 与 `thirdCohort` 嵌套事实原样保留，另增独立 `fourthCohort`25字段。旧schema1/3/4/5拒绝该新增事实；旧4活动入口关闭，历史只读校验将原grant/report/source-at-start与原08提交的Git archive互证，不要求过期来源等于当前，也不恢复旧资格。共同的完整native report、consumer和归档读取接缝在两个真实关闭历史间共享；固定批次的语义仍独立，不能由调用方选择任意新轮次。

分类器预算修复会改变资金 guard 源码，因此新human/source-review/grant显式绑定当前guard.go与estimate.go两个SHA，不能把selector仍为2解释为字节未变。新5审批还需独立requestReference，不能把旧4回复改标签。拟定十九完整进程每项最多一次，378/414，累计1358/1454；旧28.80加新6.90合计35.70≤拟请求36USD。此处是离线设计，不构成任何新调用或预算批准。三M3自身仍60/72、每路线20/24、4096/40000/120000。

新原生grant的 `firstRouteProof` 精确绑定首ZAI Responses完整十八场景的canonical父claim、原生报告及消费命令三个path→SHA。原生消费者要求同当前源码及冻结binary的18/18、原生来源固定为两个型号各自full-stream，检查真实 `regionalverify -formal` 命令、数字exit0、无Key继承及审计；Python控制器仍须实际完成严格原生消费。该proof在实际首PASS后才能生成；只读 `-recovery5-source` 不依赖新human/grant/proof，可用于首claim前历史核验。三个M3 preflight只读，不领取子进程nonce；任何新阶段失败/UNKNOWN或额度耗尽即停。实际授权、资格和最终十项门禁状态仍仅见[规格](../../.scratch/barness-provider-sites/spec.md)与[工单19](../../.scratch/barness-provider-sites/issues/19-recovery5-native-authority.md)。
