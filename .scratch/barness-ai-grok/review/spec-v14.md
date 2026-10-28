## Spec

只读复核完整差异 `1992e875136ebc332a19255e3729c448b001894b...d816aacb4538f735b76df7bdacb686b621aef382`，对照原 Spec 与工单 18/19；未测试、构建、启动 native、访问厂商或操作 claim。

**0 项发现；最高级别：无。v13 P2 已关闭。** 未发现需求缺失、范围扩大或实现错误。

工单 [19:20](/Users/cyber/.codex/worktrees/grok-native-proof/barness/.scratch/barness-ai-grok/issues/19-controlled-grant-store-isolation.md:20) 要求保留“真实执行状态”。[重启接线:66](/Users/cyber/.codex/worktrees/grok-native-proof/barness/ai/e2e/grok_history_minimax_grant_test.go:66) 现在新建 `exec.Cmd`，逐项保持原 Path/Args/Dir/Env，沿用原私有 namespace；[检查:165](/Users/cyber/.codex/worktrees/grok-native-proof/barness/ai/e2e/minimax_fresh_ledger_fixture_test.go:165) 要求不同 PID、`*exec.ExitError` 与实际数字退出 1；:185–186 再检查新 session、精确原生 consumed-approval 拒绝和零活动。

实际 RED `202218...r83ydftq` 三条均因旧本地 `Stdout already set` 失败；实际 GREEN `202256...3rrlixrk` 六项 PASS/audit0。逐一读取 Chat/Responses/Messages 新原件：第二 PID 分别为 22978/23006/23034，均不同于首进程；各有独立新 runID、原生 FAIL/精确拒绝、0 calls/HTTP/预留、空 case 清单，无新 grant/claim 原件。首尾私有 store 的四文件名称及全部 digest 相等；首进程仍完整 14 场景。证明来自真实第二 native 进程，而非本地 Go 错误或新 namespace。

v2 明确撤回 v1 重启断言，v1 摘要及旧 raw 哈希保持原值。索引 88 个引用中，按独立性要求未读取另一轴报告；其余 87 个引用哈希有效。GREEN 的 2062 个来源文件逐项与当前字节相等，受测提交标签如实为 `7e76c923...`。

相对 v13 仅两份既有 E2E 文件及新摘要变化；工单 18 的 M3 thinking-on 任务、正例及零 thinking 拒绝、14 场景/预算/消费者均未改。受控结果未取得正式资格；完整普通、重新冻结和付费/人工阶段 pending 不计为问题。
