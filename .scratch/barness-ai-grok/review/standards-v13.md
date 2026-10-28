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
