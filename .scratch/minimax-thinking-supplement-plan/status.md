# MiniMax thinking-on 补采方案当前状态

2026-10-28：本轮设计与离线核验完成，尚未实施采集器/消费者，也未授权付费执行。推荐固定单项thinking-on、复用旧13项，拟议新增上限1调用/1HTTP/零SDK、环境与场景rerun、USD0.01；原FAIL/exit1、原授权和ledger保持。

当前[最小方案](plan.md)、[范围与预算v2](../../.evidence/minimax-thinking-supplement-plan/scope-and-budget-v2.json)、[离线核验v2](../../.evidence/minimax-thinking-supplement-plan/verification-and-retention-v2.json)。v1数据/草稿及检查脚本路径版本错误的失败日志均保留，当前设计以v2为准；诊断错误不计作厂商结果。

## 设计结论

- 身份固定国际minimax-messages/MiniMax-M3/standard；仅原CRT任务、adaptive thinking、stream=true、2048输出。实际wire389字节，308为Host输入紧凑JSON；最终wire预算仍为2048输入预留。
- 精确新入口拟为^TestMiniMaxThinkingOnSupplement$、count=1、进程180s、调用120s。无需热身厂商调用、额外首等、重试或退避；失败/取消/截断/超预算立即停并消耗本次批准。
- 复用Scope.Scenes绑定单项，仅拟在既有Scope追加BaselineManifestSHA256/MaxUSD两个omitempty字段，不新增Scope.Profile。现有Grant/Claim/Consume/store及ledger链不换；单项与full14授权不能交叉使用。通用时限只改结构边界，原路线构造器仍精确原时限/预算，旧字段零值保持原规范scope/claim keys。
- 冻结M3 USD/M价格为低阶0.30/1.20、最高已知阶0.60/2.40。输入/输出各2048，基础估算USD0.003072、保守已知费率预留USD0.006144，拟议授权USD0.01。未知缓存写价不当免费，失去价格/用量依据则拒绝资格；这不是厂商账单硬封顶。真实账户与价格/输入适用性须在未来批准前核对。
- 新报告仅单项、complete=false。新13+1专用消费者保留完整旧失败包（15条Call/14个唯一场景），只取13个已验证贡献；control首次失败与恢复证明不删，旧thinking-on失败保留并排除贡献。旧producer须证明原exit1，新单项须exit0，不伪造统一来源/时间/预算或14项同进程PASS。
- 组合只有两个非嵌套引用，原件/producer/source.files+selectedInputs/binary/可信claim/ledger全部闭包；两个mode须一致，live+live才可能当前正式资格，controlled+controlled只证明机制。现有M3正式目录事实匹配时可沿用候选新测量，不追加full formal。
- 复用现有受控full14入口产生真实13成功/1失败、exit1基包，再独立运行单项成功；无需新增baseline13采集器或fixture标签。先公共CLI/进程RED/GREEN，再完整受影响普通及工具过程/vet；所有合成授权在私有Git namespace。
- 不改Client/adapter/SDK/目录，不建任意场景选择、通用N+M引擎、图/注册器/调度器、USD累计账本或历史恢复控制器。harness已493行，新逻辑应独立小文件，必要时整理模块边界，保持≤500行。
- 实施前须将受限MiniMax13+1规则补入ADR/SOP；本轮不改已接受决策或现有门禁。未来共享来源变化只重建所需二进制/重填未批准机器范围（包括必要的Kimi引用），不自动使旧厂商证据失效或恢复17路暂停批次。

## 本轮核验与保全

设计数据非可执行grant，未签ID/expiry、无未来supplement路径或新二进制；旧13项适用性直接引用前轮有效原件。预算算术、实际wire、原生产退出与引用、13+1集合、提案非执行性、本地链接、历史字节及存储集合只读核验PASS。

源码2034文件与前轮完整普通来源完全相等，182个原实际输入全相同，HEAD保持69ff4af880e68ef3ca245e5128bdae5e59a2ef7b。沿用前轮完整普通PASS/exit0，不把本轮文档核验冒充新E2E。指定旧输入8387份全部相同、六处公共存储8156文件集合/内容完全相同，无新增/删除/修改；其余历史未写入、移动或清理。

本轮付费授权、真实claim、厂商调用、live/race/压力、目录激活、暂停恢复均为0。实现、实现后普通测试、真实单项与组合正式资格均NOT_RUN，release.pass=false。实施批准和后续一次USD0.01付费选择须分别明确，失败后的另一次请求需新批准，不借旧余额。
