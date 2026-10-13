# 01: 核查 CallTimeout 与 ReadIdleTimeout 同为 100ms 时的终态竞争

**Status:** needs-triage
**Type:** task
**Updated:** 2026-10-13

工单 11 的最终普通全量（开启 Pi 差分）与全量 race 同时运行。既有
`TestDecisionsRuntimeTimeouts/body-call-deadline` 在普通全量中失败：Code/Phase 均为
`deadline_exceeded/response`，但得到读取 idle 的 StopReason:error，断言期待调用 deadline
的 StopReason:aborted。两项策略均为 100ms；该路径不使用本票的 Host CLI/collector。

原失败证据完整保留于
`.evidence/barness-ai/20261013T022246.481993000Z/P10-OD42-runtime-timeout-body-call-deadline`，
归档见 [原失败](../../barness-ai-openai-decisions/host-tooling-evidence/timeout-failure.zip)。
单项原测试复测退出 0。没有修改原断言、协议实现或将原失败改成 PASS。

```sh
BARNESS_AI_PIDIFF=1 go test ./... -count=1
go test -race ./... -count=1
go test ./ai/e2e -run '^TestDecisionsRuntimeTimeouts$/^body-call-deadline$' -count=1
```

待核查：同到期 timer 的处理顺序、并行运行负载和调度；本次仅单项复测通过，尚未建立高复现率
的最小负载闭环，不能据此宣称根因或修复。后续先用独立 bundle 的重复单项/受控负载闭环取证，
再决定是测试需要确定化还是取消语义的边界需要修复。不得通过放宽两种终态的原断言掩盖问题。
