# 受控账本回归的 Git grant store 隔离

Type: task
Status: resolved

## Contract

[Spec](../spec.md)；[SOP](../../../docs/agents/verification.md) 第2/4/8节；AGENTS隔离原则。仅修复本轮验证发现的现有GrokHistoryMiniMaxFreshLedger受控native子进程会将合成claim写入共同Git store的问题。每个受控测试必须使用独立私有Git common namespace，保留实际源/commit发现及原有断言；不改真实授权/claim合同，不删除/重置现有记录，不影响厂商或产品行为。

## Failure

18修复后的补充验证选择TestGrokHistoryMiniMaxFreshLedger，既有xaiProcessEnv未隔离GIT_DIR。受控测试实际在共同Git grant store追加3个synthetic grant ID、6个id/scope guard文件，均为synthetic-minimax-{chat,responses,messages}-45b56b12cb611d41，不是真实授权ID。原件native-proof的e2e/20261026T194750.764843Z-cryyxr84，exit0/24.717秒；该额外副作用仍属于验证选择偏差，不能以测试PASS掩盖。没有读取真实Key或请求厂商。精确文件/哈希/session由/tmp/barness-grok-issue18-controlled-store-observation.json列明，仅核对这6项，不扫描无关授权正文。

## Ownership

与18同一implementer，拥有最小fresh-ledger受控进程fixture调用隔离及公开回归；先优先局部隔离该调用，不无谓修改所有Provider helper。Root维护真实偏差与当前状态、map和工单依赖。不得更改真实store、清除synthetic记录、修改真实grant/closed账本、产品或资格消费者。

## Verification

先列隔离失败方式，先在私有测试Git parent/common namespace哨兵中建立公开进程RED，证明旧fixture继承/发现错误store；不得为了重现再次向真实共同store写入。再让受控native子进程只可写其独立Git namespace，原根命令与native预算/来源/真实执行状态保留。反例覆盖环境继承、不同子进程隔离与store文件库存，实际源码/commit必须来自本checkout，不得用合成来源冒充正式资格。修复后才可再次运行原FreshLedger回归，仍使用无Key、无真实付费grant、无厂商请求/claim的受控环境；保留完整RED/GREEN及目标回归原件、命令、退出和audit。禁止删除或重置偏差记录；失败/未知也保留。

## Answer

已完成：软件pin D816、v14独立两轴0/0及完整四包普通15,987 Go测试/1522主必交/127入口/38辅助全部PASS，633.016秒/exit0/audit0；原始失败/偏差及v13错误证明保留。真实资格及人工阶段不由本票授予。

## Comments

- 2026-10-27：本轮选择旧回归时漏查隔离约束；已停止所有同类未隔离执行。该工单不授权新的付费验证、race或压力。

- 2026-10-27：完整[普通核验](../../../.evidence/barness-ai-grok/final-ordinary-validation-v9.json)与[当前状态](../spec.md#唯一当前状态)保存受测dcfa/2062来源/129输入，局部受控正例不构成新来源正式支持。
