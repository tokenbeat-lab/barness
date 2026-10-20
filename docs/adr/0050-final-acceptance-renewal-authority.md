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
