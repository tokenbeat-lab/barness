# 28: Anthropic 服务端备用模型（可选特性，claude-fable-5）

**What to build:** 支持声明了服务端备用模型的 Anthropic 模型（pi compat `allowedFallbackModels`）：主模型不可用时由 Anthropic 改用备用模型，结果记录实际服务的模型并按其价格计费；备用模型只发送 binding 同样允许的那些（维护者决定 2026-10-02）。按 ADR-0018 这是可选特性，claude-fable-5 由工单 34 先行列入（不发送 `fallbacks`），本工单完成后对它生效。

**Blocked by:** 16

**Status:** ready-for-agent

**Context:** 基线为 pi-ai 0.87.1 `anthropic-messages.ts`：`buildParams` 的 `fallbacks`、`shouldUseServerSideFallbackBeta`、`message_start` 中的 `responseModel` 与 fallback 计价、`content_block_start` 的 `fallback` 块（输出开始后的回退已按 pi 报错，见工单 16）。spec §4"实际可调用范围是目录与 binding 授权的交集"：备用模型等于让厂商代为调用另一模型，所以只发送 `Binding.AllowedModels` 也包含的备用模型，这是相对 pi 的授权扩展，登记在 pi 差分 ledger。claude-fable-5 的备用模型 claude-opus-4-8 由工单 34 列入，claude-opus-5 由工单 27 列入。

- [ ] `ModelCompat.AllowedFallbackModels`（provider、model、价格，与 pi 数据同形，参与目录哈希与价格校验）
- [ ] 请求带 `fallbacks: [{model}]`，只含 binding 允许的备用模型；过滤后为空时既不发 `fallbacks` 也不发 `server-side-fallback-2026-07-01` beta，否则发送该 beta
- [ ] `message_start` 报告的模型不同于请求模型时记 `responseModel`；该模型是已发送的备用模型时按其价格计算成本，否则按主模型价格（pi 规则）
- [ ] payload 授权：回调仍不能新增或改写 `fallbacks`，只能保留 adapter 发出的列表
- [ ] 决定 Result/观测元数据是否另记实际服务的模型（`CallMetadata.ModelID` 仍为授权的请求模型），结论记入 ADR
- [ ] claude-fable-5 的目录条目携带 `AllowedFallbackModels`（目录版本与 pin 同步升级）；删除工单 34 为其登记的 fallbacks 差分扩展
- [ ] 离线 E2E：备用模型全部允许、部分允许、全不允许，回退发生与未发生，回退后的成本；pi 差分无待处理差异（过滤差异按扩展登记）
