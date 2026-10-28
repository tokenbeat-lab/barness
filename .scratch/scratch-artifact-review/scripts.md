# .scratch 脚本导航

审查日期：2026-10-23。共 310 个 Python、8 个 MJS；行数与 SHA-256 见 [inventory.json](/Users/cyber/RestoX/barness/.scratch/scratch-artifact-review/inventory.json)。

本表覆盖全部独立脚本，以路径、docstring 和静态入口分类导航；不是对每个分支的运行认证。活动/历史身份与风险判断以 [报告](/Users/cyber/RestoX/barness/.scratch/scratch-artifact-review/report.md) 为准。`verify` 可能调用 Go 子进程，`prepare/check` 可能构建或执行测试；名称不代表无副作用。


## barness-ai-glm

| 脚本 | 行数 | 用途导航 |
| --- | ---: | --- |
| [evidence/ticket-01/compare.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-glm/evidence/ticket-01/compare.py:1) | 116 | Compare captured public Client behavior, verifying each bundle's hashes. |
| [evidence/ticket-10/verify-evidence.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-glm/evidence/ticket-10/verify-evidence.py:1) | 96 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-11/launcher.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-glm/evidence/ticket-11/launcher.py:1) | 47 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [evidence/ticket-12/launcher.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-glm/evidence/ticket-12/launcher.py:1) | 47 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [evidence/ticket-13/commands.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-glm/evidence/ticket-13/commands.py:1) | 122 | Selected-secret subprocess isolation and finite command receipts. |
| [evidence/ticket-13/corrective.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-glm/evidence/ticket-13/corrective.py:1) | 166 | One separately reserved 36-call source-corrective round; no prior balance. |
| [evidence/ticket-13/glm_launcher.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-glm/evidence/ticket-13/glm_launcher.py:1) | 69 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [evidence/ticket-13/index.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-glm/evidence/ticket-13/index.py:1) | 151 | Index sealed current release evidence and record owned temporary cleanup. |
| [evidence/ticket-13/preserve-default-history.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-glm/evidence/ticket-13/preserve-default-history.py:1) | 100 | 既有运行结果采集、历史保留或归档生成；会写输出 |
| [evidence/ticket-13/preserve-history.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-glm/evidence/ticket-13/preserve-history.py:1) | 141 | 既有运行结果采集、历史保留或归档生成；会写输出 |
| [evidence/ticket-13/reconcile-rollback-history.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-glm/evidence/ticket-13/reconcile-rollback-history.py:1) | 83 | 历史回滚恢复/回滚证据辅助；须读源码确认输出边界 |
| [evidence/ticket-13/rollback.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-glm/evidence/ticket-13/rollback.py:1) | 133 | 历史回滚恢复/回滚证据辅助；须读源码确认输出边界 |
| [evidence/ticket-13/run.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-glm/evidence/ticket-13/run.py:1) | 211 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [evidence/ticket-13/seal.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-glm/evidence/ticket-13/seal.py:1) | 336 | 来源或证据封存/快照，可能写出哈希及索引 |

## barness-ai-minimax

