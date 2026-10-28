# MiniMax 单项实施当前状态

2026-10-28：本次实施与受控/普通验证完成。ADR-0055、SOP、工具说明及授权范围先修订，固定 thinking-on 单项与十三加一只读组合已实现。真实补项仍未执行，全部付费授权未签发；release.pass=false。

[完整验证与保全](../../.evidence/minimax-thinking-supplement/verification-and-retention-v1.json)、[七份待批准范围 v3](pending-scopes-v3/README.md)、[29条路径矩阵 v3](delivery-matrix-v3.md)。前次状态、v1/v2提案、所有历史失败和暂停状态保留；本状态只描述本次实现，不把受控事实晋升为真实交付。

## 实现

- 唯一 profile minimax-messages-thinking-on/v1；国际Messages/M3/standard/candidate；精确 TestMiniMaxThinkingOnSupplement、count1、180s，调用120s，现有20s收尾余量。复用原CRT/adaptive/2048，1调用/1HTTP、零SDK/环境/请求/工具重跑。单项 complete始终false，不冒充完整十四项。
- 复用 livegrant schema1、原claim算法/存储及MiniMax账本，仅追加两个omitempty字段；原grant的规范scope和claim key已用真实旧批准在原时刻核验。通用duration正值≤1800，各原路线构造仍精确1800。单项在追加前约束输入/输出、调用/Attempt，在发送前检查最终wire与USD预留。
- 固定13+1消费者保留旧FAIL/exit1、15条Call及原预算，贡献Plan减thinking-on的13项与独立新单项。核对封存/完整manifest、当前编码与响应解释、原授权/权威ledger、tenant/account/SDK及真实producer。旧失败回执bundles为空，按原stderr唯一精确目录和同session claim证明归属；新成功包必须有真实库存。不复制旧包、不改报告或预算、不建通用组合框架。
- releasegate inspect/native consumer/input closure/record均接入专用分派。受控组合可得CONTROLLED_COMPLETE，正式消费拒绝并保留传递输入库存；单包及混合mode不能取得完整资格。harness由493行缩至378行，请求构建独立为130行，所有新模块低于500行。

## 验证

完整普通[receipt](../../.evidence/minimax-thinking-supplement/e2e/20261028T003544.005787Z-gn7a483z/receipt.json) PASS/exit0，墙钟681.367s；完整八包包括既有七包及livegrant，count1/failfast/timeout20m，实际184项输入。主E2E 15563项：15349 PASS、214原opt-in NOT_RUN，FAIL0；本次29个有实际断言的必需case全部PASS，所有辅助库存和audit通过。

56个工具测试PASS，普通相关包及live标签包编译/vet均exit0。定向RED/GREEN与全部诊断失败保留，见[attempt索引](../../.evidence/minimax-thinking-supplement/attempts-index-v1.json)。失败包括未知profile、真实失败producer布局、来源/退出反例和null/缺失计费counter，均非新增厂商调用。

真实旧MiniMax生产消费者[复核](../../.evidence/minimax-thinking-supplement/original-baseline-v5/receipt.json)保持13项适用、原15 Call/14 Attempt、FAIL/exit1、预算不改；单独完整资格仍拒绝。原Kimi六份/54项及真实producer来源只读复核PASS。当前普通与冻结来源完整相等，HEAD保持69ff4af880e68ef3ca245e5128bdae5e59a2ef7b；2047项来源、184项实际输入匹配，32项变化仅在验证/采集/消费者，产品运行时/SDK/目录/价格变化0。既有未提交改动保留。

## 冻结和待批准范围

[冻结索引](../../.evidence/minimax-thinking-supplement/build-v1/freeze.json)：仅正常-tags live，无baseline fixture标签。MiniMax source=5ae81bd588a20f96fba22d0f63c5682bc890fee1e7ac0ff5cbcee6c8b00ef027、binary=6591a1c3afd5887ba672ba766a807a5bca1601939e711caba3ff7911d0f02814；Kimi source=0cf17ddcd89f3728fee695a865be08b077797791ca1090f4d345cde1dc4cbd6e、binary=f939167130457cc8e5ad7a6bbf4eb2ea2c93b3ba7a9878a40f9d00d55927ae6a。七条单项受控执行均与各自冻结二进制字节相同，见[对应原件](../../.evidence/minimax-thinking-supplement/frozen-producer-controlled-v1.json)。

MiniMax仅thinking-on 1调用/1HTTP、输入/输出各2048、120s/180s、拟议USD0.01，最高已知费率预留USD0.006144；四个计费counter缺失/null或正缓存写价未知拒绝，失败不退权限。Kimi六份仍各1调用/1HTTP、4096输入/输出、120s/240s、60s首发等待、各USD0.10，总USD0.60。七份合计拟议USD0.61，范围独立；全部approved/executable/grantIssued=false，ID/expiry未签发，[只读验证](../../.evidence/minimax-thinking-supplement/proposal-validation-v1/receipt.json)确认不能授权执行。未来须可信Host确认真实同账户、standard用途及实际价格/服务等级，再独立批准；不是账单硬封顶。

29条路径仍22完整适用、Kimi6条9/10、MiniMax1条13/14；TypeSafe两个版本对应23条完整路线/型号记录，不冒充23条路径。剩余最小真实补采仍7项；MiniMax只读真实13+1、Kimi真实9+1及自身目录审阅，和各项目人工阶段尚未完成。重建采集器、更新未批准范围不使原54项或其他Provider失效，不重跑所有Provider、不恢复17路暂停批次。

## 保全、未执行和回滚

[保全v2](../../.evidence/minimax-thinking-supplement/history-protection-final-v2.json)：213269项历史保护及8387项本轮保护核对，无缺失/无未解释改动；本次授权修订ADR/SOP两份原字节保存在initial副本，不能将文档修订谎报为原路径未改变。公共授权/账本六处8156文件集合与内容完全相同。新合成grant/claim仅私有Git namespace，写前断言；真实claim、付费grant、厂商调用均0，历史授权与ledger未刷新、未清理、未移动。

live/race/压力均NOT_RUN。后续race候选限单项claim/重复启动、预算追加及取消/超时/释放；压力候选限现有MiniMax共享资源和取消路径，不增加新负载框架。本次批准不启动这些阶段。只读发布收口见../../.evidence/minimax-thinking-supplement/release-v1/release-report.json；未执行事实继续阻塞release。

回滚只撤本次32项实现/测试来源及ADR/SOP/工具说明修订，停用v3未批准提案和新冻结；原文件副本、旧scope、全部失败与账本保留。无产品/SDK/数据库变更，不引入兼容shim或恢复控制器。
