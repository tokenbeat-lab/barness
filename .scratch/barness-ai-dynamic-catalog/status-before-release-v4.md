# 模型目录维护与本地加载：当前状态

Updated: 2026-11-01
Implementation: COMPLETE
Ordinary verification: PASS
Release: 见只读报告

## Scope and source

- 集成分支：`codex/barness-catalog-integration`。规格及工单01–05均已实施并解决；未合并其他用户分支或发布。
- 基准：`9a43ec0827b0ab794501682a7e66037a592a2f8d`。实际完整受测提交：`d695b3c5ddf113ac4e3c81c6659461b288bb977e`。
- 旧受测提交 d695 与当时文档收尾提交的2153个来源及150输入相等性见 `.evidence/barness-ai-dynamic-catalog/metadata/20261101T122739.037712Z-28e2e9d7/source-closure-v1.json`。新增测试后完整普通受测 c4d；后续测试helper变更后，针对性普通、race、压力均受测 f397，2154来源及150输入相等。
- 交付：独立本地加载/隔离报告；按Provider维护的唯一数据源与同源生成；完整维护CLI；冻结候选、逐身份受控合约、check/apply、导出/回退与可重放来源；维护指南见 [catalog-maintenance](../../docs/barness-ai/catalog-maintenance.md)。
- 初始118条定义（112聊天、3图片、3分类）的顺序、字段、null/缺省及价格保留。当前默认整体身份由维护内容生成；中文评测增加v3配置，当前用量断言改用当前快照。旧快照、旧配置、实测原件、pin与支持矩阵保持历史事实。
- 完整输入：49个本功能实现/数据/来源审阅文件，加101个旧公共测试实际执行脚本和导入夹具；按文件固定150项。默认来源覆盖ai、验证工具及go.mod/go.sum；旧完整普通时2153文件，当前受测来源2154文件。完整清单在阶段回执与 [发布记录](release-record-v1.json)。

## Current concurrency verification selection

用户已选择按受影响范围先 race 后压力。新增公共测试覆盖4并发加载与查询、独立报告、共享输入不变，以及4个CLI争用同一已check候选时唯一成功、busy/stale拒绝、current/历史/输出一致；已有受控取消/超时场景继续覆盖。race以显式测试flag构建子CLI，内部固定check contract子进程保持普通模式；外层直接运行候选合约。压力选择512和2048条目录各8个并发读取、4个写者2轮，记录实际墙钟及 Go test 进程的 TotalAlloc（不含 CLI 子进程资源，不是峰值）；正确性和有界完成为断言，不设新 SLA；不运行全局Loopback或不相关Provider压力。此处为选定范围，非执行通过。新测试来源须重新完整普通验证。旧状态原字节见 [快照](status-before-concurrency-v1.md)。

已提前终止来源将被修订的普通尝试：`.evidence/barness-ai-dynamic-catalog/e2e/20261101T130425.124309Z-t0psx9k6/receipt.json`，CANCELLED/exit -15；原件保留。

## Race failure and harness repair

c4d0a905 的完整普通回执 `.evidence/barness-ai-dynamic-catalog/e2e/20261101T130626.181468Z-yicjfc4x/receipt.json` 为 PASS / exit 0；主原生包16864 PASS、226 NOT_RUN，所有四包audit无发现。随后race中 `TestCatalogCandidateCheckApply` 失败（check进程返回旧helper约定的 -1，stdout/stderr为空），collector按失败即停而CANCELLED / exit -15：`.evidence/barness-ai-dynamic-catalog/race/20261101T132656.875761Z-n5x6ljbv/receipt.json`。旧证据缺单次耗时、context错误和进程错误，不能证明超时；父输出无 DATA RACE 不构成阶段通过。

仅修测试helper：显式race子进程模式中CLI调用超时60秒，普通仍10秒；保留启动状态、native退出码或null、signal、context和exec错误、时间、argv、stdout/stderr。产品check超时及受控取消断言不改变。针对性普通回执 `.evidence/barness-ai-dynamic-catalog/e2e/20261101T133048.622774Z-th0j6g53/receipt.json` PASS / exit0。机械编辑语法错误的独立失败 `.evidence/barness-ai-dynamic-catalog/e2e/20261101T132950.081295Z-ynwro7rj/receipt.json` 保留。该测试helper变更使旧完整普通来源不相等，未以旧原件冒充新来源完整验证。后续race重跑及压力结果见下；原失败事实保留。