| 脚本 | 行数 | 用途导航 |
| --- | ---: | --- |
| [evidence/ticket-08/rollback.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-08/rollback.py:1) | 142 | 历史回滚恢复/回滚证据辅助；须读源码确认输出边界 |
| [evidence/ticket-09/international-final-actual/a_core.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-09/international-final-actual/a_core.py:1) | 102 | Pure bounded archive/reference helpers. No process, environment or network use. |
| [evidence/ticket-09/international-final-actual/curate_a.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-09/international-final-actual/curate_a.py:1) | 206 | DRAFT collector: execute only after explicit parent A-CLOSED notification. |
| [evidence/ticket-09/international-final-actual/retention_core.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-09/international-final-actual/retention_core.py:1) | 293 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-09/international-final-actual/verify_a.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-09/international-final-actual/verify_a.py:1) | 158 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-10/international-complete-candidates/verify_capsule.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-10/international-complete-candidates/verify_capsule.py:1) | 81 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-10/international-formal-glm/verify_capsule.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-10/international-formal-glm/verify_capsule.py:1) | 133 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-10/international-legacy-final/curate_paid.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-10/international-legacy-final/curate_paid.py:1) | 217 | Retain closed L paid evidence under /tmp. Never execute a captured command. |
| [evidence/ticket-10/international-legacy-final/retention_core.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-10/international-legacy-final/retention_core.py:1) | 293 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-10/international-legacy-final/verify_paid.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-10/international-legacy-final/verify_paid.py:1) | 216 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-10/international-postactivation/curate.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-10/international-postactivation/curate.py:1) | 124 | 固定证据处理辅助；输入、输出和执行影响需结合调用方阅读 |
| [evidence/ticket-12/actual-corrective-1/verify_actual.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-12/actual-corrective-1/verify_actual.py:1) | 203 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-12/curation-tool.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-12/curation-tool.py:1) | 124 | 固定证据处理辅助；输入、输出和执行影响需结合调用方阅读 |
| [evidence/ticket-12/original-preflight-source-proof/verify_q.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-12/original-preflight-source-proof/verify_q.py:1) | 190 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-12/retention-bounds/test_portable.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-12/retention-bounds/test_portable.py:1) | 169 | 工具的测试、合成子进程或测试支持代码 |
| [evidence/ticket-13/actual-glm/check-boundaries.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-13/actual-glm/check-boundaries.py:1) | 63 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-13/actual-glm/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-13/actual-glm/verify.py:1) | 157 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-13/actual-legacy/check-boundaries.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-13/actual-legacy/check-boundaries.py:1) | 68 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-13/actual-legacy/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-13/actual-legacy/verify.py:1) | 190 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-13/final-gate/capsule.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-13/final-gate/capsule.py:1) | 79 | Finite streaming reader for each fixed complete ordinary/race suite. |
| [evidence/ticket-13/final-gate/check-boundaries.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-13/final-gate/check-boundaries.py:1) | 88 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-13/final-gate/snapshot-ordinary.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-13/final-gate/snapshot-ordinary.py:1) | 56 | 来源或证据封存/快照，可能写出哈希及索引 |
| [evidence/ticket-13/final-gate/snapshot-race.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-13/final-gate/snapshot-race.py:1) | 60 | 来源或证据封存/快照，可能写出哈希及索引 |
| [evidence/ticket-13/final-gate/snapshot-verdict.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-13/final-gate/snapshot-verdict.py:1) | 44 | 来源或证据封存/快照，可能写出哈希及索引 |
| [evidence/ticket-13/final-gate/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-13/final-gate/verify.py:1) | 127 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-13/implementation/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-13/implementation/verify.py:1) | 175 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-13/keyless-launch-history/capsule.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-13/keyless-launch-history/capsule.py:1) | 78 | Finite streaming reader for this large interrupted launch history only. |
| [evidence/ticket-13/keyless-launch-history/check-boundaries.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-13/keyless-launch-history/check-boundaries.py:1) | 69 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-13/keyless-launch-history/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-13/keyless-launch-history/verify.py:1) | 141 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-14/gap2/capsule.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-14/gap2/capsule.py:1) | 78 | Finite streaming reader for the ticket14 keyless repair history only. |
| [evidence/ticket-14/gap2/check-boundaries.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-14/gap2/check-boundaries.py:1) | 95 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-14/gap2/snapshot.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-14/gap2/snapshot.py:1) | 82 | 来源或证据封存/快照，可能写出哈希及索引 |
| [evidence/ticket-14/gap2/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-14/gap2/verify.py:1) | 105 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-14/repair/capsule.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-14/repair/capsule.py:1) | 78 | Finite streaming reader for the ticket14 keyless repair history only. |
| [evidence/ticket-14/repair/check-boundaries.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-14/repair/check-boundaries.py:1) | 95 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-14/repair/snapshot.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-14/repair/snapshot.py:1) | 94 | 来源或证据封存/快照，可能写出哈希及索引 |
| [evidence/ticket-14/repair/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-14/repair/verify.py:1) | 188 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-15/implementation/negative_guards.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/implementation/negative_guards.py:1) | 59 | Public verifier guard checks in private copies; immutable originals stay read-only. |
| [evidence/ticket-15/implementation/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/implementation/verify.py:1) | 72 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-15/new-actual/assembly/check-boundaries.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/assembly/check-boundaries.py:1) | 59 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-15/new-actual/assembly/freeze.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/assembly/freeze.py:1) | 54 | 来源或证据封存/快照，可能写出哈希及索引 |
| [evidence/ticket-15/new-actual/assembly/readers.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/assembly/readers.py:1) | 16 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-15/new-actual/assembly/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/assembly/verify.py:1) | 70 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-15/new-actual/effects/check-boundaries.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects/check-boundaries.py:1) | 65 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-15/new-actual/effects/freeze.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects/freeze.py:1) | 100 | 来源或证据封存/快照，可能写出哈希及索引 |
| [evidence/ticket-15/new-actual/effects/provenance.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects/provenance.py:1) | 28 | Independent custody checks for one already completed, explicitly approved effects phase. |
| [evidence/ticket-15/new-actual/effects/quality.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects/quality.py:1) | 50 | Independent finite recomputation of the fixed twelve synthetic visual samples. |
| [evidence/ticket-15/new-actual/effects/readers.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects/readers.py:1) | 16 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-15/new-actual/effects/tdd/analyse-controller.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects/tdd/analyse-controller.py:1) | 6 | 固定证据处理辅助；输入、输出和执行影响需结合调用方阅读 |
| [evidence/ticket-15/new-actual/effects/tdd/red-original-regular-scan.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects/tdd/red-original-regular-scan.py:1) | 15 | 有界诊断、负例或复现装置；可能启动负载和子进程 |
| [evidence/ticket-15/new-actual/effects/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects/verify.py:1) | 131 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-15/new-actual/final-gate/ordinary/boolean-receipt-probe.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/boolean-receipt-probe.py:1) | 62 | 有界诊断、负例或复现装置；可能启动负载和子进程 |
| [evidence/ticket-15/new-actual/final-gate/ordinary/check-boundaries.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/check-boundaries.py:1) | 46 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-15/new-actual/final-gate/ordinary/check-probe-bounds.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/check-probe-bounds.py:1) | 26 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-15/new-actual/final-gate/ordinary/field-repair/verify-negative-fixtures.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/field-repair/verify-negative-fixtures.py:1) | 15 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-15/new-actual/final-gate/ordinary/freeze.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/freeze.py:1) | 56 | 来源或证据封存/快照，可能写出哈希及索引 |
| [evidence/ticket-15/new-actual/final-gate/ordinary/readers.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/readers.py:1) | 16 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-15/new-actual/final-gate/ordinary/tdd/verify-sealed-bytes.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/tdd/verify-sealed-bytes.py:1) | 4 | 来源或证据封存/快照，可能写出哈希及索引 |
| [evidence/ticket-15/new-actual/final-gate/ordinary/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/verify.py:1) | 42 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-15/new-actual/final-gate/race/check-boundaries.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/race/check-boundaries.py:1) | 38 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-15/new-actual/final-gate/race/freeze.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/race/freeze.py:1) | 72 | 来源或证据封存/快照，可能写出哈希及索引 |
| [evidence/ticket-15/new-actual/final-gate/race/readers.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/race/readers.py:1) | 16 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-15/new-actual/final-gate/race/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/race/verify.py:1) | 42 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-15/new-actual/final-gate/verdict/check-boundaries.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/verdict/check-boundaries.py:1) | 41 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-15/new-actual/final-gate/verdict/freeze.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/verdict/freeze.py:1) | 66 | 来源或证据封存/快照，可能写出哈希及索引 |
| [evidence/ticket-15/new-actual/final-gate/verdict/readers.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/verdict/readers.py:1) | 16 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-15/new-actual/final-gate/verdict/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/verdict/verify.py:1) | 77 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-15/new-actual/glm/check-boundaries.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/glm/check-boundaries.py:1) | 83 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-15/new-actual/glm/complete-snapshot.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/glm/complete-snapshot.py:1) | 38 | 来源或证据封存/快照，可能写出哈希及索引 |
| [evidence/ticket-15/new-actual/glm/freeze.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/glm/freeze.py:1) | 110 | 来源或证据封存/快照，可能写出哈希及索引 |
| [evidence/ticket-15/new-actual/glm/provenance.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/glm/provenance.py:1) | 43 | Independent custody checks for one already completed, explicitly approved G phase. |
| [evidence/ticket-15/new-actual/glm/readers.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/glm/readers.py:1) | 16 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-15/new-actual/glm/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/glm/verify.py:1) | 101 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-16/deadline-alignment/check-boundaries.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-16/deadline-alignment/check-boundaries.py:1) | 63 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-16/deadline-alignment/snapshot.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-16/deadline-alignment/snapshot.py:1) | 41 | 来源或证据封存/快照，可能写出哈希及索引 |
| [evidence/ticket-16/deadline-alignment/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-16/deadline-alignment/verify.py:1) | 81 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-17/check-boundaries.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-17/check-boundaries.py:1) | 42 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-17/contract.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-17/contract.py:1) | 30 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-17/expand-trace.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-17/expand-trace.py:1) | 38 | One reviewed additive finite trace revision; never change prior IDs/cases. |
| [evidence/ticket-17/freeze.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-17/freeze.py:1) | 82 | 来源或证据封存/快照，可能写出哈希及索引 |
| [evidence/ticket-17/readers.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-17/readers.py:1) | 14 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-17/retention-fixes/creator-probe.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/creator-probe.py:1) | 93 | 有界诊断、负例或复现装置；可能启动负载和子进程 |
| [evidence/ticket-17/retention-fixes/factory-correction/check-factory.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/factory-correction/check-factory.py:1) | 68 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-17/retention-fixes/factory-correction/raw/controllers/check-factory.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/factory-correction/raw/controllers/check-factory.py:1) | 68 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-17/retention-fixes/factory-correction/raw/controllers/run-legitimate.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/factory-correction/raw/controllers/run-legitimate.py:1) | 11 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [evidence/ticket-17/retention-fixes/factory-correction/raw-v2/controllers/check-factory.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/factory-correction/raw-v2/controllers/check-factory.py:1) | 68 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-17/retention-fixes/factory-correction/raw-v2/controllers/run-legitimate.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/factory-correction/raw-v2/controllers/run-legitimate.py:1) | 11 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [evidence/ticket-17/retention-fixes/factory-correction/verify-factory.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/factory-correction/verify-factory.py:1) | 27 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-17/retention-fixes/positive-creator.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/positive-creator.py:1) | 13 | Reuse one isolated failed-before-write fixture for a normal bounded CLI success. |
| [evidence/ticket-17/retention-fixes/resource-probe.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/resource-probe.py:1) | 77 | 有界诊断、负例或复现装置；可能启动负载和子进程 |
| [evidence/ticket-17/retention-fixes/verify-corrections.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/verify-corrections.py:1) | 42 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-17/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-17/verify.py:1) | 89 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-19/native-home-redaction/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-19/native-home-redaction/verify.py:1) | 73 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-19/native-renewal/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-19/native-renewal/verify.py:1) | 71 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-20/jev_fixtures.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-20/jev_fixtures.py:1) | 70 | Create two explicitly derived consumer fixtures; never change native inputs. |
| [evidence/ticket-20/run.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-20/run.py:1) | 31 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [evidence/ticket-20/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-20/verify.py:1) | 73 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-21/actual-retained/public-guards.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-21/actual-retained/public-guards.py:1) | 74 | Public CLI custody negatives; private copies only, no native/captured execution. |
| [evidence/ticket-21/actual-retained/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-21/actual-retained/verify.py:1) | 192 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-21/final-gate/supplement/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-21/final-gate/supplement/verify.py:1) | 135 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [evidence/ticket-21/final-gate/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-21/final-gate/verify.py:1) | 195 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [session-analysis/build_report.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/session-analysis/build_report.py:1) | 157 | 读取既有证据，生成失败与耗时复盘材料 |
| [session-analysis/collect.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/session-analysis/collect.py:1) | 77 | 读取既有证据，生成失败与耗时复盘材料 |
| [tools/eleven_contract.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/eleven_contract.py:1) | 42 | MiniMax 固定付费轮次、续期、发布修复或其共享合同；须查对应 README |
| [tools/eleven_finalize.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/eleven_finalize.py:1) | 63 | MiniMax 固定付费轮次、续期、发布修复或其共享合同；须查对应 README |
| [tools/eleven_grant.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/eleven_grant.py:1) | 47 | MiniMax 固定付费轮次、续期、发布修复或其共享合同；须查对应 README |
| [tools/eleven_inputs.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/eleven_inputs.py:1) | 79 | MiniMax 固定付费轮次、续期、发布修复或其共享合同；须查对应 README |
| [tools/eleven_validation.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/eleven_validation.py:1) | 155 | MiniMax 固定付费轮次、续期、发布修复或其共享合同；须查对应 README |
| [tools/glm_corrective.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/glm_corrective.py:1) | 201 | MiniMax 固定付费轮次、续期、发布修复或其共享合同；须查对应 README |
| [tools/paid.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/paid.py:1) | 299 | MiniMax 固定付费轮次、续期、发布修复或其共享合同；须查对应 README |
| [tools/paid_capture.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/paid_capture.py:1) | 64 | MiniMax 固定付费轮次、续期、发布修复或其共享合同；须查对应 README |
| [tools/paid_checks.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/paid_checks.py:1) | 36 | MiniMax 固定付费轮次、续期、发布修复或其共享合同；须查对应 README |
| [tools/paid_core.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/paid_core.py:1) | 183 | MiniMax 固定付费轮次、续期、发布修复或其共享合同；须查对应 README |
| [tools/paid_history.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/paid_history.py:1) | 38 | MiniMax 固定付费轮次、续期、发布修复或其共享合同；须查对应 README |
| [tools/paid_plan.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/paid_plan.py:1) | 147 | MiniMax 固定付费轮次、续期、发布修复或其共享合同；须查对应 README |
| [tools/release_repair_validation.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/release_repair_validation.py:1) | 120 | MiniMax 固定付费轮次、续期、发布修复或其共享合同；须查对应 README |
| [tools/renewal.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/renewal.py:1) | 146 | MiniMax 固定付费轮次、续期、发布修复或其共享合同；须查对应 README |
| [tools/renewal_common.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/renewal_common.py:1) | 140 | MiniMax 固定付费轮次、续期、发布修复或其共享合同；须查对应 README |
| [tools/renewal_glm.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/renewal_glm.py:1) | 170 | MiniMax 固定付费轮次、续期、发布修复或其共享合同；须查对应 README |
| [tools/repair_anchors.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/repair_anchors.py:1) | 43 | MiniMax 固定付费轮次、续期、发布修复或其共享合同；须查对应 README |
| [tools/repair_contract.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/repair_contract.py:1) | 41 | MiniMax 固定付费轮次、续期、发布修复或其共享合同；须查对应 README |
| [tools/repair_inputs.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/repair_inputs.py:1) | 141 | MiniMax 固定付费轮次、续期、发布修复或其共享合同；须查对应 README |
| [tools/repair_publication.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/repair_publication.py:1) | 61 | MiniMax 固定付费轮次、续期、发布修复或其共享合同；须查对应 README |
| [tools/routed_contract.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/routed_contract.py:1) | 59 | MiniMax 固定付费轮次、续期、发布修复或其共享合同；须查对应 README |
| [tools/routed_evidence.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/routed_evidence.py:1) | 163 | MiniMax 固定付费轮次、续期、发布修复或其共享合同；须查对应 README |
| [tools/routed_validation.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/routed_validation.py:1) | 178 | MiniMax 固定付费轮次、续期、发布修复或其共享合同；须查对应 README |
| [tools/test_eleven_validation.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/test_eleven_validation.py:1) | 156 | 工具的测试、合成子进程或测试支持代码 |
| [tools/test_glm_corrective.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/test_glm_corrective.py:1) | 142 | 工具的测试、合成子进程或测试支持代码 |
| [tools/test_paid.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/test_paid.py:1) | 425 | 工具的测试、合成子进程或测试支持代码 |
| [tools/test_release_repair.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/test_release_repair.py:1) | 135 | 工具的测试、合成子进程或测试支持代码 |
| [tools/test_renewal.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/test_renewal.py:1) | 390 | 工具的测试、合成子进程或测试支持代码 |
| [tools/test_repair_fixtures.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/test_repair_fixtures.py:1) | 89 | 工具的测试、合成子进程或测试支持代码 |
| [tools/test_routed_validation.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/test_routed_validation.py:1) | 269 | 工具的测试、合成子进程或测试支持代码 |

