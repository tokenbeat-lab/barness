# 01：定位 race 下 unary 调用截止的终态不一致

Status: resolved

来源：Jev 图片整合票 14 在 `4310a53` 的完整 normal/race 验收；不是新增图片语义的交付声明。

当前 race 中 `P07-E08-unary-timeout-body-call-deadline` 失败：预期 `StopReasonAborted`，实际
`StopReasonError`，错误消息为 `ReadIdleTimeout (100ms)`。其余该 case 的类型、阶段、
空答案、实际 Attempt、资源释放及随后调用断言通过。原因尚未定位，不将其判为无害波动。
相关测试 `ai/e2e/unary_lifecycle_test.go` 与整合起点 `bf9e857` 一致；调用与读闲限额均为 `short`。

原始证据：`.evidence/ticket14/final-4310a53/gate/race/20261017T001525.154013000Z/P07-E08-unary-timeout-body-call-deadline/`。
完整报告：`.scratch/jev-1.14-multimodal-box/14-verification.md`，完整性回执 `14-evidence.json`。

复查：

```sh
go test -race ./ai/e2e -count=1 -run '^TestUnaryTransportTimeouts$/^body-call-deadline$'
```

先保留失败输入、时钟/实际预算、取消与读闲的时序证据，确认合同要求再修复；不通过放宽断言或
单次重跑清除本次失败。完成条件是解释并修复终态差异，在原公开入口与完整 normal/race 下
保留可重复证据、全部资源释放。此票不涉及收费服务授权。

## Comments

2026-10-17：确认 deadline 已过而取消回调传播延后，较晚读闲回调抢先记录原因。
公开 Client/HookedClient 的延迟传播反例先2RED，反向对照2PASS；f7c3aff只为严格早于调用
截止的尝试限额计时。原single-P独立race复现修复后300轮PASS，相关隔离回归通过。
本版完整normal/race各9,098 PASS、0FAIL，原P07-E08-unary-timeout-body-call-deadline
两阶段均PASS；所有受控命令/审计通过。按原验收关闭，真实服务仍NOT_RUN。

原失败及上轮证据保持不变。详见`../verification.md`、`../investigation.md`和`../evidence.json`。