## Ordinary timeout correction

补跑完整普通 `.evidence/barness-ai-dynamic-catalog/e2e/20261101T135202.402804Z-gow1f7pd/receipt.json` FAIL / exit1：`TestCatalogMaintenanceCLIAddedRollbackRecovery` 的check进程实际耗时10.00257秒，context deadline exceeded，signal killed，native退出码为null。该次有直接证据证明普通helper的10秒截止；不同于先前race失败根因仍未证实。仅将目录CLI helper总预算改为有界60秒，race原为60秒；产品check timeout、受控取消断言和全局ctxFor不变。修复后150输入针对性普通 `.evidence/barness-ai-dynamic-catalog/e2e/20261101T135707.701452Z-2fmrouyr/receipt.json` PASS / exit0。因为测试helper源码变化，既有f397 race和压力PASS仍为历史事实，不冒充新来源结果；重新按普通→race→压力执行同一已授权范围。旧状态快照 [status-before-timeout-v3](status-before-timeout-v3.md)。

## Race and pressure results (prior source)

用户选择重跑race，通过后继续目录压力。实际受测提交均为 `f3977c804bcc9fc9b8abf018e6328934c43639d9`，2154源码文件和150选定输入集合/哈希相等。原状态快照见 [status-before-race-v2](status-before-race-v2.md)。

- race：`.evidence/barness-ai-dynamic-catalog/race/20261101T133622.365335Z-72f1_lun/receipt.json` PASS / exit0，740.07秒；主包1140个case PASS，audit无发现；118身份1020直接合约全部通过，51条Go构建argv含-race、34个主CLI buildinfo含-race。父子实际输出632份扫描无DATA RACE，清单 `.evidence/barness-ai-dynamic-catalog/metadata/concurrency-verification-v1/race-output-scan-v2.json`。内层固定check contract为普通Go子进程，不声称其插桩；外层直接执行相应合约。
- pressure：`.evidence/barness-ai-dynamic-catalog/pressure/20261101T134937.254917Z-xdesnefv/receipt.json` PASS / exit0，26.19秒，四个case及audit通过。512条×8读取耗时0.0453秒、Go test进程TotalAlloc 120142424字节；2048条×8耗时0.1118秒、TotalAlloc 498608160字节；两轮4写者各唯一成功，分别0.943秒和0.880秒。TotalAlloc不包括CLI子进程，不是峰值或SLA。
- f397当前来源的针对性普通补充 `.evidence/barness-ai-dynamic-catalog/e2e/20261101T133244.463829Z-xrnu_jus/receipt.json` PASS / exit0，完整绑定150输入。c4d完整普通原件与之来源不同，不伪称完全相等。

## Final same-source verification

冻结受测提交 `ada85f43923e5ca35e53de5f478947e1386bd7b0`。普通、race和压力的2154源码文件及150选定输入完整集合和哈希相等。旧状态见 [status-before-final-v4](status-before-final-v4.md)。

- 完整普通：`.evidence/barness-ai-dynamic-catalog/e2e/20261101T135755.536051Z-xsulytv4/receipt.json` PASS / exit0，1222.76秒，主包16864 PASS、226 NOT_RUN，四包audit无发现。此前实际触发10秒截止的回滚恢复case本次通过。
- race：`.evidence/barness-ai-dynamic-catalog/race/20261101T141828.301958Z-c2ua9urt/receipt.json` PASS / exit0，752.97秒；1140原生case、118身份和1020直接合约通过。51个build argv均含-race，34个CLI buildinfo均含-race；其中只读build仅留error空且对应case通过，未保存数字退出，不补造。实际输出632文件无DATA RACE，清单 `.evidence/barness-ai-dynamic-catalog/metadata/concurrency-verification-v1/race-output-scan-v3.json`。固定内层check contract仍为普通Go，外层直接运行相关合约。
- 压力：`.evidence/barness-ai-dynamic-catalog/pressure/20261101T143149.202907Z-fbluvlt1/receipt.json` PASS / exit0，26.02秒，四case和audit通过。512条×8读取42.58ms、Go test进程TotalAlloc 120674784字节；2048条×8读取101.90ms、TotalAlloc 498615416字节；4写者两轮各唯一成功，耗时0.900秒及0.882秒。TotalAlloc不含CLI子进程，不是峰值或SLA。