## barness-ai-openai-decisions

| 脚本 | 行数 | 用途导航 |
| --- | ---: | --- |
| [candidate-live-evidence/live-authorized-20261013/launcher-used.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-openai-decisions/candidate-live-evidence/live-authorized-20261013/launcher-used.py:1) | 24 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [chinese-live-evidence/launcher-process.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-openai-decisions/chinese-live-evidence/launcher-process.py:1) | 58 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [chinese-live-evidence/run.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-openai-decisions/chinese-live-evidence/run.py:1) | 132 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [classifier-load-evidence/pack.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-openai-decisions/classifier-load-evidence/pack.py:1) | 111 | 既有运行结果采集、历史保留或归档生成；会写输出 |
| [formal-support-evidence/launcher.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-openai-decisions/formal-support-evidence/launcher.py:1) | 55 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [jev-baseline/replay.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-openai-decisions/jev-baseline/replay.py:1) | 54 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [jev-baseline/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-openai-decisions/jev-baseline/verify.py:1) | 159 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [mixed-answer-evidence/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-openai-decisions/mixed-answer-evidence/verify.py:1) | 136 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [release-evidence/collect.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-openai-decisions/release-evidence/collect.py:1) | 123 | 既有运行结果采集、历史保留或归档生成；会写输出 |
| [release-evidence/pack.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-openai-decisions/release-evidence/pack.py:1) | 188 | 既有运行结果采集、历史保留或归档生成；会写输出 |
| [release-evidence/run-gate.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-openai-decisions/release-evidence/run-gate.py:1) | 34 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [release-evidence/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-openai-decisions/release-evidence/verify.py:1) | 112 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [runtime-evidence/pack.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-openai-decisions/runtime-evidence/pack.py:1) | 119 | 既有运行结果采集、历史保留或归档生成；会写输出 |
| [text-choice-evidence/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-openai-decisions/text-choice-evidence/verify.py:1) | 136 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [ticket-review/originals/ticket-review/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-openai-decisions/ticket-review/originals/ticket-review/verify.py:1) | 130 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [ticket-review/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-openai-decisions/ticket-review/verify.py:1) | 197 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |

