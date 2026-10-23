# barness-ai-minimax 脚本边界

[分类入口](../script-boundaries.md) · [逐文件机器清单](../script-boundaries.json)

本页是静态审查说明，不是执行脚本清单。路径/行号按审查Git 99d0749 的原字节解析；L号是源码行，压缩源码的L号指解压后的文本。CLI、模块、测试支持与顶层程序分开看。全部副作用描述为代码能力，本次均未运行。

| 文件 | 身份 | 使用边界 |
| --- | --- | --- |
| [evidence/ticket-08/rollback.py](#barness-ai-minimax--evidence--ticket-08--rollback-py) | 活动工具 | 现有公共消费者 |
| [evidence/ticket-09/international-final-actual/a_core.py](#barness-ai-minimax--evidence--ticket-09--international-final-actual--a_core-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-09/international-final-actual/curate_a.py](#barness-ai-minimax--evidence--ticket-09--international-final-actual--curate_a-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-09/international-final-actual/retention_core.py](#barness-ai-minimax--evidence--ticket-09--international-final-actual--retention_core-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-09/international-final-actual/verify_a.py](#barness-ai-minimax--evidence--ticket-09--international-final-actual--verify_a-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-10/international-complete-candidates/verify_capsule.py](#barness-ai-minimax--evidence--ticket-10--international-complete-candidates--verify_capsule-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-10/international-formal-glm/verify_capsule.py](#barness-ai-minimax--evidence--ticket-10--international-formal-glm--verify_capsule-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-10/international-legacy-final/curate_paid.py](#barness-ai-minimax--evidence--ticket-10--international-legacy-final--curate_paid-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-10/international-legacy-final/retention_core.py](#barness-ai-minimax--evidence--ticket-10--international-legacy-final--retention_core-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-10/international-legacy-final/verify_paid.py](#barness-ai-minimax--evidence--ticket-10--international-legacy-final--verify_paid-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-10/international-postactivation/curate.py](#barness-ai-minimax--evidence--ticket-10--international-postactivation--curate-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-12/actual-corrective-1/verify_actual.py](#barness-ai-minimax--evidence--ticket-12--actual-corrective-1--verify_actual-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-12/curation-tool.py](#barness-ai-minimax--evidence--ticket-12--curation-tool-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-12/original-preflight-source-proof/retain_source.py.gz](#barness-ai-minimax--evidence--ticket-12--original-preflight-source-proof--retain_source-py-gz) | 历史源码 | 封存源码快照 |
| [evidence/ticket-12/original-preflight-source-proof/verify_q.py](#barness-ai-minimax--evidence--ticket-12--original-preflight-source-proof--verify_q-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-12/retention-bounds/test_portable.py](#barness-ai-minimax--evidence--ticket-12--retention-bounds--test_portable-py) | 测试 | 仅原轮次 |
| [evidence/ticket-13/actual-glm/check-boundaries.py](#barness-ai-minimax--evidence--ticket-13--actual-glm--check-boundaries-py) | 测试 | 仅原轮次 |
| [evidence/ticket-13/actual-glm/verify.py](#barness-ai-minimax--evidence--ticket-13--actual-glm--verify-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-13/actual-legacy/check-boundaries.py](#barness-ai-minimax--evidence--ticket-13--actual-legacy--check-boundaries-py) | 测试 | 仅原轮次 |
| [evidence/ticket-13/actual-legacy/verify.py](#barness-ai-minimax--evidence--ticket-13--actual-legacy--verify-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-13/final-gate/capsule.py](#barness-ai-minimax--evidence--ticket-13--final-gate--capsule-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-13/final-gate/check-boundaries.py](#barness-ai-minimax--evidence--ticket-13--final-gate--check-boundaries-py) | 测试 | 仅原轮次 |
| [evidence/ticket-13/final-gate/snapshot-ordinary.py](#barness-ai-minimax--evidence--ticket-13--final-gate--snapshot-ordinary-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-13/final-gate/snapshot-race.py](#barness-ai-minimax--evidence--ticket-13--final-gate--snapshot-race-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-13/final-gate/snapshot-verdict.py](#barness-ai-minimax--evidence--ticket-13--final-gate--snapshot-verdict-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-13/final-gate/verify.py](#barness-ai-minimax--evidence--ticket-13--final-gate--verify-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-13/implementation/verify.py](#barness-ai-minimax--evidence--ticket-13--implementation--verify-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-13/keyless-launch-history/capsule.py](#barness-ai-minimax--evidence--ticket-13--keyless-launch-history--capsule-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-13/keyless-launch-history/check-boundaries.py](#barness-ai-minimax--evidence--ticket-13--keyless-launch-history--check-boundaries-py) | 测试 | 仅原轮次 |
| [evidence/ticket-13/keyless-launch-history/verify.py](#barness-ai-minimax--evidence--ticket-13--keyless-launch-history--verify-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-14/gap2/capsule.py](#barness-ai-minimax--evidence--ticket-14--gap2--capsule-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-14/gap2/check-boundaries.py](#barness-ai-minimax--evidence--ticket-14--gap2--check-boundaries-py) | 测试 | 仅原轮次 |
| [evidence/ticket-14/gap2/snapshot.py](#barness-ai-minimax--evidence--ticket-14--gap2--snapshot-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-14/gap2/verify.py](#barness-ai-minimax--evidence--ticket-14--gap2--verify-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-14/repair/capsule.py](#barness-ai-minimax--evidence--ticket-14--repair--capsule-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-14/repair/check-boundaries.py](#barness-ai-minimax--evidence--ticket-14--repair--check-boundaries-py) | 测试 | 仅原轮次 |
| [evidence/ticket-14/repair/snapshot.py](#barness-ai-minimax--evidence--ticket-14--repair--snapshot-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-14/repair/verify.py](#barness-ai-minimax--evidence--ticket-14--repair--verify-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-15/implementation/negative_guards.py](#barness-ai-minimax--evidence--ticket-15--implementation--negative_guards-py) | 测试 | 仅原轮次 |
| [evidence/ticket-15/implementation/verify.py](#barness-ai-minimax--evidence--ticket-15--implementation--verify-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-15/new-actual/assembly/check-boundaries.py](#barness-ai-minimax--evidence--ticket-15--new-actual--assembly--check-boundaries-py) | 测试 | 仅原轮次 |
| [evidence/ticket-15/new-actual/assembly/freeze.py](#barness-ai-minimax--evidence--ticket-15--new-actual--assembly--freeze-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-15/new-actual/assembly/readers.py](#barness-ai-minimax--evidence--ticket-15--new-actual--assembly--readers-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-15/new-actual/assembly/verify.py](#barness-ai-minimax--evidence--ticket-15--new-actual--assembly--verify-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-15/new-actual/effects/additional-protocol-source.py.gz](#barness-ai-minimax--evidence--ticket-15--new-actual--effects--additional-protocol-source-py-gz) | 历史源码 | 封存源码快照 |
| [evidence/ticket-15/new-actual/effects/check-boundaries.py](#barness-ai-minimax--evidence--ticket-15--new-actual--effects--check-boundaries-py) | 测试 | 仅原轮次 |
| [evidence/ticket-15/new-actual/effects/freeze.py](#barness-ai-minimax--evidence--ticket-15--new-actual--effects--freeze-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-15/new-actual/effects/provenance.py](#barness-ai-minimax--evidence--ticket-15--new-actual--effects--provenance-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-15/new-actual/effects/quality.py](#barness-ai-minimax--evidence--ticket-15--new-actual--effects--quality-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-15/new-actual/effects/readers.py](#barness-ai-minimax--evidence--ticket-15--new-actual--effects--readers-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-15/new-actual/effects/tdd/analyse-controller.py](#barness-ai-minimax--evidence--ticket-15--new-actual--effects--tdd--analyse-controller-py) | 诊断装置 | 仅原轮次 |
| [evidence/ticket-15/new-actual/effects/tdd/red-original-regular-scan.py](#barness-ai-minimax--evidence--ticket-15--new-actual--effects--tdd--red-original-regular-scan-py) | 诊断装置 | 仅原轮次 |
| [evidence/ticket-15/new-actual/effects/verify.py](#barness-ai-minimax--evidence--ticket-15--new-actual--effects--verify-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-15/new-actual/final-gate/ordinary/boolean-receipt-probe.py](#barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--ordinary--boolean-receipt-probe-py) | 诊断装置 | 仅原轮次 |
| [evidence/ticket-15/new-actual/final-gate/ordinary/check-boundaries.py](#barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--ordinary--check-boundaries-py) | 测试 | 仅原轮次 |
| [evidence/ticket-15/new-actual/final-gate/ordinary/check-probe-bounds.py](#barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--ordinary--check-probe-bounds-py) | 测试 | 仅原轮次 |
| [evidence/ticket-15/new-actual/final-gate/ordinary/field-repair/verify-negative-fixtures.py](#barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--ordinary--field-repair--verify-negative-fixtures-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-15/new-actual/final-gate/ordinary/freeze.py](#barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--ordinary--freeze-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-15/new-actual/final-gate/ordinary/readers.py](#barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--ordinary--readers-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-15/new-actual/final-gate/ordinary/tdd/verify-sealed-bytes.py](#barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--ordinary--tdd--verify-sealed-bytes-py) | 诊断装置 | 仅原轮次 |
| [evidence/ticket-15/new-actual/final-gate/ordinary/verify.py](#barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--ordinary--verify-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-15/new-actual/final-gate/race/check-boundaries.py](#barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--race--check-boundaries-py) | 测试 | 仅原轮次 |
| [evidence/ticket-15/new-actual/final-gate/race/freeze.py](#barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--race--freeze-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-15/new-actual/final-gate/race/readers.py](#barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--race--readers-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-15/new-actual/final-gate/race/verify.py](#barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--race--verify-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-15/new-actual/final-gate/verdict/check-boundaries.py](#barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--verdict--check-boundaries-py) | 测试 | 仅原轮次 |
| [evidence/ticket-15/new-actual/final-gate/verdict/freeze.py](#barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--verdict--freeze-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-15/new-actual/final-gate/verdict/readers.py](#barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--verdict--readers-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-15/new-actual/final-gate/verdict/verify.py](#barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--verdict--verify-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-15/new-actual/glm/check-boundaries.py](#barness-ai-minimax--evidence--ticket-15--new-actual--glm--check-boundaries-py) | 测试 | 仅原轮次 |
| [evidence/ticket-15/new-actual/glm/complete-snapshot.py](#barness-ai-minimax--evidence--ticket-15--new-actual--glm--complete-snapshot-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-15/new-actual/glm/freeze.py](#barness-ai-minimax--evidence--ticket-15--new-actual--glm--freeze-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-15/new-actual/glm/provenance.py](#barness-ai-minimax--evidence--ticket-15--new-actual--glm--provenance-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-15/new-actual/glm/readers.py](#barness-ai-minimax--evidence--ticket-15--new-actual--glm--readers-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-15/new-actual/glm/verify.py](#barness-ai-minimax--evidence--ticket-15--new-actual--glm--verify-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-16/deadline-alignment/check-boundaries.py](#barness-ai-minimax--evidence--ticket-16--deadline-alignment--check-boundaries-py) | 测试 | 仅原轮次 |
| [evidence/ticket-16/deadline-alignment/snapshot.py](#barness-ai-minimax--evidence--ticket-16--deadline-alignment--snapshot-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-16/deadline-alignment/verify.py](#barness-ai-minimax--evidence--ticket-16--deadline-alignment--verify-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-17/check-boundaries.py](#barness-ai-minimax--evidence--ticket-17--check-boundaries-py) | 测试 | 仅原轮次 |
| [evidence/ticket-17/contract.py](#barness-ai-minimax--evidence--ticket-17--contract-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-17/expand-trace.py](#barness-ai-minimax--evidence--ticket-17--expand-trace-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-17/freeze.py](#barness-ai-minimax--evidence--ticket-17--freeze-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-17/readers.py](#barness-ai-minimax--evidence--ticket-17--readers-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-17/retention-fixes/creator-probe.py](#barness-ai-minimax--evidence--ticket-17--retention-fixes--creator-probe-py) | 诊断装置 | 仅原轮次 |
| [evidence/ticket-17/retention-fixes/factory-correction/check-factory.py](#barness-ai-minimax--evidence--ticket-17--retention-fixes--factory-correction--check-factory-py) | 诊断装置 | 仅原轮次 |
| [evidence/ticket-17/retention-fixes/factory-correction/raw-v2/controllers/check-factory.py](#barness-ai-minimax--evidence--ticket-17--retention-fixes--factory-correction--raw-v2--controllers--check-factory-py) | 历史源码 | 封存源码快照 |
| [evidence/ticket-17/retention-fixes/factory-correction/raw-v2/controllers/run-legitimate.py](#barness-ai-minimax--evidence--ticket-17--retention-fixes--factory-correction--raw-v2--controllers--run-legitimate-py) | 历史源码 | 封存源码快照 |
| [evidence/ticket-17/retention-fixes/factory-correction/raw/controllers/check-factory.py](#barness-ai-minimax--evidence--ticket-17--retention-fixes--factory-correction--raw--controllers--check-factory-py) | 历史源码 | 封存源码快照 |
| [evidence/ticket-17/retention-fixes/factory-correction/raw/controllers/run-legitimate.py](#barness-ai-minimax--evidence--ticket-17--retention-fixes--factory-correction--raw--controllers--run-legitimate-py) | 历史源码 | 封存源码快照 |
| [evidence/ticket-17/retention-fixes/factory-correction/verify-factory.py](#barness-ai-minimax--evidence--ticket-17--retention-fixes--factory-correction--verify-factory-py) | 诊断装置 | 仅原轮次 |
| [evidence/ticket-17/retention-fixes/positive-creator.py](#barness-ai-minimax--evidence--ticket-17--retention-fixes--positive-creator-py) | 诊断装置 | 仅原轮次 |
| [evidence/ticket-17/retention-fixes/resource-probe.py](#barness-ai-minimax--evidence--ticket-17--retention-fixes--resource-probe-py) | 诊断装置 | 仅原轮次 |
| [evidence/ticket-17/retention-fixes/verify-corrections.py](#barness-ai-minimax--evidence--ticket-17--retention-fixes--verify-corrections-py) | 诊断装置 | 仅原轮次 |
| [evidence/ticket-17/verify.py](#barness-ai-minimax--evidence--ticket-17--verify-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-19/native-home-redaction/verify.py](#barness-ai-minimax--evidence--ticket-19--native-home-redaction--verify-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-19/native-renewal/verify.py](#barness-ai-minimax--evidence--ticket-19--native-renewal--verify-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-20/jev_fixtures.py](#barness-ai-minimax--evidence--ticket-20--jev_fixtures-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-20/run.py](#barness-ai-minimax--evidence--ticket-20--run-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-20/verify.py](#barness-ai-minimax--evidence--ticket-20--verify-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-21/actual-retained/public-guards.py](#barness-ai-minimax--evidence--ticket-21--actual-retained--public-guards-py) | 测试 | 仅原轮次 |
| [evidence/ticket-21/actual-retained/verify.py](#barness-ai-minimax--evidence--ticket-21--actual-retained--verify-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-21/final-gate/supplement/verify.py](#barness-ai-minimax--evidence--ticket-21--final-gate--supplement--verify-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-21/final-gate/verify.py](#barness-ai-minimax--evidence--ticket-21--final-gate--verify-py) | 历史源码 | 仅原轮次 |
| [session-analysis/build_report.py](#barness-ai-minimax--session-analysis--build_report-py) | 历史源码 | 仅原轮次 |
| [session-analysis/collect.py](#barness-ai-minimax--session-analysis--collect-py) | 历史源码 | 仅原轮次 |
| [tools/eleven_contract.py](#barness-ai-minimax--tools--eleven_contract-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/eleven_finalize.py](#barness-ai-minimax--tools--eleven_finalize-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/eleven_grant.py](#barness-ai-minimax--tools--eleven_grant-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/eleven_inputs.py](#barness-ai-minimax--tools--eleven_inputs-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/eleven_validation.py](#barness-ai-minimax--tools--eleven_validation-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/glm_corrective.py](#barness-ai-minimax--tools--glm_corrective-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/paid.py](#barness-ai-minimax--tools--paid-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/paid_capture.py](#barness-ai-minimax--tools--paid_capture-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/paid_checks.py](#barness-ai-minimax--tools--paid_checks-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/paid_core.py](#barness-ai-minimax--tools--paid_core-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/paid_history.py](#barness-ai-minimax--tools--paid_history-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/paid_plan.py](#barness-ai-minimax--tools--paid_plan-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/release_repair_validation.py](#barness-ai-minimax--tools--release_repair_validation-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/renewal.py](#barness-ai-minimax--tools--renewal-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/renewal_common.py](#barness-ai-minimax--tools--renewal_common-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/renewal_glm.py](#barness-ai-minimax--tools--renewal_glm-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/repair_anchors.py](#barness-ai-minimax--tools--repair_anchors-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/repair_contract.py](#barness-ai-minimax--tools--repair_contract-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/repair_inputs.py](#barness-ai-minimax--tools--repair_inputs-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/repair_publication.py](#barness-ai-minimax--tools--repair_publication-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/routed_contract.py](#barness-ai-minimax--tools--routed_contract-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/routed_evidence.py](#barness-ai-minimax--tools--routed_evidence-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/routed_validation.py](#barness-ai-minimax--tools--routed_validation-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/test_eleven_validation.py](#barness-ai-minimax--tools--test_eleven_validation-py) | 测试 | 显式合成测试 |
| [tools/test_glm_corrective.py](#barness-ai-minimax--tools--test_glm_corrective-py) | 测试 | 显式合成测试 |
| [tools/test_paid.py](#barness-ai-minimax--tools--test_paid-py) | 测试 | 显式合成测试 |
| [tools/test_release_repair.py](#barness-ai-minimax--tools--test_release_repair-py) | 测试 | 显式合成测试 |
| [tools/test_renewal.py](#barness-ai-minimax--tools--test_renewal-py) | 测试 | 显式合成测试 |
| [tools/test_repair_fixtures.py](#barness-ai-minimax--tools--test_repair_fixtures-py) | 测试 | 显式合成测试 |
| [tools/test_routed_validation.py](#barness-ai-minimax--tools--test_routed_validation-py) | 测试 | 显式合成测试 |

<a id="barness-ai-minimax--evidence--ticket-08--rollback-py"></a>

## evidence/ticket-08/rollback.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-08/rollback.py](../../barness-ai-minimax/evidence/ticket-08/rollback.py)；SHA256 `ec7d752bc3d56dbdf2f873c589c4fb74b5221756e1dcb6186de1004fc891c88a`。活动工具 / 现有公共消费者。Go公共测试或原生消费者仍有实际执行/复制接线，目录年代不表示退役。

**职责**：Restore and verify the complete pre-MiniMax software outside user checkouts。

**调用方**：[ai/e2e/minimax_full_rollback_test.go](../../../ai/e2e/minimax_full_rollback_test.go) L40（confirmed-public-Go-consumer）。

**输入**：CLI新仓库相对out、完整Git基线对象/树、rollback-probe.txt；GLM还需commands.py模块。参数：`parser.add_argument('out')` L91。

**输出/写入**：out内baseline.tar/完整blob与mode库存/探针二进制/命令日志/回滚报告；系统临时目录中的恢复树。实际静态写点：L26 `path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')`；L93 `out.mkdir(parents=True, exist_ok=False, mode=448)`；L94 `dump(out / 'root-before.json', before)`；L137 `dump(out / 'root-after.json', after)`；L139 `dump(out / 'rollback-report.json', dict(passVerdict=True, baselineCommit=BASELINE, baselineTree=TREE, completeTrackedFiles=len(entries), archiveSHA256=digest(archive.read_bytes()), restoredTreeVerified=True, rootUnchanged=True, publicNoMiniMaxRefusal=True, oldProviderControlledE2E=True, baselineBinarySHA256=binary_hash, ownedTemporaryRestorationRem`；L99 `restored.mkdir()`；L105 `dump(out / 'restored-tree.json', inventory)`；L106 `probe.write_bytes((HERE / 'rollback-probe.txt').read_bytes())`；L119 `dump(out / 'commands.json', receipts)`；L121 `dump(out / 'module-pins.json', {p: digest((restored / p).read_bytes()) for p in ('go.mod', 'go.sum')})`；L123 `dump(out / 'baseline-release-index.json', {p: inventory[p] for p in selected})`；L135 `dump(out / 'baseline-old-provider-summary.json', dict(cases=cases, audit=audit, manifestSHA256=digest(manifests[0].read_bytes()), source='unaltered restored baseline'))`；L104 `tar.extractall(restored, filter='data')`；L116 `(out / (name + '.log')).write_text(scrub(proc.stdout.decode(errors='replace')))`。

**副作用**：读取Git并导出旧完整树；解包临时恢复、go build及一个受控旧Provider E2E，清理自有临时树。是普通测试辅助，不是生产回滚操作或新的live资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现行SOP与实际调用方选择的完整来源/输入；本轮文件哈希只固定审查对象，不替代原生资格。当前替代入口：[ai/e2e/minimax_full_rollback_test.go](../../../ai/e2e/minimax_full_rollback_test.go)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-08`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L13 `Path(__file__).resolve()`；L14 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-09--international-final-actual--a_core-py"></a>

## evidence/ticket-09/international-final-actual/a_core.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-09/international-final-actual/a_core.py](../../barness-ai-minimax/evidence/ticket-09/international-final-actual/a_core.py)；SHA256 `2a7c6cd1a3fc6513ad981aca1005c48f9ef649fb4b50acaa53c321cc89dfcb63`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Pure bounded archive/reference helpers. No process, environment or network use。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-09/international-final-actual`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--evidence--ticket-09--international-final-actual--curate_a-py"></a>

## evidence/ticket-09/international-final-actual/curate_a.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-09/international-final-actual/curate_a.py](../../barness-ai-minimax/evidence/ticket-09/international-final-actual/curate_a.py)；SHA256 `da808b0d31d260f5d3d937d98168847a4c1091f54f661d0f5dbd6c7b61f042de`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：DRAFT collector: execute only after explicit parent A-CLOSED notification。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：`p.add_argument('config', type=Path)` L13；`p.add_argument('--a-closed', action='store_true', required=True)` L14。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L68 `out.mkdir(parents=True, exist_ok=True)`；L198 `json.dump(result, f, ensure_ascii=False, indent=2)`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-09/international-final-actual`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--evidence--ticket-09--international-final-actual--retention_core-py"></a>

## evidence/ticket-09/international-final-actual/retention_core.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-09/international-final-actual/retention_core.py](../../barness-ai-minimax/evidence/ticket-09/international-final-actual/retention_core.py)；SHA256 `b4ebb3b209a23fc684c809e097843d4b89601ffac966c3c079b77c4a933ae810`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Bounded, no-follow, byte-preserving primitives. No captured command executes。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-09/international-final-actual`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--evidence--ticket-09--international-final-actual--verify_a-py"></a>

## evidence/ticket-09/international-final-actual/verify_a.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-09/international-final-actual/verify_a.py](../../barness-ai-minimax/evidence/ticket-09/international-final-actual/verify_a.py)；SHA256 `b74cd73a6096c0e988e328e8c54f9c740d97ff608c493305d66b26241893cc41`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：DRAFT portable A integrity verifier. Does not execute any captured command。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`p.add_argument('index', type=Path)` L32；`p.add_argument('--external-map', type=Path, required=True, help='Original external archive basenames -> immutable relocated archive files')` L33。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-09/international-final-actual`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--evidence--ticket-10--international-complete-candidates--verify_capsule-py"></a>

## evidence/ticket-10/international-complete-candidates/verify_capsule.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-10/international-complete-candidates/verify_capsule.py](../../barness-ai-minimax/evidence/ticket-10/international-complete-candidates/verify_capsule.py)；SHA256 `70fa10fd10ee9b6207ec307041351b6cf7bcf33340c8b602ffefc11f3ef80542`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Keyless portable byte-integrity verification; never executes captured commands。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`parser.add_argument('index', type=Path)` L10。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-10/international-complete-candidates`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L11 `parser.parse_args()`；L12 `json.loads(args.index.read_bytes())`；L12 `args.index.read_bytes()`；L28 `archive.read_bytes()`；L29 `archive.stat()`。

<a id="barness-ai-minimax--evidence--ticket-10--international-formal-glm--verify_capsule-py"></a>

## evidence/ticket-10/international-formal-glm/verify_capsule.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-10/international-formal-glm/verify_capsule.py](../../barness-ai-minimax/evidence/ticket-10/international-formal-glm/verify_capsule.py)；SHA256 `c9273bfdd5cea2f39b902cd632bc5b972554a53bddadce19116313dd5ea2d4e1`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Portable keyless byte verification, including truthful failed-forensic scope。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`parser.add_argument('index', type=Path)` L10；`parser.add_argument('--candidate-archive', type=Path)` L11。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-10/international-formal-glm`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L12 `parser.parse_args()`；L13 `json.loads(args.index.read_bytes())`；L13 `args.index.read_bytes()`；L29 `archive.read_bytes()`；L30 `archive.stat()`。

<a id="barness-ai-minimax--evidence--ticket-10--international-legacy-final--curate_paid-py"></a>

## evidence/ticket-10/international-legacy-final/curate_paid.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-10/international-legacy-final/curate_paid.py](../../barness-ai-minimax/evidence/ticket-10/international-legacy-final/curate_paid.py)；SHA256 `97894ea221e330e45870f3d47ab7e4dfd439d80007fdf3cddb000ccbf3ec7e59`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Retain closed L paid evidence under /tmp. Never execute a captured command。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：`p.add_argument('config', type=Path)` L12；`p.add_argument('--parent-ready', required=True, action='store_true')` L13。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L22 `out.mkdir(parents=True, exist_ok=True)`；L213 `json.dump(result, f, ensure_ascii=False, indent=2)`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-10/international-legacy-final`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L14 `p.parse_args()`；L16 `json.loads(config_raw)`；L19 `Path(c['out']).resolve()`；L20 `Path('/tmp').resolve()`；L22 `out.mkdir(parents=True, exist_ok=True)`。

<a id="barness-ai-minimax--evidence--ticket-10--international-legacy-final--retention_core-py"></a>

## evidence/ticket-10/international-legacy-final/retention_core.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-10/international-legacy-final/retention_core.py](../../barness-ai-minimax/evidence/ticket-10/international-legacy-final/retention_core.py)；SHA256 `b4ebb3b209a23fc684c809e097843d4b89601ffac966c3c079b77c4a933ae810`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Bounded, no-follow, byte-preserving primitives. No captured command executes。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-10/international-legacy-final`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--evidence--ticket-10--international-legacy-final--verify_paid-py"></a>

## evidence/ticket-10/international-legacy-final/verify_paid.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-10/international-legacy-final/verify_paid.py](../../barness-ai-minimax/evidence/ticket-10/international-legacy-final/verify_paid.py)；SHA256 `39264d0388a3bd50654eb157a2b5da8abe56bf362f38743f8829c6140fa24357`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Portable byte/reference checks for closed L evidence; never execute raw commands。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`p.add_argument('index', type=Path)` L52；`p.add_argument('--external-map', type=Path)` L53；`p.add_argument('--require-external', action='store_true')` L54。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-10/international-legacy-final`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L55 `p.parse_args()`；L56 `json.loads(stable(a.index, 32 << 20))`；L63 `archive.stat()`；L70 `json.loads(m['source-at-start.json'])`；L85 `json.loads(m['metadata/L/prerelease-source-review.json'])`。

<a id="barness-ai-minimax--evidence--ticket-10--international-postactivation--curate-py"></a>

## evidence/ticket-10/international-postactivation/curate.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-10/international-postactivation/curate.py](../../barness-ai-minimax/evidence/ticket-10/international-postactivation/curate.py)；SHA256 `69ecb97fce5e7ee022a1b870a98210129083a7d33e40a5e6cc2951b360101a87`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：curate.py：原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L7 `COMMON.mkdir(parents=True, exist_ok=True)`；L7 `DEST.mkdir(parents=True, exist_ok=True)`；L9 `p.write_text(json.dumps(v, ensure_ascii=False, indent=2) + '\n')`；L96 `partial.unlink()`；L107 `dump(DEST / (name + '.json'), index)`；L101 `json.dump(inventory, f, ensure_ascii=False, sort_keys=True, separators=(',', ':'))`；L103 `shutil.copyfile(target, DEST / target.name)`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-10/international-postactivation`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L2 `R=pathlib.Path('/tmp/barness-minimax-10-global')`；L3 `A=pathlib.Path('/tmp/barness-minimax-postactivation-red')`；L4 `COMMON=pathlib.Path('/Users/cyber/RestoX/barness/.evidence/minimax-retained')`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L7 `COMMON.mkdir(parents=True, exist_ok=True)`；L7 `DEST.mkdir(parents=True, exist_ok=True)`。

<a id="barness-ai-minimax--evidence--ticket-12--actual-corrective-1--verify_actual-py"></a>

## evidence/ticket-12/actual-corrective-1/verify_actual.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-12/actual-corrective-1/verify_actual.py](../../barness-ai-minimax/evidence/ticket-12/actual-corrective-1/verify_actual.py)；SHA256 `f143fde5cf5932442b2e29b6f69a77a52f36940fd88f7ee3de0b82dbc48047c0`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Portable byte checks only; no credentials, captured commands or vendor calls。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`parser.add_argument('index', type=Path)` L13；`parser.add_argument('--prepay-index', type=Path, required=True)` L14；`parser.add_argument('--prepay-archive', type=Path, required=True)` L15；`parser.add_argument('--prepay-inventory', type=Path, required=True)` L16；`parser.add_argument('--lineage-archive', type=Path, action='append', default=[])` L17。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-12/actual-corrective-1`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L18 `parser.parse_args()`；L41 `json.loads(args.index.read_bytes())`；L41 `args.index.read_bytes()`；L43 `archive.stat()`；L47 `tarfile.open(archive, 'r|gz')`。

<a id="barness-ai-minimax--evidence--ticket-12--curation-tool-py"></a>

## evidence/ticket-12/curation-tool.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-12/curation-tool.py](../../barness-ai-minimax/evidence/ticket-12/curation-tool.py)；SHA256 `c816c8c533584fb6911d6e64326fbfc1e9fdecb0fc81ebf4a9233a26fd057d84`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：curation-tool.py：原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L7 `COMMON.mkdir(parents=True, exist_ok=True)`；L7 `DEST.mkdir(parents=True, exist_ok=True)`；L9 `p.write_text(json.dumps(v, ensure_ascii=False, indent=2) + '\n')`；L96 `partial.unlink()`；L107 `dump(DEST / (name + '.json'), index)`；L101 `json.dump(inventory, f, ensure_ascii=False, sort_keys=True, separators=(',', ':'))`；L103 `shutil.copyfile(target, DEST / target.name)`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-12`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L2 `R=pathlib.Path('/tmp/barness-minimax-10-global')`；L3 `A=pathlib.Path('/tmp/barness-minimax-postactivation-red')`；L4 `COMMON=pathlib.Path('/Users/cyber/RestoX/barness/.evidence/minimax-retained')`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L7 `COMMON.mkdir(parents=True, exist_ok=True)`；L7 `DEST.mkdir(parents=True, exist_ok=True)`。

<a id="barness-ai-minimax--evidence--ticket-12--original-preflight-source-proof--retain_source-py-gz"></a>

## evidence/ticket-12/original-preflight-source-proof/retain_source.py.gz

原件：[.scratch/barness-ai-minimax/evidence/ticket-12/original-preflight-source-proof/retain_source.py.gz](../../barness-ai-minimax/evidence/ticket-12/original-preflight-source-proof/retain_source.py.gz)；SHA256 `638a83bf1debd452fda0cdf55a61c1f9ceb2b68ee008504d988580c140cc7389`。历史源码 / 封存源码快照。压缩封存源码，不是活动CLI。

**职责**：Add committed source bytes to an externally retained failed keyless window。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：只作为gzip封存的原Python源码字节；解压仅供静态阅读。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本定义的历史来源/协议核验或补证输出；本次仅登记压缩与解压SHA256。实际静态写点：L100 `index_path.write_text(json.dumps(index, ensure_ascii=False, indent=2) + '\n')`；L107 `(OUT / 'verification-result.json').write_text(json.dumps(summary, indent=2) + '\n')`；L125 `audit_path.write_text(text)`；L138 `provenance_path.write_text(json.dumps(provenance, ensure_ascii=False, indent=2) + '\n')`。

**副作用**：不得把源码快照当活动工具解压执行；原源码的CLI/IO/子进程静态点仍列出，不修改归档或重发历史结论。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-12/original-preflight-source-proof`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

解压文本SHA256：`21c69e0c3e4f99af7fa5e4bc0d1f121587022b0b0e64ead05964658bc372da33`；保持压缩原件，不解压执行。

**固定宿主路径**：L11 `ROOT = Path('/Users/cyber/RestoX/barness')`；L12 `OUT = Path('/tmp/barness-minimax-root-preflight-failure-capsule')`；L13 `WORKER = Path('/tmp/barness-minimax-10-global/.scratch/barness-ai-minimax/evidence/ticket-12')`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L21 `json.loads(source_raw)`；L27 `subprocess.run(['git', 'cat-file', '--batch'], cwd=ROOT, input=queries, stdout=subprocess.PIPE, stderr=subprocess.PIPE, check=True)`；L52 `RAW_INDEX.read_bytes()`；L53 `(json.dumps(audit, ensure_ascii=False, indent=2) + '\n').encode()`；L53 `json.dumps(audit, ensure_ascii=False, indent=2)`。

<a id="barness-ai-minimax--evidence--ticket-12--original-preflight-source-proof--verify_q-py"></a>

## evidence/ticket-12/original-preflight-source-proof/verify_q.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-12/original-preflight-source-proof/verify_q.py](../../barness-ai-minimax/evidence/ticket-12/original-preflight-source-proof/verify_q.py)；SHA256 `a75a9484b0ffc37a50d00cd52f03efa75462197883241ad4e9e9a17aea5f7455`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Read-only bounded verifier. It never extracts links or executes captured commands。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`parser.add_argument('--raw-archive', type=Path, required=True)` L182；`parser.add_argument('--raw-index', type=Path, required=True)` L183；`parser.add_argument('--raw-inventory', type=Path, required=True)` L184；`parser.add_argument('--source-index', type=Path)` L185。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-12/original-preflight-source-proof`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--evidence--ticket-12--retention-bounds--test_portable-py"></a>

## evidence/ticket-12/retention-bounds/test_portable.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-12/retention-bounds/test_portable.py](../../barness-ai-minimax/evidence/ticket-12/retention-bounds/test_portable.py)；SHA256 `f961ebe800794254e6e13c7234845031d7686e1fb5d66a105c3fdbf71e84714b`。测试 / 仅原轮次。原归档/原控制器的失败清单，未发现当前公共Go或现行SOP执行接线。

**职责**：Public CLI resource/failure E2E; synthetic inputs never qualify live support。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定源码/工具/归档和私有篡改副本或合成fixture；支持模块随对应测试使用。参数：`parser.add_argument('out', type=Path)` L94；`parser.add_argument('--archives', type=Path, default=COMMON)` L95；`parser.add_argument('--historical', action='store_true')` L96。

**输出/写入**：该历史失败清单的断言/拒绝记录和临时日志；不续期真实资格。实际静态写点：L39 `path.write_text(json.dumps(value, indent=2) + '\n')`；L55 `(out / (label + '.log')).write_text(safe)`；L67 `pp.write_bytes(raw)`；L86 `dump(index, idx)`；L102 `out.mkdir(parents=True, exist_ok=False)`；L115 `dump(out / 'verifier-source.json', source)`；L162 `dump(out / 'results.json', {'cases': results, 'pass': all((x['pass'] for x in results)), 'credentialsRead': False, 'networkUsed': False})`；L124 `bad.write_bytes(b'not gzip')`；L129 `dump(qi, idx)`；L141 `dump(qi, idx)`；L146 `extra.write_bytes(b'x' * ((8 << 20) + 1))`；L151 `truncated.write_bytes(gzip.compress(b'{}')[:-4])`；L153 `dump(qi, idx)`；L157 `dump(out / 'fixture.json', {'syntheticOnly': True, 'expandedBombBytes': 96 << 20, 'compressedBombBytes': bomb.stat().st_size, 'compressedBombSHA256': sha(bomb.read_bytes()), 'concatenatedExpandedBytes': 48 << 20, 'historicalVerifier': historical})`；L112 `target.write_bytes(tar.extractfile(member).read())`。

**副作用**：运行相邻公共CLI或临时程序，写/删私有fixture；部分故障测试会临时替换再恢复输入，仅原环境适用，不在本次执行。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-12/retention-bounds`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L14 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-13--actual-glm--check-boundaries-py"></a>

## evidence/ticket-13/actual-glm/check-boundaries.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-13/actual-glm/check-boundaries.py](../../barness-ai-minimax/evidence/ticket-13/actual-glm/check-boundaries.py)；SHA256 `cbd467caf593285fd09290244688cea2276a618984b4b6de2cea25b6b64e1b3c`。测试 / 仅原轮次。原归档/原控制器的失败清单，未发现当前公共Go或现行SOP执行接线。

**职责**：Replay seven integrity refusals using temporary metadata and synthetic files。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定源码/工具/归档和私有篡改副本或合成fixture；支持模块随对应测试使用。参数：`parser.add_argument('--common-root', type=Path, default=BASE.parents[4])` L17。

**输出/写入**：该历史失败清单的断言/拒绝记录和临时日志；不续期真实资格。实际静态写点：L30 `path.parent.mkdir(parents=True)`；L41 `shutil.copytree(BASE, copied)`；L42 `(root / 'implementation').mkdir()`；L42 `shutil.copyfile(BASE.parent / 'implementation/verify.py', root / 'implementation/verify.py')`；L53 `(copied / 'SHA256SUMS').write_text(''.join((hashlib.sha256(p.read_bytes()).hexdigest() + '  ' + p.relative_to(copied).as_posix() + '\n' for p in paths)))`；L45 `(copied / 'summary.json').write_text(json.dumps(value))`；L33 `path.symlink_to(a.common_root / archive['file'])`；L49 `(copied / 'scenarios.jsonl').write_text(''.join((json.dumps(v) + '\n' for v in values)))`；L36 `path.parent.symlink_to((a.common_root / archive['file']).parent, target_is_directory=True)`。

**副作用**：运行相邻公共CLI或临时程序，写/删私有fixture；部分故障测试会临时替换再恢复输入，仅原环境适用，不在本次执行。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-13/actual-glm`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L13 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-13--actual-glm--verify-py"></a>

## evidence/ticket-13/actual-glm/verify.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-13/actual-glm/verify.py](../../barness-ai-minimax/evidence/ticket-13/actual-glm/verify.py)；SHA256 `f2c6507aa27e889fb67a466ab2ea92f1ac137f2db7fb777ae81cb800cc4709ec`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Read-only actual GLM integrity/accounting; never grants PASS or replay。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`parser.add_argument('--common-root', type=Path, default=BASE.parents[4])` L139。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-13/actual-glm`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L13 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-13--actual-legacy--check-boundaries-py"></a>

## evidence/ticket-13/actual-legacy/check-boundaries.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-13/actual-legacy/check-boundaries.py](../../barness-ai-minimax/evidence/ticket-13/actual-legacy/check-boundaries.py)；SHA256 `3b16b84813d8a63d8aa0249652f4c8425e549bc35c062a17e20c6619eb8ab4f7`。测试 / 仅原轮次。原归档/原控制器的失败清单，未发现当前公共Go或现行SOP执行接线。

**职责**：Replay eight integrity refusals using temporary metadata and synthetic files。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定源码/工具/归档和私有篡改副本或合成fixture；支持模块随对应测试使用。参数：`parser.add_argument('--common-root', type=Path, default=BASE.parents[4])` L17。

**输出/写入**：该历史失败清单的断言/拒绝记录和临时日志；不续期真实资格。实际静态写点：L30 `path.parent.mkdir(parents=True)`；L42 `shutil.copytree(BASE, copied)`；L43 `(root / 'implementation').mkdir()`；L43 `shutil.copyfile(BASE.parent / 'implementation/verify.py', root / 'implementation/verify.py')`；L44 `(root / 'actual-glm').mkdir()`；L44 `shutil.copyfile(BASE.parent / 'actual-glm/summary.json', root / 'actual-glm/summary.json')`；L58 `(copied / 'SHA256SUMS').write_text(''.join((hashlib.sha256(p.read_bytes()).hexdigest() + '  ' + p.relative_to(copied).as_posix() + '\n' for p in paths)))`；L47 `(copied / 'summary.json').write_text(json.dumps(value))`；L33 `path.symlink_to(a.common_root / archive['file'])`；L51 `(copied / 'scenarios.jsonl').write_text(''.join((json.dumps(v) + '\n' for v in values)))`；L36 `path.parent.symlink_to((a.common_root / archive['file']).parent, target_is_directory=True)`；L54 `(copied / 'index.json').write_text(json.dumps(value))`。

**副作用**：运行相邻公共CLI或临时程序，写/删私有fixture；部分故障测试会临时替换再恢复输入，仅原环境适用，不在本次执行。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-13/actual-legacy`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L13 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-13--actual-legacy--verify-py"></a>

## evidence/ticket-13/actual-legacy/verify.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-13/actual-legacy/verify.py](../../barness-ai-minimax/evidence/ticket-13/actual-legacy/verify.py)；SHA256 `6aca7e2f1cd08337cc131a7f4922bb5056ee017c4f30217b4fc6818155bd31d6`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Read-only actual legacy integrity/accounting and original fixed quality metrics。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`parser.add_argument('--common-root', type=Path, default=BASE.parents[4])` L173。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-13/actual-legacy`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L13 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-13--final-gate--capsule-py"></a>

## evidence/ticket-13/final-gate/capsule.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-13/final-gate/capsule.py](../../barness-ai-minimax/evidence/ticket-13/final-gate/capsule.py)；SHA256 `3017cba1d4ac09e0a203c68ac768a4d508b861fa862aa04c9ca4550efebd9328`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Finite streaming reader for each fixed complete ordinary/race suite.  This is a distinct retention protocol: the earlier implementation capsule has 32MiB/192MiB/32k limits and remains unchanged. Here 133k complete-suite files require explicit fixed larger limits; no captured command is executed。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-13/final-gate`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--evidence--ticket-13--final-gate--check-boundaries-py"></a>

## evidence/ticket-13/final-gate/check-boundaries.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-13/final-gate/check-boundaries.py](../../barness-ai-minimax/evidence/ticket-13/final-gate/check-boundaries.py)；SHA256 `aed537cad23b9f84d1fc817765eafed376740ec96864c4ffdbebc05379a45ab3`。测试 / 仅原轮次。原归档/原控制器的失败清单，未发现当前公共Go或现行SOP执行接线。

**职责**：Fourteen finite public CLI refusals; originals and captured commands stay untouched.  Failure inventory: corrupted/oversized/link archive and linked parent, altered outer verdict/numeric exit/unit failure, redirected archive, enlarged limits, tar link, changed helper/source/binary and concealed suite failure count。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定源码/工具/归档和私有篡改副本或合成fixture；支持模块随对应测试使用。参数：`parser.add_argument('--common-root', type=Path, default=BASE.parents[4])` L29。

**输出/写入**：该历史失败清单的断言/拒绝记录和临时日志；不续期真实资格。实际静态写点：L25 `(base / 'SHA256SUMS').write_text(''.join((hashlib.sha256(p.read_bytes()).hexdigest() + '  ' + p.name + '\n' for p in files)))`；L41 `path.parent.mkdir(parents=True)`；L41 `shutil.copyfile(a.common_root / verdict['file'], root / verdict['file'])`；L54 `shutil.copytree(BASE, copied)`；L54 `helper.mkdir()`；L55 `shutil.copyfile(BASE.parent / 'implementation/verify.py', helper / 'verify.py')`；L62 `path.write_text(json.dumps(value))`；L44 `path.symlink_to(a.common_root / ordinary['file'])`；L47 `(root / verdict['file']).unlink()`；L47 `path.parent.symlink_to((a.common_root / ordinary['file']).parent, target_is_directory=True)`；L66 `path.write_text(json.dumps(value))`；L79 `path.write_text(json.dumps(value))`；L73 `destination.parent.mkdir(parents=True)`。

**副作用**：运行相邻公共CLI或临时程序，写/删私有fixture；部分故障测试会临时替换再恢复输入，仅原环境适用，不在本次执行。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-13/final-gate`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L20 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-13--final-gate--snapshot-ordinary-py"></a>

## evidence/ticket-13/final-gate/snapshot-ordinary.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-13/final-gate/snapshot-ordinary.py](../../barness-ai-minimax/evidence/ticket-13/final-gate/snapshot-ordinary.py)；SHA256 `b355709278ab8795fba0eab5047ea4aeb53cd3c052e1947d49ed304e0f547b41`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Freeze only the completed ordinary subtree and three completed numeric receipts。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L10 `BASE.mkdir()`；L10 `EXT.mkdir()`；L40 `(BASE / 'ordinary-inventory.json.gz').write_bytes(encoded)`；L41 `(BASE / 'source-at-start.json.gz').write_bytes(source_encoded)`；L54 `dump(BASE / 'ordinary-index.json', index)`；L55 `archive.chmod(292)`；L8 `p.write_text(json.dumps(v, ensure_ascii=False, separators=(',', ':')) + '\n')`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-13/final-gate`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L4 `ROOT=Path('/Users/cyber/RestoX/barness');OWN=Path('/tmp/barness-minimax-routed-validation')`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L10 `BASE.mkdir()`；L10 `EXT.mkdir()`；L17 `json.loads(p.read_text())`；L17 `p.read_text()`；L21 `p.stat()`。

<a id="barness-ai-minimax--evidence--ticket-13--final-gate--snapshot-race-py"></a>

## evidence/ticket-13/final-gate/snapshot-race.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-13/final-gate/snapshot-race.py](../../barness-ai-minimax/evidence/ticket-13/final-gate/snapshot-race.py)；SHA256 `010abcd97b363625ce946fb31ef55165fe2b1506a1051717aa07564b4c955c0d`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Read-only freeze of the completed actual race subtree and its numeric receipt。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L39 `(BASE / 'race-inventory.json.gz').write_bytes(encoded)`；L59 `dump(BASE / 'race-index.json', index)`；L59 `archive.chmod(292)`；L8 `p.write_text(json.dumps(v, ensure_ascii=False, separators=(',', ':')) + '\n')`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-13/final-gate`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L4 `ROOT=Path('/Users/cyber/RestoX/barness');OWN=Path('/tmp/barness-minimax-routed-validation')`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L16 `json.loads(record_path.read_text())`；L16 `record_path.read_text()`；L20 `p.lstat()`；L21 `json.dumps({'phase': 'archiving-completed-race-only', 'files': len(selected), 'rawBytes': sum((p.lstat().st_size for p in selected.values()))})`；L21 `p.lstat()`。

<a id="barness-ai-minimax--evidence--ticket-13--final-gate--snapshot-verdict-py"></a>

## evidence/ticket-13/final-gate/snapshot-verdict.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-13/final-gate/snapshot-verdict.py](../../barness-ai-minimax/evidence/ticket-13/final-gate/snapshot-verdict.py)；SHA256 `c9b40146dcabdb62aff983de796b76614ef8f2a1fd1d111a6d2a5becc16250c6`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Freeze only the closed compiled gate outer metadata, excluding sealed suites。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L33 `(BASE / 'verdict-inventory.json.gz').write_bytes(encoded)`；L41 `dump(BASE / 'verdict-index.json', index)`；L42 `dump(BASE / 'binary-reference.json', dict(schema=1, file='$REPO/' + binary.relative_to(ROOT).as_posix(), bytes=binary.stat().st_size, sha256=sha(binary.read_bytes()), copied=False, existingImmutableArchive='.evidence/minimax-retained/ticket-13-keyless-launch-history/keyless-launch-history-raw.tar.gz', existingMember='compiled-prepared/releasegate')`；L44 `archive.chmod(292)`；L7 `p.write_text(json.dumps(v, ensure_ascii=False, separators=(',', ':')) + '\n')`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-13/final-gate`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L4 `ROOT=Path('/Users/cyber/RestoX/barness');OWN=Path('/tmp/barness-minimax-routed-validation')`；L21 `selected['diagnosis/postgate-gap-2.json']=Path('/tmp/barness-routed-postgate-gap-2.json')`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L8 `json.loads((A / 'summary.json').read_text())`；L8 `(A / 'summary.json').read_text()`；L8 `json.loads((A / 'gate/release-report.json').read_text())`；L8 `(A / 'gate/release-report.json').read_text()`；L14 `dirs.remove(d)`。

<a id="barness-ai-minimax--evidence--ticket-13--final-gate--verify-py"></a>

## evidence/ticket-13/final-gate/verify.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-13/final-gate/verify.py](../../barness-ai-minimax/evidence/ticket-13/final-gate/verify.py)；SHA256 `2f3728842f58a68f5d05cbc448216fb8c1ce3adc4e64d73ea587060879938ad4`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Read-only finite verification of the failed compiled A2 full gate at8fbd。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`parser.add_argument('--common-root', type=Path, default=BASE.parents[4])` L120。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-13/final-gate`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L13 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-13--implementation--verify-py"></a>

## evidence/ticket-13/implementation/verify.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-13/implementation/verify.py](../../barness-ai-minimax/evidence/ticket-13/implementation/verify.py)；SHA256 `b280863f49f4a56cb507ee7e6ba753c73e38c9162236592e9f886ed9a64c3c2f`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Read-only, bounded verification of the two fixed keyless history capsules。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`parser.add_argument('--common-root', type=Path, default=BASE.parents[4])` L144。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-13/implementation`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L14 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-13--keyless-launch-history--capsule-py"></a>

## evidence/ticket-13/keyless-launch-history/capsule.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-13/keyless-launch-history/capsule.py](../../barness-ai-minimax/evidence/ticket-13/keyless-launch-history/capsule.py)；SHA256 `1d2fc4888e51b18c946dfa8d216b18e07a50551c54df900422815e87adc42e5e`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Finite streaming reader for this large interrupted launch history only.  This is a distinct retention protocol: the earlier implementation capsule has 32MiB/192MiB/32k limits and remains unchanged. Here 81k incomplete-suite files require explicit fixed larger limits; no captured command is executed。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-13/keyless-launch-history`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--evidence--ticket-13--keyless-launch-history--check-boundaries-py"></a>

## evidence/ticket-13/keyless-launch-history/check-boundaries.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-13/keyless-launch-history/check-boundaries.py](../../barness-ai-minimax/evidence/ticket-13/keyless-launch-history/check-boundaries.py)；SHA256 `f56b645e1864583271a241ef79bb5680f16acdde265751f0c61905833e6e270c`。测试 / 仅原轮次。原归档/原控制器的失败清单，未发现当前公共Go或现行SOP执行接线。

**职责**：Replay ten fixed launch-history refusals using temporary metadata and synthetic files。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定源码/工具/归档和私有篡改副本或合成fixture；支持模块随对应测试使用。参数：`parser.add_argument('--common-root', type=Path, default=BASE.parents[4])` L17。

**输出/写入**：该历史失败清单的断言/拒绝记录和临时日志；不续期真实资格。实际静态写点：L30 `path.parent.mkdir(parents=True)`；L43 `shutil.copytree(BASE, copied)`；L44 `(root / 'implementation').mkdir()`；L44 `shutil.copyfile(BASE.parent / 'implementation/verify.py', root / 'implementation/verify.py')`；L59 `(copied / 'SHA256SUMS').write_text(''.join((hashlib.sha256(p.read_bytes()).hexdigest() + '  ' + p.relative_to(copied).as_posix() + '\n' for p in paths)))`；L49 `(copied / 'summary.json').write_text(json.dumps(value))`；L33 `path.symlink_to(a.common_root / archive['file'])`；L55 `(copied / 'index.json').write_text(json.dumps(value))`；L36 `path.parent.symlink_to((a.common_root / archive['file']).parent, target_is_directory=True)`。

**副作用**：运行相邻公共CLI或临时程序，写/删私有fixture；部分故障测试会临时替换再恢复输入，仅原环境适用，不在本次执行。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-13/keyless-launch-history`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L13 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-13--keyless-launch-history--verify-py"></a>

## evidence/ticket-13/keyless-launch-history/verify.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-13/keyless-launch-history/verify.py](../../barness-ai-minimax/evidence/ticket-13/keyless-launch-history/verify.py)；SHA256 `ee67a34025960e6770e54a68e45eac7b80ca16e9fb7e865259b510b7e38fbf2e`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Read-only interrupted launch/observer/preparation history; no release certificate。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`parser.add_argument('--common-root', type=Path, default=BASE.parents[4])` L126。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-13/keyless-launch-history`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L13 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-14--gap2--capsule-py"></a>

## evidence/ticket-14/gap2/capsule.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-14/gap2/capsule.py](../../barness-ai-minimax/evidence/ticket-14/gap2/capsule.py)；SHA256 `651903be831bcff9ac172f2f8e1b3cbd5f469460760bfaa3af9af59a08c30f84`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Finite streaming reader for the ticket14 keyless repair history only.  This is a distinct retention protocol: the earlier implementation capsule has 32MiB/192MiB/32k limits and remains unchanged. Here two preflight attempts and native child fixtures require explicit fixed larger limits; no captured command is executed。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-14/gap2`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--evidence--ticket-14--gap2--check-boundaries-py"></a>

## evidence/ticket-14/gap2/check-boundaries.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-14/gap2/check-boundaries.py](../../barness-ai-minimax/evidence/ticket-14/gap2/check-boundaries.py)；SHA256 `5c8113088355a817272fca55150c32467638b6f4a14c88f9ac688895085f0991`。测试 / 仅原轮次。原归档/原控制器的失败清单，未发现当前公共Go或现行SOP执行接线。

**职责**：Replay finite read-only integrity refusals against temporary synthetic inputs.  Failure inventory: modified/oversized/link archive, linked parent, redirected archive, enlarged limits, laundered RED or interrupted race result, invented paid authority, changed link metadata, tar link, wrong binary reference, changed helper and changed source identity. No captured command, key, provider or network runs。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定源码/工具/归档和私有篡改副本或合成fixture；支持模块随对应测试使用。参数：`parser.add_argument('--common-root', type=Path, default=BASE.parents[4])` L31。

**输出/写入**：该历史失败清单的断言/拒绝记录和临时日志；不续期真实资格。实际静态写点：L27 `(base / 'SHA256SUMS').write_text(''.join((hashlib.sha256(p.read_bytes()).hexdigest() + '  ' + p.relative_to(base).as_posix() + '\n' for p in files)))`；L43 `path.parent.mkdir(parents=True)`；L58 `shutil.copytree(BASE, copied)`；L59 `helper.mkdir(parents=True)`；L60 `shutil.copyfile(BASE.parents[1] / 'ticket-13/implementation/verify.py', helper / 'verify.py')`；L67 `path.write_text(json.dumps(value))`；L46 `path.symlink_to(a.common_root / archive['file'])`；L49 `path.parent.symlink_to((a.common_root / archive['file']).parent, target_is_directory=True)`；L71 `path.write_text(json.dumps(value))`；L85 `path.write_text(json.dumps(value))`；L79 `destination.parent.mkdir(parents=True)`。

**副作用**：运行相邻公共CLI或临时程序，写/删私有fixture；部分故障测试会临时替换再恢复输入，仅原环境适用，不在本次执行。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-14/gap2`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L22 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-14--gap2--snapshot-py"></a>

## evidence/ticket-14/gap2/snapshot.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-14/gap2/snapshot.py](../../barness-ai-minimax/evidence/ticket-14/gap2/snapshot.py)；SHA256 `5673a773e8b149456888075321caa5d928c61d77619b646f9e7e06103448b29c`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Read-only bounded snapshot of gap2 offline TDD and committed six checks。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L12 `BASE.mkdir(parents=True)`；L12 `EXT.mkdir()`；L68 `(BASE / 'inventory.json.gz').write_bytes(encoded)`；L69 `(BASE / 'source-at-start.json.gz').write_bytes(sencoded)`；L79 `dump(BASE / 'index.json', index)`；L79 `dump(BASE / 'external-references.json', dict(schema=1, codeCommit=CODE, checkedSourceCommit=SOURCE, binaries=references, retentionCredentialsRead=False, retentionPaidCalls=0))`；L81 `archive.chmod(292)`；L10 `p.write_text(json.dumps(v, ensure_ascii=False, separators=(',', ':')) + '\n')`；L80 `(BASE / leaf).write_bytes((TDD / leaf).read_bytes())`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-14/gap2`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L5 `ROOT=Path('/tmp/barness-minimax-routed-validation').resolve();COMMON=Path('/Users/cyber/RestoX/barness')`；L45 `add(Path('/tmp/barness-routed-postgate-gap-2.json'),'diagnosis/postgate-gap-2.json')`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path('/tmp/barness-minimax-routed-validation').resolve()`；L12 `BASE.mkdir(parents=True)`；L12 `EXT.mkdir()`；L33 `(ROOT / '.evidence/barness-ai').glob('*/manifest.json')`；L35 `json.loads(p.read_text())`。

<a id="barness-ai-minimax--evidence--ticket-14--gap2--verify-py"></a>

## evidence/ticket-14/gap2/verify.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-14/gap2/verify.py](../../barness-ai-minimax/evidence/ticket-14/gap2/verify.py)；SHA256 `857024d022cd9c8a71c838a3536ca8b6e27a68abc8159076cdef314240dc2a18`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Bounded read-only replay of gap2 fixture/diagnostic/preflight repair; no certificate。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`parser.add_argument('--common-root', type=Path, default=BASE.parents[4])` L93。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-14/gap2`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-14--repair--capsule-py"></a>

## evidence/ticket-14/repair/capsule.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-14/repair/capsule.py](../../barness-ai-minimax/evidence/ticket-14/repair/capsule.py)；SHA256 `651903be831bcff9ac172f2f8e1b3cbd5f469460760bfaa3af9af59a08c30f84`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Finite streaming reader for the ticket14 keyless repair history only.  This is a distinct retention protocol: the earlier implementation capsule has 32MiB/192MiB/32k limits and remains unchanged. Here two preflight attempts and native child fixtures require explicit fixed larger limits; no captured command is executed。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-14/repair`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--evidence--ticket-14--repair--check-boundaries-py"></a>

## evidence/ticket-14/repair/check-boundaries.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-14/repair/check-boundaries.py](../../barness-ai-minimax/evidence/ticket-14/repair/check-boundaries.py)；SHA256 `1ee8d3a024a52fb41e1300d6b03126549830cbe177612cd3a3750a8cea7b3ad7`。测试 / 仅原轮次。原归档/原控制器的失败清单，未发现当前公共Go或现行SOP执行接线。

**职责**：Replay finite read-only integrity refusals against temporary synthetic inputs.  Failure inventory: modified/oversized/link archive, linked parent, redirected archive, enlarged limits, laundered RED or missing-fixture result, invented paid authority, changed link metadata, tar link, wrong binary reference, changed helper and changed source identity. No captured command, key, provider or network runs。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定源码/工具/归档和私有篡改副本或合成fixture；支持模块随对应测试使用。参数：`parser.add_argument('--common-root', type=Path, default=BASE.parents[4])` L31。

**输出/写入**：该历史失败清单的断言/拒绝记录和临时日志；不续期真实资格。实际静态写点：L27 `(base / 'SHA256SUMS').write_text(''.join((hashlib.sha256(p.read_bytes()).hexdigest() + '  ' + p.relative_to(base).as_posix() + '\n' for p in files)))`；L43 `path.parent.mkdir(parents=True)`；L58 `shutil.copytree(BASE, copied)`；L59 `helper.mkdir(parents=True)`；L60 `shutil.copyfile(BASE.parents[1] / 'ticket-13/implementation/verify.py', helper / 'verify.py')`；L67 `path.write_text(json.dumps(value))`；L46 `path.symlink_to(a.common_root / archive['file'])`；L49 `path.parent.symlink_to((a.common_root / archive['file']).parent, target_is_directory=True)`；L71 `path.write_text(json.dumps(value))`；L85 `path.write_text(json.dumps(value))`；L79 `destination.parent.mkdir(parents=True)`。

**副作用**：运行相邻公共CLI或临时程序，写/删私有fixture；部分故障测试会临时替换再恢复输入，仅原环境适用，不在本次执行。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-14/repair`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L22 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-14--repair--snapshot-py"></a>

## evidence/ticket-14/repair/snapshot.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-14/repair/snapshot.py](../../barness-ai-minimax/evidence/ticket-14/repair/snapshot.py)；SHA256 `b2f8add23b13154e25e6a980cd31b7da470347a1b8adcab603192e2f1ab54873`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Read-only finite ticket14 history snapshot; never executes captured code。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L11 `BASE.mkdir(parents=True)`；L11 `EXT.mkdir()`；L72 `(BASE / 'inventory.json.gz').write_bytes(encoded)`；L73 `(BASE / 'source-at-start.json.gz').write_bytes(source_gz)`；L91 `dump(BASE / 'index.json', index)`；L91 `dump(BASE / 'external-references.json', dict(schema=1, codeCommit=SOURCE, binaries=references, unusedPaidPreparedBinaries=True, retentionCredentialsRead=False, retentionPaidCalls=0))`；L93 `archive.chmod(292)`；L9 `p.write_text(json.dumps(v, ensure_ascii=False, separators=(',', ':')) + '\n')`；L92 `(BASE / leaf).write_bytes((TDD / leaf).read_bytes())`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-14/repair`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L5 `ROOT=Path('/tmp/barness-minimax-routed-validation').resolve();COMMON=Path('/Users/cyber/RestoX/barness')`；L48 `add(Path('/tmp/barness-routed-release-plan-gap.json'),'diagnosis/release-plan-gap.json')`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path('/tmp/barness-minimax-routed-validation').resolve()`；L11 `BASE.mkdir(parents=True)`；L11 `EXT.mkdir()`；L34 `(ROOT / '.evidence/barness-ai').glob('*/manifest.json')`；L36 `json.loads(p.read_text())`。

<a id="barness-ai-minimax--evidence--ticket-14--repair--verify-py"></a>

## evidence/ticket-14/repair/verify.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-14/repair/verify.py](../../barness-ai-minimax/evidence/ticket-14/repair/verify.py)；SHA256 `66fbcc8b5f84b5bfcf0b97a457e3f09b5a81bdf9d4a6fa6e38f23fdf5ce8d6da`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Bounded, read-only replay of ticket14 keyless repair history; no release certificate。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`parser.add_argument('--common-root', type=Path, default=BASE.parents[4])` L173。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-14/repair`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L16 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-15--implementation--negative_guards-py"></a>

## evidence/ticket-15/implementation/negative_guards.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/implementation/negative_guards.py](../../barness-ai-minimax/evidence/ticket-15/implementation/negative_guards.py)；SHA256 `15a266adea8f8de44a919da3078870bb5e7a5572ad51768ab6ef7c2971b582c9`。测试 / 仅原轮次。原归档/原控制器的失败清单，未发现当前公共Go或现行SOP执行接线。

**职责**：Public verifier guard checks in private copies; immutable originals stay read-only。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定源码/工具/归档和私有篡改副本或合成fixture；支持模块随对应测试使用。参数：`p.add_argument('--common-root', type=Path, default=BASE.parents[4])` L59。

**输出/写入**：该历史失败清单的断言/拒绝记录和临时日志；不续期真实资格。实际静态写点：L11 `(base / 'SHA256SUMS').write_text(''.join((sha(p) + '  ' + p.name + '\n' for p in sorted(names))))`；L21 `base.parent.mkdir(parents=True)`；L21 `shutil.copytree(BASE, base)`；L22 `target.parent.mkdir(parents=True)`；L22 `shutil.copyfile(helper, target)`；L23 `archive.parent.mkdir(parents=True)`；L23 `shutil.copyfile(common / REL, archive)`；L51 `(base / 'index.json').write_text(json.dumps(index))`；L24 `(base / 'summary.json').write_text('{}')`；L27 `(base / 'summary.json').write_text(json.dumps(v))`；L29 `p.write_bytes(gzip.compress(raw))`；L35 `archive.unlink()`；L36 `archive.symlink_to(common / REL)`。

**副作用**：运行相邻公共CLI或临时程序，写/删私有fixture；部分故障测试会临时替换再恢复输入，仅原环境适用，不在本次执行。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/implementation`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`；L7 `hashlib.sha256(p.read_bytes()).hexdigest()`；L7 `p.read_bytes()`。

<a id="barness-ai-minimax--evidence--ticket-15--implementation--verify-py"></a>

## evidence/ticket-15/implementation/verify.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/implementation/verify.py](../../barness-ai-minimax/evidence/ticket-15/implementation/verify.py)；SHA256 `ed5e84ba40309476371925e4f44c948caf513dfe36ced88f95e5bbb989500981`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Read-only fixed six CLI TDD replay; executes no captured command or vendor request。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`p.add_argument('--common-root', type=Path, default=BASE.parents[4])` L57。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/implementation`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--assembly--check-boundaries-py"></a>

## evidence/ticket-15/new-actual/assembly/check-boundaries.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/assembly/check-boundaries.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/assembly/check-boundaries.py)；SHA256 `9fe86e2adb2f0965e815433b848dbbd9826ecbc4184c052cca8cc1db1f983150`。测试 / 仅原轮次。原归档/原控制器的失败清单，未发现当前公共Go或现行SOP执行接线。

**职责**：Public process failures specified before the failed-assembly consumer。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定源码/工具/归档和私有篡改副本或合成fixture；支持模块随对应测试使用。参数：`p.add_argument('--common-root', type=Path, default=BASE.parents[5])` L59。

**输出/写入**：该历史失败清单的断言/拒绝记录和临时日志；不续期真实资格。实际静态写点：L6 `(base / 'SHA256SUMS').write_text(''.join((sha(p) + '  ' + p.relative_to(base).as_posix() + '\n' for p in sorted(base.rglob('*')) if p.is_file() and p.name != 'SHA256SUMS')))`；L14 `base.parent.mkdir(parents=True)`；L14 `shutil.copytree(BASE, base)`；L16 `shutil.copytree(BASE.parents[1] / 'finalizer-design', target)`；L50 `(base / 'index.json').write_text(json.dumps(index))`；L15 `shutil.copytree(BASE.parent / label, base.parent / label)`；L18 `p.parent.mkdir(parents=True, exist_ok=True)`；L18 `shutil.copyfile(BASE.parents[2] / 'ticket-13' / rel, p)`；L21 `p.parent.mkdir(parents=True, exist_ok=True)`；L21 `shutil.copyfile(common / c['archive']['file'], p)`；L23 `(base / 'summary.json').write_text('{}')`；L25 `(base / 'summary.json').write_text(json.dumps(v))`；L29 `archive.unlink()`；L30 `archive.symlink_to(common / REL)`；L49 `p.write_bytes(gzip.compress(raw))`。

**副作用**：运行相邻公共CLI或临时程序，写/删私有fixture；部分故障测试会临时替换再恢复输入，仅原环境适用，不在本次执行。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/assembly`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`；L5 `hashlib.sha256(p.read_bytes()).hexdigest()`；L5 `p.read_bytes()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--assembly--freeze-py"></a>

## evidence/ticket-15/new-actual/assembly/freeze.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/assembly/freeze.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/assembly/freeze.py)；SHA256 `0978b8e117f0dd01be08ce8ca040afecfd4424631dd455603eb9f6c68db96abd`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Exclusive regular-only snapshot of the explicitly closed failed A branch。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L47 `(BASE / 'inventory.json.gz').write_bytes(encoded)`；L47 `(BASE / 'source-at-start.json.gz').write_bytes(se)`；L53 `(BASE / 'index.json').write_text(json.dumps(index, separators=(',', ':')) + '\n')`；L53 `archive.chmod(292)`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/assembly`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L6 `BASE=Path(__file__).resolve().parent;COMMON=Path('/Users/cyber/RestoX/barness');A=COMMON/'.evidence/minimax-release-repair-final-1';EXT=COMMON/'.evidence/minimax-retained/ticket-15-new-actual';SOURCE='7931b163f983a8b362374163b7949d31e68bbec`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L6 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--assembly--readers-py"></a>

## evidence/ticket-15/new-actual/assembly/readers.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/assembly/readers.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/assembly/readers.py)；SHA256 `2e338ce15ef8b372f4b14ea3cc0d027650b2a3d4bea4f84e44d9ffb94576983c`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Pin and reuse the two already reviewed read-only retention protocols unchanged。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/assembly`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L4 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--assembly--verify-py"></a>

## evidence/ticket-15/new-actual/assembly/verify.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/assembly/verify.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/assembly/verify.py)；SHA256 `ff42ebe756ce8923f29b9b07860140f342ec5e449bbcd75098bea28a7602f71d`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Read only the closed A failure; successful Jev derivation is not assembly。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`p.add_argument('--common-root', type=Path, default=BASE.parents[5])` L56。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/assembly`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L6 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--effects--additional-protocol-source-py-gz"></a>

## evidence/ticket-15/new-actual/effects/additional-protocol-source.py.gz

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects/additional-protocol-source.py.gz](../../barness-ai-minimax/evidence/ticket-15/new-actual/effects/additional-protocol-source.py.gz)；SHA256 `572f51cd62e9c927ad7aef7f599d19674ad97e1b76e462fb4d3ca5045a6da44a`。历史源码 / 封存源码快照。压缩封存源码，不是活动CLI。

**职责**：Independently read candidate or explicitly selected formal protocol structure.  No network, credentials, or working-source equality required for historical runs. Use --require-live to reject controlled HTTP packages. NOT_RUN/FAIL exit nonzero. Formal support additionally requires the public Classify replay and independent effect verification performed by the jevformal command; this is one stage only。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：只作为gzip封存的原Python源码字节；解压仅供静态阅读。参数：`parser.add_argument('--require-live', action='store_true')` L356；`parser.add_argument('--formal-catalog', type=Path)` L357；`parser.add_argument('bundle', type=Path)` L358。

**输出/写入**：原脚本定义的历史来源/协议核验或补证输出；本次仅登记压缩与解压SHA256。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：不得把源码快照当活动工具解压执行；原源码的CLI/IO/子进程静态点仍列出，不修改归档或重发历史结论。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

解压文本SHA256：`0ad4b561cfe0407d8ee2943aa6543d4c56ea0f0e895e772ff111d4699a78f08e`；保持压缩原件，不解压执行。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--effects--check-boundaries-py"></a>

## evidence/ticket-15/new-actual/effects/check-boundaries.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects/check-boundaries.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/effects/check-boundaries.py)；SHA256 `988a4847ae37caedddfd4f1f494810115401a42826b1e4e2e61891511ced64b3`。测试 / 仅原轮次。原归档/原控制器的失败清单，未发现当前公共Go或现行SOP执行接线。

**职责**：Public process custody negatives specified before the effects consumer。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定源码/工具/归档和私有篡改副本或合成fixture；支持模块随对应测试使用。参数：`p.add_argument('--common-root', type=Path, default=BASE.parents[5])` L65。

**输出/写入**：该历史失败清单的断言/拒绝记录和临时日志；不续期真实资格。实际静态写点：L8 `(base / 'SHA256SUMS').write_text(''.join((sha(p) + '  ' + p.relative_to(base).as_posix() + '\n' for p in sorted(base.rglob('*')) if p.is_file() and p.name != 'SHA256SUMS')))`；L19 `base.parent.mkdir(parents=True)`；L19 `shutil.copytree(BASE, base)`；L19 `shutil.copytree(BASE.parent / 'glm', base.parent / 'glm')`；L56 `(base / 'index.json').write_text(json.dumps(index))`；L21 `p.parent.mkdir(parents=True, exist_ok=True)`；L21 `shutil.copyfile(BASE.parents[2] / 'ticket-13' / rel, p)`；L24 `p.parent.mkdir(parents=True, exist_ok=True)`；L24 `shutil.copyfile(common / cap['archive']['file'], p)`；L26 `(base / 'summary.json').write_text('{}')`；L29 `(base / 'summary.json').write_text(json.dumps(v))`；L32 `archive.unlink()`；L33 `archive.symlink_to(common / REL)`；L37 `p.write_bytes(gzip.compress(b'x' * ((8 << 20) + 1)))`；L55 `p.write_bytes(gzip.compress(raw))`。

**副作用**：运行相邻公共CLI或临时程序，写/删私有fixture；部分故障测试会临时替换再恢复输入，仅原环境适用，不在本次执行。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`；L5 `hashlib.sha256(p.read_bytes()).hexdigest()`；L5 `p.read_bytes()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--effects--freeze-py"></a>

## evidence/ticket-15/new-actual/effects/freeze.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects/freeze.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/effects/freeze.py)；SHA256 `6fa4548fce1f1e1387372af2d9db6c48975f15fcdc248c625b83e630bc085e70`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Exclusive snapshot of closed effects. Distinct fixed custody, no producer runs.  The two reviewed archive protocols remain unchanged; snapshot layout differs from GLM because protocol/visual bodies and a shared executable have own pins。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L57 `(BASE / leaf).write_bytes(encoded)`；L57 `(BASE / 'source-at-start.json.gz').write_bytes(se)`；L63 `archive.chmod(292)`；L99 `(BASE / 'index.json').write_text(json.dumps(value, separators=(',', ':')) + '\n')`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L10 `BASE=Path(__file__).resolve().parent;COMMON=Path('/Users/cyber/RestoX/barness');E=COMMON/'.evidence/minimax-release-repair-effects-1';EXT=COMMON/'.evidence/minimax-retained/ticket-15-new-actual'`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L10 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--effects--provenance-py"></a>

## evidence/ticket-15/new-actual/effects/provenance.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects/provenance.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/effects/provenance.py)；SHA256 `46634c5c644937dfd9eadc255d94847b5163ade4deef1d3311f639f0a995fd78`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Independent custody checks for one already completed, explicitly approved effects phase。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--effects--quality-py"></a>

## evidence/ticket-15/new-actual/effects/quality.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects/quality.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/effects/quality.py)；SHA256 `2cbadf5ff28ff4234f1b4e1fcad7069af9c633be18c6144a4c238ac46017ccb2`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Independent finite recomputation of the fixed twelve synthetic visual samples.  This historical frontdoor shares no mutable evaluation state with the producer. Only observed native gold/result bodies are compared; these are not population or billing estimates. Full original metrics remain in the immutable report。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--effects--readers-py"></a>

## evidence/ticket-15/new-actual/effects/readers.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects/readers.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/effects/readers.py)；SHA256 `2e338ce15ef8b372f4b14ea3cc0d027650b2a3d4bea4f84e44d9ffb94576983c`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Pin and reuse the two already reviewed read-only retention protocols unchanged。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L4 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--effects--tdd--analyse-controller-py"></a>

## evidence/ticket-15/new-actual/effects/tdd/analyse-controller.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects/tdd/analyse-controller.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/effects/tdd/analyse-controller.py)；SHA256 `056b9f524c5e61ececaaef726dc4fb3f4c570f8b3ca7a622fd6d9bf1f9d12b7a`。诊断装置 / 仅原轮次。原缺陷/资源/文件系统诊断；依赖原来源与私有fixture，不是产品资格入口。

**职责**：analyse-controller.py：原缺陷/资源/文件系统诊断；依赖原来源与私有fixture，不是产品资格入口。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定失败模型及私有源/索引/文件系统副本，部分构造器/probe包含合成外部程序。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：私有fixture、探针日志与有限断言；如处在raw目录则仅为封存源码。实际静态写点：L6 `(p / 'scenarios.jsonl').write_text(''.join((json.dumps(x, ensure_ascii=False, separators=(',', ':')) + '\n' for x in sc)))`；L5 `(p / name).write_text(json.dumps(value, ensure_ascii=False, separators=(',', ':')) + '\n')`。

**副作用**：可能启动公共CLI/构造器、修改/删除私有文件或建立FIFO；只用于原问题的诊断，不是通用维护入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects/tdd`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L3 `p=Path('/tmp/barness-minimax-routed-validation/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects').resolve();sys.path.insert(0,str(p));import verify as v`；L4 `b=v.readers.small();i=b.doc((p/'index.json').read_bytes());a,m=b.verify_archive(Path('/Users/cyber/RestoX/barness'),i['capsule']);s,sc=v.analyse(b,i,m);more=v.secondary(b,i,Path('/Users/cyber/RestoX/barness'),m)`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L3 `Path('/tmp/barness-minimax-routed-validation/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects').resolve()`；L4 `(p / 'index.json').read_bytes()`；L5 `(p / name).write_text(json.dumps(value, ensure_ascii=False, separators=(',', ':')) + '\n')`；L5 `json.dumps(value, ensure_ascii=False, separators=(',', ':'))`；L6 `(p / 'scenarios.jsonl').write_text(''.join((json.dumps(x, ensure_ascii=False, separators=(',', ':')) + '\n' for x in sc)))`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--effects--tdd--red-original-regular-scan-py"></a>

## evidence/ticket-15/new-actual/effects/tdd/red-original-regular-scan.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects/tdd/red-original-regular-scan.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/effects/tdd/red-original-regular-scan.py)；SHA256 `4d08d04410360954a962edddfdaa52dd2a1aeea47819633bcd51aa0c1b383303`。诊断装置 / 仅原轮次。原缺陷/资源/文件系统诊断；依赖原来源与私有fixture，不是产品资格入口。

**职责**：red-original-regular-scan.py：原缺陷/资源/文件系统诊断；依赖原来源与私有fixture，不是产品资格入口。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定失败模型及私有源/索引/文件系统副本，部分构造器/probe包含合成外部程序。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：私有fixture、探针日志与有限断言；如处在raw目录则仅为封存源码。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：可能启动公共CLI/构造器、修改/删除私有文件或建立FIFO；只用于原问题的诊断，不是通用维护入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects/tdd`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L3 `root=Path('/Users/cyber/RestoX/barness/.evidence/minimax-release-repair-effects-1')`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L14 `stat.S_ISREG(row.stat(follow_symlinks=False).st_mode)`；L14 `row.stat(follow_symlinks=False)`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--effects--verify-py"></a>

## evidence/ticket-15/new-actual/effects/verify.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects/verify.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/effects/verify.py)；SHA256 `04ee2c8351996297c058c99befca7e82ce7b4ac9d2ccf18bc5f73054fc7a4b10`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Read only closed protocol/visual actuals; whole release remains uncertified。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`parser.add_argument('--common-root', type=Path, default=BASE.parents[5])` L115。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L6 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--ordinary--boolean-receipt-probe-py"></a>

## evidence/ticket-15/new-actual/final-gate/ordinary/boolean-receipt-probe.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/boolean-receipt-probe.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/boolean-receipt-probe.py)；SHA256 `58f99f58927da30a0589d9ba0fb1b54efe05d613f7e22b3dd9fe175dcf7c43ed`。诊断装置 / 仅原轮次。原单项边界探针，不能泛化为支持/发布通过。

**职责**：One finite private full-archive fixture for JSON true versus numeric exit1。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定失败模型及私有源/索引/文件系统副本，部分构造器/probe包含合成外部程序。参数：`p.add_argument('--common-root', type=Path, required=True)` L62；`p.add_argument('--work-output', type=Path, required=True)` L62；`p.add_argument('--mode', choices=('prepare', 'check'), required=True)` L62；`p.add_argument('--kind', choices=('boolean', 'float'), required=True)` L62。

**输出/写入**：私有fixture、探针日志与有限断言；如处在raw目录则仅为封存源码。实际静态写点：L7 `(base / 'SHA256SUMS').write_text(''.join((sha(p.read_bytes()) + '  ' + p.relative_to(base).as_posix() + '\n' for p in sorted(base.rglob('*')) if p.is_file() and p.name != 'SHA256SUMS')))`；L10 `root.mkdir()`；L10 `base.parent.mkdir(parents=True)`；L10 `shutil.copytree(BASE, base)`；L15 `archive.parent.mkdir(parents=True)`；L48 `(base / 'inventory.json.gz').write_bytes(encoded)`；L48 `(base / 'index.json').write_text(json.dumps(index) + '\n')`；L51 `(base / 'summary.json').write_text(json.dumps(summary) + '\n')`；L51 `archive.chmod(292)`；L51 `(root / 'fixture-proof.json').write_text(json.dumps(proof) + '\n')`；L59 `(root / (mode + '-public.log')).write_text(result.stdout + result.stderr)`；L59 `(root / (mode + '-public-command.json')).write_text(json.dumps(receipt) + '\n')`；L12 `p.parent.mkdir(parents=True, exist_ok=True)`；L12 `shutil.copyfile(BASE.parents[3] / 'ticket-13' / rel, p)`；L58 `shutil.copyfile(BASE / 'verify.py', base / 'verify.py')`。

**副作用**：可能启动公共CLI/构造器、修改/删除私有文件或建立FIFO；只用于原问题的诊断，不是通用维护入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L6 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--ordinary--check-boundaries-py"></a>

## evidence/ticket-15/new-actual/final-gate/ordinary/check-boundaries.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/check-boundaries.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/check-boundaries.py)；SHA256 `4ecd88ee911f0d7af79a58faa163b746df967b6a3ffd52817e36a2f594a948c8`。测试 / 仅原轮次。原归档/原控制器的失败清单，未发现当前公共Go或现行SOP执行接线。

**职责**：Small private-index negatives over the unchanged full regular archive。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定源码/工具/归档和私有篡改副本或合成fixture；支持模块随对应测试使用。参数：`p.add_argument('--common-root', type=Path, default=BASE.parents[6])` L46。

**输出/写入**：该历史失败清单的断言/拒绝记录和临时日志；不续期真实资格。实际静态写点：L6 `(base / 'SHA256SUMS').write_text(''.join((sha(p) + '  ' + p.relative_to(base).as_posix() + '\n' for p in sorted(base.rglob('*')) if p.is_file() and p.name != 'SHA256SUMS')))`；L14 `base.parent.mkdir(parents=True)`；L14 `shutil.copytree(BASE, base)`；L17 `archive.parent.mkdir(parents=True)`；L29 `(base / 'index.json').write_text(json.dumps(index))`；L16 `p.parent.mkdir(parents=True, exist_ok=True)`；L16 `shutil.copyfile(BASE.parents[3] / 'ticket-13' / rel, p)`；L18 `(base / 'summary.json').write_text('{}')`；L28 `(base / 'summary.json').write_text(json.dumps(s))`。

**副作用**：运行相邻公共CLI或临时程序，写/删私有fixture；部分故障测试会临时替换再恢复输入，仅原环境适用，不在本次执行。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`；L5 `hashlib.sha256(p.read_bytes()).hexdigest()`；L5 `p.read_bytes()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--ordinary--check-probe-bounds-py"></a>

## evidence/ticket-15/new-actual/final-gate/ordinary/check-probe-bounds.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/check-probe-bounds.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/check-probe-bounds.py)；SHA256 `51bb941926f7501c8270dedef046f746992c2580ee2915006ff32d6c4a1e6c4e`。测试 / 仅原轮次。原归档/原控制器的失败清单，未发现当前公共Go或现行SOP执行接线。

**职责**：Exercise the public fixture creator on finite local untrusted archive inputs。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定源码/工具/归档和私有篡改副本或合成fixture；支持模块随对应测试使用。参数：`p.add_argument('--output', type=Path, required=True)` L26。

**输出/写入**：该历史失败清单的断言/拒绝记录和临时日志；不续期真实资格。实际静态写点：L8 `output.mkdir(parents=True, exist_ok=False)`；L9 `(output / 'creator-before.py.txt').write_bytes(raw)`；L23 `(output / 'results.json').write_text(json.dumps(result, indent=2) + '\n')`；L11 `archive.parent.mkdir(parents=True)`；L18 `(root / 'public.log').write_bytes(run.stdout + run.stderr)`；L15 `archive.symlink_to('missing-target')`。

**副作用**：运行相邻公共CLI或临时程序，写/删私有fixture；部分故障测试会临时替换再恢复输入，仅原环境适用，不在本次执行。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--ordinary--field-repair--verify-negative-fixtures-py"></a>

## evidence/ticket-15/new-actual/final-gate/ordinary/field-repair/verify-negative-fixtures.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/field-repair/verify-negative-fixtures.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/field-repair/verify-negative-fixtures.py)；SHA256 `9d7faaecdcf5dcd5b27fae680020342443f75953c6c8960440cb4b3e199ef998`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Read-only exact negative fixture custody; never a release qualification。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`p.add_argument('--common-root', type=Path, required=True)` L6。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/field-repair`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`；L6 `p.parse_args()`；L6 `(BASE / 'index.json').read_bytes()`；L8 `(BASE / 'field-repair' / (kind + '-negative-index.json')).read_bytes()`；L15 `json.dumps(dict(verified=True, negativeFixtures=rows, originalActualExitType='int', originalActualExit=1, credentialsRead=False, paidCalls=0, releaseCertified=False), indent=2)`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--ordinary--freeze-py"></a>

## evidence/ticket-15/new-actual/final-gate/ordinary/freeze.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/freeze.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/freeze.py)；SHA256 `dea0b0d17cb49b89f73d351956103e5b4c95d84706bded6efb9f49fb0a30db36`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Freeze only the explicitly closed ordinary subtree, receipt and immutable plan。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L48 `(BASE / 'inventory.json.gz').write_bytes(encoded)`；L48 `(BASE / 'source-at-start.json.gz').write_bytes(se)`；L55 `(BASE / 'index.json').write_text(json.dumps(cap, separators=(',', ':')) + '\n')`；L55 `archive.chmod(292)`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L6 `BASE=Path(__file__).resolve().parent;ROOT=Path('/Users/cyber/RestoX/barness');B=ROOT/'.evidence/minimax-release-repair-incomplete-gate-1';SESSION='20261018T120404.152055000Z';ORD=B/'gate/offline'/SESSION;EXT=ROOT/'.evidence/minimax-retained`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L6 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--ordinary--readers-py"></a>

## evidence/ticket-15/new-actual/final-gate/ordinary/readers.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/readers.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/readers.py)；SHA256 `b3bec7f74b5378096bb735b0d515ffd1d64bc0063ffd2506d609de6f9e30dd4a`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Pin and reuse the two already reviewed read-only retention protocols unchanged。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L4 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--ordinary--tdd--verify-sealed-bytes-py"></a>

## evidence/ticket-15/new-actual/final-gate/ordinary/tdd/verify-sealed-bytes.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/tdd/verify-sealed-bytes.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/tdd/verify-sealed-bytes.py)；SHA256 `755ca3c8a6e719c8a5f3288b9dc7fdf0ddbb13693e940765d1ff39a94a019f67`。诊断装置 / 仅原轮次。原缺陷/资源/文件系统诊断；依赖原来源与私有fixture，不是产品资格入口。

**职责**：verify-sealed-bytes.py：原缺陷/资源/文件系统诊断；依赖原来源与私有fixture，不是产品资格入口。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定失败模型及私有源/索引/文件系统副本，部分构造器/probe包含合成外部程序。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：私有fixture、探针日志与有限断言；如处在raw目录则仅为封存源码。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：可能启动公共CLI/构造器、修改/删除私有文件或建立FIFO；只用于原问题的诊断，不是通用维护入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/tdd`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L4 `b=readers.small();index=b.doc((base/'index.json').read_bytes());pins,m,archive=readers.full().verify(b,base,Path('/Users/cyber/RestoX/barness'),index);print(json.dumps(dict(verified=True,archive=archive,sourceCommit=index['sourceCommit'],cr`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L3 `Path(__file__).resolve()`；L4 `(base / 'index.json').read_bytes()`；L4 `json.dumps(dict(verified=True, archive=archive, sourceCommit=index['sourceCommit'], credentialsRead=False, paidCalls=0), indent=2)`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--ordinary--verify-py"></a>

## evidence/ticket-15/new-actual/final-gate/ordinary/verify.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/verify.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary/verify.py)；SHA256 `e8fa51864d50d61fa27359d84cd04e85b1967ffb6110cb22aa363e10ced754a3`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Verify only the completed ordinary command; race/outer remain separate。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`p.add_argument('--common-root', type=Path, default=BASE.parents[6])` L30。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L6 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--race--check-boundaries-py"></a>

## evidence/ticket-15/new-actual/final-gate/race/check-boundaries.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/race/check-boundaries.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/race/check-boundaries.py)；SHA256 `47eac8ebf4b3d512c243112c997ae2a3043631bebf97fe1fa309a2b52930e185`。测试 / 仅原轮次。原归档/原控制器的失败清单，未发现当前公共Go或现行SOP执行接线。

**职责**：Small private-index negatives over the unchanged full regular archive。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定源码/工具/归档和私有篡改副本或合成fixture；支持模块随对应测试使用。参数：`p.add_argument('--common-root', type=Path, default=BASE.parents[6])` L38。

**输出/写入**：该历史失败清单的断言/拒绝记录和临时日志；不续期真实资格。实际静态写点：L6 `(base / 'SHA256SUMS').write_text(''.join((sha(p) + '  ' + p.relative_to(base).as_posix() + '\n' for p in sorted(base.rglob('*')) if p.is_file() and p.name != 'SHA256SUMS')))`；L14 `base.parent.mkdir(parents=True)`；L14 `shutil.copytree(BASE, base)`；L17 `archive.parent.mkdir(parents=True)`；L29 `(base / 'index.json').write_text(json.dumps(index))`；L16 `p.parent.mkdir(parents=True, exist_ok=True)`；L16 `shutil.copyfile(BASE.parents[3] / 'ticket-13' / rel, p)`；L18 `(base / 'summary.json').write_text('{}')`；L28 `(base / 'summary.json').write_text(json.dumps(s))`。

**副作用**：运行相邻公共CLI或临时程序，写/删私有fixture；部分故障测试会临时替换再恢复输入，仅原环境适用，不在本次执行。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/race`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`；L5 `hashlib.sha256(p.read_bytes()).hexdigest()`；L5 `p.read_bytes()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--race--freeze-py"></a>

## evidence/ticket-15/new-actual/final-gate/race/freeze.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/race/freeze.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/race/freeze.py)；SHA256 `35bfd4349eb5f4003ed683660f321e261ce24fccd39d9c4ffd60478f893553d9`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Freeze only the explicitly closed race subtree, receipt and immutable plan。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：`p.add_argument('--common-root', type=Path, default=ROOT)` L72。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L62 `(BASE / 'inventory.json.gz').write_bytes(encoded)`；L62 `(BASE / 'source-at-start.json.gz').write_bytes(se)`；L70 `(BASE / 'index.json').write_text(json.dumps(cap, separators=(',', ':')) + '\n')`；L70 `archive.chmod(292)`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/race`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L6 `BASE=Path(__file__).resolve().parent;ROOT=Path('/Users/cyber/RestoX/barness');B=ROOT/'.evidence/minimax-release-repair-incomplete-gate-1';SESSION='20261018T121452.571280000Z';ORD=B/'gate/race'/SESSION;EXT=ROOT/'.evidence/minimax-retained/ti`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L6 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--race--readers-py"></a>

## evidence/ticket-15/new-actual/final-gate/race/readers.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/race/readers.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/race/readers.py)；SHA256 `b3bec7f74b5378096bb735b0d515ffd1d64bc0063ffd2506d609de6f9e30dd4a`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Pin and reuse the two already reviewed read-only retention protocols unchanged。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/race`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L4 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--race--verify-py"></a>

## evidence/ticket-15/new-actual/final-gate/race/verify.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/race/verify.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/race/verify.py)；SHA256 `24698beea29fce131a26a4c8ab6fa308709f185611210065744cf1f014feae4a`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Verify only the completed race command; ordinary/outer remain separate。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`p.add_argument('--common-root', type=Path, default=BASE.parents[6])` L30。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/race`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L6 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--verdict--check-boundaries-py"></a>

## evidence/ticket-15/new-actual/final-gate/verdict/check-boundaries.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/verdict/check-boundaries.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/verdict/check-boundaries.py)；SHA256 `f2819de4eaac0818ad702ff261bd8683eea1c32d8c1ccb24db493633520cedbd`。测试 / 仅原轮次。原归档/原控制器的失败清单，未发现当前公共Go或现行SOP执行接线。

**职责**：Small private-index negatives over the unchanged full regular archive。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定源码/工具/归档和私有篡改副本或合成fixture；支持模块随对应测试使用。参数：`p.add_argument('--common-root', type=Path, default=BASE.parents[6])` L41。

**输出/写入**：该历史失败清单的断言/拒绝记录和临时日志；不续期真实资格。实际静态写点：L6 `(base / 'SHA256SUMS').write_text(''.join((sha(p) + '  ' + p.relative_to(base).as_posix() + '\n' for p in sorted(base.rglob('*')) if p.is_file() and p.name != 'SHA256SUMS')))`；L14 `base.parent.mkdir(parents=True)`；L14 `shutil.copytree(BASE, base)`；L17 `archive.parent.mkdir(parents=True)`；L32 `(base / 'index.json').write_text(json.dumps(index))`；L16 `p.parent.mkdir(parents=True, exist_ok=True)`；L16 `shutil.copyfile(BASE.parents[3] / 'ticket-13' / rel, p)`；L19 `shutil.copytree(BASE.parent / label, base.parent / label)`；L21 `(base / 'summary.json').write_text('{}')`；L31 `(base / 'summary.json').write_text(json.dumps(s))`。

**副作用**：运行相邻公共CLI或临时程序，写/删私有fixture；部分故障测试会临时替换再恢复输入，仅原环境适用，不在本次执行。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/verdict`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`；L5 `hashlib.sha256(p.read_bytes()).hexdigest()`；L5 `p.read_bytes()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--verdict--freeze-py"></a>

## evidence/ticket-15/new-actual/final-gate/verdict/freeze.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/verdict/freeze.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/verdict/freeze.py)；SHA256 `150ed15b0df08fd356d4a4b53af42c9bdafb3815ffb81eb60e28e6dd353623e3`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Exclusively freeze all closed outer B originals without duplicating suite bulk。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：`p.add_argument('--common-root', type=Path, default=ROOT)` L66。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L59 `(BASE / 'inventory.json.gz').write_bytes(encoded)`；L59 `(BASE / 'source-at-start.json.gz').write_bytes(se)`；L64 `(BASE / 'index.json').write_text(json.dumps(index, separators=(',', ':')) + '\n')`；L64 `archive.chmod(292)`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/verdict`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L6 `BASE=Path(__file__).resolve().parent;ROOT=Path('/Users/cyber/RestoX/barness');B=ROOT/'.evidence/minimax-release-repair-incomplete-gate-1';EXT=ROOT/'.evidence/minimax-retained/ticket-15-new-actual';SOURCE='7931b163f983a8b362374163b7949d31e68`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L6 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--verdict--readers-py"></a>

## evidence/ticket-15/new-actual/final-gate/verdict/readers.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/verdict/readers.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/verdict/readers.py)；SHA256 `b3bec7f74b5378096bb735b0d515ffd1d64bc0063ffd2506d609de6f9e30dd4a`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Pin and reuse the two already reviewed read-only retention protocols unchanged。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/verdict`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L4 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--final-gate--verdict--verify-py"></a>

## evidence/ticket-15/new-actual/final-gate/verdict/verify.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/verdict/verify.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/verdict/verify.py)；SHA256 `0907fffddaff3bb209ebecdeb9550bad001ce89b1f4aa5273b14853145968d74`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Stream the complete historical gate custody, preserving its actual false result。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`p.add_argument('--common-root', type=Path, required=True)` L64。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/verdict`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L6 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--glm--check-boundaries-py"></a>

## evidence/ticket-15/new-actual/glm/check-boundaries.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/glm/check-boundaries.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/glm/check-boundaries.py)；SHA256 `271ffdc67666b74d04fc783960621e9c0d308659c728ee7344541d0cba8055c2`。测试 / 仅原轮次。原归档/原控制器的失败清单，未发现当前公共Go或现行SOP执行接线。

**职责**：Failure-first public process probes; no captured command or actual Key is read。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定源码/工具/归档和私有篡改副本或合成fixture；支持模块随对应测试使用。参数：`p.add_argument('--common-root', type=Path, default=BASE.parents[5])` L83。

**输出/写入**：该历史失败清单的断言/拒绝记录和临时日志；不续期真实资格。实际静态写点：L10 `(base / 'SHA256SUMS').write_text(''.join((sha(p) + '  ' + p.name + '\n' for p in sorted(base.iterdir()) if p.is_file() and p.name != 'SHA256SUMS')))`；L24 `base.parent.mkdir(parents=True)`；L24 `shutil.copytree(BASE, base)`；L27 `archive.parent.mkdir(parents=True)`；L27 `shutil.copyfile(common / REL, archive)`；L72 `(base / 'index.json').write_text(json.dumps(index))`；L26 `target.parent.mkdir(parents=True, exist_ok=True)`；L26 `shutil.copyfile(source, target)`；L28 `(base / 'summary.json').write_text('{}')`；L34 `(base / 'summary.json').write_text(json.dumps(v))`；L36 `p.write_bytes(gzip.compress(b'x' * ((8 << 20) + 1)))`；L42 `archive.unlink()`；L43 `archive.symlink_to(common / REL)`；L62 `p.write_bytes(gzip.compress(raw))`。

**副作用**：运行相邻公共CLI或临时程序，写/删私有fixture；部分故障测试会临时替换再恢复输入，仅原环境适用，不在本次执行。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/glm`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`；L7 `hashlib.sha256(p.read_bytes()).hexdigest()`；L7 `p.read_bytes()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--glm--complete-snapshot-py"></a>

## evidence/ticket-15/new-actual/glm/complete-snapshot.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/glm/complete-snapshot.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/glm/complete-snapshot.py)；SHA256 `527f43efd7a4874747acfa1687cd8a1533ab2e3b42769380b280a9912a43a78f`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Complete metadata for immutable bytes after observation selection was refused。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L19 `archive.chmod(292)`；L35 `f.dump(BASE / 'index.json', index)`；L36 `f.dump(BASE / 'snapshot-authoring-history.json', dict(initialPublicVerifier='missing entrypoint RED before implementation', initialScan='unexpected intentional fixture symlink refused; reproduced numeric1 before exact metadata-only projection', firstSnapshot='paid archive closed; preflight regular bytes/inventory/post-hashes complete, observation c`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/glm`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--glm--freeze-py"></a>

## evidence/ticket-15/new-actual/glm/freeze.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/glm/freeze.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/glm/freeze.py)；SHA256 `ee60fc38a0e656727918d3e8b92b0fdd81285bc7600e40463e40740a08afa3ba`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Snapshot only closed G and its refused publisher; no producer command runs。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L11 `path.write_text(json.dumps(value, separators=(',', ':')) + '\n')`；L58 `(BASE / file).write_bytes(encoded)`；L59 `(BASE / 'source-at-start.json.gz').write_bytes(sencoded)`；L65 `archive.chmod(292)`；L72 `EXT.mkdir()`；L99 `(EXT / 'glm-actual-raw.tar.gz').rename(EXT / 'glm-raw.tar.gz')`；L109 `dump(BASE / 'index.json', index)`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/glm`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L6 `BASE=Path(__file__).resolve().parent;OWN=BASE.parents[5];COMMON=Path('/Users/cyber/RestoX/barness');G=COMMON/'.evidence/minimax-release-repair-glm-1';PUBLICATION=COMMON/'.evidence/minimax-release-repair-glm-publication-1'`；L96 `inline['references/approved-retention-plan.md']=Path('/tmp/barness-repair-actual-retention-plan.md').read_bytes();inline['references/future-finalizer-design-index.json']=(OWN/'.scratch/barness-ai-minimax/evidence/ticket-15/finalizer-design/`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L6 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--glm--provenance-py"></a>

## evidence/ticket-15/new-actual/glm/provenance.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/glm/provenance.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/glm/provenance.py)；SHA256 `a71b0b535bc79ed24ece99738f850c57ce7a8f8fdfc4d9f6391cf4223af6fef2`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Independent custody checks for one already completed, explicitly approved G phase。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/glm`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--glm--readers-py"></a>

## evidence/ticket-15/new-actual/glm/readers.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/glm/readers.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/glm/readers.py)；SHA256 `2e338ce15ef8b372f4b14ea3cc0d027650b2a3d4bea4f84e44d9ffb94576983c`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Pin and reuse the two already reviewed read-only retention protocols unchanged。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/glm`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L4 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-15--new-actual--glm--verify-py"></a>

## evidence/ticket-15/new-actual/glm/verify.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/glm/verify.py](../../barness-ai-minimax/evidence/ticket-15/new-actual/glm/verify.py)；SHA256 `20bcd7c821bd7a855c929865923ca4d7e74203e33d75c9bb7e5c42bf4f94cf0f`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Read-only actual G custody/accounting; the original Messages failure stays failed。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`parser.add_argument('--common-root', type=Path, default=BASE.parents[5])` L80。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/glm`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L6 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-16--deadline-alignment--check-boundaries-py"></a>

## evidence/ticket-16/deadline-alignment/check-boundaries.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-16/deadline-alignment/check-boundaries.py](../../barness-ai-minimax/evidence/ticket-16/deadline-alignment/check-boundaries.py)；SHA256 `49c50b1dc49fd3d449b6e5e160a3a0eea1342fea80a7c82442de1586ca179aa6`。测试 / 仅原轮次。原归档/原控制器的失败清单，未发现当前公共Go或现行SOP执行接线。

**职责**：Finite public integrity refusals: corrupt/oversize/linked archive or parent, forged local91s outcome/paid authority/GLM expiry, redirect/source/helper, or tar link. All mutations are temporary; no captured test, key or vendor call is executed。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定源码/工具/归档和私有篡改副本或合成fixture；支持模块随对应测试使用。参数：`parser.add_argument('--common-root', type=Path, default=BASE.parents[4])` L14。

**输出/写入**：该历史失败清单的断言/拒绝记录和临时日志；不续期真实资格。实际静态写点：L11 `(base / 'SHA256SUMS').write_text(''.join((hashlib.sha256(p.read_bytes()).hexdigest() + '  ' + p.name + '\n' for p in sorted(base.iterdir()) if p.is_file() and p.name != 'SHA256SUMS')))`；L25 `p.parent.mkdir(parents=True)`；L35 `shutil.copytree(BASE, copy)`；L35 `helper.mkdir(parents=True)`；L35 `shutil.copyfile(BASE.parents[1] / 'ticket-13/implementation/verify.py', helper / 'verify.py')`；L41 `p.write_text(json.dumps(v))`；L30 `p.symlink_to(a.common_root / archive['file'])`；L31 `p.parent.symlink_to((a.common_root / archive['file']).parent, target_is_directory=True)`；L44 `(copy / 'M3-selected-README.md').write_text('changed')`；L56 `p.write_text(json.dumps(v))`；L50 `target.parent.mkdir(parents=True)`。

**副作用**：运行相邻公共CLI或临时程序，写/删私有fixture；部分故障测试会临时替换再恢复输入，仅原环境适用，不在本次执行。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-16/deadline-alignment`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L8 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-16--deadline-alignment--snapshot-py"></a>

## evidence/ticket-16/deadline-alignment/snapshot.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-16/deadline-alignment/snapshot.py](../../barness-ai-minimax/evidence/ticket-16/deadline-alignment/snapshot.py)；SHA256 `8a5c7e41ee680dd4dbdba1e456fb5edcf8edea3c3e851e6bb10a714bc44ae1b7`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Bounded immutable localTLS TDD snapshot only; does not execute captures。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L9 `BASE.mkdir(parents=True)`；L9 `EXT.mkdir()`；L35 `(BASE / 'inventory.json.gz').write_bytes(encoded)`；L36 `(BASE / 'source-at-start.json.gz').write_bytes(sencoded)`；L38 `dump(BASE / 'index.json', dict(schema=1, kind='local-tls-glm-header-policy-alignment', codeCommit=CODE, boundedHelperSHA256='b280863f49f4a56cb507ee7e6ba753c73e38c9162236592e9f886ed9a64c3c2f', capsule=capsule, originalsUnchangedAfterSnapshot=True, retentionCredentialsRead=False, retentionPaidCalls=0))`；L40 `archive.chmod(292)`；L7 `p.write_text(json.dumps(v, separators=(',', ':')) + '\n')`；L39 `(BASE / f).write_bytes((A / f).read_bytes())`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-16/deadline-alignment`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L4 `ROOT=Path('/tmp/barness-minimax-routed-validation').resolve();COMMON=Path('/Users/cyber/RestoX/barness');A=ROOT/'.evidence/minimax-glm-host-deadline-1'`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L4 `Path('/tmp/barness-minimax-routed-validation').resolve()`；L9 `BASE.mkdir(parents=True)`；L9 `EXT.mkdir()`；L20 `(COMMON / '.evidence/minimax-global-formal-glm' / ('mm-global-formal-m3-' + route)).glob('*/minimax-report.json')`；L21 `(COMMON / '.evidence/minimax-routed-glm-1' / ('glm-' + route)).glob('*/glm-execution-source.json')`。

<a id="barness-ai-minimax--evidence--ticket-16--deadline-alignment--verify-py"></a>

## evidence/ticket-16/deadline-alignment/verify.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-16/deadline-alignment/verify.py](../../barness-ai-minimax/evidence/ticket-16/deadline-alignment/verify.py)；SHA256 `f0eae33f0ac195fdfc27b1bc9f6df7f79b2bbcb344b6bf8821918f6d8c760f9c`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Bounded localTLS deadline evidence replay; never executes captured tests or calls vendors。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`p.add_argument('--common-root', type=Path, default=BASE.parents[4])` L67。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-16/deadline-alignment`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-17--check-boundaries-py"></a>

## evidence/ticket-17/check-boundaries.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-17/check-boundaries.py](../../barness-ai-minimax/evidence/ticket-17/check-boundaries.py)；SHA256 `1b6f38f12b1e0d0991291e1a51bf5a5b41708af6df336c31ec9ac7b123d7dfaa`。测试 / 仅原轮次。原归档/原控制器的失败清单，未发现当前公共Go或现行SOP执行接线。

**职责**：Public verifier failures on private metadata copies; never edit the raw archive。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定源码/工具/归档和私有篡改副本或合成fixture；支持模块随对应测试使用。参数：`p.add_argument('--common-root', type=Path, required=True)` L42。

**输出/写入**：该历史失败清单的断言/拒绝记录和临时日志；不续期真实资格。实际静态写点：L7 `(base / 'SHA256SUMS').write_text(''.join((hashlib.sha256(p.read_bytes()).hexdigest() + '  ' + p.relative_to(base).as_posix() + '\n' for p in sorted(base.rglob('*')) if p.is_file() and p.name != 'SHA256SUMS')))`；L18 `base.parent.mkdir(parents=True)`；L18 `shutil.copytree(BASE, base)`；L34 `(base / 'index.json').write_text(json.dumps(index) + '\n')`；L20 `target.parent.mkdir(parents=True, exist_ok=True)`；L20 `shutil.copyfile(BASE.parent / 'ticket-13' / rel, target)`；L35 `(base / 'summary.json').write_text('{}\n')`；L28 `(base / 'summary.json').write_text(json.dumps(summary) + '\n')`；L31 `path.parent.mkdir(parents=True)`；L32 `path.symlink_to(common / index['archive']['file'])`。

**副作用**：运行相邻公共CLI或临时程序，写/删私有fixture；部分故障测试会临时替换再恢复输入，仅原环境适用，不在本次执行。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-17`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-17--contract-py"></a>

## evidence/ticket-17/contract.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-17/contract.py](../../barness-ai-minimax/evidence/ticket-17/contract.py)；SHA256 `d5e2f957eb923b8ed985dbea956f4f1459f2d44653b002504c1d04d4d7f4d07c`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Finite observed historical stages, with fixture failures kept distinct from REDs。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-17`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--evidence--ticket-17--expand-trace-py"></a>

## evidence/ticket-17/expand-trace.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-17/expand-trace.py](../../barness-ai-minimax/evidence/ticket-17/expand-trace.py)；SHA256 `a371f61a308e550abe6d374f500555d8e13be95039d1c3a18212c0acba68355e`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：One reviewed additive finite trace revision; never change prior IDs/cases。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L37 `TRACE.write_bytes(after)`；L37 `(BASE / 'trace-before.json.gz').write_bytes(gzip.compress(before, mtime=0))`；L37 `(BASE / 'trace-after.json.gz').write_bytes(gzip.compress(after, mtime=0))`；L37 `(BASE / 'trace-expansion.json').write_text(json.dumps(proof, separators=(',', ':')) + '\n')`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-17`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-17--freeze-py"></a>

## evidence/ticket-17/freeze.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-17/freeze.py](../../barness-ai-minimax/evidence/ticket-17/freeze.py)；SHA256 `d4f971c81411a8c86db29001c7d7a4392e85b611e694fb8a6ab01b9147ed2feb`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Exclusively freeze the complete closed offline TDD, with unchanged full bounds。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：`p.add_argument('--common-root', type=Path, default=COMMON)` L82。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L44 `archive.parent.mkdir(parents=True, exist_ok=True)`；L68 `(BASE / 'inventory.json.gz').write_bytes(encoded)`；L68 `(BASE / 'source-proof.json.gz').write_bytes(source_encoded)`；L80 `(BASE / 'index.json').write_text(json.dumps(index, separators=(',', ':')) + '\n')`；L80 `archive.chmod(292)`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-17`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L6 `BASE=Path(__file__).resolve().parent;ROOT=BASE.parents[3];RAW=ROOT/'.evidence/minimax-ticket17';COMMON=Path('/Users/cyber/RestoX/barness');sha=lambda v:hashlib.sha256(v).hexdigest()`；L40 `selected['controllers/runner.py']=Path('/tmp/barness-ticket17-run.py').resolve();selected['controllers/freeze.py']=Path(__file__)`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L6 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-17--readers-py"></a>

## evidence/ticket-17/readers.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-17/readers.py](../../barness-ai-minimax/evidence/ticket-17/readers.py)；SHA256 `5e01d36494d35dcd8dc1754b2caa69455c22a51a076e05834fcf89596e5a94b6`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Reuse the existing two pinned bounded protocols without changing their limits。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-17`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L4 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-17--retention-fixes--creator-probe-py"></a>

## evidence/ticket-17/retention-fixes/creator-probe.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/creator-probe.py](../../barness-ai-minimax/evidence/ticket-17/retention-fixes/creator-probe.py)；SHA256 `dd5109648420fc588534cf1395f94bfdc66307de515f753d8ba2d3a1d033eab6`。诊断装置 / 仅原轮次。原缺陷/资源/文件系统诊断；依赖原来源与私有fixture，不是产品资格入口。

**职责**：Exercise actual public creators at real OS opens on isolated private files。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定失败模型及私有源/索引/文件系统副本，部分构造器/probe包含合成外部程序。参数：`p.add_argument('--creator', choices=('ticket17', 'race', 'verdict'), required=True)` L93；`p.add_argument('--case', required=True)` L93；`p.add_argument('--expect', choices=('unsafe', 'bounded'))` L93；`p.add_argument('--out')` L93；`p.add_argument('--child', action='store_true')` L93；`p.add_argument('--producer')` L93；`p.add_argument('--common')` L93；`p.add_argument('--events')` L93。

**输出/写入**：私有fixture、探针日志与有限断言；如处在raw目录则仅为封存源码。实际静态写点：L54 `out.mkdir(parents=True, exist_ok=False)`；L55 `shutil.copytree(t17, dest, ignore=shutil.ignore_patterns('retention-fixes', 'index.json', '__pycache__'))`；L68 `(common / '.evidence/minimax-retained/ticket-15-new-actual').mkdir(parents=True)`；L86 `(out / 'receipt.json').write_text(json.dumps(row, indent=2) + '\n')`；L35 `path.unlink()`；L57 `target.parent.mkdir(parents=True, exist_ok=True)`；L57 `shutil.copyfile(REPO / relative / 'ticket-13' / rel, target)`；L59 `shutil.copytree(raw, clone / '.evidence/minimax-ticket17', copy_function=os.link)`；L62 `shutil.copytree(actual, target, ignore=shutil.ignore_patterns('__pycache__'))`；L64 `(target / lane / 'index.json').unlink()`；L66 `shutil.copyfile(actual / 'race/index.json', target / 'race/index.json')`；L47 `sibling.write_bytes(old)`；L47 `path.symlink_to(sibling.name)`。

**副作用**：可能启动公共CLI/构造器、修改/删除私有文件或建立FIFO；只用于原问题的诊断，不是通用维护入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L5 `BASE=Path(__file__).resolve().parent;REPO=BASE.parents[4];PRIMARY=Path('/Users/cyber/RestoX/barness');sha=lambda v:hashlib.sha256(v).hexdigest()`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-17--retention-fixes--factory-correction--check-factory-py"></a>

## evidence/ticket-17/retention-fixes/factory-correction/check-factory.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/factory-correction/check-factory.py](../../barness-ai-minimax/evidence/ticket-17/retention-fixes/factory-correction/check-factory.py)；SHA256 `82da61891ec1e0afc1191139170d63b0a1cfe9d8078d5e4d191d4c45591dea30`。诊断装置 / 仅原轮次。原缺陷/资源/文件系统诊断；依赖原来源与私有fixture，不是产品资格入口。

**职责**：Actual public constructor at real OS opens, with only private FS mutations。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定失败模型及私有源/索引/文件系统副本，部分构造器/probe包含合成外部程序。参数：`p.add_argument('--case', choices=('index-size', 'inventory-compressed', 'inventory-decoded', 'seal-metadata', 'output-grow'), required=True)` L68；`p.add_argument('--expect', choices=('unsafe', 'bounded'))` L68；`p.add_argument('--out')` L68；`p.add_argument('--common-root', default='/Users/cyber/RestoX/barness')` L68；`p.add_argument('--child', action='store_true')` L68；`p.add_argument('--producer')` L68；`p.add_argument('--output')` L68；`p.add_argument('--events')` L68。

**输出/写入**：私有fixture、探针日志与有限断言；如处在raw目录则仅为封存源码。实际静态写点：L50 `out.mkdir(parents=True, exist_ok=False)`；L50 `shutil.copytree(FRONT, front, ignore=shutil.ignore_patterns('retention-fixes', '__pycache__'))`；L50 `producer.parent.mkdir()`；L50 `shutil.copyfile(FRONT / 'retention-fixes/resource-probe.py', producer)`；L54 `resource.parent.mkdir(parents=True)`；L54 `resource.write_bytes(values[RESOURCE])`；L61 `(out / 'receipt.json').write_text(json.dumps(row, indent=2) + '\n')`；L22 `path.unlink()`；L22 `path.write_bytes(data)`；L52 `target.parent.mkdir(parents=True, exist_ok=True)`；L52 `shutil.copyfile(FRONT.parent / 'ticket-13' / rel, target)`；L44 `target.unlink()`。

**副作用**：可能启动公共CLI/构造器、修改/删除私有文件或建立FIFO；只用于原问题的诊断，不是通用维护入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/factory-correction`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L68 `p=argparse.ArgumentParser();p.add_argument('--case',choices=('index-size','inventory-compressed','inventory-decoded','seal-metadata','output-grow'),required=True);p.add_argument('--expect',choices=('unsafe','bounded'));p.add_argument('--out`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-17--retention-fixes--factory-correction--raw-v2--controllers--check-factory-py"></a>

## evidence/ticket-17/retention-fixes/factory-correction/raw-v2/controllers/check-factory.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/factory-correction/raw-v2/controllers/check-factory.py](../../barness-ai-minimax/evidence/ticket-17/retention-fixes/factory-correction/raw-v2/controllers/check-factory.py)；SHA256 `82da61891ec1e0afc1191139170d63b0a1cfe9d8078d5e4d191d4c45591dea30`。历史源码 / 封存源码快照。raw/controller中的捕获源码，保留原字节，不按同名活动测试处理。

**职责**：Actual public constructor at real OS opens, with only private FS mutations。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定失败模型及私有源/索引/文件系统副本，部分构造器/probe包含合成外部程序。参数：`p.add_argument('--case', choices=('index-size', 'inventory-compressed', 'inventory-decoded', 'seal-metadata', 'output-grow'), required=True)` L68；`p.add_argument('--expect', choices=('unsafe', 'bounded'))` L68；`p.add_argument('--out')` L68；`p.add_argument('--common-root', default='/Users/cyber/RestoX/barness')` L68；`p.add_argument('--child', action='store_true')` L68；`p.add_argument('--producer')` L68；`p.add_argument('--output')` L68；`p.add_argument('--events')` L68。

**输出/写入**：私有fixture、探针日志与有限断言；如处在raw目录则仅为封存源码。实际静态写点：L50 `out.mkdir(parents=True, exist_ok=False)`；L50 `shutil.copytree(FRONT, front, ignore=shutil.ignore_patterns('retention-fixes', '__pycache__'))`；L50 `producer.parent.mkdir()`；L50 `shutil.copyfile(FRONT / 'retention-fixes/resource-probe.py', producer)`；L54 `resource.parent.mkdir(parents=True)`；L54 `resource.write_bytes(values[RESOURCE])`；L61 `(out / 'receipt.json').write_text(json.dumps(row, indent=2) + '\n')`；L22 `path.unlink()`；L22 `path.write_bytes(data)`；L52 `target.parent.mkdir(parents=True, exist_ok=True)`；L52 `shutil.copyfile(FRONT.parent / 'ticket-13' / rel, target)`；L44 `target.unlink()`。

**副作用**：可能启动公共CLI/构造器、修改/删除私有文件或建立FIFO；只用于原问题的诊断，不是通用维护入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/factory-correction/raw-v2/controllers`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L68 `p=argparse.ArgumentParser();p.add_argument('--case',choices=('index-size','inventory-compressed','inventory-decoded','seal-metadata','output-grow'),required=True);p.add_argument('--expect',choices=('unsafe','bounded'));p.add_argument('--out`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-17--retention-fixes--factory-correction--raw-v2--controllers--run-legitimate-py"></a>

## evidence/ticket-17/retention-fixes/factory-correction/raw-v2/controllers/run-legitimate.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/factory-correction/raw-v2/controllers/run-legitimate.py](../../barness-ai-minimax/evidence/ticket-17/retention-fixes/factory-correction/raw-v2/controllers/run-legitimate.py)；SHA256 `4b8bef5dde4eb03abcbc41b3b2078a9363313315dcd6978e3e8352d2084b8583`。历史源码 / 封存源码快照。raw/controller中的捕获源码，保留原字节，不按同名活动测试处理。

**职责**：Numeric receipt for one unmodified-input current public factory construction。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定失败模型及私有源/索引/文件系统副本，部分构造器/probe包含合成外部程序。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：私有fixture、探针日志与有限断言；如处在raw目录则仅为封存源码。实际静态写点：L7 `out.mkdir(exist_ok=False)`；L9 `(out / 'receipt.json').write_text(json.dumps(row, indent=2) + '\n')`。

**副作用**：可能启动公共CLI/构造器、修改/删除私有文件或建立FIFO；只用于原问题的诊断，不是通用维护入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/factory-correction/raw-v2/controllers`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L7 `root=REPO/'.evidence/minimax-ticket17-resource-factory-2';out=root/'legitimate-driver';out.mkdir(exist_ok=False);producer=root/'green-index-size/source/.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/resource-probe.py';destin`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-17--retention-fixes--factory-correction--raw--controllers--check-factory-py"></a>

## evidence/ticket-17/retention-fixes/factory-correction/raw/controllers/check-factory.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/factory-correction/raw/controllers/check-factory.py](../../barness-ai-minimax/evidence/ticket-17/retention-fixes/factory-correction/raw/controllers/check-factory.py)；SHA256 `82da61891ec1e0afc1191139170d63b0a1cfe9d8078d5e4d191d4c45591dea30`。历史源码 / 封存源码快照。raw/controller中的捕获源码，保留原字节，不按同名活动测试处理。

**职责**：Actual public constructor at real OS opens, with only private FS mutations。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定失败模型及私有源/索引/文件系统副本，部分构造器/probe包含合成外部程序。参数：`p.add_argument('--case', choices=('index-size', 'inventory-compressed', 'inventory-decoded', 'seal-metadata', 'output-grow'), required=True)` L68；`p.add_argument('--expect', choices=('unsafe', 'bounded'))` L68；`p.add_argument('--out')` L68；`p.add_argument('--common-root', default='/Users/cyber/RestoX/barness')` L68；`p.add_argument('--child', action='store_true')` L68；`p.add_argument('--producer')` L68；`p.add_argument('--output')` L68；`p.add_argument('--events')` L68。

**输出/写入**：私有fixture、探针日志与有限断言；如处在raw目录则仅为封存源码。实际静态写点：L50 `out.mkdir(parents=True, exist_ok=False)`；L50 `shutil.copytree(FRONT, front, ignore=shutil.ignore_patterns('retention-fixes', '__pycache__'))`；L50 `producer.parent.mkdir()`；L50 `shutil.copyfile(FRONT / 'retention-fixes/resource-probe.py', producer)`；L54 `resource.parent.mkdir(parents=True)`；L54 `resource.write_bytes(values[RESOURCE])`；L61 `(out / 'receipt.json').write_text(json.dumps(row, indent=2) + '\n')`；L22 `path.unlink()`；L22 `path.write_bytes(data)`；L52 `target.parent.mkdir(parents=True, exist_ok=True)`；L52 `shutil.copyfile(FRONT.parent / 'ticket-13' / rel, target)`；L44 `target.unlink()`。

**副作用**：可能启动公共CLI/构造器、修改/删除私有文件或建立FIFO；只用于原问题的诊断，不是通用维护入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/factory-correction/raw/controllers`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L68 `p=argparse.ArgumentParser();p.add_argument('--case',choices=('index-size','inventory-compressed','inventory-decoded','seal-metadata','output-grow'),required=True);p.add_argument('--expect',choices=('unsafe','bounded'));p.add_argument('--out`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-17--retention-fixes--factory-correction--raw--controllers--run-legitimate-py"></a>

## evidence/ticket-17/retention-fixes/factory-correction/raw/controllers/run-legitimate.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/factory-correction/raw/controllers/run-legitimate.py](../../barness-ai-minimax/evidence/ticket-17/retention-fixes/factory-correction/raw/controllers/run-legitimate.py)；SHA256 `4b8bef5dde4eb03abcbc41b3b2078a9363313315dcd6978e3e8352d2084b8583`。历史源码 / 封存源码快照。raw/controller中的捕获源码，保留原字节，不按同名活动测试处理。

**职责**：Numeric receipt for one unmodified-input current public factory construction。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定失败模型及私有源/索引/文件系统副本，部分构造器/probe包含合成外部程序。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：私有fixture、探针日志与有限断言；如处在raw目录则仅为封存源码。实际静态写点：L7 `out.mkdir(exist_ok=False)`；L9 `(out / 'receipt.json').write_text(json.dumps(row, indent=2) + '\n')`。

**副作用**：可能启动公共CLI/构造器、修改/删除私有文件或建立FIFO；只用于原问题的诊断，不是通用维护入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/factory-correction/raw/controllers`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L7 `root=REPO/'.evidence/minimax-ticket17-resource-factory-2';out=root/'legitimate-driver';out.mkdir(exist_ok=False);producer=root/'green-index-size/source/.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/resource-probe.py';destin`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-17--retention-fixes--factory-correction--verify-factory-py"></a>

## evidence/ticket-17/retention-fixes/factory-correction/verify-factory.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/factory-correction/verify-factory.py](../../barness-ai-minimax/evidence/ticket-17/retention-fixes/factory-correction/verify-factory.py)；SHA256 `5cbaf441e135d4834860754d2d91368ac0f63958b80b912cf15171f1c8ba1669`。诊断装置 / 仅原轮次。原缺陷/资源/文件系统诊断；依赖原来源与私有fixture，不是产品资格入口。

**职责**：Read the bounded raw repair capsule without executing any captured program。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定失败模型及私有源/索引/文件系统副本，部分构造器/probe包含合成外部程序。参数：`p.add_argument('--common-root', type=Path, required=True)` L27。

**输出/写入**：私有fixture、探针日志与有限断言；如处在raw目录则仅为封存源码。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：可能启动公共CLI/构造器、修改/删除私有文件或建立FIFO；只用于原问题的诊断，不是通用维护入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/factory-correction`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-17--retention-fixes--positive-creator-py"></a>

## evidence/ticket-17/retention-fixes/positive-creator.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/positive-creator.py](../../barness-ai-minimax/evidence/ticket-17/retention-fixes/positive-creator.py)；SHA256 `a4de26c2bdb3c1691139240baa32678ee47b0cab226ebea390e3e1c8e15cf69b`。诊断装置 / 仅原轮次。原缺陷/资源/文件系统诊断；依赖原来源与私有fixture，不是产品资格入口。

**职责**：Reuse one isolated failed-before-write fixture for a normal bounded CLI success。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定失败模型及私有源/索引/文件系统副本，部分构造器/probe包含合成外部程序。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：私有fixture、探针日志与有限断言；如处在raw目录则仅为封存源码。实际静态写点：L7 `out.mkdir(exist_ok=False)`；L7 `private.unlink()`；L11 `(out / 'receipt.json').write_text(json.dumps(row, indent=2) + '\n')`。

**副作用**：可能启动公共CLI/构造器、修改/删除私有文件或建立FIFO；只用于原问题的诊断，不是通用维护入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-17--retention-fixes--resource-probe-py"></a>

## evidence/ticket-17/retention-fixes/resource-probe.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/resource-probe.py](../../barness-ai-minimax/evidence/ticket-17/retention-fixes/resource-probe.py)；SHA256 `d0578a2e778aa139253748bcdf24201d8986c359061daea1448ba007867f85ac`。诊断装置 / 仅原轮次。原缺陷/资源/文件系统诊断；依赖原来源与私有fixture，不是产品资格入口。

**职责**：Private single-counter mutation; public verifier proves exact-integer behavior。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定失败模型及私有源/索引/文件系统副本，部分构造器/probe包含合成外部程序。参数：`p.add_argument('--common-root', default='/Users/cyber/RestoX/barness')` L77；`p.add_argument('--out', required=True)` L77；`p.add_argument('--value', choices=('false', 'float'))` L77；`p.add_argument('--recheck', action='store_true')` L77。

**输出/写入**：私有fixture、探针日志与有限断言；如处在raw目录则仅为封存源码。实际静态写点：L22 `(base / 'SHA256SUMS').write_text(''.join((line for _, line in sorted(rows))))`；L24 `out.with_suffix('.stdout').write_bytes(run.stdout)`；L24 `out.with_suffix('.stderr').write_bytes(run.stderr)`；L24 `out.with_suffix('.json').write_text(json.dumps(row, indent=2) + '\n')`；L30 `out.mkdir(parents=True, exist_ok=False)`；L30 `shutil.copytree(FRONT, base, ignore=shutil.ignore_patterns('retention-fixes', '__pycache__'))`；L33 `destination.parent.mkdir(parents=True)`；L63 `(base / 'inventory.json.gz').write_bytes(encoded)`；L75 `(base / 'index.json').write_text(json.dumps(index) + '\n')`；L75 `(out / 'construction.json').write_text(json.dumps(dict(kind='NON_AUTHORITATIVE-negative-resource-fixture', originalArchiveSHA256='5d9b5d5ad032fdcba150d294cf65f0944f6219068ec810f49a86caa45d71ef8d', value=a.value, changes=changed, files=len(pins), unchangedOtherFiles=len(pins) - 2, archive=index['archive'], limits=full.LIMITS, credentialsRead=False, `；L29 `shutil.copyfile(FRONT / 'verify.py', base / 'verify.py')`；L32 `target.parent.mkdir(parents=True, exist_ok=True)`；L32 `shutil.copyfile(FRONT.parent / 'ticket-13' / rel, target)`；L57 `(out / ('mutated-resource.json' if name == resource else 'mutated-manifest.json')).write_bytes(data)`。

**副作用**：可能启动公共CLI/构造器、修改/删除私有文件或建立FIFO；只用于原问题的诊断，不是通用维护入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L77 `p=argparse.ArgumentParser();p.add_argument('--common-root',default='/Users/cyber/RestoX/barness');p.add_argument('--out',required=True);p.add_argument('--value',choices=('false','float'));p.add_argument('--recheck',action='store_true');main`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-17--retention-fixes--verify-corrections-py"></a>

## evidence/ticket-17/retention-fixes/verify-corrections.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/verify-corrections.py](../../barness-ai-minimax/evidence/ticket-17/retention-fixes/verify-corrections.py)；SHA256 `9c78825007ca329dae2e99410259ea1bceefe9a18f0f472e6d087cb37c62e0cc`。诊断装置 / 仅原轮次。原缺陷/资源/文件系统诊断；依赖原来源与私有fixture，不是产品资格入口。

**职责**：Read-only replay of exact correction evidence; negative fixtures never certify。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定失败模型及私有源/索引/文件系统副本，部分构造器/probe包含合成外部程序。参数：`p.add_argument('--common-root', type=Path, required=True)` L42。

**输出/写入**：私有fixture、探针日志与有限断言；如处在raw目录则仅为封存源码。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：可能启动公共CLI/构造器、修改/删除私有文件或建立FIFO；只用于原问题的诊断，不是通用维护入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-17--verify-py"></a>

## evidence/ticket-17/verify.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-17/verify.py](../../barness-ai-minimax/evidence/ticket-17/verify.py)；SHA256 `d0af930d54debc32bf1ab16b34dc76a497f3dc1595c162c11f3de5a432b69b0a`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Verify complete original offline history; no captured code is executed。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`p.add_argument('--common-root', type=Path, required=True)` L74。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-17`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L6 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-19--native-home-redaction--verify-py"></a>

## evidence/ticket-19/native-home-redaction/verify.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-19/native-home-redaction/verify.py](../../barness-ai-minimax/evidence/ticket-19/native-home-redaction/verify.py)；SHA256 `4dd3e90c85695c499024af0766bdc2501dd3ab4452586548d2f0f8523de0bf92`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Read-only bounded custody of the HOME diagnostic regression; no provider calls。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`p.add_argument('--common-root', type=Path, default=BASE.parents[4])` L30。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-19/native-home-redaction`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L12 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-19--native-renewal--verify-py"></a>

## evidence/ticket-19/native-renewal/verify.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-19/native-renewal/verify.py](../../barness-ai-minimax/evidence/ticket-19/native-renewal/verify.py)；SHA256 `78262464d63824c5cf5a94aea4fa18ea639ae422a8fd68ae63ca9e6f26c851e7`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Read-only bounded custody of ticket19 native implementation; no vendor certification。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`parser.add_argument('--common-root', type=Path, default=BASE.parents[4])` L32。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-19/native-renewal`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L12 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-20--jev_fixtures-py"></a>

## evidence/ticket-20/jev_fixtures.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-20/jev_fixtures.py](../../barness-ai-minimax/evidence/ticket-20/jev_fixtures.py)；SHA256 `6dbfeb8505d774c7a5c8a16da086ef49511134537a6e2263464b81880e9f3358`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Create two explicitly derived consumer fixtures; never change native inputs。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L19 `out.mkdir(exist_ok=False)`；L40 `dest.mkdir()`；L61 `(dest / 'manifest.json').write_text(json.dumps(manifest, indent=2) + '\n')`；L62 `(out / (phase + '-provenance.json')).write_text(json.dumps(dict(kind='DERIVED_CONSUMER_FIXTURE_NOT_NATIVE_OBSERVED_PHASE', sourceReportSHA256=original['report.json'], sourceFiles=original, exactDelta=changed, manifestResealed=True, reason='Jev prepareTypeSafeImages before pin; classifier.go226/typesafe_images.go33,60; read0/unresolved/no wire retai`；L43 `p.parent.mkdir(parents=True, exist_ok=True)`；L44 `p.write_bytes(bounded.read(source, name, 16 << 20))`；L56 `p.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-20`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L11 `Path(__file__).resolve()`；L14 `Path(__file__).resolve()`；L15 `Path(sys.argv[1]).resolve()`；L16 `Path(sys.argv[2]).resolve()`；L19 `out.mkdir(exist_ok=False)`。

<a id="barness-ai-minimax--evidence--ticket-20--run-py"></a>

## evidence/ticket-20/run.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-20/run.py](../../barness-ai-minimax/evidence/ticket-20/run.py)；SHA256 `d6b1b3e0c681dbd70c48baf1f4f8c7631d041fc1f02fedcf39547261c02f8ccf`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：Record finite keyless commands and exact measured production source hashes。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：L13 `out.mkdir(parents=True, exist_ok=False)`；L22 `(out / 'source-before.json').write_text(json.dumps(before, sort_keys=True, indent=2) + '\n')`；L27 `(out / 'source-after.json').write_text(json.dumps(after, sort_keys=True, indent=2) + '\n')`；L29 `(out / 'command.json').write_text(json.dumps(receipt, indent=2) + '\n')`。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-20`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L11 `Path(__file__).resolve()`；L13 `out.mkdir(parents=True, exist_ok=False)`；L22 `(out / 'source-before.json').write_text(json.dumps(before, sort_keys=True, indent=2) + '\n')`；L22 `json.dumps(before, sort_keys=True, indent=2)`；L24 `(out / 'stdout.log').open('wb')`。

<a id="barness-ai-minimax--evidence--ticket-20--verify-py"></a>

## evidence/ticket-20/verify.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-20/verify.py](../../barness-ai-minimax/evidence/ticket-20/verify.py)；SHA256 `86f41c859e649728ff2c81892e3ccf5bfcc6dd105b3213767633e24930c87ad4`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Bounded read-only custody of ticket20 offline TDD; executes no captured command。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`p.add_argument('--common-root', type=Path, required=True)` L40。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-20`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L11 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-21--actual-retained--public-guards-py"></a>

## evidence/ticket-21/actual-retained/public-guards.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-21/actual-retained/public-guards.py](../../barness-ai-minimax/evidence/ticket-21/actual-retained/public-guards.py)；SHA256 `4f159f0d1100779789c2e89f08355ae43ad9a030bbde12e4668fd71cd30626f3`。测试 / 仅原轮次。原归档/原控制器的失败清单，未发现当前公共Go或现行SOP执行接线。

**职责**：Public CLI custody negatives; private copies only, no native/captured execution。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定源码/工具/归档和私有篡改副本或合成fixture；支持模块随对应测试使用。参数：`p.add_argument('--common-root', type=Path, required=True)` L30；`p.add_argument('--out', type=Path, required=True)` L31。

**输出/写入**：该历史失败清单的断言/拒绝记录和临时日志；不续期真实资格。实际静态写点：L32 `a.out.mkdir(parents=True, exist_ok=False, mode=448)`；L69 `(a.out / 'results.json').write_text(json.dumps(records, indent=2) + '\n')`；L43 `target.mkdir(parents=True, mode=448)`；L61 `(target / 'index.json').write_text(json.dumps(index, indent=2) + '\n')`；L65 `(case / 'raw.log').write_bytes(result.stdout)`；L45 `(target / name_to_copy).write_bytes(b.read(BASE, name_to_copy, 8 << 20))`；L48 `destination.parent.mkdir(parents=True, exist_ok=True)`；L49 `destination.write_bytes(b.read(BASE.parent.parent, relative, 64 << 10))`；L56 `(target / 'inventory.json.gz').write_bytes(b'0' * ((8 << 20) + 1))`；L59 `link.parent.mkdir(parents=True)`；L60 `link.symlink_to(a.common_root / index['capsule']['archive']['file'])`。

**副作用**：运行相邻公共CLI或临时程序，写/删私有fixture；部分故障测试会临时替换再恢复输入，仅原环境适用，不在本次执行。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-21/actual-retained`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L13 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-21--actual-retained--verify-py"></a>

## evidence/ticket-21/actual-retained/verify.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-21/actual-retained/verify.py](../../barness-ai-minimax/evidence/ticket-21/actual-retained/verify.py)；SHA256 `85ee06e783293ab35b6cbaa06081a7163f4e08adfd9ed8fa5d5083dde72926a0`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Bounded read-only custody of the closed36/36 round; no captured execution。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`p.add_argument('--common-root', type=Path, required=True)` L177。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-21/actual-retained`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L12 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--evidence--ticket-21--final-gate--supplement--verify-py"></a>

## evidence/ticket-21/final-gate/supplement/verify.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-21/final-gate/supplement/verify.py](../../barness-ai-minimax/evidence/ticket-21/final-gate/supplement/verify.py)；SHA256 `de9550f1e4458310ead810d6154a4b1e3929077fb642cdba140d50d1a88f825c`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Public readback of fixed closed publication/history supplements; no capture。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`p.add_argument('--common-root', type=Path, required=True)` L106；`p.add_argument('--metadata-root', type=Path, default=Path(__file__).resolve().parent)` L106；`p.add_argument('--repository-root', type=Path)` L106。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-21/final-gate/supplement`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--evidence--ticket-21--final-gate--verify-py"></a>

## evidence/ticket-21/final-gate/verify.py

原件：[.scratch/barness-ai-minimax/evidence/ticket-21/final-gate/verify.py](../../barness-ai-minimax/evidence/ticket-21/final-gate/verify.py)；SHA256 `c8822bb073117b87b15ef026e5a5b9055a8f415f9a66a0a382e1123a015b3c67`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Prospective public readback of the three fixed closed gate capsules; no capture。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`p.add_argument('--common-root', type=Path, required=True)` L159；`p.add_argument('--metadata-root', type=Path, default=Path(__file__).resolve().parent)` L159；`p.add_argument('--repository-root', type=Path)` L159。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/evidence/ticket-21/final-gate`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--session-analysis--build_report-py"></a>

## session-analysis/build_report.py

原件：[.scratch/barness-ai-minimax/session-analysis/build_report.py](../../barness-ai-minimax/session-analysis/build_report.py)；SHA256 `3e7618a3f9b1d0255efcfe612eb24f03c6c29c63b40584939932f512d5dce12c`。历史源码 / 仅原轮次。固定实施会话报表；派生展示不是验收。

**职责**：Render a read-only, reproducible analysis of retained FAILs and session timing。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原实施时间窗内既有归档/会话/门禁记录及固定路径；不重放捕获命令。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：同目录raw-failed-cases/scan-index/gate-reports或all-failures/summary/report.html。实际静态写点：L137 `(OUT / 'all-failures.json').write_text(json.dumps({'scope': '本次实施时间范围内保留的118份归档中的FAIL case，另补最终A两套六项失败、72条业务样本、实际finalizer与诊断门禁、已知审计/工具失败。按原始manifest语义身份去重重复归档；不同运行和父子记录保留。负例子进程FAIL且父测试PASS单独标注；NOT_RUN和成功拒绝的PASS不计入FAIL。不是独立bug总数。', 'limitations': ['这是保留证据的穷举清单，不把未留存stderr的准备失败编造为确定原文。', '测试开发期RED的根因分辨到协议/模块义务；不能仅从断言为每次中间修改推断唯一代码行。', '源代码和日志未重新执行；未访`；L143 `(OUT / 'summary.json').write_text(json.dumps(summary, ensure_ascii=False, indent=2) + '\n')`；L156 `(OUT / 'report.html').write_text(page)`。

**副作用**：只读既有原件，但会写/覆盖分析派生JSON/HTML；是历史会话报表，无实时验收或授权作用。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/session-analysis`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L125 `extra('baseline/redaction-home','go test ./...: package ai/e2e FAIL; audit: P0-GLM13-full-rollback/commands.json [environment] the home directory appears 2 time(s)','初始基线的回滚命令证据包含HOME绝对路径；测试体PASS，但审计使包返回失败。','既有基线问题；非MiniMax业务适配失败','/tmp/ba`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L4 `pathlib.Path(__file__).resolve()`；L4 `pathlib.Path(__file__).resolve()`；L6 `json.loads((OUT / 'raw-failed-cases.json').read_text())`；L6 `(OUT / 'raw-failed-cases.json').read_text()`；L7 `json.loads((ROOT / '.evidence/minimax-global-final-keyless/gate/offline/20261018T045647.697909000Z/manifest.json').read_text())`。

<a id="barness-ai-minimax--session-analysis--collect-py"></a>

## session-analysis/collect.py

原件：[.scratch/barness-ai-minimax/session-analysis/collect.py](../../barness-ai-minimax/session-analysis/collect.py)；SHA256 `589846100b8b52f3d52982611e242efc545d856e574cea2fa6104d74fcecab9a`。历史源码 / 仅原轮次。固定实施会话报表；派生展示不是验收。

**职责**：Read existing implementation evidence only; never execute captured commands。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原实施时间窗内既有归档/会话/门禁记录及固定路径；不重放捕获命令。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：同目录raw-failed-cases/scan-index/gate-reports或all-failures/summary/report.html。实际静态写点：L68 `(OUT / 'raw-failed-cases.json').write_text(json.dumps(rows, ensure_ascii=False, indent=2) + '\n')`；L75 `(OUT / 'scan-index.json').write_text(json.dumps({'archives': [{'archive': r['archive'], 'manifests': r['manifest_count'], 'fail_cases': len(r['failures']), 'audits_with_findings': [{'member': n, 'findings': d.get('findings')} for n, d in r['audits']]} for r in results], 'unique_failure_occurrences': len(rows)}, ensure_ascii=False, indent=2) + '\n')`；L76 `(OUT / 'gate-reports.json').write_text(json.dumps(selected, ensure_ascii=False, indent=2) + '\n')`。

**副作用**：只读既有原件，但会写/覆盖分析派生JSON/HTML；是历史会话报表，无实时验收或授权作用。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/session-analysis`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L4 `pathlib.Path(__file__).resolve()`；L54 `BASE.rglob('*.tar.gz')`；L68 `(OUT / 'raw-failed-cases.json').write_text(json.dumps(rows, ensure_ascii=False, indent=2) + '\n')`；L68 `json.dumps(rows, ensure_ascii=False, indent=2)`；L75 `(OUT / 'scan-index.json').write_text(json.dumps({'archives': [{'archive': r['archive'], 'manifests': r['manifest_count'], 'fail_cases': len(r['failures']), 'audits_with_findings': [{'member': n, 'findings': d.get('findings')} for n, d in r['audits']]} for r in results], 'unique_failure_occurrences': len(rows)}, ensure_ascii=False, indent=2) + '\n')`。

<a id="barness-ai-minimax--tools--eleven_contract-py"></a>

## tools/eleven_contract.py

原件：[.scratch/barness-ai-minimax/tools/eleven_contract.py](../../barness-ai-minimax/tools/eleven_contract.py)；SHA256 `6198e1057ed79031fd78d6e02309e0089c3444715a5d27821b4c2fa8b81b0e94`。活动工具 / 仅测试消费，旧真实轮次关闭。现有test_*.py覆盖固定控制器及模块；真实旧轮次仅封存，不是日常编排入口。

**职责**：One fixed eleven-group authority; its two phases share one finite budget。

**调用方**：[.scratch/barness-ai-minimax/tools/eleven_finalize.py](../../barness-ai-minimax/tools/eleven_finalize.py) L7（import）；[.scratch/barness-ai-minimax/tools/eleven_grant.py](../../barness-ai-minimax/tools/eleven_grant.py) L6（import）；[.scratch/barness-ai-minimax/tools/eleven_inputs.py](../../barness-ai-minimax/tools/eleven_inputs.py) L11（import）；[.scratch/barness-ai-minimax/tools/eleven_validation.py](../../barness-ai-minimax/tools/eleven_validation.py) L16（import）；[.scratch/barness-ai-minimax/tools/renewal_glm.py](../../barness-ai-minimax/tools/renewal_glm.py) L169（import）；[.scratch/barness-ai-minimax/tools/renewal_glm.py](../../barness-ai-minimax/tools/renewal_glm.py) L153（import）；[.scratch/barness-ai-minimax/tools/renewal_glm.py](../../barness-ai-minimax/tools/renewal_glm.py) L68（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--tools--eleven_finalize-py"></a>

## tools/eleven_finalize.py

原件：[.scratch/barness-ai-minimax/tools/eleven_finalize.py](../../barness-ai-minimax/tools/eleven_finalize.py)；SHA256 `c5bf639f10975be958d58b05bf0647f11b64289fed21000385d1ed83c4d7a2f5`。活动工具 / 仅测试消费，旧真实轮次关闭。现有test_*.py覆盖固定控制器及模块；真实旧轮次仅封存，不是日常编排入口。

**职责**：Only current own eleven collections may atomically replace seven matrix rows。

**调用方**：[.scratch/barness-ai-minimax/tools/eleven_validation.py](../../barness-ai-minimax/tools/eleven_validation.py) L144（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L21 `c.dump(intent, dict(state='STARTING', credentialsRead=False, sourceCommit=c.git('rev-parse', 'HEAD'), matrixBeforeSHA256=u.sha(before)), exclusive=True)`；L22 `destination.mkdir(mode=448)`；L23 `candidate.write_bytes(before)`；L44 `c.dump(audit, dict(schema=original['schema'], rows=[{k: row[k] for k in ('combo', 'operation', 'provider', 'api')} for row in original['rows']]))`；L62 `c.dump(destination / 'result.json', result)`；L62 `c.dump(intent, result)`；L58 `os.replace(name, target)`；L60 `os.unlink(name)`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--tools--eleven_grant-py"></a>

## tools/eleven_grant.py

原件：[.scratch/barness-ai-minimax/tools/eleven_grant.py](../../barness-ai-minimax/tools/eleven_grant.py)；SHA256 `5de6bba81f6a32228b00dd6c02aeb5f2c3f38c5f96776219e98054284471c8bd`。活动工具 / 仅测试消费，旧真实轮次关闭。现有test_*.py覆盖固定控制器及模块；真实旧轮次仅封存，不是日常编排入口。

**职责**：Canonical native M3 grant, issued once after the fresh remaining-phase review。

**调用方**：[.scratch/barness-ai-minimax/tools/eleven_validation.py](../../barness-ai-minimax/tools/eleven_validation.py) L19（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L14 `c.dump(out / 'native-renewal-facts.json', value)`；L38 `c.dump(base() / 'grant.json', grant, exclusive=True)`；L11 `c.dump(out / 'native-renewal-facts.json', dict(state='NOT_ISSUED'))`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--tools--eleven_inputs-py"></a>

## tools/eleven_inputs.py

原件：[.scratch/barness-ai-minimax/tools/eleven_inputs.py](../../barness-ai-minimax/tools/eleven_inputs.py)；SHA256 `ee274b7127369c2a97db9c796edcb533da5f24f233268beb5c931bc259b70e0c`。活动工具 / 仅测试消费，旧真实轮次关闭。现有test_*.py覆盖固定控制器及模块；真实旧轮次仅封存，不是日常编排入口。

**职责**：Immutable old actuals and the sole GLM-data transition for ticket19。

**调用方**：[.scratch/barness-ai-minimax/tools/eleven_finalize.py](../../barness-ai-minimax/tools/eleven_finalize.py) L8（import）；[.scratch/barness-ai-minimax/tools/eleven_validation.py](../../barness-ai-minimax/tools/eleven_validation.py) L17（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--tools--eleven_validation-py"></a>

## tools/eleven_validation.py

原件：[.scratch/barness-ai-minimax/tools/eleven_validation.py](../../barness-ai-minimax/tools/eleven_validation.py)；SHA256 `173823d998b55703cb6ba9e7cd1287836a7775993d61f351cd365f0b9a36d7a5`。活动工具 / 仅测试消费，旧真实轮次关闭。现有test_*.py覆盖固定控制器及模块；真实旧轮次仅封存，不是日常编排入口。

**职责**：One fixed current-source eleven renewal; old authorities remain closed.  Approval/phase logic stays separate from closed earlier authorities: this round requires three closed known GLM results, a new M3 grant, and seven-row publication only after all own GLM routes PASS. Stable execution, source freezes, native settlement and preflights use the existing paid modules。

**调用方**：[.scratch/barness-ai-minimax/tools/eleven_finalize.py](../../barness-ai-minimax/tools/eleven_finalize.py) L11（import）；[.scratch/barness-ai-minimax/tools/eleven_grant.py](../../barness-ai-minimax/tools/eleven_grant.py) L31（import）；[.scratch/barness-ai-minimax/tools/test_eleven_validation.py](../../barness-ai-minimax/tools/test_eleven_validation.py) L16（literal-reference）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：`parser.add_argument('mode', choices=('prepare', 'check', 'preview', 'run-selected', 'summarize', 'authorize', 'grant', 'finalize'))` L125；`parser.add_argument('out')` L125；`parser.add_argument('route', nargs='?')` L125；`parser.add_argument('--phase', choices=('glm', 'remaining'))` L125。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L26 `out.mkdir(parents=True, exist_ok=False, mode=448)`；L27 `c.dump(out / 'source-at-start.json', frozen)`；L29 `c.dump(out / 'plan.json', c.redact(dict(schema=1, round=r.ROUND, phase=phase, budget=r.BUDGET, accountingScope=r.ACCOUNTING_SCOPE, history=history, generic=reuse, transition=scope, allJobs=r.jobs(out), collection=e.location(out), routes=r.jobs(out, phase), automaticProcessReruns=False, noSceneSplice=True, proxy=r.PROXY, approvalProposal=r.PROPOSAL,`；L37 `c.dump(out / 'binaries.json', {n: c.digest(out / n) for n, _ in builds})`；L39 `c.dump(out / 'prepared.json', dict(state='PREPARED', credentialsRead=False, planSHA256=c.digest(out / 'plan.json'), authorizationFiles={n: c.digest(out / n) for n in auth}))`；L40 `c.dump(out / 'prerelease-source-review.template.json', dict(decision='PENDING', reviewedBy='', **pins, authorizedRoutes=[], credentialScopes={}))`；L41 `c.dump(out / 'execution-authorization.template.json', dict(decision='PENDING', authority='', requestReference='', requestText='', approvalProposalSHA256=r.PROPOSAL_SHA256, approvedAdditionalMaxCalls=216, approvedAdditionalMaxAttempts=252, round=r.ROUND, phase=phase, budget=r.BUDGET, accountingScope=r.ACCOUNTING_SCOPE, **pins, checksSHA256='', sourc`；L108 `c.dump(intent, c.redact(record), exclusive=True)`；L110 `c.dump(intent, record)`；L114 `c.dump(intent, record)`；L119 `c.dump(out / (name + '-intent.json'), record)`；L119 `c.dump(out / 'eleven-consumption.json', summarize(out))`；L28 `c.dump(out / leaf, value)`；L32 `c.dump(out / name, value)`；L69 `c.dump(path, expected, exclusive=True)`；L115 `c.dump(intent, record)`；L117 `c.dump(intent, record)`；L137 `c.dump(out / 'execution-authorization.template.json', template)`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--tools--glm_corrective-py"></a>

## tools/glm_corrective.py

原件：[.scratch/barness-ai-minimax/tools/glm_corrective.py](../../barness-ai-minimax/tools/glm_corrective.py)；SHA256 `6ed515031902960af16334779da9ef287052ab16d85283d47c63ff7856186a65`。活动工具 / 仅测试消费，旧真实轮次关闭。现有test_*.py覆盖固定控制器及模块；真实旧轮次仅封存，不是日常编排入口。

**职责**：Prospective GLM corrective-1: no authority until explicit human approval.  Exactly three new full22 processes; original canonical intents never change. Distinct preparation is deliberate: this round has a separate approval and carried-consumption contract, not a reset of the original collection。

**调用方**：[.scratch/barness-ai-minimax/tools/paid.py](../../barness-ai-minimax/tools/paid.py) L28（import）；[.scratch/barness-ai-minimax/tools/paid.py](../../barness-ai-minimax/tools/paid.py) L235（import）；[.scratch/barness-ai-minimax/tools/test_glm_corrective.py](../../barness-ai-minimax/tools/test_glm_corrective.py) L34（literal-reference）；[.scratch/barness-ai-minimax/tools/test_glm_corrective.py](../../barness-ai-minimax/tools/test_glm_corrective.py) L129（literal-reference）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：`parser.add_argument('mode', choices=('prepare', 'check', 'preview', 'run-selected', 'summarize'))` L181；`parser.add_argument('out')` L182；`parser.add_argument('route', nargs='?')` L182。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L60 `out.mkdir(parents=True, exist_ok=False, mode=448)`；L61 `c.dump(out / 'source-at-start.json', frozen)`；L62 `c.dump(out / 'plan.json', c.redact(dict(schema=1, round=ROUND, accountingScope=ACCOUNTING_SCOPE, approval='PENDING HUMAN APPROVAL', budget=BUDGET, originalReceipts=prior, routes=jobs(out), automaticProcessReruns=False, noSceneSplice=True, maxOutputTokens=4096, callSeconds=120, processSeconds=1800, defaultRetries=0)))`；L67 `c.dump(out / 'binaries.json', {'live.test': c.digest(out / 'live.test')})`；L68 `c.dump(out / 'prepared.json', dict(state='PREPARED', credentialsRead=False, planSHA256=c.digest(out / 'plan.json'), authorizationFiles={}))`；L72 `c.dump(out / 'prerelease-source-review.template.json', dict(decision='PENDING', reviewedBy='', **pins, authorizedRoutes=[], credentialScopes={}))`；L74 `c.dump(out / 'human-approval.template.json', dict(decision='PENDING', authority='', replyReference='', round=ROUND, accountingScope=ACCOUNTING_SCOPE, **pins, **BUDGET, originalReceipts=prior))`；L132 `c.dump(out / 'corrective-consumption.json', result)`；L161 `c.dump(intent, c.redact(dict(state='STARTING', route=job['name'], correctiveRound=ROUND, sourceCommit=frozen['sourceCommit'], sourceVariable='GLM_KEY', credentialsRead=False, maxCalls=24, maxAttempts=32, collectionRoot=str(out), automaticReruns=False, originalReceipts=originals(), humanApprovalSha256=c.digest(out / 'human-approval.json'), sourceMan`；L171 `c.dump(intent, record)`；L174 `c.dump(out / (job['name'] + '-intent.json'), record)`；L173 `c.dump(intent, record)`；L187 `c.dump(out / 'human-approval.template.json', template)`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--tools--paid-py"></a>

## tools/paid.py

原件：[.scratch/barness-ai-minimax/tools/paid.py](../../barness-ai-minimax/tools/paid.py)；SHA256 `5560df5569b897cc46b315288b67188557a9af47b5c04772b005dff88392cf99`。活动工具 / 仅测试消费，旧真实轮次关闭。现有test_*.py覆盖固定控制器及模块；真实旧轮次仅封存，不是日常编排入口。

**职责**：Task-10 paid collection: prepare/check/preview/run-selected/summarize/keyless-plan.  Approval is a parent-written exact-source receipt. No command creates APPROVED。

**调用方**：[.scratch/barness-ai-minimax/tools/eleven_finalize.py](../../barness-ai-minimax/tools/eleven_finalize.py) L5（import）；[.scratch/barness-ai-minimax/tools/eleven_inputs.py](../../barness-ai-minimax/tools/eleven_inputs.py) L5（import）；[.scratch/barness-ai-minimax/tools/eleven_validation.py](../../barness-ai-minimax/tools/eleven_validation.py) L14（import）；[.scratch/barness-ai-minimax/tools/glm_corrective.py](../../barness-ai-minimax/tools/glm_corrective.py) L16（import）；[.scratch/barness-ai-minimax/tools/release_repair_validation.py](../../barness-ai-minimax/tools/release_repair_validation.py) L8（import）；[.scratch/barness-ai-minimax/tools/repair_inputs.py](../../barness-ai-minimax/tools/repair_inputs.py) L7（import）；[.scratch/barness-ai-minimax/tools/routed_contract.py](../../barness-ai-minimax/tools/routed_contract.py) L6（import）；[.scratch/barness-ai-minimax/tools/routed_evidence.py](../../barness-ai-minimax/tools/routed_evidence.py) L8（import）；[.scratch/barness-ai-minimax/tools/routed_validation.py](../../barness-ai-minimax/tools/routed_validation.py) L10（import）；[.scratch/barness-ai-minimax/tools/test_glm_corrective.py](../../barness-ai-minimax/tools/test_glm_corrective.py) L89（literal-reference）；[.scratch/barness-ai-minimax/tools/test_glm_corrective.py](../../barness-ai-minimax/tools/test_glm_corrective.py) L97（literal-reference）；[.scratch/barness-ai-minimax/tools/test_glm_corrective.py](../../barness-ai-minimax/tools/test_glm_corrective.py) L122（literal-reference）；[.scratch/barness-ai-minimax/tools/test_glm_corrective.py](../../barness-ai-minimax/tools/test_glm_corrective.py) L138（literal-reference）；[.scratch/barness-ai-minimax/tools/test_paid.py](../../barness-ai-minimax/tools/test_paid.py) L90（literal-reference）；[.scratch/barness-ai-minimax/tools/test_renewal.py](../../barness-ai-minimax/tools/test_renewal.py) L308（literal-reference）；[.scratch/barness-ai-minimax/tools/test_renewal.py](../../barness-ai-minimax/tools/test_renewal.py) L322（literal-reference）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：`parser.add_argument('mode', choices=('prepare', 'check', 'preview', 'run-selected', 'summarize', 'keyless-plan'))` L280；`parser.add_argument('out')` L281；`parser.add_argument('route', nargs='?')` L281。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L32 `out.mkdir(parents=True, exist_ok=False, mode=448)`；L33 `c.dump(out / 'source-at-start.json', frozen)`；L35 `c.dump(out / 'plan.json', c.redact(dict(schema=1, reference='barness-ai-minimax-ticket10', maxLegacyCalls=440 if carried else 374, maxLegacyAttempts=464 if carried else 398, providerCaps=caps, providerAttemptCaps=attempts, carriedGLM=carried, miniMaxCandidate=dict(maxCalls=120, maxAttempts=144, maxIdentities=6), miniMaxFormal=dict(maxCalls=60, maxA`；L50 `c.dump(out / 'binaries.json', {n: c.digest(out / n) for n in ('live.test', 'chineseeval', 'visualeval')})`；L53 `c.dump(out / 'prepared.json', prepared)`；L54 `c.dump(out / 'prerelease-source-review.template.json', dict(decision='PENDING', reviewedBy='', sourceCommit=frozen['sourceCommit'], sourceManifestSHA256=c.digest(out / 'source-at-start.json'), binariesManifestSHA256=c.digest(out / 'binaries.json'), planSHA256=prepared['planSHA256'], authorizedRoutes=[], credentialScopes={}))`；L71 `c.dump(out / 'checks.json', accepted)`；L180 `c.dump(intent, c.redact(dict(state='STARTING', route=job['name'], sourceCommit=frozen['sourceCommit'], maxCalls=job['maxCalls'], maxAttempts=job['maxAttempts'], automaticReruns=False, collectionRoot=str(out), sourceVariable=job['sourceVariable'], credentialsRead=False)), exclusive=True)`；L189 `c.dump(intent, record)`；L208 `c.dump(out / (job['name'] + '-intent.json'), record)`；L274 `c.dump(out / 'consumption.json', result)`；L44 `c.dump(out / name, value)`；L194 `c.dump(intent, record)`；L198 `c.dump(intent, record)`；L201 `c.dump(intent, record)`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--tools--paid_capture-py"></a>

## tools/paid_capture.py

原件：[.scratch/barness-ai-minimax/tools/paid_capture.py](../../barness-ai-minimax/tools/paid_capture.py)；SHA256 `83d052452a350a9960140f2db33bf52374ce78dbfe064d22cc1937b4ae917c7a`。活动工具 / 仅测试消费，旧真实轮次关闭。现有test_*.py覆盖固定控制器及模块；真实旧轮次仅封存，不是日常编排入口。

**职责**：Bounded immutable research retention; no credentials or vendor path。

**调用方**：[.scratch/barness-ai-minimax/tools/paid.py](../../barness-ai-minimax/tools/paid.py) L14（import）；[.scratch/barness-ai-minimax/tools/paid_history.py](../../barness-ai-minimax/tools/paid_history.py) L4（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L50 `target.parent.mkdir(parents=True, exist_ok=True, mode=448)`；L51 `lock.mkdir(mode=448)`；L59 `os.rename(stage, target)`；L57 `path.parent.mkdir(parents=True, exist_ok=True, mode=448)`；L61 `shutil.rmtree(stage)`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--tools--paid_checks-py"></a>

## tools/paid_checks.py

原件：[.scratch/barness-ai-minimax/tools/paid_checks.py](../../barness-ai-minimax/tools/paid_checks.py)；SHA256 `fb43c9aded424d475892e1cb9c890299aca33b9e4bcb0822465f292059385034`。活动工具 / 仅测试消费，旧真实轮次关闭。现有test_*.py覆盖固定控制器及模块；真实旧轮次仅封存，不是日常编排入口。

**职责**：The six actual keyless commands shared by the two fixed renewal authorities。

**调用方**：[.scratch/barness-ai-minimax/tools/eleven_contract.py](../../barness-ai-minimax/tools/eleven_contract.py) L5（import）；[.scratch/barness-ai-minimax/tools/eleven_validation.py](../../barness-ai-minimax/tools/eleven_validation.py) L18（import）；[.scratch/barness-ai-minimax/tools/release_repair_validation.py](../../barness-ai-minimax/tools/release_repair_validation.py) L12（import）；[.scratch/barness-ai-minimax/tools/repair_contract.py](../../barness-ai-minimax/tools/repair_contract.py) L40（import）；[.scratch/barness-ai-minimax/tools/routed_validation.py](../../barness-ai-minimax/tools/routed_validation.py) L13（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L20 `c.dump(out / 'checks.json', result)`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--tools--paid_core-py"></a>

## tools/paid_core.py

原件：[.scratch/barness-ai-minimax/tools/paid_core.py](../../barness-ai-minimax/tools/paid_core.py)；SHA256 `d04b6ada7669530169debbfead314e9140f2f6f9295871ecddb42764cad14177`。活动工具 / 仅测试消费，旧真实轮次关闭。现有test_*.py覆盖固定控制器及模块；真实旧轮次仅封存，不是日常编排入口。

**职责**：Committed-source freeze, review guard and isolated selected-secret process。

**调用方**：[.scratch/barness-ai-minimax/tools/eleven_contract.py](../../barness-ai-minimax/tools/eleven_contract.py) L3（import）；[.scratch/barness-ai-minimax/tools/eleven_finalize.py](../../barness-ai-minimax/tools/eleven_finalize.py) L3（import）；[.scratch/barness-ai-minimax/tools/eleven_grant.py](../../barness-ai-minimax/tools/eleven_grant.py) L3（import）；[.scratch/barness-ai-minimax/tools/eleven_inputs.py](../../barness-ai-minimax/tools/eleven_inputs.py) L3（import）；[.scratch/barness-ai-minimax/tools/eleven_validation.py](../../barness-ai-minimax/tools/eleven_validation.py) L12（import）；[.scratch/barness-ai-minimax/tools/glm_corrective.py](../../barness-ai-minimax/tools/glm_corrective.py) L14（import）；[.scratch/barness-ai-minimax/tools/paid.py](../../barness-ai-minimax/tools/paid.py) L12（import）；[.scratch/barness-ai-minimax/tools/paid_checks.py](../../barness-ai-minimax/tools/paid_checks.py) L2（import）；[.scratch/barness-ai-minimax/tools/paid_history.py](../../barness-ai-minimax/tools/paid_history.py) L3（import）；[.scratch/barness-ai-minimax/tools/release_repair_validation.py](../../barness-ai-minimax/tools/release_repair_validation.py) L6（import）；[.scratch/barness-ai-minimax/tools/renewal.py](../../barness-ai-minimax/tools/renewal.py) L9（import）；[.scratch/barness-ai-minimax/tools/renewal_common.py](../../barness-ai-minimax/tools/renewal_common.py) L10（import）；[.scratch/barness-ai-minimax/tools/renewal_glm.py](../../barness-ai-minimax/tools/renewal_glm.py) L6（import）；[.scratch/barness-ai-minimax/tools/repair_anchors.py](../../barness-ai-minimax/tools/repair_anchors.py) L3（import）；[.scratch/barness-ai-minimax/tools/repair_contract.py](../../barness-ai-minimax/tools/repair_contract.py) L3（import）；[.scratch/barness-ai-minimax/tools/repair_inputs.py](../../barness-ai-minimax/tools/repair_inputs.py) L5（import）；[.scratch/barness-ai-minimax/tools/repair_publication.py](../../barness-ai-minimax/tools/repair_publication.py) L4（import）；[.scratch/barness-ai-minimax/tools/routed_contract.py](../../barness-ai-minimax/tools/routed_contract.py) L4（import）；[.scratch/barness-ai-minimax/tools/routed_evidence.py](../../barness-ai-minimax/tools/routed_evidence.py) L7（import）；[.scratch/barness-ai-minimax/tools/routed_validation.py](../../barness-ai-minimax/tools/routed_validation.py) L8（import）；[.scratch/barness-ai-minimax/tools/test_eleven_validation.py](../../barness-ai-minimax/tools/test_eleven_validation.py) L138（import）；[.scratch/barness-ai-minimax/tools/test_repair_fixtures.py](../../barness-ai-minimax/tools/test_repair_fixtures.py) L9（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L35 `path.parent.mkdir(parents=True, exist_ok=True, mode=448)`；L144 `dump(path, redact(record), exclusive=True)`；L45 `os.replace(name, path)`；L153 `dump(path, redact(record, key))`；L165 `(out / (name + '.log')).write_text(redact(output, key))`；L177 `dump(path, safe)`；L47 `os.unlink(name)`；L181 `dump(receipt, intent)`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L16 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--tools--paid_history-py"></a>

## tools/paid_history.py

原件：[.scratch/barness-ai-minimax/tools/paid_history.py](../../barness-ai-minimax/tools/paid_history.py)；SHA256 `a56b2bc83530829335762a46f48917a15ad268619e9e0d167c2d863cd743eae0`。活动工具 / 仅测试消费，旧真实轮次关闭。现有test_*.py覆盖固定控制器及模块；真实旧轮次仅封存，不是日常编排入口。

**职责**：One pinned CN spend inventory; confers no current target/proof authority。

**调用方**：[.scratch/barness-ai-minimax/tools/paid.py](../../barness-ai-minimax/tools/paid.py) L15（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--tools--paid_plan-py"></a>

## tools/paid_plan.py

原件：[.scratch/barness-ai-minimax/tools/paid_plan.py](../../barness-ai-minimax/tools/paid_plan.py)；SHA256 `8213089eee34ff2ff4ad973b835886c167b90fc8fd8ac0b9af7f96eb79979a5a`。活动工具 / 仅测试消费，旧真实轮次关闭。现有test_*.py覆盖固定控制器及模块；真实旧轮次仅封存，不是日常编排入口。

**职责**：Fixed task-10 collection inventory; executable argv never comes from a receipt。

**调用方**：[.scratch/barness-ai-minimax/tools/eleven_contract.py](../../barness-ai-minimax/tools/eleven_contract.py) L4（import）；[.scratch/barness-ai-minimax/tools/eleven_finalize.py](../../barness-ai-minimax/tools/eleven_finalize.py) L4（import）；[.scratch/barness-ai-minimax/tools/eleven_inputs.py](../../barness-ai-minimax/tools/eleven_inputs.py) L4（import）；[.scratch/barness-ai-minimax/tools/eleven_validation.py](../../barness-ai-minimax/tools/eleven_validation.py) L13（import）；[.scratch/barness-ai-minimax/tools/glm_corrective.py](../../barness-ai-minimax/tools/glm_corrective.py) L15（import）；[.scratch/barness-ai-minimax/tools/paid.py](../../barness-ai-minimax/tools/paid.py) L13（import）；[.scratch/barness-ai-minimax/tools/release_repair_validation.py](../../barness-ai-minimax/tools/release_repair_validation.py) L7（import）；[.scratch/barness-ai-minimax/tools/renewal.py](../../barness-ai-minimax/tools/renewal.py) L11（import）；[.scratch/barness-ai-minimax/tools/repair_contract.py](../../barness-ai-minimax/tools/repair_contract.py) L4（import）；[.scratch/barness-ai-minimax/tools/repair_inputs.py](../../barness-ai-minimax/tools/repair_inputs.py) L6（import）；[.scratch/barness-ai-minimax/tools/routed_contract.py](../../barness-ai-minimax/tools/routed_contract.py) L5（import）；[.scratch/barness-ai-minimax/tools/routed_evidence.py](../../barness-ai-minimax/tools/routed_evidence.py) L9（import）；[.scratch/barness-ai-minimax/tools/routed_validation.py](../../barness-ai-minimax/tools/routed_validation.py) L9（import）；[.scratch/barness-ai-minimax/tools/test_eleven_validation.py](../../barness-ai-minimax/tools/test_eleven_validation.py) L138（import）；[.scratch/barness-ai-minimax/tools/test_repair_fixtures.py](../../barness-ai-minimax/tools/test_repair_fixtures.py) L9（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--tools--release_repair_validation-py"></a>

## tools/release_repair_validation.py

原件：[.scratch/barness-ai-minimax/tools/release_repair_validation.py](../../barness-ai-minimax/tools/release_repair_validation.py)；SHA256 `a5ed655967041360e79a0527c7c13aca685265d7f0994b7660b18ec8711b503f`。活动工具 / 仅测试消费，旧真实轮次关闭。现有test_*.py覆盖固定控制器及模块；真实旧轮次仅封存，不是日常编排入口。

**职责**：One pending finite six-process release repair; original authorities stay closed。

**调用方**：[.scratch/barness-ai-minimax/tools/test_repair_fixtures.py](../../barness-ai-minimax/tools/test_repair_fixtures.py) L45（literal-reference）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：`parser.add_argument('mode', choices=('prepare', 'check', 'preview', 'run-selected', 'summarize'))` L97；`parser.add_argument('out')` L97；`parser.add_argument('route', nargs='?')` L97；`parser.add_argument('--phase', choices=('glm', 'effects'))` L97。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L20 `out.mkdir(parents=True, exist_ok=False, mode=448)`；L21 `c.dump(out / 'source-at-start.json', frozen)`；L23 `c.dump(out / 'plan.json', c.redact(dict(schema=1, round=r.ROUND, phase=phase, budget=r.BUDGET, accountingScope=r.ACCOUNTING_SCOPE, history=history, reuse=reuse, scope=scope, collection=e.location(out), routes=r.jobs(out, phase), automaticProcessReruns=False, noSceneSplice=True, proxy=r.PROXY, approvalProposal=r.PROPOSAL, approvalProposalSHA256=r.PR`；L31 `c.dump(out / 'binaries.json', {n: c.digest(out / n) for n, _ in builds})`；L31 `c.dump(out / 'prepared.json', dict(state='PREPARED', credentialsRead=False, planSHA256=c.digest(out / 'plan.json'), authorizationFiles={n: c.digest(out / n) for n in auth}))`；L32 `c.dump(out / 'prerelease-source-review.template.json', dict(decision='PENDING', reviewedBy='', **pins, authorizedRoutes=[], credentialScopes={}))`；L33 `c.dump(out / 'execution-authorization.template.json', dict(decision='PENDING', authority='', requestReference='', requestText='', approvalProposalSHA256=r.PROPOSAL_SHA256, approvedAdditionalMaxCalls=108, approvedAdditionalMaxAttempts=132, round=r.ROUND, phase=phase, budget=r.BUDGET, accountingScope=r.ACCOUNTING_SCOPE, **pins, checksSHA256='', sourc`；L83 `c.dump(intent, c.redact(record), exclusive=True)`；L85 `c.dump(intent, record)`；L89 `c.dump(intent, record)`；L91 `c.dump(out / (name + '-intent.json'), record)`；L91 `c.dump(out / 'repair-consumption.json', summarize(out))`；L22 `c.dump(out / leaf, value)`；L26 `c.dump(out / name, value)`；L90 `c.dump(intent, record)`；L109 `c.dump(out / 'execution-authorization.template.json', template)`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--tools--renewal-py"></a>

## tools/renewal.py

原件：[.scratch/barness-ai-minimax/tools/renewal.py](../../barness-ai-minimax/tools/renewal.py)；SHA256 `3a5eea2a0e75f0dfce561517a41c739a37b125d9ea757ca1f7b57911ef47afbc`。活动工具 / 仅测试消费，旧真实轮次关闭。现有test_*.py覆盖固定控制器及模块；真实旧轮次仅封存，不是日常编排入口。

**职责**：Task-10 keyless catalog pins, GLM renewal, MiniMax origin index and NOT_RUN。

**调用方**：[.scratch/barness-ai-minimax/tools/test_eleven_validation.py](../../barness-ai-minimax/tools/test_eleven_validation.py) L39（literal-reference）；[.scratch/barness-ai-minimax/tools/test_eleven_validation.py](../../barness-ai-minimax/tools/test_eleven_validation.py) L40（literal-reference）；[.scratch/barness-ai-minimax/tools/test_renewal.py](../../barness-ai-minimax/tools/test_renewal.py) L48（literal-reference）；[.scratch/barness-provider-sites/tools/test_final_publication.py](../../barness-provider-sites/tools/test_final_publication.py) L34（literal-reference）；[.scratch/barness-provider-sites/tools/test_final_retry.py](../../barness-provider-sites/tools/test_final_retry.py) L32（literal-reference）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：`p.add_argument('out')` L127；`p.add_argument('--round', choices=('original', 'corrective-1', 'routed-validation-1', 'release-repair-validation-1', 'current-source-eleven-renewal-1', 'provider-sites-final-acceptance-2'), default='original')` L130；`p.add_argument('--bundle', action='append', required=True)` L131；`p.add_argument('--' + route, required=True)` L129。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L93 `c.dump(target, index)`；L121 `c.dump(out / 'm31-not-run-summary.json', result)`；L36 `c.dump(c.ROOT / usage, usage_doc)`；L32 `c.dump(c.ROOT / name, value)`；L40 `(c.ROOT / name).write_bytes(raw)`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--tools--renewal_common-py"></a>

## tools/renewal_common.py

原件：[.scratch/barness-ai-minimax/tools/renewal_common.py](../../barness-ai-minimax/tools/renewal_common.py)；SHA256 `2d624c62ad9fb0e889cba161c29ab7529269fb9fadfaa4c0365a86a81cba65bf`。活动工具 / 仅测试消费，旧真实轮次关闭。现有test_*.py覆盖固定控制器及模块；真实旧轮次仅封存，不是日常编排入口。

**职责**：Bounded regular-file imports and keyless public CLI execution。

**调用方**：[.scratch/barness-ai-minimax/tools/eleven_finalize.py](../../barness-ai-minimax/tools/eleven_finalize.py) L6（import）；[.scratch/barness-ai-minimax/tools/eleven_grant.py](../../barness-ai-minimax/tools/eleven_grant.py) L5（import）；[.scratch/barness-ai-minimax/tools/eleven_inputs.py](../../barness-ai-minimax/tools/eleven_inputs.py) L7（import）；[.scratch/barness-ai-minimax/tools/renewal.py](../../barness-ai-minimax/tools/renewal.py) L10（import）；[.scratch/barness-ai-minimax/tools/renewal_glm.py](../../barness-ai-minimax/tools/renewal_glm.py) L7（import）；[.scratch/barness-ai-minimax/tools/repair_anchors.py](../../barness-ai-minimax/tools/repair_anchors.py) L5（import）；[.scratch/barness-ai-minimax/tools/repair_inputs.py](../../barness-ai-minimax/tools/repair_inputs.py) L10（import）；[.scratch/barness-ai-minimax/tools/repair_publication.py](../../barness-ai-minimax/tools/repair_publication.py) L7（import）；[.scratch/barness-ai-minimax/tools/routed_evidence.py](../../barness-ai-minimax/tools/routed_evidence.py) L11（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L57 `path.mkdir(parents=True, exist_ok=False, mode=448)`；L139 `matrix.write_bytes(file_bytes(c.ROOT / 'ai/live/support-matrix.json'))`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--tools--renewal_glm-py"></a>

## tools/renewal_glm.py

原件：[.scratch/barness-ai-minimax/tools/renewal_glm.py](../../barness-ai-minimax/tools/renewal_glm.py)；SHA256 `33e7df4cb27fa7d86101eb0d894360db07c39f6d97dc9e41b3b1a5739cf98d04`。活动工具 / 仅测试消费，旧真实轮次关闭。现有test_*.py覆盖固定控制器及模块；真实旧轮次仅封存，不是日常编排入口。

**职责**：Fresh unchanged GLM raw archives, independent known usage and support views。

**调用方**：[.scratch/barness-ai-minimax/tools/eleven_inputs.py](../../barness-ai-minimax/tools/eleven_inputs.py) L69（import）；[.scratch/barness-ai-minimax/tools/renewal.py](../../barness-ai-minimax/tools/renewal.py) L139（import）；[.scratch/barness-ai-minimax/tools/repair_inputs.py](../../barness-ai-minimax/tools/repair_inputs.py) L136（import）；[.scratch/barness-ai-minimax/tools/repair_publication.py](../../barness-ai-minimax/tools/repair_publication.py) L36（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L115 `c.dump(out / 'raw-selector-before.json', {n: raw.decode() for n, raw in before.items()})`；L126 `c.dump(out / 'raw-published.json', checkpoint)`；L138 `parent.mkdir(exist_ok=False, mode=448)`；L150 `(out / 'support-views-before.json').write_bytes(u.file_bytes(target))`；L160 `c.dump(target, view)`；L162 `c.dump(out / 'views-published.json', result)`；L117 `c.dump(c.ROOT / RAW_NAMES[0], index)`；L118 `(c.ROOT / RAW_NAMES[1]).write_text(c.digest(c.ROOT / RAW_NAMES[0]) + '\n')`；L119 `c.dump(c.ROOT / RAW_NAMES[2], usage)`；L98 `dest.parent.mkdir(parents=True, exist_ok=True)`；L98 `dest.write_bytes(raw)`；L121 `(c.ROOT / name).write_bytes(raw)`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--tools--repair_anchors-py"></a>

## tools/repair_anchors.py

原件：[.scratch/barness-ai-minimax/tools/repair_anchors.py](../../barness-ai-minimax/tools/repair_anchors.py)；SHA256 `33c279ccc32bdbbbb63c6dfd1fd0af82cf918d40111c9ef0ef224606130f6c42`。活动工具 / 仅测试消费，旧真实轮次关闭。现有test_*.py覆盖固定控制器及模块；真实旧轮次仅封存，不是日常编排入口。

**职责**：Certified original inventories are independent of newly prepared reuse hashes。

**调用方**：[.scratch/barness-ai-minimax/tools/eleven_inputs.py](../../barness-ai-minimax/tools/eleven_inputs.py) L8（import）；[.scratch/barness-ai-minimax/tools/repair_inputs.py](../../barness-ai-minimax/tools/repair_inputs.py) L11（import）；[.scratch/barness-ai-minimax/tools/test_repair_fixtures.py](../../barness-ai-minimax/tools/test_repair_fixtures.py) L9（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--tools--repair_contract-py"></a>

## tools/repair_contract.py

原件：[.scratch/barness-ai-minimax/tools/repair_contract.py](../../barness-ai-minimax/tools/repair_contract.py)；SHA256 `97ce3c09620d8beaef2f0d7ed530feb413ae6c51011bff775be8ec99eeb7cce3`。活动工具 / 仅测试消费，旧真实轮次关闭。现有test_*.py覆盖固定控制器及模块；真实旧轮次仅封存，不是日常编排入口。

**职责**：The single six-process release-repair authority; no arbitrary round or caps。

**调用方**：[.scratch/barness-ai-minimax/tools/eleven_inputs.py](../../barness-ai-minimax/tools/eleven_inputs.py) L10（import）；[.scratch/barness-ai-minimax/tools/release_repair_validation.py](../../barness-ai-minimax/tools/release_repair_validation.py) L10（import）；[.scratch/barness-ai-minimax/tools/renewal_glm.py](../../barness-ai-minimax/tools/renewal_glm.py) L153（import）；[.scratch/barness-ai-minimax/tools/renewal_glm.py](../../barness-ai-minimax/tools/renewal_glm.py) L68（import）；[.scratch/barness-ai-minimax/tools/repair_inputs.py](../../barness-ai-minimax/tools/repair_inputs.py) L134（import）；[.scratch/barness-ai-minimax/tools/repair_publication.py](../../barness-ai-minimax/tools/repair_publication.py) L6（import）；[.scratch/barness-ai-minimax/tools/test_eleven_validation.py](../../barness-ai-minimax/tools/test_eleven_validation.py) L138（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--tools--repair_inputs-py"></a>

## tools/repair_inputs.py

原件：[.scratch/barness-ai-minimax/tools/repair_inputs.py](../../barness-ai-minimax/tools/repair_inputs.py)；SHA256 `d6e38602f14b1edbcf881b0a0262a904a87859b022a55344a5cb44b92577e047`。活动工具 / 仅测试消费，旧真实轮次关闭。现有test_*.py覆盖固定控制器及模块；真实旧轮次仅封存，不是日常编排入口。

**职责**：Reopen immutable native history and prove the precise affected-source boundary。

**调用方**：[.scratch/barness-ai-minimax/tools/eleven_inputs.py](../../barness-ai-minimax/tools/eleven_inputs.py) L9（import）；[.scratch/barness-ai-minimax/tools/release_repair_validation.py](../../barness-ai-minimax/tools/release_repair_validation.py) L11（import）；[.scratch/barness-ai-minimax/tools/test_eleven_validation.py](../../barness-ai-minimax/tools/test_eleven_validation.py) L138（import）；[.scratch/barness-ai-minimax/tools/test_repair_fixtures.py](../../barness-ai-minimax/tools/test_repair_fixtures.py) L9（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--tools--repair_publication-py"></a>

## tools/repair_publication.py

原件：[.scratch/barness-ai-minimax/tools/repair_publication.py](../../barness-ai-minimax/tools/repair_publication.py)；SHA256 `116270bb06c9084e65bd3a42845630bf0e70c22bbd2447baf688cc4cb0eca849`。活动工具 / 仅测试消费，旧真实轮次关闭。现有test_*.py覆盖固定控制器及模块；真实旧轮次仅封存，不是日常编排入口。

**职责**：Only the new fixed round's same-source3PASS canonical bytes may publish data。

**调用方**：[.scratch/barness-ai-minimax/tools/renewal_glm.py](../../barness-ai-minimax/tools/renewal_glm.py) L165（import）；[.scratch/barness-ai-minimax/tools/renewal_glm.py](../../barness-ai-minimax/tools/renewal_glm.py) L169（import）；[.scratch/barness-ai-minimax/tools/renewal_glm.py](../../barness-ai-minimax/tools/renewal_glm.py) L153（import）；[.scratch/barness-ai-minimax/tools/renewal_glm.py](../../barness-ai-minimax/tools/renewal_glm.py) L68（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--tools--routed_contract-py"></a>

## tools/routed_contract.py

原件：[.scratch/barness-ai-minimax/tools/routed_contract.py](../../barness-ai-minimax/tools/routed_contract.py)；SHA256 `b6a8ba6f9daf116d664af30ffcfd96033b0d38f5d62e84f2c318e6584eb52bd8`。活动工具 / 仅测试消费，旧真实轮次关闭。现有test_*.py覆盖固定控制器及模块；真实旧轮次仅封存，不是日常编排入口。

**职责**：Fixed service routing and immutable prior accounting for ticket13。

**调用方**：[.scratch/barness-ai-minimax/tools/repair_contract.py](../../barness-ai-minimax/tools/repair_contract.py) L5（import）；[.scratch/barness-ai-minimax/tools/repair_inputs.py](../../barness-ai-minimax/tools/repair_inputs.py) L8（import）；[.scratch/barness-ai-minimax/tools/routed_evidence.py](../../barness-ai-minimax/tools/routed_evidence.py) L10（import）；[.scratch/barness-ai-minimax/tools/routed_validation.py](../../barness-ai-minimax/tools/routed_validation.py) L11（import）；[.scratch/barness-ai-minimax/tools/test_repair_fixtures.py](../../barness-ai-minimax/tools/test_repair_fixtures.py) L9（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--tools--routed_evidence-py"></a>

## tools/routed_evidence.py

原件：[.scratch/barness-ai-minimax/tools/routed_evidence.py](../../barness-ai-minimax/tools/routed_evidence.py)；SHA256 `ab1eb2247183768680c0d6315357fc9ea7394405b0028775fe8611d63f41f0a0`。活动工具 / 仅测试消费，旧真实轮次关闭。现有test_*.py覆盖固定控制器及模块；真实旧轮次仅封存，不是日常编排入口。

**职责**：Reopen launcher-pinned historical bytes without requiring current HEAD。

**调用方**：[.scratch/barness-ai-minimax/tools/eleven_grant.py](../../barness-ai-minimax/tools/eleven_grant.py) L4（import）；[.scratch/barness-ai-minimax/tools/eleven_inputs.py](../../barness-ai-minimax/tools/eleven_inputs.py) L6（import）；[.scratch/barness-ai-minimax/tools/eleven_validation.py](../../barness-ai-minimax/tools/eleven_validation.py) L15（import）；[.scratch/barness-ai-minimax/tools/paid_checks.py](../../barness-ai-minimax/tools/paid_checks.py) L3（import）；[.scratch/barness-ai-minimax/tools/release_repair_validation.py](../../barness-ai-minimax/tools/release_repair_validation.py) L9（import）；[.scratch/barness-ai-minimax/tools/repair_anchors.py](../../barness-ai-minimax/tools/repair_anchors.py) L4（import）；[.scratch/barness-ai-minimax/tools/repair_inputs.py](../../barness-ai-minimax/tools/repair_inputs.py) L9（import）；[.scratch/barness-ai-minimax/tools/repair_publication.py](../../barness-ai-minimax/tools/repair_publication.py) L5（import）；[.scratch/barness-ai-minimax/tools/routed_validation.py](../../barness-ai-minimax/tools/routed_validation.py) L12（import）；[.scratch/barness-ai-minimax/tools/test_eleven_validation.py](../../barness-ai-minimax/tools/test_eleven_validation.py) L138（import）；[.scratch/barness-ai-minimax/tools/test_repair_fixtures.py](../../barness-ai-minimax/tools/test_repair_fixtures.py) L9（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--tools--routed_validation-py"></a>

## tools/routed_validation.py

原件：[.scratch/barness-ai-minimax/tools/routed_validation.py](../../barness-ai-minimax/tools/routed_validation.py)；SHA256 `d4218ae16c85a1e679a775d78f59870b374d530e65d7f3d7060d61cc95efe4ca`。活动工具 / 仅测试消费，旧真实轮次关闭。现有test_*.py覆盖固定控制器及模块；真实旧轮次仅封存，不是日常编排入口。

**职责**：One finite routed validation, split around committed GLM selector renewal。

**调用方**：[.scratch/barness-ai-minimax/tools/test_routed_validation.py](../../barness-ai-minimax/tools/test_routed_validation.py) L33（literal-reference）；[.scratch/barness-ai-minimax/tools/test_routed_validation.py](../../barness-ai-minimax/tools/test_routed_validation.py) L137（literal-reference）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：`parser.add_argument('mode', choices=('prepare', 'check', 'preview', 'run-selected', 'summarize'))` L151；`parser.add_argument('out')` L152；`parser.add_argument('route', nargs='?')` L152；`parser.add_argument('--phase', choices=('glm', 'legacy'))` L153。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L18 `out.mkdir(parents=True, exist_ok=False, mode=448)`；L19 `c.dump(out / 'source-at-start.json', frozen)`；L20 `c.dump(out / 'plan.json', c.redact(dict(schema=1, round=r.ROUND, phase=phase, budget=r.BUDGET, accountingScope=r.ACCOUNTING_SCOPE, collection=e.location(out), priorReceipts=prior, routes=r.jobs(out, phase), automaticProcessReruns=False, authorizationBasis='explicit continuation request; finite ceiling is not an explicit numeric reply', noSceneSplic`；L35 `c.dump(out / 'binaries.json', {name: c.digest(out / name) for name, _ in builds})`；L36 `c.dump(out / 'prepared.json', dict(state='PREPARED', credentialsRead=False, planSHA256=c.digest(out / 'plan.json'), authorizationFiles={n: c.digest(out / n) for n in auth}))`；L39 `c.dump(out / 'prerelease-source-review.template.json', dict(decision='PENDING', reviewedBy='', **pins, authorizedRoutes=[], credentialScopes={}))`；L41 `c.dump(out / 'execution-authorization.template.json', dict(decision='PENDING', authority='', requestReference='', requestText=r.REQUEST, round=r.ROUND, phase=phase, budget=r.BUDGET, accountingScope=r.ACCOUNTING_SCOPE, **pins, checksSHA256='', sourceReviewSHA256=''))`；L119 `c.dump(out / 'routed-consumption.json', result)`；L127 `c.dump(intent, c.redact(dict(state='STARTING', route=name, baseName=base, validationRound=r.ROUND, phase=c.doc(out / 'plan.json')['phase'], sourceCommit=frozen['sourceCommit'], sourceVariable=job['sourceVariable'], credentialsRead=False, maxCalls=job['maxCalls'], maxAttempts=job['maxAttempts'], networkRoute=job['networkRoute'], collectionRoot=str(o`；L139 `c.dump(intent, record)`；L141 `c.dump(intent, record)`；L144 `c.dump(out / (name + '-intent.json'), record)`；L28 `c.dump(out / name, value)`；L143 `c.dump(intent, record)`；L162 `c.dump(out / 'execution-authorization.template.json', template)`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--tools--test_eleven_validation-py"></a>

## tools/test_eleven_validation.py

原件：[.scratch/barness-ai-minimax/tools/test_eleven_validation.py](../../barness-ai-minimax/tools/test_eleven_validation.py)；SHA256 `fdb15670df9611aaf1a9649b8f08b6e0f5cbee481f6809943aedcfdd4ab99b33`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Public fixed-eleven tool E2E; external executables and keys are synthetic。

**调用方**：[.scratch/barness-ai-minimax/tools](../../barness-ai-minimax/tools) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L22 `code.replace("route=os.environ.get('BARNESS_GLM_ROUTE')", "\nif '-verify' in argv:sys.exit(0)\nif '-formal' in argv and '-out' not in argv:\n bundle=pathlib.Path(argv[-1])\n if '-publish' in argv:(bundle/'live-report.json').write_text('{}')\n sys.exit(0)\nroute=os.environ.get('BARNESS_GLM_ROUTE')")`；L24 `code.replace('print(json.dumps(dict(os.environ)))', "\nif 'BARNESS_MINIMAX_ROUTE' in os.environ:\n for n in ('minimax-integrity.json','minimax-ledger.json'):(bundle/n).write_text('{}')\nprint(json.dumps(dict(os.environ)))")`；L26 `producer.write_text(code)`；L29 `go.write_text(code)`；L35 `(self.root / '.env').write_text('GLM_KEY=SYNTHETIC-GLM\nMINIMAX_KEY=SYNTHETIC-M3\nOPENAI_KEY=SYNTHETIC-OA\nTYPESAFE_KEY=SYNTHETIC-TS\n')`；L54 `key.unlink()`；L54 `key.mkdir()`；L70 `(self.root / '.env').write_text('GLM_KEY=SYNTHETIC-GLM\n')`；L77 `(self.root / '.env').mkdir()`；L98 `(base / 'prerelease-source-review.json').write_text(json.dumps(review))`；L100 `(base / 'execution-authorization.json').write_text(json.dumps(auth))`；L102 `(self.root / '.env').mkdir()`；L110 `(self.root / '.env').write_text('GLM_KEY=SYNTHETIC-GLM\nMINIMAX_KEY=SYNTHETIC-UNSELECTED\n')`；L118 `(self.root / '.env').mkdir()`；L141 `dest.parent.mkdir(parents=True, exist_ok=True)`；L141 `dest.write_bytes(raw)`；L56 `target.write_text(json.dumps(changed))`；L57 `target.write_bytes(raw)`；L108 `auth.write_text(json.dumps(changed))`；L108 `auth.write_bytes(raw)`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--tools--test_glm_corrective-py"></a>

## tools/test_glm_corrective.py

原件：[.scratch/barness-ai-minimax/tools/test_glm_corrective.py](../../barness-ai-minimax/tools/test_glm_corrective.py)；SHA256 `b6a3377dbd79e4c021916f84a4dc1c1f8c5480b24b1732678b6c08f0ea1ca1a8`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Executable corrective-1 boundary; all accounts/captures are synthetic。

**调用方**：[.scratch/barness-ai-minimax/tools](../../barness-ai-minimax/tools) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L28 `(self.bin / 'producer').write_text('#!' + sys.executable + '\nimport json,os,pathlib,subprocess,time\ntime.sleep(0.1)\nr=os.environ["BARNESS_GLM_ROUTE"];p=pathlib.Path(os.environ["BARNESS_AI_EVIDENCE_DIR"])/"session";p.mkdir(parents=True)\nd=dict(schema=1,kind="glm-formal",mode="live",route=r,outcome="PASS",complete=True,auditPassed=True,gitCommit=`；L44 `(base / 'prerelease-source-review.json').write_text(json.dumps(value))`；L51 `(self.root / '.env').mkdir()`；L65 `(self.root / '.env').mkdir()`；L71 `file.write_bytes(raw)`；L72 `receipt.write_text('{}')`；L73 `receipt.write_bytes(old)`；L78 `(self.root / '.env').write_text('GLM_KEY=SYNTHETIC-SELECTED\nMINIMAX_KEY=SYNTHETIC-UNSELECTED\n')`；L105 `(self.root / '.env').unlink()`；L105 `(self.root / '.env').mkdir()`；L111 `(self.root / '.env').write_text('GLM_KEY=SYNTHETIC-SELECTED\n')`；L117 `report.unlink()`；L119 `marker.write_text(json.dumps(value))`；L130 `p.write_text(json.dumps(value))`；L131 `(self.root / '.env').mkdir()`；L135 `(self.root / '.env').mkdir()`；L136 `marker.write_text('{}')`；L24 `(self.root / '.evidence/minimax-launches' / ('glm-' + route + '.json')).write_text(json.dumps(value))`；L48 `(base / 'human-approval.json').write_text(json.dumps(value))`；L68 `file.write_text(json.dumps(altered))`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--tools--test_paid-py"></a>

## tools/test_paid.py

原件：[.scratch/barness-ai-minimax/tools/test_paid.py](../../barness-ai-minimax/tools/test_paid.py)；SHA256 `0bcce1eb02e3ae013b017b026ce644cf86e64e87f1d55e46f4ca2a051a770023`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Controlled executable-boundary checks; credentials are synthetic fixtures。

**调用方**：[.scratch/barness-ai-minimax/tools/test_glm_corrective.py](../../barness-ai-minimax/tools/test_glm_corrective.py) L9（import）；[.scratch/barness-ai-minimax/tools/test_routed_validation.py](../../barness-ai-minimax/tools/test_routed_validation.py) L11（import）；[.scratch/barness-ai-minimax/tools](../../barness-ai-minimax/tools) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L72 `shutil.copytree(TOOLS, self.tools, ignore=shutil.ignore_patterns('__pycache__'))`；L73 `(self.root / 'ai/live').mkdir(parents=True)`；L74 `(self.root / 'ai/release').mkdir()`；L75 `(self.root / 'ai/runtime.go').write_text('package ai\n')`；L76 `(self.root / 'go.mod').write_text('module controlled\n')`；L77 `(self.root / 'go.sum').write_text('')`；L79 `(self.root / 'ai/release/catalog-snapshot.json').write_text('{"hash":"sha256:controlled"}')`；L81 `self.bin.mkdir()`；L82 `(self.bin / 'go').write_text('#!/bin/sh\nif [ "$1" = test ] || [ "$1" = build ]; then\n  while [ "$#" -gt 0 ]; do\n    if [ "$1" = -o ]; then shift; cp "' + str(self.bin / 'producer') + '" "$1"; chmod +x "$1"; exit 0; fi\n    shift\n  done\nfi\nexit 0\n')`；L84 `(self.bin / 'go').chmod(493)`；L85 `(self.bin / 'producer').write_text('#!' + sys.executable + '\nimport json,os\nprint(json.dumps(dict(os.environ)))\n')`；L94 `(base / 'minimax-launches').mkdir(parents=True, exist_ok=True)`；L95 `(base / 'minimax-launches/mm-research-m3-chat.json').write_bytes(CN_INTENT)`；L96 `(base / 'minimax-stages').mkdir(parents=True, exist_ok=True)`；L97 `(base / 'minimax-stages/candidate.jsonl').write_bytes(CN_PREFIX)`；L110 `(self.root / '.env').mkdir()`；L126 `(base / 'prerelease-source-review.json').write_text(json.dumps(review))`；L134 `(self.bin / 'producer').write_text(code)`；L139 `(self.root / '.env').write_text('MINIMAX_KEY=SELECTED-CONTROLLED-SECRET\nOPENAI_KEY=UNSELECTED-CONTROLLED-SECRET\n')`；L152 `path.write_text(json.dumps({'outcome': 'UNEXPECTED_STATUS_SENTINEL', 'budget': {'calls': 3, 'attempts': 3, 'private': 'REPORT_SECRET'}}))`；L186 `(self.root / '.env').unlink()`；L186 `(self.root / '.env').mkdir()`；L197 `(self.root / '.env').unlink()`；L197 `(self.root / '.env').mkdir()`；L204 `index.parent.mkdir(parents=True)`；L205 `index.write_text('{"schema":1,"entries":{}}')`；L207 `index.write_text('{"schema":1,"entries":{"curated-verifier-only":{}}}')`；L213 `source.parent.mkdir(parents=True)`；L214 `source.write_text('Synthetic fixed USD rate document\n')`；L219 `(self.root / '.env').mkdir()`；L220 `source.write_text('Changed after reviewed binary build\n')`；L229 `runtime.write_text('package changed\n')`；L231 `runtime.write_text('package ai\n')`；L232 `extra.write_text('package ai\n')`；L233 `extra.unlink()`；L234 `binary.write_bytes(original + b'# altered\n')`；L235 `binary.write_bytes(original)`；L237 `review.write_text(json.dumps(data))`；L245 `base.mkdir(parents=True)`；L246 `(base / 'minimax-report.json').write_text(json.dumps({'outcome': 'PASS', 'budget': {'calls': 20, 'attempts': 20}}))`；L247 `(self.root / 'out/mm-global-candidate-m3-chat-command.json').write_text('{"state":"COMPLETED"}')`；L248 `(self.root / '.evidence/minimax-launches/mm-global-candidate-m3-chat.json').write_text('{"state":"COMPLETED"}')`；L254 `receipt.write_text(json.dumps(data))`；L256 `intent.write_text(json.dumps(data))`；L262 `(self.root / '.env').mkdir()`；L282 `marker.write_text(json.dumps({'state': 'COMPLETED', 'reportedFileSha256': hashlib.sha256(report.read_bytes()).hexdigest()}))`；L283 `report.write_text(json.dumps(value))`；L289 `(self.root / '.env').mkdir()`；L311 `(base / 'support').mkdir()`；L311 `(base / 'support/live-report.json').write_text('{}')`；L325 `(self.bin / 'go').write_text('#!' + sys.executable + '\nimport json,sys\nfrom pathlib import Path\nargs=sys.argv[1:]\nwith Path(' + repr(str(log)) + ').open("a") as stream:stream.write(json.dumps(args)+"\\n")\nif "./ai/live/cmd/jevformal" in args and "-out" in args:\n    target=Path(args[args.index("-out")+1]);target.mkdir()\n')`；L343 `(self.root / '.env').mkdir()`；L357 `review.write_text(json.dumps(value))`；L362 `review.write_text(json.dumps(value))`；L386 `(self.root / '.env').mkdir()`；L404 `(self.root / '.env').mkdir()`；L405 `base.mkdir(parents=True)`；L407 `(self.root / 'out/mm-global-candidate-m3-chat-command.json').write_text('{"state":"COMPLETED"}')`；L408 `marker.write_text('{"state":"COMPLETED"}')`；L417 `stale.mkdir(parents=True)`；L418 `(stale / 'minimax-report.json').write_text('{"outcome":"PASS","budget":{"calls":0,"attempts":0}}')`；L419 `(self.root / '.env').unlink()`；L419 `(self.root / '.env').mkdir()`；L270 `target.parent.mkdir(parents=True, exist_ok=True)`；L270 `target.write_text(json.dumps(report))`；L271 `(self.root / 'out' / (route + '-command.json')).write_text('{"state":"COMPLETED"}')`；L272 `marker.parent.mkdir(parents=True, exist_ok=True)`；L272 `marker.write_text('{"state":"COMPLETED"}')`；L306 `path.mkdir(parents=True)`；L306 `(path / 'live-report.json').write_text('{}')`；L309 `path.mkdir(parents=True)`；L412 `report.write_text(json.dumps({'outcome': 'PASS', 'budget': budget}))`；L310 `(path / leaf).write_text('{}')`；L392 `intent.write_text(json.dumps(value))`；L393 `ledger.write_bytes(CN_PREFIX[:-1])`；L394 `intent.unlink()`；L395 `intent.unlink()`；L395 `ledger.unlink()`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L12 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--tools--test_release_repair-py"></a>

## tools/test_release_repair.py

原件：[.scratch/barness-ai-minimax/tools/test_release_repair.py](../../barness-ai-minimax/tools/test_release_repair.py)；SHA256 `20ed2ab31668c85c4390e67e917c754a212bd2906e3259ec7824ded16232cb5e`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：New finite authority public CLI failure inventory before implementation.  Seam: isolated subprocesses, fixed native DTOs and canonical root-owned receipts. No producer/vendor request or actual credential is needed for planning/preflight。

**调用方**：[.scratch/barness-ai-minimax/tools](../../barness-ai-minimax/tools) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L11 `(self.root / '.env').mkdir()`；L36 `(self.root / '.env').mkdir()`；L49 `target.write_bytes(original + b'\n')`；L57 `(self.root / '.env').mkdir()`；L72 `(self.root / '.env').mkdir()`；L78 `(self.root / '.env').write_text('GLM_KEY=SYNTHETIC-GLM\nTYPESAFE_KEY=SYNTHETIC-TS\nOPENAI_KEY=SYNTHETIC-OA\nMINIMAX_KEY=SYNTHETIC-UNSELECTED\n')`；L83 `chat.unlink()`；L88 `chat.write_bytes(original)`；L92 `(self.root / '.env').unlink()`；L92 `(self.root / '.env').mkdir()`；L98 `(self.root / '.env').mkdir()`；L108 `receipt.write_text(json.dumps(v))`；L109 `(base / 'checks.json').write_text(json.dumps(checks))`；L114 `(self.root / '.env').mkdir()`；L120 `(self.root / '.env').mkdir()`；L131 `matrix.write_text(json.dumps(value))`；L132 `(self.root / '.env').mkdir()`；L77 `auth.write_text(json.dumps(changed))`；L77 `auth.write_bytes(raw)`；L115 `(bundle / ('empty-' + str(n))).mkdir()`；L48 `capture.write_bytes(raw)`；L102 `(self.root / 'fail-check-name').write_text(name)`；L106 `(self.root / 'fail-check-name').unlink()`；L127 `path.write_bytes(raw)`；L47 `capture.unlink()`；L126 `path.write_text(json.dumps(value))`；L41 `capture.write_bytes(raw + b' ')`；L42 `capture.unlink()`；L44 `copy.write_bytes(raw)`；L44 `capture.unlink()`；L44 `capture.symlink_to(copy)`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-minimax--tools--test_renewal-py"></a>

## tools/test_renewal.py

原件：[.scratch/barness-ai-minimax/tools/test_renewal.py](../../barness-ai-minimax/tools/test_renewal.py)；SHA256 `a28d8ba1d389f7fc96c42d83b89dc01f2f269a7540aa6b9d3b90dc391819dc30`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Public executable tests for keyless renewal; all captures are synthetic。

**调用方**：[.scratch/barness-ai-minimax/tools](../../barness-ai-minimax/tools) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L19 `shutil.copytree(REAL, self.tools, ignore=shutil.ignore_patterns('__pycache__'))`；L36 `self.bin.mkdir()`；L37 `self.write('fake-bin/go', ('#!' + sys.executable + '\nimport sys,os,json,pathlib,shutil\na=sys.argv[1:]\nif "-o" in a:\n p=pathlib.Path(a[a.index("-o")+1]);shutil.copy(pathlib.Path(__file__).with_name("producer"),p);p.chmod(0o755)\nelif "-write-snapshot" in a:\n p=pathlib.Path("ai/release/catalog-snapshot.json");p.write_text(json.dumps({"hash":"sha`；L41 `(self.bin / 'go').chmod(493)`；L42 `self.write('fake-bin/producer', ('#!' + sys.executable + '\nimport os,json,pathlib,subprocess\nr=os.environ["BARNESS_MINIMAX_ROUTE"];p=pathlib.Path(os.environ["BARNESS_AI_EVIDENCE_DIR"])/"session";p.mkdir(parents=True)\nd={"outcome":"NOT_RUN","complete":False,"mode":"live","stage":"candidate","identity":{"route":r,"model":"MiniMax-M3.1-Flash-Previe`；L46 `(self.root / '.env').mkdir()`；L50 `p.parent.mkdir(parents=True, exist_ok=True)`；L50 `p.write_bytes(raw)`；L94 `old.write_bytes(b'IMMUTABLE-HISTORY')`；L126 `mixed.write_bytes((self.root / '.evidence/minimax-launches/glm-messages.json').read_bytes())`；L195 `marker.unlink()`；L197 `marker.write_bytes(raw)`；L197 `marker.write_text(json.dumps(value))`；L208 `report.write_text(json.dumps(value))`；L223 `(self.bin / 'go').chmod(493)`；L226 `shutil.copytree(self.root / out, seed)`；L254 `target.symlink_to(self.root / 'go.mod')`；L259 `target.unlink()`；L259 `p.write_text(json.dumps(r))`；L321 `producer.write_text(producer.read_text().replace('https://api.minimax.io/', 'https://api.minimax.cn/'))`；L336 `(base / 'nested/a/b/c').mkdir(parents=True)`；L21 `(self.root / p).mkdir(parents=True)`；L37 `('#!' + sys.executable + '\nimport sys,os,json,pathlib,shutil\na=sys.argv[1:]\nif "-o" in a:\n p=pathlib.Path(a[a.index("-o")+1]);shutil.copy(pathlib.Path(__file__).with_name("producer"),p);p.chmod(0o755)\nelif "-write-snapshot" in a:\n p=pathlib.Path("ai/release/catalog-snapshot.json");p.write_text(json.dumps({"hash":"sha256:new","catalog":{"versi`；L42 `('#!' + sys.executable + '\nimport os,json,pathlib,subprocess\nr=os.environ["BARNESS_MINIMAX_ROUTE"];p=pathlib.Path(os.environ["BARNESS_AI_EVIDENCE_DIR"])/"session";p.mkdir(parents=True)\nd={"outcome":"NOT_RUN","complete":False,"mode":"live","stage":"candidate","identity":{"route":r,"model":"MiniMax-M3.1-Flash-Preview","purpose":"mplan","api":{"cha`；L148 `path.write_text(json.dumps(value))`；L151 `path.write_bytes(original)`；L206 `receipt.write_text(json.dumps(value))`；L207 `receipt.write_bytes(original)`；L330 `(base / f'empty-{n}').mkdir()`；L366 `receipt.write_bytes(old_receipt)`；L383 `(parent / f'empty-{n}').mkdir()`；L231 `shutil.copytree(seed, self.root / clone)`；L248 `usage.write_bytes(original)`；L355 `p.unlink()`；L365 `(base / n).write_bytes(raw)`；L368 `(base / 'formal-glm-report.json').write_text(json.dumps(formal))`；L369 `(base / 'live-report.json').write_text(json.dumps(support))`；L370 `(base / 'glm-integrity.json').write_text(json.dumps(seal))`；L233 `usage.write_text(json.dumps(value))`；L250 `index_path.write_bytes(index_raw)`；L250 `(index_path.parent / 'live-captures.sha256').write_text(hashlib.sha256(index_raw).hexdigest() + '\n')`；L250 `archive.write_bytes(archive_raw)`；L346 `p.write_bytes(b'')`；L372 `receipt.write_text(json.dumps(value))`；L244 `index_path.write_text(json.dumps(index))`；L244 `(index_path.parent / 'live-captures.sha256').write_text(hashlib.sha256(index_path.read_bytes()).hexdigest() + '\n')`；L245 `usage.write_text(json.dumps(value))`；L373 `(base / 'live-report.json').write_text('{"raw":"changed"}')`；L374 `(base / 'glm-execution-source.json').write_text('{"controlled-source":"edited"}')`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L13 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--tools--test_repair_fixtures-py"></a>

## tools/test_repair_fixtures.py

原件：[.scratch/barness-ai-minimax/tools/test_repair_fixtures.py](../../barness-ai-minimax/tools/test_repair_fixtures.py)；SHA256 `b61a79b10f307ad9f7edd5715c0fe7718d52c1f7ffafb527fa1d6b697d91d522`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Synthetic immutable native history in an isolated local Git clone; no credentials。

**调用方**：[.scratch/barness-ai-minimax/tools/test_eleven_validation.py](../../barness-ai-minimax/tools/test_eleven_validation.py) L4（import）；[.scratch/barness-ai-minimax/tools/test_release_repair.py](../../barness-ai-minimax/tools/test_release_repair.py) L7（import）；[.scratch/barness-ai-minimax/tools](../../barness-ai-minimax/tools) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L14 `target.parent.mkdir(parents=True, exist_ok=True)`；L14 `target.write_bytes(raw)`；L37 `shutil.copytree(TOOLS, self.tools, dirs_exist_ok=True, ignore=shutil.ignore_patterns('__pycache__'))`；L40 `self.bin.mkdir()`；L40 `(self.bin / 'go').write_text('#!' + sys.executable + '\nimport pathlib,sys,os,json,shutil\na=sys.argv[1:];r=pathlib.Path(' + repr(str(self.root)) + ')\nwith (r/"go-probe.jsonl").open("a") as f:f.write(json.dumps(dict(argv=a,environmentKeys=sorted(os.environ)))+"\\n")\nif "-o" in a:\n p=pathlib.Path(a[a.index("-o")+1]);shutil.copyfile(r/".fake-bin/p`；L40 `(self.bin / 'go').chmod(493)`；L41 `(self.bin / 'producer').write_text('#!' + sys.executable + '\nprint("synthetic-no-selected-job")\n')`；L47 `(self.bin / 'producer').write_text('#!' + sys.executable + '\n' + "import json,os,pathlib,subprocess,sys,time\nroot=pathlib.Path.cwd()\nwhile not (root/'go.mod').is_file():root=root.parent\nargv=sys.argv[1:];route=os.environ.get('BARNESS_GLM_ROUTE');failed=False\nif route:\n base='glm-'+route;bundle=pathlib.Path(os.environ['BARNESS_AI_EVIDENCE_DIR'`；L52 `(base / 'prerelease-source-review.json').write_text(json.dumps(review))`；L54 `(base / 'execution-authorization.json').write_text(json.dumps(auth))`；L64 `target.mkdir(parents=True, exist_ok=False)`；L83 `(target / 'public-cli-journal.json').write_text(json.dumps(self.journal, indent=2) + '\n')`；L84 `(target / 'inventory.json').write_text(json.dumps(dict(files=inventory, bytes=total, credentials='synthetic test values only; no .env copied', paidVendorCalls=0), indent=2) + '\n')`；L71 `dest.parent.mkdir(parents=True, exist_ok=True)`；L71 `dest.write_bytes(raw)`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L4 `Path(__file__).resolve()`。

<a id="barness-ai-minimax--tools--test_routed_validation-py"></a>

## tools/test_routed_validation.py

原件：[.scratch/barness-ai-minimax/tools/test_routed_validation.py](../../barness-ai-minimax/tools/test_routed_validation.py)；SHA256 `1d8c0bab92234d6343051635d61b3a49f600e26780608ed76cb66aa383f094be`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Public routed-launcher CLI checks; synthetic accounts and local processes。

**调用方**：[.scratch/barness-ai-minimax/tools](../../barness-ai-minimax/tools) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L36 `(self.bin / 'producer').write_text('#!' + sys.executable + '\nimport json,os,pathlib,subprocess,time\ntime.sleep(0.1)\nr=os.environ["BARNESS_GLM_ROUTE"];p=pathlib.Path(os.environ["BARNESS_AI_EVIDENCE_DIR"])/"session";p.mkdir(parents=True)\nd=dict(schema=1,kind="glm-formal",mode="live",route=r,outcome="PASS",complete=True,auditPassed=True,gitCommit=`；L53 `review.write_text(json.dumps(value))`；L57 `(base / 'execution-authorization.json').write_text(json.dumps(value))`；L60 `(self.root / '.env').mkdir()`；L86 `(self.root / '.env').mkdir()`；L103 `(self.root / '.env').write_text('GLM_KEY=SYNTHETIC-SELECTED\nMINIMAX_KEY=SYNTHETIC-UNSELECTED\n')`；L116 `(self.root / '.env').write_text('GLM_KEY=SYNTHETIC-SELECTED\n')`；L124 `report.write_bytes(raw + b' ')`；L127 `report.write_bytes(raw)`；L129 `(self.root / '.env').unlink()`；L129 `(self.root / '.env').mkdir()`；L153 `(self.root / '.env').mkdir()`；L155 `source.write_text('{"fresh":true}')`；L168 `(self.root / '.env').write_text('GLM_KEY=SYNTHETIC-SELECTED\n')`；L170 `source.write_text('{"fresh":true}')`；L192 `path.write_bytes(b'x' * ((16 << 20) + 1))`；L194 `path.unlink()`；L196 `path.unlink()`；L196 `path.write_bytes(raw)`；L198 `shutil.copytree(base, self.root / 'copied-proof')`；L225 `(self.bin / 'go').write_text(code)`；L229 `(self.root / '.env').mkdir()`；L249 `(self.root / '.env').mkdir()`；L250 `(self.root / marker).write_text('controlled public command failure')`；L257 `auth.write_text(json.dumps(value))`；L32 `(self.root / '.evidence/minimax-launches' / (name + '.json')).write_text(json.dumps(value))`；L97 `path.write_text(json.dumps(doc))`；L98 `path.write_bytes(raw)`；L101 `path.write_bytes(raw + b' ')`；L102 `path.write_bytes(raw)`；L203 `intent.write_text(json.dumps(doc))`；L205 `intent.write_bytes(original)`；L190 `path.write_bytes(raw)`；L183 `path.write_bytes(raw + b' ')`；L189 `path.unlink()`；L184 `path.unlink()`；L186 `target.write_bytes(raw)`；L186 `path.unlink()`；L186 `path.symlink_to(target)`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-minimax/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。
