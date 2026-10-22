# Ticket 06 qualification ordinary verification

Kimi 复用原生 regional 生命周期，拥有独立六路线、候选事实、9 场景合同、USD 上界和 one-use ledger；旧路线 JSON、预算与账本合同保持原值。新增 `regionalpromote -kimi` 独立 proof-gated proposal/activation；当前 `ai/catalog_kimi.json` 保持 inactive，BuiltinCatalog 无 Kimi。

最终受测提交：`9c3409d15dea4e6df52cd3a1a0d1db48ece45ef1`。最终命令：

```sh
python3 .scratch/barness-ai-test-process-review/tools/run.py --feature barness-ai-kimi-qualification --stage e2e -- go test -json ./ai/e2e -run '^Test(KimiQualification|Regional)' -count=1 -timeout=10m
```

最终退出 0：814 个叶测试 PASS，其中 Kimi qualification 74 个。主 E2E bundle 与 audit 通过；六个实际受控原生进程各完成 9 场景、独立只读消费者通过。验证覆盖 grant/cap/source/binary/expiry/one-use、实际 L01 前缀、工具配对、取消与恢复、原始响应 replay 和 CostEstimate/Observer 快照；篡改及受控/缺失/重复/部分/过期来源晋升均拒绝。

[最终原件回执](../../../.evidence/barness-ai-kimi-qualification/e2e/20261022T151141.555350Z-m5y7s4_e/receipt.json)；[不可变收尾索引](ticket-06-closeout-20261022T1512.json)；[先前 30 次 attempt 索引](ticket-06-retention-20261022T1509.json)。1339 个完整执行来源文件与最终提交集合/哈希相等，选定外部输入为空。索引保留实际测试 SHA；后续 scratch 文档提交不修改执行来源。全部 31 次 attempt、失败/构建失败、原生子进程、grant/ledger receipt 和消费者原件已复制到主工作区原相对路径并逐文件核验，共 170374 文件，inventory SHA-256 `61d42b22bf0795291f2c6528cd3c5e877fab14970fd8e0cb2512eb6b05497b93`。

补充未过滤命令 `go test -json ./ai/e2e ./ai ./ai/internal/testkit/supportmatrix ./ai/internal/testkit/release ./ai/live/cmd/regionalverify ./ai/live/cmd/regionalpromote -count=1 -timeout=20m` 的 attempt `20261022T150224.972917Z-6t6buhp5` 退出 1，完整保留。该 checkout 缺少既有 Decisions text-choice/jev/runtime ZIP 与 GLM formal 不可变档案，导致对应历史证据测试失败；不将其当成完整普通范围 PASS。最终主工作区完整包验收由工单07负责。

live、race、压力、真实候选/正式资格与实际晋升均 NOT_RUN；本工单未请求厂商。受控 proposal 不产生正式身份，release.pass=false。回滚先停用绑定，再回滚本工单代码与仍 inactive 的独立目录载入；保留全部历史原件。