## barness-ai-pi-1.0

| 脚本 | 行数 | 用途导航 |
| --- | ---: | --- |
| [chinese-evaluation-evidence/run-live.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/chinese-evaluation-evidence/run-live.py:1) | 60 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [chinese-evaluation-evidence/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/chinese-evaluation-evidence/verify.py:1) | 28 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [google-images-live-evidence/capture.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/google-images-live-evidence/capture.py:1) | 82 | 既有运行结果采集、历史保留或归档生成；会写输出 |
| [google-images-live-evidence/run-live.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/google-images-live-evidence/run-live.py:1) | 31 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [google-images-live-evidence/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/google-images-live-evidence/verify.py:1) | 80 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [google-interactions-evidence/capture.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/google-interactions-evidence/capture.py:1) | 87 | 既有运行结果采集、历史保留或归档生成；会写输出 |
| [google-interactions-evidence/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/google-interactions-evidence/verify.py:1) | 45 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [google-reference-images-evidence/capture.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/google-reference-images-evidence/capture.py:1) | 87 | 既有运行结果采集、历史保留或归档生成；会写输出 |
| [google-reference-images-evidence/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/google-reference-images-evidence/verify.py:1) | 47 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [mixed-operations-evidence/capture.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/mixed-operations-evidence/capture.py:1) | 149 | 既有运行结果采集、历史保留或归档生成；会写输出 |
| [mixed-operations-evidence/run_commands.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/mixed-operations-evidence/run_commands.py:1) | 59 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [mixed-operations-evidence/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/mixed-operations-evidence/verify.py:1) | 79 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [nine-route-release-evidence/capture.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/nine-route-release-evidence/capture.py:1) | 151 | 既有运行结果采集、历史保留或归档生成；会写输出 |
| [nine-route-release-evidence/run-live.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/nine-route-release-evidence/run-live.py:1) | 50 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [nine-route-release-evidence/run-release.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/nine-route-release-evidence/run-release.py:1) | 54 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [nine-route-release-evidence/test-run-release.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/nine-route-release-evidence/test-run-release.py:1) | 69 | 工具的测试、合成子进程或测试支持代码 |
| [nine-route-release-evidence/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/nine-route-release-evidence/verify.py:1) | 116 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [openai-editing-evidence/capture.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/openai-editing-evidence/capture.py:1) | 67 | 既有运行结果采集、历史保留或归档生成；会写输出 |
| [openai-editing-evidence/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/openai-editing-evidence/verify.py:1) | 34 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [openai-generation-evidence/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/openai-generation-evidence/verify.py:1) | 34 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [openai-images-live-evidence/run_live.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/openai-images-live-evidence/run_live.py:1) | 33 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [openai-images-live-evidence/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/openai-images-live-evidence/verify.py:1) | 102 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [pressure-loopback-evidence/reproduce-retention.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/pressure-loopback-evidence/reproduce-retention.py:1) | 64 | 有界诊断、负例或复现装置；可能启动负载和子进程 |
| [pressure-loopback-evidence/run.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/pressure-loopback-evidence/run.py:1) | 70 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [pressure-loopback-evidence/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/pressure-loopback-evidence/verify.py:1) | 92 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [release-verification/check.mjs](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/release-verification/check.mjs:1) | 142 | 历史发布件和源码构建产物比较，写出核验结果 |
| [release-verification/failures.mjs](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/release-verification/failures.mjs:1) | 75 | 历史等价核验的故障场景 |
| [release-verification/faults.mjs](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/release-verification/faults.mjs:1) | 64 | 历史下载超限/证据写失败的进程故障场景 |
| [release-verification/run.mjs](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/release-verification/run.mjs:1) | 159 | 历史 pi 发布件/源码下载、安装与等价核验入口 |
| [typesafe-live-evidence/run-live.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/typesafe-live-evidence/run-live.py:1) | 31 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [typesafe-live-evidence/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/typesafe-live-evidence/verify.py:1) | 68 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [unary-failures-evidence/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-pi-1.0/unary-failures-evidence/verify.py:1) | 42 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |

