# 19: Google × Gemini Developer API（P03）

**What to build:** 租户可通过 Gemini Developer API binding 完成文本、function call/result、thought signature 回放与图片输入，通过全部适用横向场景（spec I4、I5、P03）。

**Blocked by:** 18

**Status:** ready-for-agent

- [ ] 使用 Google GenAI Go SDK（研究锁定 v1.71.0 为起点，纳入时重新核实）；先核查必需字段能否经公开扩展或原始响应无损保留，否则该组合改用直接 HTTP，并跑同一组场景
- [ ] ToolCall 增加区分缺失/null/空值的 thoughtSignature；跨模型回放按 pi 原 truthy 条件删除非空值、不合并缺失/null/空值（自 08 移交，见 `replay.go` replayContent）
- [ ] 区分 thought 与 thoughtSignature；回放保留签名；未知必需字段走原始通道，不先丢字段再声称差分一致
- [ ] 无 finishReason 的 EOF 为错误；各结束原因映射到 StopReason
- [ ] function call/result 往返；工具结果图片路由；支持图片与占位降级分别覆盖
- [ ] Google level/budget reasoning 映射；忽略 samplingParams
- [ ] 保留 onPayload；不调用 onResponse；拒绝非默认 fetch 的基线行为不被悄悄改写（新增支持须登记扩展）
- [ ] usage 与成本；不以 Vertex 用例替代 Developer API
- [ ] 复用 E01–E09、E11 适用用例全部通过；pi 差分无待处理差异
