# Provider 交付准备当前状态

2026-10-27：本次授权范围完成。MiniMax国际Messages其余13个成功场景已取得当前逐项适用性，thinking-on断流已受控复现；六份Kimi待批准范围v2与29条路径矩阵已交付。原真实失败、授权、账本和暂停状态保持。没有真实厂商调用、付费授权签发、真实claim、race/压力、目录激活或暂停恢复，release.pass=false。

当前[交付矩阵](delivery-matrix-v2.md)、[最小补采清单](minimal-capture.md)、[六份Kimi范围v2](kimi-scopes-v2/README.md)。机器原件：[decisions-v1](../../.evidence/provider-final-readiness/decisions-v1.json)、[矩阵v2](../../.evidence/provider-final-readiness/delivery-matrix-v2.json)、[资格索引v5](../../.evidence/provider-final-readiness/current-qualification-index-v5.json)、[完整验证与保全](../../.evidence/provider-final-readiness/verification-and-retention-v1.json)。旧索引/失败和v1草稿均保留，以本状态指向的v2范围为当前准备材料。

## MiniMax 国际 Messages

| 事实 | 当前结论 |
| --- | --- |
| 原测量 | formal/live；15条Call、14个唯一场景；原producer exit1、报告FAIL/complete=false |
| 当前可复用 | 13个成功场景全部通过当前既有原生逐项检查；源/SDK/型号目录投影、封存/审计、原授权与权威ledger均PASS |
| control初次失败 | 同进程合法rerun已成功；原首次失败保留，不列第二个缺项 |
| 唯一新增事实缺口 | thinking-on的成功真实测量，保持原CRT任务、adaptive thinking、2048输出及成功文本/终态门槛 |
| 当前完整资格 | 仍拒绝“MiniMax report is incomplete”，没有从13项分析产生完整PASS |
| 执行准备 | 现有完整入口固定14项；尚无MiniMax单项采集/13+1组合合同。本轮仅列缺口，付费调用上限与USD未核定，不默认重采13项或整组14项 |

原件可定位审计共2227项：2055项为留存相同字节、172项从原507601386083f1d21d83875a8ce46c892117ed54 Git blob匹配，缺原件0；原producer数字退出、捕获日志、manifest/audit、实际env -C启动的二进制及其grant哈希均核对。[原件审计v2](../../.evidence/provider-final-readiness/originals-audit-v2.json)。

[当前原生逐项审计](../../.evidence/provider-final-readiness/native-analysis-v3/receipt.json)使用明确标记的分析overlay，只暴露现有校验函数，不安装生产partial消费者，不改写原包。tool-result-native仍绑定该包实际tool-call前序；cancel/preflight各按自身原合同检查。所选13项和恢复前初次control错误的既有边界均有效，原budget/15条调用记录不被改成一次新的13调用测量。

四个公共Client受控场景分别证明：原793字节合法推理前缀后非EOF读错→transport/stream；相同前缀普通EOF缺终态→protocol/stream；补齐合法结束→成功；初始request EOF→transport/request。当前编码wire与原请求相同，部分Thinking保留、无虚构正文/成功、一次Attempt无流内重放、唯一终态、Observer/Result一致、body与资源释放。原read error具体字符串未留存；合成注入不证明真实中断来自厂商/代理/SDK/网络某一方。产品/SDK未改，不增重试。

## Kimi 六份待批准范围

各路线仅补kimi-k3/xai-origin-history，1调用/1HTTP/零SDK与环境重试，输入预留4096/输出4096，USD0.10，六路总USD0.60；调用120s、进程240s、首发等待60s、串行原生时间上限24分钟。仅为冻结价格快照下的提案，不是实际账单或执行批准。

每份v2提案绑定自己的原九项manifest/report/seal、当前native source/SDK/候选目录、正常live-tag二进制SHA、原账户/地区/tenant/account scope的待确认范围。未签发ID/有效期，不复用旧余额。[提案索引](../../.evidence/provider-final-readiness/kimi-pending-scopes-v2.json)及[只读授权检查](../../.evidence/provider-final-readiness/proposal-validation-v1/receipt.json)确认六份均不能作为可执行grant。