## barness-ai-release-verification

| 脚本 | 行数 | 用途导航 |
| --- | ---: | --- |
| [native-pressure/reproduce.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-release-verification/native-pressure/reproduce.py:1) | 176 | 有界诊断、负例或复现装置；可能启动负载和子进程 |
| [seal-evidence.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-release-verification/seal-evidence.py:1) | 254 | 来源或证据封存/快照，可能写出哈希及索引 |

## barness-ai-test-process-review

| 脚本 | 行数 | 用途导航 |
| --- | ---: | --- |
| [reviews/20261022-section1-rereview-repros/consumer-final-write.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/reviews/20261022-section1-rereview-repros/consumer-final-write.py:1) | 22 | 固定证据处理辅助；输入、输出和执行影响需结合调用方阅读 |
| [reviews/20261022-section1-rereview-repros/relative-dependency.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/reviews/20261022-section1-rereview-repros/relative-dependency.py:1) | 21 | 固定证据处理辅助；输入、输出和执行影响需结合调用方阅读 |
| [reviews/20261022-section1-rereview-repros/zero-call-chat.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/reviews/20261022-section1-rereview-repros/zero-call-chat.py:1) | 34 | 固定证据处理辅助；输入、输出和执行影响需结合调用方阅读 |
| [tools/cleanup.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/tools/cleanup.py:1) | 195 | 当前 SOP：清理批次命令入口，含移动/恢复/永久删除 |
| [tools/cleanup_scan.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/tools/cleanup_scan.py:1) | 265 | 当前 SOP：引用/库存扫描与保留分类 |
| [tools/cleanup_state.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/tools/cleanup_state.py:1) | 126 | 当前 SOP：清理批次锁、状态和原件检查 |
| [tools/consume_live.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/tools/consume_live.py:1) | 69 | 当前 SOP：采集一次原生 live 资格消费者；不请求厂商 |
| [tools/live_environment.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/tools/live_environment.py:1) | 38 | 当前 SOP：转发单一选定 live 路线的环境 |
| [tools/receipt_files.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/tools/receipt_files.py:1) | 134 | 当前 SOP：有界文件、原子保存、来源和原生包核验 |
| [tools/run.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/tools/run.py:1) | 146 | 当前 SOP：采集一条已选命令的回执 |
| [tools/run_process.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/tools/run_process.py:1) | 142 | 当前 SOP：隔离环境、有界采集、取消和脱敏 |
| [tools/tests/fixture_child.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/tools/tests/fixture_child.py:1) | 62 | 工具的测试、合成子进程或测试支持代码 |
| [tools/tests/live_fixture.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/tools/tests/live_fixture.py:1) | 117 | 工具的测试、合成子进程或测试支持代码 |
| [tools/tests/process_support.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/tools/tests/process_support.py:1) | 70 | 工具的测试、合成子进程或测试支持代码 |
| [tools/tests/test_cleanup.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/tools/tests/test_cleanup.py:1) | 187 | 工具的测试、合成子进程或测试支持代码 |
| [tools/tests/test_live_release.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/tools/tests/test_live_release.py:1) | 239 | 工具的测试、合成子进程或测试支持代码 |
| [tools/tests/test_ordinary.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/tools/tests/test_ordinary.py:1) | 77 | 工具的测试、合成子进程或测试支持代码 |
| [tools/tests/test_release.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/tools/tests/test_release.py:1) | 174 | 工具的测试、合成子进程或测试支持代码 |
| [tools/tests/test_run.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/tools/tests/test_run.py:1) | 192 | 工具的测试、合成子进程或测试支持代码 |

