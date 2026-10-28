# Issue18/19 两轴复核 v13

Fixed base: 1992e875136ebc332a19255e3729c448b001894b
Reviewed pin: cd91dbfa257692b744ba92708d0770522561ba07

## Standards

# Standards review v13

Base: `1992e875136ebc332a19255e3729c448b001894b`  
Head: `cd91dbfa257692b744ba92708d0770522561ba07`  
Command: `git diff 1992e875136ebc332a19255e3729c448b001894b...HEAD` in `/Users/cyber/.codex/worktrees/grok-native-proof/barness`.

Reviewed the ten-file delta against current canonical AGENTS.md, verification SOP, issue/domain conventions, ADR-0051/0053 and the supplied smell baseline. Read-only review; no test, build, native process, vendor request or claim executed.

## Documented-standard finding

- **[P2] Restart evidence never starts a second native process.** [grok_history_minimax_grant_test.go:64](/Users/cyber/.codex/worktrees/grok-native-proof/barness/ai/e2e/grok_history_minimax_grant_test.go:64): the new `restartedStore` / `maps.Equal` assertions follow a second `cmd.CombinedOutput()` on the already-run command (line 61). Go forbids reusing `exec.Cmd`; its `CombinedOutput` returns `exec: Stdout already set` before starting any process (`/usr/local/go/src/os/exec/exec.go:1038`). Thus `e != nil` passes without exercising native one-use rejection, and an unchanged store is inevitable. The retained 200652 attempt's three `restart.json` files show only empty output and `refused:true`, with no actual second exit. This contradicts AGENTS §9 (behavior verifiable/diagnosable) and SOP §2 (actual process completion and assertions, rather than a parent PASS). Construct a fresh command with the same executable, arguments, working directory and **existing private environment**; require an actual `*exec.ExitError` / expected native exit and admission-refusal evidence before comparing the retained store inventory. Do not call the allocating fixture builder again for that restart.

## Smell judgements

No additional actionable smell. The local private namespace helper and separate protocol fixture preserve real-use boundaries; no speculative production abstraction is warranted.

Summary: **1 documented finding (P2); 0 additional smell findings.**


## Spec

## Spec

只读审查 `1992e875136ebc332a19255e3729c448b001894b...cd91dbfa257692b744ba92708d0770522561ba07`，对照 Spec R23/E18、工单 18/19；未测试、构建、启动 native、请求厂商或操作 claim。

- **[P2] 新增“拒绝重启后库存不变”检查没有真正重启进程（部分实现/验证错误）。** 工单 [19:8](/Users/cyber/.codex/worktrees/grok-native-proof/barness/.scratch/barness-ai-grok/issues/19-controlled-grant-store-isolation.md:8) 要求“保留实际源/commit发现及原有断言”，[19:20](/Users/cyber/.codex/worktrees/grok-native-proof/barness/.scratch/barness-ai-grok/issues/19-controlled-grant-store-isolation.md:20) 要求保留“真实执行状态”。[grok_history_minimax_grant_test.go:64](/Users/cyber/.codex/worktrees/grok-native-proof/barness/ai/e2e/grok_history_minimax_grant_test.go:64) 新增库存断言依赖 :61 对已执行过的同一 `exec.Cmd` 再次调用 `CombinedOutput()`；该调用会直接返回 `exec: Stdout already set`，没有第二个 native 进程。因此 :62 的非 nil 错误与 :65 的库存相等不能证明同一 grant 的原生拒绝或重启仍使用原私有 namespace。留存三份 `restart.json` 也均为空 output、`refused:true`，没有原生数字退出或拒绝原件。应新建命令对象，保持原 Args/Dir/Env/同一 namespace，实际启动并核对原生拒绝、零新增活动及库存不变；不能再次调用会创建新 namespace 的 fixture。旧代码虽已有命令复用，本次新增库存验证及摘要仍把它当成真实重启证明，未满足本票边界。

未发现其他需求缺失、范围扩大或实现错误。工单 18 的请求仅改变 M3 thinking-on；实际子归档确认三协议正例通过、同任务零 thinking 仍为原生 FAIL，完整 14 场景/原预算保留，受控结果没有取得正式资格。工单 19 的私有 common namespace、真实 HEAD/工作树、父子 store 一致与环境清理均有实现及留存检查。两个索引共 34 个引用哈希有效；历史偏差原件保留。完整普通、重新冻结和未来付费/人工阶段的 pending 不计为 finding。

合计：1 项；最高 P2。


Standards：1项P2；Spec：1项P2。同一真实restart验证缺口交由原单一实现者修复，不代表新来源普通、付费或发布完成。