## Task graph and review

01加载器、02维护源、03维护CLI、04候选合约/check/apply、05集成审查均 `resolved`，见 [issues](issues/)。双轴审查原问题及后续回归均已修复，最终0open，见 [v4](implementation-review-v4.md)、[v3](implementation-review-v3.md)、[v2](implementation-review-v2.md)。原问题报告 [v1](implementation-review-v1.md)保留。

最终静态原件：`.evidence/barness-ai-dynamic-catalog/metadata/20261101T122739.037712Z-28e2e9d7/review-standards-v4.md`（SHA256 `175819bb981a1eda3b878f999e9b06ca585d029d86f80ed14ab9f1c0d82cc685`）、`review-spec-v4.md`（SHA256 `d501f0231aa27d0970f112ec2db87c8af3542a0129d2a3c9669411cde5160781`）。审查不代替执行或厂商事实。

## Stages

| 阶段 | 实际状态 | 范围/依据 |
| --- | --- | --- |
| ordinary | PASS / exit0 | ada85f43完整受影响包，17708 Go PASS、244 SKIP，1222.76秒；主包16864 PASS、226 NOT_RUN，4包audit无发现 |
| native evidence | PASS | 唯一主包16862 PASS、222 NOT_RUN；3个受控辅助包115 PASS；全部audit无发现 |
| build | PASS / exit0 | `go build ./ai/... ./catalog/... ./cmd/...`，2.271177秒 |
| vet | PASS / exit0 | `go vet ./ai/... ./catalog/... ./cmd/...`，1.761515秒 |
| gate binary build | PASS / exit0 | 冻结来源编译releasegate，2.140633秒；二进制/实际argv见元数据回执 |
| live | NOT_APPLICABLE | 无新增厂商事实需求；既有定义精确迁移，路由/分类判断为已审等价提取，新增接纳验证不改变有效旧请求；来源有限审阅见source-review及v2/v3。受控live-tag子包不取得真实资格 |
| race | PASS / exit0 | ada85f43同范围752.97秒；118身份1020合约，1140原生case，输出632文件无DATA RACE |
| pressure | PASS / exit0 | ada85f43四case通过，26.02秒；512/2048条各8读取，4写者两轮 |
| read-only release | 见报告 | 同源完整普通、race与压力的新记录及实际结果见 `.evidence/barness-ai-dynamic-catalog/metadata/concurrency-verification-v1/releasegate-report-v4/release-report.json`；旧v3因来源差异拒绝的报告保留 |

完整普通实际命令：

```sh
go test -json -failfast ./ai/e2e ./ai ./catalog/... ./cmd/... -count=1 -timeout=30m
```

由`tools/verification/run.py --stage e2e --timeout 2100`采集；全部150个`--input`及实际collector argv见 `.evidence/barness-ai-dynamic-catalog/metadata/20261101T122739.037712Z-28e2e9d7/ordinary-invocation.json`。GOENV=off、空GOFLAGS、隔离HOME及清除凭据/live/race/压力/冻结oracle配置。旧完整普通240个Go跳过项和222个原生NOT_RUN仍保持未运行，不将其计为PASS；新完整普通为244个Go跳过项和226个原生NOT_RUN，增加的4项是新增目录压力case，已在压力阶段实际PASS。独立审计报告及JSON保存在 `.evidence/barness-ai-dynamic-catalog/metadata/concurrency-verification-v1/`，SHA256分别为 `1d4537f40cc155fb56a74c2307b178bee8a17f52daa219f51be425d3872e5ab0` 和 `f898113d35497570191057403ebeb50641940f19efbb8f96766fed67befbf368`。

