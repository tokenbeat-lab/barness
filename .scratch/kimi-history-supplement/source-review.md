# Kimi 六路线旧九项的当前来源审阅

2026-10-27：原测量提交 d291e9b5eba87f717d3eb3fcee56c95649806a2e、9c0e5205788b0fdcf73cda95d9191b386868b1a2、92d4f63fbcfbaa26f029931b327e0dffc47e99f4。每个差异原字节从固定 Git 提交读取并与原 regional-execution-source 哈希相等；完整 diff 与78条确切前后哈希分别见 `.evidence/kimi-history-supplement/*-runtime.patch`、`source-review-deltas-v1.json`、`reviewed-sources-added-v1.json`。

按当前协议主入口归属，Chat/Messages 各10个保守差异，Responses各19个：

- binding 新增 xAI 常量；catalog_validate / preflight 的 xaiModelProblem / xaiBindingProblem 在非xAI入口立即无操作；call/history/hooks 的新分支均由目标ProviderXAI约束。
- cost_estimate 和 openai_sdk 是注释；retry 只为xAI增加unreported初态，原Kimi路径相同。
- Responses 的 capabilities/options/assistant/history/errors/request/stream 新增xAI实现；所有xAI最终校验/状态构造对Kimi返回nil，Kimi自身的分支和summary投影不变。responses_request 的 system/developer角色新排除xAI，不改变原Kimi排除条件。
- opaque_history_ids 确实扩展了“源为xAI”的归类：这属于新增行为，由缺失的 xai-origin-history 单项另行验证；原九项均为自身Kimi来源，没有新增xAI源，不能用它们证明这项新行为。

当前公共 Client 在无 dialer 的内存 RoundTripper 中重建原54个请求，无整包OnPayload替换；助手状态只从同bundle实际Result且tenant/account一致恢复TrustNativeState。54/54 wire结构相同，48个正常响应的content/usage/stop/responseID相同，6取消结果在原捕获的解码内容/用量前缀内。原完整性、独立wire/Result/Observer、固定native/tool前序及取消语义继续在原生校验中使用，回放不增加任何厂商事实。

回放原件：`.evidence/kimi-history-supplement/public-replay-v1/{receipt.json,stdout.jsonl}`。来源原先拒绝的RED及修正后的GREEN见source-red/green-command.log对应的新attempt；只追加Kimi Provider/API的确切字节审阅，不把文件名或未来修改永久放行，依赖/新增运行时/错来源/错Provider继续拒绝。

因此旧九项可进入各自组合的baseline贡献，原候选标签、9-call预算、时间、授权和账本不变。六个新增真实场景仍缺失，目录仍inactive，不能由本回放取得十项或正式支持。最后随完整普通验证再绑定最终消费者来源。
