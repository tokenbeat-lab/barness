# 29: SDK 修复后移除 deliveredBody 兜底

**What to build:** 删除 `ai/sdk_middleware.go` 的 `deliveredBody`/`closeOnce` 及其在 `ai/responses.go`、`ai/anthropic.go` 的接线，前提是两家 SDK 自己会关闭“调用方 context 结束时恰好到达”的响应体。

**Blocked by:** —

**Status:** needs-triage

**Context:** 工单 17 的 E06 场景 `a-canceled-on-arrival` 发现：openai-go v3.66.0 与 anthropic-sdk-go v1.75.0 的 `internal/requestconfig` 在 `handler(req)` 返回响应后，若调用方 context 已结束，直接 `return ctx.Err()`，既不交回响应（`WithResponseInto` 未赋值）也不关闭响应体，宿主 transport 为该响应持有的资源因此不被释放。barness 以 SDK 中间件记录交付的响应体，并在失败尝试上关闭（关闭一次语义）。

**影响与风险：** 兜底只多一层 body 包装与一次幂等 Close；若 SDK 改变中间件顺序，所有 Close 仍经过 `closeOnce`，不会对宿主 body 重复关闭。

**移除条件：** 升级到在上述路径关闭 `res.Body` 的 SDK 版本后，删除兜底，`go test -race -count=20 -run 'TestTenantIsolation/.*/a-canceled-on-arrival' ./ai/e2e` 仍全部通过（无未关闭响应体）。

- [ ] 核实两家 SDK 的新版本已在 caller ctx 结束分支关闭响应体
- [ ] 删除 `deliveredBody`、`closeOnce` 与接线，E06 `a-canceled-on-arrival` 在 `-race` 下重复通过
