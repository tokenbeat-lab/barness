# 区域路由无 Key 诊断

2026-10-18：7个固定官方主机全部连接通过。GLM、DeepSeek 直连；OpenAI、Anthropic、Google、TypeSafe、MiniMax 国际站经 `http://127.0.0.1:7890` HTTP CONNECT 后进行普通 TLS 证书校验。

只执行 DNS/TCP、CONNECT 和 TLS handshake，没有读取 Key、发送 API 请求或生成调用。MiniMax 仅检查网络，不重跑其既有正式 PASS。失败原因与时长按主机记录；本结果不证明地区、账户权限、服务可用型号、支持资格或付费用量。

[计划](plan.json)、[实际结果](results.json)、[SHA256](checksums.json)。原始收据另存主仓库 `.evidence/minimax-routed-network-diagnostics-1/`。后续实际验收仍按票13单独冻结与计量。