准备时修复一处真实消费者遗漏：旧run.py把两份README记在selectedInputs，原生来源含二者；此前消费者只查source.files，受控夹具又未覆盖分栏。现已精确核对两个域，缺失、改动或冲突仍拒绝；公共反例先RED后GREEN，[真实六份旧producer](../../.evidence/provider-final-readiness/kimi-original-producers-green-v3/receipt.json)全部通过，包含原54项完整原生核验。修正只影响消费者，未使旧厂商事实失效。

该修正改变采集器冻结来源，因此重新构建了正常live-tag二进制供v2引用；新native source=aed5ea7f3187a4cebb189d4b7ed446aea8beb42112eec996e36d059a596596dc，binary=16d68b87bfecd3d0110309670e570882d4e7e94cf03239832e2b37664729b70a。[实际构建元数据](../../.evidence/provider-final-readiness/build-metadata-v2/receipt.json)证明仅-tags=live，不含baseline fixture标签。重建不产生新授权或付费复验。

六路线仍各9/10、目录inactive。将来明确批准并取得六个真实补项后，才能只读组合各自9+1，再独立审阅正式目录/资格；不自动跑六组full formal、其他Provider或旧17路暂停批次。

## 普通验证与来源

最终完整[receipt](../../.evidence/provider-final-readiness/e2e/20261027T133206.932782Z-6ffnipwr/receipt.json) PASS/exit0，墙钟652.483s；[实际选择v2](../../.evidence/provider-final-readiness/full-ordinary-selection-v2.json)含182项输入，完整七包`./ai/e2e ./ai ./ai/internal/testkit/{release,supportmatrix,minimaxlive,jevformal} ./ai/examples/visualeval -count=1 -failfast -timeout=20m`。

主E2E 15533项：15319 PASS、214原opt-in NOT_RUN；新增5个必需case（MMREADY四项、KIMIREADY一项）全部PASS，辅助库存25+32+77+6全部PASS，所有audit零发现。相关e2e/supportmatrix/release的go vet退出0。前次普通PASS、公共布局RED、定向GREEN和所有诊断失败均保留，见[attempt索引](../../.evidence/provider-final-readiness/attempts-index-v1.json)；诊断wrapper的相对路径错误与真实厂商失败明确分开。

实际HEAD保持69ff4af880e68ef3ca245e5128bdae5e59a2ef7b，未提交实现按2034个完整Source.Files和182项实际输入绑定；最终来源/输入与当前完全相等。相对本轮开始仅6个来源文件变化：新增两个E2E与三个最小夹具，修改Kimi producer一个来源域比较；产品运行时、SDK、go.mod/go.sum、正式目录变化0。没有因补测试或消费者修改牵动所有Provider付费复验。

## 保全与交付边界

[最终保护](../../.evidence/provider-final-readiness/history-protection-final-v1.json)：已枚举213269个历史文件全部相同，无缺失/改动；六处公共存储共8156文件，集合及内容完全相同，无增删改。新受控授权/claim仅在私有Git namespace，真实grant/claim为0；全部新旧失败与暂停原件保留，不清理、不移动、不冲销。

矩阵为12个Provider身份、29条操作/Provider/API路径：22条完整适用、6条Kimi9/10、1条MiniMax13/14；TypeSafe两个版本的23条完整路线/型号记录归并为22条路径，不冒称23/29交付。其他Provider已有交付与适用性保留；xAI/地区项目/Kimi的最终发布与人工阶段按各自实际影响收口。

本任务为只读分析、文档和消费者修正，不增加产品共享写状态或负载；本任务不新增race/压力影响，执行仍NOT_RUN。原大功能已有的并发/负载候选继续待人选择，不以本任务不适用将其关闭。最小真实事实缺口共7个，但当前仅Kimi6次/USD0.60具备具体待批准执行范围；MiniMax合同/预算、Kimi晋升及所需人工阶段仍阻塞整体release。

回滚只撤本次两个E2E/三个夹具及producer分栏修正；v1提案和旧拒绝原件保留，停用v2准备范围。保留既有未提交改动及所有原测量、授权和账本，不恢复旧权限。
