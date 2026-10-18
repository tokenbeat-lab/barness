# MiniMax 交付状态

Delivery: PARTIAL
Updated: 2026-10-18
Integration branch: `codex/minimax-global`
Historical CN checkpoint: `codex/minimax-cn` / `4295ee8`

用户已明确将当前目标改为国际站。票11正在转换Provider/端点/USD价格、证明身份与共享预算，随后继续10/09；原中国区离线通过和实际401保留为历史。**当前完整发布门禁未通过，spec、10和09保持未关闭。**

原CN检查点见[历史机器可读状态](evidence/ticket-09/partial-delivery-status.json)与[ticket10证据入口](evidence/ticket-10/README.md)。

| 范围 | 当前状态 | 证据与边界 |
| --- | --- | --- |
| 原CN三协议与六型号离线约束 | PASS，历史离线范围 | 最新完整MiniMax普通/race均1307/1307；含冻结pi、压力、完整软件回滚，审计0发现。不能证明厂商实际接受请求。 |
| 原CN M3 Chat实际研究 | FAIL，历史范围 | 两次Logical Call、两次Attempt均HTTP401 `authorized_error`；没有实际型号或已知usage。依赖原生工具续接NOT_RUN。 |
| 国际站M3完整候选三身份 | 三项NOT_RUN | Chat研究失败不等于完整候选执行；Responses/Messages完整候选也未执行。 |
| 国际站M3正式三身份及目录激活 | 三项NOT_RUN | BuiltinCatalog未激活M3；没有成功actual origin或正式支持投影。 |
| M3.1线上候选三身份 | 三项NOT_RUN | 未获M Plan用途/资格授权；零实际调用，不虚构公开价格，离线约束已完成。 |
| 原十四身份与四业务效果续期 | NOT_RUN | 尚待最终正式目录/来源确定后重新获取实际证据；六个GLM当前来源适用性校验仍不通过。 |
| 完整ordinary/race发布门禁 | NOT_RUN，未取得发布认证 | 已保留原ordinary失败/中断记录；46个可本地修复case的定向普通/race重验均通过，不能推出完整门禁PASS。 |

原两次实际调用后，另有一次仅本地语法诊断读取同一已选`MINIMAX_KEY`，网络请求为0，未输出Key值、长度或Hash；Bearer前缀、内部空白和控制字符标记均为false。此诊断不增加收费账本，实际消费仍为2次调用/2次Attempt；无Key验证进程各自保持未读凭据。

## 已保存的验证

完整MiniMax离线套件：[普通](evidence/ticket-10/minimax-ordinary-1205300.json)43.000秒、[race](evidence/ticket-10/minimax-race-1205300.json)152.736秒，均1307个case通过；冻结pi差分、压力与完整软件回滚均运行。46个修复后的定向旧case[普通](evidence/ticket-10/repaired-ordinary-1205300.json)/[race](evidence/ticket-10/repaired-race-1205300.json)分别46/46。[无Key live和两种vet](evidence/ticket-10/offline-checks-1205300.json)通过。来源1205300及其后矩阵数据修正的适用范围以10的原始manifest、审计、命令和归档Hash为准，不给历史捕获刷新来源。

无Key live的主manifest为323 PASS/62 NOT_RUN，另有两个辅助PASS；进程正常结束不意味着未执行的真实场景已通过。[当前M3.1三份NOT_RUN](evidence/ticket-10/m31-not-run-1205300-summary.json)及关键来源/子捕获保存在[有明确选择范围的无Key胶囊](evidence/ticket-10/postauth-keyless-1205300.tar.gz)。

旧9481a6a完整ordinary的测试主体已结束，[原manifest](evidence/ticket-10/whole-ordinary-failed-9481a6a.json)为12304 PASS/52 FAIL；进程在后续阶段因已知harness修复被停止，exit=-15，不能声称完整普通套件通过。独立redaction audit随后完成，0发现；这是独立派生审计，不修改原退出状态。[计时调查](evidence/ticket-09/whole-ordinary-timing-investigation.md)没有支持E03或精确小数解析死循环的证据。52个失败中的46个本地可修复case已定向重验，其余六个GLM来源校验需真实续期，不能放宽适用性规则。

## 历史证据完整性

原完整失败bundle已在主工作区本地保留，110MB归档不纳入Git；ticket10的[有限索引](evidence/ticket-10/whole-ordinary-failed-9481a6a.json)与[完整文件Hash清单](evidence/ticket-10/whole-ordinary-failed-9481a6a-files.json.gz)记录其位置和SHA。原131505个文件未改写，归档额外保留一个21字节的派生审计文件。

98236个声明的artifact引用全部匹配，零缺失、零artifact不符。另349项fixture来源归属Hash与脱敏后的artifact文件Hash不同，二者按既有语义分别记录；不能合并两种Hash后推断捕获被篡改。[独立只读复核](evidence/ticket-09/whole-ordinary-artifact-integrity.md)通过单case来源和登记替换的逐字节复现解释了该区别，见[版本与来源](evidence/ticket-09/whole-ordinary-artifact-integrity-provenance.json)。该完整性结论不改变原52个FAIL或中断退出。

复核脚本记录的是调查时路径。可将保留归档解压到临时目录，并把脚本的`bundle`指向该目录后只读复核；不需要Key、厂商调用或重跑测试。

## 待继续工作

1. 按[票11](issues/11-international-target.md)完成国际站身份、三端点、USD价格、frozen pi与离线复验；沿用户指示使用单一MINIMAX_KEY做国际站标准M3三协议验收，不再等待中国区修复，不推断M3.1资格。
2. 先对新的已提交源码重新prepare、无Key检查和精确来源审查；旧9481a6a批准不适用于新HEAD。随后使用独立国际站研究intent与自身前驱继续三候选真实验收；保留原CN401及所有失败，不将其作为跨区native前驱，不覆盖既有intent。候选阶段已消费2/120调用、2/144 Attempts；该Chat身份2/20、2/24；正式阶段0/60、0/72。不得重置账本或降低验收。
3. 三候选各自通过后才作数据目录激活，获取三份自身正式报告；再续期原十四身份和四业务效果，运行完整ordinary/race/压力/pi、两种vet、逐包脱敏和回滚门禁。
4. 对固定基点`ecfd4b4`完成最终独立Standards/Spec审查及必要修复重验，全部必交通过后才关闭spec和10/09。

历史、受控或回放结果均不转成actual/formal PASS。历史追溯表继续保留其原来源和状态；当前完整交付判断只依据实际适用证据。
