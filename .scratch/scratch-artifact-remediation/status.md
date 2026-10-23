# 普通测试输入治理当前状态

日期：2026-10-24（Asia/Shanghai）。当前工单01：IN_PROGRESS。

- 基线 HEAD：571ca112859211ca5d987906e16ae135f89f54f5。原有未跟踪 SOP 回归目录及审查目录保持。
- 影响：仅固定普通测试输入及其读取路径。公共 seam 为现有 Go E2E/CLI/证据消费者。
- RED 初选：TestJevBaselineEvidence、TestDecisionsEvidence（含 Snapshot/ManifestBudget）、TestReleaseRollbackArtifacts、TestReleaseRevocationSemanticIntegrity。
- 普通交付范围：完整 ai/e2e、ai；测试证据消费依赖涉及 release/supportmatrix 时覆盖对应包，最终范围随依赖清单固定。
- 来源：ai/、SOP工具、go.mod/go.sum；保守固定 Git 跟踪的历史 scratch 输入，用户未跟踪文档不作为测试输入。新testdata纳入来源选择器。
- live：计划 NOT_APPLICABLE；不改厂商运行行为、型号或正式支持资格，最终以实际 diff 复核。
- race：NOT_RUN；候选为受影响 Go 包对应 -race 命令，须另行给出具体来源/用例与人类选择。
- 压力：计划 NOT_APPLICABLE；不改运行时资源策略或设计负载，普通消费者只读固定夹具。
- 付费调用/真实清理：NOT_RUN。release.pass=false；未取得普通验收和人工race事实。
- 验证、实际受测来源、命令与原件指针随进展追加，失败原件不覆盖。

## 初次 RED/GREEN

- 干净独立 clone 为基线HEAD，仅 Git 内容，工作区开始干净，无 .env/本机忽略 ZIP。
- RED：.evidence/scratch-artifact-remediation/red-clean/.evidence/scratch-artifact-remediation/e2e/20261023T222133.479214Z-yv2tja9a/receipt.json；实际exit1，46个原生场景为45FAIL/1PASS，audit0。
- 初次定向GREEN：.evidence/scratch-artifact-remediation/e2e/20261023T222316.990694Z-7ufuosop/receipt.json；实际exit0，46/46 PASS，audit0，墙钟约10秒。未增加或删除场景，两个实际ID集合完全相等。
- 实现仅三处Go测试输入读取与新固定夹具：两份原字节ZIP/manifest及536文件的六场景runtime派生bundle；历史8757份跟踪输入哈希未变。
- 最终普通范围固定为 ai/e2e、ai、release、releasegate、supportmatrix 五包。其余scratch传递依赖保守固定8757份现有跟踪原件，新夹具由ai来源集合覆盖。最终干净工作区及完整普通验收仍待执行。

- 输入独立核验：`.evidence/scratch-artifact-remediation/e2e/input-audit-qnvshf7a/summary.json`；构建auditor、Jev原生verifier --audit、Decisions原生verifier --audit、runtime临时副本脱敏审计四命令均exit0。固定夹具没有被审计器写回。
- 新固定输入共544文件、10,201,369字节；536文件runtime bundle仅含原六场景（1,289,208字节），两个原ZIP/manifest按原始哈希逐字节保留。
- 已冻结46个定向必交case及五包最终普通argv，见 `.evidence/scratch-artifact-remediation/planning/selection.json`。准备提交候选并从全新干净clone执行，尚不声明最终通过。

## 干净全包发现的追加输入缺口

- 候选640476c的完整五包实际FAIL，首个失败 TestGLMFormalReports/chat 为 immutable archive unavailable，原件在 green-clean/.evidence/scratch-artifact-remediation/e2e/20261023T222503.230183Z-9fwf2u8f。其他三个实际测试包已PASS；不能把部分包成功写成全量通过。
- GLM四组公共进程定向RED另存20261023T222711.380852Z-lqtkxntt；读取的三raw/三support-view ZIP原本只有宿主.evidence。
- 将六份原字节ZIP按原SHA放入 ai/live/testdata/glm-formal/archives，消费者及测试helper直接按固定digest读取；不回填宿主路径，不改既有原生index/pin/已测报告。该消费者源码变化保留旧来源拒绝语义，不刷新正式资格。新的定向GREEN与干净最终全包仍待取得。

- GLM定向GREEN：20261023T222744.357964Z-shw5hgib，实际exit0；12/12原生case PASS，audit0；相同ID集合的独立定向RED已保留。最终必交case共58项，另要求完整五包和全部已发现E2E顶层测试实际执行。
