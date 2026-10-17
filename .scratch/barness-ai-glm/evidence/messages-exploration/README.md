# Messages 公共 Client 契约探索（2026-10-17）

这些是 R13/T02 的先行 wire 取证，使用现有公共 Client、显式 Host 候选和可信 payload 回调，不是 GLM 默认生产路径或正式支持通过报告。计划、逐次预算、SDK 与目录 Hash、源码、最终请求/SSE/事件/Result/Attempt/usage、脱敏审计和文件校验均保留在各独立 bundle；[索引](index.json) 记录归档 Hash。原始完整产物复制到根目录 `.evidence/barness-ai-glm/protocol-exploration/messages/`，同目录 ZIP 按项目约定不纳入 Git。

累计 13 次 Logical Call、13 次保守本地 Attempt 预留、11 次外部 HTTP 请求；零重试。首次 2 次被本地精确 URL guard 拒绝（SDK 自带 `?beta=true`），没有外部网络请求；该失败原样保留。随后只放开这一 SDK 查询参数，两型号基础调用、工具往返、签名历史回放均通过。最后三次只改变预登记变量，验证 5.2 默认不带 thinking 参数，以及两型号回放时省略 signature 字段。全部 bundle 通过 Key 脱敏审计，不能把本地失败记作 Provider 成功。

已证实：两型号工具定义/auto、工具参数分片和 Host 工具结果续接；5.3 默认可观察 thinking；5.2 的确切 enabled/budget1024 形状返回 signed thinking 并可续接；实际签名回放可用；两型号均接受只省略 signature 字段的 thinking 历史。没有证明 Provider 会返回 unsigned thinking，也没有测试 signature 空字符串、redacted thinking、任意 budget 范围、budget 强制执行、effort/adaptive 映射、强制工具选择或缓存控制。5.2 同一复杂提示省略 thinking 后只返回文本。

实际用量在 message_delta 更新，cache_creation 缺失应保持 partial；初始 message_start 的零不能覆盖最终值。实际响应头有 X-Request-Id，现有原生 Anthropic request-id lookup 未捕获它，需生产适配后再由公共 Client 验证。签名与思考是低权限合成任务的原生 wire 证据；Observer 仍只记录元数据。账户/区域别名不构成真实身份、权益或账单证明。

后续工单 06/07 必须先将确切形状转成失败的离线公共 E2E，再完成生产 Provider 配置和回放；本探索不能替代三组合两型号的正式 live、内置目录或发布门禁。
