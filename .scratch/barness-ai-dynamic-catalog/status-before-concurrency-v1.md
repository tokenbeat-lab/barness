# 模型目录维护与本地加载：当前状态

Updated: 2026-11-01
Implementation: COMPLETE
Ordinary verification: PASS
Release: false

## Scope and source

- 集成分支：`codex/barness-catalog-integration`。规格及工单01–05均已实施并解决；未合并其他用户分支或发布。
- 基准：`9a43ec0827b0ab794501682a7e66037a592a2f8d`。实际完整受测提交：`d695b3c5ddf113ac4e3c81c6659461b288bb977e`。
- 当前提交与受测提交：收尾仅修改本功能 Markdown/索引。最终提交见只读报告的 `releaseCommit`；其2153个执行来源文件及150个选定输入的完整集合/摘要相等性，见 `.evidence/barness-ai-dynamic-catalog/metadata/20261101T122739.037712Z-28e2e9d7/source-closure-v1.json`。
- 交付：独立本地加载/隔离报告；按Provider维护的唯一数据源与同源生成；完整维护CLI；冻结候选、逐身份受控合约、check/apply、导出/回退与可重放来源；维护指南见 [catalog-maintenance](../../docs/barness-ai/catalog-maintenance.md)。
- 初始118条定义（112聊天、3图片、3分类）的顺序、字段、null/缺省及价格保留。当前默认整体身份由维护内容生成；中文评测增加v3配置，当前用量断言改用当前快照。旧快照、旧配置、实测原件、pin与支持矩阵保持历史事实。
- 完整输入：49个本功能实现/数据/来源审阅文件，加101个旧公共测试实际执行脚本和导入夹具；按文件固定150项。默认来源覆盖ai、验证工具及go.mod/go.sum，共2153文件。完整清单在阶段回执与 [发布记录](release-record-v1.json)。

## Task graph and review

01加载器、02维护源、03维护CLI、04候选合约/check/apply、05集成审查均 `resolved`，见 [issues](issues/)。双轴审查原问题及后续回归均已修复，最终0open，见 [v4](implementation-review-v4.md)、[v3](implementation-review-v3.md)、[v2](implementation-review-v2.md)。原问题报告 [v1](implementation-review-v1.md)保留。

最终静态原件：`.evidence/barness-ai-dynamic-catalog/metadata/20261101T122739.037712Z-28e2e9d7/review-standards-v4.md`（SHA256 `175819bb981a1eda3b878f999e9b06ca585d029d86f80ed14ab9f1c0d82cc685`）、`review-spec-v4.md`（SHA256 `d501f0231aa27d0970f112ec2db87c8af3542a0129d2a3c9669411cde5160781`）。审查不代替执行或厂商事实。

## Stages

| 阶段 | 实际状态 | 范围/依据 |
| --- | --- | --- |
| ordinary | PASS / exit0 | 完整受影响包；17705个Go测试事件PASS、240个SKIP，无FAIL；墙钟1210.921192秒 |
| native evidence | PASS | 唯一主包16862 PASS、222 NOT_RUN；3个受控辅助包115 PASS；全部audit无发现 |
| build | PASS / exit0 | `go build ./ai/... ./catalog/... ./cmd/...`，2.271177秒 |
| vet | PASS / exit0 | `go vet ./ai/... ./catalog/... ./cmd/...`，1.761515秒 |
| gate binary build | PASS / exit0 | 冻结来源编译releasegate，2.140633秒；二进制/实际argv见元数据回执 |
| live | NOT_APPLICABLE | 无新增厂商事实需求；既有定义精确迁移，路由/分类判断为已审等价提取，新增接纳验证不改变有效旧请求；来源有限审阅见source-review及v2/v3。受控live-tag子包不取得真实资格 |
| race | NOT_RUN | 新增存储锁、状态读写和check取消/收尾须另行人工选择；实施批准不等于执行批准 |
| pressure | NOT_RUN | 加载/执行资源及并发负载尚未人工选择；race前置未完成，无真实压力生成 |
| read-only release | false | 适用race/压力未运行；实际数字退出、五项检查及报告见最终机器判定原件 |

完整普通实际命令：

```sh
go test -json -failfast ./ai/e2e ./ai ./catalog/... ./cmd/... -count=1 -timeout=30m
```

由`tools/verification/run.py --stage e2e --timeout 2100`采集；全部150个`--input`及实际collector argv见 `.evidence/barness-ai-dynamic-catalog/metadata/20261101T122739.037712Z-28e2e9d7/ordinary-invocation.json`。GOENV=off、空GOFLAGS、隔离HOME及清除凭据/live/race/压力/冻结oracle配置。240个Go跳过项和222个原生NOT_RUN仍保持未运行，不将其计为PASS；完整名称与原因见原始stdout和原生清单。

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
