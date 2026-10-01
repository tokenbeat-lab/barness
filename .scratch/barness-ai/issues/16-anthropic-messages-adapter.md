# 16: Anthropic × Messages（P02）

**What to build:** 租户可通过 Anthropic Messages binding 完成文本、工具往返、thinking 签名回放、图片输入与缓存计费，全部经同一 Client 入口，并通过已在 Responses 上建立的所有适用横向场景（spec I4、I5、P02）。

**Blocked by:** 05, 06, 08, 09, 10, 11, 15

**Status:** ready-for-agent

- [ ] 使用 Anthropic Go SDK（研究锁定 v1.75.0 为起点，纳入时重新核实），关闭默认重试，transport 约束与 01 一致
- [ ] 以 message_stop 判定成功终态；缺终态 EOF、SSE error 事件为错误终态
- [ ] signed empty thinking、redacted thinking、交错块、input_json_delta / signature_delta 正确归一与回放
- [ ] 工具往返、图片与不支持图片占位分别覆盖
- [ ] Anthropic adaptive effort / token budget 映射；非 OpenAI-compatible 路径忽略 samplingParams
- [ ] 替换 payload 后仍强制 stream=true；onResponse 时点与 Responses 一致
- [ ] 缓存读写、1h 写入与计价
- [ ] 复用 E01–E05、E11 适用用例（full/simple × Stream/Complete、不读事件的 Result、成功与失败终态）全部通过
- [ ] pi 差分无待处理差异