## barness-ai-thinking-sampling

| 脚本 | 行数 | 用途导航 |
| --- | ---: | --- |
| [fixture-cases.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-thinking-sampling/fixture-cases.py:1) | 130 | 历史生成器：直接改写 ai/e2e/testdata 下四路线夹具 |
| [pressure-stream-integrity-evidence/native-causality/reproduce.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-thinking-sampling/pressure-stream-integrity-evidence/native-causality/reproduce.py:1) | 173 | 有界诊断、负例或复现装置；可能启动负载和子进程 |
| [pressure-stream-integrity-evidence/native-causality/tool-red.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-thinking-sampling/pressure-stream-integrity-evidence/native-causality/tool-red.py:1) | 66 | Replay the removed retry through a temporary Go overlay; never edit runtime source. |
| [pressure-stream-integrity-evidence/replay-red.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-thinking-sampling/pressure-stream-integrity-evidence/replay-red.py:1) | 66 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [pressure-stream-integrity-evidence/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-thinking-sampling/pressure-stream-integrity-evidence/verify.py:1) | 79 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [release-evidence/pack.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-thinking-sampling/release-evidence/pack.py:1) | 153 | 既有运行结果采集、历史保留或归档生成；会写输出 |
| [release-evidence/run-evaluation.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-thinking-sampling/release-evidence/run-evaluation.py:1) | 62 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [release-evidence/run-live.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-thinking-sampling/release-evidence/run-live.py:1) | 54 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [release-evidence/run-release.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-thinking-sampling/release-evidence/run-release.py:1) | 39 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [release-evidence/run-smokes.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-thinking-sampling/release-evidence/run-smokes.py:1) | 30 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [release-evidence/verify.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-thinking-sampling/release-evidence/verify.py:1) | 307 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [release-verification/check.mjs](/Users/cyber/RestoX/barness/.scratch/barness-ai-thinking-sampling/release-verification/check.mjs:1) | 136 | 历史发布件和源码构建产物比较，写出核验结果 |
| [release-verification/failures.mjs](/Users/cyber/RestoX/barness/.scratch/barness-ai-thinking-sampling/release-verification/failures.mjs:1) | 75 | 历史等价核验的故障场景 |
| [release-verification/faults.mjs](/Users/cyber/RestoX/barness/.scratch/barness-ai-thinking-sampling/release-verification/faults.mjs:1) | 64 | 历史下载超限/证据写失败的进程故障场景 |
| [release-verification/run.mjs](/Users/cyber/RestoX/barness/.scratch/barness-ai-thinking-sampling/release-verification/run.mjs:1) | 165 | 历史 pi 发布件/源码下载、安装与等价核验入口 |
| [update-ledger.py](/Users/cyber/RestoX/barness/.scratch/barness-ai-thinking-sampling/update-ledger.py:1) | 29 | 退役差分账本改写器；当前 --help 已复现失败 |

