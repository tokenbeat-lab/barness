# Ticket17独立两轴复核 v12

完整固定差分`4c9cd7c6...b77859c0a7b94b7ef874cd8a52196f544d0e01e1`，生产268faeba与测试补齐b77859c0。完整11-scene原生包公共CLI及三份真实RED证明原故障与两条生产接线被锁定；合成总资格保持false，未产生真实授权/claim/厂商调用。审查者各自只读。

## Standards

## Standards

Reviewed the complete fixed diff `4c9cd7c6...b77859c0a7b94b7ef874cd8a52196f544d0e01e1` against `AGENTS.md`, `docs/agents/verification.md`, `docs/agents/issue-tracker.md`, ADR-0051/0053, and the labelled smell baseline. Read-only review; no tests, builds, vendor requests, or repository mutations.

**No actionable standards findings. Highest severity: none.**

**v11 P2 is substantively closed.** The new public compiled `releasegate -record` regression supplies all 11 fixed OpenAI Chat scenarios, ten captured calls, matching budgets, and the assertions-only zero-activity `UNSUPPORTED` case. The positive reaches exactly `live qualification stale or consumer receipt missing`, which occurs after successful native artifact/report qualification; total `release.pass` remains false. Negative cases reject mandatory NOT_RUN, altered unsupported reasons, hidden captures, changed inventory bytes, and absent authority namespace.

Three retained actual RED runs independently fail when restoring the all-PASS rule, removing report-validator wiring, or removing unsupported-artifact-validator wiring. Their stdout records the expected incorrect boundary, followed by the retained final GREEN: exit 0, 61 PASS cases, audit findings 0, elapsed 7.719 seconds. All 135 indexed evidence references match their hashes; all 2057 frozen execution-source entries match the fixed commit, including final test bodies.

Authority isolation is concrete: each fixture seeds only its fresh bare `TEST_ONLY.git` with exact compact id/scope receipt bytes. The consumer receives this path through child-only `GIT_DIR`; `xaiProcessEnv` excludes inherited Git variables and credentials. The native verifier resolves that namespace and only reads its receipts; the test never invokes `Claim` or a live producer. Controlled history uses the existing loopback Client and registered cleanup. No production authorization bypass or live qualification is added.

The changed Go test files are 121, 168, and 205 lines. Existing fixture/source helpers are reused; the separate Chat history helper documents its encoding and independent evolution. No actionable baseline smells identified. Root-owned full ordinary acceptance and paid/manual stages remain separate pending work, not findings.

**Count: 0; highest severity: none; v11 P2 closed.**

## Spec

## Spec

只读复核固定完整差异 `4c9cd7c6...b77859c0a7b94b7ef874cd8a52196f544d0e01e1`，对照工单 17、Spec R23/E18 与验证 SOP。未运行测试、构建或厂商调用。

**0 项发现；最高级别：无。** 未发现需求缺失、范围扩大或实现错误。v11 的 P2 已关闭。

工单要求“通过公共CLI建立完整OpenAI Chat native fixture的RED”。新增 [fixture:44](/Users/cyber/RestoX/barness/ai/e2e/generic_unsupported_native_fixture_test.go:44) 构造完整 11 场景、10 PASS/1 UNSUPPORTED、完整捕获与一致预算；[CLI:104](/Users/cyber/RestoX/barness/ai/e2e/generic_unsupported_cli_test.go:104) 正例精确到达 `live qualification stale or consumer receipt missing`。该错误位于 [record_live:93](/Users/cyber/RestoX/barness/ai/internal/testkit/release/record_live.go:93)，证明 `qualifiedLive` 中的报告、库存、audit、Merge、目录和 SDK 检查均已通过。三个留存 RED 的原始日志分别证明恢复全 PASS 规则、删除报告接线、删除库存接线都会失败，补齐了原先仅断言统一拒绝的缺口。

reasoning-history 仍为精确既有原因的 UNSUPPORTED、零 calls/attempts/retries、仅 assertions.json；其余必交场景仍须 PASS，mandatory xAI history 的 NOT_RUN/UNSUPPORTED 不获放行，既有矩阵检查保留无新增 LastPassedAt。

合成 grant/receipt 只静态写入 fixture 自有 bare Git namespace（[fixture:70](/Users/cyber/RestoX/barness/ai/e2e/generic_unsupported_native_fixture_test.go:70)），`GIT_DIR` 仅传给 releasegate 子进程（[CLI:78](/Users/cyber/RestoX/barness/ai/e2e/generic_unsupported_cli_test.go:78)）；无 Claim/真实 store 写入或厂商调用，missing-namespace 反例仍被拒绝。生产代码相对 268faeba 无新增变化。

只读核对 summary 的 135 个文件引用，哈希全部一致；最终留存为 61 PASS、audit 0。所有公开消费者命令仍 exit 1、总 release.pass=false，empty source/缺 consumer 均明确保留。完整普通验收、最终冻结、后续付费及人工阶段的 pending 不计为 finding。

Standards 0项（最高级别无）；Spec 0项（最高级别无）。v11两轴P2均实质关闭，完整普通/新source资格仍须独立实际执行。
