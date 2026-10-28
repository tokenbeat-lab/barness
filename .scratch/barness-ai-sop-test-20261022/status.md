# SOP 普通回归当前状态

日期：2026-10-22（Asia/Shanghai）。请求及影响范围见 [spec.md](spec.md)。本轮普通回归已完成，不修改产品实现，不覆盖原功能的历史状态。

## 来源与范围

- 实际受测 SHA 与当前 HEAD：`68b7da0a3400f909dc535d7f585c93771df1569c`。开始前工作区干净；本轮仅新增测试规格、当前状态和独立证据记录，未提交产品代码变更。
- 1,247 份执行来源与上一普通受测 `3bce7a61cb85662e25faf13137feee723c226503` 的集合/哈希相等；8,636 份历史输入完整固定，前后集合及内容相等。当前提交直接重跑，没有用旧回执替代新执行。
- 执行前冻结的计划、来源、选定输入及必交 case ID：[.evidence/barness-ai-sop-test-20261022/planning/20261022T044225.760126Z-pzsg2xw4/selection.json](../../.evidence/barness-ai-sop-test-20261022/planning/20261022T044225.760126Z-pzsg2xw4/selection.json)。既有测试流程 spec/plan 和 ADR-0051 为范围依据。
- 公共入口实际 argv：`go test -json ./ai/e2e ./ai ./ai/internal/testkit/release ./ai/release/cmd/releasegate ./ai/internal/testkit/supportmatrix -count=1 -failfast -timeout=20m`。采集器普通阶段固定 `GOENV=off`、空 `GOFLAGS`，隔离 HOME，清除继承的差分、压力、live、路线与凭据配置。

## 本轮结果

- E2E：PASS；实际退出 `0`，墙钟 `376.27` 秒；开始 `2026-10-22T04:42:26.487662+00:00`，结束 `2026-10-22T04:48:43.143504+00:00`（原件 UTC 时间）。
- 共 `12,890` 个 Go 通过节点；原生主包 `12,086` PASS / `214` NOT_RUN，脱敏 audit 发现 0。预选的全部 PASS case、NOT_RUN case 和通过测试名集合与实际完全相等；NOT_RUN 不计为完成。原生主 bundle 唯一，本次选择中其他包的退出事件完整，releasegate 包无测试文件。
- 命令与来源回执：[.evidence/barness-ai-sop-test-20261022/e2e/20261022T044226.445355Z-g51l9h61/receipt.json](../../.evidence/barness-ai-sop-test-20261022/e2e/20261022T044226.445355Z-g51l9h61/receipt.json)，SHA256 `7c1a8f49c5c511ae3a878ef269d6bda35d081504a60b3991947b6988c1f31a24`；全部命令、stdout/stderr、数字 exit-code、工具身份及输入哈希由该回执绑定。
- 主原生 manifest：[.evidence/barness-ai-sop-test-20261022/e2e/20261022T044226.445355Z-g51l9h61/bundles/20261022T044231.048899000Z/manifest.json](../../.evidence/barness-ai-sop-test-20261022/e2e/20261022T044226.445355Z-g51l9h61/bundles/20261022T044231.048899000Z/manifest.json)，SHA256 `56d70f46c057d379a17a09349e34b7b1f271b9a1886a6f5abc9569742bb928d3`；audit：[.evidence/barness-ai-sop-test-20261022/e2e/20261022T044226.445355Z-g51l9h61/bundles/20261022T044231.048899000Z/audit.json](../../.evidence/barness-ai-sop-test-20261022/e2e/20261022T044226.445355Z-g51l9h61/bundles/20261022T044231.048899000Z/audit.json)，SHA256 `8aa973100a15bf817985395637131e838cf84e427be4792583def923f020f5fa`。
- 完整统计、包退出、预期集合核对与构建回执：[.evidence/barness-ai-sop-test-20261022/record-20261022T044928.648896Z-oufe7i2v/summary.json](../../.evidence/barness-ai-sop-test-20261022/record-20261022T044928.648896Z-oufe7i2v/summary.json)。

## 独立阶段与只读判定

- live：NOT_APPLICABLE；本轮只执行现有普通回归，没有厂商行为、型号、协议或支持资格变更，离线通过不产生真实资格。
- race：NOT_RUN；本次没有人类对具体 race 范围的选择。候选 argv：`go test -race -json ./ai/e2e ./ai ./ai/internal/testkit/release ./ai/release/cmd/releasegate ./ai/internal/testkit/supportmatrix -count=1 -failfast -timeout=75m`；沿用本轮来源、8,636 份输入及 12,086 个必交 case，具体集合见计划。按普通约六分钟粗估 10–20 分钟，尚未实测 race 开销，原生上限 75 分钟。实际回执已齐，须由人明确选择后才能发起。
- 压力：NOT_APPLICABLE；本轮没有队列、取消、共享状态、内存、限额、复杂度或大负载行为改动。[原测试流程功能](../barness-ai-test-process-review/status.md)尚缺的人工 race、正式压力与首次完整大功能验收保持原状。
- 真实证据清理：NOT_RUN；没有隔离、恢复、观察期或永久删除。
- release.pass：false；人工 race 尚未选择。新不可变索引：[.evidence/barness-ai-sop-test-20261022/record-20261022T044928.648896Z-oufe7i2v/record.json](../../.evidence/barness-ai-sop-test-20261022/record-20261022T044928.648896Z-oufe7i2v/record.json)；只读报告：[.evidence/barness-ai-sop-test-20261022/record-20261022T044928.648896Z-oufe7i2v/report/release-report.md](../../.evidence/barness-ai-sop-test-20261022/record-20261022T044928.648896Z-oufe7i2v/report/release-report.md)；该次真实 argv、数字退出及日志见 [.evidence/barness-ai-sop-test-20261022/record-20261022T044928.648896Z-oufe7i2v/releasegate-command/command-receipt.json](../../.evidence/barness-ai-sop-test-20261022/record-20261022T044928.648896Z-oufe7i2v/releasegate-command/command-receipt.json)。只读命令不会启动后续测试阶段。
- 本次只读 argv：`/Users/cyber/RestoX/barness/.evidence/barness-ai-sop-test-20261022/e2e/releasegate-build-20261022T044307.667662Z-cszm163g/releasegate -record .evidence/barness-ai-sop-test-20261022/record-20261022T044928.648896Z-oufe7i2v/record.json -out .evidence/barness-ai-sop-test-20261022/record-20261022T044928.648896Z-oufe7i2v/report`。
- 上一当前状态快照：[.evidence/barness-ai-sop-test-20261022/record-20261022T044928.648896Z-oufe7i2v/previous-status.md](../../.evidence/barness-ai-sop-test-20261022/record-20261022T044928.648896Z-oufe7i2v/previous-status.md)。回滚只涉及新增测试记录，无产品代码回滚；保留新旧失败及依赖，不执行清理。
