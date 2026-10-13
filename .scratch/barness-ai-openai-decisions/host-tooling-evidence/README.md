# 工单 11：Host 迁移与中文评估工具交付

2026-10-13 完成；真实质量为 NOT_RUN，由票 12 执行。fixture 只证明流程，不写支持矩阵、
不推断型号效果、不共享 Jev threshold。原 Jev 默认/schema 1 和 24 样本数据集 bytes 保持。
公开命令与全部报告字段见 [CLI README](../../../ai/examples/chineseeval/README.md)，
local/cloud 示例见 [迁移文档](../../../ai/examples/classifiermigration/README.md)，
架构取舍见 [ADR-0029](../../../docs/adr/0029-barness-ai-host-evaluation-routes-and-refusals.md)。

## 最终独立报告

`final/` 下每个 bundle 都有自己的 route/config/catalog/code Hash、实际 ResponseModel、
完整样本、Observer、实际传输 captures、manifest/files inventory 和 audit：

| 目录 | 身份与状态 | 普通/拒答 | 总准确率 | 条件准确率 | 已知费用 USD |
|---|---|---:|---:|---:|---:|
| final/jev | Jev fixture COMPLETE | 24/0 | 0.25 | 0.25 | 0.0001008 |
| final/luna | Luna mixed fixture COMPLETE | 12/12 | 0.125 | 0.25 | 0.00024 |
| final/all-refusal | Luna fixture COMPLETE | 0/24 | 0 | null | 0.00024 |
| final/not-run | Luna live identity NOT_RUN | 0/0 | null | null | 0（没有 Attempt） |

精确金额与 bins/latency 在各报告；表格舍入不替代核验。普通 fixture 答案固定 billing，
不读取 gold 标签。全拒答率 1，条件准确率/ECE/Brier 为 null，费用照实保留。
`fixture-summary.json` 汇总上述各自 Hash 与指标。

```sh
for bundle in .scratch/barness-ai-openai-decisions/host-tooling-evidence/final/*/*; do
  go run ./ai/examples/chineseeval/cmd/chineseeval -verify "$bundle"
done
go run ./ai/examples/chineseeval/cmd/chineseeval \
  -verify .scratch/barness-ai-pi-1.0/chinese-evaluation-evidence/live
```

旧 Jev 原始历史包未改变，按自身 2026-10-08.5 recorded snapshot 继续核验；可核验不赋予
当前发布适用性。`initial/` 是审查前保留的中间包，早于最终 captures 合同，不能当作最终验收。

## 验证和反例

- `target-e2e.zip`：97 PASS 的完整实际专项 E2E bundle，含 CLI 两 route、混合/全拒答、
  usage partial/unreported、零 Attempt/已发送取消、有限预算、旧默认/历史以及输入/报告/verifier
  反例；local/cloud 类型 switch、nil/协议 Options、结构化说明、图片三重授权也在其中。
- `failure-iterations.zip`：达到 manifest 的 red 迭代原始证据，成功没有覆盖失败。
- `full-suite-summaries.zip`：完整普通/race 的原 manifest 与 audit 汇总；所有原始 case 仍在
  `validation.json` 指明的本地 `.evidence` 路径，此 zip 不含全部 case。
- `timeout-failure.zip`：普通全量中既有 body-call-deadline 的原失败 case；另建
  [needs-triage issue](../../barness-ai-timeout-ordering/issues/01-call-and-read-idle-deadline-order.md)。
- `REVIEW.md`：独立 Standards/Spec 审查与所有修复，最终无剩余发现。

```sh
go test ./ai/e2e -run '^Test(ChineseEvaluation|ClassifierHostMigration)' -count=1
BARNESS_AI_PIDIFF=1 go test ./... -count=1
go test -race ./... -count=1
go vet ./...
go vet -tags live ./...
```

最终专项 97 PASS；完整 race 退出 0，6,494 PASS/1,050 原有 opt-in NOT_RUN。普通全量开启
冻结 Pi，7,356 PASS/187 NOT_RUN/1 FAIL：既有测试的 CallTimeout 与 ReadIdleTimeout 同为
100ms，得到 idle/error 终态而非断言期待的 call/aborted。原单项复测 PASS；没有修改断言或
协议实现，原全量仍标 FAIL。精确命令、退出状态、完整源 manifest Hash 和审计见 `validation.json`。
两种 vet 通过，所有包审计无发现。没有声明发布门禁通过或实际型号质量已验证。

`SHA256SUMS` 覆盖交付文件；`manifest.json` 是本交付目录 inventory，`audit.json` 是重新
审计结果。归档源包分别已审计；校验归档后可解压重放其 assertions 中的精确命令。