## barness-provider-sites

| 脚本 | 行数 | 用途导航 |
| --- | ---: | --- |
| [tools/final_contract.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/final_contract.py:1) | 64 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/final_execute.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/final_execute.py:1) | 235 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/final_finalize.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/final_finalize.py:1) | 101 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/final_grants.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/final_grants.py:1) | 62 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/final_minimax.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/final_minimax.py:1) | 59 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/final_native.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/final_native.py:1) | 159 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/final_predecessor.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/final_predecessor.py:1) | 186 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/final_process.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/final_process.py:1) | 82 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/final_publication.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/final_publication.py:1) | 65 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/final_retry.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/final_retry.py:1) | 153 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/final_retry_execute.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/final_retry_execute.py:1) | 61 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/final_round.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/final_round.py:1) | 74 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/final_state.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/final_state.py:1) | 212 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery4_contract.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery4_contract.py:1) | 53 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery4_history.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery4_history.py:1) | 134 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery4_native.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery4_native.py:1) | 72 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery4_round.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery4_round.py:1) | 39 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery5_contract.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery5_contract.py:1) | 55 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery5_history.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery5_history.py:1) | 96 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery5_native.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery5_native.py:1) | 78 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery5_retained.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery5_retained.py:1) | 86 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery5_round.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery5_round.py:1) | 39 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery6_contract.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery6_contract.py:1) | 57 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery6_files.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery6_files.py:1) | 44 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery6_history.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery6_history.py:1) | 116 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery6_native.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery6_native.py:1) | 78 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery6_retained.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery6_retained.py:1) | 97 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery6_round.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery6_round.py:1) | 39 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery7_contract.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery7_contract.py:1) | 63 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery7_files.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery7_files.py:1) | 44 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery7_history.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery7_history.py:1) | 127 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery7_native.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery7_native.py:1) | 82 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery7_retained.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery7_retained.py:1) | 97 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery7_round.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery7_round.py:1) | 33 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery_contract.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery_contract.py:1) | 53 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery_flow.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery_flow.py:1) | 35 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery_flow_execute.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery_flow_execute.py:1) | 165 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery_flow_files.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery_flow_files.py:1) | 55 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery_flow_finalize.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery_flow_finalize.py:1) | 104 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery_flow_native.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery_flow_native.py:1) | 120 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery_flow_publication.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery_flow_publication.py:1) | 86 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery_flow_source.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery_flow_source.py:1) | 75 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery_flow_state.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery_flow_state.py:1) | 141 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery_history.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery_history.py:1) | 84 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery_native.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery_native.py:1) | 48 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/recovery_round.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/recovery_round.py:1) | 40 | 地区验收固定轮次控制器/共享实现；当前恢复7暂停，须查 spec |
| [tools/test_closed_fixture.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/test_closed_fixture.py:1) | 49 | 工具的测试、合成子进程或测试支持代码 |
| [tools/test_final_publication.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/test_final_publication.py:1) | 117 | 工具的测试、合成子进程或测试支持代码 |
| [tools/test_final_retry.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/test_final_retry.py:1) | 77 | 工具的测试、合成子进程或测试支持代码 |
| [tools/test_final_round.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/test_final_round.py:1) | 400 | 工具的测试、合成子进程或测试支持代码 |
| [tools/test_recovery4_fixture.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/test_recovery4_fixture.py:1) | 91 | 工具的测试、合成子进程或测试支持代码 |
| [tools/test_recovery4_round.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/test_recovery4_round.py:1) | 256 | 工具的测试、合成子进程或测试支持代码 |
| [tools/test_recovery5_fixture.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/test_recovery5_fixture.py:1) | 100 | 工具的测试、合成子进程或测试支持代码 |
| [tools/test_recovery5_intent_io.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/test_recovery5_intent_io.py:1) | 69 | 工具的测试、合成子进程或测试支持代码 |
| [tools/test_recovery5_io.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/test_recovery5_io.py:1) | 56 | 工具的测试、合成子进程或测试支持代码 |
| [tools/test_recovery5_round.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/test_recovery5_round.py:1) | 170 | 工具的测试、合成子进程或测试支持代码 |
| [tools/test_recovery6_fixture.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/test_recovery6_fixture.py:1) | 88 | 工具的测试、合成子进程或测试支持代码 |
| [tools/test_recovery6_round.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/test_recovery6_round.py:1) | 196 | 工具的测试、合成子进程或测试支持代码 |
| [tools/test_recovery7_fixture.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/test_recovery7_fixture.py:1) | 81 | 工具的测试、合成子进程或测试支持代码 |
| [tools/test_recovery7_round.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/test_recovery7_round.py:1) | 227 | 工具的测试、合成子进程或测试支持代码 |
| [tools/test_recovery_fixture.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/test_recovery_fixture.py:1) | 100 | 工具的测试、合成子进程或测试支持代码 |
| [tools/test_recovery_review_fix.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/test_recovery_review_fix.py:1) | 104 | 工具的测试、合成子进程或测试支持代码 |
| [tools/test_recovery_round.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/test_recovery_round.py:1) | 189 | 工具的测试、合成子进程或测试支持代码 |
| [tools/test_retry_fixture.py](/Users/cyber/RestoX/barness/.scratch/barness-provider-sites/tools/test_retry_fixture.py:1) | 79 | 工具的测试、合成子进程或测试支持代码 |

