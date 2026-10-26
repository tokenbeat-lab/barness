## Spec

只读复核固定完整差异 `4c9cd7c6...b77859c0a7b94b7ef874cd8a52196f544d0e01e1`，对照工单 17、Spec R23/E18 与验证 SOP。未运行测试、构建或厂商调用。

**0 项发现；最高级别：无。** 未发现需求缺失、范围扩大或实现错误。v11 的 P2 已关闭。

工单要求“通过公共CLI建立完整OpenAI Chat native fixture的RED”。新增 [fixture:44](/Users/cyber/RestoX/barness/ai/e2e/generic_unsupported_native_fixture_test.go:44) 构造完整 11 场景、10 PASS/1 UNSUPPORTED、完整捕获与一致预算；[CLI:104](/Users/cyber/RestoX/barness/ai/e2e/generic_unsupported_cli_test.go:104) 正例精确到达 `live qualification stale or consumer receipt missing`。该错误位于 [record_live:93](/Users/cyber/RestoX/barness/ai/internal/testkit/release/record_live.go:93)，证明 `qualifiedLive` 中的报告、库存、audit、Merge、目录和 SDK 检查均已通过。三个留存 RED 的原始日志分别证明恢复全 PASS 规则、删除报告接线、删除库存接线都会失败，补齐了原先仅断言统一拒绝的缺口。

reasoning-history 仍为精确既有原因的 UNSUPPORTED、零 calls/attempts/retries、仅 assertions.json；其余必交场景仍须 PASS，mandatory xAI history 的 NOT_RUN/UNSUPPORTED 不获放行，既有矩阵检查保留无新增 LastPassedAt。

合成 grant/receipt 只静态写入 fixture 自有 bare Git namespace（[fixture:70](/Users/cyber/RestoX/barness/ai/e2e/generic_unsupported_native_fixture_test.go:70)），`GIT_DIR` 仅传给 releasegate 子进程（[CLI:78](/Users/cyber/RestoX/barness/ai/e2e/generic_unsupported_cli_test.go:78)）；无 Claim/真实 store 写入或厂商调用，missing-namespace 反例仍被拒绝。生产代码相对 268faeba 无新增变化。

只读核对 summary 的 135 个文件引用，哈希全部一致；最终留存为 61 PASS、audit 0。所有公开消费者命令仍 exit 1、总 release.pass=false，empty source/缺 consumer 均明确保留。完整普通验收、最终冻结、后续付费及人工阶段的 pending 不计为 finding。