## Evidence

- [最终不可变普通索引](final-verification-v1.json)，SHA256 `47a3b0269303b88214fc9e06d0df1ebee25b50519c4e47ec21f33b2c3b4e2c3c`。
- 主阶段：`.evidence/barness-ai-dynamic-catalog/e2e/20261101T122739.442924Z-ztjtxwmz/receipt.json`，SHA256 `5f2f28cd00598b8338a07d2cbafd3fce3898994d2bb791929808eadf0359b52a`，PASS/exit0；起止与墙钟、工具身份、实际argv、全部输入/来源和四个原生包均在原件。
- 主manifest：`.evidence/barness-ai-dynamic-catalog/e2e/20261101T122739.442924Z-ztjtxwmz/bundles/20261101T122744.713185000Z/manifest.json`，SHA256 `27cfbac9a0e0a4e5f0fcfd6b5d1fca6ff9dd84fac23b76cb7f7da1f6b4a25ea6`；audit SHA256 `8aa973100a15bf817985395637131e838cf84e427be4792583def923f020f5fa`。
- 已核对118个身份、1020个逐场景Go测试及原生case实际PASS，完整计划与加载报告匹配；各前序工单1299个必需case均在最终主包PASS。空目录另由公共完整check/apply及固定空合约证据核验。
- build/vet/判定工具编译的元数据回执均位于 `.evidence/barness-ai-dynamic-catalog/metadata/20261101T122739.037712Z-28e2e9d7/`，来源和150个输入与完整普通一致；它们不冒充主E2E。
- [只读发布输入](release-record-v1.json)绑定当前状态内容摘要、阶段原件和有限完整requiredCases。最终机器判定：`.evidence/barness-ai-dynamic-catalog/metadata/20261101T122739.037712Z-28e2e9d7/release-evaluation-v1.json`；报告：`.evidence/barness-ai-dynamic-catalog/metadata/20261101T122739.037712Z-28e2e9d7/releasegate-report-v1/release-report.json`。判定不会启动后续测试或厂商请求。
- race候选的具体完整argv、150输入、预计耗时与保留/未覆盖范围：`.evidence/barness-ai-dynamic-catalog/metadata/20261101T122739.037712Z-28e2e9d7/manual-race-proposal-v1.json`；它尚未被人类选择，CLI子进程插桩及压力负载仍未选定，不作完整race资格承诺。

## Retained failures and process facts

所有RED、失败、中断、原件及传递输入保留，不覆盖或清理。首次全量中文旧配置失败/中断见 [记录](integration-interrupted-v1.json)；主agent在回滚观察期间创建文档导致的独立FAIL见 [记录](integration-workspace-mutation-v1.json)，已改为运行前提交全部文档；最终旧用量pin FAIL见 [记录](integration-pin-consumer-failure-v1.json)。后二者及中文流程均已在最终完整运行PASS。

工单04专门apply E2E晚于该存储实现，组合CLI RED先于完整流程；原顺序如实保留，未补造专门RED或补写单元测试，见 [04](issues/04-candidate-contract-check.md)。早期加载器全包来源拒绝与诊断panic、后续精确来源修复和全部RED/GREEN索引保持原样。

## Worktree cleanup and rollback

修复工作树已由应用归档，list_artifacts确认；其余5个工作树归档请求均被应用拒绝：`Another active task still uses this worktree.`。所需终态证据/传递输入已复制并核验，全部代码已合并；未绕过占用保护或手工删除。原始操作与归档观察见 `.evidence/barness-ai-dynamic-catalog/metadata/20261101T122739.037712Z-28e2e9d7/worktree-cleanup-v2.json`。

本轮仅交付集成分支。维护数据回退按指南创建恢复旧内容的新候选，重新check后apply，保留追加历史；不手工覆盖current。软件回滚可切回原功能分支/固定基准审阅恢复；本轮未执行真实数据清理或发布。普通完成、未运行人工阶段与完整发布通过分别表达：`release.pass=false`。