## jev-1.14-multimodal-box

| 脚本 | 行数 | 用途导航 |
| --- | ---: | --- |
| [14-remaining-real/recheck.py](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/14-remaining-real/recheck.py:1) | 69 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [14-remaining-real/recover.py](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/14-remaining-real/recover.py:1) | 247 | The one reserved post-fix full round; retain every initial verdict. |
| [14-remaining-real/run.py](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/14-remaining-real/run.py:1) | 300 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [14-remaining-real/seal.py](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/14-remaining-real/seal.py:1) | 337 | 来源或证据封存/快照，可能写出哈希及索引 |
| [14-typesafe-real/run.py](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/14-typesafe-real/run.py:1) | 249 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [14-typesafe-real/seal.py](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/14-typesafe-real/seal.py:1) | 404 | 来源或证据封存/快照，可能写出哈希及索引 |
| [native-snapshot-evidence/replay.py](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/native-snapshot-evidence/replay.py:1) | 127 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [run-14-verification.py](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/run-14-verification.py:1) | 137 | 执行/回放/采集入口；先查本目录合同及是否 live/压力 |
| [seal-11-evidence.py](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/seal-11-evidence.py:1) | 99 | 来源或证据封存/快照，可能写出哈希及索引 |
| [seal-12-evidence.py](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/seal-12-evidence.py:1) | 123 | 来源或证据封存/快照，可能写出哈希及索引 |
| [seal-13-evidence.py](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/seal-13-evidence.py:1) | 175 | 来源或证据封存/快照，可能写出哈希及索引 |
| [seal-14-evidence.py](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/seal-14-evidence.py:1) | 223 | 来源或证据封存/快照，可能写出哈希及索引 |
| [verify-02-evidence.py](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/verify-02-evidence.py:1) | 55 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [verify-03-evidence.py](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/verify-03-evidence.py:1) | 81 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [verify-04-evidence.py](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/verify-04-evidence.py:1) | 134 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [verify-05-evidence.py](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/verify-05-evidence.py:1) | 153 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [verify-06-evidence.py](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/verify-06-evidence.py:1) | 164 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [verify-07-evidence.py](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/verify-07-evidence.py:1) | 185 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [verify-08-evidence.py](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/verify-08-evidence.py:1) | 178 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [verify-09-evidence.py](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/verify-09-evidence.py:1) | 234 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [verify-10-evidence.py](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/verify-10-evidence.py:1) | 157 | 证据读取、完整性/语义核验或其固定合同；可能启动本地子进程 |
| [verify-11-evidence.py](/Users/cyber/RestoX/barness/.scratch/jev-1.14-multimodal-box/verify-11-evidence.py:1) | 364 | 仍被 Go 原生消费者调用的 Jev 协议/预算/完整性核验器 |

## 另存的压缩 Python 源码快照

这两份是历史来源字节，不计入上方 318 个独立脚本入口；仅做有界解压与 AST 解析，未执行。

| 快照 | 解压后行数 | 用途 |
| --- | ---: | --- |
| [retain_source.py.gz](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-12/original-preflight-source-proof/retain_source.py.gz) | 142 | Add committed source bytes to an externally retained failed keyless window. |
| [additional-protocol-source.py.gz](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects/additional-protocol-source.py.gz) | 364 | Independently read candidate or explicitly selected formal protocol structure. |
