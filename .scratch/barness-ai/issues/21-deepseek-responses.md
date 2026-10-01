# 21: DeepSeek × Responses（P05）

**What to build:** 租户可通过独立的 DeepSeek Responses binding 完成文本、工具往返、推理与完整历史回放；复用 Responses adapter，但 Provider 身份、能力配置、endpoint 与 key 和 OpenAI 完全独立（spec I4 末两段、P05）。

**Blocked by:** 18

**Status:** ready-for-agent

- [ ] 独立能力配置（endpoint `https://api.deepseek.com`）；Result 中 ProviderID 为 DeepSeek，与 OpenAI Responses 共享 adapter 时不串配置
- [ ] 处理 response.reasoning_text.delta / done、incomplete、failed
- [ ] 请求不自动注入 previous_response_id、conversation、store；历史通过完整输入回放；不支持字段的行为单列 fixture
- [ ] 工具往返与推理历史回放
- [ ] 本组合属于冻结 pi 路由之外的扩展：使用研究记录的官方协议 fixture 证明，登记为扩展，不计为 pi 差分通过
- [ ] 复用 E01–E09、E11 适用用例全部通过
