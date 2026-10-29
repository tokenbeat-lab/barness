# xAI Images 本地代理增量审查

## Standards

已完成：代理小修复的静态审查。行动问题为 0。

审查范围为 `414a2c36..fb1dbb85` 的三个 Go 文件。依据为 AGENTS.md、测试流程 SOP、ADR-0056 与既有 xAI 代理接线。

已验证：静态审查确认以下边界。

- `ai/live/xai_images_config_test.go:110` 复用 `smokeProxy`。它只接受显式的 loopback HTTP 地址。userinfo、query、fragment、远程地址和路径变体均拒绝。
- `ai/live/xai_images_harness_test.go:52` 只修改原生验证 transport。生产 Client 保持原状。`newRecorder` 默认禁用环境代理。
- `ai/live/xai_images_harness_test.go:58` 在受控代理返回 CONNECT 响应时立即终止。客户端不会建立 TLS 隧道或发送厂商 HTTP。受控模式继续拒绝真实凭据。
- 固定官方目标、六场景预算、零重试与一次性授权合同保持。报告记录显式代理选择。旧失败证据没有改写路径。
- 新增公共子进程检查覆盖代理拒绝、真实本地 CONNECT 和原六场景 fixture。三个 Go 文件均少于 500 行。代码复用既有代理校验，没有新增生产抽象。

未验证：未运行测试或厂商请求。运行结果由本次独立验证回执证明。

Standards 共 0 项，无待修复问题。

## Spec

已完成：`414a2c36..fb1dbb85199d2b47a4ce333c8d583cff95bcbab7` 代理增量审查。可行动缺陷：0。

- 工单 05 要求「复用现有 smokeProxy 的严格 loopback HTTP 校验」。`ai/live/xai_images_config_test.go:110–120` 校验显式代理，并额外拒绝空片段和 RawPath。`ai/live/proxy_test.go:15–27` 拒绝非 loopback IP、无效端口、非 HTTP、userinfo、query 与额外路径。controlled proxy 缺代理时在 claim 前拒绝。
- 工单要求「不扩大目标/预算」。`ai/live/xai_images_harness_test.go:52–60` 只配置现有 transport。受控代理的 OnProxyConnectResponse 对所有响应返回错误，HTTP 200 也不能建立 TLS 隧道。`:78–85` 仍固定 xAI endpoint 和 6 次发送上限；场景与 USD0.60 上限未变。生产 Client 默认代理无改动。
- 规格第 208 行要求「固定自身身份、来源、二进制、目录和实际输入」。`ai/live/xai_images_config_test.go:31`、`:116` 将代理选择写入原生报告。既有新授权、源码注入和一次性 claim 仍生效。旧失败及旧授权未复用。
- 工单要求「无厂商请求」「普通 fixture 仍完成 6 场景」。`ai/e2e/xai_images_live_process_test.go:188–215` 断言代理只收到一次 CONNECT api.x.ai:443，Authorization 与 Proxy-Authorization 均为空。失败后其余场景保持 NOT_RUN。`:182–186`、`:219–228` 保留代理配置和完整 fixture 的有限计划验收。

已验证：只读核对既有回执 `20261029T011128.326343Z-huqny4b5`。命令 `go test -json ./ai/e2e -run '^TestXAIImagesLiveProcess$' -count=1` 退出 0。29 个测试事件 PASS，audit findings 为 0。3 份 Go 增量文件的实际来源哈希一致。HTTP 200/502 的代理场景均确认零厂商 HTTP；fixture 完成 6/6。

未验证：真实代理链路与真实厂商响应。本次未重跑测试或发起厂商调用。此结论不授予新的付费执行权限，不改变原 live 失败结论。

Spec：0 项。
